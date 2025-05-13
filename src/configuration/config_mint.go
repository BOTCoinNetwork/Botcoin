package configuration

var (
	DefaultValidatorRewardRound  = 7200
	DefaultValidatorRewardPool   = 5000000
	DefaultValidatorHalvingRound = 15000000
	DefaultValidatorRate         = 60
	DefaultStakerRate            = 35
	DefaultStableRate            = 5
)

// MintConfig
type MintConfig struct {
	// Validator reward round
	ValidatorRewardRound int `mapstructure:"validator-reward-round"`

	// Validator reward pool
	ValidatorRewardPool int `mapstructure:"validator-reward-pool"`

	// Validator halving round
	ValidatorHalvingRound int `mapstructure:"validator-halving-round"`

	// validator rate
	ValidatorRate int `mapstructure:"validator-rate"`

	// staker rate
	StakerRate int `mapstructure:"staker-rate"`

	// history rate
	StableRate int `mapstructure:"stable-rate"`
}

// DefaultEthConfig return the default configuration for Mint services
func DefaultMintConfig() *MintConfig {
	return &MintConfig{
		ValidatorRewardPool:   DefaultValidatorRewardPool,
		ValidatorRewardRound:  DefaultValidatorRewardRound,
		ValidatorHalvingRound: DefaultValidatorHalvingRound,
		ValidatorRate:         DefaultValidatorRate,
		StakerRate:            DefaultStakerRate,
		StableRate:            DefaultStableRate,
	}
}
