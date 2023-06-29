package e2e

import (
	"math/rand"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"

	"github.com/onsi/gomega"

	// Register tests
	//nolint:goimports
	_ "github.com/loft-sh/devpod-provider-ssh/e2e/tests/smoke"
	//nolint:goimports
	_ "github.com/loft-sh/devpod-provider-ssh/e2e/tests/provider"
	//nolint:goimports
	_ "github.com/loft-sh/devpod-provider-ssh/e2e/tests/integration"
)

// TestRunE2ETests checks configuration parameters (specified through flags) and then runs
// E2E tests using the Ginkgo runner.
// If a "report directory" is specified, one or more JUnit test reports will be
// generated in this directory, and cluster logs will also be saved.
// This function is called on each Ginkgo node in parallel mode.
func TestRunE2ETests(t *testing.T) {
	rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DevPod SSH Provider e2e suite")
}
