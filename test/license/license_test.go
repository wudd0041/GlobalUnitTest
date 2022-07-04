package license

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/gorp.v1"
	"license_testing/services/license"
	"license_testing/utils/uuid"
	"strings"
	"testing"
	"time"
)

var dbm *gorp.DbMap

type testSuite struct {
	// 这里放全局变量
	suite.Suite
	db          *sqlx.DB
	sqlExecutor gorp.SqlExecutor
	orgIds      []string
	autoFlag    string
}

type orgLicenseDefaultGrant struct {
	Org_UUID      string
	Default_Grant int
	Type          int
}

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

// setup 初始化数据库
func (suite *testSuite) SetupSuite() {
	// 数据库地址
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

	dbm := &gorp.DbMap{
		Db: database.DB,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "utf8mb4",
		},
	}

	// 基础数据
	suite.sqlExecutor = dbm
	suite.orgIds, _ = suite.BatchInsertLicensAndGrant()

}

// teardown 关闭数据库链接
func (suite *testSuite) TearDownSuite() {
	defer suite.db.Close()
}

func (suite *testSuite) SetupTest() {
	// 设置唯一标识位
	timeStamp := fmt.Sprintf("%d", time.Now().UnixNano())
	suite.autoFlag = "autoTest" + timeStamp
}

// 数据库批量插入license、licenseGrant，作为初始化数据
func (suite *testSuite) BatchInsertLicensAndGrant() ([]string, error) {
	orgId1 := uuid.UUID()
	orgId2 := uuid.UUID()
	orgIds := []string{orgId1, orgId2}
	l1 := orgLicense{orgId1, 1, license.EditionEnterpriseTrial, 0, 100, -1, 1655714729}
	l2 := orgLicense{orgId1, 2, license.EditionEnterpriseTrial, 0, 100, -1, 1655714729}
	l3 := orgLicense{orgId2, 1, license.EditionEnterpriseTrial, 0, 100, -1, 1655714729}
	l4 := orgLicense{orgId2, 2, license.EditionEnterpriseTrial, 0, 100, -1, 1655714729}
	license := []*orgLicense{&l1, &l2, &l3, &l4}
	_, err := suite.db.NamedExec("INSERT INTO license (org_uuid,type,edition,add_type,scale,expire_time,update_time) "+
		"VALUES (:org_uuid,:type,:edition,:add_type,:scale,:expire_time,:update_time)", license)

	userId1 := uuid.UUID()
	userId2 := uuid.UUID()
	userId3 := uuid.UUID()
	userId4 := uuid.UUID()
	ug1 := licenseGrant{orgId1, 1, userId1, 1}
	ug2 := licenseGrant{orgId1, 1, userId2, 1}
	ug3 := licenseGrant{orgId1, 2, userId1, 1}
	ug4 := licenseGrant{orgId1, 2, userId2, 1}
	ug5 := licenseGrant{orgId2, 1, userId3, 1}
	ug6 := licenseGrant{orgId2, 1, userId4, 1}
	ug7 := licenseGrant{orgId2, 2, userId3, 1}
	ug8 := licenseGrant{orgId2, 2, userId4, 1}
	licenseGrant := []*licenseGrant{&ug1, &ug2, &ug3, &ug4, &ug5, &ug6, &ug7, &ug8}
	_, err2 := suite.db.NamedExec("INSERT INTO license_grant (org_uuid, type, user_uuid, status) "+
		"VALUES (:org_uuid, :type, :user_uuid, :status)", licenseGrant)

	if err != nil {
		fmt.Printf("BatchInsertLicense执行失败: %v", err)
	} else if err2 != nil {
		fmt.Printf("BatchInsertLicense执行失败: %v", err2)
	}

	return orgIds, err
}

func (suite *testSuite) ManulAddLicense(orgUUID string, typeInt int, edition string, scale int, expire_time int64) {
	_, err := license.AddLicense(suite.sqlExecutor, license.AddTypePay, &license.LicenseAdd{orgUUID, scale, expire_time, AddLicenseTag(typeInt, edition)})
	if err != nil {
		fmt.Println("构造license发生错误：", err)
		panic(err)
	}
}

// 数据库查询license_default_grant
func (suite *testSuite) GetLicenseDefaultGrant(orgUUID string) []*license.LicenseDefaultGrant {

	var orgLicenseDefaultGrants []orgLicenseDefaultGrant
	err := suite.db.Select(&orgLicenseDefaultGrants, "SELECT org_uuid,default_grant,type from license_default_grant  where org_uuid =? ", orgUUID)
	if err != nil {
		fmt.Println("数据库执行失败: ", err)
		panic(err)
	}
	fmt.Printf("查询成功-表license_default_grant:%v\n", orgLicenseDefaultGrants)

	licenseDefaultGrants := []*license.LicenseDefaultGrant{}
	for _, v := range orgLicenseDefaultGrants {
		licenseDefaultGrant := license.LicenseDefaultGrant{OrgUUID: v.Org_UUID, LicenseType: license.GetLicenseType(v.Type)}
		if v.Default_Grant == 1 {
			licenseDefaultGrant.DefaultGrant = license.DefaultGrantTrue
		} else if v.Default_Grant == 0 {
			licenseDefaultGrant.DefaultGrant = license.DefaultGrantFalse
		}
		licenseDefaultGrants = append(licenseDefaultGrants, &licenseDefaultGrant)
	}

	return licenseDefaultGrants

}

// 构造licenseTag·
func AddLicenseTag(typeInt int, editionName string) license.LicenseTag {
	licenseType := license.GetLicenseType(typeInt)
	licenseTag := license.LicenseTag{licenseType, editionName}
	return licenseTag
}

// 添加一个License, 若LicenseTag已存在，则报错
func (suite *testSuite) TestAddLicense() {

	type test struct {
		sql        gorp.SqlExecutor
		addType    int
		addLicense *license.LicenseAdd
		expected   interface{}
	}

	// 获取已有应用的licenseTag
	licenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	licenseTag := licenseEntity.LicenseTag

	data_suite := map[string]test{
		"传入完全不存在的licenseTag: 添加证书成功": {suite.sqlExecutor,
			license.AddTypePay,
			&license.LicenseAdd{
				suite.orgIds[0],
				100,
				-1,
				AddLicenseTag(license.LicenseTypeProject, license.EditionEnterprise), // edition格式：auto+时间戳
			},
			AddLicenseTag(license.LicenseTypeProject, license.EditionEnterprise),
		},
		"传入已经存在的licenseTag: 返回报错信息": {
			suite.sqlExecutor,
			license.AddTypePay,
			&license.LicenseAdd{
				suite.orgIds[0],
				100,
				-1,
				licenseTag,
			},
			"",
		},
	}

	for name, tc := range data_suite {
		result, err := license.AddLicense(tc.sql, tc.addType, tc.addLicense)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, result.LicenseTag, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}

}

// 批量添加多个License, 若其中一个LicenseTag已存在，则报错
func (suite *testSuite) TestBatchAddLicenses() {

	type test struct {
		sql        gorp.SqlExecutor
		addType    int
		addLicense []*license.LicenseAdd
		expected   interface{}
	}

	// 获取已有应用的licenseTag
	licenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	licenseTag := licenseEntity.LicenseTag
	existOrgUUID := suite.orgIds[0]

	data_suite := map[string]test{
		"传入2个完全不存在的licenseTag：添加证书成功": {suite.sqlExecutor,
			license.AddTypePay,
			[]*license.LicenseAdd{
				{
					existOrgUUID,
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypeTestCase, license.EditionEnterprise),
				},
				{
					existOrgUUID,
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypePerformance, license.EditionEnterpriseTrial),
				},
			},
			[]license.LicenseTag{
				AddLicenseTag(license.LicenseTypeTestCase, license.EditionEnterprise),
				AddLicenseTag(license.LicenseTypePerformance, license.EditionEnterpriseTrial),
			},
		},
		"传入2个重复且不存在的licenseTag：返回报错信息": {suite.sqlExecutor,
			license.AddTypePay,
			[]*license.LicenseAdd{
				{
					existOrgUUID,
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypePipeline, license.EditionEnterprise),
				},
				{
					existOrgUUID,
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypePipeline, license.EditionEnterprise),
				},
			},
			"",
		},
		"传入1个不存在和1个存在的licenseTag：返回报错信息": {suite.sqlExecutor,
			license.AddTypePay,
			[]*license.LicenseAdd{
				{
					existOrgUUID,
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypePlan, license.EditionEnterprise),
				},
				{
					existOrgUUID,
					100,
					license.UnLimitExpire,
					licenseTag,
				},
			},
			"",
		},
		"传入1个不存在addType：返回报错信息": {suite.sqlExecutor,
			1000,
			[]*license.LicenseAdd{
				{
					existOrgUUID,
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypeDesk, license.EditionEnterprise),
				},
			},
			"",
		},
	}

	for name, tc := range data_suite {
		_, err := license.BatchAddLicenses(tc.sql, tc.addType, tc.addLicense)

		if strings.Contains(name, "传入2个完全不存在的licenseTag：添加证书成功") {
			testcaseLicenseEntity, _ := license.GetOrgLicenseByType(tc.sql, existOrgUUID, license.GetLicenseType(license.LicenseTypeTestCase))
			performanceLicenseEntity, _ := license.GetOrgLicenseByType(tc.sql, existOrgUUID, license.GetLicenseType(license.LicenseTypePerformance))

			orgLicenseTags := []license.LicenseTag{}
			orgLicenseTags = append(orgLicenseTags, testcaseLicenseEntity.LicenseTag, performanceLicenseEntity.LicenseTag)
			assert.Subset(suite.T(), orgLicenseTags, tc.expected, name)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)

		}

	}

}

// 更新组织下对应LicenseTag的最大授予人数
// 若组织下不存在该LicenseTag则报错
func (suite *testSuite) TestUpdateOrgLicenseScale() {
	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		tag      *license.LicenseTag
		newScale int
		expected interface{}
	}

	orgUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	exitLicenseTag := AddLicenseTag(license.LicenseTypeProject, license.EditionTeam)
	notExitLicenseTag := AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1")

	data_suite := map[string]test{
		"对存在的License修改scale为10：修改成功，scale为5": {
			suite.sqlExecutor,
			orgUUID,
			&exitLicenseTag,
			5,
			5,
		},
		"对不存在的License修改scale为10：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			&notExitLicenseTag,
			10,
			"",
		},
		"对不存在的orgUUID修改scale为10：返回nil": {
			suite.sqlExecutor,
			"不存在的orgUUID",
			&exitLicenseTag,
			10,
			"",
		},
	}

	for name, tc := range data_suite {
		licenseAlter, err := license.UpdateOrgLicenseScale(tc.sql, tc.orgUUID, tc.tag, tc.newScale)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, licenseAlter.New.Scale, name)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)

		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseAlter, name)
		}
	}

}

// 批量更新组织下多个LicenseTag的最大授予人数
// 若组织下不存在某个LicenseTag则报错
func (suite *testSuite) TestBatchUpdateOrgLicenseScale() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		tag      []*license.LicenseTag
		newScale int
		expected interface{}
	}

	orgUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	suite.ManulAddLicense(orgUUID, license.LicenseTypeWiki, license.EditionTeam, 10, -1)
	exitLicenseTag01 := AddLicenseTag(license.LicenseTypeProject, license.EditionTeam)
	exitLicenseTag02 := AddLicenseTag(license.LicenseTypeWiki, license.EditionTeam)
	exitLicenseTags := []*license.LicenseTag{&exitLicenseTag01, &exitLicenseTag02}
	notExitLicenseTag := AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1")

	data_suite := map[string]test{
		"对存在的2个License修改scale为20，修改成功": {
			suite.sqlExecutor,
			orgUUID,
			exitLicenseTags,
			20,
			20,
		},
		"对1个不存在的licenseTag和1个存在的licenseTag修改scale为30：返回nil": {
			suite.sqlExecutor,
			suite.orgIds[0],
			[]*license.LicenseTag{
				&exitLicenseTag01,
				&notExitLicenseTag,
			},
			30,
			nil,
		},
		"对1个不存在的orgUUID修改scale为30：返回nil": {
			suite.sqlExecutor,
			"1234org",
			exitLicenseTags,
			40,
			nil,
		},
	}

	for name, tc := range data_suite {
		licenseAlters, _ := license.BatchUpdateOrgLicenseScale(tc.sql, tc.orgUUID, tc.tag, tc.newScale)
		if strings.Contains(name, "成功") {
			for _, v := range licenseAlters {
				assert.EqualValues(suite.T(), tc.expected, v.New.Scale, name)
			}
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseAlters, name)
		}

	}

}

// 更新组织下对应LicenseTag的过期时间
// 若组织下不存在该LicenseTag则报错
func (suite *testSuite) TestRenewalOrgLicenseExpire() {
	type test struct {
		sql        gorp.SqlExecutor
		orgUUID    string
		tag        *license.LicenseTag
		expireTime int64
		expected   int64
	}

	orgUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	exitLicenseTag := AddLicenseTag(license.LicenseTypeProject, license.EditionTeam)
	notExitLicenseTag := AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1")
	timeStamp := time.Now().Unix() + 600

	data_suite := map[string]test{
		"更新组织过期时间为当前时间+1分钟后：修改过期时间成功": {
			suite.sqlExecutor,
			orgUUID,
			&exitLicenseTag,
			timeStamp,
			timeStamp,
		},
		"更新组织不存在license的过期时间为当前时间+1分钟：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			&notExitLicenseTag,
			timeStamp,
			timeStamp,
		},
		"更新不存在的组织license过期时间为当前时间+1分钟：返回报错信息": {
			suite.sqlExecutor,
			"123org",
			&exitLicenseTag,
			timeStamp,
			timeStamp,
		},
	}

	for name, tc := range data_suite {
		licenseAlter, err := license.RenewalOrgLicenseExpire(tc.sql, tc.orgUUID, tc.tag, tc.expireTime)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, licenseAlter.New.ExpireTime, name)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)

		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseAlter, name)
		}
	}

}

// 批量更新组织下多个LicenseTag的过期时间
// 若组织下不存在某个LicenseTag则报错----
func (suite *testSuite) TestBatchRenewalOrgLicenseExpire() {
	type test struct {
		sql        gorp.SqlExecutor
		orgUUID    string
		tag        []*license.LicenseTag
		expireTime int64
		expected   interface{}
	}

	orgUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	suite.ManulAddLicense(orgUUID, license.LicenseTypeWiki, license.EditionTeam, 10, -1)
	tag01 := AddLicenseTag(license.LicenseTypeProject, license.EditionTeam)
	tag02 := AddLicenseTag(license.LicenseTypeWiki, license.EditionTeam)
	exitLicenseTags := []*license.LicenseTag{&tag01, &tag02}
	notExitLicenseTag := AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1")

	data_suite := map[string]test{
		"更新组织的2个license过期时间为当前时间+1分钟后：修改过期时间成功": {
			suite.sqlExecutor,
			orgUUID,
			exitLicenseTags,
			time.Now().Unix() + 60,
			[]*license.LicenseAlter{
				{New: &license.LicenseAlterInfo{ExpireTime: time.Now().Unix() + 60}},
				{New: &license.LicenseAlterInfo{ExpireTime: time.Now().Unix() + 60}},
			},
		}, "分别更新组织的1个存在和不存在的license过期时间为当前时间+1分钟后：返回报错信息": {
			suite.sqlExecutor,
			orgUUID,
			[]*license.LicenseTag{
				exitLicenseTags[0],
				&notExitLicenseTag,
			},
			time.Now().Unix() + 120,
			"",
		},
		"更新不存在的组织license过期时间为当前时间+1分钟后：返回报错信息": {
			suite.sqlExecutor,
			"123auto",
			exitLicenseTags,
			time.Now().Unix() + 180,
			"",
		},
	}

	for name, tc := range data_suite {
		licenseExpireTimes := []*license.LicenseAlter{}
		licenseAlters, err := license.BatchRenewalOrgLicenseExpire(tc.sql, tc.orgUUID, tc.tag, tc.expireTime)

		if strings.Contains(name, "成功") {
			for _, v := range licenseAlters {
				licenseExpireTimes = append(licenseExpireTimes, &license.LicenseAlter{New: &license.LicenseAlterInfo{ExpireTime: v.New.ExpireTime}})
			}
			assert.EqualValues(suite.T(), tc.expected, licenseExpireTimes, name)

		} else if strings.Contains(name, "分别更新组织的1个存在和不存在的license过期时间为当前时间+1分钟后：返回报错信息") {
			exitLicenseEntitys, _ := license.GetOrgLicenses(suite.sqlExecutor, tc.orgUUID)
			assert.NotEqualValues(suite.T(), tc.expireTime, exitLicenseEntitys[0].ExpireTime, name)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)

		}
	}

}

// 新增或者更新对应LicenseTag的默认授权配置
func (suite *testSuite) TestAddOrUpdateOrgDefaultGrant() {

	type test struct {
		sql          gorp.SqlExecutor
		orgUUID      string
		defaultGrant *license.LicenseDefaultGrant
		expected     interface{}
	}

	orgUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	defaultGrant := license.LicenseDefaultGrant{
		orgUUID,
		license.DefaultGrantTrue,
		license.GetLicenseType(license.LicenseTypeProject),
	}

	// 测试数据
	data_suite := map[string]test{
		"新增组织的1个license默认授权配置：修改授权配置成功": {
			suite.sqlExecutor,
			suite.orgIds[0],
			&defaultGrant,
			license.DefaultGrantTrue,
		},
		"新增不存在的组织1个license默认授权配置：返回nil": {
			suite.sqlExecutor,
			"123",
			&defaultGrant,
			nil,
		},
	}

	for name, tc := range data_suite {
		err := license.AddOrUpdateOrgDefaultGrant(tc.sql, tc.orgUUID, tc.defaultGrant)
		if strings.Contains(name, "成功") {
			licenseDefaultGrants, _ := license.GetOrgDefaultGrantLicenses(suite.sqlExecutor, tc.orgUUID)
			for _, v := range licenseDefaultGrants {
				if v.LicenseType == tc.defaultGrant.LicenseType {
					assert.True(suite.T(), v.DefaultGrant, name)
				}
			}
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), err, name)
		}

	}

}

// 批量新增或者更新组织下多个LicenseTag的默认授权配置--
func (suite *testSuite) TestBatchAddOrUpdateOrgDefaultGrants() {
	type test struct {
		sql           gorp.SqlExecutor
		orgUUID       string
		defaultGrants []*license.LicenseDefaultGrant
		expected      interface{}
	}

	orgUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	defaultGrants := []*license.LicenseDefaultGrant{
		{
			orgUUID,
			license.DefaultGrantTrue,
			license.GetLicenseType(license.LicenseTypeProject),
		},
	}

	// 测试数据
	data_suite := map[string]test{
		"新增组织的1个license默认授权配置:修改授权配置成功": {
			suite.sqlExecutor,
			orgUUID,
			defaultGrants,
			license.DefaultGrantTrue,
		},
		"新增不存在组织的1个license默认授权配置:返回nil信息": {
			suite.sqlExecutor,
			"123auto",
			defaultGrants,
			nil,
		},
	}

	for name, tc := range data_suite {
		err := license.BatchAddOrUpdateOrgDefaultGrants(tc.sql, tc.orgUUID, tc.defaultGrants)
		if strings.Contains(name, "成功") {
			licenseDefaultGrants, _ := license.GetOrgDefaultGrantLicenses(suite.sqlExecutor, tc.orgUUID)
			for _, v := range licenseDefaultGrants {
				if v.LicenseType == tc.defaultGrants[0].LicenseType {
					assert.EqualValues(suite.T(), tc.expected, v.DefaultGrant, name)
				}
			}
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), err, name)
		}

	}

}

// 获取组织下所有LicenseTag的默认授权配置
func (suite *testSuite) TestGetOrgDefaultGrantLicenses() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected interface{}
	}

	// 测试数据
	data_suite := map[string]test{
		"查询组织所有license默认授权配置：查询成功": {
			suite.sqlExecutor,
			suite.orgIds[0],
			suite.GetLicenseDefaultGrant(suite.orgIds[0]),
		},
		"查询不存在组织的默认授权配置：返回nil": {
			suite.sqlExecutor,
			"123",
			nil,
		},
	}

	for name, tc := range data_suite {
		licenseDefaultGrants, err := license.GetOrgDefaultGrantLicenses(tc.sql, tc.orgUUID)

		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, licenseDefaultGrants, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), err, name)
		}

	}

}

// 获取组织下某个时间点装配的LicenseType实体信息, 无授权则返回nil
func (suite *testSuite) TestGetOrgLicenseByTypeAndTimestamp() {
	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		sec         int64
		expected    interface{}
	}

	timeStamp := time.Now().Unix()
	orgUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, timeStamp+60)
	licenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, orgUUID, license.GetLicenseType(license.LicenseTypeProject))

	// 测试数据
	data_suite := map[string]test{
		"传入正确组织UUID和时间戳：查询成功，可查到该组织下过期时间大于传入时间戳的license": {
			suite.sqlExecutor,
			orgUUID,
			license.GetLicenseType(license.LicenseTypeProject),
			timeStamp,
			licenseEntity,
		},
		"传入错误组织UUID：返回nil": {
			suite.sqlExecutor,
			"123auto",
			license.GetLicenseType(license.LicenseTypeProject),
			timeStamp,
			nil,
		},
		"传入不存在的licenseType：返回nil": {
			suite.sqlExecutor,
			orgUUID,
			license.GetLicenseType(100),
			timeStamp,
			nil,
		},
	}

	for name, tc := range data_suite {
		actualLicenseEntity, _ := license.GetOrgLicenseByTypeAndTimestamp(tc.sql, tc.orgUUID, tc.licenseType, tc.sec)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, actualLicenseEntity, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), actualLicenseEntity, name)
		}
	}
}

// 查询一段时间区间内会过期的LicenseEntity
func (suite *testSuite) TestListExpireInTimeStampRange() {

	type test struct {
		sql        gorp.SqlExecutor
		startStamp int64
		endStamp   int64
		expected   interface{}
	}

	startStamp := time.Now().Unix() - 600
	endStamp := time.Now().Unix() + 600
	greaterStartStamp := endStamp + 600

	// 获取licenseTag
	orgUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 0, time.Now().Unix()+300)
	exitLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, orgUUID, license.GetLicenseType(license.LicenseTypeProject))

	// 测试数据
	data_suite := map[string]test{
		"传入起始和终止时间：查询成功，查询所有组织下过期时间在范围时间内的license": {
			suite.sqlExecutor,
			startStamp,
			endStamp,
			exitLicenseEntity,
		},
		"传入起始大于于终止时间：返回内容为空": {
			suite.sqlExecutor,
			greaterStartStamp,
			endStamp,
			[]*license.LicenseEntity{},
		},
	}

	for name, tc := range data_suite {
		licenseEntities, _ := license.ListExpireInTimeStampRange(tc.sql, tc.startStamp, tc.endStamp)
		if strings.Contains(name, "成功") {
			assert.Contains(suite.T(), licenseEntities, tc.expected, name)
		} else if strings.Contains(name, "内容为空") {
			assert.EqualValues(suite.T(), tc.expected, licenseEntities, name)
		}
	}
}

// 获取组织下当前已经装配的所有Licenses
func (suite *testSuite) TestGetOrgLicenses() {
	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected interface{}
	}

	// todo 手工创建
	orgLicenses := []*license.LicenseEntity{}
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 8} {
		licenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(v))
		if licenseEntity != nil {
			orgLicenses = append(orgLicenses, licenseEntity)
		}
	}

	// 测试数据
	data_suite := map[string]test{
		"传入正确组织UUID：查询成功，可查到该组织下所有最高优先级的license": {
			suite.sqlExecutor,
			suite.orgIds[0],
			orgLicenses,
		},
		"传入错误组织UUID：返回空对象": {
			suite.sqlExecutor,
			"123auto",
			[]*license.LicenseEntity{},
		},
	}

	for name, tc := range data_suite {
		licenseEntities, _ := license.GetOrgLicenses(tc.sql, tc.orgUUID)
		assert.ElementsMatch(suite.T(), tc.expected, licenseEntities, name)
	}
}

// 获取组织下当前已经装配的所有Licenses, map[LicenseTypeInt]*LicenseEntity 返回
func (suite *testSuite) TestGetOrgLicensesMap() {
	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected interface{}
	}
	orgLicensesMap := map[int]*license.LicenseEntity{}
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 8} {
		t, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(v))
		if t != nil {
			orgLicensesMap[v] = t
		}
	}

	// 测试数据
	data_suite := map[string]test{
		"传入正确组织UUID：查询成功，可查到该组织下每个应用下的最高优先级的license": {
			suite.sqlExecutor,
			suite.orgIds[0],
			orgLicensesMap,
		},
		"传入错误组织UUID：返回内容为空": {
			suite.sqlExecutor,
			"123auto",
			map[int]*license.LicenseEntity{},
		},
	}

	for name, tc := range data_suite {
		licenseEntitiesMap, _ := license.GetOrgLicensesMap(tc.sql, tc.orgUUID)
		assert.EqualValues(suite.T(), tc.expected, licenseEntitiesMap, name)
	}
}

// 批量获取组织下当前已经装配的所有Licenses, map[orgUUID]map[LicenseTypeInt]*LicenseEntity 返回
func (suite *testSuite) TestBatchGetOrgLicensesMaps() {
	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  []string
		expected interface{}
	}

	orgUUID01 := uuid.UUID()
	suite.ManulAddLicense(orgUUID01, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	suite.ManulAddLicense(orgUUID01, license.LicenseTypeProject, license.EditionEnterprise, 10, -1)

	orgUUID02 := uuid.UUID()
	suite.ManulAddLicense(orgUUID02, license.LicenseTypeProject, license.EditionTeam, 10, -1)
	orgUUIDs := []string{orgUUID01, orgUUID02}

	orgLicensesMap := map[string]map[int]*license.LicenseEntity{}
	orgLicensesMap[orgUUID01], _ = license.GetOrgLicensesMap(suite.sqlExecutor, orgUUID01)
	orgLicensesMap[orgUUID02], _ = license.GetOrgLicensesMap(suite.sqlExecutor, orgUUID02)

	otherOrgLicensesMap := make(map[string]map[int]*license.LicenseEntity)
	otherOrgLicensesMap[orgUUID01], _ = license.GetOrgLicensesMap(suite.sqlExecutor, orgUUID01)

	// 测试数据
	data_suite := map[string]test{
		"传入正确2个组织UUID：查询成功，可查到每个组织下每个应用下的最高优先级的license": {
			suite.sqlExecutor,
			orgUUIDs,
			orgLicensesMap,
		},
		"传入1个正确和1个错误的组织UUID：查询成功，只查到正确组织对应的license": {
			suite.sqlExecutor,
			[]string{orgUUID01, "123"},
			otherOrgLicensesMap,
		},
		"传入1个错误的组织UUID：返回内容为空": {
			suite.sqlExecutor,
			[]string{"123"},
			map[string]map[int]*license.LicenseEntity{},
		},
	}

	for name, tc := range data_suite {
		orgLicensesMaps, _ := license.BatchGetOrgLicensesMaps(tc.sql, tc.orgUUID)
		assert.EqualValues(suite.T(), tc.expected, orgLicensesMaps, name)

	}
}

// 获取组织下所有LicenseEntity(包含过期, 一般业务不关注) map返回
func (suite *testSuite) TestGetOrgAllLicensesMap() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected interface{}
	}

	orgUUID := uuid.UUID()
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionEnterpriseTrial, 100, -1)

	// 测试数据
	data_suite := map[string]test{
		"传入正确的组织UUID：查询成功，可查到project应用存在新配置的license": {
			suite.sqlExecutor,
			orgUUID,
			license.EditionEnterpriseTrial,
		},
		"传入错误的组织UUID：返回空对象": {
			suite.sqlExecutor,
			"123auto",
			map[int][]*license.LicenseEntity{},
		},
	}

	for name, tc := range data_suite {
		orgAllLicensesMap, _ := license.GetOrgAllLicensesMap(tc.sql, tc.orgUUID)
		if strings.Contains(name, "成功") {
			licenseEdition := orgAllLicensesMap[license.LicenseTypeProject][0].EditionName
			assert.EqualValues(suite.T(), tc.expected, licenseEdition, name)
		} else {
			assert.EqualValues(suite.T(), tc.expected, orgAllLicensesMap, name)
		}

	}

}

// 获取组织下某种类型的所有LicenseEntity(包含过期, 一般业务不关注)
func (suite *testSuite) TestGetOrgAllLicensesByType() {
	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		expected    interface{}
	}

	orgAllLicensesMap, _ := license.GetOrgAllLicensesMap(suite.sqlExecutor, suite.orgIds[0])

	// 测试数据
	data_suite := map[string]test{
		"传入正确的组织UUID和project应用类型：查询成功，可查到该组织project应用所有license": {
			suite.sqlExecutor,
			suite.orgIds[0],
			license.GetLicenseType(license.LicenseTypeProject),
			orgAllLicensesMap[license.LicenseTypeProject],
		},
		"传入错误的组织UUID：返回内容为空": {
			suite.sqlExecutor,
			"123",
			license.GetLicenseType(license.LicenseTypeProject),
			[]*license.LicenseEntity{},
		},
		"传入错误的应用类型LicenseType：返回报错": {
			suite.sqlExecutor,
			suite.orgIds[0],
			license.GetLicenseType(100),
			"",
		},
	}

	for name, tc := range data_suite {
		orgAllLicenses, err := license.GetOrgAllLicensesByType(tc.sql, tc.orgUUID, tc.licenseType)
		if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		} else {

			assert.EqualValues(suite.T(), tc.expected, orgAllLicenses, name)
		}
	}

}

// 获取组织下当前装配的LicenseType实体信息, 无授权则返回nil
func (suite *testSuite) TestGetOrgLicenseByType() {

	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		expected    interface{}
	}

	orgUUID := uuid.UUID()
	licenseType := license.GetLicenseType(license.LicenseTypeProject)
	suite.ManulAddLicense(orgUUID, license.LicenseTypeProject, license.EditionTeam, 10, -1)

	data_suite := map[string]test{
		"传入project的licenseType: 查询成功，返回应用免费版licenseEntity,scale为10": {suite.sqlExecutor,
			orgUUID,
			licenseType,
			10},

		"传入错误的orgUUID：返回nil": {suite.sqlExecutor,
			orgUUID + "123",
			licenseType,
			nil},

		"orgUUID为空串：返回nil": {suite.sqlExecutor,
			"",
			licenseType,
			nil},

		"licenseType为nil：返回nil": {suite.sqlExecutor,
			orgUUID,
			nil,
			nil},
	}

	for name, tc := range data_suite {
		result, _ := license.GetOrgLicenseByType(tc.sql, tc.orgUUID, tc.licenseType)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), result.Scale, tc.expected, name)
		} else {
			assert.Nil(suite.T(), result, name)
		}

	}

}

// 测试套件，批量执行测试用例
func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}
