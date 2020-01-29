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

package entities

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-authx-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-user-go"
	"github.com/nalej/grpc-user-manager-go"
	"regexp"
)

const (
	emptyOrganizationId = "organization_id cannot be empty"
	emptyEmail          = "email cannot be empty"
	emptyName           = "name cannot be empty"
	emptyRoleID         = "role_id cannot be empty"
	emptyPassword       = "password cannot be empty"
	invalidEmail        = "invalid email"
)

func ValidOrganizationID(organizationID *grpc_organization_go.OrganizationId) derrors.Error {
	if organizationID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	return nil
}

func ValidUserID(userID *grpc_user_go.UserId) derrors.Error {
	if userID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if userID.Email == "" {
		return derrors.NewInvalidArgumentError(emptyEmail)
	}
	return nil
}

func ValidAddRoleRequest(addRoleRequest *grpc_user_manager_go.AddRoleRequest) derrors.Error {
	if addRoleRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if addRoleRequest.Name == "" {
		return derrors.NewInvalidArgumentError(emptyName)
	}
	if len(addRoleRequest.Primitives) == 0 {
		return derrors.NewInvalidArgumentError("at least one primitive is expected")
	}
	return nil
}

func ValidRoleID(roleID *grpc_authx_go.RoleId) derrors.Error {
	if roleID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if roleID.RoleId == "" {
		return derrors.NewInvalidArgumentError(emptyRoleID)
	}
	return nil
}

func ValidAssignRoleRequest(assignRoleRequest *grpc_user_manager_go.AssignRoleRequest) derrors.Error {
	if assignRoleRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if assignRoleRequest.Email == "" {
		return derrors.NewInvalidArgumentError(emptyEmail)
	}
	if assignRoleRequest.RoleId == "" {
		return derrors.NewInvalidArgumentError(emptyRoleID)
	}
	return nil
}

func ValidAddUserRequest(addUserRequest *grpc_user_manager_go.AddUserRequest) derrors.Error {
	if addUserRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if addUserRequest.Email == "" {
		return derrors.NewInvalidArgumentError(emptyEmail)
	}
	var rxEmail = regexp.MustCompile(`^([a-zA-Z0-9_\-.]+)@([a-zA-Z0-9_\-.]+)\.([a-zA-Z]{2,30})$`)
	if len(addUserRequest.Email) > 254 || !rxEmail.MatchString(addUserRequest.Email) {
		return derrors.NewInvalidArgumentError(invalidEmail)
	}
	if addUserRequest.Password == "" {
		return derrors.NewInvalidArgumentError(emptyPassword)
	}
	if addUserRequest.Name == "" {
		return derrors.NewInvalidArgumentError(emptyName)
	}
	if addUserRequest.LastName == "" {
		return derrors.NewInvalidArgumentError(emptyName)
	}
	if addUserRequest.Title == "" {
		return derrors.NewInvalidArgumentError(emptyName)
	}
	if addUserRequest.RoleId == "" {
		return derrors.NewInvalidArgumentError(emptyRoleID)
	}
	return nil
}

func ValidChangePasswordRequest(request *grpc_user_manager_go.ChangePasswordRequest) derrors.Error {
	if request.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if request.Email == "" {
		return derrors.NewInvalidArgumentError(emptyEmail)
	}
	if request.Password == "" {
		return derrors.NewInvalidArgumentError("password cannot be empty")
	}
	if request.NewPassword == "" {
		return derrors.NewInvalidArgumentError("new_password cannot be empty")
	}
	return nil
}

func ValidUpdateUserRequest(updateUserRequest *grpc_user_go.UpdateUserRequest) derrors.Error {
	if updateUserRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if updateUserRequest.Email == "" {
		return derrors.NewInvalidArgumentError(emptyEmail)
	}
	if updateUserRequest.UpdateName && updateUserRequest.Name == "" {
		return derrors.NewInvalidArgumentError(emptyName)
	}
	if updateUserRequest.UpdateLastName && updateUserRequest.LastName == "" {
		return derrors.NewInvalidArgumentError(emptyName)
	}
	if updateUserRequest.UpdateTitle && updateUserRequest.Title == "" {
		return derrors.NewInvalidArgumentError(emptyName)
	}
	return nil
}
