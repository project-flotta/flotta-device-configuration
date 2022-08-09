package ignition_test

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/project-flotta/flotta-device-configuration/pkg/ignition"
)

var (
	sampleConfig = `
{
  "ignition": { "version": "3.0.0" },
  "storage": {
    "files": [{
      "path": "/tmp/someconfig",
      "mode": 420,
      "contents": { "source": "data:,example%20file%0A" },
      "overwrite": true
    }]
  },
  "systemd": {
    "units": [{
      "name": "flotta.service",
      "enabled": true,
      "contents": "[Service]\nExecStart=/usr/bin/sleep 1h\n\n[Install]\nWantedBy=multi-user.target"
    }]
  }
}
`
)

var _ = Describe("Ignition", func() {

	execCommand := func(cmd string, exitCode int) {
		command := exec.Command("bash", "-c", fmt.Sprintf("/usr/bin/systemctl %s", cmd))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		session.Wait()
		fmt.Printf("Content: %s \n", session.Out.Contents())
		ExpectWithOffset(1, session.ExitCode()).Should(Equal(exitCode))
	}

	AfterEach(func() {
		err := os.RemoveAll("/opt/someconfig")
		Expect(err).NotTo(HaveOccurred())

		execCommand("disable --now flotta.service", 0)

		err = os.RemoveAll("/etc/systemd/system/flotta.service")
		Expect(err).NotTo(HaveOccurred())
	})

	It("Test", func() {
		Expect(true).To(BeTrue())

		cfg, err := ignition.ParseConfig(sampleConfig)
		Expect(cfg).NotTo(BeNil())
		Expect(err).NotTo(HaveOccurred())

		// Checking that files are not in place
		Expect("/opt/someconfig").ShouldNot(BeAnExistingFile())
		execCommand("is-active flotta.service", 3)
		execCommand("is-enabled flotta.service", 1)

		err = ignition.RunConfig(cfg)
		Expect(err).NotTo(HaveOccurred())

		Expect("/opt/someconfig").Should(BeAnExistingFile())

		execCommand("is-active flotta.service", 0)
		execCommand("is-enabled flotta.service", 0)
	})

})
