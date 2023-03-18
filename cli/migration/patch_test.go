package migration

import "testing"

func TestPath(t *testing.T) {
	patchDatabase("mysql://root:nYp5ztg4@tcp(10.233.25.208:3306)/db_tenant_mgt")
}