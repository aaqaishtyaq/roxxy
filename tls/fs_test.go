package tls

import (
	"crypto/tls"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/check.v1"
)

var fileCertPEM = `-----BEGIN CERTIFICATE-----
MIICDDCCAXUCFDXpci2fF+Ze2PAh8461caB850JxMA0GCSqGSIb3DQEBCwUAMEUx
CzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRl
cm5ldCBXaWRnaXRzIFB0eSBMdGQwHhcNMjIwNDA5MTMxOTQ5WhcNMjMwNDA5MTMx
OTQ5WjBFMQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UE
CgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GN
ADCBiQKBgQDWFEydXdLYRPHVuWY8wdDEc/RdPU9m4h0cXySCcB2b12a6PPvWW+9a
0V+VWKk1MQY3DMYSUTg1bXmNHcuh2OG6/5Uou/BKfjpHA27CtEAsdOrlw8FtKGDW
ftPMA4K1sWsnA1po/SXU0X7RCRibE+s4g+VPJrlzx98oXrsID2pNVwIDAQABMA0G
CSqGSIb3DQEBCwUAA4GBALCDcEqGUnjCHbvnhiLtPaN7s67k8Jo2zSpla7euxtyY
vdR6Dv5bcpEDdlQ9BWlUO16PQn3m87keO426GaIgRIQ6TERb/FalPVpTd+VRpZvA
1vkVSkTxgnsxjRxPuqd2QXJhdIdtXgRoA1sj2fLNZ51XuR7uQSFK54IflqooRBRB
-----END CERTIFICATE-----
`

var fileKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDWFEydXdLYRPHVuWY8wdDEc/RdPU9m4h0cXySCcB2b12a6PPvW
W+9a0V+VWKk1MQY3DMYSUTg1bXmNHcuh2OG6/5Uou/BKfjpHA27CtEAsdOrlw8Ft
KGDWftPMA4K1sWsnA1po/SXU0X7RCRibE+s4g+VPJrlzx98oXrsID2pNVwIDAQAB
AoGARTBIVq/lHgqiUl3aQhat32BOgPf4upqnp+zEAvgzSZPDWrus9Om/oQ18I+uE
vHE8vfv95Bul2/amy0nu7z8GLO9dApL86oBWZYbJvhS29WaUOu9lIRhRpyINeRsZ
IpMVQNjkADJFZC8zm1k6lamjnuN1w+SeXjh2v21KRU+z0aECQQDw6jCFQdkNVZt6
v1M3fMQe2IIdT7uDjLOIiuSQufQVLbOe+fhLuWMyAKbJf0qLE0fTyPBJwx6+rQtR
Km7SpWmxAkEA43vxScWYNfYvS+yYAdhaCzCGuqnp2hJmVay1+/wpzb9kNb/lk5A2
uzvBmFlzrcadOAy261d+9Ah0DLwKpLvhhwJAHaln4fBSjg69Puaxk0JcT0Pu+Tbo
6nB3Zldbfuo2QClJVUiHpqMjsHNeFa8DeY4dKNkzpJFOhsF9hDfKP0s4cQJAaeja
fa4xH25utrqAStufkHYXQ/C3n3/RhTHTyG2uSMxCq4OcLweFc8Zua6+5274MlHvW
7drekF8fKI6jpe6TIQJBANMAzan0h0ikpoUCZrL9M43t5lDn1NliI1z3dVARWLhN
fDOSKjcG8gFG3/mb581MK8vMQJCu2pFEA1eqRfz/fHY=
-----END RSA PRIVATE KEY-----
`

type FSSuite struct {
	path string
	be   CertificateLoader
}

var _ = check.Suite(&FSSuite{})

func (s *FSSuite) SetUpTest(c *check.C) {
	var err error
	s.path, err = ioutil.TempDir("", "cert-path")
	c.Assert(err, check.IsNil)

	tmpfn := filepath.Join(s.path, "*.dvito.sh.key")
	err = ioutil.WriteFile(tmpfn, []byte(fileKeyPEM), 0666)
	c.Assert(err, check.IsNil)

	tmpfn = filepath.Join(s.path, "*.dvito.sh.crt")
	err = ioutil.WriteFile(tmpfn, []byte(fileCertPEM), 0666)
	c.Assert(err, check.IsNil)

	tmpfn = filepath.Join(s.path, "aaqa.dev.key")
	err = ioutil.WriteFile(tmpfn, []byte(fileKeyPEM), 0666)
	c.Assert(err, check.IsNil)

	tmpfn = filepath.Join(s.path, "aaqa.dev.crt")
	err = ioutil.WriteFile(tmpfn, []byte(fileCertPEM), 0666)
	c.Assert(err, check.IsNil)

	s.be = NewFSCertificateLoader(s.path)
}

func (s *FSSuite) TearDownTest(c *check.C) {
	os.RemoveAll(s.path)
}

func (s *FSSuite) TestCertificateNotFound(c *check.C) {
	clientHello := &tls.ClientHelloInfo{
		ServerName: "test.test",
	}
	_, err := s.be.GetCertificate(clientHello)
	c.Assert(err, check.ErrorMatches, `Certificate for \"test.test\" not is found`)
}

func (s *FSSuite) TestCertificateFound(c *check.C) {
	clientHello := &tls.ClientHelloInfo{
		ServerName: "aaqa.dev",
	}
	_, err := s.be.GetCertificate(clientHello)
	c.Assert(err, check.IsNil)
}

func (s *FSSuite) TestWildCardCertificate(c *check.C) {
	clientHello := &tls.ClientHelloInfo{
		ServerName: "hello.dvito.sh",
	}

	_, err := s.be.GetCertificate(clientHello)
	c.Assert(err, check.IsNil)
}
