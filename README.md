<h2 align="center">Planejamento de viagens para recargas de veículos elétricos</h2>
<h4 align="center">Projeto da disciplina TEC502 - Concorrência e Conectividade.</h4>

<p align="center">Este projeto foi desenvolvido para facilitar a comunicação entre veículos elétricos e pontos de recarga. Utilizando arquiteturas MQTT e API's REST, o sistema permite que veículos programem viagens reservando pontos necessários para recarregar, informem a cidade de origem e a cidade de destino, o estado atual da bateria e sua autonomia e recebam recomendações de pontos de recarga para atender ao seu percurso.</p>
<p align="center">O projeto consiste em um sistema distribuído, para simular / gerenciar veículos e empresas na região nordeste, utilizando comunicação via MQTT e API REST, tudo orquestrado com Docker. Ele é composto por múltiplos serviços (servidores e veículos), que comunicam entre si e com um broker MQTT. O objetivo é otimizar o processo de recarga, garantindo eficiência e gerenciamento adequado da concorrência.</p>

[Relatório](https://docs.google.com/document/d/1NYiV0I9dxnWGn_qsMTTqNb5xW55k6mtO/edit?pli=1)

## Sumário
- [Sumário](#sumário)
- [Introdução](#introdução)
- [Arquitetura do Sistema](#arquitetura-do-sistema)
  - [Broker MQTT](#broker-mqtt)
  - [Servidor](#servidor)
  - [API REST](#api-rest)
  - [Veículo](#veículo)
  - [Fluxo de Comunicação](#fluxo-de-comunicação)
  - [Funcionalidades Principais](#funcionalidades-principais)
- [Protocolo de Comunicação](#protocolo-de-comunicação)
  - [Dados e Estado](#dados-e-estado)
- [Conexões Simultâneas](#conexões-simultâneas)
- [Gerenciamento de Concorrência](#gerenciamento-de-concorrência)
  - [Garantia de Reserva e Integridade](#garantia-de-reserva-e-integridade)
- [Execução com Docker](#execução-com-docker)
- [Como Executar](#como-executar)
  - [Pré-requisitos](#pré-requisitos)
  - [Passo a passo](#passo-a-passo)
- [Tecnologias Utilizadas](#tecnologias-utilizadas)
- [Conclusão](#conclusão)
- [Desenvolvedoras](#desenvolvedoras)
- [Referências](#referências)

## Introdução

O sistema simula um ambiente distribuído de recarga de veículos elétricos, com múltiplos servidores (empresas) e veículos, utilizando comunicação via MQTT e API REST. O objetivo é permitir que veículos planejem viagens de longa distância, reservando pontos de recarga de diferentes empresas de forma otimizada, com controle de concorrência e integridade dos dados.

A aplicação é composta por:
- **Broker**: serviço de mensageria MQTT (Eclipse Mosquitto), permitindo a troca de mensagens entre servidores e veículos.
- **Servidores**: cada servidor representa uma empresa, expõe uma API REST e se comunica com outros serviços para gerenciar reservas, pré-reservas e cancelamentos.
- **Veículo**: simula um cliente que planeja viagens, solicita reservas e interage com o sistema via terminal.

Todos os serviços são orquestrados com Docker Compose, garantindo isolamento, escalabilidade e fácil simulação de concorrência distribuída.

## Arquitetura do Sistema

A arquitetura utiliza MQTT para comunicação assíncrona e API REST para coordenação entre servidores. Os dados são persistidos em arquivos JSON montados como volumes nos containers.

### Broker MQTT
- Utiliza a imagem oficial do Eclipse Mosquitto.
- Viabiliza a comunicação entre todos os serviços.
- Exposto na porta 1883.

### Servidor
- Desenvolvido em Go.
- Expõe uma API REST.
- Gerencia dados de empresas, regiões e veículos em arquivos JSON.
- Recebe solicitações de veículos via MQTT, coordena reservas locais e remotas.
- Utiliza goroutines e mutexes para concorrência segura.

### API REST
- Usada para coordenação de reservas/pré-reservas/cancelamentos entre servidores.
- Endpoints principais: `/api/confirmar-prereserva`, `/api/reserva`, `/api/cancelamento`.
- Recebe e responde requisições em JSON.

### Veículo
- Implementado em Go, com interface via terminal.
- Permite ao usuário:
  - Informar origem, destino, bateria e autonomia.
  - Receber rota sugerida com pontos de recarga necessários.
  - Solicitar pré-reserva, confirmar reserva.
- Comunica-se via MQTT, publicando solicitações e recebendo respostas em tópicos exclusivos.

### Fluxo de Comunicação
1. **Veículo** publica solicitação (ex: pré-reserva) via MQTT.
2. **Servidor** recebe, processa e responde via MQTT.
3. Se necessário, servidor coordena com outros servidores via REST.
4. Resposta final é enviada ao veículo.

### Funcionalidades Principais
- Programação de viagem com sugestão de pontos de recarga.
- Pré-reserva e confirmação de pontos.
- Cancelamento e liberação automática por timeout.
- Concorrência segura e atomicidade nas operações distribuídas.

## Protocolo de Comunicação
- Mensagens estruturadas em JSON.
- MQTT para comunicação assíncrona entre veículos e servidores.
- REST para coordenação entre servidores.

### Dados e Estado
- Dados de empresas, regiões e veículos em arquivos JSON.
- Carregados em memória ao iniciar o servidor.
- Atualizados e persistidos conforme operações.

## Conexões Simultâneas
- Servidores suportam múltiplas conexões simultâneas usando goroutines.
- Concorrência controlada com mutexes para evitar condições de corrida.

## Gerenciamento de Concorrência
- Uso de mutexes (locks) para garantir exclusão mútua em operações críticas.
- Cada ponto de recarga possui um lock próprio.
- Exemplo:
  - Antes de reservar um ponto, o servidor executa `lock.Lock()`.
  - Após a operação, libera com `lock.Unlock()`.
- Garante que dois veículos não reservem o mesmo ponto simultaneamente.

### Garantia de Reserva e Integridade
- Operações de reserva são atômicas: ou todos os pontos são reservados, ou nenhum.
- Se algum ponto falhar, as reservas temporárias são canceladas.
- Timeout automático libera pontos não utilizados.

## Execução com Docker
- O sistema é simulado com Docker Compose.
- Cada serviço (broker, servidores, veículos) roda em um container isolado.
- Volumes mapeiam arquivos de dados para persistência.

## Como Executar
### Pré-requisitos
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://go.dev/) (opcional, para testes locais)

### Passo a passo
1. Clone o repositório:
   ```bash
   git clone https://github.com/usuario/nome-do-repositorio.git
   cd nome-do-repositorio
   ```
2. Compile as imagens Docker e inicie o sistema:
   ```bash
   docker compose up --build -d
   ```
3. Acesse o terminal do veículo ou servidor:
   ```bash
   docker compose exec veiculo sh
   # ou
   docker compose exec servidor1 sh
   ```
4. Execute a aplicação dentro do container:
   ```bash
   ./veiculo
   # ou
   ./servidor
   ```
5. Para encerrar:
   ```bash
   docker compose down
   ```
6. Para ver logs:
   ```bash
   docker compose logs -f servidor1
   # ou
   docker compose logs -f veiculo
   ```

## Tecnologias Utilizadas
- Go (Golang)
- MQTT (Eclipse Mosquitto)
- REST (API HTTP)
- Docker e Docker Compose
- JSON para persistência de dados

## Conclusão
O sistema demonstra na prática conceitos de concorrência distribuída, comunicação em tempo real e integração de múltiplos serviços. O uso de MQTT e REST permite flexibilidade e robustez na troca de mensagens, enquanto Docker garante portabilidade e fácil simulação. O controle de concorrência com mutexes assegura integridade nas operações de reserva, mesmo com múltiplos veículos e servidores atuando simultaneamente.

## Desenvolvedoras
<table>
  <tr>
    <td align="center"><img style="" src="https://avatars.githubusercontent.com/u/142849685?v=4" width="100px;" alt=""/><br /><sub><b> Brenda Araújo </b></sub></a><br />👨‍💻</a></td>
    <td align="center"><img style="" src="https://avatars.githubusercontent.com/u/89545660?v=4" width="100px;" alt=""/><br /><sub><b> Naylane Ribeiro </b></sub></a><br />👨‍💻</a></td>
    <td align="center"><img style="" src="https://avatars.githubusercontent.com/u/124190885?v=4" width="100px;" alt=""/><br /><sub><b> Letícia Gonçalves </b></sub></a><br />👨‍💻</a></td>    
  </tr>
</table>

## Referências
Donovan, A. A. and Kernighan, B. W. (2016). The Go Programming Language. Addison-Wesley.   
Merkel, D. (2014). Docker: lightweight Linux containers for consistent development and deployment. Linux Journal, 2014(239), 2.    
Silberschatz, A., Galvin, P. B., and Gagne, G. (2018). Operating System Concepts (10th ed.). Wiley.   
Stevens, W. R. (1998). UNIX Network Programming, Volume 1: The Sockets Networking API (2nd ed.). Prentice Hall.    
Tanenbaum, A. S. and Van Steen, M. (2007). Distributed Systems: Principles and Paradigms (2nd ed.). Pearson Prentice Hall.
