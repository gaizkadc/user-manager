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
