package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	devpodssh "github.com/loft-sh/devpod/pkg/ssh"
	"github.com/pkg/errors"

	"github.com/loft-sh/devpod-provider-ssh/pkg/options"
	"github.com/loft-sh/devpod/pkg/log"
)

type SSHProvider struct {
	Config           *options.Options
	Log              log.Logger
	WorkingDirectory string
}

func NewProvider(logs log.Logger) (*SSHProvider, error) {
	config, err := options.FromEnv()
	if err != nil {
		return nil, err
	}

	// create provider
	provider := &SSHProvider{
		Config: config,
		Log:    logs,
	}

	return provider, nil
}

func returnSSHError(provider *SSHProvider, command string) error {
	var sshError string = "Please make sure you have configured the correct SSH host\nand the following command can be executed on your system:\n"
	return fmt.Errorf(sshError + "ssh" + strings.Join(getSSHCommand(provider), " ") + " " + command)
}

func getSSHCommand(provider *SSHProvider) []string {
	result := []string{"-oStrictHostKeyChecking=no",
		"-p", provider.Config.Port}

	if provider.Config.ExtraFlags != "" {
		result = append(result, provider.Config.ExtraFlags)
	}

	result = append(result, provider.Config.Host)
	return result
}

func execSSHCommand(provider *SSHProvider, command string, output io.Writer) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	user := currentUser.Username
	ip := provider.Config.Host

	if strings.Contains(provider.Config.Host, "@") {
		user = strings.Split(provider.Config.Host, "@")[0]
		ip = strings.Split(provider.Config.Host, "@")[1]
	}

	if strings.HasPrefix(provider.Config.PrivateKeyPath, "~/") {
		provider.Config.PrivateKeyPath = filepath.Join(currentUser.HomeDir, provider.Config.PrivateKeyPath[2:])
	}

	privateKey, err := os.ReadFile(provider.Config.PrivateKeyPath)
	if err != nil {
		return errors.Wrap(err, "read private ssh key")
	}

	sshClient, err := devpodssh.NewSSHClient(user, ip+":"+provider.Config.Port, privateKey)
	if err != nil {
		return errors.Wrap(err, "create ssh client")
	}
	defer sshClient.Close()

	// run command
	return devpodssh.Run(context.Background(), sshClient, command, os.Stdin, output, os.Stderr)
}

func Init(provider *SSHProvider) error {

	out := new(bytes.Buffer)
	// check that we can do outputs
	err := execSSHCommand(provider, "echo Devpod Test", out)
	if err != nil {
		return returnSSHError(provider, "echo Devpod Test")
	}
	if out.String() != "Devpod Test\n" {
		return fmt.Errorf("error: ssh output mismatch")
	}

	// If we're root, we won't have problems
	err = execSSHCommand(provider, "id -ru", out)
	if err != nil {
		return returnSSHError(provider, "id -ru")
	}
	if out.String() == "0\n" {
		return nil
	}

	// check that we have access to AGENT_PATH
	agentDir := path.Dir(provider.Config.AgentPath)
	err1 := execSSHCommand(provider, "mkdir -p "+agentDir, out)
	err2 := execSSHCommand(provider, "test -w "+agentDir, out)
	if err1 != nil || err2 != nil {
		err = execSSHCommand(provider, "sudo -nl", out)
		if err != nil {
			return fmt.Errorf(agentDir + " is not writable, passwordless sudo or root user required")
		}
	}

	// check that we have access to DOCKER_PATH
	err = execSSHCommand(provider, provider.Config.DockerPath+" ps", out)
	if err != nil {
		err = execSSHCommand(provider, "sudo -nl", out)
		if err != nil {
			return fmt.Errorf(provider.Config.DockerPath + " not found, passwordless sudo or root user required")
		}
	}

	return nil
}

func Command(provider *SSHProvider, command string) error {
	return execSSHCommand(provider, command, os.Stdout)
}
