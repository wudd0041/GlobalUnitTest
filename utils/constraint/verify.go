package constraint

import (
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"unsafe"
)

var ONESCAPublicKey = []byte(``)

const CertFile = "conf/constraint_cert"

func verifyCert() (error, []*x509.Certificate) {

	return nil, nil
}

//getPublicCert 获取公钥信息
func getPublicCert() (*x509.Certificate, error) {
	publicCert, err := parseCert(ONESCAPublicKey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return publicCert, nil
}

// verifyPem 验证证书
func verifyPem(publicCert *x509.Certificate, certFile string, cerSlice *[]*x509.Certificate) {
	caCrt, err := ioutil.ReadFile(certFile)
	if err != nil {
		fmt.Printf("%s证书读取失败", certFile)
		return
	}
	caCert, err := parseCert(caCrt)
	if err != nil {
		fmt.Printf("%s证书解析失败", certFile)
		return
	}
	err = caCert.CheckSignatureFrom(publicCert)
	if err != nil {
		fmt.Printf("%s证书校验失败", certFile)
		return
	}
	if caCert.NotAfter.Before(time.Now()) {
		fmt.Printf("%s证书过期", certFile)
		return
	}
	*cerSlice = append(*cerSlice, caCert)
}

// verifyFiles 读取目录下所有pem证书，然后解析验证
func verifyFiles(filePath string, publicCert *x509.Certificate, cerSlice *[]*x509.Certificate) error {

	return nil
}

func parseCert(cert []byte) (*x509.Certificate, error) {
	return nil, nil
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}

func BytesToString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
