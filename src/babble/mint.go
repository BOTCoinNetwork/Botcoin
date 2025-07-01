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

func (p *InmemProxy) validatorsForAddress(block hashgraph.Block, validators []*peers.Peer) []common.Address {

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

func (p *InmemProxy) rewardValidators(block hashgraph.Block, validators []*peers.Peer, currentRewardPool *big.Int) (rewardData, error) {
	if len(validators) == 0 {
		p.logger.WithFields(logrus.Fields{
			"validators": validators,
		}).Info("validators")
		return nil, nil
	}

	// totalRewardPool := p.getTotalRewardPool(block)

	currentReward := new(big.Int).Div(new(big.Int).Mul(currentRewardPool, big.NewInt(int64(p.rewardRule.validatorRate))), big.NewInt(100))

	rewardData := map[string]*big.Int{
		"verifyCurrentReward": currentReward,
	}

	return rewardData, nil
}

func (p *InmemProxy) rewardStakers(currentRewardPool *big.Int) (rewardData, map[string]peers.PeerReward, error) {

	peerRewards := make(map[string]peers.PeerReward)
	stakerArray, err := p.getStakerArray()
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getStakerArray err")
		return nil, peerRewards, err
	}

	if len(stakerArray) == 0 {
		return nil, peerRewards, nil
	}

	currentReward := new(big.Int).Div(new(big.Int).Mul(currentRewardPool, big.NewInt(int64(p.rewardRule.stakerRate))), big.NewInt(100))
	totalStakeAmount, err := p.getTotalStaked()

	// sum online's
	for _, vAddress := range stakerArray {

		stakerAmount, err := p.checkStake(vAddress)
		if err == nil {
			stakeReward := new(big.Int).Div(new(big.Int).Mul(currentReward, stakerAmount), totalStakeAmount)

			peerRewards[vAddress.Hex()] = peers.PeerReward{
				StakeAmount: stakerAmount,
				StakeReward: stakeReward,
			}
		}
	}

	rewardData := map[string]*big.Int{
		"stakerCurrentReward": currentReward,
		"totalStakeAmount":    totalStakeAmount,
	}

	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getTotalStaked err")
		return rewardData, peerRewards, err
	}

	return rewardData, peerRewards, nil
}

type stablePeer struct {
	addr  common.Address
	index int
}

func (p *InmemProxy) rewardStablePeer(block hashgraph.Block, validators []*peers.Peer, currentRewardPool *big.Int) (rewardData, error) {

	if len(validators) == 0 {
		p.logger.WithFields(logrus.Fields{
			"validators": validators,
		}).Info("validators")
		return nil, nil
	}

	// currentRewardPool := p.getTotalRewardPool(block)

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

	// Check if it's time to distribute rewards
	if p.checkRound(block) {
		return peers.Mint{}
	}

	totalCurrentReward := p.getTotalRewardPool(block)
	if totalCurrentReward.Cmp(big.NewInt(0)) <= 0 {
		return peers.Mint{}
	}
	p.logger.WithFields(logrus.Fields{
		"totalRewardPool": totalCurrentReward,
	}).Info("Total_rewardData")

	rewardData_Validators, err := p.rewardValidators(block, validators, totalCurrentReward)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward validators")
	}

	var reward_Validators = new(big.Int)
	var reward_Stake = new(big.Int)
	var reward_History = new(big.Int)
	var totalStakeAmount = new(big.Int)

	vAddress := p.validatorsForAddress(block, validators)

	// Reward for Stake
	rewardData_Stake, peerRewards, err := p.rewardStakers(totalCurrentReward)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward stakers")
	}

	if rewardData_Stake != nil {
		reward_Stake = rewardData_Stake["stakerCurrentReward"]
		totalStakeAmount = rewardData_Stake["totalStakeAmount"]

		p.logger.WithFields(logrus.Fields{
			"staker_CurrentReward": reward_Stake,
			"totalStakeAmount":     totalStakeAmount,
		}).Info("Total_rewardData")
	}

	// Reward for anyone Validators
	if rewardData_Validators != nil {
		reward_Validators = rewardData_Validators["verifyCurrentReward"]

		p.logger.WithFields(logrus.Fields{
			"verify_CurrentReward": reward_Validators,
		}).Info("Total_rewardData")
	}

	// Reward for some validators, With History
	rewardData_StablePeers, err := p.rewardStablePeer(block, validators, totalCurrentReward)
	if err != nil {
		p.logger.WithError(err).Error("Failed to reward stable peers")
	}

	if rewardData_StablePeers != nil {
		reward_History = rewardData_StablePeers["stablePeersCurrentReward"]
		p.logger.WithFields(logrus.Fields{
			"history_CurrentReward": reward_History,
		}).Info("Total_rewardData")
	}

	mint := peers.Mint{
		MintRewards:      totalCurrentReward,
		TotalStakeAmount: totalStakeAmount,
		PeersCount:       len(validators),
	}

	var mintSet = peers.MintSet{}
	mintSet.Mint = mint

	allIndexSum, rewardPeersMap := p.getAllIndexSum(block, validators)
	currentIndex := block.Index()

	for _, addr := range vAddress {

		addrString := addr.String()
		peerReward := peerRewards[addrString]
		if peerReward.StakeAmount == nil {
			peerReward = peers.PeerReward{}
		}

		this_Validator_RewardAmount := new(big.Int).Div(reward_Validators, big.NewInt(int64(len(validators))))
		peerReward.VerifyReward = this_Validator_RewardAmount

		if allIndexSum > 0 {
			v, ok := rewardPeersMap[addrString]
			if ok {
				tihs_History_RewardAmount := new(big.Int).Div(new(big.Int).Mul(reward_History, big.NewInt(int64(currentIndex-v.index))), big.NewInt(int64(allIndexSum)))
				peerReward.HistoryReward = tihs_History_RewardAmount
			}
		}

		peerRewards[addrString] = peerReward

	}
	// All peerRewards
	mintSet.PeerReward = peerRewards

	// Reward ing
	for addr, peerReward := range peerRewards {

		tishaddr_Can_Reward := new(big.Int)

		// Validator Reward
		if peerReward.VerifyReward != nil && peerReward.VerifyReward.Cmp(big.NewInt(0)) > 0 {
			tishaddr_Can_Reward = tishaddr_Can_Reward.Add(tishaddr_Can_Reward, peerReward.VerifyReward)

			p.logger.WithFields(logrus.Fields{
				"verifyReward": peerReward.VerifyReward,
				"verifyAddr":   addr,
			}).Info("Rewarding")
		}

		// History Reward
		if allIndexSum > 0 && peerReward.HistoryReward != nil && peerReward.HistoryReward.Cmp(big.NewInt(0)) > 0 {
			v, ok := rewardPeersMap[addr]
			if ok {
				tishaddr_Can_Reward = tishaddr_Can_Reward.Add(tishaddr_Can_Reward, peerReward.HistoryReward)

				p.logger.WithFields(logrus.Fields{
					"historyReward": peerReward.HistoryReward,
					"joinIndex":     v.index,
					"currentIndex":  currentIndex,
				}).Info("Rewarding")
			}
		}

		// Stake Reward
		if peerReward.StakeReward != nil && peerReward.StakeReward.Cmp(big.NewInt(0)) > 0 {

			tishaddr_Can_Reward = tishaddr_Can_Reward.Add(tishaddr_Can_Reward, peerReward.StakeReward)

			//  stakerAmount on totalStakeAmount percent
			stakerRate := new(big.Float).Quo(
				new(big.Float).SetInt(peerReward.StakeAmount).Mul(new(big.Float).SetInt(peerReward.StakeAmount), big.NewFloat(100)),
				new(big.Float).SetInt(totalStakeAmount),
			)

			p.logger.WithFields(logrus.Fields{
				"stakerReward": peerReward.StakeReward,
				"stakerRate":   stakerRate.String() + "%",
				"stakerAmount": peerReward.StakeAmount,
			}).Info("Rewarding")
		}

		Address := common.HexToAddress(addr)
		p.state.AddBalance(Address, tishaddr_Can_Reward)
	}

	// p.babble.Store.SetMinthistory(currentIndex, mintSet)
	jsMint := peers.NewJSONMint(p.babble.Config.DataDir)
	jsMint.SetMinthistoryForJson(currentIndex, mintSet)

	return mint
}
