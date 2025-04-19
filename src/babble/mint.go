package babble

import (
	"github.com/BOTCoinNetwork/Botcoin/src/configuration"
	"github.com/BOTCoinNetwork/babble/src/hashgraph"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"math/big"
)

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

func (p *InmemProxy) rewardValidators(block hashgraph.Block) error {
	// Check if it's time to distribute rewards
	if p.checkRound(block) {
		return nil
	}

	currentRewardPool := p.getTotalRewardPool(block)
	validatorSet, err := p.babble.Node.GetValidatorSet(block.RoundReceived())
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to GetValidatorSet err")
		return err
	}

	currentReward := new(big.Int).Div(new(big.Int).Mul(currentRewardPool, big.NewInt(int64(p.rewardRule.validatorRate))), big.NewInt(100))

	avgReward := new(big.Int).Div(currentReward, big.NewInt(int64(len(validatorSet))))

	for _, peer := range validatorSet {
		pubKey, err := crypto.UnmarshalPubkey(peer.PubKeyBytes())
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to UnmarshalPubkey err")
			return err
		}

		address := crypto.PubkeyToAddress(*pubKey)
		p.logger.WithFields(logrus.Fields{
			"currentRewardPool":  currentReward,
			"coinbase":           address.String(),
			"blockRoundReceived": block.RoundReceived(),
			"reward":             avgReward,
		}).Info("Rewarding validator")
		p.state.AddBalance(address, avgReward)

	}

	return nil
}

func (p *InmemProxy) rewardStakers(block hashgraph.Block) error {
	if p.checkRound(block) {
		return nil
	}

	stakerArray, err := p.getStakerArray()
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getStakerArray err")
		return err
	}

	totalRewardPool := p.getTotalRewardPool(block)

	currentReward := new(big.Int).Div(new(big.Int).Mul(totalRewardPool, big.NewInt(int64(p.rewardRule.stakerRate))), big.NewInt(100))

	totalStakeAmount, err := p.getTotalStaked()
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getTotalStaked err")
		return err
	}

	p.logger.WithFields(logrus.Fields{
		"totalStakeAmount":   totalStakeAmount,
	}).Info("Rewarding staker")

	for _, staker := range stakerArray {
		p.logger.WithFields(logrus.Fields{
			"currentStaker": staker.String(),
		}).Info("Rewarding staker")
		amount, err := p.checkStake(staker)
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to checkStake err")
			return err
		}

		stakerReward := new(big.Int).Div(new(big.Int).Mul(currentReward, amount), totalStakeAmount)

		p.logger.WithFields(logrus.Fields{
			"currentRewardPool":  currentReward,
			"coinbase":           staker.String(),
			"blockRoundReceived": block.RoundReceived(),
			"reward":             stakerReward,
			"stakerRate":         new(big.Int).Div(new(big.Int).Mul(amount, big.NewInt(100)), totalStakeAmount),
			"stakeAmount":        amount,
		}).Info("Rewarding staker")
		p.state.AddBalance(staker, stakerReward)
	}

	return nil
}
