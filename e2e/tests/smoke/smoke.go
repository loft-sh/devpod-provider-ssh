package smoke

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/loft-sh/devpod/e2e/framework"
	"github.com/onsi/ginkgo/v2"
)

var _ = ginkgo.Describe("[smoke]: devpod provider ssh test suite", ginkgo.Ordered, func() {

	ginkgo.Context("testing /kubeletinfo endpoint", ginkgo.Label("smoke"), ginkgo.Ordered, func() {
		ginkgo.It("should compile the provider", func() {
			cmd := exec.Command("bash", "hack/build.sh")
			cmd.Env = append(cmd.Environ(), "RELEASE_VERSION=0.0.0")
			cmd.Dir = "../"

			err := cmd.Run()
			framework.ExpectNoError(err)

			// Replace binary path in manifest to point to freshly built binaries
			input, err := os.ReadFile("../release/provider.yaml")
			framework.ExpectNoError(err)
			//
			output := bytes.Replace(input, []byte("https://github.com/loft-sh/devpod-provider-ssh/releases/download/0.0.0/"), []byte(os.Getenv("PWD")+"/../release/"), -1)

			err = os.WriteFile("../release/provider.yaml", output, 0666)
			framework.ExpectNoError(err)
		})

		ginkgo.It("should generate ssh keypairs", func() {
			_, err := os.Stat(os.Getenv("HOME") + "/.ssh/id_rsa")
			if err != nil {
				fmt.Println("generating ssh keys")
				cmd := exec.Command("ssh-keygen", "-q", "-t", "rsa", "-N", "", "-f", os.Getenv("HOME")+"/.ssh/id_rsa")
				err = cmd.Run()
				framework.ExpectNoError(err)

				cmd = exec.Command("ssh-keygen", "-y", "-f", os.Getenv("HOME")+"/.ssh/id_rsa")
				output, err := cmd.Output()
				framework.ExpectNoError(err)

				err = os.WriteFile(os.Getenv("HOME")+"/.ssh/id_rsa.pub", output, 0600)
				framework.ExpectNoError(err)
			}

			cmd := exec.Command("ssh-keygen", "-y", "-f", os.Getenv("HOME")+"/.ssh/id_rsa")
			publicKey, err := cmd.Output()
			framework.ExpectNoError(err)

			_, err = os.Stat(os.Getenv("HOME") + "/.ssh/authorized_keys")
			if err != nil {
				err = os.WriteFile(os.Getenv("HOME")+"/.ssh/authorized_keys", publicKey, 0600)
				framework.ExpectNoError(err)
			} else {
				f, err := os.OpenFile(os.Getenv("HOME")+"/.ssh/authorized_keys",
					os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
				framework.ExpectNoError(err)

				defer f.Close()
				_, err = f.Write(publicKey)
				framework.ExpectNoError(err)
			}
		})

		ginkgo.It("should download latest devpod", func() {
			resp, err := http.Get("https://github.com/loft-sh/devpod/releases/latest/download/devpod-linux-amd64")
			framework.ExpectNoError(err)
			defer resp.Body.Close()

			err = os.MkdirAll("bin/", 0755)
			framework.ExpectNoError(err)

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
	})
})
