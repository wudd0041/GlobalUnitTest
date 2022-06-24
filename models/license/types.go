package license

import "strconv"

type License struct {
	OrgUUID    string `db:"org_uuid"`
	Type       int    `db:"type" json:"type"`
	Edition    string `db:"edition" json:"edition"`
	AddType    int    `db:"add_type" json:"add_type"`
	Scale      int    `db:"scale" json:"scale"`
	ExpireTime int64  `db:"expire_time" json:"expire_time"`
	UpdateTime int64  `db:"update_time" json:"update_time"`
}

func (l *License) UniqueKey() string {
	return l.OrgUUID + strconv.Itoa(l.Type) + l.Edition
}

type LicenseGrant struct {
	OrgUUID     string `db:"org_uuid"`
	UserUUID    string `db:"user_uuid" json:"user_uuid"`
	LicenseType int    `db:"type" json:"type"`
	Status      int    `db:"status" json:"status"`
}

type OrgDefaultGrant struct {
	OrgUUID      string `db:"org_uuid"`
	LicenseType  int    `db:"type" json:"type"`
	DefaultGrant bool   `db:"default_grant" json:"default_grant"` // 默认授权开关设置
}

type LicenseAlter struct {
	UUID        string `db:"uuid"`
	OrgUUID     string `db:"org_uuid"`
	LicenseType int    `db:"type"`
	Edition     string `db:"edition"`
	AddType     int    `db:"add_type"`
	Scale       int    `db:"scale"`
	ExpireTime  int64  `db:"expire_time"`
	CreateTime  int64  `db:"create_time"`
}
