/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package user

import (
	"context"
	"fmt"
	"github.com/nalej/grpc-authx-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-role-go"
	"github.com/nalej/grpc-user-go"
	"github.com/nalej/grpc-user-manager-go"
	"github.com/nalej/grpc-utils/pkg/test"
	"github.com/nalej/user-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"math/rand"
	"os"
)


func CreateOrganization(name string, orgClient grpc_organization_go.OrganizationsClient) * grpc_organization_go.Organization {
	toAdd := &grpc_organization_go.AddOrganizationRequest{
		Name:                 fmt.Sprintf("%s-%d-%d", name, ginkgo.GinkgoRandomSeed(), rand.Int()),
	}
	added, err := orgClient.AddOrganization(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(added).ToNot(gomega.BeNil())
	return added
}

func CreateRole(name string, organizationID string, roleClient grpc_role_go.RolesClient) * grpc_role_go.Role {
	toAdd := &grpc_role_go.AddRoleRequest{
		OrganizationId:       organizationID,
		Name:                 name,
		Description:          "user-manager-it",
	}
	added, err := roleClient.AddRole(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(added).ToNot(gomega.BeNil())
	return added
}


var _ = ginkgo.Describe("User service", func() {

	if ! utils.RunIntegrationTests() {
		log.Warn().Msg("Integration tests are skipped")
		return
	}

	var (
		systemModelAddress = os.Getenv("IT_SM_ADDRESS")
		authxAddress = os.Getenv("IT_AUTHX_ADDRESS")
	)

	if systemModelAddress == "" || authxAddress == ""{
		ginkgo.Fail("missing environment variables")
	}

	// gRPC server
	var server *grpc.Server
	// grpc test listener
	var listener *bufconn.Listener
	// client
	var client grpc_user_manager_go.UserManagerClient

	// Providers
	var orgClient grpc_organization_go.OrganizationsClient
	var userClient grpc_user_go.UsersClient
	var roleClient grpc_role_go.RolesClient
	var smConn * grpc.ClientConn
	var authxClient grpc_authx_go.AuthxClient
	var authxConn * grpc.ClientConn

	// Target organization.
	var targetOrganization *grpc_organization_go.Organization
	var targetRole * grpc_role_go.Role

	ginkgo.BeforeSuite(func() {
		listener = test.GetDefaultListener()
		server = grpc.NewServer()
		test.LaunchServer(server, listener)

		smConn = utils.GetConnection(systemModelAddress)
		userClient = grpc_user_go.NewUsersClient(smConn)
		roleClient = grpc_role_go.NewRolesClient(smConn)
		orgClient = grpc_organization_go.NewOrganizationsClient(smConn)

		authxConn = utils.GetConnection(authxAddress)
		authxClient = grpc_authx_go.NewAuthxClient(authxConn)

		// Register the service
		manager := NewManager(authxClient, userClient, roleClient)
		handler := NewHandler(manager)
		grpc_user_manager_go.RegisterUserManagerServer(server, handler)

		conn, err := test.GetConn(*listener)
		gomega.Expect(err).Should(gomega.Succeed())
		client = grpc_user_manager_go.NewUserManagerClient(conn)
	})

	ginkgo.AfterSuite(func() {
		server.Stop()
		listener.Close()
	})

	ginkgo.BeforeEach(func(){
		ginkgo.By("creating target entities", func(){
			// Initial data
			targetOrganization = CreateOrganization("app-manager-it", orgClient)
			targetRole = CreateRole("test", targetOrganization.OrganizationId, roleClient)
		})
	})

	ginkgo.It("should be able to add a new user", func(){
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId:       targetOrganization.OrganizationId,
			Email:                "random@email.com",
			Password:             "password",
			Name:                 "user",
			RoleId:               targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(added.Email).ShouldNot(gomega.BeEmpty())
		gomega.Expect(added.RoleId).ShouldNot(gomega.BeEmpty())
		gomega.Expect(added.RoleName).Should(gomega.Equal(targetRole.Name))
	})

	ginkgo.It("should be able to retrieve a user", func(){
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId:       targetOrganization.OrganizationId,
			Email:                "random@email.com",
			Password:             "password",
			Name:                 "user",
			RoleId:               targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		userID := &grpc_user_go.UserId{
			OrganizationId:       added.OrganizationId,
			Email:                added.Email,
		}
		retrieved, err := client.GetUser(context.Background(), userID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(retrieved.Email).Should(gomega.Equal(added.Email))
	})

	ginkgo.It("should be able to remove a user", func(){
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId:       targetOrganization.OrganizationId,
			Email:                "random@email.com",
			Password:             "password",
			Name:                 "user",
			RoleId:               targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		userID := &grpc_user_go.UserId{
			OrganizationId:       added.OrganizationId,
			Email:                added.Email,
		}
		success, err := client.RemoveUser(context.Background(), userID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(success).ShouldNot(gomega.BeNil())
	})

	ginkgo.It("should be able to change the password of a user", func(){
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId:       targetOrganization.OrganizationId,
			Email:                "random@email.com",
			Password:             "password",
			Name:                 "user",
			RoleId:               targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		changeRequest := &grpc_authx_go.ChangePasswordRequest{
			Username:             added.Email,
			Password:             toAdd.Password,
			NewPassword:          "newPassword",
		}
		success, err := client.ChangePassword(context.Background(), changeRequest)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(success).ShouldNot(gomega.BeNil())
	})

	ginkgo.It("should be able to add a new role", func(){
		primitives := make([] grpc_authx_go.AccessPrimitive, 0)
		primitives = append(primitives, grpc_authx_go.AccessPrimitive_ORG)
		addRoleRequest := &grpc_user_manager_go.AddRoleRequest{
			OrganizationId:       targetOrganization.OrganizationId,
			Name:                 "newRole",
			Description:          "newRole",
			Primitives:           primitives,
		}
		added, err := client.AddRole(context.Background(), addRoleRequest)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(added.RoleId).ToNot(gomega.BeEmpty())
		gomega.Expect(added.Name).Should(gomega.Equal(addRoleRequest.Name))
	})

	ginkgo.It("should be able to assign a role to an existing user", func(){
		// Add the user
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId:       targetOrganization.OrganizationId,
			Email:                "random@email.com",
			Password:             "password",
			Name:                 "user",
			RoleId:               targetRole.RoleId,
		}
		user, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())
		// Create the new role
		newRole := CreateRole("newTestRole", targetOrganization.OrganizationId, roleClient)

		// Assign role
		assignRoleRequest := &grpc_user_manager_go.AssignRoleRequest{
			OrganizationId:       user.OrganizationId,
			Email:                user.Email,
			RoleId:               newRole.RoleId,
		}
		retrieved, err := client.AssignRole(context.Background(), assignRoleRequest)
		// Check role
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(retrieved.Email).Should(gomega.Equal(user.Email))
		gomega.Expect(retrieved.RoleId).Should(gomega.Equal(newRole.RoleId))
	})

	ginkgo.It("should be able to list the roles in an organization", func(){
		organizationID := &grpc_organization_go.OrganizationId{
			OrganizationId:       targetOrganization.OrganizationId,
		}
		roles, err := client.ListRoles(context.Background(), organizationID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(len(roles.Roles)).Should(gomega.Equal(1))
	})


})