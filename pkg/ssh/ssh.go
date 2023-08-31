package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

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
	sshError := "Please make sure you have configured the correct SSH host\nand the following command can be executed on your system:\n"
	return fmt.Errorf(sshError + "ssh " + strings.Join(getSSHCommand(provider), " ") + " " + command)
}

func getSSHCommand(provider *SSHProvider) []string {
	result := []string{"-oStrictHostKeyChecking=no", "-oBatchMode=yes"}

	if provider.Config.Port != "22" {
		result = append(result, []string{"-p", provider.Config.Port}...)
	}

	if provider.Config.ExtraFlags != "" {
		result = append(result, strings.Split(provider.Config.ExtraFlags, " ")...)
	}

	result = append(result, provider.Config.Host)
	return result
}

func execSSHCommand(provider *SSHProvider, command string, output io.Writer) error {
	commandToRun := getSSHCommand(provider)
	commandToRun = append(commandToRun, command)

	cmd := exec.Command("ssh", commandToRun...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
			return fmt.Errorf(
				agentDir + " is not writable, passwordless sudo or root user required",
			)
		}
	}

	// check that we have access to DOCKER_PATH
	err = execSSHCommand(provider, provider.Config.DockerPath+" ps", out)
	if err != nil {
		err = execSSHCommand(provider, "sudo -nl", out)
		if err != nil {
			return fmt.Errorf(
				provider.Config.DockerPath + " not found, passwordless sudo or root user required",
			)
		}
	}

	return nil
}

func Command(provider *SSHProvider, command string) error {
	return execSSHCommand(provider, command, os.Stdout)
}
