package main

import (
	"fmt"
	"os"
)

func main() {
	AbreArquivoEmpresas()
	GetPontosDeRecargaJson()
	// Le variável de ambiente do docker-compose
	idEmpresa := os.Getenv("ID")
	if idEmpresa == "" {
		fmt.Println("Erro: ID não definido")
		return
	}
	//Inicia o servidor REST
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
	empresa = GetEmpresaPorId(idEmpresa)
	fmt.Printf("[Iniciando servidor] %s - %d pontos de recarga\n", empresa.Nome, len(empresa.Pontos))
	//comunicação servidor-servidor
	inicializa_rest(porta)
	inicializaMonitoramentoDosPontos()
	//comunicação com clientes
	inicializaMqtt(idEmpresa)
	//Mantem o servidor em execução
	select {}
}
