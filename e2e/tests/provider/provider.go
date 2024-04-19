package provider

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/loft-sh/devpod/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("[e2e]: devpod provider ssh test suite", ginkgo.Ordered, func() {
	ginkgo.Context("testing /kubeletinfo endpoint", ginkgo.Label("e2e"), ginkgo.Ordered, func() {
		ginkgo.It("should fail the init", func() {
			cmd := exec.Command("../release/devpod-provider-ssh-linux-amd64", "init")
			cmd.Env = append(cmd.Environ(), []string{
				"AGENT_PATH=/tmp/devpod/agent",
				"COMMAND=ls",
				"DOCKER_PATH=docker",
				"HOST=localhost",
				"PORT=1234",
			}...)
			err := cmd.Run()
			framework.ExpectError(err)
		})

		ginkgo.It("should run the init", func() {
			cmd := exec.Command("../release/devpod-provider-ssh-linux-amd64", "init")
			cmd.Env = append(cmd.Environ(), []string{
				"AGENT_PATH=/tmp/devpod/agent",
				"COMMAND=ls",
				"DOCKER_PATH=docker",
				"HOST=localhost",
				"PORT=22",
			}...)
			err := cmd.Run()
			framework.ExpectNoError(err)
		})

		ginkgo.It("should run a command", func() {
			cmd := exec.Command("../release/devpod-provider-ssh-linux-amd64", "command")
			cmd.Env = append(cmd.Environ(), []string{
				"AGENT_PATH=/tmp/devpod/agent",
				"COMMAND=ls",
				"DOCKER_PATH=docker",
				"HOST=localhost",
				"PORT=22",
			}...)
			err := cmd.Run()
			framework.ExpectNoError(err)
		})

		ginkgo.It("should run a command and verify the output", func() {
			cmd := exec.Command("ls", "/")
			controlOutput, err := cmd.Output()
			framework.ExpectNoError(err)

			cmd = exec.Command("../release/devpod-provider-ssh-linux-amd64", "command")
			cmd.Env = append(cmd.Environ(), []string{
				"AGENT_PATH=/tmp/devpod/agent",
				"COMMAND=ls /",
				"DOCKER_PATH=docker",
				"HOST=localhost",
				"PORT=22",
			}...)
			output, err := cmd.Output()
			framework.ExpectNoError(err)

			gomega.Expect(output).To(gomega.Equal(controlOutput))
		})

		ginkgo.It("should run a multiline command and verify the output", func() {
			cmd := exec.Command("echo", `line1
line2
line3`)
			controlOutput, err := cmd.Output()
			framework.ExpectNoError(err)

			os.Setenv("COMMAND", `echo line1
echo line2
echo line3`)

			cmd = exec.Command("../release/devpod-provider-ssh-linux-amd64", "command")
			cmd.Env = append(cmd.Environ(), []string{
				"AGENT_PATH=/tmp/devpod/agent",
				"DOCKER_PATH=docker",
				"COMMAND=" + `echo line1
echo line2
echo line3`,
				"HOST=localhost",
				"PORT=22",
			}...)
			output, err := cmd.CombinedOutput()
			framework.ExpectNoError(err)

			gomega.Expect(output).To(gomega.Equal(controlOutput))
		})

		ginkgo.It("should run a failing command and fail", func() {
			controlOutput := []byte("bash: line 1: not-a-command: command not found")

			cmd := exec.Command("../release/devpod-provider-ssh-linux-amd64", "command")
			cmd.Env = append(cmd.Environ(), []string{
				"AGENT_PATH=/tmp/devpod/agent",
				"COMMAND=not-a-command",
				"DOCKER_PATH=docker",
				"HOST=localhost",
				"PORT=22",
			}...)
			output, err := cmd.CombinedOutput()
			framework.ExpectError(err)

			output = bytes.TrimSpace(output)

			gomega.Expect(output).To(gomega.Equal(controlOutput))
		})
	})
})
