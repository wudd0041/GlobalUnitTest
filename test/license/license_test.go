package license

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/gorp.v1"
	"license_testing/services/license"
	"math"
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
	licenses    map[string][]License
	autoFlag    string
}

type License struct {
	Org_UUID    string
	Type        int
	Edition     string
	Add_Type    int
	Scale       int
	Expire_Time int64
}

type orgLicenseDefaultGrant struct {
	Org_UUID      string
	Default_Grant int
	Type          int
}

// setup 初始化数据库
func (suite *testSuite) SetupSuite() {
	// 数据库地址
	database, err := sqlx.Open("mysql", "root:admin123456@tcp(127.0.0.1:3306)/test")
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
	suite.orgIds = suite.GetOrgIds()

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

// 数据库查询组织ID和用户ID记录
func (suite *testSuite) GetOrgIds() []string {
	var orgIds []string
	err := suite.db.Select(&orgIds, "SELECT distinct org_uuid from license_grant order by org_uuid limit 2 ")
	if err != nil {
		fmt.Println("数据库执行失败: ", err)
		panic(err)
	}
	fmt.Printf("查询成功-组织id:%v\n", orgIds)
	return orgIds

}

//数据库查询组织license记录
func (suite *testSuite) GetOrgLicense(orgUUID string) (OrgLicense []License) {
	var orgLicenses []License
	err := suite.db.Select(&orgLicenses, "SELECT org_uuid,type,edition,add_type,scale,expire_time from license where org_uuid = ? ", orgUUID)
	if err != nil {
		fmt.Println("数据库执行失败: ", err)
		panic(err)
	}
	fmt.Printf("查询成功-组织license:%v\n", orgLicenses)
	return orgLicenses

}

// 数据库删除license
func (suite *testSuite) DeleteLicense(t []interface{}) {

	sqlStr := "delete from license where edition IN (?) "
	result, err := suite.db.Exec(sqlStr, t...)
	rows, _ := result.RowsAffected()
	if err != nil {
		fmt.Println("数据库执行失败: ", err)
		panic(err)
	}
	fmt.Printf("删除成功-影响行数:%v\n", rows)

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

//构造LicenseEntity
func (suite *testSuite) GetLicenseEntity(orgUUID string, typeIntOrEdition ...interface{}) []*license.LicenseEntity {

	var scale int
	var expireTime int64
	var addType int
	var typeIntSlice []int
	var edition string
	var licenseTag license.LicenseTag
	licenseEntitySlice := []*license.LicenseEntity{}

	// 获取组织下所有License
	OrgLicense := suite.GetOrgLicense(orgUUID)

	// 判断不定长参数类型
	if typeIntOrEdition == nil {
		typeIntSlice = []int{1, 2, 3, 4, 5, 6, 7, 8}
	} else {
		for _, value := range typeIntOrEdition {
			switch value.(type) {
			case int:
				typeIntSlice = append(typeIntSlice, value.(int))

			case string:
				edition = value.(string)
			}
		}
	}

	// 否则只遍历传入的typeIntSlice
	for _, typeInt := range typeIntSlice {

		// 获取License不同应用的值
		for _, v := range OrgLicense {
			if v.Type == typeInt {
				scale = v.Scale
				expireTime = v.Expire_Time
				addType = v.Add_Type

				// 如果没有额外传入edition则默认取suite.license中定义的edition
				if len(edition) != 0 {
					licenseTag = AddLicenseTag(typeInt, edition)
				} else {
					licenseTag = AddLicenseTag(typeInt, v.Edition)
				}

			}
		}

		// 构造LicenseEntity
		licenseEntity := license.LicenseEntity{
			LicenseTag: licenseTag,
			OrgUUID:    orgUUID,
			Scale:      scale,
			ExpireTime: expireTime,
			AddType:    addType,
		}

		licenseEntitySlice = append(licenseEntitySlice, &licenseEntity)

	}

	fmt.Println("查询成功，", *licenseEntitySlice[0])
	return licenseEntitySlice

}

func (suite *testSuite) TestAddLicense() {

	type test struct {
		sql        gorp.SqlExecutor
		addType    int
		addLicense *license.LicenseAdd
		expected   interface{}
	}

	// 获取已有应用的licenseTag
	licenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	licenseTag := (*licenseEntity).LicenseTag

	data_suite := map[string]test{
		"传入完全不存在的licenseTag：添加证书成功": {suite.sqlExecutor,
			license.AddTypePay,
			&license.LicenseAdd{
				suite.orgIds[0],
				100,
				-1,
				AddLicenseTag(license.LicenseTypeProject, suite.autoFlag), // edition格式：auto+时间戳
			},
			AddLicenseTag(license.LicenseTypeProject, suite.autoFlag),
		},
		"传入已经存在的licenseTag，返回报错信息": {
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
			assert.EqualValues(suite.T(), tc.expected, (*result).LicenseTag, name)
			suite.DeleteLicense([]interface{}{suite.autoFlag})

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), result, name)
		}
	}

}

func (suite *testSuite) TestBatchAddLicenses() {

	type test struct {
		sql        gorp.SqlExecutor
		addType    int
		addLicense []*license.LicenseAdd
		expected   interface{}
	}

	// 获取已有应用的licenseTag
	licenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	licenseTag := (*licenseEntity).LicenseTag

	data_suite := map[string]test{
		"传入2个完全不存在的licenseTag：添加证书成功": {suite.sqlExecutor,
			license.AddTypePay,
			[]*license.LicenseAdd{
				{
					suite.orgIds[0],
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypeProject, suite.autoFlag),
				},
				{
					suite.orgIds[0],
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1"),
				},
			},
			[]license.LicenseTag{
				AddLicenseTag(license.LicenseTypeProject, suite.autoFlag),
				AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1"),
			},
		},
		"传入1个不存在和1个存在的licenseTag：返回报错信息": {suite.sqlExecutor,
			license.AddTypePay,
			[]*license.LicenseAdd{
				{
					suite.orgIds[0],
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"2"),
				},
				{
					suite.orgIds[0],
					100,
					license.UnLimitExpire,
					licenseTag,
				},
			},
			"",
		},
		"传入1个不存在addType：返回nil": {suite.sqlExecutor,
			license.AddTypePay,
			[]*license.LicenseAdd{
				{
					suite.orgIds[0],
					100,
					license.UnLimitExpire,
					AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"3"),
				},
			},
			"",
		},
	}

	for name, tc := range data_suite {
		result, err := license.BatchAddLicenses(tc.sql, tc.addType, tc.addLicense)
		if strings.Contains(name, "成功") {

			licenseEntities, _ := license.GetOrgAllLicensesByType(tc.sql, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
			LicenseTags := []license.LicenseTag{}
			for _, entity := range licenseEntities {
				LicenseTags = append(LicenseTags, (*entity).LicenseTag)
			}
			assert.EqualValues(suite.T(), tc.expected, LicenseTags, name)
			suite.DeleteLicense([]interface{}{suite.autoFlag})
			suite.DeleteLicense([]interface{}{suite.autoFlag + "1"})

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)

		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), result, name)
		}

	}

}

func (suite *testSuite) TestUpdateOrgLicenseScale() {
	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		tag      *license.LicenseTag
		newScale int
		expected interface{}
	}

	exitLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	exitLicenseTag := (*exitLicenseEntity).LicenseTag
	oldScale := (*exitLicenseEntity).Scale

	notExitLicenseTag := AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1")

	data_suite := map[string]test{
		"对存在的License修改scale为10：修改成功，scale为10": {
			suite.sqlExecutor,
			suite.orgIds[0],
			&exitLicenseTag,
			10,
			10,
		},
		"对不存在的License修改scale为10：返回报错信息": {
			suite.sqlExecutor,
			suite.orgIds[0],
			&notExitLicenseTag,
			10,
			"",
		},
		"对不存在的orgUUID修改scale为10：返回nil": {
			suite.sqlExecutor,
			"不存在的orgUUID",
			&notExitLicenseTag,
			10,
			"",
		},
	}

	for name, tc := range data_suite {
		licenseAlter, err := license.UpdateOrgLicenseScale(tc.sql, tc.orgUUID, tc.tag, tc.newScale)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, (*licenseAlter).New.Scale, name)
			// 还原scale
			license.UpdateOrgLicenseScale(tc.sql, tc.orgUUID, tc.tag, oldScale)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)

		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseAlter, name)
		}
	}

}

func (suite *testSuite) TestBatchUpdateOrgLicenseScale() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		tag      []*license.LicenseTag
		newScale int
		expected interface{}
	}

	exitLicenseEntitys, _ := license.GetOrgLicenses(suite.sqlExecutor, suite.orgIds[0])
	exitLicenseTags := []*license.LicenseTag{}
	for _, exitLicenseEntity := range exitLicenseEntitys {
		exitLicenseTag := (*exitLicenseEntity).LicenseTag
		exitLicenseTags = append(exitLicenseTags, &exitLicenseTag)
	}

	oldScale := int(math.Max(float64(exitLicenseEntitys[0].ExpireTime), float64(exitLicenseEntitys[1].ExpireTime)))

	notExitLicenseTag := AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1")

	data_suite := map[string]test{
		"对存在的2个License修改scale为20，修改成功": {
			suite.sqlExecutor,
			suite.orgIds[0],
			exitLicenseTags[:2],
			20,
			[]*license.LicenseAlter{
				{New: &license.LicenseAlterInfo{Scale: 20}},
				{New: &license.LicenseAlterInfo{Scale: 20}},
			},
		},
		"对1个不存在的licenseTag和存在的licenseTag修改scale为30：返回报错信息": {
			suite.sqlExecutor,
			suite.orgIds[0],
			[]*license.LicenseTag{
				&(exitLicenseEntitys[0].LicenseTag),
				&notExitLicenseTag,
			},
			30,
			"",
		},
		"对1个不存在的orgUUID修改scale为30：返回nil": {
			suite.sqlExecutor,
			"1234",
			exitLicenseTags[:2],
			30,
			"",
		},
	}

	for name, tc := range data_suite {
		licenseScales := []*license.LicenseAlter{}
		licenseAlters, err := license.BatchUpdateOrgLicenseScale(tc.sql, tc.orgUUID, tc.tag, tc.newScale)
		if strings.Contains(name, "成功") {
			for _, v := range licenseAlters {
				licenseScales = append(licenseScales, &license.LicenseAlter{New: v.New})
			}
			assert.EqualValues(suite.T(), tc.expected, licenseScales, name)
			// 还原
			license.BatchUpdateOrgLicenseScale(tc.sql, tc.orgUUID, tc.tag, oldScale)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
			exitLicenseEntitys, _ := license.GetOrgLicenses(suite.sqlExecutor, suite.orgIds[0])
			assert.NotEqualValues(suite.T(), tc.newScale, exitLicenseEntitys[0].Scale, name)

		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseAlters, name)
		}

	}

}

func (suite *testSuite) TestRenewalOrgLicenseExpire() {
	type test struct {
		sql        gorp.SqlExecutor
		orgUUID    string
		tag        *license.LicenseTag
		expireTime int64
		expected   int64
	}

	exitLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	exitLicenseTag := exitLicenseEntity.LicenseTag
	oldExpireTime := exitLicenseEntity.ExpireTime
	timeStamp := time.Now().Unix() + 60
	notExitLicenseTag := AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1")

	data_suite := map[string]test{
		"更新组织过期时间为当前时间+1分钟后：修改过期时间成功": {
			suite.sqlExecutor,
			suite.orgIds[0],
			&exitLicenseTag,
			timeStamp,
			timeStamp,
		},
		"更新组织不存在license的过期时间为当前时间+1分钟：返回报错信息": {
			suite.sqlExecutor,
			suite.orgIds[0],
			&notExitLicenseTag,
			timeStamp,
			timeStamp,
		},
		"更新不存在的组织license过期时间为当前时间+1分钟：返回nil": {
			suite.sqlExecutor,
			"123",
			&exitLicenseTag,
			timeStamp,
			timeStamp,
		},
	}

	for name, tc := range data_suite {
		licenseAlter, err := license.RenewalOrgLicenseExpire(tc.sql, tc.orgUUID, tc.tag, tc.expireTime)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, licenseAlter.New.ExpireTime, name)
			license.RenewalOrgLicenseExpire(tc.sql, tc.orgUUID, tc.tag, oldExpireTime)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)

		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseAlter, name)
		}
	}

}

func (suite *testSuite) TestBatchRenewalOrgLicenseExpire() {
	type test struct {
		sql        gorp.SqlExecutor
		orgUUID    string
		tag        []*license.LicenseTag
		expireTime int64
		expected   interface{}
	}

	exitLicenseEntitys, _ := license.GetOrgLicenses(suite.sqlExecutor, suite.orgIds[0])
	exitLicenseTags := []*license.LicenseTag{}
	for _, exitLicenseEntity := range exitLicenseEntitys {
		exitLicenseTag := exitLicenseEntity.LicenseTag
		exitLicenseTags = append(exitLicenseTags, &exitLicenseTag)
	}
	oldExpireTime := int64(math.Max(float64(exitLicenseEntitys[0].ExpireTime), float64(exitLicenseEntitys[1].ExpireTime)))

	notExitLicenseTag := AddLicenseTag(license.LicenseTypeProject, suite.autoFlag+"1")
	timeStamp := time.Now().Unix() + 60

	data_suite := map[string]test{
		"更新组织的2个license过期时间为当前时间+1分钟后：修改过期时间成功": {
			suite.sqlExecutor,
			suite.orgIds[0],
			exitLicenseTags[:2],
			timeStamp,
			[]*license.LicenseAlter{
				{New: &license.LicenseAlterInfo{ExpireTime: timeStamp}},
				{New: &license.LicenseAlterInfo{ExpireTime: timeStamp}},
			},
		}, "分别更新组织的1个存在和不存在的license过期时间为当前时间+1分钟后：返回报错信息": {
			suite.sqlExecutor,
			suite.orgIds[0],
			[]*license.LicenseTag{
				exitLicenseTags[0],
				&notExitLicenseTag,
			},
			timeStamp,
			"",
		},
		"更新不存在的组织license过期时间为当前时间+1分钟后：返回nil": {
			suite.sqlExecutor,
			suite.orgIds[0],
			exitLicenseTags,
			timeStamp,
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
			license.BatchRenewalOrgLicenseExpire(tc.sql, tc.orgUUID, tc.tag, oldExpireTime)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
			exitLicenseEntitys, _ := license.GetOrgLicenses(suite.sqlExecutor, suite.orgIds[0])
			assert.NotEqualValues(suite.T(), tc.expireTime, exitLicenseEntitys[0].ExpireTime, name)

		} else if strings.Contains(name, "nil") {
			assert.Nil(suite.T(), licenseAlters, name)
		}

	}

}

func (suite *testSuite) TestAddOrUpdateOrgDefaultGrant() {

	type test struct {
		sql          gorp.SqlExecutor
		orgUUID      string
		defaultGrant *license.LicenseDefaultGrant
		expected     bool
	}

	// 查询现有的默认授权
	licenseDefaultGrants, _ := license.GetOrgDefaultGrantLicenses(suite.sqlExecutor, suite.orgIds[0])
	var defaultGrant license.LicenseDefaultGrant

	if licenseDefaultGrants != nil {
		defaultGrant = *licenseDefaultGrants[0]
		defaultGrant.DefaultGrant = license.DefaultGrantFalse
	} else {
		defaultGrant = license.LicenseDefaultGrant{
			suite.orgIds[0],
			license.DefaultGrantFalse,
			license.GetLicenseType(license.LicenseTypeProject),
		}
	}

	// 测试数据
	data_suite := map[string]test{
		"新增组织的1个license默认授权配置：修改授权配置成功": {
			suite.sqlExecutor,
			suite.orgIds[0],
			&defaultGrant,
			license.DefaultGrantFalse,
		},
		"新增不存在的组织1个license默认授权配置：返回报错信息": {
			suite.sqlExecutor,
			"123",
			&defaultGrant,
			license.DefaultGrantFalse,
		},
	}

	for name, tc := range data_suite {
		err := license.AddOrUpdateOrgDefaultGrant(tc.sql, tc.orgUUID, tc.defaultGrant)
		if strings.Contains(name, "成功") {
			licenseDefaultGrants, _ := license.GetOrgDefaultGrantLicenses(suite.sqlExecutor, suite.orgIds[0])
			for _, v := range licenseDefaultGrants {
				if v.LicenseType == (*tc.defaultGrant).LicenseType {
					assert.False(suite.T(), v.DefaultGrant, name)
				}
			}

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)

		}

	}

}

func (suite *testSuite) TestBatchAddOrUpdateOrgDefaultGrants() {
	type test struct {
		sql           gorp.SqlExecutor
		orgUUID       string
		defaultGrants []*license.LicenseDefaultGrant
		expected      bool
	}

	// 查询现有的默认授权
	licenseDefaultGrants, _ := license.GetOrgDefaultGrantLicenses(suite.sqlExecutor, suite.orgIds[0])
	var defaultGrants []*license.LicenseDefaultGrant
	if licenseDefaultGrants != nil {
		defaultGrants = append(defaultGrants, licenseDefaultGrants[0])
		defaultGrants[0].DefaultGrant = !defaultGrants[0].DefaultGrant
	} else {
		defaultGrants = []*license.LicenseDefaultGrant{
			{
				suite.orgIds[0],
				license.DefaultGrantFalse,
				license.GetLicenseType(license.LicenseTypeProject),
			},
		}
	}

	// 测试数据
	data_suite := map[string]test{
		"新增组织的1个license默认授权配置:修改授权配置成功": {
			suite.sqlExecutor,
			suite.orgIds[0],
			defaultGrants,
			defaultGrants[0].DefaultGrant,
		},
		"新增不存在组织的1个license默认授权配置:返回报错信息": {
			suite.sqlExecutor,
			suite.orgIds[0],
			defaultGrants,
			defaultGrants[0].DefaultGrant,
		},
	}

	for name, tc := range data_suite {
		err := license.BatchAddOrUpdateOrgDefaultGrants(tc.sql, tc.orgUUID, tc.defaultGrants)
		if strings.Contains(name, "成功") {
			licenseDefaultGrants, _ := license.GetOrgDefaultGrantLicenses(suite.sqlExecutor, suite.orgIds[0])
			for _, v := range licenseDefaultGrants {
				if v.LicenseType == tc.defaultGrants[0].LicenseType {
					assert.EqualValues(suite.T(), tc.expected, v.DefaultGrant, name)
				}
			}

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)

		}

	}

}

func (suite *testSuite) TestGetOrgDefaultGrantLicenses() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected []*license.LicenseDefaultGrant
	}

	// 测试数据
	data_suite := map[string]test{
		"查询组织所有license默认授权配置：查询成功": {
			suite.sqlExecutor,
			suite.orgIds[0],
			suite.GetLicenseDefaultGrant(suite.orgIds[0]),
		},
		"查询不存在组织的默认授权配置：返回报错信息": {
			suite.sqlExecutor,
			suite.orgIds[0],
			suite.GetLicenseDefaultGrant(suite.orgIds[0]),
		},
	}

	for name, tc := range data_suite {
		licenseDefaultGrants, err := license.GetOrgDefaultGrantLicenses(tc.sql, tc.orgUUID)

		if strings.Contains(name, "成功") {
			assert.Equal(suite.T(), tc.expected, licenseDefaultGrants, name)
		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		} else {
			assert.Nil(suite.T(), licenseDefaultGrants, name)
		}

	}

}

func (suite *testSuite) TestGetOrgLicenseByTypeAndTimestamp() {
	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		sec         int64
		expected    *license.LicenseEntity
	}

	timeStamp := time.Now().Unix()

	licenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	license.RenewalOrgLicenseExpire(suite.sqlExecutor, suite.orgIds[0], &(licenseEntity.LicenseTag), timeStamp+60)
	expectedLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))

	// 测试数据
	data_suite := map[string]test{
		"传入正确组织UUID和时间戳：查询成功，可查到该组织下过期时间大于传入时间戳的license": {
			suite.sqlExecutor,
			suite.orgIds[0],
			license.GetLicenseType(license.LicenseTypeProject),
			timeStamp,
			expectedLicenseEntity,
		},
		"传入错误组织UUID：返回报错信息": {
			suite.sqlExecutor,
			suite.orgIds[0],
			license.GetLicenseType(license.LicenseTypeProject),
			timeStamp,
			expectedLicenseEntity,
		},
		"传入不存在的licenseType：返回报错信息": {
			suite.sqlExecutor,
			suite.orgIds[0],
			license.GetLicenseType(license.LicenseTypeProject),
			timeStamp,
			expectedLicenseEntity,
		},
	}

	for name, tc := range data_suite {
		actualLicenseEntity, err := license.GetOrgLicenseByTypeAndTimestamp(tc.sql, suite.orgIds[0], tc.licenseType, tc.sec)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, actualLicenseEntity, name)

		} else if strings.Contains(name, "报错") {
			assert.Error(suite.T(), err, name)
		}
	}
}

func (suite *testSuite) TestListExpireInTimeStampRange() {

	type test struct {
		sql        gorp.SqlExecutor
		startStamp int64
		endStamp   int64
		expected   *license.LicenseEntity
	}

	startStamp := time.Now().Unix() - 600
	endStamp := startStamp + 600

	var renewLicenseEntity *license.LicenseEntity
	// 获取licenseTag
	exitLicenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))
	exitLicenseTag := exitLicenseEntity.LicenseTag
	oldExpireTime := exitLicenseEntity.ExpireTime
	// 更改到期时间
	license.RenewalOrgLicenseExpire(suite.sqlExecutor, suite.orgIds[0], &exitLicenseTag, startStamp+30)
	renewLicenseEntity, _ = license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(license.LicenseTypeProject))

	// 测试数据
	data_suite := map[string]test{
		"传入起始和终止时间：查询成功，查询所有组织下过期时间在范围时间内的license": {
			suite.sqlExecutor,
			startStamp,
			endStamp,
			renewLicenseEntity,
		},
	}

	for name, tc := range data_suite {
		licenseEntities, _ := license.ListExpireInTimeStampRange(tc.sql, tc.startStamp, tc.endStamp)
		if strings.Contains(name, "成功") {
			assert.Contains(suite.T(), licenseEntities, renewLicenseEntity, name)
			license.RenewalOrgLicenseExpire(suite.sqlExecutor, suite.orgIds[0], &exitLicenseTag, oldExpireTime)
		}

	}
}

func (suite *testSuite) TestGetOrgLicenses() {
	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected interface{}
	}
	orgLicenses := []*license.LicenseEntity{}
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 8} {
		licenseEntity, _ := license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(v))
		orgLicenses = append(orgLicenses, licenseEntity)
	}

	// 测试数据
	data_suite := map[string]test{
		"传入正确组织UUID：查询成功，可查到该组织下所有最高优先级的license": {
			suite.sqlExecutor,
			suite.orgIds[0],
			orgLicenses,
		},
		"传入错误组织UUID：返回nil": {
			suite.sqlExecutor,
			suite.orgIds[0],
			orgLicenses,
		},
	}

	for name, tc := range data_suite {
		licenseEntities, _ := license.GetOrgLicenses(tc.sql, tc.orgUUID)
		if strings.Contains(name, "成功") {
			assert.Equal(suite.T(), tc.expected, licenseEntities, name)
		} else {
			assert.Nil(suite.T(), licenseEntities, name)
		}

	}
}

func (suite *testSuite) TestGetOrgLicensesMap() {
	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected interface{}
	}
	orgLicensesMap := map[int]*license.LicenseEntity{}
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 8} {
		orgLicensesMap[v], _ = license.GetOrgLicenseByType(suite.sqlExecutor, suite.orgIds[0], license.GetLicenseType(v))
	}

	// 测试数据
	data_suite := map[string]test{
		"传入正确组织UUID：查询成功，可查到该组织下每个应用下的最高优先级的license": {
			suite.sqlExecutor,
			suite.orgIds[0],
			orgLicensesMap,
		},
		"传入错误组织UUID：返回nil": {
			suite.sqlExecutor,
			suite.orgIds[0],
			"",
		},
	}

	for name, tc := range data_suite {
		licenseEntitiesMap, _ := license.GetOrgLicensesMap(tc.sql, tc.orgUUID)
		if strings.Contains(name, "成功") {
			assert.Equal(suite.T(), tc.expected, licenseEntitiesMap, name)
		} else {
			assert.Nil(suite.T(), licenseEntitiesMap, name)
		}

	}
}

func (suite *testSuite) TestBatchGetOrgLicensesMaps() {
	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  []string
		expected interface{}
	}
	orgLicensesMap := map[string]map[int]*license.LicenseEntity{}
	for _, orgUUID := range suite.orgIds {
		orgLicensesMap[orgUUID], _ = license.GetOrgLicensesMap(suite.sqlExecutor, orgUUID)

	}
	firstLicensesMap, _ := license.GetOrgLicensesMap(suite.sqlExecutor, suite.orgIds[0])

	// 测试数据
	data_suite := map[string]test{
		"传入正确2个组织UUID：查询成功，可查到每个组织下每个应用下的最高优先级的license": {
			suite.sqlExecutor,
			suite.orgIds,
			orgLicensesMap,
		}, "传入1个正确和1个错误的组织UUID：查询成功，只查到正确组织对应的license": {
			suite.sqlExecutor,
			[]string{suite.orgIds[0], "123"},
			map[string]map[int]*license.LicenseEntity{suite.orgIds[0]: firstLicensesMap},
		},
	}

	for name, tc := range data_suite {
		orgLicensesMaps, _ := license.BatchGetOrgLicensesMaps(tc.sql, tc.orgUUID)
		if strings.Contains(name, "成功") {
			assert.EqualValues(suite.T(), tc.expected, orgLicensesMaps, name)
		} else {
			assert.Nil(suite.T(), orgLicensesMaps, name)
		}

	}
}

func (suite *testSuite) TestGetOrgAllLicensesMap() {

	type test struct {
		sql      gorp.SqlExecutor
		orgUUID  string
		expected string
	}

	licenseTag := AddLicenseTag(license.LicenseTypeProject, suite.autoFlag)
	license.AddLicense(
		suite.sqlExecutor,
		license.AddTypePay,
		&license.LicenseAdd{suite.orgIds[0], 100, -1, licenseTag})

	// 测试数据
	data_suite := map[string]test{
		"传入正确的组织UUID：查询成功，可查到project应用存在新配置的license": {
			suite.sqlExecutor,
			suite.orgIds[0],
			suite.autoFlag,
		},
		"传入错误的组织UUID：返回nil": {
			suite.sqlExecutor,
			suite.orgIds[0],
			suite.autoFlag,
		},
	}

	for name, tc := range data_suite {
		orgAllLicensesMap, _ := license.GetOrgAllLicensesMap(tc.sql, tc.orgUUID)
		licenseEdition := []string{}
		if strings.Contains(name, "成功") {
			for _, LicenseEntity := range orgAllLicensesMap[license.LicenseTypeProject] {
				licenseEdition = append(licenseEdition, LicenseEntity.EditionName)
			}
			assert.Contains(suite.T(), licenseEdition, tc.expected, name)
		} else {
			assert.Nil(suite.T(), orgAllLicensesMap, name)
		}

	}

}

func (suite *testSuite) TestGetOrgAllLicensesByType() {
	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		expected    []*license.LicenseEntity
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
		"传入错误的组织UUID：返回nil": {
			suite.sqlExecutor,
			"123",
			license.GetLicenseType(license.LicenseTypeProject),
			orgAllLicensesMap[license.LicenseTypeProject],
		},
		"传入错误的应用类型LicenseType：返回nil": {
			suite.sqlExecutor,
			suite.orgIds[0],
			license.GetLicenseType(100),
			orgAllLicensesMap[license.LicenseTypeProject],
		},
	}

	for name, tc := range data_suite {
		orgAllLicenses, _ := license.GetOrgAllLicensesByType(tc.sql, tc.orgUUID, tc.licenseType)
		if strings.Contains(name, "成功") {
			assert.Equal(suite.T(), tc.expected, orgAllLicenses, name)
		} else {
			assert.Nil(suite.T(), orgAllLicenses, name)
		}

	}

}

func (suite *testSuite) TestGetOrgLicenseByType() {

	// 正向出参,project 试用版
	type test struct {
		sql         gorp.SqlExecutor
		orgUUID     string
		licenseType license.LicenseType
		expected    interface{}
	}

	// 正向入参
	orgUUID := suite.orgIds[0]
	licenseType := license.GetLicenseType(license.LicenseTypeProject)

	orgLicenses := suite.GetOrgLicense(orgUUID)
	var licenseTag license.LicenseTag
	for _, licenses := range orgLicenses {
		if licenses.Edition == license.EditionEnterprise {
			licenseTag = license.LicenseTag{licenseType, licenses.Edition}
			license.UpdateOrgLicenseScale(suite.sqlExecutor, orgUUID, &licenseTag, 101)
			break
		} else {
			license.AddLicense(suite.sqlExecutor,
				license.AddTypePay,
				&license.LicenseAdd{
					suite.orgIds[0],
					101,
					-1,
					license.LicenseTag{licenseType, license.EditionEnterprise}})
		}
	}

	data_suite := map[string]test{
		"传入project的licenseType: 查询成功，返回应用企业版licenseEntity,scale为101": {suite.sqlExecutor, orgUUID, licenseType, 101},
		"传入错误的orgUUID：返回nil":    {suite.sqlExecutor, orgUUID + "123", licenseType, nil},
		"orgUUID为空串：返回nil":      {suite.sqlExecutor, "", licenseType, nil},
		"licenseType为nil：返回nil": {suite.sqlExecutor, orgUUID, nil, nil},
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
func (suite *testSuite) Test01() {
	fmt.Println("1")
}

// 测试套件，批量执行测试用例
func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}
