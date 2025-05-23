package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Estrutura que representa um ponto de recarga no sistema
type Ponto struct {
	ID        int     `json:"id"`
	Cidade    string  `json:"cidade"`
	Estado    string  `json:"estado"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Reservado string  `json:"reservado"`
}

// Dados completos da região com pontos e rota principal
type DadosRegiao struct {
	PontosDeRecarga     []Ponto  `json:"pontos_de_recarga"`
	RotaSalvadorSaoLuis []string `json:"rota_salvador_saoLuis"`
}

// Registra uma operação de recarga realizada
type Recarga struct {
	Data    string  `json:"data"`
	PontoID int     `json:"ponto_id"`
	Valor   float64 `json:"valor"`
}

// Representa um veículo no sistema
type Veiculo struct {
	Placa             string    `json:"placa"`
	Autonomia         float64   `json:"autonomia"`
	NivelBateriaAtual float64   `json:"batery_level"`
	Recargas          []Recarga `json:"recargas,omitempty"`
}

// Contém todos os veículos ativos no sistema
type DadosVeiculos struct {
	Veiculos []Veiculo `json:"veiculos"`
}

// ok
// Carrega os dados dos veículos do arquivo JSON
func AbreArquivoVeiculos() (DadosVeiculos, error) {
	file, erro := os.Open("/app/veiculos.json")
	if erro != nil {
		return DadosVeiculos{}, fmt.Errorf("erro ao abrir: %v", erro)
	}
	defer file.Close()

	var dadosVeiculos DadosVeiculos
	erro = json.NewDecoder(file).Decode(&dadosVeiculos)
	if erro != nil {
		return DadosVeiculos{}, fmt.Errorf("erro ao ler: %v", erro)
	}
	return dadosVeiculos, nil
}

// ok
// Adiciona um novo veículo ao arquivo JSON
func EscreveArquivoVeiculos(veiculo Veiculo) error {
	dadosVeiculos, erro := AbreArquivoVeiculos()
	if erro != nil {
		fmt.Printf("Erro ao abrir arquivo: %v\n", erro)
		return erro
	}
	dadosVeiculos.Veiculos = append(dadosVeiculos.Veiculos, veiculo)

	file, erro := os.OpenFile("/app/veiculos.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if erro != nil {
		fmt.Printf("Erro ao criar arquivo: %v\n", erro)
		return erro
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	erro = encoder.Encode(dadosVeiculos)
	return erro
}

// ok
// Localiza um veículo pela sua placa
// Retorna o veículo e um código (0=sucesso, 1=erro, 2=não encontrado)
func GetVeiculoPorPlaca(placa string) (Veiculo, int) {
	dadosVeiculos, erro := AbreArquivoVeiculos()
	if erro != nil {
		return Veiculo{}, 1
	}

	for _, veiculo := range dadosVeiculos.Veiculos {
		if veiculo.Placa == placa {
			return veiculo, 0
		}
	}
	return Veiculo{}, 2
}

// ok
// Remove um veículo do arquivo de veículos ativos
func RemovePlacaVeiculo(placa string) error {
	dadosVeiculos, erro := AbreArquivoVeiculos()
	if erro != nil {
		return fmt.Errorf("erro ao abrir arquivo: %v", erro)
	}

	var listaAtualizada []Veiculo
	for _, v := range dadosVeiculos.Veiculos {
		if !strings.EqualFold(v.Placa, placa) {
			listaAtualizada = append(listaAtualizada, v)
		}
	}
	dadosVeiculos.Veiculos = listaAtualizada

	file, erro := os.OpenFile("/app/veiculos.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if erro != nil {
		return fmt.Errorf("erro ao salvar arquivo: %v", erro)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(dadosVeiculos)
}

// ok
// Obtém a lista de todos os veículos ativos
func GetVeiculosAtivosJson() ([]Veiculo, error) {
	DadosVeiculos, erro := AbreArquivoVeiculos()
	if erro != nil {
		return DadosVeiculos.Veiculos, fmt.Errorf("erro ao carregar dados JSON: %v", erro)
	}
	return DadosVeiculos.Veiculos, nil
}

// ok
// Carrega os dados da região do arquivo JSON
func AbreArquivoRegiao() (DadosRegiao, error) {
	file, erro := os.Open("/app/regiao.json")
	if erro != nil {
		return DadosRegiao{}, fmt.Errorf("erro ao abrir: %v", erro)
	}
	defer file.Close()

	var dadosRegiao DadosRegiao
	erro = json.NewDecoder(file).Decode(&dadosRegiao)
	if erro != nil {
		return DadosRegiao{}, fmt.Errorf("erro ao ler: %v", erro)
	}
	return dadosRegiao, nil
}

// ok
// Obtém a rota principal Salvador-São Luís
func GetRotaJson() ([]string, error) {
	dadosRegiao, erro := AbreArquivoRegiao()
	if erro != nil {
		return dadosRegiao.RotaSalvadorSaoLuis, fmt.Errorf("erro ao carregar dados JSON: %v", erro)
	}
	return dadosRegiao.RotaSalvadorSaoLuis, nil
}

// ok
// Obtém a lista de todos os pontos de recarga
func GetPontosDeRecargaJson() ([]Ponto, error) {
	dadosRegiao, erro := AbreArquivoRegiao()
	if erro != nil {
		return dadosRegiao.PontosDeRecarga, fmt.Errorf("erro ao carregar dados JSON: %v", erro)
	}
	return dadosRegiao.PontosDeRecarga, nil
}

// ok
// Localiza pontos de recarga por cidade
// Retorna pontos correspondentes às cidades da lista
func GetPontosPorCidades(cidades []string) []Ponto {
	var pontos []Ponto
	pontosJson, erro := GetPontosDeRecargaJson()
	if erro != nil {
		return []Ponto{}
	}
	for _, cidade := range cidades {
		for _, ponto := range pontosJson {
			if strings.EqualFold(cidade, ponto.Cidade) {
				pontos = append(pontos, ponto)
			}
		}
	}
	return pontos
}

// ok
// Calcula trecho de rota entre origem e destino
// Retorna a lista de cidades do trecho e índices da origem/destino
func GetTrechoRotaCompleta(origem string, destino string, rotaCompleta []string) ([]string, int, int) {
	var trechoViagem []string

	indexOrigem, err1 := strconv.Atoi(origem)
	indexDestino, err2 := strconv.Atoi(destino)

	if err1 != nil || err2 != nil || 1 > indexOrigem || 9 < indexOrigem || 1 > indexDestino || 9 < indexDestino {
		return []string{}, -1, -1
	}

	if indexOrigem-1 <= indexDestino-1 {
		trechoViagem = rotaCompleta[indexOrigem-1 : indexDestino]
	} else {
		for i := indexOrigem - 1; i >= indexDestino-1; i-- {
			trechoViagem = append(trechoViagem, rotaCompleta[i])
		}
	}
	return trechoViagem, indexOrigem - 1, indexDestino - 1
}

// ok
// Localiza um ponto de recarga pelo ID
func GetPontoId(id int) (Ponto, int) {
	dadosRegiao, erro := AbreArquivoRegiao()
	if erro != nil {
		return Ponto{}, 1
	}

	for _, ponto := range dadosRegiao.PontosDeRecarga {
		if ponto.ID == id {
			return ponto, 0
		}
	}
	return Ponto{}, 2
}

// ok
// Retorna o número total de pontos de recarga cadastrados
func GetTotalPontosJson() int {
	pontos, erro := GetPontosDeRecargaJson()
	if erro != nil {
		return -1
	}
	return len(pontos)
}
