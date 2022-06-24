package license

import (
	"gopkg.in/gorp.v1"
	licenseModel "license_testing/models/license"
)

// GetUserGrantByType 获取用户指定LicenseType授权 无授权不报错，返回nil
func GetUserGrantByType(src gorp.SqlExecutor, orgUUID string, userUUID string, licenseType LicenseType) (*LicenseUserGrant, error) {

	return nil, nil

}

// ListUserGrants 获取用户所有LicenseType授权列表
func ListUserGrants(src gorp.SqlExecutor, orgUUID string, userUUID string) ([]*LicenseUserGrant, error) {

	return nil, nil
}

//ListUserGrantTypeInts 获取用户所有LicenseType授权列表（int类型）
func ListUserGrantTypeInts(src gorp.SqlExecutor, orgUUID string, userUUID string) ([]int, error) {

	return nil, nil
}

/*
	批量获取用户所有LicenseType授权map
	ret: map[userUUID]map[licenseTypeInt][status]
*/
func MapUserGrantTypeIntsByUserUUIDs(src gorp.SqlExecutor, orgUUID string, userUUIDs []string) (map[string]map[int]int, error) {
	return licenseModel.MapUserLicenseByUserUUID(src, orgUUID, userUUIDs)
}

// ListOrgUserGrantsByType 获取组织下指定LicenseType类型的用户授权列表
func ListOrgUserGrantsByType(src gorp.SqlExecutor, orgUUID string, licenseType LicenseType) ([]*LicenseUserGrant, error) {

	return nil, nil
}

// 获取组织下所有LicenseType的数量
// ret: map[licenseTypeInt]count
func MapOrgLicenseGrantCount(src gorp.SqlExecutor, orgUUID string) (map[int]int, error) {
	return licenseModel.MapOrgLicenseUsage(src, orgUUID)
}

//GrantLicenseToUser  授予组织下某个用户对应LicenseType
func GrantLicenseToUser(src gorp.SqlExecutor, orgUUID string, userUUID string, licenseType LicenseType) error {

	return nil
}

// 批量授予组织下多个用户对应LicenseType
// ret1: 授予成功用户列表 []userUUIDs
// ret2: 授予失败用户列表 []userUUIDs
func BatchGrantLicenseToUsers(
	src gorp.SqlExecutor,
	orgUUID string,
	userUUIDs []string,
	licenseType LicenseType) (succeedUsers []string, failedUsers []string, err error) {

	return nil, nil, nil
}

//GrantLicensesToUser 授予组织下某个用户多个LicenseType
// ret1: 授予成功Type列表 []LicenseTypeInt
// ret2: 授予失败Type列表 []LicenseTypeInt
func GrantLicensesToUser(
	tx *gorp.Transaction,
	orgUUID string,
	userUUID string,
	types []LicenseType) (succeedTypes []int, failedTypes []int, err error) {

	return nil, nil, nil
}

// ReclaimUserGrant 回收组织下某个用户的指定LicenseType授权
func ReclaimUserGrant(src gorp.SqlExecutor, orgUUID string, userUUID string, tpe LicenseType) error {
	return nil
}

// ReclaimUserGrants 回收组织下某个用户的多个LicenseType授权
func ReclaimUserGrants(src gorp.SqlExecutor, orgUUID string, userUUID string, types []LicenseType) error {

	return nil
}

// ReclaimUserAllGrant 回收组织下用户的所有授权
func ReclaimUserAllGrant(src gorp.SqlExecutor, orgUUID string, userUUIDs ...string) error {
	return nil
}

// BatchReclaimUsersGrant 批量回收组织下多个用户的指定LicenseType授权
func BatchReclaimUsersGrant(src gorp.SqlExecutor, orgUUID string, userUUIDs []string, tpe LicenseType) error {
	return nil
}

// BatchReclaimUsersGrants 批量回收组织下多个用户的多个LicenseType授权
func BatchReclaimUsersGrants(src gorp.SqlExecutor, orgUUID string, userUUIDs []string, types []LicenseType) error {

	return nil
}
