package license

import (
	gorp "gopkg.in/gorp.v1"
)

const (
	licenseColumns = `org_uuid, type, add_type, scale, expire_time, update_time, edition`
)

func ListOrgTypeLicenses(src gorp.SqlExecutor, orgUUID string, licenseType int) ([]*License, error) {
	return nil, nil
}

func ListByOrgUUID(src gorp.SqlExecutor, orgUUID string) ([]*License, error) {

	return nil, nil
}

func BatchListByOrgUUIDs(src gorp.SqlExecutor, orgUUIDs ...string) ([]*License, error) {

	return nil, nil
}

func ListExpire(src gorp.SqlExecutor, addType int, startStamp, endStamp int64) ([]*License, error) {

	return nil, nil
}

func AddOrUpdateLicenses(src gorp.SqlExecutor, licenses ...*License) error {

	return nil
}

func ListUserGrantedLicensesByType(src gorp.SqlExecutor, orgUUID string, userUUID string, licenseType int) ([]*License, error) {

	return nil, nil
}

func ListUserGrantedLicenses(src gorp.SqlExecutor, orgUUID string, userUUID string) ([]*License, error) {

	return nil, nil
}
