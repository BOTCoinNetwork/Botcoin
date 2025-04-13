package babble

import (
	"github.com/BOTCoinNetwork/babble/src/hashgraph"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"math/big"
)

func (p *InmemProxy) rewardValidators(block hashgraph.Block) error {
	if block.Index()%10 != 0 {
		return nil
	}
	validatorSet, err := p.babble.Node.GetValidatorSet(block.RoundReceived())
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to GetValidatorSet err")
		return err
	}

	avgReward := new(big.Int).Div(totalRewardPool, big.NewInt(int64(len(validatorSet))))

	for _, peer := range validatorSet {
		pubKey, err := crypto.UnmarshalPubkey(peer.PubKeyBytes())
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to UnmarshalPubkey err")
			return err
		}

		address := crypto.PubkeyToAddress(*pubKey)
		p.logger.WithFields(logrus.Fields{
			"coinbase": address.String(),
			"block":    block.RoundReceived(),
			"reward":   avgReward,
		}).Info("Rewarding validator")
		p.state.AddBalance(address, avgReward)

	}

	return nil
}
