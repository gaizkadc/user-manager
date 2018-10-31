/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package server

import (
"github.com/nalej/derrors"
"github.com/rs/zerolog/log"
)

type Config struct {
	// Port where the gRPC API service will listen requests.
	Port int
	// AuthxAddress with the host:port to connect to the Authx manager.
	AuthxAddress string
	// SystemModelAddress with the host:port to connect to System Model
	SystemModelAddress string
}

func (conf * Config) Validate() derrors.Error {

	if conf.AuthxAddress == "" {
		return derrors.NewInvalidArgumentError("authxAddress must be set")
	}

	if conf.SystemModelAddress == "" {
		return derrors.NewInvalidArgumentError("systemModelAddress must be set")
	}

	return nil
}

func (conf *Config) Print() {
	log.Info().Int("port", conf.Port).Msg("gRPC port")
	log.Info().Str("URL", conf.AuthxAddress).Msg("Authx")
	log.Info().Str("URL", conf.SystemModelAddress).Msg("System Model")
}
