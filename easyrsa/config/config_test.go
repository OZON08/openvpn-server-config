package easyrsaconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	easyrsaconfig "github.com/OZON08/openvpn-server-config/easyrsa/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	assert.Equal(t, 2048, easyrsaconfig.New().EasyRSAKeySize)
}

func TestTemplateGeneration(t *testing.T) {
	txt, err := os.ReadFile("./templates/easyrsa-vars.tpl")
	require.NoError(t, err)

	_, err = easyrsaconfig.GetText(string(txt), easyrsaconfig.New())
	assert.NoError(t, err)
}

func TestBrokenTemplate(t *testing.T) {
	_, err := easyrsaconfig.GetText("{{ ", easyrsaconfig.New())
	assert.Error(t, err, "Parser should fail on broken template")
}

// Every string field ends up inside a shell script that Easy-RSA sources.
func allFields(c *easyrsaconfig.Config) map[string]*string {
	return map[string]*string{
		"EasyRSADN":          &c.EasyRSADN,
		"EasyRSAReqCountry":  &c.EasyRSAReqCountry,
		"EasyRSAReqProvince": &c.EasyRSAReqProvince,
		"EasyRSAReqCity":     &c.EasyRSAReqCity,
		"EasyRSAReqOrg":      &c.EasyRSAReqOrg,
		"EasyRSAReqEmail":    &c.EasyRSAReqEmail,
		"EasyRSAReqOu":       &c.EasyRSAReqOu,
		"EasyRSAReqCn":       &c.EasyRSAReqCn,
	}
}

func TestGetTextRejectsShellInjection(t *testing.T) {
	payloads := []string{"$(id)", "`id`", `x"; id #`, "x\nid", "x\rid", `a\b`, "a}b", "${HOME}", "a;b", "a|b", "a/CN=b", "O'Neill"}
	for _, payload := range payloads {
		c := easyrsaconfig.New()
		for name, field := range allFields(&c) {
			original := *field
			*field = payload
			_, err := easyrsaconfig.GetText("{{ .EasyRSAReqOrg }}", c)
			assert.Error(t, err, "%s must reject %q", name, payload)
			if err != nil {
				assert.Contains(t, err.Error(), name)
			}
			*field = original
		}
	}
}

func TestGetTextAcceptsRealisticValues(t *testing.T) {
	c := easyrsaconfig.New()
	c.EasyRSAReqOrg = "Müller & Söhne GmbH (Köln), Inc."
	c.EasyRSAReqOu = "IT-Services"
	c.EasyRSAReqEmail = "first.last+vpn@example.org"
	c.EasyRSAReqCn = "vpn_server-01:main"
	c.EasyRSAReqCity = ""

	out, err := easyrsaconfig.GetText("{{ .EasyRSAReqOrg }}", c)
	require.NoError(t, err)
	assert.Equal(t, c.EasyRSAReqOrg, out)
}

func TestSaveToFileWritesNothingForInvalidConfig(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "vars")
	c := easyrsaconfig.New()
	c.EasyRSAReqOrg = "$(id)"

	err := easyrsaconfig.SaveToFile("./templates/easyrsa-vars.tpl", c, dest)

	assert.Error(t, err)
	assert.NoFileExists(t, dest)
}
