.. _tests_rst:

Tests
=====

Included in the botcoin distribution are numerous tests. There are unit tests,
which test individual components, and end to end tests.

Unit Tests
----------

These can be run as follows:

.. code:: bash

    [...]/botcoin$ make test

    Botcoin Tests

    ?       .../botcoin/cmd/giverny  [no test files]
    ?       .../botcoin/cmd/giverny/commands [no test files]
    ?       .../botcoin/cmd/giverny/commands/keys    [no test files]
    ?       .../botcoin/cmd/giverny/commands/network [no test files]
    ?       .../botcoin/cmd/giverny/commands/server  [no test files]
    ?       .../botcoin/cmd/giverny/commands/transactions    [no test files]
    ?       .../botcoin/cmd/giverny/configuration    [no test files]
    ?       .../botcoin/cmd/botcoin   [no test files]
    ?       .../botcoin/cmd/botcoin/commands  [no test files]
    ?       .../botcoin/cmd/botcoin/commands/config   [no test files]
    ?       .../botcoin/cmd/botcoin/commands/keys     [no test files]
    ok      .../botcoin/src/babble   0.077s
    ok      .../botcoin/src/common   0.003s
    ?       .../botcoin/src/config   [no test files]
    ?       .../botcoin/src/configuration    [no test files]
    ?       .../botcoin/src/contract [no test files]
    ?       .../botcoin/src/crypto   [no test files]
    ?       .../botcoin/src/docker   [no test files]
    ?       .../botcoin/src/files    [no test files]
    ?       .../botcoin/src/peers    [no test files]
    ?       .../botcoin/src/types    [no test files]
    ?       .../botcoin/src/version  [no test files]

    EVM-Lite Tests

    ?       .../vendor/.../evm-lite/src/common      [no test files]
    ?       .../vendor/.../evm-lite/src/config      [no test files]
    ?       .../vendor/.../evm-lite/src/consensus   [no test files]
    ?       .../vendor/.../evm-lite/src/consensus/solo      [no test files]
    ok      .../vendor/.../evm-lite/src/currency    0.003s
    ?       .../vendor/.../evm-lite/src/engine      [no test files]
    ?       .../vendor/.../evm-lite/src/service     [no test files]
    ok      .../vendor/.../evm-lite/src/state       3.148s
    ?       .../vendor/.../evm-lite/src/version     [no test files]

    Babble Tests

    ok      .../vendor/.../babble/src/babble        0.149s
    ok      .../vendor/.../babble/src/common        0.024s
    ?       .../vendor/.../babble/src/config        [no test files]
    ?       .../vendor/.../babble/src/crypto        [no test files]
    ok      .../vendor/.../babble/src/crypto/keys   0.097s
    ok      .../vendor/.../babble/src/hashgraph     11.385s
    ?       .../vendor/.../babble/src/mobile        [no test files]
    ok      .../vendor/.../babble/src/net   0.092s
    ok      .../vendor/.../babble/src/node  36.339s
    ok      .../vendor/.../babble/src/peers 0.082s
    ?       .../vendor/.../babble/src/proxy [no test files]
    ok      .../vendor/.../babble/src/proxy/dummy   0.038s
    ok      .../vendor/.../babble/src/proxy/inmem   0.037s
    ok      .../vendor/.../babble/src/proxy/socket  0.043s
    ?       .../vendor/.../babble/src/proxy/socket/app      [no test files]
    ?       .../vendor/.../babble/src/proxy/socket/babble   [no test files]
    ?       .../vendor/.../babble/src/service       [no test files]
    ?       .../vendor/.../babble/src/version       [no test files]



They will take some seconds to run. If any test fails an error message will be
displayed.

End to End Tests
----------------

End to end tests are in the subfolder ``e2e`` of the repository. All tests
can be run using either of the commands below. Note the different paths:

.. code:: bash

    [...]/botcoin/e2e$ make tests
    [...]/botcoin$ make e2e


An individual test can be run as follows:

.. code:: bash

    [...]/botcoin/e2e$ make test TEST=crowdfundnet

To prevent the test net being destroyed on completion, add ``NOSTOP=nostop``.
This allows you to interrogate the network after the test has completed:

.. code:: bash

    [...]/botcoin/e2e$ make test TEST=transfer_03_10 NOSTOP=nostop


Tests output logs to ``...botcoin/e2e/tests/<TESTNAME>.out``


Transfer Tests
--------------

As well as standalone tests, the transaction generation tools can be used
against extant networks.

You can get the list of options (and defaults) by using the ``--help`` or
``-h`` option:

.. code:: bash

    $ e2e/tools/build-trans.sh -h
    e2e/tools/build-trans.sh [-v] [--accounts=10] [--transactions=200] [--faucet="Faucet"] [--faucet-config-dir=] [--prefix=Test] [--node-name=Node] [--node-host=172.77.5.11] [--node-port=8080] [--config-dir=/home/jon/.monettest] [--temp-dir=/tmp] [-h|--help]


+ **-v** turns on verbose output
+ **--accounts=10** sets the number of accounts to transfer tokens between
+ **--transactions=200** sets the number of transactions to generate
+ **--faucet="Faucet"** sets the account to fund the transfers
+ **--faucet-config-dir=** where the faucet account is stored.
  ``$HOME/.monet/keystore`` or ``$HOME/.giverny/networks/<net name>/keystore``
  are the likely values
+ **--prefix=Test** is the prefix for the moniker of the accounts for transfers
+ **--node-name=Node** is the Node Name
+ **--node-host=172.77.5.11** is the Node address
+ **--node-port=8080** is the port for EVM-Lite endpoints
+ **--config-dir=/home/user/.monettest** is the config directory to use


