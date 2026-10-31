package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OZON08/openvpn-server-config/client/config"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	c := config.New()
	assert.Equal(t, "SHA512", c.Auth)
}

func TestTemplateGeneration(t *testing.T) {
	c := config.New()
	txt, err := os.ReadFile("./templates/openvpn-client-config.tpl")
	assert.Nil(t, err)

	_, err = config.GetText(string(txt), c)
	assert.Nil(t, err)
}

func TestBrokenTemplate(t *testing.T) {
	c := config.New()

	_, err := config.GetText("{{ ", c)
	assert.NotNil(t, err, "Parser should fail on broken template")
}

func TestRejectsDirectiveInjection(t *testing.T) {
	c := config.New()
	c.ServerAddress = "vpn.example.org\nscript-security 2"

	_, err := config.GetText("{{ .ServerAddress }}", c)
	assert.Error(t, err)

	dest := filepath.Join(t.TempDir(), "client.ovpn")
	assert.Error(t, config.SaveToFile("./templates/openvpn-client-config.tpl", c, dest))
	assert.NoFileExists(t, dest)
}

func TestAllowsMultilineKeyMaterial(t *testing.T) {
	c := config.New()
	c.Ca, c.Cert, c.Key, c.Ta = "a\nb", "c\nd", "e\nf", "g\nh"

	_, err := config.GetText("{{ .Ca }}", c)
	assert.NoError(t, err)
}

func renderTemplate(t *testing.T, c config.Config) string {
	t.Helper()
	txt, err := os.ReadFile("./templates/openvpn-client-config.tpl")
	assert.NoError(t, err)
	out, err := config.GetText(string(txt), c)
	assert.NoError(t, err)
	return out
}

func TestTemplateIncludesTLSCryptKey(t *testing.T) {
	c := config.New()
	c.Ta = "ta-key-material"

	assert.Contains(t, renderTemplate(t, c), "<tls-crypt>\nta-key-material\n</tls-crypt>")
}

func TestTemplateOmitsEmptyTLSCryptBlock(t *testing.T) {
	assert.NotContains(t, renderTemplate(t, config.New()), "tls-crypt")
}

func TestRejectsBreakoutFromKeyBlock(t *testing.T) {
	c := config.New()
	c.Cert = "x\n</cert>\nscript-security 2\n<cert>"

	_, err := config.GetText("{{ .Cert }}", c)
	assert.Error(t, err)
}
