package alter

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

type orgLicense struct {
	Org_uuid    string `db:"org_uuid"`
	Type        int    `db:"type"`
	Edition     string `db:"edition"`
	Add_type    int    `db:"add_type"`
	Scale       int    `db:"scale"`
	Expire_time int64  `db:"expire_time"`
	Update_time int64  `db:"update_time"`
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

	dbm := &gorp.DbMap{
		Db: database.DB,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "utf8mb4",
		},
	}

	// 基础数据
	suite.sqlExecutor = dbm
	suite.orgIds, _ = suite.BatchInsertLicense()
}

// teardown 关闭数据库链接
func (suite *testSuite) TearDownSuite() {
	defer suite.db.Close()
}

// 数据库批量插入license，作为初始化数据
func (suite *testSuite) BatchInsertLicense() ([]string, error) {
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
	if err != nil {
		fmt.Printf("BatchInsertLicense执行失败: %v", err)
	}
	return orgIds, err
}

// 根据alterUUID批量查询
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
		"传入不存在的alterUUID，查询alter记录：返回空记录": {
			suite.sqlExecutor,
			[]string{"1234auto"},
			map[string]*license.LicenseAlter{},
		},
	}

	for name, tc := range data_suite {
		mapLicenseAlters, _ := license.MapLicenseAltersByUUIDs(tc.sql, tc.alterUUIDs)
		assert.EqualValues(suite.T(), tc.expected, mapLicenseAlters, name)
	}

}

// 查询组织内所有LicenseAlter
func (suite *testSuite) TestListLicenseAltersByOrgUUID() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected interface{}
	}

	orgLicenseAlters := []*license.LicenseAlter{}
	for count := 1; count <= 8; count++ {
		licenseAlters, _ := license.ListLicenseAltersByOrgUUIDAndType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(count))
		if licenseAlters != nil {
			orgLicenseAlters = append(orgLicenseAlters, licenseAlters...)
		}
	}

	data_suite := map[string]test{
		"传入正确的组织id，查询alter记录：可以查到组织所有的alter记录": {
			suite.sqlExecutor,
			suite.orgIds[0],
			orgLicenseAlters,
		},
		"传入错误的组织id，查询alter记录：返回空记录": {
			suite.sqlExecutor,
			"123auto",
			[]*license.LicenseAlter{},
		},
	}

	for name, tc := range data_suite {
		licenseAlters, _ := license.ListLicenseAltersByOrgUUID(tc.sql, tc.orgUUID)
		assert.EqualValues(suite.T(), tc.expected, licenseAlters, name)
	}

}

// 查询组织内指定licenseType的所有LicenseAlter
func (suite *testSuite) TestListLicenseAltersByOrgUUIDAndType() {

	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		expected    interface{}
	}

	exitLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	exitLicenseTag := exitLicenseEntity.LicenseTag
	exitLicenseAlter, _ := license.UpdateOrgLicenseScale(suite.sqlExecutor, suite.orgIds[0], &exitLicenseTag, 888)

	data_suite := map[string]test{
		"传入正确的组织id，查询alter记录：成功，可以查到组织所有的更新scale的alter记录": {
			suite.sqlExecutor,
			suite.orgIds[0],
			exitLicenseAlter.LicenseType,
			[]*license.LicenseAlter{exitLicenseAlter},
		},
		"传入错误的组织id，查询alter记录：返回空记录": {
			suite.sqlExecutor,
			"123",
			exitLicenseAlter.LicenseType,
			[]*license.LicenseAlter{},
		}, "传入错误的licenseType，查询alter记录：返回空记录": {
			suite.sqlExecutor,
			suite.orgIds[0],
			license.GetLicenseType(100),
			[]*license.LicenseAlter{},
		},
	}

	for name, tc := range data_suite {
		licenseAlters, _ := license.ListLicenseAltersByOrgUUIDAndType(tc.sql, tc.orgUUID, tc.licenseType)
		assert.EqualValues(suite.T(), tc.expected, licenseAlters, name)
	}

}

// 查询组织内指定licenseType和Edition的所有LicenseAlter
func (suite *testSuite) TestListLicenseAltersByOrgUUIDAndTypeEdition() {

	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		edition     string
		expected    interface{}
	}
	licenseType := license.GetLicenseType(license.LicenseTypeProject)
	exitLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], licenseType)
	exitLicenseTag := exitLicenseEntity.LicenseTag
	exitLicenseAlter, _ := license.UpdateOrgLicenseScale(suite.sqlExecutor, suite.orgIds[0], &exitLicenseTag, 777)

	data_suite := map[string]test{
		"传入正确的组织id和edition，查询alter记录：成功，可以查询组织内指定licenseType和Edition的所有记录": {
			suite.sqlExecutor,
			suite.orgIds[0],
			licenseType,
			exitLicenseAlter.EditionName,
			exitLicenseAlter,
		},
		"传入错误的组织id，查询alter记录：返回空记录": {
			suite.sqlExecutor,
			"123auto",
			licenseType,
			exitLicenseAlter.EditionName,
			[]*license.LicenseAlter{},
		},
		"传入错误的licenseType，查询alter记录：返回空记录": {
			suite.sqlExecutor,
			suite.orgIds[0],
			license.GetLicenseType(100),
			exitLicenseAlter.EditionName,
			[]*license.LicenseAlter{},
		},
		"传入错误的edition，查询alter记录：返回空记录": {
			suite.sqlExecutor,
			suite.orgIds[0],
			licenseType,
			"123autotest",
			[]*license.LicenseAlter{},
		},
	}

	for name, tc := range data_suite {
		licenseAlters, _ := license.ListLicenseAltersByOrgUUIDAndTypeEdition(tc.sql, tc.orgUUID, tc.licenseType, tc.edition)
		if strings.Contains(name, "成功") {
			assert.Contains(suite.T(), licenseAlters, tc.expected, name)
		} else {
			assert.EqualValues(suite.T(), tc.expected, licenseAlters, name)
		}
	}

}

// 测试套件，批量执行测试用例
func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}
