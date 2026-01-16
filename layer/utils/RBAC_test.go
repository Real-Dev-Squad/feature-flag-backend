package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		permission Permission
		expected   bool
	}{
		{
			name:       "ADMIN has CREATE_FEATURE_FLAG permission",
			role:       ROLE_ADMIN,
			permission: PermissionCreateFeatureFlag,
			expected:   true,
		},
		{
			name:       "DEVELOPER has CREATE_FEATURE_FLAG permission",
			role:       ROLE_DEVELOPER,
			permission: PermissionCreateFeatureFlag,
			expected:   true,
		},
		{
			name:       "VIEWER does not have CREATE_FEATURE_FLAG permission",
			role:       ROLE_VIEWER,
			permission: PermissionCreateFeatureFlag,
			expected:   false,
		},
		{
			name:       "ADMIN has DELETE_FEATURE_FLAG permission",
			role:       ROLE_ADMIN,
			permission: PermissionDeleteFeatureFlag,
			expected:   true,
		},
		{
			name:       "DEVELOPER does not have DELETE_FEATURE_FLAG permission",
			role:       ROLE_DEVELOPER,
			permission: PermissionDeleteFeatureFlag,
			expected:   false,
		},
		{
			name:       "VIEWER has READ_FEATURE_FLAG permission",
			role:       ROLE_VIEWER,
			permission: PermissionReadFeatureFlag,
			expected:   true,
		},
		{
			name:       "ADMIN has UPDATE_USER permission",
			role:       ROLE_ADMIN,
			permission: PermissionUpdateUser,
			expected:   true,
		},
		{
			name:       "DEVELOPER does not have UPDATE_USER permission",
			role:       ROLE_DEVELOPER,
			permission: PermissionUpdateUser,
			expected:   false,
		},
		{
			name:       "Unknown role has no permissions",
			role:       "UNKNOWN_ROLE",
			permission: PermissionReadFeatureFlag,
			expected:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := HasPermission(test.role, test.permission)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name           string
		userContext    *UserContext
		permission     Permission
		expectedStatus int
		expectedError  error
	}{
		{
			name: "ADMIN with valid permission",
			userContext: &UserContext{
				UserId: "admin-123",
				Role:   ROLE_ADMIN,
				Email:  "admin@example.com",
			},
			permission:     PermissionCreateFeatureFlag,
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name: "VIEWER without required permission",
			userContext: &UserContext{
				UserId: "viewer-123",
				Role:   ROLE_VIEWER,
				Email:  "viewer@example.com",
			},
			permission:     PermissionCreateFeatureFlag,
			expectedStatus: http.StatusForbidden,
			expectedError:  nil,
		},
		{
			name:           "Nil user context",
			userContext:    nil,
			permission:     PermissionReadFeatureFlag,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  nil,
		},
		{
			name: "DEVELOPER with valid permission",
			userContext: &UserContext{
				UserId: "dev-123",
				Role:   ROLE_DEVELOPER,
				Email:  "dev@example.com",
			},
			permission:     PermissionUpdateFeatureFlag,
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := RequirePermission(test.userContext, test.permission)
			assert.Equal(t, test.expectedStatus, response.StatusCode)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestRequireAnyPermission(t *testing.T) {
	tests := []struct {
		name           string
		userContext    *UserContext
		permissions    []Permission
		expectedStatus int
		expectedError  error
	}{
		{
			name: "ADMIN with one of the permissions",
			userContext: &UserContext{
				UserId: "admin-123",
				Role:   ROLE_ADMIN,
				Email:  "admin@example.com",
			},
			permissions:    []Permission{PermissionCreateFeatureFlag, PermissionDeleteFeatureFlag},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name: "VIEWER without any of the permissions",
			userContext: &UserContext{
				UserId: "viewer-123",
				Role:   ROLE_VIEWER,
				Email:  "viewer@example.com",
			},
			permissions:    []Permission{PermissionCreateFeatureFlag, PermissionDeleteFeatureFlag},
			expectedStatus: http.StatusForbidden,
			expectedError:  nil,
		},
		{
			name:           "Nil user context",
			userContext:    nil,
			permissions:    []Permission{PermissionReadFeatureFlag},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  nil,
		},
		{
			name: "DEVELOPER with one valid permission",
			userContext: &UserContext{
				UserId: "dev-123",
				Role:   ROLE_DEVELOPER,
				Email:  "dev@example.com",
			},
			permissions:    []Permission{PermissionCreateFeatureFlag, PermissionDeleteFeatureFlag},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := RequireAnyPermission(test.userContext, test.permissions...)
			assert.Equal(t, test.expectedStatus, response.StatusCode)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestCanAccessUserResource(t *testing.T) {
	tests := []struct {
		name           string
		userContext    *UserContext
		resourceUserId string
		expected       bool
	}{
		{
			name: "User accessing own resource",
			userContext: &UserContext{
				UserId: "user-123",
				Role:   ROLE_VIEWER,
				Email:  "user@example.com",
			},
			resourceUserId: "user-123",
			expected:       true,
		},
		{
			name: "ADMIN accessing other user's resource",
			userContext: &UserContext{
				UserId: "admin-123",
				Role:   ROLE_ADMIN,
				Email:  "admin@example.com",
			},
			resourceUserId: "user-456",
			expected:       true,
		},
		{
			name: "VIEWER accessing other user's resource",
			userContext: &UserContext{
				UserId: "viewer-123",
				Role:   ROLE_VIEWER,
				Email:  "viewer@example.com",
			},
			resourceUserId: "user-456",
			expected:       false,
		},
		{
			name: "DEVELOPER accessing other user's resource",
			userContext: &UserContext{
				UserId: "dev-123",
				Role:   ROLE_DEVELOPER,
				Email:  "dev@example.com",
			},
			resourceUserId: "user-456",
			expected:       false,
		},
		{
			name:           "Nil user context",
			userContext:    nil,
			resourceUserId: "user-123",
			expected:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := CanAccessUserResource(test.userContext, test.resourceUserId)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestRolePermissions(t *testing.T) {
	// Test that all roles have expected permissions
	tests := []struct {
		role       string
		permission Permission
		shouldHave bool
	}{
		// ADMIN permissions
		{ROLE_ADMIN, PermissionCreateFeatureFlag, true},
		{ROLE_ADMIN, PermissionUpdateFeatureFlag, true},
		{ROLE_ADMIN, PermissionReadFeatureFlag, true},
		{ROLE_ADMIN, PermissionDeleteFeatureFlag, true},
		{ROLE_ADMIN, PermissionCreateUserMapping, true},
		{ROLE_ADMIN, PermissionUpdateUserMapping, true},
		{ROLE_ADMIN, PermissionReadUserMapping, true},
		{ROLE_ADMIN, PermissionReadUser, true},
		{ROLE_ADMIN, PermissionUpdateUser, true},
		{ROLE_ADMIN, PermissionDeleteUser, true},

		// DEVELOPER permissions
		{ROLE_DEVELOPER, PermissionCreateFeatureFlag, true},
		{ROLE_DEVELOPER, PermissionUpdateFeatureFlag, true},
		{ROLE_DEVELOPER, PermissionReadFeatureFlag, true},
		{ROLE_DEVELOPER, PermissionDeleteFeatureFlag, false},
		{ROLE_DEVELOPER, PermissionCreateUserMapping, true},
		{ROLE_DEVELOPER, PermissionUpdateUserMapping, true},
		{ROLE_DEVELOPER, PermissionReadUserMapping, true},
		{ROLE_DEVELOPER, PermissionReadUser, true},
		{ROLE_DEVELOPER, PermissionUpdateUser, false},
		{ROLE_DEVELOPER, PermissionDeleteUser, false},

		// VIEWER permissions
		{ROLE_VIEWER, PermissionCreateFeatureFlag, false},
		{ROLE_VIEWER, PermissionUpdateFeatureFlag, false},
		{ROLE_VIEWER, PermissionReadFeatureFlag, true},
		{ROLE_VIEWER, PermissionDeleteFeatureFlag, false},
		{ROLE_VIEWER, PermissionCreateUserMapping, false},
		{ROLE_VIEWER, PermissionUpdateUserMapping, false},
		{ROLE_VIEWER, PermissionReadUserMapping, true},
		{ROLE_VIEWER, PermissionReadUser, true},
		{ROLE_VIEWER, PermissionUpdateUser, false},
		{ROLE_VIEWER, PermissionDeleteUser, false},
	}

	for _, test := range tests {
		t.Run(test.role+"_"+string(test.permission), func(t *testing.T) {
			result := HasPermission(test.role, test.permission)
			assert.Equal(t, test.shouldHave, result, "Role %s should %s have permission %s",
				test.role, map[bool]string{true: "", false: "not"}[test.shouldHave], test.permission)
		})
	}
}

