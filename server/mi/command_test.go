package mi_test

import (
	"bufio"
	"net"
	"strings"
	"testing"

	mi "github.com/OZON08/openvpn-server-config/server/mi"
	"github.com/stretchr/testify/assert"
)

var cResponsePid = `SUCCESS: pid=10869
`

func TestReadResponsePid(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(cResponsePid))
	response, err := mi.ReadResponse(reader)
	assert.Nil(t, err)
	pid, err := mi.ParsePid(response)
	assert.Nil(t, err)
	assert.Equal(t, int64(10869), pid)
}

func TestReadResponseFailure(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(""))
	_, err := mi.ReadResponse(reader)
	assert.NotNil(t, err)
}

func TestSendCommandFailure(t *testing.T) {
	conn := &net.IPConn{}
	err := mi.SendCommand(conn, "dummy")
	assert.NotNil(t, err)
}

func TestReadResponseSkipsAsyncNotifications(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(">HOLD:Waiting for hold release:0\r\n>BYTECOUNT:1,2\r\nSUCCESS: pid=42\r\n"))

	response, err := mi.ReadResponse(reader)

	assert.Nil(t, err)
	pid, err := mi.ParsePid(response)
	assert.Nil(t, err)
	assert.Equal(t, int64(42), pid)
}
