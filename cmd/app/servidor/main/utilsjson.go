package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Empresa struct {
	Id     string
	Nome   string
	Pontos []string
}

type Ponto struct {
	ID        int     `json:"id"`
	Cidade    string  `json:"cidade"`
	Estado    string  `json:"estado"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Reservado string  `json:"reservado"`
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
	Placa             string    `json:"placa"`
	Autonomia         float64   `json:"autonomia"`
	NivelBateriaAtual float64   `json:"batery_level"`
	Recargas          []Recarga `json:"recargas,omitempty"`
}

type DadosVeiculos struct {
	Veiculos []Veiculo `json:"veiculos"`
}
type DadosEmpresas struct {
	Empresas []Empresa `json:"empresas"`
}

var dadosEmpresas DadosEmpresas
var dadosRegiao DadosRegiao

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

func GetVeiculosAtivosJson() ([]Veiculo, error) {
	DadosVeiculos, erro := AbreArquivoVeiculos()
	if erro != nil {
		return DadosVeiculos.Veiculos, fmt.Errorf("erro ao carregar dados JSON: %v", erro)
	}
	return DadosVeiculos.Veiculos, nil
}

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

func GetRotaJson() ([]string, error) {
	dadosRegiao, erro := AbreArquivoRegiao()
	if erro != nil {
		return dadosRegiao.RotaSalvadorSaoLuis, fmt.Errorf("erro ao carregar dados JSON: %v", erro)
	}
	return dadosRegiao.RotaSalvadorSaoLuis, nil
}

func GetPontosDeRecargaJson() ([]Ponto, error) {
	dadosRegiao, erro := AbreArquivoRegiao()
	if erro != nil {
		return dadosRegiao.PontosDeRecarga, fmt.Errorf("erro ao carregar dados JSON: %v", erro)
	}
	return dadosRegiao.PontosDeRecarga, nil
}

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

func GetTotalPontosJson() int {
	pontos, erro := GetPontosDeRecargaJson()
	if erro != nil {
		return -1
	}
	return len(pontos)
}

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

func leArquivoJsonEmpresas() {
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

// 001 = N-Sul, 002 = N-Centro, 003 = N-Norte
func getEmpresaPorId(id string) Empresa {
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

func getPontoPorCidade(cidade string) (Ponto, int) {
	var ponto Ponto
	var index int
	pontos := dadosRegiao.PontosDeRecarga
	if len(pontos) > 0 {
		for i, pont := range pontos {
			if pont.Cidade == cidade {
				ponto = pont
				index = i
			}
		}
	}
	return ponto, index
}