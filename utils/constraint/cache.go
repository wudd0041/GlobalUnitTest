package constraint

import (
	"gopkg.in/gorp.v1"
)

const (
	GLSUserUuidKey = "constraint_user_uuid_key"
	GLSOrgUuidKey  = "constraint_org_uuid_key"
)

var dbExec gorp.SqlExecutor

func SetDBExec(db gorp.SqlExecutor) {
	dbExec = db
}
func GetDBExec() gorp.SqlExecutor {
	return dbExec
}

func GlsSetUuid(key, uuid string) {
	//gls.Set(key, uuid)
}

func GlsGetUuid(key string) string {
	return "uuid"
}

func glsSetAllConstraint(key string, value map[string]interface{}) {
	//gls.Set(key, value)
}

func glsGetAllConstraint(key string) (map[string]interface{}, bool) {
	return nil, true
}

// findUserAllConstraint 获取用户所有的配置集包含实例/应用级别,并存入map
func findUserAllConstraint(orgUUID string, userUUID string) (map[string]interface{}, error) {
	resultMap := make(map[string]interface{})
	resultMap = MergeMap(keys_default_value, resultMap) //加入默认配置
	return resultMap, nil
}

func findOrgAllConstraints(orgUUID string) (map[string]interface{}, error) {
	resultMap := make(map[string]interface{})
	resultMap = MergeMap(resultMap, hub.instance.dict)  //merge实例配置集
	resultMap = MergeMap(keys_default_value, resultMap) //加入默认配置
	return resultMap, nil
}

//默认配置集+实例配置集
func defaultConstraint() map[string]interface{} {
	resultMap := make(map[string]interface{})
	resultMap = MergeMap(resultMap, hub.instance.dict)  //merge实例配置集
	resultMap = MergeMap(keys_default_value, resultMap) //加入默认配置
	return resultMap
}

func MergeMap(mObj ...map[string]interface{}) map[string]interface{} {
	newObj := map[string]interface{}{}
	for _, m := range mObj {
		for k, v := range m {
			newObj[k] = v
		}
	}
	return newObj
}
