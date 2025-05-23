package main

import (
	"fmt"
	"os"
)

func main() {
	// Carrega dados das empresas do arquivo JSON
	AbreArquivoEmpresas()
	// Carrega dados dos pontos de recarga
	GetPontosDeRecargaJson()

	// Obtém o identificador da empresa das variáveis de ambiente
	idEmpresa := os.Getenv("ID")
	if idEmpresa == "" {
		fmt.Println("Erro: ID não definido")
		return
	}

	// Configura porta do servidor com base no ID da empresa
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

	// Carrega dados específicos da empresa
	empresa = GetEmpresaPorId(idEmpresa)
	fmt.Printf("[Iniciando servidor] %s - %d pontos de recarga\n", empresa.Nome, len(empresa.Pontos))

	// Inicializa componentes do sistema
	inicializa_rest(porta)
	inicializaMonitoramentoDosPontos()
	//comunicação com clientes
	inicializaMqtt(idEmpresa)
	//Mantem o servidor em execução
	select {}
}
