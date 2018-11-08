/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package user

import (
	"context"
	"github.com/nalej/grpc-authx-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-user-go"
	"github.com/nalej/grpc-user-manager-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/user-manager/internal/pkg/entities"
	"github.com/rs/zerolog/log"
)

// Handler structure for the user requests.
type Handler struct {
	Manager Manager
}

// NewHandler creates a new Handler with a linked manager.
func NewHandler(manager Manager) *Handler{
	return &Handler{manager}
}

// AddUser adds a new user to an organization.
func (h*Handler) AddUser(ctx context.Context, addUserRequest *grpc_user_manager_go.AddUserRequest) (*grpc_user_manager_go.User, error){
	log.Debug().Str("organizationID", addUserRequest.OrganizationId).Str("roleID", addUserRequest.RoleId).
		Str("email", addUserRequest.Email).Msg("add user")
	err := entities.ValidAddUserRequest(addUserRequest)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	user, aErr := h.Manager.AddUser(addUserRequest)
	if aErr != nil{
		return nil, aErr
	}
	log.Debug().Str("organizationID", addUserRequest.OrganizationId).
		Str("email", addUserRequest.Email).Msg("user has been added")
	return user, nil
}

// GetUser retrieves the information of a user including role information.
func (h*Handler) GetUser(ctx context.Context, userID * grpc_user_go.UserId) (*grpc_user_manager_go.User, error){
	err := entities.ValidUserID(userID)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.GetUser(userID)
}

// RemoveUser removes a given user from the system.
func (h*Handler) RemoveUser(ctx context.Context, userID *grpc_user_go.UserId) (*grpc_common_go.Success, error){
	log.Debug().Str("organizationID", userID.OrganizationId).Str("email", userID.Email).Msg("remove user")
	err := entities.ValidUserID(userID)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	rErr := h.Manager.RemoveUser(userID)
	if rErr != nil{
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_common_go.Success{}, nil
}

func (h*Handler) ListUsers(ctx context.Context, organizationID * grpc_organization_go.OrganizationId) (*grpc_user_manager_go.UserList, error){
	err := entities.ValidOrganizationID(organizationID)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.ListUsers(organizationID)
}

// ChangePassword updates the password of a user.
func (h*Handler) ChangePassword(ctx context.Context, request *grpc_authx_go.ChangePasswordRequest) (*grpc_common_go.Success, error){
	err := entities.ValidChangePasswordRequest(request)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	cErr := h.Manager.ChangePassword(request)
	if cErr != nil{
		return nil, err
	}
	return &grpc_common_go.Success{}, nil
}

// AddRole adds a new role to an organization.
func (h*Handler) AddRole(ctx context.Context, addRoleRequest *grpc_user_manager_go.AddRoleRequest) (*grpc_authx_go.Role, error){
	log.Debug().Str("organizationID", addRoleRequest.OrganizationId).Str("name", addRoleRequest.Name).Msg("add role")
	err := entities.ValidAddRoleRequest(addRoleRequest)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	role, aErr := h.Manager.AddRole(addRoleRequest)
	if aErr != nil{
		return nil, aErr
	}
	log.Debug().Str("organizationID", addRoleRequest.OrganizationId).Str("roleID", role.RoleId).Msg("role has been created")
	return role, nil
}

// RemoveRole removes a role from an organization.
func (h*Handler) RemoveRole(ctx context.Context, roleID *grpc_authx_go.RoleId) (*grpc_common_go.Success, error){
	err := entities.ValidRoleID(roleID)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	rErr := h.Manager.RemoveRole(roleID)
	if rErr != nil{
		return nil, rErr
	}
	return &grpc_common_go.Success{}, nil
}

// AssignRole assigns a role to an existing user.
func (h*Handler) AssignRole(ctx context.Context, assignRoleRequest *grpc_user_manager_go.AssignRoleRequest) (*grpc_user_manager_go.User, error){
	err := entities.ValidAssignRoleRequest(assignRoleRequest)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	user, aErr := h.Manager.AssignRole(assignRoleRequest)
	if aErr != nil{
		return nil, aErr
	}
	return user, nil
}

// ListRoles obtains a list of roles in an organization.
func (h*Handler) ListRoles(ctx context.Context, organizationID *grpc_organization_go.OrganizationId) (*grpc_authx_go.RoleList, error){
	err := entities.ValidOrganizationID(organizationID)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	roles, lErr := h.Manager.ListRoles(organizationID)
	if lErr != nil{
		return nil, lErr
	}
	return roles, nil
}