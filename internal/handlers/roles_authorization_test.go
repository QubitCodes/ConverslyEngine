package handlers_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shridarpatil/whatomate/internal/handlers"
	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/shridarpatil/whatomate/test/testutil"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

func wm001RoleAdministrator(t *testing.T, app *handlers.App, org uuid.UUID) *uuid.UUID {
	t.Helper()
	role := testutil.CreateTestRoleWithKeys(t, app.DB, org, "Role administrator", []string{"users:read", "users:write"})
	return &role.ID
}

func TestWM001_RestrictedRoleHandlersDenyAccess(t *testing.T) {
	for _, action := range []string{"list", "read", "create", "update", "delete"} {
		t.Run(action, func(t *testing.T) {
			app := newTestApp(t)
			org := testutil.CreateTestOrganization(t, app.DB)
			role := testutil.CreateTestRoleWithKeys(t, app.DB, org.ID, "Restricted messaging", []string{"contacts:read"})
			user := testutil.CreateTestUser(t, app.DB, org.ID, testutil.WithRoleID(&role.ID))
			req := testutil.NewJSONRequest(t, handlers.RoleRequest{Name: "Escalated", Permissions: []string{"users:read", "users:write"}})
			req.RequestCtx.SetUserValue("user_id", user.ID)
			req.RequestCtx.SetUserValue("organization_id", org.ID)
			req.RequestCtx.SetUserValue("id", role.ID.String())
			var call func(*fastglue.Request) error
			switch action {
			case "list":
				call = app.ListRoles
			case "read":
				call = app.GetRole
			case "create":
				call = app.CreateRole
			case "update":
				call = app.UpdateRole
			case "delete":
				call = app.DeleteRole
			}
			require.NoError(t, call(req))
			require.Equal(t, fasthttp.StatusForbidden, testutil.GetResponseStatusCode(req))
			var stored models.CustomRole
			require.NoError(t, app.DB.Preload("Permissions").First(&stored, "id = ?", role.ID).Error)
			require.Equal(t, role.Name, stored.Name)
			require.Len(t, stored.Permissions, 1)
			require.Equal(t, "contacts", stored.Permissions[0].Resource)
			require.Equal(t, "read", stored.Permissions[0].Action)
		})
	}
}

func TestWM001_RoleAdministratorCannotCrossOrganisation(t *testing.T) {
	for _, action := range []string{"read", "update", "delete"} {
		t.Run(action, func(t *testing.T) {
			app := newTestApp(t)
			own := testutil.CreateTestOrganization(t, app.DB)
			foreign := testutil.CreateTestOrganization(t, app.DB)
			user := testutil.CreateTestUser(t, app.DB, own.ID, testutil.WithRoleID(wm001RoleAdministrator(t, app, own.ID)))
			role := testutil.CreateTestRoleWithKeys(t, app.DB, foreign.ID, "Foreign role", []string{"contacts:read"})
			req := testutil.NewJSONRequest(t, handlers.RoleRequest{Name: "Do not change", Permissions: []string{"users:write"}})
			req.RequestCtx.SetUserValue("user_id", user.ID)
			req.RequestCtx.SetUserValue("organization_id", own.ID)
			req.RequestCtx.SetUserValue("id", role.ID.String())
			var call func(*fastglue.Request) error
			switch action {
			case "read":
				call = app.GetRole
			case "update":
				call = app.UpdateRole
			case "delete":
				call = app.DeleteRole
			}
			require.NoError(t, call(req))
			require.Equal(t, fasthttp.StatusNotFound, testutil.GetResponseStatusCode(req))
			var stored models.CustomRole
			require.NoError(t, app.DB.First(&stored, "id = ?", role.ID).Error)
			require.Equal(t, role.Name, stored.Name)
		})
	}
}
