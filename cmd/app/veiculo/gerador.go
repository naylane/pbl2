package main

import (
	"fmt"
	"math/rand"
)

func setDadosVeiculo(veiculo *Veiculo) {
	level := float64(int(10 + rand.Float64()*(91)))
	autonomia := float64(int(500 + rand.Float64()*(201)))

	fmt.Printf("Bateria atual do veiculo placa[%s]: %.0f%%\n", veiculo.Placa, level)
	fmt.Printf("Autonomia do veiculo placa[%s]: %.0fkm\n", veiculo.Placa, autonomia)

	veiculo.NivelBateriaAtual = level
	veiculo.Autonomia = autonomia
}