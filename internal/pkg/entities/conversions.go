package entities

import (
"github.com/nalej/grpc-authx-go"
"github.com/nalej/grpc-user-manager-go"
)

func ToChangePasswordRequest(source *grpc_user_manager_go.ChangePasswordRequest) *grpc_authx_go.ChangePasswordRequest {
	request := &grpc_authx_go.ChangePasswordRequest{
		Username:    source.Email,
		Password:    source.Password,
		NewPassword: source.NewPassword,
	}
	return request
}
