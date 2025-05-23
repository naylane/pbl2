<h2 align="center">Planejamento de viagens com suporte à recargas de veículos elétricos</h2>
<h4 align="center">Projeto da disciplina TEC502 - Concorrência e Conectividade.</h4>

<p align="center">Este projeto foi desenvolvido para facilitar a comunicação entre veículos elétricos e pontos de recarga. Utilizando arquiteturas MQTT, API's REST e Clientes, o sistema permite que veículos programem viagens reservando pontos necessário para recarregar, informem a cidade de origem e a cidade de destino, estado atual da bateria e sua autonomia e recebam recomendações para pontos de recarga para atender ao seu percurso.</p>
<p align="center">O projeto consiste em um sistema distribuído, para simular / gerenciar uma frota de veículos e empresas na região nordeste, utilizando comunicação via MQTT e API REST, tudo orquestrado com Docker. Ele é composto por múltiplos serviços (servidores e veículos), que se comunicam entre si e com um broker MQTT. Cujo objetivo é otimizar o processo de recarga, garantindo eficiência e gerenciamento adequado da concorrência. </p>


[Relatório](https://docs.google.com/document/d/1NYiV0I9dxnWGn_qsMTTqNb5xW55k6mtO/edit?pli=1)
## Sumário

- [Sumário](#sumário)
- [Introdução](#introdução)
- [Arquitetura do Sistema](#arquitetura-do-sistema)
  - [Servidor](#servidor)
  - [MQTT](#mqtt)
  - [API REST](#api-rest)
  - [Veículo](#veículo)
  - [Comunicação](#comunicação)
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

O presente sistema foi desenvolvido para implementar gerenciamento de concorrência distribuída entre veículos e empresas na região nordeste utilizando comunicação via MQTT e REST, simulando requisições atômicas para diferentes empresas no contexto de recarga de veículos elétricos. O projeto viabiliza a solicitação e gestão de recargas por parte dos veículos, utilizando MQTT e API REST e desenvolvimento em Go, com suporte para múltiplas conexões simultâneas. Simulando um ambiente realista onde múltiplos servidores e veículos trocam informações em tempo real, como em sistemas de transporte.

A aplicação está contida em containers Docker, que isolam e orquestram a execução dos serviços. Onde:
- broker: serviço de mensaegns MQTT, usando a imagem do Eclipse Mosquitto. Permite que os outros serviços troquem mensagens de forma assíncrona.
- servidores: expõe uma API REST na sua respectiva porta e se comunica com o broker MQTT. Recebe variáveis de ambiente para identificar servidores e portas.
- veiculo: Simula um veículo elétrico, que também se comunica com o broker MQTT.

Porcionando então, uma solução que permite aos veículos planejar viagens de longa distância com múltiplas recargas, reservar e utilizar pontos de recarga de diferentes empresas de forma otimizada com uma única solicitação.

## Arquitetura do Sistema

A solução foi desenvolvida utilizando a arquitetura de comunicação MQTT e API REST, onde a comunicação entre as partes ocorre ... Seu uso garante a ... proporcionando uma comunicação confiável entre os módulos do sistema: servidores e veículos. 

A troca de dados ocorre via ..., com mensagens estruturadas em formato JSON. O sistema foi projetado para funcionar em ambiente de containers Docker interconectados por uma rede interna definida no docker-compose, garantindo isolamento, escalabilidade e simulação de concorrência distribuída. Onde:

- **Servidores**: Gerencia as solicitações, consulta os pontos, calcula distâncias e gerenciar as solicitações de reservas.
- **Veículo**: Responsável por programar viagens, informar cidade de origem e destino e confirmar reservas.

### Servidor
O servidor atua como o ... do sistema, responsável por intermediar a comunicação entre veículos e outros servidores, escutando conexões ... em uma porta definida. As principais responsabilidades do servidor incluem:
- Gerenciar conexões ... de veículos e outros servidores.
- Gerenciar solicitações de recarga dos veículos, ...
- Gerenciar as reservas, ...  
O servidor foi desenvolvido em Go, utilizando recursos como goroutines para o tratamento concorrente de conexões. Isso garante maior performance e segurança no acesso aos dados compartilhados.

### MQTT

### API REST

### Veículo
O veículo é implementado como cliente ... onde o usuário interage por meio de um menu via terminal que permite:
- Enviar cidade de origem e destino, bateria atual e autonomia ao programar viagem.
- Receber rota de viagem, com fila de espera e distância.
- Escolher um ponto de recarga para reservar e efetuar recarga  
O sistema é capaz de manter sessões interativas com o servidor, permitindo que o usuário envie solicitações de recarga e consulte seu histórico de recargas pendentes para efetuar o pagamento posteriormente.  

A comunicação entre as partes ocorre via **sockets TCP/IP** conforme ilustração da arquitetura à seguir:

<div align="center">  
  <img align="center" width=100% src= public/sistema-recarga.png alt="Comunicação sistema">
  <p><em>Arquitetura do Sistema</em></p>
</div>

### Comunicação

- Veículo programa viagem enviando seus dados.
-
- 

### Funcionalidades Principais

- **Programação de Viagem**: O veículo pode programar uma viagem.
-
-

## Protocolo de Comunicação
A comunicação entre os clientes e o servidor é realizada por meio de ... utilizando mensagens estruturadas em JSON. A escolha do formato JSON foi decorrente da necessidade de garantia de entrega confiável e legível, além do formato ser leve, compatível com diversos ambientes e amplamente adotado em sistemas distribuídos. Cada mensagem permite a troca de dados e encapsulam ações como identificação dos clientes, solicitação de recarga, envio de disponibilidade, confirmação de reservas, entre outros.

### Dados e Estado
Os dados do sistema como região de cobertura e localização dos pontos de recarga cadastrados, são carregados a partir de arquivos JSON ao iniciar o servidor e permanecem em memória, funcionando como um cache de alta performance para as operações. Isso reduz a latência e permite respostas rápidas às requisições.  

## Conexões Simultâneas
O servidor foi projetado para suportar múltiplas conexões simultâneas utilizando goroutines, nativas da linguagem Go. ...

## Gerenciamento de Concorrência
Para garantir a integridade dos dados durante operações concorrentes como por exemplo a atualizações das disponibilidades dos pontos de recarga, registro de reservas, modificação em estruturas de dados, entre outras. Foi implementado o uso de mutexes - locks de exclusão mútua. 

O controle de exclusão mútua assegura que múltiplas goroutines não modifiquem simultaneamente estruturas de dados compartilhadas, como a disponibilidade de um ponto de recarga.  

Funcionamento:  
- Lock: Antes da operação crítica, a goroutine realiza um mutex.Lock().  
- Seção Crítica: A operação crítica é executada de forma exclusica onde os dados são validados e atualizados de forma segura.
- Unlock: Após a operação, o mutex é liberado com mutex.Unlock(), permitindo que outras goroutines acessem os dados.  

Essa abordagem impede condições de corrida, evitando problemas como múltiplos veículos tentando ocupar a mesma posição na reserva de um determinado ponto de recarga simultaneamente.

### Garantia de Reserva e Integridade
Ao solicitar uma recarga, o veículo envia sua bateria e autonomia atual, cidade de origem e cidade de destino ao servidor. O servidor, então:

- 
-
-

Após a confirmar a reserva, o veículo é adicionado à reserva do ponto selecionado. Para garantir a integridade da operação, cada etapa é realizada com controle de concorrência utilizando mutexes, impedindo que dois veículos reservem a mesma posição simultaneamente.

## Execução com Docker
A simulação do sistema é feita utilizando Docker-Compose, com containers para os Servidores e os Veículos. O Docker Compose permite as partes do sistema compartilhar uma rede interna privada, proporcionando a troca de mensagens ... entre os containers.  

A imagem Docker do sistema é construída com base nos Dockerfiles que inclui as dependências necessárias, mantendo o ambiente leve e eficiente.

## Como Executar
### Pré-requisitos
Certifique-se de ter os seguintes softwares instalados na máquina:
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://go.dev/) *Opcional – Para testes locais fora dos contêineres

### Passo a passo
1. Clone o repositório:
   ```bash
   git clone https://github.com/usuario/nome-do-repositorio.git
   cd nome-do-repositorio
   ```
2. Compile as imagens Docker e inicie o sistema:
   ```bash
   docker-compose up build -d
   ```
Isso iniciará os contêineres do servidor, pontos de recarga e veículos, todos conectados em uma rede Docker interna.

3. Em seguida execute para ter acesso a interface dos clientes.
    ```bash
    docker-compose exec veiculo sh
    ```
    ou
    ```bash
    docker-compose exec servidor sh
    ```
4. Por fim ao entrar no terminal do cotainer, executa o último comando, para executar a aplicação.
    ```bash
    ./veiculo
    ```
    ou 
    ```bash
    ./servidor
    ```
5. Para encerrar:
   ```bash
   docker-compose down
   ```

Caso deseje ver os logs do servidor, execute em outro terminal:  
    ```
    docker compose logs -f servidor
    ```  
    (servidor ou veiculo)
## Tecnologias Utilizadas
- Linguagem: Go (Golang)
- Comunicação: sockets TCP/IP
- Execução: Docker, Docker Compose
- Mock de dados: JSON

## Conclusão
O desenvolvimento deste sistema permitiu aplicar na prática conceitos fundamentais de redes de computadores, comunicação baseada em MQTT e API REST e concorrência distribuída. A arquitetura ... foi estruturada para garantir escalabilidade, paralelismo e integridade na troca de mensagens entre veículo e o servidores.  

Com o uso de mutexes foi possível garantir o controle adequado de concorrência, especialmente no gerenciamento das reservas dos pontos e acesso as estruturas de dados. O sistema também se beneficiou da persistência temporária de dados em memória, otimizando a resposta às requisições.  

Além disso, a utilização do Docker e do Docker Compose tornou possível a simulação de múltiplos componentes operando simultaneamente em um ambiente isolado, facilitando os testes e validações da aplicação.  

Como resultado, o sistema atendeu aos requisitos propostos, oferecendo uma solução eficiente e didática para o gerenciamento de recargas de veículos elétricos com requisições atômicas. A experiência proporcionou uma compreensão mais profunda sobre infraestrutura de comunicação em tempo real, concorrência segura, e práticas de desenvolvimento com conteinerização.  

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
