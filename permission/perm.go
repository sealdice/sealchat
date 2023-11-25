package permission

import (
	"github.com/mikespook/gorbac"
	"github.com/samber/lo"
)

var perm *gorbac.RBAC

func GetPerm() *gorbac.RBAC {
	return perm
}

func Init() {
	perm = gorbac.New()

	roleAdmin := gorbac.NewStdRole("role-admin")
	gorbac.NewStdRole("role-user")
	gorbac.NewStdRole("role-visitor")

	pA := gorbac.NewStdPermission("permission-a")

	lo.Must0(roleAdmin.Assign(pA))
	lo.Must0(perm.Add(roleAdmin))
}
