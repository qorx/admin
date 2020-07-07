package admin_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/qor/admin"
	. "github.com/qor/admin/tests/dummy"
	"github.com/qor/admin/tests/utils"
	"github.com/qor/qor"
	qorTestUtils "github.com/qor/qor/test/utils"
	"github.com/qor/roles"
)

func TestGroupMenuPermission(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	group := admin.Group{Name: "test group", Users: fmt.Sprintf("%d", user.ID), AllowList: "Company,Credit Card"}
	utils.AssertNoErr(t, db.Save(&group).Error)

	// setup Admin
	ctx := &admin.Context{Context: &qor.Context{CurrentUser: user, DB: Admin.DB}, Admin: Admin, Settings: map[string]interface{}{}}

	companyMenu := Admin.GetMenu("Companies")
	if !companyMenu.HasPermission(roles.Read, ctx) {
		t.Error("user should have permission to access allowed Company resource")
	}

	// Check no permission menu
	noPermissionMenu := Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin", Priority: 1})
	if !noPermissionMenu.HasPermission(roles.Read, ctx) {
		t.Error("menu with no permission set should be always accessible")
	}

	// check no group permission menu
	group.AllowList = ""
	utils.AssertNoErr(t, db.Save(&group).Error)
	if companyMenu.HasPermission(roles.Read, ctx) {
		t.Error("user should not have permission to access company when it is not allowed")
	}

	individualMenuWithPermission := Admin.AddMenu(&admin.Menu{Name: "ExternalURL", Permission: roles.Allow(roles.CRUD, Role_developer)})
	if individualMenuWithPermission.HasPermission(roles.Read, ctx) {
		t.Error("admin user should not have permission to access menu which is visible to Developer only")
	}
}

func TestGroupMenuPermissionShouldHasLowerPriorityThanRole(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{})
	user := User{Name: LoggedInUserName, Role: Role_system_administrator}
	utils.AssertNoErr(t, db.Save(&user).Error)

	// setup Admin, group enabled but this user has no group registered
	ctx := &admin.Context{Context: &qor.Context{CurrentUser: user, DB: Admin.DB},
		Admin: Admin, Settings: map[string]interface{}{}}
	ctx.Context.Roles = []string{Role_system_administrator}

	Admin.AddResource(&Profile{}, &admin.Config{Permission: roles.Allow(roles.CRUD, Role_system_administrator)})
	profileMenu := Admin.GetMenu("Profiles")
	if !profileMenu.HasPermission(roles.Read, ctx) {
		t.Error("user should have permission to access roles allowed resource")
	}
}

func TestGroupRouterPermission(t *testing.T) {
	qorTestUtils.ResetDBTables(db, &admin.Group{}, &User{})
	user := User{Name: LoggedInUserName, Role: "admin"}
	utils.AssertNoErr(t, db.Save(&user).Error)

	group := admin.Group{Name: "test group", Users: fmt.Sprintf("%d", user.ID), AllowList: "Company,CreditCard"}
	utils.AssertNoErr(t, db.Save(&group).Error)

	// TODO: C R U D should all be test covered.
	req, err := http.Get(server.URL + "/admin/companies")
	utils.AssertNoErr(t, err)

	if req.StatusCode != 200 {
		t.Errorf("Expect user with group permission to have the ability to visit companies")
	}

	group.AllowList = "CreditCard"
	utils.AssertNoErr(t, db.Save(&group).Error)
	req, err = http.Get(server.URL + "/admin/companies")
	utils.AssertNoErr(t, err)

	if req.StatusCode != 404 {
		t.Errorf("Expect user without group permission not have the ability to visit companies")
	}
}
