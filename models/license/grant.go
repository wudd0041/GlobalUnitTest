package license

import (
	"strings"

	gorp "gopkg.in/gorp.v1"
)

const (
	GrantStatusNormal      = 1
	GrantStatusExpired     = 2
	licenseGrantColumnsStr = "org_uuid, type, user_uuid, status"
)

var (
	licenseGrantColumns = strings.Split(licenseGrantColumnsStr, ", ")
)

func AddGrant(src gorp.SqlExecutor, orgUUID string, tp int, userUUIDs ...string) error {

	return nil
}

func BatchAddGrant(tx *gorp.Transaction, grants []*LicenseGrant) error {

	return nil
}

func DeleteGrant(src gorp.SqlExecutor, orgUUID string, tp int, userUUIDs ...string) error {

	return nil
}

func BatchDeleteGrants(src gorp.SqlExecutor, orgUUID, userUUID string, types ...int) error {

	return nil
}

func BatchDeleteUsersGrants(src gorp.SqlExecutor, orgUUID string, userUUIDs []string, types ...int) error {

	return nil
}

func DeleteAllGrant(src gorp.SqlExecutor, orgUUID string, userUUIDs ...string) error {

	return nil
}

func MapOrgLicenseUsage(src gorp.SqlExecutor, orgUUID string) (map[int]int, error) {

	return nil, nil
}

func OrgLicenseUsage(src gorp.SqlExecutor, orgUUID string, tp int) (int, error) {

	return 0, nil
}

func ListUserGrantedTypes(src gorp.SqlExecutor, orgUUID string, userUUID string) ([]int, error) {

	return nil, nil
}

func ListGrantUser(src gorp.SqlExecutor, orgUUID string, tp int) ([]*LicenseGrant, error) {

	return nil, nil
}

func ListGrantUserLimit(src gorp.SqlExecutor, orgUUID string, tp int, limit int) ([]string, error) {

	return nil, nil
}

func MapUserLicenseByUserUUID(src gorp.SqlExecutor, orgUUID string, userUUIDs []string) (map[string]map[int]int, error) {

	return nil, nil
}
