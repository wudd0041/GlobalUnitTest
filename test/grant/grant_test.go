package grant

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/gorp.v1"
	"license_testing/services/license"
	"license_testing/utils/uuid"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

type testSuite struct {
	// 这里放全局变量
	suite.Suite
	db          *sqlx.DB
	sqlExecutor gorp.SqlExecutor
	Transaction *gorp.Transaction
	orgIds      []string
	userIds     []string
	orgUserMap  map[string][]string
	autoFlag    string
}

var dbm *gorp.DbMap

type orgLicense struct {
	Org_uuid    string `db:"org_uuid"`
	Type        int    `db:"type"`
	Edition     string `db:"edition"`
	Add_type    int    `db:"add_type"`
	Scale       int    `db:"scale"`
	Expire_time int64  `db:"expire_time"`
	Update_time int64  `db:"update_time"`
}

type licenseGrant struct {
	Org_uuid  string `db:"org_uuid"`
	Type      int    `db:"type"`
	User_uuid string `db:"user_uuid"`
	Status    int    `db:"status"`
}

// setup 初始化数据库(帐号需要修改)
func (suite *testSuite) SetupSuite() {

	database, err := sqlx.Open("mysql", "bang:bang@tcp(127.0.0.1:3306)/U0015_LOCAL")
	if err != nil {
		panic(err)
	}
	err = database.Ping()
	if err != nil {
		panic(err)
	}
	suite.db = database
	fmt.Println("数据库连接成功！")

	dbm = &gorp.DbMap{
		Db: database.DB,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "utf8mb4",
		},
	}

	// 基础数据
	suite.sqlExecutor = dbm
	suite.orgIds, suite.orgUserMap, _ = suite.BatchInsertLicensAndGrant()

}

// teardown 关闭数据库链接
func (suite *testSuite) TearDownSuite() {
	defer suite.db.Close()
}

// 数据库批量插入license、licenseGrant，作为初始化数据
func (suite *testSuite) BatchInsertLicensAndGrant() ([]string, map[string][]string, error) {
	orgId1 := uuid.UUID()
	orgId2 := uuid.UUID()
	orgIds := []string{orgId1, orgId2}
	l1 := orgLicense{orgId1, license.LicenseTypeProject, license.EditionEnterpriseTrial, 0, 100, -1, 1655714729}
	l2 := orgLicense{orgId1, license.LicenseTypeWiki, license.EditionEnterpriseTrial, 0, 100, -1, 1655714729}
	l3 := orgLicense{orgId2, license.LicenseTypeProject, license.EditionEnterpriseTrial, 0, 100, -1, 1655714729}
	l4 := orgLicense{orgId2, license.LicenseTypeWiki, license.EditionEnterpriseTrial, 0, 100, -1, 1655714729}
	myOrgLicense := []*orgLicense{&l1, &l2, &l3, &l4}
	_, err := suite.db.NamedExec("INSERT INTO license (org_uuid,type,edition,add_type,scale,expire_time,update_time) "+
		"VALUES (:org_uuid,:type,:edition,:add_type,:scale,:expire_time,:update_time)", myOrgLicense)

	userId1 := uuid.UUID()
	userId2 := uuid.UUID()
	userId3 := uuid.UUID()
	userId4 := uuid.UUID()
	ug1 := licenseGrant{orgId1, license.LicenseTypeProject, userId1, 1}
	ug2 := licenseGrant{orgId1, license.LicenseTypeProject, userId2, 1}
	ug3 := licenseGrant{orgId1, license.LicenseTypeWiki, userId1, 1}
	ug4 := licenseGrant{orgId1, license.LicenseTypeWiki, userId2, 1}
	ug5 := licenseGrant{orgId2, license.LicenseTypeProject, userId3, 1}
	ug6 := licenseGrant{orgId2, license.LicenseTypeProject, userId4, 1}
	ug7 := licenseGrant{orgId2, license.LicenseTypeWiki, userId3, 1}
	ug8 := licenseGrant{orgId2, license.LicenseTypeWiki, userId4, 1}
	myLicenseGrant := []*licenseGrant{&ug1, &ug2, &ug3, &ug4, &ug5, &ug6, &ug7, &ug8}
	_, err2 := suite.db.NamedExec("INSERT INTO license_grant (org_uuid, type, user_uuid, status) "+
		"VALUES (:org_uuid, :type, :user_uuid, :status)", myLicenseGrant)

	if err != nil {
		fmt.Printf("BatchInsertLicense执行失败: %v", err)
	} else if err2 != nil {
		fmt.Printf("BatchInsertLicense执行失败: %v", err2)
	}

	var orgUserMap = make(map[string][]string)
	orgUserMap[orgId1] = []string{userId1, userId2}
	orgUserMap[orgId2] = []string{userId3, userId4}
	return orgIds, orgUserMap, err
}

// 构造licenseTag
func AddLicenseTag(typeInt int, editionName string) license.LicenseTag {
	licenseType := license.GetLicenseType(typeInt)
	licenseTag := license.LicenseTag{licenseType, editionName}
	return licenseTag
}

// 构造License
func (suite *testSuite) ManulAddLicense(orgUUID string, typeInt int, edition string, scale int, expire_time int64) {
	_, err := license.AddLicense(suite.sqlExecutor, license.AddTypePay, &license.LicenseAdd{orgUUID, scale, expire_time, AddLicenseTag(typeInt, edition)})
	if err != nil {
		fmt.Println("构造license发生错误：", err)
		panic(err)
	}
}

// 获取组织下所有LicenseType的数量
func (suite *testSuite) GetOrgLicenseTypeCount(orgUUID string) map[int]int {

	type orgLicenseType struct {
		Type  int
		Count int
	}
	var orgTypeCounts = []orgLicenseType{}
	err := suite.db.Select(&orgTypeCounts, "select type,count(type) as count from license_grant where org_uuid=? group by type ", orgUUID)
	if err != nil {
		fmt.Println("数据库执行失败: ", err)
		panic(err)
	}

	var TypeCountMap = make(map[int]int)
	for _, typeCount := range orgTypeCounts {
		TypeCountMap[typeCount.Type] = typeCount.Count
	}

	return TypeCountMap
}

// 获取用户指定LicenseType授权 无授权不报错，返回nil
func (suite *testSuite) TestGetUserGrantByType() {

	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		userUUID    string
		licenseType license.LicenseType
		expected    interface{}
	}

	newOrgUUID01 := uuid.UUID()
	newUserUUID01 := uuid.UUID()
	newOrgUUID02 := uuid.UUID()
	newUserUUID02 := uuid.UUID()
	licenseTypeProject := license.GetLicenseType(license.LicenseTypeProject)

	// 新增组织 -- 新增2个license -- 设置一个 license 的 scale 上限为1
	licenseTag := AddLicenseTag(license.LicenseTypeProject, license.EditionTeam)
	newLicenseAdd01 := license.LicenseAdd{newOrgUUID01, 1, -1, licenseTag}

	// 新增组织 -- 新增2个license -- 设置一个 license 的 scale 上限为0
	newLicenseAdd02 := license.LicenseAdd{newOrgUUID02, 0, -1, licenseTag}
	license.BatchAddLicenses(suite.sqlExecutor, license.AddTypeFree, []*license.LicenseAdd{&newLicenseAdd01, &newLicenseAdd02})
	// 新增授权 -- 可以查到新增授权
	license.GrantLicenseToUser(suite.sqlExecutor, newOrgUUID01, newUserUUID01, licenseTypeProject)
	license.GrantLicenseToUser(suite.sqlExecutor, newOrgUUID02, newUserUUID02, licenseTypeProject)

	// 测试数据
	data_suite := map[string]test{
		"传入1个应用licenseType，其中A应用scale未到上限：成功，返回对应内容": {
			suite.sqlExecutor,
			newOrgUUID01,
			newUserUUID01,
			licenseTypeProject,
			&license.LicenseUserGrant{licenseTag, newUserUUID01, 1},
		},
		"传入1个应用licenseType，其中A应用scale达到上限：失败，返回nil": {
			suite.sqlExecutor,
			newOrgUUID02,
			newUserUUID02,
			licenseTypeProject,
			nil,
		},
		"传入非空且不存在的orgUUID：返回nil": {
			suite.sqlExecutor,
			"123org",
			newUserUUID01,
			licenseTypeProject,
			nil,
		},
		"传入不存在的licenseType：返回报错": {
			suite.sqlExecutor,
			newOrgUUID01,
			newUserUUID01,
			license.GetLicenseType(100),
			"",
		},
		"传入非空且不存在的userUUID：返回nil": {
			suite.sqlExecutor,
			newOrgUUID01,
			"123user",
			licenseTypeProject,
			nil,
		},
		"传入orgUUID为空串：返回报错信息": {
			suite.sqlExecutor,
			"",
			newUserUUID01,
			licenseTypeProject,
			"",
		},
		"传入userUUID为空串：返回报错信息": {
			suite.sqlExecutor,
			newOrgUUID01,
			"",
			licenseTypeProject,
			"",
		},
		"传入licenseType为nil：返回报错信息": {
			suite.sqlExecutor,
			newOrgUUID01,
			newUserUUID01,
			nil,
			"",
		},
	}
	for name, tc := range data_suite {
		userGrant, err := license.GetUserGrantByType(tc.sql, tc.orgUUID, tc.userUUID, tc.licenseType)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, userGrant, name)
		} else if strings.Contains(name, "返回nil") {
			assert.Nil(suite.T(), err, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}

}

// ListUserGrants 获取用户所有LicenseType授权列表
func (suite *testSuite) TestListUserGrants() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		userUUID string
		expected []*license.LicenseUserGrant
	}

	// 遍历type,调用GetUserGrantByType，获取集合
	orgUUID := suite.orgIds[0]
	userUUID := suite.orgUserMap[orgUUID][0]
	var myLicenseUserGrants []*license.LicenseUserGrant
	for _, typeInt := range []int{1, 2, 3, 4, 5, 6, 7, 8} {
		licenseType := license.GetLicenseType(typeInt)
		licenseUserGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
		if licenseUserGrant != nil {
			myLicenseUserGrants = append(myLicenseUserGrants, licenseUserGrant)
		}
	}

	data_suite := map[string]test{
		"传入正确的组织id和用户id，查询用户所有LicenseType授权列表：成功，获取用户所有LicenseType授权列表": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			myLicenseUserGrants,
		},
		"传入错误的组织id和正确的用户id，查询用户所有LicenseType授权列表：返回内容为空": {
			suite.sqlExecutor,
			"123auto",
			userUUID,
			[]*license.LicenseUserGrant{},
		},
		"传入正确的组织id和错误的用户id，查询用户所有LicenseType授权列表：返回内容为空": {
			suite.sqlExecutor,
			orgUUID,
			"123auto",
			[]*license.LicenseUserGrant{},
		},
		"传入空串的组织id和正确的用户id，查询用户所有LicenseType授权列表：返回内容为空": {
			suite.sqlExecutor,
			"",
			userUUID,
			[]*license.LicenseUserGrant{},
		},
		"传入正确的组织id和空串的用户id，查询用户所有LicenseType授权列表：返回内容为空": {
			suite.sqlExecutor,
			orgUUID,
			"",
			[]*license.LicenseUserGrant{},
		},
	}

	for name, tc := range data_suite {
		licenseUserGrants, _ := license.ListUserGrants(tc.sql, tc.orgUUID, tc.userUUID)
		assert.ElementsMatch(suite.T(), tc.expected, licenseUserGrants, name)
	}

}

//ListUserGrantTypeInts 获取用户所有LicenseType授权列表（int类型）
func (suite *testSuite) TestListUserGrantTypeInts() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		userUUID string
		expected interface{}
	}

	// 遍历type,调用GetUserGrantByType，获取集合，组装成[]int
	orgUUID := suite.orgIds[0]
	userUUID := suite.orgUserMap[orgUUID][0]
	var userGrantTypeInts []int
	for _, typeInt := range []int{1, 2, 3, 4, 5, 6, 7, 8} {
		licenseType := license.GetLicenseType(typeInt)
		licenseUserGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
		if licenseUserGrant != nil {
			userGrantTypeInts = append(userGrantTypeInts, typeInt)
		}
	}

	data_suite := map[string]test{
		"传入正确的组织id和用户id，查询用户所有LicenseType授权列表：成功，获取用户所有LicenseType授权列表": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			userGrantTypeInts,
		},
		"传入错误的组织id和正确的用户id，查询用户所有LicenseType授权列表：返回内容为空": {
			suite.sqlExecutor,
			"123",
			userUUID,
			[]int{},
		},
		"传入正确的组织id和错误的用户id，查询用户所有LicenseType授权列表：返回内容为空": {
			suite.sqlExecutor,
			orgUUID,
			"555",
			[]int{},
		},
		"传入空串的组织id和正确的用户id，查询用户所有LicenseType授权列表：返回内容为空": {
			suite.sqlExecutor,
			"",
			userUUID,
			[]int{},
		},
		"传入正确组织id和空串的用户id，查询用户所有LicenseType授权列表：返回内容为空": {
			suite.sqlExecutor,
			orgUUID,
			"",
			[]int{},
		},
	}

	for name, tc := range data_suite {
		licenseUserGrantTypeInts, _ := license.ListUserGrantTypeInts(tc.sql, tc.orgUUID, tc.userUUID)
		assert.ElementsMatch(suite.T(), tc.expected, licenseUserGrantTypeInts, name)
	}

}

// 批量获取用户所有LicenseType授权map
func (suite *testSuite) TestMapUserGrantTypeIntsByUserUUIDs() {

	type test struct {
		sql       gorp.SqlExecutor
		orgUUID   string
		userUUIDs []string
		expected  interface{}
	}

	orgUUID := suite.orgIds[0]
	userUUIDs := suite.orgUserMap[orgUUID]

	UserGrantTypeIntMap := make(map[string]map[int]int, 0) // 多用户
	for _, userUUID := range userUUIDs {
		GrantTypeIntMap := make(map[int]int, 0)
		typeInts, _ := license.ListUserGrantTypeInts(suite.sqlExecutor, orgUUID, userUUID)
		for _, typeInt := range typeInts {
			if len(typeInts) != 0 {
				licenseType := license.GetLicenseType(typeInt)
				userGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
				GrantTypeIntMap[typeInt] = userGrant.Status
			}
		}
		UserGrantTypeIntMap[userUUID] = GrantTypeIntMap
	}

	projectStatusMap := map[int]int{license.LicenseTypeProject: 1, license.LicenseTypeWiki: 1}
	onlyUserGrantTypeIntMap := map[string]map[int]int{userUUIDs[0]: projectStatusMap} // 单用户

	data_suite := map[string]test{
		"传入正确组织id和用户id：成功，返回用户下各个licenseType的状态。": {
			suite.sqlExecutor,
			orgUUID,
			userUUIDs,
			UserGrantTypeIntMap,
		},
		"传入1个正确和1个错误的用户id：成功，只返回正确用户id下各个licenseType的状态。": {
			suite.sqlExecutor,
			orgUUID,
			[]string{"1234", userUUIDs[0]},
			onlyUserGrantTypeIntMap,
		},

		"传入错误的组织id：返回内容为空": {
			suite.sqlExecutor,
			"123auto",
			userUUIDs,
			map[string]map[int]int{},
		},
		"传入错误的用户id：返回内容为空": {
			suite.sqlExecutor,
			orgUUID,
			[]string{"123", "456"},
			map[string]map[int]int{},
		},
		"传入组织id为空串：返回内容为空": {
			suite.sqlExecutor,
			"",
			userUUIDs,
			map[string]map[int]int{}, // todo
		},
		"传入用户id为空串：返回内容为空": {
			suite.sqlExecutor,
			orgUUID,
			[]string{""},
			map[string]map[int]int{},
		},
	}

	for name, tc := range data_suite {
		mapUserGrantTypeInts, _ := license.MapUserGrantTypeIntsByUserUUIDs(tc.sql, tc.orgUUID, tc.userUUIDs)
		assert.EqualValues(suite.T(), tc.expected, mapUserGrantTypeInts, name)

	}

}

// 获取组织下指定LicenseType类型的用户授权列表
func (suite *testSuite) TestListOrgUserGrantsByType() {

	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		expected    interface{}
	}

	// 获取组织和用户的关系--调用GetUserGrantByType，组成用户集合
	orgUUID := suite.orgIds[0]
	userUUIDs := suite.orgUserMap[orgUUID]
	licenseType := license.GetLicenseType(license.LicenseTypeProject)
	orgUserGrants := []*license.LicenseUserGrant{}
	for _, userUUID := range userUUIDs {
		userGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
		orgUserGrants = append(orgUserGrants, userGrant)
	}

	data_suite := map[string]test{
		"传入正确组织id：成功，返回用户下各个licenseType的状态。": {
			suite.sqlExecutor,
			orgUUID,
			licenseType,
			orgUserGrants,
		},
		"传入非空且没有license的组织id：返回内容为空": {
			suite.sqlExecutor,
			"123",
			licenseType,
			[]*license.LicenseUserGrant{},
		},
		"传入错误的license：返回内容为空": {
			suite.sqlExecutor,
			orgUUID,
			license.GetLicenseType(100),
			[]*license.LicenseUserGrant{},
		},
		"传入组织id为空串：返回内容为空": {
			suite.sqlExecutor,
			"",
			licenseType,
			[]*license.LicenseUserGrant{},
		},
		"传入license为nil：返回内容为空": {
			suite.sqlExecutor,
			orgUUID,
			nil,
			[]*license.LicenseUserGrant{},
		},
	}

	for name, tc := range data_suite {
		userGrants, _ := license.ListOrgUserGrantsByType(tc.sql, tc.orgUUID, tc.licenseType)
		assert.ElementsMatch(suite.T(), tc.expected, userGrants, name)
	}

}

// 获取组织下所有LicenseType的数量
func (suite *testSuite) TestMapOrgLicenseGrantCount() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected interface{}
	}

	orgUUID := suite.orgIds[0]
	data_suite := map[string]test{
		"传入正确组织id：成功，返回用户下各个licenseType的状态。": {
			suite.sqlExecutor,
			orgUUID,
			suite.GetOrgLicenseTypeCount(orgUUID), // 直接查库
		},
		"传入非空且没有license的组织id：返回内容为空": {
			suite.sqlExecutor,
			"123",
			map[int]int{},
		},
		"传入组织id为空串：返回内容为空": {
			suite.sqlExecutor,
			"",
			map[int]int{},
		},
	}

	for name, tc := range data_suite {
		orgLicenseGrantCountMap, _ := license.MapOrgLicenseGrantCount(tc.sql, tc.orgUUID)
		assert.EqualValues(suite.T(), tc.expected, orgLicenseGrantCountMap, name)

	}

}

// 授予组织下某个用户对应LicenseType
func (suite *testSuite) TestGrantLicenseToUser() {

	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		userUUID    string
		licenseType license.LicenseType
		expected    interface{}
	}

	orgUUID := uuid.UUID()
	userUUID := uuid.UUID()
	licenseType := license.GetLicenseType(license.LicenseTypeProject)
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)

	userGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	if userGrant != nil {
		license.ReclaimUserGrant(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	}

	data_suite := map[string]test{
		"传入正确组织id、用户id、licenseType：成功，授予用于project应用权限": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			licenseType,
			nil,
		},
		"传入非空且没有license的orgUUID：返回报错信息": {
			suite.sqlExecutor,
			"123org",
			userUUID,
			licenseType,
			"",
		},
		"传入错误的license：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			license.GetLicenseType(100),
			"",
		},
		"传入userUUID为空串：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			"",
			licenseType,
			"",
		},
		"传入licenseType为nil：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			nil,
			"",
		},
	}

	for name, tc := range data_suite {
		err := license.GrantLicenseToUser(tc.sql, tc.orgUUID, tc.userUUID, tc.licenseType)
		if strings.Contains(name, "成功") {
			assert.Nil(suite.T(), err, name)
			licenseUserGrant, _ := license.GetUserGrantByType(tc.sql, tc.orgUUID, tc.userUUID, tc.licenseType)
			assert.EqualValues(suite.T(), licenseType, licenseUserGrant.LicenseType, name)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}

}

// 批量授予组织下多个用户对应LicenseType
func (suite *testSuite) TestBatchGrantLicenseToUsers() {

	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		userUUIDs   []string
		licenseType license.LicenseType
		expected    interface{}
	}

	orgUUID := uuid.UUID()
	userUUID01 := uuid.UUID()
	userUUID02 := uuid.UUID()
	licenseType := license.GetLicenseType(license.LicenseTypeProject)
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 1, -1)

	// 测试数据
	data_suite := map[string]test{
		"传入2个用户id，其中A应用scale为1：用户1授权成功，用户2授权失败": {
			suite.sqlExecutor,
			orgUUID,
			[]string{userUUID01, userUUID02},
			licenseType,
			[]string{userUUID02},
		},
		"传入非空且不存在的orgUUID：返回报错信息": {
			suite.sqlExecutor,
			"123org",
			[]string{userUUID01, userUUID02},
			licenseType,
			"",
		},
		"传入userUUID为空串：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			[]string{""},
			licenseType,
			"",
		},
		"传入错误的licenseType：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			[]string{userUUID01, userUUID02},
			license.GetLicenseType(100),
			"",
		},
		"传入licenseType为nil：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			[]string{userUUID01, userUUID02},
			nil,
			"",
		},
	}
	for name, tc := range data_suite {
		_, failedUsers, err := license.BatchGrantLicenseToUsers(tc.sql, tc.orgUUID, tc.userUUIDs, tc.licenseType)
		if strings.Contains(name, "失败") {
			assert.Nil(suite.T(), err, name)
			assert.Subset(suite.T(), []string{userUUID01, userUUID02}, failedUsers, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}

}

// 授予组织下某个用户多个LicenseType
func (suite *testSuite) TestGrantLicensesToUser() {

	type test struct {
		tx       *gorp.Transaction
		orgUUID  string
		userUUID string
		types    []license.LicenseType
		expected interface{}
	}

	orgUUID01 := uuid.UUID()
	userUUID01 := uuid.UUID()
	suite.ManulAddLicense(orgUUID01, license.LicenseTypeProject, license.EditionTeam, 0, -1) //设置一个 license 的 scale 上限为0
	suite.ManulAddLicense(orgUUID01, license.LicenseTypeWiki, license.EditionTeam, 10, -1)

	orgUUID02 := uuid.UUID()
	userUUID02 := uuid.UUID()
	suite.ManulAddLicense(orgUUID02, license.LicenseTypeProject, license.EditionTeam, 10, time.Now().Unix()-60) //设置一个 license 的 过期时间 为一分钟前
	suite.ManulAddLicense(orgUUID02, license.LicenseTypeWiki, license.EditionTeam, 10, -1)

	licenseTypes := []license.LicenseType{license.GetLicenseType(license.LicenseTypeProject), license.GetLicenseType(license.LicenseTypeWiki)}

	// 开启事务
	tx, err := dbm.Begin()
	if err != nil {
		fmt.Println("事务错误：", err)
		return
	}
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("%s: %s", p, debug.Stack())
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = fmt.Errorf("%s", p)
			}
		}
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	// 测试数据
	data_suite := map[string]test{
		"传入2个应用licenseType，其中A应用scale已到上限：成功，A应用授权失败，B应用授权成功": {
			tx,
			orgUUID01,
			userUUID01,
			licenseTypes,
			[]int{license.LicenseTypeProject}, // 失败授权应用
		},
		"传入2个应用licenseType，其中A应用已过期：成功，A应用授权失败，B应用授权成功": {
			tx,
			orgUUID02,
			userUUID02,
			licenseTypes,
			[]int{license.LicenseTypeProject}, // 失败授权应用
		},
		"传入非空且不存在的orgUUID：应用授权失败": {
			tx,
			"123org",
			userUUID01,
			licenseTypes,
			[]int{license.LicenseTypeProject, license.LicenseTypeWiki},
		},
		"传入2个licenseType，1个已授权和1个错误的license应用授权失败": {
			tx,
			orgUUID01,
			userUUID01,
			[]license.LicenseType{licenseTypes[0], license.GetLicenseType(100)},
			[]int{license.LicenseTypeProject, license.LicenseTypeInvalid},
		},
		"传入非空且不存在的userUUID：应用授权成功": {
			tx,
			orgUUID01,
			"123user",
			licenseTypes,
			[]int{license.LicenseTypeProject},
		},
		"传入userUUID为空串：返回报错信息": {
			tx,
			orgUUID01,
			"",
			licenseTypes,
			"",
		},
		"传入orgUUID为空串：返回报错信息": {
			tx,
			"",
			userUUID01,
			licenseTypes,
			"",
		},
		"传入licenseType为nil：返回报错信息": {
			tx,
			"",
			userUUID01,
			nil,
			"",
		},
		"传入licenseType数组为空：返回报错信息": {
			tx,
			"",
			userUUID01,
			[]license.LicenseType{},
			"",
		},
	}
	for name, tc := range data_suite {
		_, failedTypes, err := license.GrantLicensesToUser(tc.tx, tc.orgUUID, tc.userUUID, tc.types)
		if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		} else {
			assert.Nil(suite.T(), err, name)
			assert.ElementsMatch(suite.T(), tc.expected, failedTypes, name)
		}

	}

}

//回收组织下某个用户的指定LicenseType授权
func (suite *testSuite) TestReclaimUserGrant() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		userUUID string
		tpe      license.LicenseType
		expected interface{}
	}

	orgUUID := uuid.UUID()
	userUUID := uuid.UUID()
	licenseType := license.GetLicenseType(license.LicenseTypeProject)
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)

	myLicenseUserGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	if myLicenseUserGrant == nil {
		license.GrantLicenseToUser(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	}

	data_suite := map[string]test{
		"传入正确的组织id、用户id和type，回收project应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			licenseType,
			nil,
		},
		"传入错误的组织id回收授权：返回nil": {
			suite.sqlExecutor,
			"123",
			userUUID,
			licenseType,
			nil,
		},
		"传入错误的用户id回收授权：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			"555",
			licenseType,
			nil,
		},
		"传入错误的licenseType回收授权：返回报错": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			license.GetLicenseType(100),
			nil,
		},
		"传入组织id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			"",
			userUUID,
			licenseType,
			nil,
		},
		"传入用户id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			orgUUID,
			"",
			licenseType,
			nil,
		},
	}

	for name, tc := range data_suite {
		err := license.ReclaimUserGrant(tc.sql, tc.orgUUID, tc.userUUID, tc.tpe)
		if strings.Contains(name, "成功") {
			assert.Nil(suite.T(), err, name)
			licenseUserGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, tc.orgUUID, tc.userUUID, tc.tpe)
			assert.Nil(suite.T(), licenseUserGrant, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), err, name)
		}
	}

}

//回收组织下某个用户的多个LicenseType授权
func (suite *testSuite) TestReclaimUserGrants() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		userUUID string
		types    []license.LicenseType
		expected interface{}
	}

	orgUUID := uuid.UUID()
	userUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	suite.ManulAddLicense(orgUUID, license.LicenseTypeWiki, license.EditionTeam, 10, -1)

	licenseTypes := []license.LicenseType{license.GetLicenseType(license.LicenseTypeProject), license.GetLicenseType(license.LicenseTypeWiki)}
	for _, licenseType := range licenseTypes {
		license.GrantLicenseToUser(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	}

	data_suite := map[string]test{
		"传入正确的组织id、用户id、type，回收多个应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			licenseTypes,
			[]*license.LicenseUserGrant{},
		},
		"传入错误的组织id回收授权：返回nil": {
			suite.sqlExecutor,
			"123",
			userUUID,
			licenseTypes,
			nil,
		},
		"传入错误的用户id回收授权：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			"555",
			licenseTypes,
			nil,
		},
		"传入错误的licenseType回收授权：返回报错": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			[]license.LicenseType{license.GetLicenseType(100)},
			"",
		},
		"传入组织id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			"",
			userUUID,
			licenseTypes,
			nil,
		},
		"传入用户id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			orgUUID,
			"",
			licenseTypes,
			nil,
		},
	}

	for name, tc := range data_suite {
		err := license.ReclaimUserGrants(tc.sql, tc.orgUUID, tc.userUUID, tc.types)
		if strings.Contains(name, "成功") {
			licenseUserGrants, _ := license.ListUserGrants(suite.sqlExecutor, tc.orgUUID, tc.userUUID)
			assert.EqualValues(suite.T(), tc.expected, licenseUserGrants, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), err, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}

}

//回收组织下用户的所有授权
func (suite *testSuite) TestReclaimUserAllGrant() {

	type test struct {
		sql       gorp.SqlExecutor
		orgUUID   string
		userUUIDs string
		expected  interface{}
	}

	orgUUID := uuid.UUID()
	userUUID := uuid.UUID()

	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	suite.ManulAddLicense(orgUUID, license.LicenseTypeWiki, license.EditionTeam, 10, -1)
	licenseTypes := []license.LicenseType{license.GetLicenseType(license.LicenseTypeProject), license.GetLicenseType(license.LicenseTypeWiki)}
	for _, licenseType := range licenseTypes {
		license.GrantLicenseToUser(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	}

	data_suite := map[string]test{
		"传入正确的组织id和用户id，回收全部应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			[]*license.LicenseUserGrant{},
		},
		"传入错误的组织id回收授权：返回nil": {
			suite.sqlExecutor,
			"123",
			userUUID,
			nil,
		},
		"传入错误的用户id回收授权：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			"555",
			nil,
		},
		"传入组织id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			"",
			userUUID,
			nil,
		},
		"传入用户id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			orgUUID,
			"",
			nil,
		},
	}

	for name, tc := range data_suite {
		err := license.ReclaimUserAllGrant(tc.sql, tc.orgUUID, tc.userUUIDs)
		if strings.Contains(name, "成功") {
			licenseUserGrants, _ := license.ListUserGrants(suite.sqlExecutor, orgUUID, userUUID)
			assert.EqualValues(suite.T(), tc.expected, licenseUserGrants, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), err, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}
}

//批量回收组织下多个用户的指定LicenseType授权
func (suite *testSuite) TestBatchReclaimUsersGrant() {

	type test struct {
		sql       gorp.SqlExecutor
		orgUUID   string
		userUUIDs []string
		tpe       license.LicenseType
		expected  interface{}
	}

	orgUUID := uuid.UUID()
	userUUID01 := uuid.UUID()
	userUUID02 := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	licenseType := license.GetLicenseType(license.LicenseTypeProject)
	hasGrantUsers := []string{userUUID01, userUUID02}
	license.BatchGrantLicenseToUsers(suite.sqlExecutor, orgUUID, hasGrantUsers, licenseType) // 授权

	data_suite := map[string]test{
		"传入正确的组织id和用户id，回收project应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			hasGrantUsers,
			licenseType,
			nil,
		},
		"传入错误的组织id回收授权：返回nil": {
			suite.sqlExecutor,
			"123",
			hasGrantUsers,
			licenseType,
			"",
		},
		"传入错误的用户id回收授权：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			[]string{"123", "456"},
			licenseType,
			"",
		},
		"传入错误的licenseType回收授权：返回报错": {
			suite.sqlExecutor,
			orgUUID,
			hasGrantUsers,
			license.GetLicenseType(100),
			"",
		},
		"传入用户id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			orgUUID,
			[]string{""},
			licenseType,
			"",
		},
		"传入组织id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			"",
			hasGrantUsers,
			licenseType,
			"",
		},
	}

	for name, tc := range data_suite {
		err := license.BatchReclaimUsersGrant(tc.sql, tc.orgUUID, tc.userUUIDs, tc.tpe)
		if strings.Contains(name, "成功") {
			licenseUserGrant := &license.LicenseUserGrant{}
			for _, userUUID := range tc.userUUIDs {
				// 遍历user判断grant为nil
				licenseUserGrant, _ = license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
				assert.Nil(suite.T(), licenseUserGrant, name)
			}
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), err, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}

}

//批量回收组织下多个用户的多个LicenseType授权
func (suite *testSuite) TestBatchReclaimUsersGrants() {

	type test struct {
		sql       gorp.SqlExecutor
		orgUUID   string
		userUUIDs []string
		tpe       []license.LicenseType
		expected  interface{}
	}

	orgUUID := uuid.UUID()
	userUUID01 := uuid.UUID()
	userUUID02 := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	suite.ManulAddLicense(orgUUID, license.LicenseTypeWiki, license.EditionTeam, 10, -1)
	licenseTypes := []license.LicenseType{license.GetLicenseType(license.LicenseTypeProject), license.GetLicenseType(license.LicenseTypeWiki)}
	hasGrantUsers := []string{userUUID01, userUUID02}
	for _, licenseType := range licenseTypes {
		license.BatchGrantLicenseToUsers(suite.sqlExecutor, orgUUID, hasGrantUsers, licenseType) // 授权
	}

	// 测试数据
	data_suite := map[string]test{
		"传入正确的组织id、用户id和licenseTypes，回收project和wiki应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			hasGrantUsers,
			licenseTypes,
			nil,
		},
		"传入错误的组织id回收授权：返回nil": {
			suite.sqlExecutor,
			"123",
			hasGrantUsers,
			licenseTypes,
			"",
		},
		"传入错误的用户id回收授权：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			[]string{"123", "456"},
			licenseTypes,
			"",
		},
		"传入错误的licenseType回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			hasGrantUsers,
			[]license.LicenseType{license.GetLicenseType(100)},
			"",
		},
		"传入组织id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			"",
			hasGrantUsers,
			licenseTypes,
			"",
		},
		"传入用户id为空串回收授权：返回报错": {
			suite.sqlExecutor,
			orgUUID,
			[]string{""},
			licenseTypes,
			"",
		},
	}

	for name, tc := range data_suite {
		err := license.BatchReclaimUsersGrants(tc.sql, tc.orgUUID, tc.userUUIDs, tc.tpe)
		if strings.Contains(name, "成功") {
			for _, userUUID := range tc.userUUIDs {
				grants, _ := license.ListUserGrants(tc.sql, tc.orgUUID, userUUID)
				assert.Len(suite.T(), grants, 0, name)
			}
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), err, name)
		}
	}

}

// 有account的license情况下，直接查询ListUserGrantTypeInts会带有account。
func (suite *testSuite) TestHasAccountThenGetListUserGrantTypeInts() {
	orgUUID := uuid.UUID()
	userUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeAccount, license.EditionEnterprise, 0, -1)
	ints, err := license.ListUserGrantTypeInts(suite.sqlExecutor, orgUUID, userUUID)
	assert.Nil(suite.T(), err)
	assert.EqualValues(suite.T(), []int{6}, ints, "有account的license情况: 查询成功，返回数组带有account信息")
}

// 有account的license情况下，直接查询MapUserGrantTypeIntsByUserUUIDs会带有account。
func (suite *testSuite) TestHasAccountThenGetMapUserGrantTypeIntsByUserUUIDs() {
	orgUUID := uuid.UUID()
	userUUID01 := uuid.UUID()
	userUUID02 := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeAccount, license.EditionEnterprise, 0, -1)
	// map[userUUID]map[licenseTypeInt][status]
	mapUserGrantTypeIntsByUserUUIDs, err := license.MapUserGrantTypeIntsByUserUUIDs(suite.sqlExecutor, orgUUID, []string{userUUID01, userUUID02})
	myExpectedMap := make(map[string]map[int]int)
	myExpectedMap[userUUID01] = map[int]int{license.LicenseTypeAccount: 1}
	myExpectedMap[userUUID02] = map[int]int{license.LicenseTypeAccount: 1}
	assert.Nil(suite.T(), err)
	assert.EqualValues(suite.T(), myExpectedMap, mapUserGrantTypeIntsByUserUUIDs, "有account的license情况: 查询成功，返回map带有account信息")
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}
