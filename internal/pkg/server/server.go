/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package server

import (
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-authx-go"
	"github.com/nalej/grpc-role-go"
	"github.com/nalej/grpc-user-go"
	"github.com/nalej/grpc-user-manager-go"
	"github.com/nalej/user-manager/internal/pkg/server/user"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

// Service structure with the configuration and the gRPC server.
type Service struct {
	Configuration Config

}

// NewService creates a new system model service.
func NewService(conf Config) *Service {
	return &Service{
		conf,
	}
}

// Clients structure with the gRPC clients for remote services.
type Clients struct {
	AuthxClient grpc_authx_go.AuthxClient
	UsersClient grpc_user_go.UsersClient
	RolesClient grpc_role_go.RolesClient
}

// GetClients creates the required connections with the remote clients.
func (s * Service) GetClients() (* Clients, derrors.Error) {
	authxConn, err := grpc.Dial(s.Configuration.AuthxAddress, grpc.WithInsecure())
	if err != nil{
		return nil, derrors.AsError(err, "cannot create connection with the authx component")
	}

	smConn, err := grpc.Dial(s.Configuration.SystemModelAddress, grpc.WithInsecure())
	if err != nil{
		return nil, derrors.AsError(err, "cannot create connection with the system model component")
	}

	aClient := grpc_authx_go.NewAuthxClient(authxConn)
	uClient := grpc_user_go.NewUsersClient(smConn)
	rClient := grpc_role_go.NewRolesClient(smConn)

	return &Clients{aClient, uClient, rClient}, nil
}

// Run the service, launch the REST service handler.
func (s *Service) Run() error {
	cErr := s.Configuration.Validate()
	if cErr != nil{
		log.Fatal().Str("err", cErr.DebugReport()).Msg("invalid configuration")
	}
	s.Configuration.Print()
	clients, cErr := s.GetClients()
	if cErr != nil{
		log.Fatal().Str("err", cErr.DebugReport()).Msg("Cannot create clients")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	// Create handlers
	manager := user.NewManager(clients.AuthxClient, clients.UsersClient, clients.RolesClient)
	handler := user.NewHandler(manager)

	grpcServer := grpc.NewServer()

	grpc_user_manager_go.RegisterUserManagerServer(grpcServer, handler)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
	return nil
}