package license

type ProjectLicense struct{}

func (l *ProjectLicense) Type() int {
	return LicenseTypeProject
}

func (l *ProjectLicense) Name() string {
	return LicenseNameProject
}

type WikiLicense struct{}

func (l *WikiLicense) Type() int {
	return LicenseTypeWiki
}

func (l *WikiLicense) Name() string {
	return LicenseNameWiki
}

type TestcaseLicense struct{}

func (l *TestcaseLicense) Type() int {
	return LicenseTypeTestCase
}

func (l *TestcaseLicense) Name() string {
	return LicenseNameTestcase
}

type PipelineLicense struct{}

func (l *PipelineLicense) Type() int {
	return LicenseTypePipeline
}

func (l *PipelineLicense) Name() string {
	return LicenseNamePipeline
}

type PlanLicense struct{}

func (l *PlanLicense) Type() int {
	return LicenseTypePlan
}

func (l *PlanLicense) Name() string {
	return LicenseNamePlan
}

type AccountLicense struct{}

func (l *AccountLicense) Type() int {
	return LicenseTypeAccount
}

func (l *AccountLicense) Name() string {
	return LicenseNameAccount
}

type DeskLicense struct{}

func (l *DeskLicense) Type() int {
	return LicenseTypeDesk
}

func (l *DeskLicense) Name() string {
	return LicenseNameDesk
}

type PerformanceLicense struct{}

func (l *PerformanceLicense) Type() int {
	return LicenseTypePerformance
}

func (l *PerformanceLicense) Name() string {
	return LicenseNamePerformance
}
