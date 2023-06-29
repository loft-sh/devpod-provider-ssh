package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/loft-sh/devpod/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("[integration]: devpod provider ssh test suite", ginkgo.Ordered, func() {

	ginkgo.Context("testing /kubeletinfo endpoint", ginkgo.Label("integration"), ginkgo.Ordered, func() {
		ginkgo.It("should download latest devpod", func() {
			resp, err := http.Get("https://github.com/loft-sh/devpod/releases/latest/download/devpod-linux-amd64")
			framework.ExpectNoError(err)
			defer resp.Body.Close()

			out, err := os.Create("bin/devpod")
			framework.ExpectNoError(err)
			defer out.Close()

			err = out.Chmod(0755)
			framework.ExpectNoError(err)

			_, err = io.Copy(out, resp.Body)
			framework.ExpectNoError(err)

			err = out.Close()
			framework.ExpectNoError(err)

			// test that devpod works
			cmd := exec.Command("bin/devpod")
			err = cmd.Run()
			framework.ExpectNoError(err)
		})

		ginkgo.It("should add provider to devpod", func() {
			// ensure we don't have the ssh provider present
			cmd := exec.Command("bin/devpod", "provider", "delete", "ssh")
			err := cmd.Run()
			if err != nil {
				fmt.Println("warning: " + err.Error())
			}

			cmd = exec.Command("bin/devpod", "provider", "add", "../release/provider.yaml", "-o", "HOST=localhost")
			err = cmd.Run()
			framework.ExpectNoError(err)
		})

		ginkgo.It("should run devpod up", func() {
			// ensure we don't have the ssh provider present
			cmd := exec.Command("bin/devpod", "up", "--debug", "--ide=none", "../")
			err := cmd.Run()
			framework.ExpectNoError(err)
		})

		ginkgo.It("should run commands to workspace via ssh", func() {
			// ensure we don't have the ssh provider present
			cmd := exec.Command("ssh", "devpod-provider-ssh.devpod", "echo", "test")
			output, err := cmd.Output()
			framework.ExpectNoError(err)

			gomega.Expect(output).To(gomega.Equal([]byte("test\n")))
		})

		ginkgo.It("should cleanup devpod workspace", func() {
			// ensure we don't have the ssh provider present
			cmd := exec.Command("bin/devpod", "delete", "--debug", "--force", "devpod-provider-ssh")
			err := cmd.Run()
			framework.ExpectNoError(err)
		})
	})
})
