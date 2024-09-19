package e2e

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"

	"fxp/test/e2e/framework"
	"fxp/test/e2e/framework/consts"
)

var _ = ginkgo.Describe("[Feature: Example]", func() {
	f := framework.NewDefaultFramework()

	ginkgo.Describe("TCP", func() {
		ginkgo.It("Expose a TCP echo server", func() {
			serverConf := consts.DefaultServerConfig
			clientConf := consts.DefaultClientConfig

			remotePort := f.AllocPort()
			clientConf += fmt.Sprintf(`
			[[proxies]]
			name = "tcp"
			type = "tcp"
			localPort = {{ .%s }}
			remotePort = %d
			`, framework.TCPEchoServerPort, remotePort)

			f.RunProcesses([]string{serverConf}, []string{clientConf})

			framework.NewRequestExpect(f).Port(remotePort).Ensure()
		})
	})
})
