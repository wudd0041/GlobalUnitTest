package license

import (
	"gopkg.in/gorp.v1"
)

// 根据alterUUID批量查询
func MapLicenseAltersByUUIDs(src gorp.SqlExecutor, alterUUIDs []string) (map[string]*LicenseAlter, error) {
	return nil, nil
}

// 查询组织内所有LicenseAlter
func ListLicenseAltersByOrgUUID(src gorp.SqlExecutor, orgUUID string) ([]*LicenseAlter, error) {
	return nil, nil
}

// 查询组织内指定licenseType的所有LicenseAlter
func ListLicenseAltersByOrgUUIDAndType(src gorp.SqlExecutor, orgUUID string, licenseType LicenseType) ([]*LicenseAlter, error) {
	return nil, nil
}

// 查询组织内指定licenseType和Edition的所有LicenseAlter
func ListLicenseAltersByOrgUUIDAndTypeEdition(src gorp.SqlExecutor, orgUUID string, licenseType LicenseType, edition string) ([]*LicenseAlter, error) {
	return nil, nil
}
