package config

import (
	"errors"

	"github.com/OZON08/openvpn-server-config/internal/render"
)

var defaultConfig = Config{
	Device:            "tun",
	Proto:             "udp",
	ServerAddress:     "127.0.0.1",
	Port:              1194,
	OpenVpnServerPort: "1194",
	Cipher:            "AES-256-GCM",
	Auth:              "SHA512",
	Ca:                "ca.crt",
}

// Config model
type Config struct {
	Device            string
	ServerAddress     string
	Port              int
	OpenVpnServerPort string
	Proto             string

	Ca   string
	Cert string
	Key  string
	Ta   string

	Cipher string
	Auth   string
}

// New returns config object with default values
func New() Config {
	return defaultConfig
}

// GetText injects config values into template
func GetText(tpl string, c Config) (string, error) {
	return render.Text(tpl, c)
}

// SaveToFile reads template and writes result to destination file
func SaveToFile(tplPath string, c Config, destPath string) error {
	return render.SaveToFile(tplPath, c, destPath)
}

// Validate rejects line breaks outside the key material fields and values that
// could close an inline key block
func (c Config) Validate() error {
	return errors.Join(
		render.RequireSingleLine(c, "Ca", "Cert", "Key", "Ta"),
		render.RequireKeyMaterial(c.Ca, c.Cert, c.Key, c.Ta),
	)
}
