package babble

import (
	"github.com/BOTCoinNetwork/babble/src/hashgraph"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"math/big"
)

type RewardRule struct {
	halving     int
	rewardRound int
	rewardPool  *big.Int
}

func NewRewardRule(reward int64, halving, rewardRound int) *RewardRule {
	return &RewardRule{
		halving:     halving,
		rewardRound: rewardRound,
		rewardPool: new(big.Int).Mul(
			big.NewInt(reward),
			new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
	}
}

func (p *InmemProxy) rewardValidators(block hashgraph.Block) error {
	// Check if it's time to distribute rewards
	if block.Index()%p.rewardRule.rewardRound != 0 || block.Index() == 0 {
		return nil
	}

	// Calculate the current reward amount
	halvingCount := block.Index() / p.rewardRule.halving
	currentReward := new(big.Int).Set(p.rewardRule.rewardPool)
	for i := 0; i < halvingCount; i++ {
		currentReward.Div(currentReward, big.NewInt(2))
	}
	validatorSet, err := p.babble.Node.GetValidatorSet(block.RoundReceived())
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to GetValidatorSet err")
		return err
	}

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
