package license

import (
	"gopkg.in/gorp.v1"
)

// 获取组织下当前装配的LicenseType实体信息, 无授权则返回nil
func GetOrgLicenseByType(src gorp.SqlExecutor, orgUUID string, licenseType LicenseType) (*LicenseEntity, error) {

	return nil, nil
}

// 获取组织下某个时间点装配的LicenseType实体信息, 无授权则返回nil
func GetOrgLicenseByTypeAndTimestamp(src gorp.SqlExecutor, orgUUID string, licenseType LicenseType, sec int64) (*LicenseEntity, error) {
	return nil, nil

}

// 获取组织下当前已经装配的所有Licenses
func GetOrgLicenses(src gorp.SqlExecutor, orgUUID string) ([]*LicenseEntity, error) {
	return nil, nil

}

// 获取组织下当前已经装配的所有Licenses, map[LicenseTypeInt]*LicenseEntity 返回
func GetOrgLicensesMap(src gorp.SqlExecutor, orgUUID string) (map[int]*LicenseEntity, error) {
	return nil, nil

}

// 批量获取组织下当前已经装配的所有Licenses, map[orgUUID]map[LicenseTypeInt]*LicenseEntity 返回
func BatchGetOrgLicensesMaps(src gorp.SqlExecutor, orgUUIDs []string) (map[string]map[int]*LicenseEntity, error) {
	return nil, nil

}

// 获取组织下某种类型的所有LicenseEntity(包含过期, 一般业务不关注)
func GetOrgAllLicensesByType(src gorp.SqlExecutor, orgUUID string, licenseType LicenseType) ([]*LicenseEntity, error) {
	return nil, nil

}

// 获取组织下所有LicenseEntity(包含过期, 一般业务不关注) map返回
func GetOrgAllLicensesMap(src gorp.SqlExecutor, orgUUID string) (map[int][]*LicenseEntity, error) {
	return nil, nil

}

// 添加一个License, 若LicenseTag已存在，则报错
func AddLicense(src gorp.SqlExecutor, addType int, addLicense *LicenseAdd) (*LicenseAlter, error) {
	return nil, nil

}

// 批量添加多个License, 若其中一个LicenseTag已存在，则报错
func BatchAddLicenses(src gorp.SqlExecutor, addType int, addLicenses []*LicenseAdd) ([]*LicenseAlter, error) {
	return nil, nil

}

// 更新组织下对应LicenseTag的最大授予人数
// 若组织下不存在该LicenseTag则报错
func UpdateOrgLicenseScale(src gorp.SqlExecutor, orgUUID string, tag *LicenseTag, newScale int) (*LicenseAlter, error) {
	return nil, nil

}

// 批量更新组织下多个LicenseTag的最大授予人数
// 若组织下不存在某个LicenseTag则报错
func BatchUpdateOrgLicenseScale(src gorp.SqlExecutor, orgUUID string, tags []*LicenseTag, newScale int) ([]*LicenseAlter, error) {
	return nil, nil

}

// 更新组织下对应LicenseTag的过期时间
// 若组织下不存在该LicenseTag则报错
func RenewalOrgLicenseExpire(src gorp.SqlExecutor, orgUUID string, tag *LicenseTag, expireTime int64) (*LicenseAlter, error) {
	return nil, nil

}

// 批量更新组织下多个LicenseTag的过期时间
// 若组织下不存在某个LicenseTag则报错
func BatchRenewalOrgLicenseExpire(src gorp.SqlExecutor, orgUUID string, tags []*LicenseTag, expireTime int64) ([]*LicenseAlter, error) {
	return nil, nil

}

// 新增或者更新对应LicenseTag的默认授权配置
func AddOrUpdateOrgDefaultGrant(src gorp.SqlExecutor, orgUUID string, defaultGrant *LicenseDefaultGrant) error {
	return nil

}

// 获取组织下所有LicenseTag的默认授权配置
func GetOrgDefaultGrantLicenses(src gorp.SqlExecutor, orgUUID string) ([]*LicenseDefaultGrant, error) {
	return nil, nil

}

// 批量新增或者更新组织下多个LicenseTag的默认授权配置
func BatchAddOrUpdateOrgDefaultGrants(src gorp.SqlExecutor, orgUUID string, defaultGrants []*LicenseDefaultGrant) error {
	return nil

}

// 查询一段时间区间内会过期的LicenseEntity
func ListExpireInTimeStampRange(src gorp.SqlExecutor, startStamp, endStamp int64) ([]*LicenseEntity, error) {
	return nil, nil

}

func batchCheckLicenseTagsAndReturnEntities(src gorp.SqlExecutor, orgUUID string, tags []*LicenseTag) ([]*LicenseEntity, error) {
	return nil, nil
}
