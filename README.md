<img width="75px" height="75px" align="right" alt="BOTCoin Logo" src="docs/_static/Botcoin_logo.png" title="Botcoin"/>

# BOTCoin

BOTCoin uses peer-to-peer network system technology at the bottom layer, allowing intelligent robots to become autonomous nodes in the decentralized network. It also integrates BVM to implement hardware-level smart contract programming, allowing the robot's own data and programs to be executed safely and reliably. Based on the Hashgraph consensus algorithm, It is a Directed Acyclic Graph (DAG) . 
[BVM](https://github.com/BOTCoinNetwork/BVM) and
[Babble](https://github.com/BOTCoinNetwork/babble).

The **BOTCoin Tools** underpins the
[BOTCoin Hub](https://botcoin.network/), but it is also available for use in 
other projects. You can read more about BOTCoin in the 
[Whitepaper](https://botcoin.network/BOTCoin_en.pdf).

This repository contains the code for `botcoin` and `giverny`:

- `botcoin` is the server process that validators are expected to run.
- `giverny` is the CLI for advanced testnet creation and management.

## Documentation


Full documentation is availble on
[ReadTheDocs](https://botcoin.network/docs/html/index.html).

## Quick Start

For the impatient, we recommend you start here:

-   [Installation
    Documents](https://botcoin.network/docs/html/install.html)
-   [Quick
    Start](https://botcoin.network/docs/html/quickstart.html#quickstart-rst)


## Build botcoin
need linux environment
```shell
go mod tidy
cd cmd/botcoin
go build -o botcoin
```

## License

BOTCoin is released under the MIT license. See [LICENSE](LICENSE).