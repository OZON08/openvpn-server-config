package mi_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	mi "github.com/OZON08/openvpn-server-config/server/mi"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	welcome = ">INFO:OpenVPN Management Interface Version 5\r\n"
	prompt  = "ENTER PASSWORD:"
)

// serveMI runs handler for one connection on a fake Management Interface and
// returns the client together with a channel that closes when handler is done.
func serveMI(t *testing.T, handler func(net.Conn, *bufio.Reader)) (*mi.Client, <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		handler(conn, bufio.NewReader(conn))
	}()
	return mi.NewClient("tcp", ln.Addr().String()), done
}

func TestExecuteWorksOverUnixSocket(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mi.sock")
	ln, err := net.Listen("unix", path)
	require.NoError(t, err)
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		io.WriteString(conn, welcome)
		bufio.NewReader(conn).ReadString('\n')
		io.WriteString(conn, "SUCCESS: pid=42\r\n")
	}()

	pid, err := mi.NewClient("unix", path).GetPid()

	require.NoError(t, err)
	assert.Equal(t, int64(42), pid)
}

func TestSendCommandSendsPercentVerbatim(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	go mi.SendCommand(client, "kill a%sb")

	line, err := bufio.NewReader(server).ReadString('\n')

	require.NoError(t, err)
	assert.Equal(t, "kill a%sb\n", line)
}

func TestSendCommandRejectsLineBreaks(t *testing.T) {
	for _, cmd := range []string{"status\nsignal SIGTERM", "status\rsignal SIGTERM"} {
		client, server := net.Pipe()
		err := mi.SendCommand(client, cmd)
		assert.Error(t, err, "%q", cmd)
		client.Close()
		server.Close()
	}
}

func TestExecuteNeverSendsInjectedCommand(t *testing.T) {
	connected := make(chan struct{}, 1)
	c, _ := serveMI(t, func(net.Conn, *bufio.Reader) { connected <- struct{}{} })

	_, err := c.Execute("status\nsignal SIGTERM")

	assert.Error(t, err)
	assert.Empty(t, connected, "invalid command must not open a connection")
}

func TestExecuteLogsInWithPassword(t *testing.T) {
	c, done := serveMI(t, func(conn net.Conn, r *bufio.Reader) {
		io.WriteString(conn, prompt)
		pw, _ := r.ReadString('\n')
		if pw != "s3cret\n" {
			io.WriteString(conn, "ERROR: bad password\r\n")
			return
		}
		io.WriteString(conn, "SUCCESS: password is correct\r\n"+welcome)
		cmd, _ := r.ReadString('\n')
		if cmd == "pid\n" {
			io.WriteString(conn, "SUCCESS: pid=42\r\n")
		}
	})
	c.MIPassword = "s3cret"

	pid, err := c.GetPid()

	<-done
	require.NoError(t, err)
	assert.Equal(t, int64(42), pid)
}

func TestExecuteFailsOnWrongPassword(t *testing.T) {
	commandSeen := make(chan bool, 1)
	c, done := serveMI(t, func(conn net.Conn, r *bufio.Reader) {
		io.WriteString(conn, prompt)
		r.ReadString('\n')
		// OpenVPN 2.7 answers a wrong password only with a new prompt, no line ending
		io.WriteString(conn, prompt)
		_, err := r.ReadString('\n')
		commandSeen <- err == nil
	})
	c.MIPassword = "wrong"

	_, err := c.Execute("pid")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password rejected")
	assert.NotContains(t, err.Error(), "wrong")
	<-done
	assert.False(t, <-commandSeen, "command must not be sent after failed login")
}

func TestExecuteFailsWhenServerDoesNotPrompt(t *testing.T) {
	c, done := serveMI(t, func(conn net.Conn, r *bufio.Reader) {
		io.WriteString(conn, welcome)
	})
	c.MIPassword = "s3cret"

	_, err := c.Execute("pid")

	assert.Error(t, err)
	<-done
}

func TestInvalidInputIsRejectedBeforeConnecting(t *testing.T) {
	c := mi.NewClient("tcp", "127.0.0.1:1")

	_, err := c.KillSession("a\nb")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "newline")
}

func TestExecuteReportsMissingPassword(t *testing.T) {
	c, done := serveMI(t, func(conn net.Conn, r *bufio.Reader) {
		io.WriteString(conn, prompt)
	})

	_, err := c.Execute("pid")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "MIPassword")
	<-done
}

func TestRejectedCommandIsNotLogged(t *testing.T) {
	var logs bytes.Buffer
	level, out := log.GetLevel(), log.StandardLogger().Out
	log.SetLevel(log.DebugLevel)
	log.SetOutput(&logs)
	defer func() { log.SetLevel(level); log.SetOutput(out) }()
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	assert.Error(t, mi.SendCommand(client, "x\nforged"))
	assert.NotContains(t, logs.String(), "forged")
}

func TestFormattingClientHidesPassword(t *testing.T) {
	c := mi.NewClient("tcp", "127.0.0.1:2080")
	c.MIPassword = "s3cret"

	assert.NotContains(t, fmt.Sprintf("%v %+v", c, *c), "s3cret")
}

func TestPasswordIsNotLogged(t *testing.T) {
	var logs bytes.Buffer
	level, out := log.GetLevel(), log.StandardLogger().Out
	log.SetLevel(log.DebugLevel)
	log.SetOutput(&logs)
	defer func() { log.SetLevel(level); log.SetOutput(out) }()

	c, done := serveMI(t, func(conn net.Conn, r *bufio.Reader) {
		io.WriteString(conn, prompt)
		r.ReadString('\n')
		io.WriteString(conn, "SUCCESS: password is correct\r\n"+welcome)
		r.ReadString('\n')
		io.WriteString(conn, "SUCCESS: pid=42\r\n")
	})
	c.MIPassword = "s3cret"

	_, err := c.GetPid()

	<-done
	require.NoError(t, err)
	assert.True(t, strings.Contains(logs.String(), "pid"), "debug logging should be active")
	assert.NotContains(t, logs.String(), "s3cret")
}

func TestMissingPasswordIsReportedWithoutWaiting(t *testing.T) {
	release := make(chan struct{})
	c, done := serveMI(t, func(conn net.Conn, r *bufio.Reader) {
		io.WriteString(conn, prompt)
		<-release // the real server keeps the connection open
	})

	start := time.Now()
	_, err := c.Execute("pid")
	elapsed := time.Since(start)
	close(release)
	<-done

	require.Error(t, err)
	assert.Contains(t, err.Error(), "MIPassword")
	assert.Less(t, elapsed, 2*time.Second)
}

func TestPasswordStaysOutOfGoSyntaxAndJSON(t *testing.T) {
	c := mi.NewClient("tcp", "127.0.0.1:2080")
	c.MIPassword = "s3cret"

	encoded, err := json.Marshal(c)

	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "s3cret")
	assert.NotContains(t, fmt.Sprintf("%#v %#v", c, *c), "s3cret")
}
