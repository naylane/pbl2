package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Ponto struct {
	ID        int     `json:"id"`
	Cidade    string  `json:"cidade"`
	Estado    string  `json:"estado"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type DadosRegiao struct {
	PontosDeRecarga     []Ponto  `json:"pontos_de_recarga"`
	RotaSalvadorSaoLuis []string `json:"rota_salvador_saoLuis"`
}

type Recarga struct {
	Data    string  `json:"data"`
	PontoID int     `json:"ponto_id"`
	Valor   float64 `json:"valor"`
}

type Veiculo struct {
	Placa    string    `json:"placa"`
	Recargas []Recarga `json:"recargas,omitempty"`
}

type DadosVeiculos struct {
	Veiculos []Veiculo `json:"veiculos"`
}

func GetTrechoRotaCompleta(origem string, destino string, rotaCompleta []string) []string {
	var trechoViagem []string
	indexOrigem, indexDestino := -1, -1

	for i, cidade := range rotaCompleta {
		if strings.ToUpper(cidade) == strings.ToUpper(origem) {
			indexOrigem = i
		}
		if strings.ToUpper(cidade) == strings.ToUpper(destino) {
			indexDestino = i
		}
	}

	if indexOrigem == -1 || indexDestino == -1 {
		return []string{} //Cidade nao encontrada
	}

	if indexOrigem <= indexDestino {
		trechoViagem = rotaCompleta[indexOrigem : indexDestino+1]
	} else {
		for i := indexOrigem; i >= indexDestino; i-- {
			trechoViagem = append(trechoViagem, rotaCompleta[i])
		}
	}
	return trechoViagem
}

func OpenFile(arquivo string) (DadosRegiao, error) {
	path := filepath.Join("app", "internal", "data", arquivo)
	file, erro := os.Open(path)
	if erro != nil {
		return DadosRegiao{}, (fmt.Errorf("Erro ao abrir: %v", erro))
	}
	defer file.Close()

	var dadosRegiao DadosRegiao
	erro = json.NewDecoder(file).Decode(&dadosRegiao)
	if erro != nil {
		return DadosRegiao{}, (fmt.Errorf("Erro ao ler: %v", erro))
	}
	return dadosRegiao, nil
}

func GetRotaSalvadorSaoLuis() ([]string, error) {
	dadosRegiao, erro := OpenFile("regiao.json")
	if erro != nil {
		return dadosRegiao.RotaSalvadorSaoLuis, fmt.Errorf("Erro ao carregar dados JSON: %v", erro)
	}
	return dadosRegiao.RotaSalvadorSaoLuis, nil
}

func GetPontosDeRecargaJson() ([]Ponto, error) {
	dadosRegiao, erro := OpenFile("regiao.json")
	if erro != nil {
		return dadosRegiao.PontosDeRecarga, fmt.Errorf("Erro ao carregar dados JSON: %v", erro)
	}
	return dadosRegiao.PontosDeRecarga, nil
}

func GetPontoId(id int) (Ponto, int) {
	dadosRegiao, erro := OpenFile("regiao.json")
	if erro != nil {
		return Ponto{}, 1 //Erro ao carregar dados JSON
	}

	for _, ponto := range dadosRegiao.PontosDeRecarga {
		if ponto.ID == id {
			return ponto, 0
		}
	}
	return Ponto{}, 2 //Erro ao localizar ponto
}

func GetTotalPontosJson() int {
	pontos, erro := GetPontosDeRecargaJson()
	if erro != nil {
		return -1
	}
	return len(pontos)
}

func main() {
	rotaNordeste, erro := GetRotaSalvadorSaoLuis()

	if erro != nil {
		fmt.Printf("Erro ao carregar dados JSON: %v", erro)
	}

	origem := "salvador"
	destino := "natal"
	trecho := GetTrechoRotaCompleta(origem, destino, rotaNordeste)

	fmt.Printf("Trecho da viagem de %s ate %s:\n", origem, destino)
	for _, cidade := range trecho {
		fmt.Println(" -", cidade)
	}
}
