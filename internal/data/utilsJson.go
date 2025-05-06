package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	return trechoViagem, indexOrigem-1, indexDestino-1
}

func OpenFile(arquivo string) (DadosRegiao, error) {
	path := filepath.Join("internal", "data", arquivo) //"app", "internal", "data", arquivo
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