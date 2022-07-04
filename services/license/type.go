package license

import (
	"strconv"
)

const (
	LicenseTypeInvalid     = 0
	LicenseTypeProject     = 1
	LicenseTypeWiki        = 2
	LicenseTypeTestCase    = 3
	LicenseTypePipeline    = 4
	LicenseTypePlan        = 5
	LicenseTypeAccount     = 6
	LicenseTypeDesk        = 7
	LicenseTypePerformance = 8

	LicenseNameInvalid     = ""
	LicenseNameProject     = "project"
	LicenseNameWiki        = "wiki"
	LicenseNameTestcase    = "testcase"
	LicenseNamePipeline    = "pipeline"
	LicenseNamePlan        = "plan"
	LicenseNameAccount     = "account"
	LicenseNameDesk        = "desk"
	LicenseNamePerformance = "performance"
)

var appTypeNameMap = map[string]int{
	LicenseNameProject:     LicenseTypeProject,
	LicenseNameWiki:        LicenseTypeWiki,
	LicenseNameTestcase:    LicenseTypeTestCase,
	LicenseNamePipeline:    LicenseTypePipeline,
	LicenseNamePlan:        LicenseTypePlan,
	LicenseNameAccount:     LicenseTypeAccount,
	LicenseNameDesk:        LicenseTypeDesk,
	LicenseNamePerformance: LicenseTypePerformance,
}

var appNameTypeMap = map[int]string{
	LicenseTypeProject:     LicenseNameProject,
	LicenseTypeWiki:        LicenseNameWiki,
	LicenseTypeTestCase:    LicenseNameTestcase,
	LicenseTypePipeline:    LicenseNamePipeline,
	LicenseTypePlan:        LicenseNamePlan,
	LicenseTypeAccount:     LicenseNameAccount,
	LicenseTypeDesk:        LicenseNameDesk,
	LicenseTypePerformance: LicenseNamePerformance,
}

const (
	AddTypeTrail = 0 // 试用获得
	AddTypePay   = 1 // 购买获得
	AddTypeFree  = 2 // 免费获得

	DefaultGrantTrue  = true
	DefaultGrantFalse = false

	UnLimitExpire = -1

	GrantStatusNormal  = 1
	GrantStatusExpired = 2
)

const (
	EditionTeam            = "team"
	EditionEnterprise      = "enterprise"
	EditionEnterpriseTrial = "enterprise-trial"
)

type LicenseType interface {
	Type() int
	Name() string
}

// 「版本」
type LicenseEdition struct {
	LicenseTag
	Priority    int
	InvalidTime int64 //秒
}

func (le *LicenseEdition) GetPriority() int {
	return le.Priority
}

// 标示一个License对应的「应用」和「应用版本」
type LicenseTag struct {
	LicenseType `db:"type" json:"type"`
	EditionName string
}

func (lt *LicenseTag) UniqueKey() string {
	return strconv.Itoa(lt.Type()) + lt.EditionName
}

// 授予给组织的License实体
type LicenseEntity struct {
	LicenseTag
	OrgUUID    string
	Scale      int
	ExpireTime int64
	AddType    int
	edition    *LicenseEdition
}

func (le *LicenseEntity) UniqueKey() string {
	return le.OrgUUID + le.LicenseTag.UniqueKey()
}

func (le *LicenseEntity) Valid() bool {
	if le.edition == nil {
		le.edition = editionHub.findConfigByTypeEdition(le.Type(), le.LicenseTag.EditionName)
		if le.edition == nil {
			return false
		}
	}
	return true
}

func (le *LicenseEntity) IsExpire() bool {
	return true
}

func (le *LicenseEntity) Priority() int {
	return le.edition.GetPriority()
}

// 添加License结构
type LicenseAdd struct {
	OrgUUID    string
	Scale      int
	ExpireTime int64
	LicenseTag
}

func (la *LicenseAdd) UniqueKey() string {
	return la.OrgUUID + la.LicenseTag.UniqueKey()
}

func (la *LicenseAdd) Valid() bool {
	return true
}

// 组织默认授予用户配置
type LicenseDefaultGrant struct {
	OrgUUID      string
	DefaultGrant bool
	LicenseType
}

// 授予给用户的License实体
type LicenseUserGrant struct {
	LicenseTag
	UserUUID string
	Status   int
}

const (
	AlterTypeNew    = "new"
	AlterTypeUpdate = "update"
)

type LicenseAlterInfo struct {
	ExpireTime int64 `json:"expire_time"`
	Scale      int   `json:"scale"`
}

// Lincense变更后的信息，UUID用于标识（相当于流水号）
type LicenseAlter struct {
	LicenseTag
	UUID       string
	OrgUUID    string
	AddType    int
	AlterType  string
	Old        *LicenseAlterInfo
	New        *LicenseAlterInfo
	CreateTime int64
}
