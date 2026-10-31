package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OZON08/openvpn-server-config/server/config"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	c := config.New()
	assert.Equal(t, "SHA512", c.Auth)
}

func TestTemplateGeneration(t *testing.T) {
	c := config.New()
	txt, err := os.ReadFile("./templates/openvpn-server-config.tpl")
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
	c.Device = "tun\nscript-security 2\nup /tmp/evil.sh"

	_, err := config.GetText("{{ .Device }}", c)
	assert.Error(t, err)

	dest := filepath.Join(t.TempDir(), "server.conf")
	assert.Error(t, config.SaveToFile("./templates/openvpn-server-config.tpl", c, dest))
	assert.NoFileExists(t, dest)
}

func TestAllowsMultilineCustomOptions(t *testing.T) {
	c := config.New()
	c.CustomOptOne, c.CustomOptTwo, c.CustomOptThree = "a\nb", "c\nd", "e\nf"

	_, err := config.GetText("{{ .CustomOptOne }}", c)
	assert.NoError(t, err)
}
