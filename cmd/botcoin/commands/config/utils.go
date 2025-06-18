package config

import (
	"fmt"
	"path/filepath"

	"github.com/BOTCoinNetwork/Botcoin/src/common"
	"github.com/BOTCoinNetwork/Botcoin/src/configuration"
	"github.com/BOTCoinNetwork/Botcoin/src/files"
)

// CreateMonetConfigFolders creates the standard directory layout for a monet
// configuration folder
func CreateMonetConfigFolders(configDir string) error {
	return files.CreateDirsIfNotExists([]string{
		configDir,
		filepath.Join(configDir, configuration.BabbleDir),
		filepath.Join(configDir, configuration.EthDir),
		filepath.Join(configDir, configuration.EthDir, configuration.POADir),
	})
}

// ShowIPWarnings outputs warnings if IP addresses are local and propably not
// reachable from the outside.
func ShowIPWarnings() {
	api := configuration.Global.APIAddr
	listen := configuration.Global.Babble.BindAddr
	advertise := configuration.Global.Babble.AdvertiseAddr

	if common.CheckIP(api, true) {
		common.MessageWithType(common.MsgWarning, fmt.Sprintf("Botcoin service API address in botcoin.toml may be internal: %s", api))
	}

	if advertise != "" && common.CheckIP(advertise, false) {
		common.MessageWithType(common.MsgWarning, fmt.Sprintf("babble.advertise address in botcoin.toml may be internal: %s \n", listen))
	} else if common.CheckIP(listen, false) {
		common.MessageWithType(
			common.MsgWarning,
			fmt.Sprintf("babble.listen address in botcoin.toml may be internal: %s. Consider setting an advertise address.", listen),
		)
	}
}
