package license

import (
	gorp "gopkg.in/gorp.v1"
)

const (
	licenseAlterColumns = `uuid, org_uuid, license_type, edition, add_type, scale, expire_time, create_time`
)

func BatchInsertOrgLicenseAlters(src gorp.SqlExecutor, licenseAlters ...*LicenseAlter) error {

	return nil

}
