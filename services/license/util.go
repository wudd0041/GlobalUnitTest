package license

import (
	licenseModel "license_testing/models/license"
)

func GetLicenseTypeIntByName(name string) int {
	return appTypeNameMap[name]
}

func GetLicenseNameByTypeInt(appType int) string {
	return appNameTypeMap[appType]
}

func GetLicenseI18NNameByTypeInt(appType int) string {
	return ""
}

func GetLicenseI18NDescByTypeInt(appType int) string {
	return ""
}

func initEnableEntityFromModelLicenses(licenses []*licenseModel.License, sec int64) *LicenseEntity {
	var entity *LicenseEntity
	entity = nil
	for _, lic := range licenses {
		curEntity := initLicenseEntityByModelLicense(lic)
		if !curEntity.Valid() {
			continue
		}
		if entity == nil || entity.Priority() < curEntity.Priority() {
			entity = curEntity
		}
	}

	return entity
}

func convertModelLicenseListToLicenseEntityMap(modelLicenses []*licenseModel.License) map[int]*LicenseEntity {
	licenseMap := make(map[int]*LicenseEntity)
	for _, lic := range modelLicenses {
		entity := &LicenseEntity{
			OrgUUID:    lic.OrgUUID,
			Scale:      lic.Scale,
			ExpireTime: lic.ExpireTime,
			AddType:    lic.AddType,
			LicenseTag: LicenseTag{
				LicenseType: GetLicenseType(lic.Type),
				EditionName: lic.Edition,
			},
		}
		if !entity.Valid() {
			continue
		}

		mapEntity, ok := licenseMap[lic.Type]
		if !ok {
			licenseMap[lic.Type] = entity
			continue
		}

		if entity.Priority() > mapEntity.Priority() {
			licenseMap[lic.Type] = entity
		}

	}

	return licenseMap
}

func initLicenseEntityByModelLicense(license *licenseModel.License) *LicenseEntity {
	return &LicenseEntity{
		OrgUUID: license.OrgUUID,
		LicenseTag: LicenseTag{
			LicenseType: GetLicenseType(license.Type),
			EditionName: license.Edition,
		},
		Scale:      license.Scale,
		AddType:    license.AddType,
		ExpireTime: license.ExpireTime,
	}
}

func convertLicenseEntityToModelLicense(entity *LicenseEntity, sec int64) *licenseModel.License {
	return &licenseModel.License{
		OrgUUID:    entity.OrgUUID,
		Type:       entity.Type(),
		Scale:      entity.Scale,
		AddType:    entity.AddType,
		ExpireTime: entity.ExpireTime,
		Edition:    entity.EditionName,
		UpdateTime: sec,
	}
}

func convertLicenseEntityToLicenseAlter(entity *LicenseEntity) *LicenseAlter {
	return &LicenseAlter{
		UUID:       "",
		OrgUUID:    entity.OrgUUID,
		LicenseTag: entity.LicenseTag,
		AddType:    entity.AddType,
	}
}

func convertLicenseEntityToModelAlter(entity *LicenseEntity, sec int64) *licenseModel.LicenseAlter {
	return &licenseModel.LicenseAlter{
		UUID:        "",
		OrgUUID:     entity.OrgUUID,
		LicenseType: entity.Type(),
		Edition:     entity.EditionName,
		AddType:     entity.AddType,
		Scale:       entity.Scale,
		ExpireTime:  entity.ExpireTime,
		CreateTime:  sec,
	}
}
