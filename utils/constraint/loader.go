package constraint

import (
	"crypto/x509"
	"encoding/asn1"
	"license_testing/services/license"
	"time"
)

var instanceCertKey = asn1.ObjectIdentifier{1, 2, 1, 3} //实例配置
var appCertKey = asn1.ObjectIdentifier{1, 2, 1, 4}      //应用配置

type baseConfig struct {
	dict map[string]interface{}
}

type instanceConfig struct {
	baseConfig
}

type appConfig struct {
	baseConfig
	appName  string
	edition  string
	priority int
	deadline time.Time
}

var hub *configHub

type configHub struct {
	instance instanceConfig
	apps     []appConfig
}

func (c *configHub) loadInstanceConfigFromAgent(dict map[string]interface{}) {
	c.instance = instanceConfig{
		baseConfig{
			dict,
		},
	}
}

func (c *configHub) loadAppConfigsFromAgent(
	dict map[string]interface{},
	appName string,
	edition string,
	priority int,
	deadline time.Time,
) {
	c.apps = append(c.apps, appConfig{
		baseConfig{dict},
		appName,
		edition,
		priority,
		deadline,
	})
	licenseType := license.GetLicenseTypeByName(appName)
	//再注册一遍到license里面
	licenseEdition := &license.LicenseEdition{
		LicenseTag: license.LicenseTag{
			LicenseType: licenseType,
			EditionName: edition,
		},
		Priority:    priority,
		InvalidTime: deadline.Unix(), //秒
	}
	license.RegistEdition(licenseEdition)
}

func init() {
	hub = &configHub{
		apps: make([]appConfig, 0),
	}
}

func LoadCertificate() error {
	err, certs := verifyCert()
	if err != nil {
		return nil
	}
	for _, caCert := range certs {
		resolveCertificate(caCert)
	}
	return nil
}

//resolveCertificate 解析证书
func resolveCertificate(caCert *x509.Certificate) {
	var certType string
	var collMap = make(map[string]interface{})
	for _, ext := range caCert.Extensions {
		if ext.Id.Equal(instanceCertKey) {
			certType = "instance"
		} else if ext.Id.Equal(appCertKey) {
			certType = "app"
		}
	}
	if certType == "instance" {
		hub.loadInstanceConfigFromAgent(collMap)
	} else if certType == "app" {
		var configMap = make(map[string]interface{})
		var edition, appName string
		var priority int
		if v, ok := collMap["appName"]; ok {
			appName = v.(string)
		}
		if v, ok := collMap["edition"]; ok {
			edition = v.(string)
		}
		if v, ok := collMap["priority"]; ok {
			priority = v.(int)
		}
		if v, ok := collMap["config"]; ok {
			configMap = v.(map[string]interface{})
		}
		hub.loadAppConfigsFromAgent(configMap, appName, edition, priority, caCert.NotAfter)
	}
}
