package babble

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

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

	currentRewardPool := p.getTotalRewardPool(block)

	currentReward := new(big.Int).Div(new(big.Int).Mul(currentRewardPool, big.NewInt(int64(p.rewardRule.stableRate))), big.NewInt(100))

	rewardData := map[string]*big.Int{
		"stablePeersCurrentReward": currentReward,
	}

	return rewardData, nil
}

func (p *InmemProxy) getAllIndexSum(block hashgraph.Block, validators []*peers.Peer) (int, map[string]*stablePeer) {
	currentIndex := block.Index()
	allIndexSum := 0
	rewardPeersMap := make(map[string]*stablePeer)

	for _, validator := range validators {
		index, err := p.babble.Store.GetPeerJoinIndex(validator.PubKeyHex)
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to GetPeerJoinIndex err")
			return allIndexSum, rewardPeersMap
		}

		if index.PubKeyHex == "" {
			continue
		}
		pubKey, err := crypto.UnmarshalPubkey(validator.PubKeyBytes())
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to UnmarshalPubkey err")
			return allIndexSum, rewardPeersMap
		}

		address := crypto.PubkeyToAddress(*pubKey)

		rewardPeersMap[address.String()] = &stablePeer{
			addr:  address,
			index: index.Index,
		}

		allIndexSum += currentIndex - index.Index
	}
	// p.logger.WithFields(logrus.Fields{
	// 	"rewardPeersMap": rewardPeersMap,
	// }).Info("rewardPeersMap")

	return allIndexSum, rewardPeersMap
}

func (p *InmemProxy) makeRewards(block hashgraph.Block, validators []*peers.Peer) peers.Mint {
	rewardData_Validators, err := p.rewardValidators(block, validators)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward validators")
	}

	var totalCurrentReward = new(big.Int)
	var reward_Validators = new(big.Int)
	var reward_Stake = new(big.Int)
	var reward_StablePeers = new(big.Int)
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
		reward_StablePeers = rewardData_StablePeers["stablePeersCurrentReward"]
		p.logger.WithFields(logrus.Fields{
			"stablePeersCurrentReward": reward_StablePeers,
		}).Info("rewardData_StablePeers")
	}

	mint := peers.Mint{
		MintRewards:      totalCurrentReward.String(),
		TotalStakeAmount: totalStakeAmount.String(),
		PeersCount:       len(validators),
	}

	var mintSet = peers.MintSet{}
	mintSet.Mint = mint

	// block.Body.MintRewards = currentReward.String()
	// block.AppendTransactions([][]byte{bytesData})
	if totalCurrentReward.Cmp(big.NewInt(0)) > 0 {

		peerRewards := map[string]peers.PeerReward{}
		allIndexSum, rewardPeersMap := p.getAllIndexSum(block, validators)
		currentIndex := block.Index()

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

			peerReward = peers.PeerReward{
				VerifyReward: rewardAmount.String(),
			}
			p.logger.WithFields(logrus.Fields{
				"verifyReward": rewardAmount,
				"verifyAddr":   addr,
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

			if allIndexSum > 0 {
				v, ok := rewardPeersMap[addr]
				if ok {
					historyReward := new(big.Int).Div(new(big.Int).Mul(reward_StablePeers, big.NewInt(int64(currentIndex-v.index))), big.NewInt(int64(allIndexSum)))
					rewardAmount = rewardAmount.Add(rewardAmount, historyReward)
					p.logger.WithFields(logrus.Fields{
						"historyReward": historyReward,
						"joinIndex":     v.index,
						"currentIndex":  currentIndex,
					}).Info("Rewarding")
					peerReward.HistoryReward = historyReward.String()
				}
			}

			p.state.AddBalance(address, rewardAmount)
			peerRewards[addr] = peerReward

			mintSet.PeerReward = peerRewards

		}
		// p.babble.Store.SetMinthistory(currentIndex, mintSet)
		jsMint := peers.NewJSONMint(p.babble.Config.DataDir)
		jsMint.SetMinthistoryForJson(currentIndex, mintSet)
		return mint
	}
	return mint
}
