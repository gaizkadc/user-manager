/*
 * Copyright 2020 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package user

import (
	"context"
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-authx-go"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-role-go"
	"github.com/nalej/grpc-user-go"
	"github.com/nalej/grpc-user-manager-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/user-manager/internal/pkg/entities"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	accessClient grpc_authx_go.AuthxClient
	usersClient  grpc_user_go.UsersClient
	roleClient   grpc_role_go.RolesClient

	usersCache UsersCache
}

// NewManager creates a Manager using a set of clients.
func NewManager(
	accessClient grpc_authx_go.AuthxClient,
	usersClient grpc_user_go.UsersClient,
	roleClient grpc_role_go.RolesClient,
) Manager {
	return Manager{accessClient: accessClient, usersClient: usersClient, roleClient: roleClient,
		usersCache: NewUsersCache(accessClient, usersClient, roleClient)}
}

// AddUser adds a new user to an organization.
func (m *Manager) AddUser(addUserRequest *grpc_user_manager_go.AddUserRequest) (*grpc_user_manager_go.User, error) {

	// clear userCache
	m.usersCache.Clear(addUserRequest.OrganizationId)

	addRequest := &grpc_user_go.AddUserRequest{
		OrganizationId: addUserRequest.OrganizationId,
		Email:          addUserRequest.Email,
		Name:           addUserRequest.Name,
		PhotoBase64:    addUserRequest.PhotoBase64,
		LastName:       addUserRequest.LastName,
		Location:       addUserRequest.Location,
		Phone:          addUserRequest.Phone,
		Title:          addUserRequest.Title,
	}
	// 1. Add the user to system model
	user, err := m.usersClient.AddUser(context.Background(), addRequest)
	if err != nil {
		return nil, err
	}
	// 2. Register the credentials on authx
	addBasicCredentialsRequest := &grpc_authx_go.AddBasicCredentialRequest{
		OrganizationId: addUserRequest.OrganizationId,
		Username:       addUserRequest.Email,
		Password:       addUserRequest.Password,
		RoleId:         addUserRequest.RoleId,
	}
	_, err = m.accessClient.AddBasicCredentials(context.Background(), addBasicCredentialsRequest)
	if err != nil {
		return nil, err
	}
	userID := &grpc_user_go.UserId{
		OrganizationId: user.OrganizationId,
		Email:          user.Email,
	}
	return m.GetUser(userID)
}

// RemoveUser removes a given user from the system.
func (m *Manager) RemoveUser(userID *grpc_user_go.UserId) error {

	// check if the operation can be done
	canRemove, vErr := m.usersCache.CanRemoveUser(userID)
	if vErr != nil {
		return conversions.ToGRPCError(vErr)
	}
	if !canRemove {
		return derrors.NewInvalidArgumentError(fmt.Sprintf("can not remove user, last %d user in the system", grpc_authx_go.AccessPrimitive_ORG))
	}
	// clear userCache
	m.usersCache.Clear(userID.OrganizationId)

	// 1. Remove user from authx
	deleteCredentialsRequest := &grpc_authx_go.DeleteCredentialsRequest{
		Username: userID.Email,
	}
	_, err := m.accessClient.DeleteCredentials(context.Background(), deleteCredentialsRequest)
	if err != nil {
		return err
	}
	// 2. Remove user from system model
	removeUserRequest := &grpc_user_go.RemoveUserRequest{
		OrganizationId: userID.OrganizationId,
		Email:          userID.Email,
	}
	_, err = m.usersClient.RemoveUser(context.Background(), removeUserRequest)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) ListUsers(organizationID *grpc_organization_go.OrganizationId) (*grpc_user_manager_go.UserList, error) {
	users, err := m.usersClient.GetUsers(context.Background(), organizationID)
	if err != nil {
		return nil, err
	}
	result := make([]*grpc_user_manager_go.User, 0)
	for _, u := range users.Users {
		uID := &grpc_user_go.UserId{
			OrganizationId: u.OrganizationId,
			Email:          u.Email,
		}
		info, err := m.GetUser(uID)
		if err != nil {
			return nil, err
		}
		result = append(result, info)
	}
	return &grpc_user_manager_go.UserList{
		Users: result,
	}, nil
}

// ChangePassword updates the password of a user.
func (m *Manager) ChangePassword(request *grpc_user_manager_go.ChangePasswordRequest) error {
	authxRequest := entities.ToChangePasswordRequest(request)
	_, err := m.accessClient.ChangePassword(context.Background(), authxRequest)
	return err
}

// AddRole adds a new role to an organization.
func (m *Manager) AddRole(addRoleRequest *grpc_user_manager_go.AddRoleRequest) (*grpc_authx_go.Role, error) {

	// clear userCache
	m.usersCache.Clear(addRoleRequest.OrganizationId)

	// 1. Add the role to the organization in SM
	addRequest := &grpc_role_go.AddRoleRequest{
		OrganizationId: addRoleRequest.OrganizationId,
		Name:           addRoleRequest.Name,
		Description:    addRoleRequest.Description,
		Internal:       addRoleRequest.Internal,
	}
	role, err := m.roleClient.AddRole(context.Background(), addRequest)
	if err != nil {
		return nil, err
	}
	// 2. Add the role in Authx
	toAdd := &grpc_authx_go.Role{
		OrganizationId: role.OrganizationId,
		RoleId:         role.RoleId,
		Name:           role.Name,
		Internal:       role.Internal,
		Primitives:     addRoleRequest.Primitives,
	}
	_, err = m.accessClient.AddRole(context.Background(), toAdd)
	if err != nil {
		return nil, err
	}
	return toAdd, nil
}

// RemoveRole removes a role from an organization.
func (m *Manager) RemoveRole(roleID *grpc_authx_go.RoleId) error {
	// 1. Check if users with the role exists
	// 2. Remove role from SM
	// 3. Remove role from authx
	panic("RemoveRole")
}

// AssignRole assigns a role to an existing user.
func (m *Manager) AssignRole(assignRoleRequest *grpc_user_manager_go.AssignRoleRequest) (*grpc_user_manager_go.User, error) {

	canAssign, err := m.usersCache.CanAssignRole(assignRoleRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	if !canAssign {
		return nil, conversions.ToDerror(derrors.NewInvalidArgumentError(fmt.Sprintf("can not assign role, last %d user in the system", grpc_authx_go.AccessPrimitive_ORG)))
	}

	// clear userCache
	m.usersCache.Clear(assignRoleRequest.OrganizationId)

	// 1. Update on authx
	editRequest := &grpc_authx_go.EditUserRoleRequest{
		Username:  assignRoleRequest.Email,
		NewRoleId: assignRoleRequest.RoleId,
	}
	_, eErr := m.accessClient.EditUserRole(context.Background(), editRequest)
	if eErr != nil {
		return nil, eErr
	}
	userID := &grpc_user_go.UserId{
		OrganizationId: assignRoleRequest.OrganizationId,
		Email:          assignRoleRequest.Email,
	}
	return m.GetUser(userID)
}

// GetUser retrieves the information of a user including role information.
func (m *Manager) GetUser(userID *grpc_user_go.UserId) (*grpc_user_manager_go.User, error) {
	smUser, err := m.usersClient.GetUser(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	authxUserInfo, err := m.accessClient.GetUserAuthxInfo(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	role, err := m.roleClient.GetRole(context.Background(), &grpc_role_go.RoleId{
		OrganizationId: userID.OrganizationId,
		RoleId:         authxUserInfo.RoleId,
	})

	return &grpc_user_manager_go.User{
		OrganizationId: smUser.OrganizationId,
		Email:          smUser.Email,
		Name:           smUser.Name,
		PhotoBase64:    smUser.PhotoBase64,
		MemberSince:    smUser.MemberSince,
		RoleId:         authxUserInfo.RoleId,
		RoleName:       role.Name,
		InternalRole:   authxUserInfo.InternalRole,
		LastName:       smUser.LastName,
		Title:          smUser.Title,
		LastLogin:      authxUserInfo.LastLogin,
		Phone:          smUser.Phone,
		Location:       smUser.Location,
	}, nil
}

// ListRoles obtains a list of roles in an organization.
func (m *Manager) ListRoles(organizationID *grpc_organization_go.OrganizationId) (*grpc_authx_go.RoleList, error) {
	return m.accessClient.ListRoles(context.Background(), organizationID)
}

func (m *Manager) UpdateUser(updateUserRequest *grpc_user_go.UpdateUserRequest) (*grpc_common_go.Success, error) {
	return m.usersClient.Update(context.Background(), updateUserRequest)
}
