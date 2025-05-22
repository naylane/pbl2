package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// comunicação API REST
type ReservaRequest struct {
	PlacaVeiculo string   `json:"placa_veiculo"`
	Pontos       []string `json:"pontos"`
	EmpresaID    string   `json:"empresa_id"`
}

type ReservaResponse struct {
	Status    string `json:"status"`
	Ponto     string `json:"ponto"`
	Mensagem  string `json:"mensagem"`
	EmpresaID string `json:"empresa_id"`
}

var reservas_mutex sync.Mutex
var reservas = make(map[string]map[string]string)

var status_ponto = struct {
	sync.RWMutex
	status map[string]bool
}{status: make(map[string]bool)}

var ponto_locks = make(map[string]*sync.Mutex)

// IP do pc lab em uso
var ip_pc string = "172.16.103.14"

func inicializa_rest(porta string) {
	http.HandleFunc("/api/regiao", handleRegiaoJson)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/admin/ponto/status", handleStatusPonto)
	http.HandleFunc("/api/confirmar-prereserva", handleConfirmaPreReserva)
	http.HandleFunc("/api/reserva", handleReserva)
	http.HandleFunc("/api/cancelamento", handleCancelamento)

	fmt.Printf("[Servidor REST] iniciado - porta %s\n", porta)
	endereco := "0.0.0.0:" + porta
	go func() {
		if err := http.ListenAndServe(endereco, nil); err != nil {
			fmt.Printf("Erro ao iniciar servidor REST: %v\n", err)
		}
	}()

}

func handleRegiaoJson(responseW http.ResponseWriter, request *http.Request) {
	responseW.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseW).Encode(dadosRegiao)
}

func handleStatusPonto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método inválido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Ponto  string `json:"ponto"`
		Status bool   `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Atualiza status do ponto
	status_ponto.Lock()
	status_ponto.status[req.Ponto] = req.Status
	status_ponto.Unlock()

	// Se desconectou, cancela reservas existentes
	if !req.Status {
		cancelaReservasPorPontosDesconectados(req.Ponto)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ponto":      req.Ponto,
		"status":     req.Status,
		"empresa_id": empresa.Id,
	})

	fmt.Printf("[INFO] Status do ponto %s alterado para: %v.\n", req.Ponto, req.Status)
}

// status dos pontos
func inicializaMonitoramentoDosPontos() {
	for _, ponto := range dadosRegiao.PontosDeRecarga {
		ponto_locks[ponto.Cidade] = &sync.Mutex{}
	}

	//pontos da própria empresa como conectados
	status_ponto.Lock()
	for _, ponto := range dadosRegiao.PontosDeRecarga {
		pertenceEmpresa := false
		for _, pontoDaEmpresa := range empresa.Pontos {
			if ponto.Cidade == pontoDaEmpresa {
				pertenceEmpresa = true
				break
			}
		}
		if pertenceEmpresa {
			status_ponto.status[ponto.Cidade] = true
		} else {
			status_ponto.status[ponto.Cidade] = false
		}
	}
	status_ponto.Unlock()
	//imediatamente após inicialização
	statusDosPontos()

	//goroutine para verificar periodicamente conexão dos pontos
	go func() {
		for {
			time.Sleep(30 * time.Second)
			statusDosPontos()
		}
	}()
}

func statusDosPontos() {
	for _, ponto := range empresa.Pontos {
		ponto_conectado := pontoEstaConectado(ponto)

		status_ponto.Lock()
		status_anterior := status_ponto.status[ponto]
		status_ponto.status[ponto] = ponto_conectado
		status_ponto.Unlock()

		if status_anterior != ponto_conectado {
			if ponto_conectado {
				fmt.Printf("[INFO] Ponto %s está conectado.\n", ponto)
			} else {
				fmt.Printf("[AVISO] Ponto %s está desconectado.\n", ponto)
				cancelaReservasPorPontosDesconectados(ponto)
			}
		}
	}
}

func pontoEstaConectado(ponto string) bool {
	pertenceEmpresa := false
	for _, pontoDaEmpresa := range empresa.Pontos {
		if ponto == pontoDaEmpresa {
			pertenceEmpresa = true
			break
		}
	}

	// Se o ponto pertence a esta empresa - está conectado
	if pertenceEmpresa {
		return true
	}

	pontoObj, _ := GetPontoPorCidade(ponto)
	return pontoObj.ID%2 == 0
}

func cancelaReservasPorPontosDesconectados(ponto string) {
	reservas_mutex.Lock()
	defer reservas_mutex.Unlock()

	for placa, pontosMap := range reservas {
		//este veículo tem reserva para este ponto
		if _, reservado := pontosMap[ponto]; reservado {
			// Obtem o índice do ponto
			pontoObj, i := GetPontoPorCidade(ponto)
			if pontoObj.Reservado == placa {
				// Limpa a reserva
				dadosRegiao.PontosDeRecarga[i].Reservado = ""
				delete(pontosMap, ponto)
				salvaDadosPontos()

				fmt.Printf("[AVISO] Reserva para %s no ponto %s cancelada devido à desconexão.\n", placa, ponto)

				// Notifica o cliente
				client := getClienteMqtt()
				publicaMensagemMqtt(client, "mensagens/cliente/"+placa, fmt.Sprintf("ponto_desconectado,%s,Reserva cancelada por desconexão", ponto))
			}
		}
	}
}

// entre servidores
func requisicaoRest(metodo, url string, corpo interface{}, resposta interface{}) error {
	json_corpo, erro := json.Marshal(corpo)
	if erro != nil {
		return fmt.Errorf("erro ao codificar JSON: %v", erro)
	}

	req, erro := http.NewRequest(metodo, url, bytes.NewBuffer(json_corpo))
	if erro != nil {
		return fmt.Errorf("erro ao criar requisição: %v", erro)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, erro := client.Do(req)
	if erro != nil {
		return fmt.Errorf("erro ao fazer requisição: %v", erro)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status de resposta inválido: %d", resp.StatusCode)
	}

	if resposta != nil {
		if err := json.NewDecoder(resp.Body).Decode(resposta); err != nil {
			return fmt.Errorf("erro ao decodificar resposta: %v", err)
		}
	}

	return nil
}

// coordenar reservas com outros servidores via REST
func handleReservaRest(placaVeiculo string, pontos []string) bool {
	// Filtra pontos que não pertencem a este servidor
	var pontos_de_outros_servidores []string
	for _, ponto := range pontos {
		pertenceEmpresaAtual := false
		for _, pontoDaEmpresa := range empresa.Pontos {
			if ponto == pontoDaEmpresa {
				pertenceEmpresaAtual = true
				break
			}
		}
		if !pertenceEmpresaAtual {
			pontos_de_outros_servidores = append(pontos_de_outros_servidores, ponto)
		}
	}

	//não há pontos para outros servidores -> sucesso
	if len(pontos_de_outros_servidores) == 0 {
		return true
	}

	ip_servidor_atual := ip_pc
	var porta string
	switch empresa.Id {
	case "001":
		porta = "8081"
	case "002":
		porta = "8082"
	case "003":
		porta = "8083"
	default:
		porta = "8080"
	}
	meuEndereco := fmt.Sprintf("http://%s:%s", ip_servidor_atual, porta)

	servidores := []string{
		"http://172.16.103.11:8081",
		"http://172.16.103.12:8082",
		"http://172.16.103.14:8083",
	}

	// Remove o próprio servidor
	var outros_servidores []string
	for _, serv := range servidores {
		if serv != meuEndereco {
			outros_servidores = append(outros_servidores, serv)
			fmt.Printf("[INFO] Servidor %s sendo adicionado para coordenação.\n", serv)
		}
	}

	//requisição para outros servidores
	req := ReservaRequest{
		PlacaVeiculo: placaVeiculo,
		Pontos:       pontos_de_outros_servidores,
		EmpresaID:    empresa.Id,
	}

	//Envia
	todasConfirmadas := true
	var respostasServidores []ReservaResponse
	for _, servidor := range outros_servidores {
		var resposta ReservaResponse
		url := servidor + "/api/reserva"
		fmt.Printf("[INFO] Requisição enviada para %s.\n", url)

		erro := requisicaoRest("POST", url, req, &resposta)
		if erro != nil {
			fmt.Printf("[ERRO] comunicação com o servidor %s: %v.\n", servidor, erro)
			todasConfirmadas = false
			break
		}

		fmt.Printf("[Resposta] %s: %s.\n", servidor, resposta.Status)
		if resposta.Status == "falha" {
			fmt.Printf("[ERRO] Reserva não realizada em %s: %s.\n", servidor, resposta.Mensagem)
			todasConfirmadas = false
			respostasServidores = append(respostasServidores, resposta)
			break
		} else if resposta.Status == "confirmado" {
			fmt.Printf("[SUCESSO] Reserva confirmada em %s para o ponto %s.\n", servidor, resposta.Ponto)
			respostasServidores = append(respostasServidores, resposta)
		} else if resposta.Status == "ignorado" {
			//fmt.Printf("[AVISO] Servidor %s ignorou a solicitação: %s.\n", servidor, resposta.Mensagem)
		}
	}

	if !todasConfirmadas {
		fmt.Printf("[AVISO] Cancelando reservas. Falha em algum servidor\n")
		for _, resposta := range respostasServidores {
			if resposta.Status == "confirmado" {
				fmt.Printf("[INFO] Cancelando reserva no servidor %s.\n", resposta.EmpresaID)
				cancelaReservaRest(resposta.EmpresaID, placaVeiculo, pontos)
			}
		}
	}

	return todasConfirmadas
}

func reservaPontosEmOutrosServidores(placaVeiculo string, pontos []string) bool {
	ip_servidor_atual := ip_pc
	var porta string
	switch empresa.Id {
	case "001":
		porta = "8081"
	case "002":
		porta = "8082"
	case "003":
		porta = "8083"
	default:
		porta = "8080"
	}
	endereco_atual := fmt.Sprintf("http://%s:%s", ip_servidor_atual, porta)

	servidores := []string{
		"http://172.16.103.11:8081",
		"http://172.16.103.12:8082",
		"http://172.16.103.14:8083",
	}

	req := ReservaRequest{
		PlacaVeiculo: placaVeiculo,
		Pontos:       pontos,
		EmpresaID:    empresa.Id,
	}

	// Envia
	sucesso_em_algum := false

	for _, servidor := range servidores {
		if servidor == endereco_atual {
			continue // Pula o próprio servidor
		}

		var resposta ReservaResponse
		url := servidor + "/api/reserva"

		err := requisicaoRest("POST", url, req, &resposta)
		if err != nil {
			fmt.Printf("Erro na comunicação com servidor %s: %v\n", servidor, err)
			continue
		}

		if resposta.Status == "confirmado" {
			sucesso_em_algum = true
			fmt.Printf("[SUCESSO] Servidor %s reservou ponto %s.\n", resposta.EmpresaID, resposta.Ponto)
		}
	}

	return sucesso_em_algum
}

func handleConfirmaPreReserva(responseW http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseW, "Método inválido", http.StatusMethodNotAllowed)
		return
	}

	var req ReservaRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(responseW, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	placaVeiculo := ""
	if strings.HasPrefix(req.PlacaVeiculo, "CONFIRM_") {
		placaVeiculo = req.PlacaVeiculo[8:] // Remove "CONFIRM_"
	} else {
		placaVeiculo = req.PlacaVeiculo
	}

	ponto_localizado := false
	resposta := ReservaResponse{
		EmpresaID: empresa.Id,
	}

	for _, ponto_solicitado := range req.Pontos {
		lock := ponto_locks[ponto_solicitado]
		lock.Lock()
		for _, pontoDaEmpresa := range empresa.Pontos {
			if ponto_solicitado == pontoDaEmpresa {
				ponto_localizado = true
				pontoRecarga, index := GetPontoPorCidade(ponto_solicitado)
				if pontoRecarga.Reservado == "PRE_"+placaVeiculo || pontoRecarga.Reservado == placaVeiculo {
					dadosRegiao.PontosDeRecarga[index].Reservado = placaVeiculo
					salvaDadosPontos()
					resposta.Status = "confirmado"
					resposta.Ponto = ponto_solicitado
					resposta.Mensagem = fmt.Sprintf("Ponto %s reserva confirmada", ponto_solicitado)
					fmt.Printf("[SUCESSO] Pré-reserva do ponto %s convertida para reserva para %s.\n", ponto_solicitado, placaVeiculo)
				} else {
					resposta.Status = "falha"
					resposta.Ponto = ponto_solicitado
					resposta.Mensagem = fmt.Sprintf("Ponto %s não estava pré-reservado para %s", ponto_solicitado, placaVeiculo)
					fmt.Printf("[ERRO] O ponto %s não estava pré-reservado para %s.\n", ponto_solicitado, placaVeiculo)
				}
			}
		}
		lock.Unlock()
	}
	if !ponto_localizado {
		resposta.Status = "ignorado"
		resposta.Mensagem = "Pontos solicitados não pertencem a esta empresa"
	}
	responseW.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseW).Encode(resposta)
}

// com outros servidores via REST
func handlePreReservaRest(placaVeiculo string, pontos []string) bool {
	var pontos_em_outros_servidores []string
	for _, ponto := range pontos {
		pertence_esta_empresa := false
		for _, pontoDaEmpresa := range empresa.Pontos {
			if ponto == pontoDaEmpresa {
				pertence_esta_empresa = true
				break
			}
		}
		if !pertence_esta_empresa {
			pontos_em_outros_servidores = append(pontos_em_outros_servidores, ponto)
		}
	}

	//não há pontos para outros servidores - sucesso
	if len(pontos_em_outros_servidores) == 0 {
		return true
	}

	ip_servidor_atual := ip_pc
	var porta string
	switch empresa.Id {
	case "001":
		porta = "8081"
	case "002":
		porta = "8082"
	case "003":
		porta = "8083"
	default:
		porta = "8080"
	}
	endereco_atual := fmt.Sprintf("http://%s:%s", ip_servidor_atual, porta)

	servidores := []string{
		"http://172.16.103.11:8081",
		"http://172.16.103.12:8082",
		"http://172.16.103.14:8083",
	}

	// Remove o próprio servidor da lista
	var outros_servidores []string
	for _, serv := range servidores {
		if serv != endereco_atual {
			outros_servidores = append(outros_servidores, serv)
			fmt.Printf("[INFO] Servidor %s adicionado para coordenação da pré-reserva.\n", serv)
		}
	}

	//requisição para outros servidores
	req := ReservaRequest{
		PlacaVeiculo: "PRE_" + placaVeiculo, // Adiciona prefixo para identificar pre-reserva
		Pontos:       pontos,
		EmpresaID:    empresa.Id,
	}

	//Envia
	todas_confirmadas := true
	var respostasServidores []ReservaResponse

	fmt.Printf("[INFO] Pré-reservando pontos em %d outros servidores.\n", len(outros_servidores))
	for _, servidor := range outros_servidores {
		var resposta ReservaResponse
		url := servidor + "/api/reserva"
		fmt.Printf("[INFO] Enviando requisição para %s.\n", url)

		err := requisicaoRest("POST", url, req, &resposta)
		if err != nil {
			fmt.Printf("[ERRO] comunicação com o servidor %s: %v.\n", servidor, err)
			todas_confirmadas = false
			break
		}

		fmt.Printf("[Resposta] %s: %s.\n", servidor, resposta.Status)
		if resposta.Status == "falha" {
			fmt.Printf("[ERRO] Pré-reserva não realizada em %s: %s.\n", servidor, resposta.Mensagem)
			todas_confirmadas = false
			respostasServidores = append(respostasServidores, resposta)
			break
		} else if resposta.Status == "confirmado" {
			fmt.Printf("[SUCESSO] Pré-reserva confirmada em %s para o ponto %s.\n", servidor, resposta.Ponto)
			respostasServidores = append(respostasServidores, resposta)
		} else if resposta.Status == "ignorado" {
			//fmt.Printf("[AVISO] Servidor %s ignorou a solicitação de pré-reserva: %s.\n", servidor, resposta.Mensagem)
		}
	}

	if !todas_confirmadas {
		fmt.Printf("[AVISO] Cancelando pré-reservas devido a falha.\n")
		for _, resposta := range respostasServidores {
			if resposta.Status == "confirmado" {
				fmt.Printf("[INFO] Cancelando pré-reserva no servidor %s.\n", resposta.EmpresaID)
				cancelaReservaRest(resposta.EmpresaID, placaVeiculo, pontos)
			}
		}
	}

	return todas_confirmadas
}

func handleConfirmacaoPreReservaRest(placaVeiculo string, pontos []string) bool {
	var pontos_em_outros_servidores []string
	for _, ponto := range pontos {
		pertence_esta_empresa := false
		for _, pontoDaEmpresa := range empresa.Pontos {
			if ponto == pontoDaEmpresa {
				pertence_esta_empresa = true
				break
			}
		}
		if !pertence_esta_empresa {
			pontos_em_outros_servidores = append(pontos_em_outros_servidores, ponto)
		}
	}

	//não há pontos para outros servidores - retorna sucesso
	if len(pontos_em_outros_servidores) == 0 {
		return true
	}

	ip_servidor_atual := ip_pc
	var porta string
	switch empresa.Id {
	case "001":
		porta = "8081"
	case "002":
		porta = "8082"
	case "003":
		porta = "8083"
	default:
		porta = "8080"
	}
	meuEndereco := fmt.Sprintf("http://%s:%s", ip_servidor_atual, porta)

	servidores := []string{
		"http://172.16.103.11:8081",
		"http://172.16.103.12:8082",
		"http://172.16.103.14:8083",
	}

	var outrosServidores []string
	for _, serv := range servidores {
		if serv != meuEndereco {
			outrosServidores = append(outrosServidores, serv)
			fmt.Printf("[INFO] Servidor %s adicionado para confirmação de pré-reserva.\n", serv)
		}
	}

	//requisição para outros servidores com flag de confirmação
	req := ReservaRequest{
		PlacaVeiculo: "CONFIRM_" + placaVeiculo,
		Pontos:       pontos_em_outros_servidores,
		EmpresaID:    empresa.Id,
	}

	// Envia
	todasConfirmadas := true
	var respostasFalhas []string

	fmt.Printf("[INFO] Confirmando pré-reservas em %d outros servidores.\n", len(outrosServidores))
	for _, servidor := range outrosServidores {
		var resposta ReservaResponse
		url := servidor + "/api/confirmar-prereserva" // Novo endpoint específico
		fmt.Printf("[INFO] Enviando requisição de confirmação para %s.\n", url)

		err := requisicaoRest("POST", url, req, &resposta)
		if err != nil {
			fmt.Printf("[ERRO] Comunicação com o servidor %s: %v.\n", servidor, err)
			todasConfirmadas = false
			respostasFalhas = append(respostasFalhas, fmt.Sprintf("Erro de comunicação: %v", err))
			continue
		}

		if resposta.Status != "confirmado" {
			fmt.Printf("[ALERTA] Falha na confirmação em %s: %s.\n", servidor, resposta.Mensagem)
			todasConfirmadas = false
			respostasFalhas = append(respostasFalhas, resposta.Mensagem)
		} else {
			fmt.Printf("[SUCESSO] Pré-reserva confirmada em %s para ponto %s.\n", servidor, resposta.Ponto)
		}
	}

	if !todasConfirmadas {
		fmt.Printf("[AVISO] Falhas na confirmação: %v.\n", respostasFalhas)
	}

	return todasConfirmadas
}

func handleStatus(responseW http.ResponseWriter, request *http.Request) {
	responseW.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseW).Encode(map[string]string{
		"status":     "online",
		"empresa_id": empresa.Id,
	})
}

func handleReserva(responseW http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseW, "Método inválido", http.StatusMethodNotAllowed)
		return
	}

	var req ReservaRequest
	if erro := json.NewDecoder(request.Body).Decode(&req); erro != nil {
		http.Error(responseW, "[Erro] decodificar JSON", http.StatusBadRequest)
		return
	}

	ponto_localizado := false
	for _, ponto_solicitado := range req.Pontos {
		lock := ponto_locks[ponto_solicitado]
		lock.Lock()
		for _, pontoDaEmpresa := range empresa.Pontos {
			if ponto_solicitado == pontoDaEmpresa {
				ponto_localizado = true
				status_ponto.RLock()
				estaConectado := status_ponto.status[ponto_solicitado]
				status_ponto.RUnlock()

				if !estaConectado {
					responseW.Header().Set("Content-Type", "application/json")
					json.NewEncoder(responseW).Encode(ReservaResponse{
						Status:    "falha",
						Ponto:     ponto_solicitado,
						Mensagem:  fmt.Sprintf("Ponto %s desconectado", ponto_solicitado),
						EmpresaID: empresa.Id,
					})
					fmt.Printf("[ERRO] Tentativa para o ponto %s rejeitada: ponto desconectado.\n", ponto_solicitado)
					return
				}

				//Ponto encontrado, verificar disponibilidade
				ponto_recarga, i := GetPontoPorCidade(ponto_solicitado)
				reservas_mutex.Lock()
				resposta := ReservaResponse{
					Ponto:     ponto_solicitado,
					EmpresaID: empresa.Id,
				}
				if ponto_recarga.Reservado == "" || ponto_recarga.Reservado == req.PlacaVeiculo {
					//Livre ou reservado temp pelo próprio veículo
					dadosRegiao.PontosDeRecarga[i].Reservado = req.PlacaVeiculo
					salvaDadosPontos()
					//Registra a reserva no mapa de controle
					if _, existe := reservas[req.PlacaVeiculo]; !existe {
						reservas[req.PlacaVeiculo] = make(map[string]string)
					}
					reservas[req.PlacaVeiculo][ponto_solicitado] = "confirmado"

					resposta.Status = "confirmado"
					resposta.Mensagem = fmt.Sprintf("Ponto %s reservado com sucesso", ponto_solicitado)
					fmt.Printf("[SUCESSO] Ponto %s da empresa %s reservado para %s.\n", ponto_solicitado, empresa.Nome, req.PlacaVeiculo)

				} else {
					resposta.Status = "falha"
					resposta.Mensagem = fmt.Sprintf("Ponto %s reservado no momento", ponto_solicitado)
					fmt.Printf("[ERRO] O ponto %s da empresa %s encontra-se reservado no momento.\n", ponto_solicitado, empresa.Nome)
					reservas_mutex.Unlock()
					responseW.Header().Set("Content-Type", "application/json")
					json.NewEncoder(responseW).Encode(resposta)
					lock.Unlock()
					return
				}
				reservas_mutex.Unlock()
				responseW.Header().Set("Content-Type", "application/json")
				json.NewEncoder(responseW).Encode(resposta)
			}
		}
		lock.Unlock()
	}

	if !ponto_localizado {
		//Pontos não pertencem a este servidor
		responseW.Header().Set("Content-Type", "application/json")
		json.NewEncoder(responseW).Encode(ReservaResponse{
			Status:    "ignorado",
			Mensagem:  "Os pontos solicitados não pertencem a esta empresa.",
			EmpresaID: empresa.Id,
		})
	}
}

func handleCancelamento(responseW http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseW, "Método inválido", http.StatusMethodNotAllowed)
		return
	}
	var req ReservaRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(responseW, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}
	reservas_mutex.Lock()
	defer reservas_mutex.Unlock()
	resposta := ReservaResponse{
		EmpresaID: empresa.Id,
	}
	// Verifica se tem reservas para o veiculo
	if pontos_map, existe := reservas[req.PlacaVeiculo]; existe {
		for _, ponto_solicitado := range req.Pontos {
			lock := ponto_locks[ponto_solicitado]
			lock.Lock()
			if _, reservado := pontos_map[ponto_solicitado]; reservado {
				//reservado -> Cancela a reserva
				pontoRecarga, index := GetPontoPorCidade(ponto_solicitado)
				if pontoRecarga.Reservado == req.PlacaVeiculo {
					dadosRegiao.PontosDeRecarga[index].Reservado = ""
					salvaDadosPontos()
					delete(pontos_map, ponto_solicitado)
					fmt.Printf("[INFO] Reserva cancelada do ponto %s da empresa %s para %s.\n", ponto_solicitado, empresa.Nome, req.PlacaVeiculo)
					resposta.Status = "cancelado"
					resposta.Ponto = ponto_solicitado
					resposta.Mensagem = "Reserva cancelada"
				}
			}
			lock.Unlock()
		}
	}

	if resposta.Status == "" {
		resposta.Status = "nao_encontrado"
		resposta.Mensagem = "Reserva não encontrada para cancelamento"
	}

	responseW.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseW).Encode(resposta)
}

// em outro servidor
func cancelaReservaRest(empresaID string, placaVeiculo string, pontos []string) {
	ip_servidor_atual := ip_pc
	var porta string
	switch empresa.Id {
	case "001":
		porta = "8081"
	case "002":
		porta = "8082"
	case "003":
		porta = "8083"
	default:
		porta = "8080"
	}
	url := fmt.Sprintf("http://%s:%s", ip_servidor_atual, porta)

	req := ReservaRequest{
		PlacaVeiculo: placaVeiculo,
		Pontos:       pontos,
		EmpresaID:    empresa.Id,
	}

	var resposta ReservaResponse
	err := requisicaoRest("POST", url, req, &resposta)
	if err != nil {
		fmt.Printf("[ERRO] Reserva não realizada no servidor %s: %v.\n", empresaID, err)
	} else {
		fmt.Printf("[INFO] Reserva cancelada no servidor %s: %s.\n", empresaID, resposta.Status)
	}
}

func liberaPorTimeout(placa string, pontos []string, tempo time.Duration) {
	go func() {
		time.Sleep(tempo)
		for _, ponto := range pontos {
			ponto_locks[ponto].Lock()
			//reserva ainda existe
			if pontosMap, existe := reservas[placa]; existe {
				for _, ponto := range pontos {
					if _, reservado := pontosMap[ponto]; reservado {
						pontoRecarga, index := GetPontoPorCidade(ponto)
						if pontoRecarga.Reservado == placa {
							dadosRegiao.PontosDeRecarga[index].Reservado = ""
							delete(pontosMap, ponto)
							fmt.Printf("\nReserva para %s no ponto %s expirada por timeout\n", placa, ponto)
						}
					}
					ponto_locks[ponto].Unlock()
				}
			}
		}
		salvaDadosPontos()
	}()
}

func handleCancelaPreReservaRest(placaVeiculo string, pontos []string) bool {
	//cancelar pré-reservas em todos os servidores
	req := ReservaRequest{
		PlacaVeiculo: placaVeiculo,
		Pontos:       pontos,
		EmpresaID:    empresa.Id,
	}

	//requisições para todos os servidores para cancelar a pré-reserva
	sucessoEmTodos := true

	for _, servidor := range []string{"http://server1:8081", "http://server2:8082", "http://server3:8083"} {
		if servidor == fmt.Sprintf("http://server%s:808%s", empresa.Id[3:], empresa.Id[3:]) {
			continue // passa o próprio servidor
		}

		var resposta ReservaResponse
		url := servidor + "/api/cancelamento"

		err := requisicaoRest("POST", url, req, &resposta)
		if err != nil {
			fmt.Printf("[Erro] Comunicação com servidor %s: %v\n", servidor, err)
			sucessoEmTodos = false
			continue
		}

		if resposta.Status != "cancelado" {
			fmt.Printf("[ERRO] Cancelar pré-reserva em %s: %s.\n", servidor, resposta.Mensagem)
			sucessoEmTodos = false
		}
	}

	return sucessoEmTodos
}
