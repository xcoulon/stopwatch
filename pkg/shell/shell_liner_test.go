package shell_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("shell", func() {
	var session *gexec.Session

	BeforeEach(func() {
		shellPath := build()
		session = run(shellPath)
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})

	It("exits with status code 0", func() {
		Eventually(session).Should(gexec.Exit(0))
	})

	It("prints 'Hello World' to stdout", func() {
		Eventually(session).Should(gbytes.Say("Hello World"))
		Eventually(session).Should(gbytes.Say("Hello World"))
	})
})

func build() string {
	p, err := gexec.Build("github.com/vatriathlon/stopwatch")
	Expect(err).NotTo(HaveOccurred())

	return p
}

func run(path string) *gexec.Session {
	cmd := exec.Command(path, "shell")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}
