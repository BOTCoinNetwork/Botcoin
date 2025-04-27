package babble

import (
	"math/big"

	"github.com/BOTCoinNetwork/Botcoin/src/configuration"
	"github.com/BOTCoinNetwork/babble/src/hashgraph"
	"github.com/BOTCoinNetwork/babble/src/peers"
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

	totalRewardPool := p.getTotalRewardPool(block)

	currentReward := new(big.Int).Div(new(big.Int).Mul(totalRewardPool, big.NewInt(int64(p.rewardRule.validatorRate))), big.NewInt(100))

	rewardData := map[string]*big.Int{
		"verifyCurrentReward": currentReward,
		"totalRewardPool":     totalRewardPool,
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

	return rewardData, nil
}
