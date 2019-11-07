/*
 * Copyright 2019 Nalej
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
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-role-go"
	"github.com/nalej/grpc-user-go"
	"github.com/nalej/grpc-user-manager-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
)

type UsersCache struct {
	// all owner roles indexed by organization_id
	ownerRoleIds map[string][]string
	// all owner users - indexed by organization_id
	ownerUsers map[string][]string

	accessClient grpc_authx_go.AuthxClient
	usersClient  grpc_user_go.UsersClient
	roleClient   grpc_role_go.RolesClient
}

func NewUsersCache(accessClient grpc_authx_go.AuthxClient, usersClient grpc_user_go.UsersClient,
	roleClient grpc_role_go.RolesClient) UsersCache {
	rolesIds := make(map[string][]string, 0)
	users := make(map[string][]string, 0)
	return UsersCache{accessClient: accessClient,
		usersClient:  usersClient,
		roleClient:   roleClient,
		ownerRoleIds: rolesIds,
		ownerUsers:   users}
}

func (uc *UsersCache) Clear(organizationID string) derrors.Error {
	// clear ownerRoleIds
	delete(uc.ownerRoleIds, organizationID)
	// Clear users-roles
	delete(uc.ownerUsers, organizationID)
	return nil
}

// check if thr removeUser operaton can be done
// 1.- If the user Role is not ORG -> the operation can be done
// 2.- If the user Role is ORG:
// 2.1.- If there are more ORG users -> the operation can be done
// 2.2.- If there are not more ORG users -> the operation cannot be done
func (uc *UsersCache) CanRemoveUser(userID *grpc_user_go.UserId) (bool, derrors.Error) {

	isOwner, err := uc.userIsOwner(userID.OrganizationId, userID.Email)
	if err != nil {
		return false, err
	}
	if isOwner {
		hasMoreOwner, err := uc.hasMoreOwner(userID.OrganizationId, userID.Email)
		if err != nil {
			return false, err
		}
		if !hasMoreOwner {
			return false, nil
		}
	}
	return true, nil

}

// check if the assignRole operation can be done.
// If new Role is ORG -> nothing to check
// If new Role is not ORG:
// If old Role was no ORG -> noting to check
// It old Role was ORG:
// If there are more ORG users -> pass the validation
// It there aren't more ORG users -> not pass the validation
func (uc *UsersCache) CanAssignRole(assignRoleRequest *grpc_user_manager_go.AssignRoleRequest) (bool, derrors.Error) {

	// 1.- If newRole != ORG and oldRole == ORG -> check if the change can be made
	isOwner, err := uc.roleIsOwner(assignRoleRequest.OrganizationId, assignRoleRequest.RoleId)
	if err != nil {
		return false, err
	}

	if !isOwner {
		wasOwnerBefore, err := uc.userIsOwner(assignRoleRequest.OrganizationId, assignRoleRequest.Email)
		if err != nil {
			return false, err
		}
		if wasOwnerBefore {
			hasMoreOwner, err := uc.hasMoreOwner(assignRoleRequest.OrganizationId, assignRoleRequest.Email)
			if err != nil {
				return false, err
			}
			if !hasMoreOwner {
				return false, nil
			}
		}
	}

	return true, nil

}

// Add load into the cache the users and roles in the organizationID
func (uc *UsersCache) add(organizationID string) derrors.Error {

	// ---------------
	// Owner Roles Ids
	// ---------------
	roles, err := uc.accessClient.ListRoles(context.Background(), &grpc_organization_go.OrganizationId{
		OrganizationId: organizationID,
	})
	if err != nil {
		return conversions.ToDerror(err)
	}
	roleIds := make([]string, 0)
	for _, rol := range roles.Roles {
		for _, primitive := range rol.Primitives {
			if primitive == grpc_authx_go.AccessPrimitive_ORG {
				roleIds = append(roleIds, rol.RoleId)
			}
		}
	}
	if len(roleIds) > 0 {
		uc.ownerRoleIds[organizationID] = roleIds
	}
	// ---------
	// userRoles
	// ---------
	// get all the users in the organization
	organizationUsers, err := uc.usersClient.GetUsers(context.Background(), &grpc_organization_go.OrganizationId{
		OrganizationId: organizationID,
	})
	if err != nil {
		return conversions.ToDerror(err)
	}

	userEmails := make([]string, 0)
	for _, user := range organizationUsers.Users {
		credentials, err := uc.accessClient.GetUserRole(context.Background(), &grpc_user_go.UserId{
			OrganizationId: user.OrganizationId,
			Email:          user.Email,
		})
		if err != nil {
			return conversions.ToDerror(err)
		}
		isOwner, err := uc.roleIsOwner(credentials.OrganizationId, credentials.RoleId)
		if err != nil {
			return conversions.ToDerror(err)
		}
		if isOwner {
			userEmails = append(userEmails, user.Email)
		}
	}
	if len(userEmails) > 0 {
		uc.ownerUsers[organizationID] = userEmails
	}

	return nil
}

// IsOwner checks if UserEmail allows to role with 'ORG' primitive
func (uc *UsersCache) userIsOwner(organizationID string, email string) (bool, derrors.Error) {

	users, exists := uc.ownerUsers[organizationID]
	if !exists {
		err := uc.add(organizationID)
		if err != nil {
			return false, err
		}

		users, exists = uc.ownerUsers[organizationID]
		if !exists {
			return false, derrors.NewInvalidArgumentError(fmt.Sprintf("no %d users found in the system", grpc_authx_go.AccessPrimitive_ORG))
		}
	}

	for _, user := range users {
		if user == email {
			return true, nil
		}
	}

	return false, nil
}

// IsOwner checks if roleID allows to role with 'ORG' primitive
func (uc *UsersCache) roleIsOwner(organizationID string, roleID string) (bool, derrors.Error) {

	roles, exists := uc.ownerRoleIds[organizationID]
	if !exists {
		err := uc.add(organizationID)
		if err != nil {
			return false, err
		}

		roles, exists = uc.ownerRoleIds[organizationID]
		if !exists {
			return false, derrors.NewInvalidArgumentError(fmt.Sprintf("no %d roles found in the system", grpc_authx_go.AccessPrimitive_ORG))
		}
	}

	for _, role := range roles {
		if role == roleID {
			return true, nil
		}
	}

	return false, nil
}

// HasMoreOwner checks if exists another user with an owner role
func (uc *UsersCache) hasMoreOwner(organizationID string, email string) (bool, derrors.Error) {

	users, exists := uc.ownerUsers[organizationID]
	if !exists {
		err := uc.add(organizationID)
		if err != nil {
			return false, err
		}
		users, exists = uc.ownerUsers[organizationID]
		if !exists {
			return false, derrors.NewInvalidArgumentError(fmt.Sprintf("cannot chech if there is more %d users in the system. No users found ", grpc_authx_go.AccessPrimitive_ORG))
		}
	}

	for _, user := range users {
		if user != email {
			return true, nil
		}
	}

	return false, nil
}
