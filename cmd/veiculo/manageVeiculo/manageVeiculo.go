package manageveiculo

import (
	"bufio"
	"fmt"
	"os"
	"pbl2/internal/data"
	"pbl2/internal/distancia"
	"strings"
)

func IdentificacaoInicialPlaca() string {
	leitor := bufio.NewReader(os.Stdin)
	placa := ""
	placaValida := false

	for !placaValida {
		fmt.Println("Por favor, informe a placa do veiculo (6-8 caracteres): ")
		input, _ := leitor.ReadString('\n')
		placa = strings.TrimSpace(input)

		if len(placa) < 6 || len(placa) > 8 { // Validar formato da placa
			fmt.Println("Placa invalida! A placa deve ter entre 6 e 8 caracteres.")
			continue
		}
		placaValida = true
		// Enviar solicitação ao servidor para verificar se a placa está disponível
		// Esperar resposta do servidor
		/*
			if resposta.Tipo == "placa-disponivel" {
				placaValida = true
			} else if resposta.Tipo == "placa-indisponivel" {
				fmt.Println("Esta placa já está em uso por outro veículo!")
			} else {
				logger.Erro(fmt.Sprintf("Resposta inesperada do servidor: %s", resposta.Tipo))
				return ""
			}*/
	}
	// Agora que sabemos que a placa é válida, enviar a identificação final
	return placa
}

func ConsultarPagamentosPendentes() {
}

func CancelarReserva() {
}

func listCapitaisNordeste() {
	fmt.Println("\n==== Cidades com Servico de Recarga ====")
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

func GetCidade(tipo string) string {
	leitor := bufio.NewReader(os.Stdin)
	on := true
	for on {
		listCapitaisNordeste()
		fmt.Printf("Selecione a cidade de %s: \n", tipo)
		opcao, _ := leitor.ReadString('\n')
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

func GetDistanciaRota(origem, destino int) float64 {
	var pontosViagem []data.Ponto
	pontos, erro := data.GetPontosDeRecargaJson()
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
			fmt.Printf("origem: %s", ponto.Cidade)
			latitudeOrigem = ponto.Latitude
			longitudeOrigem = ponto.Longitude
		} else if i == max {
			fmt.Printf("destino: %s", ponto.Cidade)
			latitudeDestino = ponto.Latitude
			longitudeDestino = ponto.Longitude
		}
	}
	distanciaTotal := distancia.GetDistancia(latitudeOrigem, longitudeOrigem, latitudeDestino, longitudeDestino)
	return distanciaTotal / 1000
}

func ProgramarViagem() {
	origem := GetCidade("Origem")
	if origem == "-2" {
		return
	}
	destino := GetCidade("Destino")
	rotaNordeste, erro := data.GetRotaSalvadorSaoLuis()
	if erro != nil {
		fmt.Printf("Erro ao carregar rota: %v", erro)
		return
	}
	rota, indexOrigem, indexDestino := data.GetTrechoRotaCompleta(origem, destino, rotaNordeste)
	distancia := GetDistanciaRota(indexOrigem, indexDestino)
	fmt.Printf("Rota da viagem: \n")
	for i, cidade := range rota {
		max := len(rota)
		if i == 0 {
			fmt.Printf("Origem: %s -> ", cidade)
		} else if i == max-1 {
			fmt.Printf("%s: Destino Final.\n", cidade)
		} else {
			fmt.Print(cidade, " -> ")
		}
	}
	fmt.Printf("Distancia total: %.2f km\n", distancia)
}

func listMenu() {
	fmt.Println("\n==== Menu Veiculo ====")
	fmt.Println("(1) - Programar viagem")
	fmt.Println("(2) - Cancelar reserva")
	fmt.Println("(3) - Consultar pagamentos pendentes")
	fmt.Println("(0) - Sair")
}

func MenuVeiculo() {
	leitor := bufio.NewReader(os.Stdin)
	on := true
	placa := IdentificacaoInicialPlaca()

	// Verificar se a identificação falhou
	if placa == "" {
		fmt.Println("Falha na identificacao do veiculo")
		return
	}

	fmt.Printf("Veiculo com placa %s registrado com sucesso!\n", placa)

	for on {
		listMenu()
		fmt.Println("Selecione uma opcao: ")
		opcao, _ := leitor.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		switch opcao {
		case "1":
			ProgramarViagem()
		case "2":
			CancelarReserva()
		case "3":
			ConsultarPagamentosPendentes()
		case "0":
			fmt.Println("Saindo...")
			on = false
		default:
			fmt.Println("Opcao invalida. Tente novamente!")
		}
	}
}
