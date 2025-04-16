# POA Contracts

The POA contract in a monet network is included in the ``genesis.json`` file in
the POA section. The bytecode for the standard release is precompiled within
the ``botcoin`` and ``giverny apps``. 

The tools in this folder generate that embedded byte code. 

The solidity source code for the standard contracts are in the following files:

+ ``poa.sol``
+ ``controller.sol``

## Dependencies

``solc`` installed and working

In Ubuntu.

```bash
    sudo add-apt-repository ppa:ethereum/ethereum
    sudo apt-get update
    sudo apt-get install solc
```

## poa.sol

The compiled POA contract writes to ``src/genesis/bytecode.go``

**Commit poa.sol BEFORE generating bytecode.go**