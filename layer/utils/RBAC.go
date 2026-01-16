package utils

import (
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Permission represents an action that can be performed
type Permission string

const (
	// Feature Flag Permissions
	PermissionCreateFeatureFlag Permission = "CREATE_FEATURE_FLAG"
	PermissionUpdateFeatureFlag Permission = "UPDATE_FEATURE_FLAG"
	PermissionReadFeatureFlag   Permission = "READ_FEATURE_FLAG"
	PermissionDeleteFeatureFlag Permission = "DELETE_FEATURE_FLAG"

	// User Feature Flag Mapping Permissions
	PermissionCreateUserMapping Permission = "CREATE_USER_MAPPING"
	PermissionUpdateUserMapping Permission = "UPDATE_USER_MAPPING"
	PermissionReadUserMapping   Permission = "READ_USER_MAPPING"

	// User Management Permissions
	PermissionReadUser   Permission = "READ_USER"
	PermissionUpdateUser Permission = "UPDATE_USER"
	PermissionDeleteUser Permission = "DELETE_USER"
)

// RolePermissions maps roles to their allowed permissions
var RolePermissions = map[string][]Permission{
	ROLE_ADMIN: {
		// Feature Flags - Full Access
		PermissionCreateFeatureFlag,
		PermissionUpdateFeatureFlag,
		PermissionReadFeatureFlag,
		PermissionDeleteFeatureFlag,
		// User Mappings - Full Access
		PermissionCreateUserMapping,
		PermissionUpdateUserMapping,
		PermissionReadUserMapping,
		// User Management - Full Access
		PermissionReadUser,
		PermissionUpdateUser,
		PermissionDeleteUser,
	},
	ROLE_DEVELOPER: {
		// Feature Flags - Create and Update
		PermissionCreateFeatureFlag,
		PermissionUpdateFeatureFlag,
		PermissionReadFeatureFlag,
		// User Mappings - Full Access
		PermissionCreateUserMapping,
		PermissionUpdateUserMapping,
		PermissionReadUserMapping,
		// User Management - Read Only
		PermissionReadUser,
	},
	ROLE_VIEWER: {
		// Feature Flags - Read Only
		PermissionReadFeatureFlag,
		// User Mappings - Read Only (own mappings)
		PermissionReadUserMapping,
		// User Management - Read Own Profile
		PermissionReadUser,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role string, permission Permission) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		log.Printf("Unknown role: %s", role)
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// RequirePermission is a middleware helper that checks if user has required permission
func RequirePermission(userContext *UserContext, permission Permission) (events.APIGatewayProxyResponse, error) {
	if userContext == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "User context not available",
		}, nil
	}

	if !HasPermission(userContext.Role, permission) {
		log.Printf("User %s with role %s does not have permission %s", userContext.UserId, userContext.Role, permission)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusForbidden,
			Body:       "Insufficient permissions",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
	}, nil
}

// RequireAnyPermission checks if user has any of the provided permissions
func RequireAnyPermission(userContext *UserContext, permissions ...Permission) (events.APIGatewayProxyResponse, error) {
	if userContext == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "User context not available",
		}, nil
	}

	for _, permission := range permissions {
		if HasPermission(userContext.Role, permission) {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
			}, nil
		}
	}

	log.Printf("User %s with role %s does not have any of the required permissions", userContext.UserId, userContext.Role)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusForbidden,
		Body:       "Insufficient permissions",
	}, nil
}

// CanAccessUserResource checks if user can access a resource belonging to another user
// Users can access their own resources, or if they're ADMIN
func CanAccessUserResource(userContext *UserContext, resourceUserId string) bool {
	if userContext == nil {
		return false
	}

	// Users can always access their own resources
	if userContext.UserId == resourceUserId {
		return true
	}

	// Only ADMIN can access other users' resources
	return userContext.Role == ROLE_ADMIN
}

