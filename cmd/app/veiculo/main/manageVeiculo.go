package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var placa string

//ok
func listCapitaisNordeste() {
	fmt.Println("\n======= Cidades com Servico de Recarga =======")
	fmt.Println("(1) - Salvador")
	fmt.Println("(2) - Aracaju")
	fmt.Println("(3) - Maceio")
	fmt.Println("(4) - Recife")
	fmt.Println("(5) - Joao Pessoa")
	fmt.Println("(6) - Natal")
	fmt.Println("(7) - Fortaleza")
	fmt.Println("(8) - Teresina")
	fmt.Println("(9) - Sao Luis")
	fmt.Println("(0) - Retornar ao Menu")
}

//ok
func opcoesMenu() {
	fmt.Println("\n======= Menu Inicial =======")
	fmt.Println("(1) - Programar viagem")
	fmt.Println("(0) - Sair")
}

func MenuInicial() {
	input := bufio.NewReader(os.Stdin)
	on := true
	placa = IdentificacaoInicialPlaca()
	var veiculo Veiculo
	if placa == "" {
		fmt.Println("Falha na identificacao do veiculo")
		return
	}
	veiculo.Placa = placa
	setDadosIniciais(&veiculo)
	erro := EscreveArquivoVeiculos(veiculo)
	if erro != nil {
		fmt.Printf("Erro ao escrever no arquivo: %v\n", erro)
		return
	}
	defer RemovePlacaVeiculo(veiculo.Placa)
	fmt.Printf("Placa[%s] registrada\n", veiculo.Placa)

	canalSinal := make(chan os.Signal, 1)
	signal.Notify(canalSinal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-canalSinal
		RemovePlacaVeiculo(veiculo.Placa)
		os.Exit(0)
	}()

	for on {
		opcoesMenu()
		fmt.Println("Selecione uma opcao: ")
		opcao, _ := input.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		switch opcao {
		case "1":
			ProgramarReservas(&veiculo)
		case "0":
			fmt.Println("Sistema encerrado! Até a próxima")
			RemovePlacaVeiculo(veiculo.Placa)
			on = false
		default:
			fmt.Println("Opcao invalida, tente novamente!")
		}
	}
}

//ok
func GetCidade(tipo string) string {
	input := bufio.NewReader(os.Stdin)
	on := true
	for on {
		listCapitaisNordeste()
		fmt.Printf("Selecione a cidade de %s: \n", tipo)
		opcao, _ := input.ReadString('\n')
		opcao = strings.TrimSpace(opcao)
		switch opcao {
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			return opcao
		case "0":
			on = false
			return "-2"
		default:
			fmt.Println("Opcao invalida. Tente novamente!")
			continue
		}
	}
	return "-1"
}

//ok
func IdentificacaoInicialPlaca() string {
	input := bufio.NewReader(os.Stdin)
	placa := ""

	placa_validada := false
	for !placa_validada {
		veiculos_json, erro := GetVeiculosAtivosJson()
		if erro != nil {
			fmt.Printf("Erro ao localizar veiculos no arquivo json: %v", erro)
		}
		fmt.Println("Ola! Para iniciar informe a placa do seu veiculo: ")
		input, _ := input.ReadString('\n')
		placa = strings.TrimSpace(input)

		if len(placa) < 6 || len(placa) > 8 {
			fmt.Println("Placa invalida! A placa deve ter entre 6 e 8 caracteres.")
			continue
		}

		placa_ativa := false
		for _, veiculo := range veiculos_json {
			if strings.EqualFold(veiculo.Placa, placa) {
				fmt.Println("A placa informada ja esta em uso. Tente novamente!")
				placa_ativa = true
				break
			}
		}
		if placa_ativa {
			continue
		}
		placa_validada = true
	}
	return placa
}

//ok
func RecargasNecessarias(veiculo *Veiculo, cidades_da_viagem []string) []Ponto {
	var recargas_necessarias []Ponto
	var cap_km_bateria_restante, dist_ate_prox_ponto float64
	pontos_da_viagem := GetPontosPorCidades(cidades_da_viagem)
	max := len(pontos_da_viagem)
	fmt.Printf("\nSimulação da viagem para o veículo [%s]:\n", veiculo.Placa)
	for i, ponto := range pontos_da_viagem {
		ponto_atual := ponto
		if i < max-1 {
			proximo_ponto := pontos_da_viagem[i+1]

			dist_ate_prox_ponto = GetDistancia(ponto_atual.Latitude, ponto_atual.Longitude, proximo_ponto.Latitude, proximo_ponto.Longitude)
			cap_km_bateria_restante = (veiculo.Autonomia * veiculo.NivelBateriaAtual) / 100

			if i == 0 {
				fmt.Printf("Saindo de %s - Bateria restante: %.2f%% Capacidade (%.2f km). Próximo ponto em: %.2f km\n", ponto.Cidade, veiculo.NivelBateriaAtual, cap_km_bateria_restante, dist_ate_prox_ponto)
			} else {
				fmt.Printf("Passando em %s - Bateria restante: %.2f%% Capacidade (%.2f km). Próximo ponto em: %.2f km\n", ponto.Cidade, veiculo.NivelBateriaAtual, cap_km_bateria_restante, dist_ate_prox_ponto)
			}

			if dist_ate_prox_ponto > cap_km_bateria_restante {
				recargas_necessarias = append(recargas_necessarias, ponto_atual)
				veiculo.NivelBateriaAtual = 100
				cap_km_bateria_restante = (veiculo.Autonomia * veiculo.NivelBateriaAtual) / 100
				fmt.Printf("Recarregará em %s! Bateria restante: %.2f%% Capacidade(%.2fkm). Próximo ponto em: %.2fkm\n", ponto.Cidade, veiculo.NivelBateriaAtual, cap_km_bateria_restante, dist_ate_prox_ponto)
			}
			percentualConsumido := (dist_ate_prox_ponto / veiculo.Autonomia) * 100
			veiculo.NivelBateriaAtual -= percentualConsumido
		} else {
			fmt.Printf("Chegará ao seu destino em %s! Bateria restante: %.2f%% Capacidade(%.2f km).\n", ponto.Cidade, veiculo.NivelBateriaAtual, (veiculo.Autonomia*veiculo.NivelBateriaAtual)/100)
		}
	}
	return recargas_necessarias
}

//ok
func GetDistanciaRota(origem, destino int) float64 {
	var pontosViagem []Ponto
	pontos, erro := GetPontosDeRecargaJson()
	if erro != nil {
		fmt.Printf("Erro ao carregar pontos: %v", erro)
		return -1
	}
	var latitudeOrigem, longitudeOrigem, latitudeDestino, longitudeDestino float64

	if origem <= destino {
		pontosViagem = pontos[origem : destino+1]
	} else {
		for i := origem; i >= destino; i-- {
			pontosViagem = append(pontosViagem, pontos[i])
		}
	}

	for i, ponto := range pontosViagem {
		max := len(pontosViagem) - 1
		if i == 0 {
			latitudeOrigem = ponto.Latitude
			longitudeOrigem = ponto.Longitude
		} else if i == max {
			latitudeDestino = ponto.Latitude
			longitudeDestino = ponto.Longitude
		}
	}
	distanciaTotal := GetDistancia(latitudeOrigem, longitudeOrigem, latitudeDestino, longitudeDestino)
	return distanciaTotal
}

//ok
func ProgramarReservas(veiculo *Veiculo) {
	origem := GetCidade("Origem")
	if origem == "-2" {
		return
	}
	destino := GetCidade("Destino")
	rota_nordeste, erro := GetRotaJson()
	if erro != nil {
		fmt.Printf("Erro ao carregar rota: %v", erro)
		return
	}
	rota_viagem, indexOrigem, indexDestino := GetTrechoRotaCompleta(origem, destino, rota_nordeste)
	distancia := GetDistanciaRota(indexOrigem, indexDestino)

	fmt.Printf("\nRota da viagem: \n")
	for i, cidade := range rota_viagem {
		max := len(rota_viagem)
		if i == 0 {
			fmt.Printf("Origem: %s -> ", cidade)
		} else if i == max-1 {
			fmt.Printf("%s: Destino Final.\n", cidade)
		} else {
			fmt.Print(cidade, " -> ")
		}
	}
	fmt.Printf("Distancia total: %.2fkm\n\n", distancia)

	pontos_necessarios := RecargasNecessarias(veiculo, rota_viagem)

	if len(pontos_necessarios) == 0 {
		fmt.Printf("\nPara este trajeto não será necessário reservar pontos!\n")
		return
	}

	fmt.Printf("\nPontos necessários para recarregar: \n")
	for i, ponto := range pontos_necessarios {
		fmt.Printf("[%d°] Parada: Ponto %s\n", i+1, ponto.Cidade)
	}

	//pontos para pre-reserva
	var listPontosFinal []string
	for _, ponto := range pontos_necessarios {
		listPontosFinal = append(listPontosFinal, ponto.Cidade)
	}
	pontosString := strings.Join(listPontosFinal, ",")

	//faz a pré-reserva automaticamente
	fmt.Printf("\nRealizando pré-reserva dos pontos...\n")
	msg := "4," + placa + "," + pontosString
	preReservaSucesso := conecta(msg, placa)

	//confirma a reserva
	if preReservaSucesso {
		leitor := bufio.NewReader(os.Stdin)
		fmt.Print("\nGostaria de confirmar a reserva? (S/N): ")
		opcao, _ := leitor.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		if strings.ToLower(opcao) == "s" || strings.ToLower(opcao) == "sim" {
			// 5 - confirma
			msg = "5," + placa + "," + pontosString
			fmt.Println("\nConfirmando reserva dos pontos de recarga...")
			conecta(msg, placa)

			//simula viagem e recargas
			fmt.Println("\nSimulando viagem...")
			for i, ponto := range listPontosFinal {
				fmt.Printf("\nParada %d/%d\n", i+1, len(listPontosFinal))
				fmt.Printf("Veiculo viajando para %s...\n", ponto)
				time.Sleep(2 * time.Second)
				fmt.Printf("Chegou em %s. Carregando...\n", ponto)
				time.Sleep(2 * time.Second)
				fmt.Printf("Recarga concluida em %s.\n", ponto)
			}
			fmt.Println("\nViagem concluida com sucesso. Agradecemos por utilizar nossos serviços!")
			time.Sleep(15 * time.Second)
			// Enviar mensagem para liberar pontos
			liberarPontosMQTT(placa, listPontosFinal)
			fmt.Println("Todos os pontos foram liberados.")
		} else {
			// 6 -> cancela
			msg = "6," + placa + "," + pontosString
			fmt.Println("\n A pré-reserva esta sendo cancelada...")
			conecta(msg, placa)
		}
	} else {
		fmt.Printf("\n[FALHA] Não foi possível prosseguir com a pré-reserva. Tente novamente mais tarde!\n")
	}
}

//ok
// Libera os pontos ao concluir viagem
func liberarPontosMQTT(placa string, pontos []string) {
	mensagem := "7," + placa + "," + strings.Join(pontos, ",")
	conecta(mensagem, placa)
}
