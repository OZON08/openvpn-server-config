package clientconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	clientconfig "github.com/OZON08/openvpn-server-config/client/client-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateGeneration(t *testing.T) {
	txt, err := os.ReadFile("./templates/openvpn-client-config.tpl")
	require.NoError(t, err)

	out, err := clientconfig.GetText(string(txt), clientconfig.New())
	require.NoError(t, err)
	assert.Contains(t, out, "remote-cert-tls server")
}

func TestRejectsDirectiveInjection(t *testing.T) {
	c := clientconfig.New()
	c.ServerAddress = "vpn.example.org\nscript-security 2"

	_, err := clientconfig.GetText("{{ .ServerAddress }}", c)
	assert.Error(t, err)

	dest := filepath.Join(t.TempDir(), "client.ovpn")
	assert.Error(t, clientconfig.SaveToFile("./templates/openvpn-client-config.tpl", c, dest))
	assert.NoFileExists(t, dest)
}

func TestAllowsMultilineKeysAndCustomOptions(t *testing.T) {
	c := clientconfig.New()
	c.Ca, c.Cert, c.Key, c.Ta = "a\nb", "c\nd", "e\nf", "g\nh"
	c.CustomConfOne, c.CustomConfTwo, c.CustomConfThree = "i\nj", "k\nl", "m\nn"

	_, err := clientconfig.GetText("{{ .Ca }}", c)
	assert.NoError(t, err)
}

func TestRejectsBreakoutFromKeyBlock(t *testing.T) {
	c := clientconfig.New()
	c.Key = "x\n</key>\nscript-security 2\n<key>"

	_, err := clientconfig.GetText("{{ .Key }}", c)
	assert.Error(t, err)
}

func TestTemplateOmitsEmptyTLSCryptBlock(t *testing.T) {
	txt, err := os.ReadFile("./templates/openvpn-client-config.tpl")
	require.NoError(t, err)

	out, err := clientconfig.GetText(string(txt), clientconfig.New())

	require.NoError(t, err)
	assert.NotContains(t, out, "tls-crypt")
}
