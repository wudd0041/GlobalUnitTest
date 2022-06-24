package constraint

//GetConstraint 获取当前goroutine处理当前user的配置项
func GetConstraint(key string) *Constraint {
	return GetUserConstraint(GlsGetUuid(GLSOrgUuidKey), GlsGetUuid(GLSUserUuidKey), key)
}

// BatchGetConstraints 批量获取当前goroutine处理当前user的配置项
func BatchGetConstraints(keys []string) *Constraints {
	return GetUserConstraints(GlsGetUuid(GLSOrgUuidKey), GlsGetUuid(GLSUserUuidKey), keys)
}

// GetAllConstraints 获取当前goroutine处理当前user的所有配置项
func GetAllConstraints() *Constraints {
	return GetUserAllConstraints(GlsGetUuid(GLSOrgUuidKey), GlsGetUuid(GLSUserUuidKey))
}

// GetUserConstraint 获取指定user的配置项
func GetUserConstraint(orgUUID, userUUID, key string) *Constraint {
	return nil
}

// GetUserConstraints 批量获取user配置项
func GetUserConstraints(orgUUID, userUUID string, keys []string) *Constraints {
	return nil
}

// GetUserAllConstraints 获取user全部配置项
func GetUserAllConstraints(orgUUID, userUUID string) *Constraints {
	return nil
}

// GetOrgAllConstraints 获取org全部配置项
func GetOrgAllConstraints(orgUUID string) *Constraints {
	return nil
}
