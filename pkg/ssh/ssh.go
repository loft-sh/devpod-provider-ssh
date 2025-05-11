package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/loft-sh/devpod-provider-ssh/pkg/options"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/ssh"
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

	sshcmd, err := getSSHCommand(provider)
	if err != nil {
		return err
	}

	return fmt.Errorf("%s ssh %s %s", sshError, strings.Join(sshcmd, " "), command)
}

func getSSHCommand(provider *SSHProvider) ([]string, error) {
	result := []string{"-oStrictHostKeyChecking=no", "-oBatchMode=yes"}

	if provider.Config.Port != "22" {
		result = append(result, []string{"-p", provider.Config.Port}...)
	}

	if provider.Config.ExtraFlags != "" {
		flags, err := shellquote.Split(provider.Config.ExtraFlags)
		if err != nil {
			return nil, fmt.Errorf("error managing EXTRA_ARGS, %v", err)
		}

		result = append(result, flags...)
	}

	result = append(result, provider.Config.Host)
	return result, nil
}

func execSSHCommand(provider *SSHProvider, command string, output io.Writer) error {
	// The SSH protocol's command is a single string which is interpreted by
	// the remote user's default login shell. This causes problems if the
	// remote shell is not POSIX-compliant. To get around this, we wrap the
	// command in a call to sh -c which should always give a POSIX shell.
	command = fmt.Sprintf("sh -c %s", shellquote.Join(command))

	if provider.Config.UseBuiltinSSH {
		// get ssh config for host
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		sshConfig, err := exec.CommandContext(ctx, "ssh", "-G", provider.Config.Host).Output()
		if err != nil {
			return fmt.Errorf("read ssh config for host %s: %w", provider.Config.Host, err)
		}
		hostname, user, port, identityfile := parseConfig(string(sshConfig))
		if hostname == "" || user == "" || port == "" {
			return fmt.Errorf("resolve ssh config. Hostname='%s', User='%s', Port='%s'", hostname, user, port)
		}

		// expand identityfile path
		if strings.HasPrefix(identityfile, "~") {
			identityfile = strings.Replace(identityfile, "~", "$userprofile", 1)
			identityfile = os.ExpandEnv(identityfile)
		}
		abs, err := filepath.Abs(identityfile)
		if err != nil {
			return fmt.Errorf("absolute filepath: %w", err)
		}
		key, err := os.ReadFile(abs)
		if err != nil {
			return fmt.Errorf("read identifiyfile: %w", err)
		}

		if provider.Config.Port != "" {
			port = provider.Config.Port
		}
		// create ssh session
		addr := net.JoinHostPort(hostname, port)
		client, err := ssh.NewSSHClient(user, addr, key)
		if err != nil {
			return fmt.Errorf("create ssh client: %w", err)
		}
		sess, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("create ssh session: %w", err)
		}
		sess.Stdin = os.Stdin
		sess.Stdout = output

		return sess.Run(command)
	}

	commandToRun, err := getSSHCommand(provider)
	if err != nil {
		return err
	}

	commandToRun = append(commandToRun, command)

	var stderrBuf bytes.Buffer

	cmd := exec.Command("ssh", commandToRun...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = output
	cmd.Stderr = io.Writer(&stderrBuf)

	err = cmd.Run()
	if err != nil {
		provider.Log.Error(stderrBuf.String())
		return err
	}

	// A non-POSIX shell has been detected: falling back to copy and execute scripts
	if strings.Contains(stderrBuf.String(), "fish: Unsupported") {
		provider.Log.Warn("A non-POSIX shell has been detected: falling back to copy and execute scripts")

		return copyAndExecSSHCommand(provider, command, output)
	}

	return err
}

func copyAndExecSSHCommand(provider *SSHProvider, command string, output io.Writer) error {
	commandToRun, err := getSSHCommand(provider)
	if err != nil {
		return err
	}

	script, err := copyCommandToRemote(provider, command)
	if err != nil {
		return err
	}

	commandToRun = append(commandToRun, []string{
		"/bin/sh", script, ";", "rm", "-f", script,
	}...)

	cmd := exec.Command("ssh", commandToRun...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func copyCommandToRemote(provider *SSHProvider, command string) (string, error) {
	script, err := os.CreateTemp("", "devpod-command-*")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = script.Close()
		_ = os.Remove(script.Name())
	}()

	commandToRun, err := getSCPCommand(provider, script.Name())
	if err != nil {
		return "", err
	}

	_, err = script.WriteString(command)
	if err != nil {
		return "", err
	}

	return script.Name(), exec.Command("scp", commandToRun...).Run()
}

func getSCPCommand(provider *SSHProvider, sourcefile string) ([]string, error) {
	result := []string{"-oStrictHostKeyChecking=no", "-oBatchMode=yes"}

	if provider.Config.Port != "22" {
		result = append(result, []string{"-p", provider.Config.Port}...)
	}

	if provider.Config.ExtraFlags != "" {
		flags, err := shellquote.Split(provider.Config.ExtraFlags)
		if err != nil {
			return nil, fmt.Errorf("error managing EXTRA_ARGS, %v", err)
		}

		result = append(result, flags...)
	}

	destfile := "/tmp/" + filepath.Base(sourcefile)

	result = append(result, sourcefile)
	result = append(result, provider.Config.Host+":"+destfile)
	return result, nil
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

	// We only support running on Linux ssh servers
	out = new(bytes.Buffer)
	err = execSSHCommand(provider, "uname", out)
	if err != nil {
		return returnSSHError(provider, "uname")
	}
	if out.String() != "Linux\n" {
		fmt.Println(out.String())
		return fmt.Errorf("error: SSH provider only works on Linux servers")
	}

	// If we're root, we won't have problems
	out = new(bytes.Buffer)
	err = execSSHCommand(provider, "id -ru", out)
	if err != nil {
		return returnSSHError(provider, "id -ru")
	}
	if out.String() == "0\n" {
		return nil
	}

	// check that we have access to AGENT_PATH
	out = new(bytes.Buffer)
	agentDir := path.Dir(provider.Config.AgentPath)
	err1 := execSSHCommand(provider, "mkdir -p "+agentDir, out)
	err2 := execSSHCommand(provider, "test -w "+agentDir, out)
	if err1 != nil || err2 != nil {
		err = execSSHCommand(provider, "sudo -nl", out)
		if err != nil {
			return fmt.Errorf("%s is not writable, passwordless sudo or root user required", agentDir)
		}
	}

	// check that we have access to DOCKER_PATH
	err = execSSHCommand(provider, provider.Config.DockerPath+" ps", out)
	if err != nil {
		err = execSSHCommand(provider, "sudo -nl", out)
		if err != nil {
			return fmt.Errorf("%s not found, passwordless sudo or root user required. If using another user please add to the docker group", provider.Config.DockerPath)
		}
	}

	return nil
}

func Command(provider *SSHProvider, command string) error {
	return execSSHCommand(provider, command, os.Stdout)
}

func parseConfig(config string) (hostname string, user string, port string, identityfile string) {
	for _, line := range strings.Split(config, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if fields[0] == "hostname" {
			hostname = fields[1]
			continue
		}
		if fields[0] == "user" {
			user = fields[1]
			continue
		}
		if fields[0] == "port" {
			port = fields[1]
			continue
		}
		// just take the first one
		if fields[0] == "identityfile" && identityfile == "" {
			identityfile = fields[1]
			continue
		}
	}

	return hostname, user, port, identityfile
}
