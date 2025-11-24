package server

import (
	"github.com/hashicorp/nomad/api"
)

func generateNomadClient(cfg *NomadConfig) (*api.Client, error) {

	nomadConfig := api.DefaultConfig()
	if cfg != nil {
		if cfg.Addr != "" {
			nomadConfig.Address = cfg.Addr
		}
		if cfg.Token != "" {
			nomadConfig.SecretID = cfg.Token
		}
	}

	return api.NewClient(nomadConfig)
}
