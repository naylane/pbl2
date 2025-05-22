package main

import (
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var empresa Empresa
var mqttClient mqtt.Client

//ok
func publicaMensagemMqtt(client mqtt.Client, topico string, mensagem string) {
	token := client.Publish(topico, 0, false, mensagem)
	token.Wait()
	fmt.Printf("\nMensagem enviada para %s: %s\n", topico, mensagem)
}

//ok
func getClienteMqtt() mqtt.Client {
	return mqttClient
}

//OK
// Inicializa MQTT para comunicação com cliente.
func inicializaMqtt(idCliente string) {
	empresa = GetEmpresaPorId(idCliente)

	//O servidor se conecta via TCP ao broker - este pc é o broker
	opts := mqtt.NewClientOptions().AddBroker("tcp://broker:1883")
	opts.SetClientID(idCliente)

	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("[Servidor MQTT] inicializado - Empresa " + idCliente + " e conectado ao broker")

		//O handle de mensagens será chamado sempre que uma nova mensagem for publicada nesse tópico
		if token := c.Subscribe("mensagens/cliente", 0, handleMensagens); token.Wait() && token.Error() != nil {
			fmt.Println("Erro ao assinar tópico:", token.Error())
		}
	}

	//se conecta ao broker (Mosquitto) e age como conexão para o servidor
	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

// Verifica se um ponto pertence a empresa ou não. Recebe como parâmetro o ponto e retorna true ou false.
func pertenceAEstaEmpresa(ponto string) bool {
	for _, p := range empresa.Pontos {
		if p == ponto {
			return true
		}
	}
	return false
}

//ok
// Handler para tratar mensagens recebidas pelo cliente. Avalia o tipo de mensagem de acordo com o seu código.
func handleMensagens(client mqtt.Client, msg mqtt.Message) {
	list := strings.Split(string(msg.Payload()), ",")
	fmt.Printf("[Servidor recebeu]: %s\n", msg.Payload())

	if len(list) < 2 {
		fmt.Println("\nMensagem com formatação inválida!")
		return
	}

	codigo := list[0]
	placaVeiculo := list[1]
	pontos := list[2:]

	switch codigo {
	case "1": // Reserva
		fmt.Printf("\n Solicitação de reserva recebida - placa [%s]\n", placaVeiculo)
		if pertenceAEstaEmpresa(pontos[0]) {
			fmt.Printf("Processando reserva - ponto %s pertence a esta empresa.\n", pontos[0])
			reservaMqtt(client, pontos, placaVeiculo)
		} else {
			fmt.Printf("Aguardando confirmação via REST - placa [%s]\n", placaVeiculo)
			// Os outros servidores receberão pela API REST
		}
	case "3": // Cancelamento
		fmt.Printf("Solicitação de cancelamento recebida - placa [%s]\n", placaVeiculo)
		cancelaMqtt(client, placaVeiculo)

	case "4": // Pré-reserva
		fmt.Printf("Solicitação de pré-reserva recebida - placa [%s] nos pontos: %v\n", placaVeiculo, pontos)
		if len(pontos) > 0 && pertenceAEstaEmpresa(pontos[0]) {
			//primeiro ponto pertence a esta empresa - processa a reserva
			preReservaMqtt(client, pontos, placaVeiculo)
		} else if len(pontos) > 0 {
			//coordenação de reservas consultando outros servidores via api rest
			handlePreReservaRest(placaVeiculo, pontos)
		}

	case "5": // Confirmar pré-reserva
		fmt.Printf("Confirmação de pré-reserva recebida - placa [%s] nos pontos: %v\n", placaVeiculo, pontos)
		if len(pontos) > 0 && pertenceAEstaEmpresa(pontos[0]) {
			confirmaPreReservaMqtt(client, pontos, placaVeiculo)
		} else if len(pontos) > 0 {
			// Coordena com outros servidores e responde ao cliente
			sucesso := handleConfirmacaoPreReservaRest(placaVeiculo, pontos)
			if sucesso {
				publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
					"reserva_confirmada,Reserva confirmada em todos os servidores")
				fmt.Printf("\nReserva confirmada em todos os pontos solicitados - placa [%s]\n", placaVeiculo)
			} else {
				publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
					"reserva_falhou,Não foi possível confirmar a reserva em todos os pontos")
				fmt.Printf("\nReserva não confirmada - placa [%s]\n", placaVeiculo)
			}
		}

	case "6": // Cancelar pré-reserva
		fmt.Printf("Solicitação de cancelamento de pré-reserva recebida - placa [%s] nos pontos: %v\n", placaVeiculo, pontos)
		if len(pontos) > 0 && pertenceAEstaEmpresa(pontos[0]) {
			cancelaPreReservaMqtt(client, pontos, placaVeiculo)
		} else if len(pontos) > 0 {
			handleCancelaPreReservaRest(placaVeiculo, pontos)
		}

	case "7":
		//fmt.Printf("Solicitação de liberação de pontos recebida - placa [%s] nos pontos: %v\n", placaVeiculo, pontos)
		liberaPontosConcluiuViagem(client, placaVeiculo, pontos)
	default:
		fmt.Printf("[AVISO] Código desconhecido recebido: '%s'. Conteúdo: %s\n", codigo, string(msg.Payload()))
	}
}

//ok
// Faz a pré-reserva do(s) ponto(s).
func preReservaMqtt(client mqtt.Client, pontosParaReservar []string, placaVeiculo string) {
	pontos_locais := false
	falhaLocal := false
	var pontosReservadosTemp []string
	var indexesReservados []int

	for _, ponto := range pontosParaReservar {
		lock := ponto_locks[ponto]
		lock.Lock()

		for _, pontoDaEmpresa := range empresa.Pontos {
			if ponto == pontoDaEmpresa {
				status_ponto.RLock()
				conectado := status_ponto.status[ponto]
				status_ponto.RUnlock()

				if !conectado {
					publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
						fmt.Sprintf("ponto_desconectado,%s,Ponto %s está desconectado", ponto, ponto))
					falhaLocal = true
					lock.Unlock()
					return
				}

				pontos_locais = true
				ponto_recarga, i := GetPontoPorCidade(ponto)

				if ponto_recarga.Reservado == "" {
					pontosReservadosTemp = append(pontosReservadosTemp, ponto)
					indexesReservados = append(indexesReservados, i)
				} else if ponto_recarga.Reservado == "PRE_"+placaVeiculo || ponto_recarga.Reservado == placaVeiculo {
					pontosReservadosTemp = append(pontosReservadosTemp, ponto)
					indexesReservados = append(indexesReservados, i)
				} else {
					if strings.HasPrefix(ponto_recarga.Reservado, "PRE_") {
						outra_placa := ponto_recarga.Reservado[4:]
						publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
							fmt.Sprintf("falha_prereserva,%s,Ponto %s já está pré-reservado pelo veículo [%s]",
								ponto, ponto, outra_placa))
					} else {
						publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
							fmt.Sprintf("falha_prereserva,%s,Ponto %s já está reservado pelo veículo [%s]",
								ponto, ponto, ponto_recarga.Reservado))
					}
					falhaLocal = true
					lock.Unlock()
					return
				}
			}
		}

		lock.Unlock()
	}

	if falhaLocal {
		return
	}

	if pontos_locais {
		for i, ponto := range pontosReservadosTemp {
			lock := ponto_locks[ponto]
			lock.Lock()

			index := indexesReservados[i]
			dadosRegiao.PontosDeRecarga[index].Reservado = "PRE_" + placaVeiculo
			fmt.Printf("[INFO] Ponto %s da empresa %s pré-reservado para %s.\n",
				ponto, empresa.Nome, placaVeiculo)

			lock.Unlock()
		}
		salvaDadosPontos()
		publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
			"prereserva_confirmada,Os pontos foram pré-reservados com sucesso")
		liberaPreReservaTimeout(placaVeiculo, pontosReservadosTemp, 15*time.Minute)
	}
}

//ok
func confirmaPreReservaMqtt(client mqtt.Client, pontosParaReservar []string, placaVeiculo string) {
	pontos_locais := false
	sucesso := true

	for _, ponto := range pontosParaReservar {
		lock := ponto_locks[ponto]
		lock.Lock()

		for _, pontoDaEmpresa := range empresa.Pontos {
			if ponto == pontoDaEmpresa {
				pontos_locais = true
				pontoRecarga, index := GetPontoPorCidade(ponto)

				if pontoRecarga.Reservado == "PRE_"+placaVeiculo {
					dadosRegiao.PontosDeRecarga[index].Reservado = placaVeiculo
					fmt.Printf("[SUCESSO] Ponto %s pré-reserva convertida para reserva para %s.\n",
						ponto, placaVeiculo)

				} else if pontoRecarga.Reservado == placaVeiculo {
					fmt.Printf("[INFO] Ponto %s já está reservado para %s.\n",
						ponto, placaVeiculo)
				} else {
					if pontoRecarga.Reservado == "" {
						fmt.Printf("[ERRO] Ponto %s não está pré-reservado (está vazio).\n", ponto)
					} else if strings.HasPrefix(pontoRecarga.Reservado, "PRE_") {
						outra_placa := pontoRecarga.Reservado[4:]
						fmt.Printf("[ERRO] Ponto %s está pré-reservado para outro veículo: %s.\n",
							ponto, outra_placa)
					} else {
						fmt.Printf("[ERRO] Ponto %s está reservado para outro veículo: %s.\n",
							ponto, pontoRecarga.Reservado)
					}
					sucesso = false
				}
			}
		}
		lock.Unlock()
	}

	if pontos_locais {
		if sucesso {
			salvaDadosPontos()
			reservas_mutex.Lock()
			if _, existe := reservas[placaVeiculo]; !existe {
				reservas[placaVeiculo] = make(map[string]string)
			}
			for _, ponto := range pontosParaReservar {
				reservas[placaVeiculo][ponto] = "confirmado"
			}
			reservas_mutex.Unlock()

			publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
				"reserva_confirmada,Reserva confirmada com sucesso")
			liberaPorTimeout(placaVeiculo, pontosParaReservar, 3*time.Hour)
		} else {
			publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
				"reserva_falhou,Falha ao confirmar pré-reserva")
		}
	} else {
		publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
			"reserva_confirmada,Confirmação processada")
	}
}

//ok
func cancelaPreReservaMqtt(client mqtt.Client, pontosParaReservar []string, placaVeiculo string) {
	pontos_locais := false
	cancelou := false

	for _, ponto := range pontosParaReservar {
		lock := ponto_locks[ponto]
		lock.Lock()

		for _, pontoDaEmpresa := range empresa.Pontos {
			if ponto == pontoDaEmpresa {
				pontos_locais = true
				pontoRecarga, index := GetPontoPorCidade(ponto)

				if pontoRecarga.Reservado == "PRE_"+placaVeiculo {
					dadosRegiao.PontosDeRecarga[index].Reservado = ""
					cancelou = true
					fmt.Printf("[INFO] Ponto %s pré-reserva cancelada para %s.\n", ponto, placaVeiculo)
				}
			}
		}

		lock.Unlock()
	}

	if pontos_locais && cancelou {
		salvaDadosPontos()
		publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
			"prereserva_cancelada,Pré-reserva cancelada com sucesso")
	} else if pontos_locais {
		publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
			"prereserva_cancelada,Nenhuma pré-reserva encontrada para cancelar")
	}
}

//ok
func liberaPreReservaTimeout(placa string, pontos []string, tempo time.Duration) {
	go func() {
		time.Sleep(tempo)
		fmt.Printf("Verificando timeout para pré-reservas do veículo %s...\n", placa)

		for _, ponto := range pontos {
			lock := ponto_locks[ponto]
			lock.Lock()

			pontoRecarga, i := GetPontoPorCidade(ponto)
			if pontoRecarga.Reservado == "PRE_"+placa {
				dadosRegiao.PontosDeRecarga[i].Reservado = ""
				fmt.Printf("[INFO] Pré-reserva para %s no ponto %s expirou e foi liberada automaticamente.\n", placa, ponto)
			}

			lock.Unlock()
		}
		salvaDadosPontos()
	}()
}

//ok
func reservaMqtt(client mqtt.Client, pontos_a_reservar []string, placaVeiculo string) {
	//pontos locais primeiro
	pontos_locais := false
	falha_local := false
	var pontos_reservados_temp []string //reservas temporárias
	var i_reservados []int              //índices reservados
	for _, ponto := range pontos_a_reservar {
		lock := ponto_locks[ponto]
		lock.Lock()

		for _, ponto := range pontos_a_reservar {
			for _, pontoDaEmpresa := range empresa.Pontos {
				if ponto == pontoDaEmpresa {
					status_ponto.RLock()
					estaConectado := status_ponto.status[ponto]
					status_ponto.RUnlock()

					if !estaConectado {
						fmt.Printf("[ERRO] Tentativa de reserva no ponto %s falhou: ponto desconectado.\n", ponto)
						publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
							fmt.Sprintf("ponto_desconectado,%s,Ponto %s desconectado", ponto, ponto))
						falha_local = true
						lock.Unlock()
						return
					}

					pontos_locais = true
					pontoRecarga, index := GetPontoPorCidade(ponto)

					if pontoRecarga.Reservado == "" || pontoRecarga.Reservado == placaVeiculo {
						//pontos que serão reservados
						pontos_reservados_temp = append(pontos_reservados_temp, ponto)
						i_reservados = append(i_reservados, index)
					} else {
						// Falha na reserva local - ponto já reservado por outro veículo
						fmt.Printf("[ERRO] O ponto %s da empresa %s está reservado no momento.\n",
							ponto, empresa.Nome)
						publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
							fmt.Sprintf("falha_reserva,%s,Ponto %s reservado no momento", ponto, ponto))
						falha_local = true
						lock.Unlock()
						return
					}
				}
			}
		}

		if falha_local {
			lock.Unlock()
			return
		}

		// coordena reserva com outras empresas via REST
		if pontos_locais {
			//reserva temporariamente os pontos
			for i, ponto := range pontos_reservados_temp {
				index := i_reservados[i]
				dadosRegiao.PontosDeRecarga[index].Reservado = placaVeiculo
				fmt.Printf("[INFO] Ponto %s da empresa %s reservado temporariamente para [%s].\n",
					ponto, empresa.Nome, placaVeiculo)
			}

			// outros servidores via REST
			sucesso := handleReservaRest(placaVeiculo, pontos_a_reservar)

			if sucesso {
				// Confirma reservas locais
				salvaDadosPontos()
				publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
					"reserva_confirmada,Todos os pontos foram reservados")
				fmt.Printf("[SUCESSO] Reserva confirmada para a placa [%s] em todos os pontos solicitados.\n", placaVeiculo)
				//libera a reserva automaticamente após 3 horas
				liberaPorTimeout(placaVeiculo, pontos_a_reservar, 3*time.Hour)
			} else {
				// Cancela todas as reservas temporárias locais
				for i, ponto := range pontos_reservados_temp {
					index := i_reservados[i]
					dadosRegiao.PontosDeRecarga[index].Reservado = ""
					fmt.Printf("[INFO] Reserva temporária cancelada no ponto %s.\n", ponto)
				}
				salvaDadosPontos()
				// Notificar o cliente sobre a falha
				publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
					"reserva_falhou,Não foi possível reservar todos os pontos solicitados")
				fmt.Printf("[ERRO] Reserva para a placa [%s].\n", placaVeiculo)
			}
		} else {
			//não tem pontos locais, tenta reservar em outros servidores via REST
			sucesso := reservaPontosEmOutrosServidores(placaVeiculo, pontos_a_reservar)
			if sucesso {
				publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
					"reserva_confirmada,Todos os pontos foram reservados")
				fmt.Printf("[SUCESSO] Reserva confirmada para a placa [%s] em servidores externos.\n", placaVeiculo)
			} else {
				publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
					"reserva_falhou,Não foi possível reservar todos os pontos solicitados")
				fmt.Printf("[ERRO] Reserva para a placa [%s].\n", placaVeiculo)
			}
		}
		lock.Unlock()
	}
}

//ok
// Cancela reservas vinculadas a placa do veículo.
func cancelaMqtt(client mqtt.Client, placaVeiculo string) {
	reservas_mutex.Lock()
	defer reservas_mutex.Unlock()

	cancelou := false
	//existe reservas para essa placa
	if pontosMap, existe := reservas[placaVeiculo]; existe {
		for ponto := range pontosMap {
			lock := ponto_locks[ponto]
			lock.Lock()
			ponto_desta_empresa := false
			for _, pontoDaEmpresa := range empresa.Pontos {
				if ponto == pontoDaEmpresa {
					ponto_desta_empresa = true
					break
				}
			}

			if ponto_desta_empresa {
				// Cancelar a reserva local
				pontoObj, index := GetPontoPorCidade(ponto)
				if pontoObj.Reservado == placaVeiculo {
					dadosRegiao.PontosDeRecarga[index].Reservado = ""
					delete(pontosMap, ponto)
					salvaDadosPontos()
					cancelou = true
					fmt.Printf("[INFO] Cancelamento de reserva para %s no ponto %s.\n", placaVeiculo, ponto)
				}
			}
			lock.Unlock()
		}
		if len(pontosMap) == 0 {
			delete(reservas, placaVeiculo)
		}
	}

	if cancelou {
		publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo, "cancelamento_confirmado")
	} else {
		publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo,
			"cancelamento_falhou,Nenhuma reserva encontrada neste servidor")
	}

}

//ok
// Libera os pontos reservados após o cliente concluir a viagem.
func liberaPontosConcluiuViagem(client mqtt.Client, placaVeiculo string, pontos []string) {
	reservas_mutex.Lock()
	defer reservas_mutex.Unlock()
	liberou := false

	for _, ponto := range pontos {
		lock := ponto_locks[ponto]
		lock.Lock()
		pontoObj, index := GetPontoPorCidade(ponto)
		if pontoObj.Reservado == placaVeiculo {
			dadosRegiao.PontosDeRecarga[index].Reservado = ""
			liberou = true
			fmt.Printf("[INFO] Ponto %s liberado para a placa %s.\n", ponto, placaVeiculo)
		}
		lock.Unlock()
	}

	if liberou {
		salvaDadosPontos()
		delete(reservas, placaVeiculo)
		publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo, "pontos_liberados,Pontos liberados")
	} else {
		publicaMensagemMqtt(client, "mensagens/cliente/"+placaVeiculo, "pontos_liberados,Nenhum ponto estava reservado para esta placa")
	}
}
