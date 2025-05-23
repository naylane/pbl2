package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Representa uma empresa com seus pontos de operação
type Empresa struct {
	Id     string
	Nome   string
	Pontos []string
}

// Representa um ponto de recarga com suas coordenadas e status
type Ponto struct {
	ID        int     `json:"id"`
	Cidade    string  `json:"cidade"`
	Estado    string  `json:"estado"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Reservado string  `json:"reservado"`
}

// Estrutura que contém pontos de recarga e informações de rota
type DadosRegiao struct {
	PontosDeRecarga     []Ponto  `json:"pontos_de_recarga"`
	RotaSalvadorSaoLuis []string `json:"rota_salvador_saoLuis"`
}

// Registro de uma operação de recarga realizada
type Recarga struct {
	Data    string  `json:"data"`
	PontoID int     `json:"ponto_id"`
	Valor   float64 `json:"valor"`
}

// Lista de empresas cadastradas no sistema
type DadosEmpresas struct {
	Empresas []Empresa `json:"empresas"`
}

// Variáveis globais que armazenam os dados carregados dos arquivos
var dadosEmpresas DadosEmpresas
var dadosRegiao DadosRegiao

// ok
// Carrega os dados da região a partir do arquivo JSON
func AbreArquivoRegiao() (DadosRegiao, error) {
	file, erro := os.Open("/app/regiao.json")
	if erro != nil {
		return DadosRegiao{}, fmt.Errorf("erro ao abrir: %v", erro)
	}
	defer file.Close()

	erro = json.NewDecoder(file).Decode(&dadosRegiao)
	if erro != nil {
		return DadosRegiao{}, fmt.Errorf("erro ao ler: %v", erro)
	}
	return dadosRegiao, nil
}

// ok
// Obtém os pontos de recarga disponíveis na região
func GetPontosDeRecargaJson() ([]Ponto, error) {
	dadosRegiao, erro := AbreArquivoRegiao()
	if erro != nil {
		return dadosRegiao.PontosDeRecarga, fmt.Errorf("erro ao carregar dados JSON: %v", erro)
	}
	return dadosRegiao.PontosDeRecarga, nil
}

// ok
// Persiste alterações nos dados de pontos de recarga
func salvaDadosPontos() {
	bytes, err := json.MarshalIndent(dadosRegiao, "", "  ")
	if err != nil {
		fmt.Println("Erro ao converter dados para JSON:", err)
		return
	}

	err = os.WriteFile("regiao.json", bytes, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar no arquivo regiao.json:", err)
		return
	}

	fmt.Println("\nDados salvos no arquivo Região!")
}

// ok
// Carrega os dados das empresas a partir do arquivo JSON
func AbreArquivoEmpresas() {
	bytes, err := os.ReadFile("empresas.json")
	if err != nil {
		fmt.Println("Erro ao abrir arquivo JSON:", err)
		return
	}

	err = json.Unmarshal(bytes, &dadosEmpresas)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}
}

// ok
// Localiza uma empresa pelo seu ID
// 001 = N-Sul, 002 = N-Centro, 003 = N-Norte
func GetEmpresaPorId(id string) Empresa {
	var empresa Empresa
	if len(dadosEmpresas.Empresas) > 0 {
		for _, emp := range dadosEmpresas.Empresas {
			if emp.Id == id {
				empresa = emp
			}
		}
	}
	return empresa
}

// ok
// Localiza um ponto de recarga pelo nome da cidade
// Retorna o ponto e seu índice no array
func GetPontoPorCidade(cidade string) (Ponto, int) {
	var ponto Ponto
	var index int
	pontos := dadosRegiao.PontosDeRecarga
	if len(pontos) > 0 {
		for i, p := range pontos {
			if p.Cidade == cidade {
				ponto = p
				index = i
			}
		}
	}
	return ponto, index
}
