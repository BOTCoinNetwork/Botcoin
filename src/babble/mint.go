package babble

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"

	"github.com/BOTCoinNetwork/Botcoin/src/configuration"
	"github.com/BOTCoinNetwork/babble/src/hashgraph"
	"github.com/BOTCoinNetwork/babble/src/peers"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

type rewardData map[string]*big.Int

type RewardRule struct {
	halving       int
	rewardRound   int
	rewardPool    *big.Int
	validatorRate int
	stakerRate    int
	stableRate    int
}

func NewRewardRule(config *configuration.BaseConfig) *RewardRule {
	return &RewardRule{
		halving:     config.ValidatorHalvingRound,
		rewardRound: config.ValidatorRewardRound,
		rewardPool: new(big.Int).Mul(
			big.NewInt(int64(config.ValidatorRewardPool)),
			new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
		validatorRate: config.ValidatorRate,
		stakerRate:    config.StakerRate,
		stableRate:    config.StableRate,
	}
}

func (p *InmemProxy) checkRound(block hashgraph.Block) bool {
	return block.Index()%p.rewardRule.rewardRound != 0 || block.Index() == 0
}

func (p *InmemProxy) getTotalRewardPool(block hashgraph.Block) *big.Int {
	halvingCount := block.Index() / p.rewardRule.halving
	currentReward := new(big.Int).Set(p.rewardRule.rewardPool)
	for i := 0; i < halvingCount; i++ {
		currentReward.Div(currentReward, big.NewInt(2))
	}

	return currentReward
}

func (p *InmemProxy) rewardValidators(block hashgraph.Block, validators []*peers.Peer) (rewardData, error) {
	// Check if it's time to distribute rewards
	if p.checkRound(block) {
		return nil, nil
	}
	if validators == nil || len(validators) == 0 {
		p.logger.WithFields(logrus.Fields{
			"validators": validators,
		}).Info("validators")
		return nil, nil
	}

	currentRewardPool := p.getTotalRewardPool(block)

	currentReward := new(big.Int).Div(new(big.Int).Mul(currentRewardPool, big.NewInt(int64(p.rewardRule.validatorRate))), big.NewInt(100))

	avgReward := new(big.Int).Div(currentReward, big.NewInt(int64(len(validators))))

	rewardData := map[string]*big.Int{
		"verifyCurrentReward": currentReward,
	}

	for _, peer := range validators {
		pubKey, err := crypto.UnmarshalPubkey(peer.PubKeyBytes())
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to UnmarshalPubkey err")
			return rewardData, err
		}

		address := crypto.PubkeyToAddress(*pubKey)
		p.logger.WithFields(logrus.Fields{
			"coinbase":     address.String(),
			"verifyReward": avgReward,
		}).Info("Rewarding verify")
		p.state.AddBalance(address, avgReward)

	}

	return rewardData, nil
}

func (p *InmemProxy) rewardStakers(block hashgraph.Block) (rewardData, error) {
	if p.checkRound(block) {
		return nil, nil
	}

	stakerArray, err := p.getStakerArray()
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getStakerArray err")
		return nil, err
	}

	if len(stakerArray) == 0 {
		return nil, nil
	}

	totalRewardPool := p.getTotalRewardPool(block)

	currentReward := new(big.Int).Div(new(big.Int).Mul(totalRewardPool, big.NewInt(int64(p.rewardRule.stakerRate))), big.NewInt(100))

	totalStakeAmount, err := p.getTotalStaked()

	rewardData := map[string]*big.Int{
		"stakerCurrentReward": currentReward,
		"totalStakeAmount":    totalStakeAmount,
	}

	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getTotalStaked err")
		return rewardData, err
	}

	for _, staker := range stakerArray {

		amount, err := p.checkStake(staker)
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to checkStake err")
			return rewardData, err
		}

		stakerReward := new(big.Int).Div(new(big.Int).Mul(currentReward, amount), totalStakeAmount)

		p.logger.WithFields(logrus.Fields{
			"currentStaker": staker.String(),
			"stakerReward":  stakerReward,
			"stakerRate%":   new(big.Float).Quo(new(big.Float).Mul(big.NewFloat(float64(amount.Int64())), big.NewFloat(100)), big.NewFloat(float64(totalStakeAmount.Int64()))),
			"stakeAmount":   amount,
		}).Info("Rewarding staker")
		p.state.AddBalance(staker, stakerReward)
	}

	return rewardData, nil
}

type stablePeer struct {
	addr  common.Address
	index int
}

func (p *InmemProxy) rewardStablePeer(block hashgraph.Block, validators []*peers.Peer) (rewardData, error) {
	if p.checkRound(block) {
		return nil, nil
	}

	if validators == nil || len(validators) == 0 {
		p.logger.WithFields(logrus.Fields{
			"validators": validators,
		}).Info("validators")
		return nil, nil
	}

	currentIndex := block.Index()

	currentRewardPool := p.getTotalRewardPool(block)

	currentReward := new(big.Int).Div(new(big.Int).Mul(currentRewardPool, big.NewInt(int64(p.rewardRule.stableRate))), big.NewInt(100))

	rewardData := map[string]*big.Int{
		"stablePeersCurrentReward": currentReward,
	}

	allIndexSum := 0
	rewardPeersMap := make(map[string]*stablePeer)

	for _, validator := range validators {
		index, err := p.babble.Store.GetPeerJoinIndex(validator.PubKeyHex)
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to GetPeerJoinIndex err")
			return nil, err
		}

		if index == nil {
			continue
		}
		pubKey, err := crypto.UnmarshalPubkey(validator.PubKeyBytes())
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to UnmarshalPubkey err")
			return rewardData, err
		}

		address := crypto.PubkeyToAddress(*pubKey)

		rewardPeersMap[address.String()] = &stablePeer{
			addr:  address,
			index: index.Index,
		}

		allIndexSum += currentIndex - index.Index
	}

	if allIndexSum == 0 {
		return nil, nil
	}

	for _, v := range rewardPeersMap {
		reward := new(big.Int).Div(new(big.Int).Mul(currentReward, big.NewInt(int64(currentIndex-v.index))), big.NewInt(int64(allIndexSum)))
		p.logger.WithFields(logrus.Fields{
			"stablePeer":   v.addr.String(),
			"reward":       reward,
			"joinIndex":    v.index,
			"currentIndex": currentIndex,
		})
		p.state.AddBalance(v.addr, reward)
	}

	return rewardData, nil
}
