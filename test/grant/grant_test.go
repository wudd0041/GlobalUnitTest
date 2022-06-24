package grant

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/gorp.v1"
	"license_testing/services/license"
	"math/rand"
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

// setup 初始化数据库
func (suite *testSuite) SetupSuite() {

	database, err := sqlx.Open("mysql", "onesdev:onesdev@tcp(119.23.130.213:3306)/project_u0015")
	if err != nil {
		panic(err)
	}
	err = database.Ping()
	if err != nil {
		panic(err)
	}
	suite.db = database
	fmt.Println("数据库连接成功！")

	dbm := &gorp.DbMap{
		Db: database.DB,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "utf8mb4",
		},
	}

	// 元数据
	suite.sqlExecutor = dbm
	suite.orgIds, suite.orgUserMap = suite.GetOrgAndUserIds()
	suite.Transaction = &gorp.Transaction{}

}

// teardown 关闭数据库链接
func (suite *testSuite) TearDownSuite() {
	defer suite.db.Close()
}

// 构造licenseTag·
func AddLicenseTag(typeInt int, editionName string) license.LicenseTag {
	licenseType := license.GetLicenseType(typeInt)
	licenseTag := license.LicenseTag{licenseType, editionName}
	return licenseTag
}

// 获取组织用户关系
func (suite *testSuite) GetOrgAndUserIds() ([]string, map[string][]string) {

	var orgIds []string
	err := suite.db.Select(&orgIds, "SELECT org_uuid FROM license_grant GROUP BY org_uuid ")
	if err != nil {
		fmt.Println("数据库执行失败: ", err)
		panic(err)
	}

	var userIds []string
	var orgUserMap = make(map[string][]string)
	var err2 error
	for _, orgId := range orgIds {
		err2 = suite.db.Select(&userIds, "SELECT user_uuid FROM license_grant where org_uuid = ? GROUP BY user_uuid", orgId)
		orgUserMap[orgId] = userIds
	}
	if err2 != nil {
		fmt.Println("数据库执行失败: ", err2)
		panic(err)
	}
	return orgIds, orgUserMap

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

	fmt.Printf("查询成功-组织下type-count:%v\n", TypeCountMap)
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

	licenseTypeProject := license.GetLicenseType(license.LicenseTypeProject)
	randInt := fmt.Sprintf("%d", rand.Intn(100))
	newOrgUUID01 := "autotest1-" + randInt
	newUserUUID := "dd"
	// 新增组织 -- 新增2个license -- 设置一个 license 的 scale 上限为0
	licenseTag := AddLicenseTag(license.LicenseTypeProject, license.EditionTeam)
	licenseAdd01 := license.LicenseAdd{newOrgUUID01, 10, -1, licenseTag}
	license.BatchAddLicenses(suite.sqlExecutor, license.AddTypeFree, []*license.LicenseAdd{&licenseAdd01})
	// 新增授权 -- 可以查到新增授权
	license.GrantLicenseToUser(suite.sqlExecutor, newOrgUUID01, newUserUUID, licenseTypeProject)

	// 测试数据
	data_suite := map[string]test{
		"传入2个应用licenseType，其中A应用scale已到上限：成功，A应用授权失败，B应用授权成功": {
			suite.sqlExecutor,
			newOrgUUID01,
			newUserUUID,
			licenseTypeProject,
			&license.LicenseUserGrant{licenseTag, newUserUUID, 1},
		},
		"传入错误的orgUUID：返回nil": {
			suite.sqlExecutor,
			"123org",
			newUserUUID,
			licenseTypeProject,
			nil,
		},
		"传入错误的license：返回nil": {
			suite.sqlExecutor,
			newOrgUUID01,
			newUserUUID,
			license.GetLicenseType(100),
			nil,
		},
		"传入错误的userUUID：返回nil": {
			suite.sqlExecutor,
			newOrgUUID01,
			"123user",
			licenseTypeProject,
			nil,
		},
	}
	for name, tc := range data_suite {
		userGrant, err := license.GetUserGrantByType(tc.sql, tc.orgUUID, tc.userUUID, tc.licenseType)
		if strings.Contains(name, "成功") {
			assert.Nil(suite.T(), err, name)
			assert.EqualValues(suite.T(), tc.expected, userGrant, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}

}

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
	var licenseUserGrants []*license.LicenseUserGrant
	for _, typeInt := range []int{1, 2, 3, 4, 5, 6, 7, 8} {
		licenseType := license.GetLicenseType(typeInt)
		licenseUserGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
		licenseUserGrants = append(licenseUserGrants, licenseUserGrant)
	}

	data_suite := map[string]test{
		"传入正确的组织id和用户id，查询用户所有LicenseType授权列表：成功，获取用户所有LicenseType授权列表": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			licenseUserGrants,
		}, "传入错误的组织id和正确的用户id，查询用户所有LicenseType授权列表：返回nil": {
			suite.sqlExecutor,
			"123",
			userUUID,
			nil,
		},
		"传入正确的组织id和错误的用户id，查询用户所有LicenseType授权列表：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			"555",
			nil,
		},
	}

	for name, tc := range data_suite {
		licenseUserGrants, _ := license.ListUserGrants(tc.sql, tc.orgUUID, tc.userUUID)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, licenseUserGrants, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseUserGrants, name)
		}
	}

}

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
		}, "传入错误的组织id和正确的用户id，查询用户所有LicenseType授权列表：返回nil": {
			suite.sqlExecutor,
			"123",
			userUUID,
			nil,
		},
		"传入正确的组织id和错误的用户id，查询用户所有LicenseType授权列表：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			"555",
			nil,
		},
	}

	for name, tc := range data_suite {
		licenseUserGrantTypeInts, _ := license.ListUserGrantTypeInts(tc.sql, tc.orgUUID, tc.userUUID)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, licenseUserGrantTypeInts, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseUserGrantTypeInts, name)
		}
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
	userUUIDs := suite.orgUserMap[orgUUID][:2]

	// 调用ListUserGrantTypeInts，获取user下的ints
	// 根据int，调用GetUserGrantByType，获取int-status的关系
	UserGrantTypeIntMap := make(map[string]map[int]int, 0)

	for _, userUUID := range userUUIDs {

		GrantTypeIntMap := make(map[int]int, 0)
		typeInts, _ := license.ListUserGrantTypeInts(suite.sqlExecutor, orgUUID, userUUID)
		for _, typeInt := range typeInts {
			licenseType := license.GetLicenseType(typeInt)
			userGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
			GrantTypeIntMap[typeInt] = userGrant.Status
		}

		UserGrantTypeIntMap[userUUID] = GrantTypeIntMap
	}

	onlyUserGrantTypeIntMap := UserGrantTypeIntMap
	delete(onlyUserGrantTypeIntMap, userUUIDs[1])

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
			[]string{suite.userIds[0], "1234"},
			onlyUserGrantTypeIntMap,
		},

		"传入错误的组织id：返回nil": {
			suite.sqlExecutor,
			"123",
			userUUIDs,
			"",
		},
		"传入错误的用户id：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			[]string{"123", "456"},
			"",
		},
	}

	for name, tc := range data_suite {
		mapUserGrantTypeInts, _ := license.MapUserGrantTypeIntsByUserUUIDs(tc.sql, tc.orgUUID, tc.userUUIDs)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, mapUserGrantTypeInts, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), mapUserGrantTypeInts, name)
		}
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
		"传入正确组织id和用户id：成功，返回用户下各个licenseType的状态。": {
			suite.sqlExecutor,
			orgUUID,
			licenseType,
			orgUserGrants,
		},
		"传入错误的组织id：返回nil": {
			suite.sqlExecutor,
			"123",
			licenseType,
			"",
		},
		"传入错误的license：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			license.GetLicenseType(100),
			"",
		},
	}

	for name, tc := range data_suite {
		userGrants, _ := license.ListOrgUserGrantsByType(tc.sql, tc.orgUUID, tc.licenseType)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, userGrants, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), userGrants, name)
		}
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
		"传入错误的组织id：返回nil": {
			suite.sqlExecutor,
			"123",
			"",
		},
	}

	for name, tc := range data_suite {
		orgLicenseGrantCountMap, _ := license.MapOrgLicenseGrantCount(tc.sql, tc.orgUUID)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, orgLicenseGrantCountMap, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), orgLicenseGrantCountMap, name)
		}
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

	// 调用GetUserGrantByType查询授予情况
	// 有授予则回收--授予--调用GetUserGrantByType查询授予情况
	// 无授予--授予-调用GetUserGrantByType查询授予情况

	flag := 0
	orgUUID := suite.orgIds[0]
	userUUID := suite.orgUserMap[orgUUID][0]
	licenseType := license.GetLicenseType(license.LicenseTypeProject)
	userGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	if userGrant != nil {
		flag = 1
		license.ReclaimUserGrant(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	}

	data_suite := map[string]test{
		"传入正确组织id和用户id：成功，授予用于project应用权限": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			licenseType,
			nil,
		},
		"传入错误的orgUUID：返回报错信息": {
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
		"传入错误的userUUID：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			"123user",
			licenseType,
			"",
		},
	}

	for name, tc := range data_suite {
		err := license.GrantLicenseToUser(tc.sql, tc.orgUUID, tc.userUUID, tc.licenseType)
		if strings.Contains(name, "成功") {
			assert.Nil(suite.T(), err, name)
			licenseUserGrant, _ := license.GetUserGrantByType(tc.sql, tc.orgUUID, tc.userUUID, tc.licenseType)
			assert.EqualValues(suite.T(), licenseType, licenseUserGrant.LicenseType, name)
			// 后置动作
			if flag == 0 {
				license.ReclaimUserGrant(suite.sqlExecutor, orgUUID, userUUID, licenseType)
			}
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
	}

	// 调用ListOrgUserGrantsByType查询证书1、2用户列表，获取差集
	// 差集有内容，则删除对应-再授权-调用ListOrgUserGrantsByType查询证书1、2用户列表
	// 差集无内容，则授权-调用ListOrgUserGrantsByType查询证书1、2用户列表

	userUUIDs := []string{"dd", "kb"}
	licenseTypeProject := license.GetLicenseType(license.LicenseTypeProject)

	randInt := fmt.Sprintf("%d", rand.Intn(100))
	newOrgUUID01 := "autotest1-" + randInt

	licenseAdd01 := license.LicenseAdd{newOrgUUID01, 1, -1, AddLicenseTag(license.LicenseTypeProject, license.EditionTeam)}
	license.BatchAddLicenses(suite.sqlExecutor, license.AddTypeFree, []*license.LicenseAdd{&licenseAdd01})

	// 测试数据
	data_suite := map[string]test{
		"传入2个用户id，其中A应用scale已到上限：成功，A应用授权失败，B应用授权成功": {
			suite.sqlExecutor,
			newOrgUUID01,
			userUUIDs,
			licenseTypeProject, // 失败授权应用
		},
	}
	for name, tc := range data_suite {
		_, failedUsers, err := license.BatchGrantLicenseToUsers(tc.sql, tc.orgUUID, tc.userUUIDs, tc.licenseType)
		if strings.Contains(name, "成功") {
			assert.Nil(suite.T(), err, name)
			assert.NotEmpty(suite.T(), failedUsers, name)
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

	orgUUID := suite.orgIds[0]
	userUUID := suite.orgUserMap[orgUUID][0]
	licenseTypeProject := license.GetLicenseType(license.LicenseTypeProject)
	licenseTypeWiki := license.GetLicenseType(license.LicenseTypeWiki)
	types := []license.LicenseType{licenseTypeProject, licenseTypeWiki}

	randInt := fmt.Sprintf("%d", rand.Intn(100))
	newOrgUUID01 := "autotest1-" + randInt
	newOrgUUID02 := "autotest2-" + randInt

	// 新增组织 -- 新增2个license -- 设置一个 license 的 scale 上限为0
	licenseAdd01 := license.LicenseAdd{newOrgUUID01, 0, -1, AddLicenseTag(license.LicenseTypeProject, license.EditionTeam)}
	licenseAdd02 := license.LicenseAdd{newOrgUUID01, 1, -1, AddLicenseTag(license.LicenseTypeWiki, license.EditionTeam)}
	// 新增组织 -- 新增2个license -- 设置一个 license 的 过期时间 为一分钟前
	licenseAdd03 := license.LicenseAdd{newOrgUUID02, 10, time.Now().Unix() - 60, AddLicenseTag(license.LicenseTypeProject, license.EditionTeam)}
	licenseAdd04 := license.LicenseAdd{newOrgUUID02, 10, -1, AddLicenseTag(license.LicenseTypeWiki, license.EditionTeam)}

	license.BatchAddLicenses(suite.sqlExecutor, license.AddTypeFree, []*license.LicenseAdd{&licenseAdd01, &licenseAdd02})
	license.BatchAddLicenses(suite.sqlExecutor, license.AddTypeFree, []*license.LicenseAdd{&licenseAdd03, &licenseAdd04})

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
			newOrgUUID01,
			userUUID,
			types,
			[]int{license.LicenseTypeProject}, // 失败授权应用
		},
		"传入2个应用licenseType，其中A应用已过期：成功，A应用授权失败，B应用授权成功": {
			tx,
			newOrgUUID02,
			userUUID,
			types,
			[]int{license.LicenseTypeProject}, // 失败授权应用
		},
		"传入错误的orgUUID：返回报错信息": {
			tx,
			"123org",
			userUUID,
			types,
			"",
		},
		"传入错误的license：返回报错信息": {
			tx,
			orgUUID,
			userUUID,
			[]license.LicenseType{types[0], license.GetLicenseType(100)},
			"",
		},
		"传入错误的userUUID：返回报错信息": {
			tx,
			orgUUID,
			"123user",
			types,
			"",
		},
	}
	for name, tc := range data_suite {
		_, failedTypes, err := license.GrantLicensesToUser(tc.tx, tc.orgUUID, tc.userUUID, tc.types)
		if strings.Contains(name, "成功") {
			assert.Nil(suite.T(), err, name)
			assert.EqualValues(suite.T(), tc.expected, failedTypes, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
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

	orgUUID := suite.orgIds[0]
	userUUID := suite.orgUserMap[orgUUID][0]
	licenseType := license.GetLicenseType(license.LicenseTypeProject)

	// 查询是否存在授权--不存在则授予，然后回收--存在则直接回收--查询是否存在授权
	licenseUserGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	if licenseUserGrant == nil {
		license.GrantLicenseToUser(suite.sqlExecutor, orgUUID, userUUID, licenseType)
	}

	data_suite := map[string]test{
		"传入正确的组织id和用户id，回收project应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			licenseType,
			nil,
		}, "传入错误的组织id回收授权：返回报错信息": {
			suite.sqlExecutor,
			"123",
			userUUID,
			licenseType,
			"",
		},
		"传入错误的用户id回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			"555",
			licenseType,
			"",
		}, "传入错误的licenseType回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			license.GetLicenseType(100),
			"",
		},
	}

	for name, tc := range data_suite {
		err := license.ReclaimUserGrant(tc.sql, tc.orgUUID, tc.userUUID, tc.tpe)
		if strings.Contains(name, "成功") {
			licenseUserGrant, _ := license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
			assert.Nil(suite.T(), licenseUserGrant, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
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

	orgUUID := suite.orgIds[0]
	userUUID := suite.orgUserMap[orgUUID][0]
	licenseTypes := []license.LicenseType{}

	// 查询是否存在授权--不存在则授予，然后回收--存在则直接回收--查询是否存在授权
	licenseUserGrants, _ := license.ListUserGrants(suite.sqlExecutor, orgUUID, userUUID)
	if licenseUserGrants == nil {
		licenseTypes = append(licenseTypes,
			license.GetLicenseType(license.LicenseTypeProject),
			license.GetLicenseType(license.LicenseTypeWiki))

		// todo
		license.GrantLicensesToUser(&gorp.Transaction{}, orgUUID, userUUID, licenseTypes)
	} else {
		for _, grant := range licenseUserGrants {
			licenseTypes = append(licenseTypes, grant.LicenseType)
		}
	}

	data_suite := map[string]test{
		"传入正确的组织id和用户id，回收多个应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			licenseTypes,
			nil,
		}, "传入错误的组织id回收授权：返回报错信息": {
			suite.sqlExecutor,
			"123",
			userUUID,
			licenseTypes,
			"",
		},
		"传入错误的用户id回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			"555",
			licenseTypes,
			"",
		}, "传入错误的licenseType回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			[]license.LicenseType{license.GetLicenseType(100)},
			"",
		},
	}

	for name, tc := range data_suite {
		err := license.ReclaimUserGrants(tc.sql, tc.orgUUID, tc.userUUID, tc.types)
		if strings.Contains(name, "成功") {
			licenseUserGrants, _ := license.ListUserGrants(suite.sqlExecutor, orgUUID, userUUID)
			allLicenseType := []license.LicenseType{}
			for _, grant := range licenseUserGrants {
				allLicenseType = append(allLicenseType, grant.LicenseType)
			}
			assert.Subset(suite.T(), allLicenseType, tc.types)
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

	orgUUID := suite.orgIds[0]
	userUUID := suite.orgUserMap[orgUUID][0]
	licenseTypes := []license.LicenseType{}

	// 查询是否存在授权--不存在则授予，然后回收--存在则直接回收--查询是否存在授权
	licenseUserGrants, _ := license.ListUserGrants(suite.sqlExecutor, orgUUID, userUUID)
	if licenseUserGrants == nil {
		licenseTypes = append(licenseTypes,
			license.GetLicenseType(license.LicenseTypeProject),
			license.GetLicenseType(license.LicenseTypeWiki))

		// todo
		license.GrantLicensesToUser(&gorp.Transaction{}, orgUUID, userUUID, licenseTypes)
	} else {
		for _, grant := range licenseUserGrants {
			licenseTypes = append(licenseTypes, grant.LicenseType)
		}
	}

	data_suite := map[string]test{
		"传入正确的组织id和用户id，回收全部应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			userUUID,
			nil,
		}, "传入错误的组织id回收授权：返回报错信息": {
			suite.sqlExecutor,
			"123",
			userUUID,
			"",
		},
		"传入错误的用户id回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			"555",
			"",
		},
	}

	for name, tc := range data_suite {
		err := license.ReclaimUserAllGrant(tc.sql, tc.orgUUID, tc.userUUIDs)
		if strings.Contains(name, "成功") {
			assert.Nil(suite.T(), err, name)
			licenseUserGrants, _ := license.ListUserGrants(suite.sqlExecutor, orgUUID, userUUID)
			assert.Nil(suite.T(), licenseUserGrants, name)

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

	flag := 1
	orgUUID := suite.orgIds[0]
	userUUIDs := suite.orgUserMap[orgUUID][:2]
	licenseType := license.GetLicenseType(license.LicenseTypeProject)

	// 查询指定LicenseType类型的用户授权列表，获取用户--不存在则授权
	userGrants, _ := license.ListOrgUserGrantsByType(suite.sqlExecutor, orgUUID, licenseType)
	hasGrantUsers := []string{}
	for _, grant := range userGrants {
		hasGrantUsers = append(hasGrantUsers, grant.UserUUID)
	}
	if len(hasGrantUsers) == 0 {
		flag = 0
		hasGrantUsers = append(hasGrantUsers, userUUIDs...)
		license.BatchGrantLicenseToUsers(suite.sqlExecutor, orgUUID, hasGrantUsers, licenseType)
	}

	data_suite := map[string]test{
		"传入正确的组织id和用户id，回收project应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			hasGrantUsers,
			licenseType,
			nil,
		}, "传入错误的组织id回收授权：返回报错信息": {
			suite.sqlExecutor,
			"123",
			hasGrantUsers,
			licenseType,
			"",
		},
		"传入错误的用户id回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			[]string{"123", "456"},
			licenseType,
			"",
		}, "传入错误的licenseType回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			hasGrantUsers,
			license.GetLicenseType(100),
			"",
		},
	}

	for name, tc := range data_suite {
		err := license.BatchReclaimUsersGrant(tc.sql, tc.orgUUID, tc.userUUIDs, tc.tpe)
		if strings.Contains(name, "成功") {
			// 遍历user判断grant为nil
			licenseUserGrant := &license.LicenseUserGrant{}
			for _, userUUID := range tc.userUUIDs {
				licenseUserGrant, _ = license.GetUserGrantByType(suite.sqlExecutor, orgUUID, userUUID, licenseType)
			}
			assert.Nil(suite.T(), licenseUserGrant, name)

			// 还原数据
			if flag == 1 {
				license.BatchGrantLicenseToUsers(suite.sqlExecutor, orgUUID, tc.userUUIDs, licenseType)
			}

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

	flag := 1
	orgUUID := suite.orgIds[0]
	userUUIDs := suite.orgUserMap[orgUUID][:2]
	licenseTypes := []license.LicenseType{
		license.GetLicenseType(license.LicenseTypeProject),
		license.GetLicenseType(license.LicenseTypeWiki)}

	// 查询两个应用的用户授权，取交集用户
	userGrants4Project, _ := license.ListOrgUserGrantsByType(suite.sqlExecutor, orgUUID, licenseTypes[0])
	userGrants4Wiki, _ := license.ListOrgUserGrantsByType(suite.sqlExecutor, orgUUID, licenseTypes[1])

	hasGrantUsers4Project := []string{}
	hasGrantUsers4Wiki := []string{}
	for _, grant := range userGrants4Project {
		hasGrantUsers4Project = append(hasGrantUsers4Project, grant.UserUUID)
	}
	for _, grant := range userGrants4Wiki {
		hasGrantUsers4Wiki = append(hasGrantUsers4Wiki, grant.UserUUID)
	}

	UserCount4ProjectMap := make(map[string]int)
	intersectUsers := make([]string, 0)

	for _, v := range hasGrantUsers4Project {
		UserCount4ProjectMap[v]++
	}

	for _, v := range hasGrantUsers4Wiki {
		count := UserCount4ProjectMap[v]
		if count > 0 {
			intersectUsers = append(intersectUsers, v)
		}
	}

	if len(intersectUsers) == 0 {
		flag = 0
		intersectUsers = append(intersectUsers, userUUIDs...)
		license.BatchGrantLicenseToUsers(suite.sqlExecutor, orgUUID, intersectUsers, licenseTypes[0])
		license.BatchGrantLicenseToUsers(suite.sqlExecutor, orgUUID, intersectUsers, licenseTypes[1])
	}

	// 测试数据
	data_suite := map[string]test{
		"传入正确的组织id和用户id，回收project和wiki应用权限：回收成功": {
			suite.sqlExecutor,
			orgUUID,
			intersectUsers,
			licenseTypes,
			nil,
		}, "传入错误的组织id回收授权：返回报错信息": {
			suite.sqlExecutor,
			"123",
			intersectUsers,
			licenseTypes,
			"",
		},
		"传入错误的用户id回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			[]string{"123", "456"},
			licenseTypes,
			"",
		}, "传入错误的licenseType回收授权：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			intersectUsers,
			[]license.LicenseType{license.GetLicenseType(100)},
			"",
		},
	}

	for name, tc := range data_suite {
		err := license.BatchReclaimUsersGrants(tc.sql, tc.orgUUID, tc.userUUIDs, tc.tpe)
		if strings.Contains(name, "成功") {
			for _, userUUID := range tc.userUUIDs {
				grants, _ := license.ListUserGrants(tc.sql, tc.orgUUID, userUUID)
				allLicenseType := []license.LicenseType{}
				for _, grant := range grants {
					allLicenseType = append(allLicenseType, grant.LicenseType)
				}
				assert.NotSubset(suite.T(), allLicenseType, licenseTypes, name)
			}
			// 还原数据
			if flag == 1 {
				license.BatchGrantLicenseToUsers(suite.sqlExecutor, orgUUID, tc.userUUIDs, licenseTypes[0])
				license.BatchGrantLicenseToUsers(suite.sqlExecutor, orgUUID, tc.userUUIDs, licenseTypes[1])
			}

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}

}

func (suite *testSuite) TestNa() {

}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}
