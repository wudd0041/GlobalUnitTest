package license

import (
	gorp "gopkg.in/gorp.v1"
)

const (
	DefaultGrantTrue  = true
	DefaultGrantFalse = false

	defaultGrantColumns = `org_uuid, license_type, default_grant`
)

func AddOrUpdateDefaultGrant(src gorp.SqlExecutor, dg *OrgDefaultGrant) error {

	return nil
}

func BatchAddOrUpdateOrgDefaultGrants(src gorp.SqlExecutor, defaultGrants []*OrgDefaultGrant) error {

	return nil
}

func ListOrgDefaultGrantTypes(src gorp.SqlExecutor, orgUUID string) ([]int, error) {

	return nil, nil
}
