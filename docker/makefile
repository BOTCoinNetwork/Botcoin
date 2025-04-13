BOTCOIN_VERSION=0.3.6
GIVERNY_VERSION=0.3.5

all: solidity botcoin-image giverny-image

solidity:
	wget -O solc https://github.com/ethereum/solidity/releases/download/v0.5.10/solc-static-linux
	chmod +x solc

botcoin-image:
	go build \
		--ldflags '-extldflags "-static"' \
		-o botcoin/botcoin ../cmd/botcoin/
	cp solc botcoin/solc
	docker build --no-cache=true -t botcoinnetwork/botcoin:$(BOTCOIN_VERSION) -t botcoinnetwork/Botcoin:latest ./botcoin/

giverny-image:
	go build \
		--ldflags '-extldflags "-static"' \
		-o giverny/giverny ../cmd/giverny/
	cp solc giverny/solc
	docker build --no-cache=true -t botcoinnetwork/giverny:$(GIVERNY_VERSION) -t botcoinnetwork/giverny:latest ./giverny/

.PHONY: solidity botcoin-image giverny-image
