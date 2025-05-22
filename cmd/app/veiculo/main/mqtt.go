package main

import (
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// envio e recepção com conexão única
func conecta(mensagem string, idClient string) bool {
	opts := mqtt.NewClientOptions().AddBroker("tcp://172.16.103.3:1883")
	opts.SetClientID(idClient)

	var respostaRecebida bool
	var operacaoSucesso bool

	//Configurar handler antes de conectar para mensagens de resposta
	topicResponse := "mensagens/cliente/" + idClient
	opts.OnConnect = func(c mqtt.Client) {
		fmt.Printf("\nConectado ao broker: Cliente %s.\n", idClient)

		//Subscreve ao tópico de resposta ANTES de publicar
		if token := c.Subscribe(topicResponse, 0, func(client mqtt.Client, msg mqtt.Message) {
			mensagemRecebida := string(msg.Payload())
			fmt.Printf("\n[Resposta]: %s\n", mensagemRecebida)
			parts := strings.Split(mensagemRecebida, ",")
			if len(parts) < 1 {
				return
			}
			switch parts[0] {
			case "reserva_confirmada":
				fmt.Println("\n[sucesso] Reserva efetuada!")
				respostaRecebida = true
				operacaoSucesso = true
			case "reserva_falhou":
				fmt.Println("\n[falha] Não foi possível reservar todos os pontos.")
				respostaRecebida = true
				operacaoSucesso = false
			case "ponto_desconectado":
				if len(parts) >= 2 {
					fmt.Printf("\n[info] Ponto %s desconectado.\n", parts[1])
				} else {
					fmt.Println("\n[info] Algum ponto está desconectado.")
				}
				fmt.Println("Escolha outra rota ou aguarde a reconexão.")
				respostaRecebida = true
				operacaoSucesso = false
			case "falha_reserva":
				if len(parts) >= 3 {
					fmt.Printf("\n[Falha] reserva: %s\n", parts[2])
				} else {
					fmt.Println("\n[Falha] Não foi possível reservar todos os pontos.")
				}
				respostaRecebida = true
				operacaoSucesso = false
			case "cancelamento_confirmado":
				fmt.Println("\n[info] Cancelamento de reserva realizado.")
				respostaRecebida = true
				operacaoSucesso = true
			case "cancelamento_falhou":
				if len(parts) >= 2 {
					fmt.Printf("\n [Falha] ao cancelar reserva: %s\n", parts[1])
				} else {
					fmt.Println("\n [Falha] ao cancelar a reserva.")
				}
				respostaRecebida = true
				operacaoSucesso = false
			case "prereserva_confirmada":
				fmt.Println("\n [sucesso] Pré-reserva realizada! Confirme em até 15 minutos.")
				respostaRecebida = true
				operacaoSucesso = true
			case "prereserva_cancelada":
				fmt.Println("\n [info] Pré-reserva cancelada.")
				respostaRecebida = true
				operacaoSucesso = true
			case "falha_prereserva":
				if len(parts) >= 3 {
					fmt.Printf("\n [Falha] pré-reserva: %s\n", parts[2])
				} else {
					fmt.Println("\n [Falha] na pré-reserva dos pontos.")
				}
				respostaRecebida = true
				operacaoSucesso = false
			case "pontos_liberados":
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
	//aguarda resposta
	fmt.Println("Aguardando resposta do servidor...")
	timeout := time.After(10 * time.Second)
	ticker := time.Tick(500 * time.Millisecond)

	for !respostaRecebida {
		select {
		case <-timeout:
			fmt.Printf("\nTempo esgotado. Aguardando resposta do servidor...\n")
			return false
		case <-ticker:
			// Continua aguardando
		}
	}
	// Desconecta após receber resposta ou timeout
	client.Disconnect(250)
	return operacaoSucesso
}
