package test

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"io"
	"os/exec"
	"sync"
	"testing"

	testssh "github.com/gliderlabs/ssh"
)

func RunKeySshServer(t *testing.T, allowedPublicKey []byte) func() {
	publicKeyOption := testssh.PublicKeyAuth(func(ctx testssh.Context, key testssh.PublicKey) bool {
		allowed, _, _, _, err := ssh.ParseAuthorizedKey(allowedPublicKey)
		require.NoError(t, err)
		return testssh.KeysEqual(allowed, key)
	})
	return runsServer(t, publicKeyOption)
}

func RunBasicSshServer(t *testing.T, password string) func() {
	basicAuthOption := testssh.PasswordAuth(func(ctx testssh.Context, pass string) bool {
		return pass == password
	})
	return runsServer(t, basicAuthOption)
}

func runsServer(t *testing.T, authOption testssh.Option) func() {
	logrus.Info("starting ssh server on port 2222...")
	srv := &testssh.Server{Addr: ":2222", Handler: handlerSSH}
	if err := srv.SetOption(authOption); err != nil {
		require.NoError(t, err)
	}

	go func() {
		logrus.Info(srv.ListenAndServe())
	}()
	return func() {
		err := srv.Shutdown(context.Background())
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

func handlerSSH(s testssh.Session) {
	cmd, stdin, stderr, stdout, err := buildCommand(s.Command())
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, s)
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(s.Stderr(), stderr)
	}()

	go func() {
		defer wg.Done()
		io.Copy(s, stdout)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return
	}
}

func buildCommand(c []string) (cmd *exec.Cmd, stdin io.WriteCloser, stderr, stdout io.ReadCloser, err error) {
	if len(c) != 2 {
		err = fmt.Errorf("invalid command")
		return
	}

	cmd = exec.Command(c[0], c[1])
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}

	stdin, err = cmd.StdinPipe()
	if err != nil {
		return
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		return
	}

	return
}
