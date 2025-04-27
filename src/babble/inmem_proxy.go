package babble

import (
	"math/big"

	"github.com/BOTCoinNetwork/BVM/src/service"
	"github.com/BOTCoinNetwork/BVM/src/state"
	"github.com/BOTCoinNetwork/babble/src/babble"
	"github.com/BOTCoinNetwork/babble/src/crypto/keys"
	"github.com/BOTCoinNetwork/babble/src/hashgraph"
	"github.com/BOTCoinNetwork/babble/src/peers"
	"github.com/BOTCoinNetwork/babble/src/proxy"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

// InmemProxy implements the Babble AppProxy interface
type InmemProxy struct {
	service    *service.Service
	state      *state.State
	babble     *babble.Babble
	submitCh   chan []byte
	logger     *logrus.Entry
	rewardRule *RewardRule
}

// NewInmemProxy initializes and return a new InmemProxy
func NewInmemProxy(state *state.State,
	service *service.Service,
	babble *babble.Babble,
	submitCh chan []byte,
	logger *logrus.Entry,
	rule *RewardRule) *InmemProxy {

	return &InmemProxy{
		service:    service,
		state:      state,
		babble:     babble,
		submitCh:   submitCh,
		logger:     logger,
		rewardRule: rule,
	}
}

/*
******************************************************************************
Implement Babble AppProxy Interface
******************************************************************************
*/

// SubmitCh is the channel through which the Service sends transactions to the
// node.
func (p *InmemProxy) SubmitCh() chan []byte {
	return p.submitCh
}

// CommitBlock applies the block's transactions to the state and commits. All
// transaction fees are sent to the coinbase address, which is computed from the
// block and the current validator-set. It also checks the block's internal
// transactions against the POA smart-contract to verify if joining peers are
// authorised to become validators in Babble. It returns the resulting
// state-hash and internal transaction receips.
func (p *InmemProxy) CommitBlock(block hashgraph.Block) (proxy.CommitResponse, error) {

	coinbaseAddress, validators, err := p.getCoinbase(block)
	if err != nil {
		return proxy.CommitResponse{}, err
	}

	p.logger.WithFields(logrus.Fields{
		"validators":    validators,
		"coinbase":      coinbaseAddress.String(),
		"blockIndex":    block.Index(),
		"RoundReceived": block.RoundReceived(),
	}).Info("Commit")

	blockHashBytes, err := block.Hash()
	blockHash := ethCommon.BytesToHash(blockHashBytes)

	for i, tx := range block.Transactions() {
		if err := p.state.ApplyTransaction(tx, i, blockHash, coinbaseAddress); err != nil {
			p.logger.WithError(err).Errorf("Failed to apply tx %d of %d", i+1, len(block.Transactions()))
		}
	}

	hash, err := p.state.Commit()
	if err != nil {
		return proxy.CommitResponse{}, err
	}

	internalTransactionReceipts := p.processInternalTransactions(block.InternalTransactions())

	evictionReceipts := p.processEvictions(block)

	receipts := append(internalTransactionReceipts, evictionReceipts...)

	// reward ing
	rewardReceipts := p.processRewardInternalTransactionsReceipts(block, validators)
	receipts = append(receipts, rewardReceipts...)

	res := proxy.CommitResponse{
		StateHash:                   hash.Bytes(),
		InternalTransactionReceipts: receipts,
	}

	return res, nil
}

func (p *InmemProxy) processRewardInternalTransactionsReceipts(block hashgraph.Block, validators []*peers.Peer) []hashgraph.InternalTransactionReceipt {

	receipts := []hashgraph.InternalTransactionReceipt{}

	rewardData_Validators, err := p.rewardValidators(block, validators)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward validators")
	}
	var totalCurrentReward = new(big.Int)
	var reward_Validators = new(big.Int)
	var reward_Stake = new(big.Int)
	var totalStakeAmount = new(big.Int)
	// create a new Transaction and add it to the block
	if rewardData_Validators != nil {
		totalCurrentReward = rewardData_Validators["totalRewardPool"]
		reward_Validators = rewardData_Validators["verifyCurrentReward"]

		p.logger.WithFields(logrus.Fields{
			"totalRewardPool":     totalCurrentReward,
			"verifyCurrentReward": reward_Validators,
		}).Info("Total_rewardData")
	}

	rewardData_Stake, err := p.rewardStakers(block)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward stakers")
	}
	// create a new Transaction and add it to the block
	if rewardData_Stake != nil {
		reward_Stake = rewardData_Stake["stakerCurrentReward"]
		totalStakeAmount = rewardData_Stake["totalStakeAmount"]
		p.logger.WithFields(logrus.Fields{
			"stakerCurrentReward": reward_Stake,
			"totalStakeAmount":    totalStakeAmount,
		}).Info("Total_rewardData")
	}

	rewardData_StablePeers, err := p.rewardStablePeer(block, validators)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward stable peers")
	}
	// create a new Transaction and add it to the block
	if rewardData_StablePeers != nil {
		currentReward = currentReward.Add(currentReward, rewardData_StablePeers["stablePeersCurrentReward"])
		p.logger.WithFields(logrus.Fields{
			"stablePeersCurrentReward": currentReward,
		}).Info("rewardData_StablePeers")
	}

	// block.Body.MintRewards = currentReward.String()
	// block.AppendTransactions([][]byte{bytesData})
	if totalCurrentReward.Cmp(big.NewInt(0)) > 0 {

		peerRewards := map[string]peers.PeerReward{}

		for _, peer := range validators {
			pubKey, err := crypto.UnmarshalPubkey(peer.PubKeyBytes())
			if err != nil {
				p.logger.WithError(err).Errorf("Failed to UnmarshalPubkey err")
			}

			rewardAmount := new(big.Int).Div(reward_Validators, big.NewInt(int64(len(validators))))
			address := crypto.PubkeyToAddress(*pubKey)
			var addr = address.String()

			var peerReward = peers.PeerReward{
				VerifyReward: rewardAmount.String(),
			}

			p.logger.WithFields(logrus.Fields{
				"verifyAddr":   addr,
				"verifyReward": rewardAmount,
			}).Info("Rewarding")

			stakerAmount, err := p.checkStake(address)
			if err == nil && stakerAmount.Cmp(big.NewInt(0)) > 0 {

				stakeReward := new(big.Int).Div(new(big.Int).Mul(reward_Stake, stakerAmount), totalStakeAmount)
				rewardAmount = rewardAmount.Add(rewardAmount, stakeReward)

				peerReward.StakeReward = stakeReward.String()
				peerReward.StakeAmount = stakerAmount.String()

				p.logger.WithFields(logrus.Fields{
					"stakerReward": stakeReward,
					"stakerRate%":  new(big.Float).Quo(new(big.Float).Mul(big.NewFloat(float64(stakerAmount.Int64())), big.NewFloat(100)), big.NewFloat(float64(totalStakeAmount.Int64()))),
				}).Info("Rewarding")
			}

			p.state.AddBalance(address, rewardAmount)
			peerRewards[addr] = peerReward
			// hashgraph.Store.SetMinthistory(block.RoundReceived(), )
		}

		var mintInfo = peers.Mint{
			MintRewards:      totalCurrentReward.String(),
			TotalStakeAmount: totalStakeAmount.String(),
			PeersCount:       len(validators),
		}

		var mintTransactions = hashgraph.NewMintInternalTransaction(hashgraph.Mint_Rewards, mintInfo)
		var mintReceipts = hashgraph.InternalTransactionReceipt{
			InternalTransaction: mintTransactions,
			Accepted:            false,
		}
		block.Body.InternalTransactions = append(block.Body.InternalTransactions, mintTransactions)
		// block.Body.InternalTransactionReceipts = append(block.Body.InternalTransactionReceipts, mintReceipts)
		return []hashgraph.InternalTransactionReceipt{mintReceipts}
	}
	return receipts
}

// getCoinbase returns the coinbase address which will receive all the
// transaction fees from the block. It is meant to be a safe and fair selection
// process from the current Babble validator-set. We use the block hash, which
// is pseudo-random, but equal for all validators, to select a validator from
// the current validator-set.
func (p *InmemProxy) getCoinbase(block hashgraph.Block) (ethCommon.Address, []*peers.Peer, error) {
	coinbaseAddress := ethCommon.Address{}
	validators := []*peers.Peer{}

	if p.babble != nil {
		babbleValidators, err := p.babble.Node.GetValidatorSet(block.RoundReceived())

		if err != nil {
			return coinbaseAddress, babbleValidators, err
		}

		blockHash, _ := block.Hash()
		blockRand := keys.Hash32(blockHash)

		coinbaseValidator := babbleValidators[blockRand%uint32(len(babbleValidators))]

		coinbasePubKey, err := crypto.UnmarshalPubkey(coinbaseValidator.PubKeyBytes())
		if err != nil {
			return coinbaseAddress, babbleValidators, err
		}

		validators = make([]*peers.Peer, len(babbleValidators))
		copy(validators, babbleValidators)
		if err != nil {
			return coinbaseAddress, babbleValidators, err
		}
		coinbaseAddress = crypto.PubkeyToAddress(*coinbasePubKey)
	}

	return coinbaseAddress, validators, nil
}

// processInternalTransactions decides if InternalTransactions should be
// accepted. For PEER_ADD transactions, it checks if the peer is authorised in
// the POA smart-contract. All PEER_REMOVE transactions are accepted.
func (p *InmemProxy) processInternalTransactions(internalTransactions []hashgraph.InternalTransaction) []hashgraph.InternalTransactionReceipt {
	receipts := []hashgraph.InternalTransactionReceipt{}

	for _, tx := range internalTransactions {
		switch tx.Body.Type {
		case hashgraph.PEER_ADD:
			pk, err := crypto.UnmarshalPubkey(tx.Body.Peer.PubKeyBytes())
			if err != nil {
				p.logger.Warningf("couldn't unmarshal pubkey bytes: %v", err)
			}

			addr := crypto.PubkeyToAddress(*pk)

			ok, err := p.state.CheckAuthorised(addr)

			if err != nil {
				p.logger.WithError(err).Error("Error in checkAuthorised")
				receipts = append(receipts, tx.AsRefused())
			} else {
				if ok {
					p.logger.WithField("addr", addr.String()).Info("Accepted peer")
					receipts = append(receipts, tx.AsAccepted())
				} else {
					p.logger.WithField("addr", addr.String()).Info("Rejected peer")
					receipts = append(receipts, tx.AsRefused())
				}
			}
		case hashgraph.PEER_REMOVE:
			receipts = append(receipts, tx.AsAccepted())
		}
	}

	return receipts
}

// processEvictions compares the current validator-set to the whitelist and
// creates InternalTransactionReceipts to evict any current validator which is
// not in the whitelist.
func (p *InmemProxy) processEvictions(block hashgraph.Block) []hashgraph.InternalTransactionReceipt {
	receipts := []hashgraph.InternalTransactionReceipt{}

	if p.babble != nil {
		babbleValidators, err := p.babble.Node.GetValidatorSet(block.RoundReceived())
		if err != nil {
			p.logger.WithError(err).Error("Error GetValidatorSet")
			return receipts
		}

		for _, val := range babbleValidators {
			pk, err := crypto.UnmarshalPubkey(val.PubKeyBytes())
			if err != nil {
				p.logger.Warningf("couldn't unmarshal pubkey bytes: %v", err)
				continue
			}

			addr := crypto.PubkeyToAddress(*pk)

			ok, err := p.state.CheckAuthorised(addr)

			if err != nil {
				p.logger.WithError(err).Error("Error in checkAuthorised")
			} else {
				if !ok {
					p.logger.WithField("addr", addr.String()).Info("Ejected peer")
					receipts = append(receipts,
						hashgraph.InternalTransactionReceipt{
							InternalTransaction: hashgraph.InternalTransaction{
								Body: hashgraph.InternalTransactionBody{
									Type: hashgraph.PEER_REMOVE,
									Peer: *val,
								},
							},
							Accepted: true,
						})
				}
			}
		}
	}

	return receipts
}

//TODO - Implement these two functions

// GetSnapshot will generate a snapshot
func (p *InmemProxy) GetSnapshot(blockIndex int) ([]byte, error) {
	return []byte{}, nil
}

// Restore will restore a snapshot
func (p *InmemProxy) Restore(snapshot []byte) error {
	return nil
}
