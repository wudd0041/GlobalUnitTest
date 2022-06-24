package alter

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/gorp.v1"
	"license_testing/services/license"
	"strings"
	"testing"
)

type testSuite struct {
	// 这里放全局变量
	suite.Suite
	db          *sqlx.DB
	sqlExecutor gorp.SqlExecutor
	orgIds      []string
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
	suite.orgIds = suite.GetOrgIds()

}

// 数据库查询组织ID和用户ID记录
func (suite *testSuite) GetOrgIds() []string {
	var orgIds []string
	err := suite.db.Select(&orgIds, "SELECT distinct org_uuid from license_grant order by org_uuid limit 2 ")
	if err != nil {
		fmt.Println("数据库执行失败: ", err)
		panic(err)
	}
	fmt.Printf("查询成功:%v\n", orgIds)
	return orgIds

}

// teardown 关闭数据库链接
func (suite *testSuite) TearDownSuite() {
	defer suite.db.Close()
}

func (suite *testSuite) TestMapLicenseAltersByUUIDs() {

	type test struct {
		sql        gorp.SqlExecutor
		alterUUIDs []string
		expected   interface{}
	}

	exitLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	exitLicenseTag := exitLicenseEntity.LicenseTag
	exitLicenseAlter, _ := license.UpdateOrgLicenseScale(suite.sqlExecutor, suite.orgIds[0], &exitLicenseTag, 999)
	exitUUID := exitLicenseAlter.UUID

	data_suite := map[string]test{
		"对存在的License修改scale为10，查询alter记录：可以查到修改scale的alter记录": {
			suite.sqlExecutor,
			[]string{exitUUID},
			map[string]*license.LicenseAlter{exitUUID: exitLicenseAlter},
		},
		"传入不存在的alterUUID，查询alter记录：返回nil": {
			suite.sqlExecutor,
			[]string{"1234auto"},
			nil,
		},
	}

	for name, tc := range data_suite {
		mapLicenseAlters, _ := license.MapLicenseAltersByUUIDs(tc.sql, tc.alterUUIDs)
		assert.EqualValues(suite.T(), tc.expected, mapLicenseAlters, name)
	}

}

func (suite *testSuite) TestListLicenseAltersByOrgUUID() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected interface{}
	}

	orgLicenseAlters := []*license.LicenseAlter{}
	for count := 1; count <= 8; count++ {
		licenseAlters, _ := license.ListLicenseAltersByOrgUUIDAndType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(count))
		orgLicenseAlters = append(orgLicenseAlters, licenseAlters...)
	}

	data_suite := map[string]test{
		"传入正确的组织id，查询alter记录：可以查到组织所有的alter记录": {
			suite.sqlExecutor,
			suite.orgIds[0],
			orgLicenseAlters,
		},
		"传入错误的组织id，查询alter记录：返回nil": {
			suite.sqlExecutor,
			"123auto",
			nil,
		},
	}

	for name, tc := range data_suite {
		licenseAlters, _ := license.ListLicenseAltersByOrgUUID(tc.sql, tc.orgUUID)
		assert.EqualValues(suite.T(), tc.expected, licenseAlters, name)
	}

}

func (suite *testSuite) TestListLicenseAltersByOrgUUIDAndType() {

	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		expected    interface{}
	}

	exitLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	exitLicenseTag := exitLicenseEntity.LicenseTag
	exitLicenseAlter, _ := license.UpdateOrgLicenseScale(suite.sqlExecutor, suite.orgIds[0], &exitLicenseTag, 999)

	data_suite := map[string]test{
		"传入正确的组织id，查询alter记录：成功，可以查到组织所有的更新scale的alter记录": {
			suite.sqlExecutor,
			suite.orgIds[0],
			exitLicenseAlter.LicenseType,
			exitLicenseAlter,
		},
		"传入错误的组织id，查询alter记录：返回nil": {
			suite.sqlExecutor,
			"123",
			exitLicenseAlter.LicenseType,
			nil,
		}, "传入错误的licenseType，查询alter记录：返回nil": {
			suite.sqlExecutor,
			suite.orgIds[0],
			license.GetLicenseType(100),
			nil,
		},
	}

	for name, tc := range data_suite {
		licenseAlters, _ := license.ListLicenseAltersByOrgUUIDAndType(tc.sql, tc.orgUUID, tc.licenseType)
		if strings.Contains(name, "成功") {
			assert.Contains(suite.T(), licenseAlters, tc.expected, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseAlters, name)
		}
	}

}

func (suite *testSuite) TestListLicenseAltersByOrgUUIDAndTypeEdition() {

	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		edition     string
		expected    *license.LicenseAlter
	}
	licenseType := license.GetLicenseType(license.LicenseTypeProject)
	exitLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], licenseType)
	exitLicenseTag := exitLicenseEntity.LicenseTag
	exitLicenseAlter, _ := license.UpdateOrgLicenseScale(suite.sqlExecutor, suite.orgIds[0], &exitLicenseTag, 999)

	data_suite := map[string]test{
		"传入正确的组织id，查询alter记录：成功，可以查到组织所有的更新scale的alter记录": {
			suite.sqlExecutor,
			suite.orgIds[0],
			licenseType,
			exitLicenseAlter.EditionName,
			exitLicenseAlter,
		},
	}

	for name, tc := range data_suite {
		licenseAlters, _ := license.ListLicenseAltersByOrgUUIDAndTypeEdition(tc.sql, tc.orgUUID, tc.licenseType, tc.edition)
		assert.Contains(suite.T(), licenseAlters, tc.expected, name)
	}

}

// 测试套件，批量执行测试用例
func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}
