/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

/*
RUN_INTEGRATION_TEST=true
IT_SM_ADDRESS=192.168.99.100:31089
IT_AUTHX_ADDRESS=192.168.99.100:31810
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
	"github.com/nalej/grpc-utils/pkg/conversions"
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

func CreateOrganization(name string, orgClient grpc_organization_go.OrganizationsClient) *grpc_organization_go.Organization {
	toAdd := &grpc_organization_go.AddOrganizationRequest{
		Name: fmt.Sprintf("%s-%d-%d", name, ginkgo.GinkgoRandomSeed(), rand.Int()),
	}
	added, err := orgClient.AddOrganization(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(added).ToNot(gomega.BeNil())
	return added
}

func CreateRole(name string, organizationID string,
	roleClient grpc_role_go.RolesClient, accessClient grpc_authx_go.AuthxClient) *grpc_role_go.Role {
	toAdd := &grpc_role_go.AddRoleRequest{
		OrganizationId: organizationID,
		Name:           name,
		Description:    "user-manager-it",
	}
	added, err := roleClient.AddRole(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(added).ToNot(gomega.BeNil())

	accessRoleRequest := &grpc_authx_go.Role{
		OrganizationId: organizationID,
		RoleId:         added.RoleId,
		Name:           added.Name,
		Primitives:     []grpc_authx_go.AccessPrimitive{grpc_authx_go.AccessPrimitive_ORG},
	}
	_, err = accessClient.AddRole(context.Background(), accessRoleRequest)
	gomega.Expect(err).To(gomega.Succeed())
	return added
}

func CreateResourcesRole(name string, organizationID string,
	roleClient grpc_role_go.RolesClient, accessClient grpc_authx_go.AuthxClient) *grpc_role_go.Role {
	toAdd := &grpc_role_go.AddRoleRequest{
		OrganizationId: organizationID,
		Name:           name,
		Description:    "user-manager-it(resource)",
	}
	added, err := roleClient.AddRole(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(added).ToNot(gomega.BeNil())

	accessRoleRequest := &grpc_authx_go.Role{
		OrganizationId: organizationID,
		RoleId:         added.RoleId,
		Name:           added.Name,
		Primitives:     []grpc_authx_go.AccessPrimitive{grpc_authx_go.AccessPrimitive_RESOURCES},
	}
	_, err = accessClient.AddRole(context.Background(), accessRoleRequest)
	gomega.Expect(err).To(gomega.Succeed())
	return added
}

func GetRandomEmail() string {
	return fmt.Sprintf("random-%d@mail.com", rand.Int())
}

var _ = ginkgo.Describe("User service", func() {

	if !utils.RunIntegrationTests() {
		log.Warn().Msg("Integration tests are skipped")
		return
	}

	var (
		systemModelAddress = os.Getenv("IT_SM_ADDRESS")
		authxAddress       = os.Getenv("IT_AUTHX_ADDRESS")
	)

	if systemModelAddress == "" || authxAddress == "" {
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
	var smConn *grpc.ClientConn
	var authxClient grpc_authx_go.AuthxClient
	var authxConn *grpc.ClientConn

	// Target organization.
	var targetOrganization *grpc_organization_go.Organization
	var targetRole *grpc_role_go.Role

	ginkgo.BeforeSuite(func() {
		listener = test.GetDefaultListener()
		server = grpc.NewServer()

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
		test.LaunchServer(server, listener)

		conn, err := test.GetConn(*listener)
		gomega.Expect(err).Should(gomega.Succeed())
		client = grpc_user_manager_go.NewUserManagerClient(conn)
		rand.Seed(ginkgo.GinkgoRandomSeed())
		

	})

	ginkgo.AfterSuite(func() {
		server.Stop()
		listener.Close()
	})


	ginkgo.BeforeEach(func() {

		ginkgo.By("creating target entities", func() {
			// Initial data
			targetOrganization = CreateOrganization("app-manager-it", orgClient)
			targetRole = CreateRole("test", targetOrganization.OrganizationId, roleClient, authxClient)
		})
	})

	ginkgo.It("should be able to add a new user", func() {
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user",
			RoleId:         targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		if err != nil {
			log.Info().Str("trace", conversions.ToDerror(err).DebugReport()).Msg("error returned")
		}
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(added.Email).ShouldNot(gomega.BeEmpty())
		gomega.Expect(added.RoleId).ShouldNot(gomega.BeEmpty())
		gomega.Expect(added.RoleName).Should(gomega.Equal(targetRole.Name))
	})

	ginkgo.It("should be able to retrieve a user", func() {
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user",
			RoleId:         targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		userID := &grpc_user_go.UserId{
			OrganizationId: added.OrganizationId,
			Email:          added.Email,
		}
		retrieved, err := client.GetUser(context.Background(), userID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(retrieved.Email).Should(gomega.Equal(added.Email))
	})

	ginkgo.It("should be able to list the users in an organizationID", func() {
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user",
			RoleId:         targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		organizationID := &grpc_organization_go.OrganizationId{
			OrganizationId: targetOrganization.OrganizationId,
		}
		users, err := client.ListUsers(context.Background(), organizationID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(len(users.Users)).Should(gomega.Equal(1))
		gomega.Expect(users.Users[0].Email).Should(gomega.Equal(added.Email))
	})

	ginkgo.It("should be able to remove a user", func() {
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user",
			RoleId:         targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		userID := &grpc_user_go.UserId{
			OrganizationId: added.OrganizationId,
			Email:          added.Email,
		}
		success, err := client.RemoveUser(context.Background(), userID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(success).ShouldNot(gomega.BeNil())
	})

	ginkgo.It("should be able to change the password of a user", func() {
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user",
			RoleId:         targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		changeRequest := &grpc_user_manager_go.ChangePasswordRequest{
			Email:    added.Email,
			Password:    toAdd.Password,
			NewPassword: "newPassword",
			OrganizationId: targetOrganization.OrganizationId,
		}
		success, err := client.ChangePassword(context.Background(), changeRequest)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(success).ShouldNot(gomega.BeNil())
	})

	ginkgo.It("should not be able to change the password of a user", func() {
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user",
			RoleId:         targetRole.RoleId,
		}
		added, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		changeRequest := &grpc_user_manager_go.ChangePasswordRequest{
			Email:    added.Email,
			Password:    "WrongPassword",
			NewPassword: "newPassword",
			OrganizationId: targetOrganization.OrganizationId,
		}
		_, err = client.ChangePassword(context.Background(), changeRequest)
		gomega.Expect(err).NotTo(gomega.Succeed())
	})

	ginkgo.It("should be able to add a new role", func() {
		primitives := make([]grpc_authx_go.AccessPrimitive, 0)
		primitives = append(primitives, grpc_authx_go.AccessPrimitive_ORG)
		addRoleRequest := &grpc_user_manager_go.AddRoleRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Name:           "newRole",
			Description:    "newRole",
			Internal:       false,
			Primitives:     primitives,
		}
		added, err := client.AddRole(context.Background(), addRoleRequest)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(added.RoleId).ToNot(gomega.BeEmpty())
		gomega.Expect(added.Name).Should(gomega.Equal(addRoleRequest.Name))
	})

	ginkgo.It("should be able to assign a role to an existing user", func() {
		// Add the user
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user",
			RoleId:         targetRole.RoleId,
		}
		user, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())
		// Create the new role
		newRole := CreateRole("newTestRole", targetOrganization.OrganizationId, roleClient, authxClient)

		// Assign role
		assignRoleRequest := &grpc_user_manager_go.AssignRoleRequest{
			OrganizationId: user.OrganizationId,
			Email:          user.Email,
			RoleId:         newRole.RoleId,
		}
		retrieved, err := client.AssignRole(context.Background(), assignRoleRequest)
		// Check role
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(retrieved.Email).Should(gomega.Equal(user.Email))
		gomega.Expect(retrieved.RoleId).Should(gomega.Equal(newRole.RoleId))
	})
	ginkgo.It("should NOT be able to assign a role to an existing user (no more ORG users in the system)", func() {
		// Add the user
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user",
			RoleId:         targetRole.RoleId,
		}
		user, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())
		// Create the new role
		newRole := CreateResourcesRole("newResourceTestRole", targetOrganization.OrganizationId, roleClient, authxClient)

		// Assign role
		assignRoleRequest := &grpc_user_manager_go.AssignRoleRequest{
			OrganizationId: user.OrganizationId,
			Email:          user.Email,
			RoleId:         newRole.RoleId,
		}
		_, err = client.AssignRole(context.Background(), assignRoleRequest)
		// Check role
		gomega.Expect(err).NotTo(gomega.Succeed())
	})
	ginkgo.It("should be able to assign a role to an existing user (more ORG users in the system)", func() {
		// Add the user
		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user1",
			RoleId:         targetRole.RoleId,
		}
		user, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		// Create the new role and add new user
		newOrgRole := CreateRole("orgRole2", targetOrganization.OrganizationId, roleClient, authxClient)
		toAdd2 := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user2",
			RoleId:         newOrgRole.RoleId,
		}
		_, err = client.AddUser(context.Background(), toAdd2)
		gomega.Expect(err).To(gomega.Succeed())

		newRole := CreateResourcesRole("newResourceTestRole", targetOrganization.OrganizationId, roleClient, authxClient)

		// Assign role
		assignRoleRequest := &grpc_user_manager_go.AssignRoleRequest{
			OrganizationId: user.OrganizationId,
			Email:          user.Email,
			RoleId:         newRole.RoleId,
		}
		_, err = client.AssignRole(context.Background(), assignRoleRequest)
		// Check role
		gomega.Expect(err).To(gomega.Succeed())
	})
	ginkgo.It("should be able to assign a role to an existing user (it wasn't ORG user)", func() {
		// Add the user
		newRole := CreateResourcesRole("newResourceTestRole", targetOrganization.OrganizationId, roleClient, authxClient)

		toAdd := &grpc_user_manager_go.AddUserRequest{
			OrganizationId: targetOrganization.OrganizationId,
			Email:          GetRandomEmail(),
			Password:       "password",
			Name:           "user1",
			RoleId:         newRole.RoleId,
		}
		user, err := client.AddUser(context.Background(), toAdd)
		gomega.Expect(err).To(gomega.Succeed())

		// Create the new role and add new user
		newOrgRole := CreateRole("orgRole2", targetOrganization.OrganizationId, roleClient, authxClient)

		// Assign role
		assignRoleRequest := &grpc_user_manager_go.AssignRoleRequest{
			OrganizationId: user.OrganizationId,
			Email:          user.Email,
			RoleId:         newOrgRole.RoleId,
		}
		_, err = client.AssignRole(context.Background(), assignRoleRequest)
		// Check role
		gomega.Expect(err).To(gomega.Succeed())
	})

	ginkgo.It("should be able to list the roles in an organization", func() {
		organizationID := &grpc_organization_go.OrganizationId{
			OrganizationId: targetOrganization.OrganizationId,
		}
		roles, err := client.ListRoles(context.Background(), organizationID)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(len(roles.Roles)).Should(gomega.Equal(1))
	})

	ginkgo.Context("userCache tests", func() {
		ginkgo.It("should be able to add an organization Roles", func(){

			toAdd := &grpc_user_manager_go.AddUserRequest{
				OrganizationId: targetOrganization.OrganizationId,
				Email:          GetRandomEmail(),
				Password:       "password",
				Name:           "user",
				RoleId:         targetRole.RoleId,
			}
			added, err := client.AddUser(context.Background(), toAdd)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(added.Email).ShouldNot(gomega.BeEmpty())

			userCache := NewUsersCache(authxClient, userClient, roleClient)
			isOwner, err := userCache.roleIsOwner(targetOrganization.OrganizationId, targetRole.RoleId)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(isOwner).To(gomega.BeTrue())

			isOwner, err = userCache.roleIsOwner(targetOrganization.OrganizationId, "WrongID")
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(isOwner).NotTo(gomega.BeTrue())


		})
	})
})
