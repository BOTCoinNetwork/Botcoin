BUILD_TAGS?=evml

all: vendor install

# vendor uses Glide to install all the Go dependencies in vendor/
vendor:
	 (rm glide.lock || rm -rf vendor ) && glide install

# install compiles and places the binary in GOPATH/bin
install: installd installgiv

installd:
	go install \
		--ldflags "-X github.com/BOTCoinNetwork/Botcoin/src/version.GitCommit=`git rev-parse HEAD` -X github.com/BOTCoinNetwork/Botcoin/src/version.GitBranch=`git symbolic-ref --short HEAD`" \
		./cmd/botcoin

installgiv:
	go install \
		--ldflags "-X github.com/BOTCoinNetwork/Botcoin/src/version.GitCommit=`git rev-parse HEAD` -X github.com/BOTCoinNetwork/Botcoin/src/version.GitBranch=`git symbolic-ref --short HEAD`" \
		./cmd/giverny

docker:
	$(MAKE) -C docker


e2e: 
	$(MAKE) -C e2e tests    

test: testbotcoin testevml testbabble

testbotcoin:
	@echo "\nBotcoin Tests\n\n" ; glide novendor | xargs go test | sed -e 's?github.com/BOTCoinNetwork/?.../?g'

testevml:
	@echo "\nEVM-Lite Tests\n\n" ; cd vendor/github.com/BOTCoinNetwork/BVM ; go test ./src/... -count=1 -tags=unit | sed -e 's?github.com/BOTCoinNetwork/Botcoin/vendor/github.com/BOTCoinNetwork/?.../vendor/.../?g'

testbabble:
	@echo "\nBabble Tests\n\n" ; cd vendor/github.com/BOTCoinNetwork/babble ;   go test ./src/... -count=1 -tags=unit | sed -e 's?github.com/BOTCoinNetwork/Botcoin/vendor/github.com/BOTCoinNetwork/?.../vendor/.../?g'

dist:
	xgo --targets=*/amd64 --dest=build/  ./cmd/botcoin/ 
 
lint:
	glide novendor | xargs golint

.PHONY: all vendor install installd installcli installgiv test update docker testbotcoin testevml testbabble lint e2e
