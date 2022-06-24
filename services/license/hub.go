package license

var editionHub *LicenseEditionHub
var typeHub *LicenseTypeHub

func init() {
	editionHub = &LicenseEditionHub{
		editions: make([]*LicenseEdition, 0),
	}
	typeHub = &LicenseTypeHub{
		types: make([]*LicenseType, 0),
	}

	typeHub.registType(&ProjectLicense{})
	typeHub.registType(&WikiLicense{})
	typeHub.registType(&TestcaseLicense{})
	typeHub.registType(&PipelineLicense{})
	typeHub.registType(&PlanLicense{})
	typeHub.registType(&AccountLicense{})
	typeHub.registType(&DeskLicense{})
	typeHub.registType(&PerformanceLicense{})
}

func RegistEdition(edition *LicenseEdition) {
	if edition == nil {
		return
	}
	editionHub.editions = append(editionHub.editions, edition)
}

type LicenseEditionHub struct {
	editions []*LicenseEdition
}

func (h *LicenseEditionHub) listConfigsByType(licenseType int) []*LicenseEdition {
	res := make([]*LicenseEdition, 0)
	for _, c := range h.editions {
		if c.Type() == licenseType {
			res = append(res, c)
		}
	}
	return res
}

func (h *LicenseEditionHub) findConfigByTypeEdition(licenseType int, edition string) *LicenseEdition {
	for _, c := range h.editions {
		if c.Type() == licenseType && c.EditionName == edition {
			return c
		}
	}
	return nil
}

type LicenseTypeHub struct {
	types []*LicenseType
}

func (h *LicenseTypeHub) registType(in interface{}) {
	if t, ok := in.(LicenseType); ok {
		h.types = append(h.types, &t)
	}
}

func (h *LicenseTypeHub) findLicenseType(findType int) LicenseType {
	for _, licenseType := range h.types {
		if (*licenseType).Type() == findType {
			return *licenseType
		}
	}
	return nil
}

func (h *LicenseTypeHub) findLicenseTypeByName(name string) LicenseType {
	for _, licenseType := range h.types {
		if (*licenseType).Name() == name {
			return *licenseType
		}
	}
	return nil
}

func AllLicenseTypeMap() map[int]LicenseType {
	res := make(map[int]LicenseType)
	for _, licenseType := range typeHub.types {
		res[(*licenseType).Type()] = *licenseType
	}
	return res
}

func NormalAppMap() map[int]LicenseType {
	res := make(map[int]LicenseType)
	for _, licenseType := range typeHub.types {
		if (*licenseType).Type() == LicenseTypeAccount {
			continue
		}
		res[(*licenseType).Type()] = *licenseType
	}
	return res
}

func GetLicenseType(findType int) LicenseType {
	return typeHub.findLicenseType(findType)
}

func GetLicenseTypeByName(name string) LicenseType {
	return typeHub.findLicenseTypeByName(name)
}

func ListLicenseEditions() []*LicenseEdition {
	return editionHub.editions
}
