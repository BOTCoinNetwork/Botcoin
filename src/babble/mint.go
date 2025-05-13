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

func NewRewardRule(config *configuration.MintConfig) *RewardRule {
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

func (p *InmemProxy) stringAddress(validators []*peers.Peer) []common.Address {

	var vAddress []common.Address
	for _, peer := range validators {
		pubKey, err := crypto.UnmarshalPubkey(peer.PubKeyBytes())
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to UnmarshalPubkey err")
		}
		address := crypto.PubkeyToAddress(*pubKey)
		vAddress = append(vAddress, address)
	}

	return vAddress
}

func (p *InmemProxy) rewardValidators(block hashgraph.Block) (rewardData, error) {

	totalRewardPool := p.getTotalRewardPool(block)

	currentReward := new(big.Int).Div(new(big.Int).Mul(totalRewardPool, big.NewInt(int64(p.rewardRule.validatorRate))), big.NewInt(100))

	rewardData := map[string]*big.Int{
		"verifyCurrentReward": currentReward,
		"totalRewardPool":     totalRewardPool,
	}

	return rewardData, nil
}

func (p *InmemProxy) rewardStakers(block hashgraph.Block, validatorsAddress []common.Address) (rewardData, error) {

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

	var onlineTotalStakeAmount = new(big.Int)
	// sum online's totalStakeAmount
	for _, vAddress := range validatorsAddress {

		amount, err := p.checkStake(vAddress)
		if err == nil {
			onlineTotalStakeAmount.Add(onlineTotalStakeAmount, amount)
		}
	}

	rewardData := map[string]*big.Int{
		"stakerCurrentReward":    currentReward,
		"totalStakeAmount":       totalStakeAmount,
		"onlineTotalStakeAmount": onlineTotalStakeAmount,
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

func (p *InmemProxy) rewardStablePeer(block hashgraph.Block) (rewardData, error) {

	currentRewardPool := p.getTotalRewardPool(block)

	currentReward := new(big.Int).Div(new(big.Int).Mul(currentRewardPool, big.NewInt(int64(p.rewardRule.stableRate))), big.NewInt(100))

	rewardData := map[string]*big.Int{
		"stablePeersCurrentReward": currentReward,
	}

	return rewardData, nil
}

func (p *InmemProxy) getAllAbleValidatorsAndHistorys(block hashgraph.Block, validators []*peers.Peer) ([]common.Address, map[string]*stablePeer, int) {
	currentIndex := block.Index()
	allIndexSum := 0
	historysPeersMap := make(map[string]*stablePeer)
	validatorsPeers := []common.Address{}

	for _, validator := range validators {
		peerIndex, err := p.babble.Store.GetPeerJoinIndex(validator.PubKeyHex)
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to GetPeerJoinIndex err")
			return validatorsPeers, historysPeersMap, allIndexSum
		}

		if peerIndex.PubKeyHex == "" {
			continue
		}
		pubKey, err := crypto.UnmarshalPubkey(validator.PubKeyBytes())
		if err != nil {
			p.logger.WithError(err).Errorf("Failed to UnmarshalPubkey err")
			return validatorsPeers, historysPeersMap, allIndexSum
		}

		address := crypto.PubkeyToAddress(*pubKey)

		historysPeersMap[address.String()] = &stablePeer{
			addr:  address,
			index: peerIndex.Index,
		}

		allIndexSum += currentIndex - peerIndex.Index

		if currentIndex-peerIndex.Index >= configuration.Global.Mint.ValidatorRewardRound {
			validatorsPeers = append(validatorsPeers, address)
		}
	}

	return validatorsPeers, historysPeersMap, allIndexSum
}

func (p *InmemProxy) makeRewards(block hashgraph.Block, validators []*peers.Peer) peers.Mint {

	// Check if it's time to distribute rewards
	if p.checkRound(block) {
		return peers.Mint{}
	}

	ableValidators, ableHistorys, allIndexSum := p.getAllAbleValidatorsAndHistorys(block, validators)

	rewardData_Validators, err := p.rewardValidators(block)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward validators")
	}

	var totalCurrentReward = new(big.Int)
	var reward_Validators = new(big.Int)
	var reward_Stake = new(big.Int)
	var reward_StablePeers = new(big.Int)
	var totalStakeAmount = new(big.Int)
	var onlineTotalStakeAmount = new(big.Int)

	// create a new Transaction and add it to the block
	if rewardData_Validators != nil {
		totalCurrentReward = rewardData_Validators["totalRewardPool"]
		reward_Validators = rewardData_Validators["verifyCurrentReward"]

		p.logger.WithFields(logrus.Fields{
			"totalRewardPool": totalCurrentReward,
		}).Info("Total_rewardData")
		p.logger.WithFields(logrus.Fields{
			"verify_CurrentReward": reward_Validators,
		}).Info("Total_rewardData")
	}

	// stakersAddress must be use validators
	stakersAddress := p.stringAddress(validators)
	rewardData_Stake, err := p.rewardStakers(block, stakersAddress)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward stakers")
	}
	// create a new Transaction and add it to the block
	if rewardData_Stake != nil {
		reward_Stake = rewardData_Stake["stakerCurrentReward"]
		totalStakeAmount = rewardData_Stake["totalStakeAmount"]
		onlineTotalStakeAmount = rewardData_Stake["onlineTotalStakeAmount"]

		p.logger.WithFields(logrus.Fields{
			"staker_CurrentReward":   reward_Stake,
			"totalStakeAmount":       totalStakeAmount,
			"onlineTotalStakeAmount": onlineTotalStakeAmount,
		}).Info("Total_rewardData")
	}

	rewardData_StablePeers, err := p.rewardStablePeer(block)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward stable peers")
	}
	// create a new Transaction and add it to the block
	if rewardData_StablePeers != nil {
		reward_StablePeers = rewardData_StablePeers["stablePeersCurrentReward"]
		p.logger.WithFields(logrus.Fields{
			"history_CurrentReward": reward_StablePeers,
		}).Info("Total_rewardData")
	}

	vsLen := len(ableValidators)
	if len(validators) > vsLen {
		vsLen = len(validators)
	}
	mint := peers.Mint{
		MintRewards:      totalCurrentReward.String(),
		TotalStakeAmount: totalStakeAmount.String(),
		PeersCount:       vsLen,
	}

	var mintSet = peers.MintSet{}
	mintSet.Mint = mint

	// Reward ing
	if totalCurrentReward.Cmp(big.NewInt(0)) > 0 {

		peerRewards := map[string]peers.PeerReward{}

		currentIndex := block.Index()

		for _, addr := range ableValidators {

			addrString := addr.String()
			rewardAmount := new(big.Int).Div(reward_Validators, big.NewInt(int64(len(ableValidators))))
			var peerReward = peers.PeerReward{
				VerifyReward: rewardAmount.String(),
			}

			peerReward = peers.PeerReward{
				VerifyReward: rewardAmount.String(),
			}
			p.logger.WithFields(logrus.Fields{
				"verifyReward": rewardAmount,
				"verifyAddr":   addrString,
			}).Info("Rewarding")

			stakerAmount, err := p.checkStake(addr)
			if err == nil && stakerAmount.Cmp(big.NewInt(0)) > 0 {

				stakeReward := new(big.Int).Div(new(big.Int).Mul(reward_Stake, stakerAmount), onlineTotalStakeAmount)
				rewardAmount = rewardAmount.Add(rewardAmount, stakeReward)

				peerReward.StakeReward = stakeReward.String()
				peerReward.StakeAmount = stakerAmount.String()

				// 计算 stakerAmount 占 onlineTotalStakeAmount 的百分比
				stakerRate := new(big.Float).Quo(
					new(big.Float).SetInt(stakerAmount).Mul(new(big.Float).SetInt(stakerAmount), big.NewFloat(100)),
					new(big.Float).SetInt(onlineTotalStakeAmount),
				)

				p.logger.WithFields(logrus.Fields{
					"stakerReward": stakeReward,
					"stakerRate%":  stakerRate.String(),
					"stakeAmount":  stakerAmount,
				}).Info("Rewarding")
			}

			if allIndexSum > 0 {
				v, ok := ableHistorys[addrString]
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

			p.state.AddBalance(addr, rewardAmount)
			peerRewards[addrString] = peerReward

			mintSet.PeerReward = peerRewards

		}
		// p.babble.Store.SetMinthistory(currentIndex, mintSet)
		jsMint := peers.NewJSONMint(p.babble.Config.DataDir)
		jsMint.SetMinthistoryForJson(currentIndex, mintSet)
		return mint
	}
	return mint
}
