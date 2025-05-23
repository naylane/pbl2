package main

import (
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Faz envio e recepção com conexão única
func conecta(mensagem string, id_veiculo string) bool {
	opts := mqtt.NewClientOptions().AddBroker("tcp://broker:1883")
	opts.SetClientID(id_veiculo)

	var respostaRecebida bool
	var operacaoSucesso bool

	topicResponse := "mensagens/cliente/" + id_veiculo
	opts.OnConnect = func(c mqtt.Client) {
		fmt.Printf("\nConectado ao broker: Cliente %s.\n", id_veiculo)

		if token := c.Subscribe(topicResponse, 0, func(client mqtt.Client, msg mqtt.Message) {
			mensagemRecebida := string(msg.Payload())
			fmt.Printf("\n[Resposta]: %s\n", mensagemRecebida)
			parts := strings.Split(mensagemRecebida, ",")
			if len(parts) < 1 {
				return
			}
			// Processa diferentes tipos de respostas do servidor
			switch parts[0] {
			case "reserva_confirmada":
				// Processamento de confirmação de reserva
				fmt.Println("\n[sucesso] Reserva efetuada!")
				respostaRecebida = true
				operacaoSucesso = true
			case "reserva_falhou":
				// Processamento de falha na reserva
				fmt.Println("\n[falha] Não foi possível reservar todos os pontos.")
				respostaRecebida = true
				operacaoSucesso = false
			case "ponto_desconectado":
				// Processamento de notificação de ponto desconectado
				if len(parts) >= 2 {
					fmt.Printf("\n[info] Ponto %s desconectado.\n", parts[1])
				} else {
					fmt.Println("\n[info] Algum ponto está desconectado.")
				}
				fmt.Println("Escolha outra rota ou aguarde a reconexão.")
				respostaRecebida = true
				operacaoSucesso = false
			case "falha_reserva":
				// Processamento de erro específico na reserva
				if len(parts) >= 3 {
					fmt.Printf("\n[Erro] reserva: %s\n", parts[2])
				} else {
					fmt.Println("\n[Erro] Não foi possível reservar todos os pontos.")
				}
				respostaRecebida = true
				operacaoSucesso = false
			case "cancelamento_confirmado":
				// Processamento de confirmação de cancelamento
				fmt.Println("\n[info] Cancelamento de reserva realizado.")
				respostaRecebida = true
				operacaoSucesso = true
			case "cancelamento_falhou":
				// Processamento de falha no cancelamento
				if len(parts) >= 2 {
					fmt.Printf("\n [Erro] ao cancelar reserva: %s\n", parts[1])
				} else {
					fmt.Println("\n [Erro] ao cancelar a reserva.")
				}
				respostaRecebida = true
				operacaoSucesso = false
			case "prereserva_confirmada":
				// Processamento de confirmação de pré-reserva
				fmt.Println("\n [sucesso] Pré-reserva realizada! Confirme em até 15 minutos.")
				respostaRecebida = true
				operacaoSucesso = true
			case "prereserva_cancelada":
				// Processamento de cancelamento de pré-reserva
				fmt.Println("\n [info] Pré-reserva cancelada.")
				respostaRecebida = true
				operacaoSucesso = true
			case "erro_prereserva":
				if len(parts) >= 3 {
					fmt.Printf("\n [Falha] pré-reserva: %s\n", parts[2])
				} else {
					fmt.Println("\n [Falha] na pré-reserva dos pontos.")
				}
				respostaRecebida = true
				operacaoSucesso = false
			case "pontos_liberados":
				// Processamento de confirmação de liberação de pontos
				if len(parts) >= 2 {
					fmt.Printf("\n%s\n", parts[1])
				} else {
					fmt.Println("\n [sucesso] Pontos liberados!")
				}
				respostaRecebida = true
				operacaoSucesso = true
			default:
				fmt.Printf("\n [Mensagem recebida]: %s\n", mensagemRecebida)
			}
		}); token.Wait() && token.Error() != nil {
			fmt.Println(" [Erro] assinatura do tópico:", token.Error())
		}
	}

	// Estabelece conexão com o broker MQTT
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	//garante que a conexão foi estabelecida
	time.Sleep(1 * time.Second)

	// Publica a mensagem no canal de solicitações
	fmt.Printf("\n[Enviando solicitação]: %s\n", mensagem)
	token := client.Publish("mensagens/cliente", 0, false, mensagem)
	token.Wait()

	// Implementação de timeout para aguardar resposta do servidor
	fmt.Println("Aguardando resposta do servidor...")
	timeout := time.After(10 * time.Second)
	ticker := time.Tick(500 * time.Millisecond)

	// Loop que aguarda a resposta ou timeout
	for !respostaRecebida {
		select {
		case <-timeout:
			fmt.Printf("\nTempo esgotado. Aguardando resposta do servidor...\n")
			return false
		case <-ticker:
			// Continua aguardando
		}
	}
	// Encerra a conexão MQTT após receber resposta
	client.Disconnect(250)
	return operacaoSucesso
}
