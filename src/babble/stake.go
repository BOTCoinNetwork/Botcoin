package babble

import (
	"github.com/BOTCoinNetwork/BVM/src/state"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

func (p *InmemProxy) getStakerArray() ([]common.Address, error) {
	// Pack the function call data
	callData, err := state.POAABI.Pack("getStakerArray")
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getStakerArray err")
		return nil, err
	}

	// Create ethereum message for contract call
	ethMsg := ethTypes.NewMessage(state.POAADDR,
		&state.POAADDR,
		uint64(1),
		big.NewInt(0),
		p.state.GetGasLimit(),
		big.NewInt(0),
		callData,
		false)

	// Execute the call
	res, err := p.state.Call(ethMsg)
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getStakerArray err")
		return nil, err
	}

	// Unpack the result
	var stakerArray []common.Address
	err = state.POAABI.Unpack(&stakerArray, "getStakerArray", res)
	if err != nil {
		return nil, err
	}

	return stakerArray, nil
}

func (p *InmemProxy) getStakeList(staker common.Address) (*big.Int, error) {
	callData, err := state.POAABI.Pack("getStakeList", staker)
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getStakeList err")
		return nil, err
	}

	ethMsg := ethTypes.NewMessage(state.POAADDR,
		&state.POAADDR,
		uint64(1),
		big.NewInt(0),
		p.state.GetGasLimit(),
		big.NewInt(0),
		callData,
		false)

	res, err := p.state.Call(ethMsg)
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getStakeList err")
		return nil, err
	}

	var result interface{}
	err = state.POAABI.Unpack(&result, "getStakeList", res)
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to getStakeList err")
		return nil, err
	}

	results, ok := result.([]interface{})
	if!ok {
	    p.logger.WithError(err).Errorf("Failed to convert result to []interface{}")
	    return nil, err
	}

	if len(results) < 2 {
	    p.logger.WithError(err).Errorf("Result does not contain enough elements")
	    return nil, err
	}

	rate, ok := results[1].(*big.Int)
	if!ok {
	    p.logger.WithError(err).Errorf("Failed to convert second element to *big.Int")
	    return nil, err
	}

	return rate, nil
}

func (p *InmemProxy) checkStake(staker common.Address) (*big.Int, error) {
	callData, err := state.POAABI.Pack("checkStake", staker)
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to checkStake pack err")
		return nil, err
	}

	ethMsg := ethTypes.NewMessage(state.POAADDR,
		&state.POAADDR,
		uint64(1),
		big.NewInt(0),
		p.state.GetGasLimit(),
		big.NewInt(0),
		callData,
		false)

	res, err := p.state.Call(ethMsg)
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to checkStake Call err")
		return nil, err
	}

	var result = new(big.Int)
	err = state.POAABI.Unpack(&result, "checkStake", res)
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to checkStake Unpack err")
		return nil, err
	}

	return result, nil
}

func (p *InmemProxy) getTotalStaked() (*big.Int, error) {
	// Pack the function call data
	callData, err := state.POAABI.Pack("gettotalStaked")
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to gettotalStaked err")
		return nil, err
	}

	// Create ethereum message for contract call
	ethMsg := ethTypes.NewMessage(state.POAADDR,
		&state.POAADDR,
		uint64(1),
		big.NewInt(0),
		p.state.GetGasLimit(),
		big.NewInt(0),
		callData,
		false)

	// Execute the call
	res, err := p.state.Call(ethMsg)
	if err != nil {
		p.logger.WithError(err).Errorf("Failed to gettotalStaked err")
		return nil, err
	}

	// Unpack the result
	var result = new(big.Int)
	err = state.POAABI.Unpack(&result, "gettotalStaked", res)
	if err != nil {
		return nil, err
	}

	return result, nil
}
