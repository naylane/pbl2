package main

import (
	"fmt"
	"os"
)

// Função principal: inicializa o servidor e seus componentes
func main() {
	AbreArquivoEmpresas()
	GetPontosDeRecargaJson()
	// Le variável de ambiente do docker-compose
	idEmpresa := os.Getenv("ID")
	if idEmpresa == "" {
		fmt.Println("Erro: ID não definido")
		return
	}
	// Define porta com base no ID da empresa
	var porta string
	switch idEmpresa {
	case "001":
		porta = "8081"
	case "002":
		porta = "8082"
	case "003":
		porta = "8083"
	default:
		porta = "8080"
	}
	// Recupera detalhes da empresa pelo ID
	empresa = GetEmpresaPorId(idEmpresa)
	fmt.Printf("[Iniciando servidor] %s - %d pontos de recarga\n", empresa.Nome, len(empresa.Pontos))
	// Inicializa comunicação REST entre servidores
	inicializa_rest(porta)
	inicializaMonitoramentoDosPontos()
	// Inicializa comunicação MQTT com clientes
	inicializaMqtt(idEmpresa)
	//Mantem o servidor em execução
	select {}
}
