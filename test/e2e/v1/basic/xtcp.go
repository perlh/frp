package basic

import (
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"

	"fxp/test/e2e/framework"
	"fxp/test/e2e/framework/consts"
	"fxp/test/e2e/pkg/port"
	"fxp/test/e2e/pkg/request"
)

var _ = ginkgo.Describe("[Feature: XTCP]", func() {
	f := framework.NewDefaultFramework()

	ginkgo.It("Fallback To STCP", func() {
		serverConf := consts.DefaultServerConfig
		clientConf := consts.DefaultClientConfig

		bindPortName := port.GenName("XTCP")
		clientConf += fmt.Sprintf(`
			[[proxies]]
			name = "foo"
			type = "stcp"
			localPort = {{ .%s }}

			[[visitors]]
			name = "foo-visitor"
			type = "stcp"
			serverName = "foo"
			bindPort = -1

			[[visitors]]
			name = "bar-visitor"
			type = "xtcp"
			serverName = "bar"
			bindPort = {{ .%s }}
			keepTunnelOpen = true
			fallbackTo = "foo-visitor"
			fallbackTimeoutMs = 200
			`, framework.TCPEchoServerPort, bindPortName)

		f.RunProcesses([]string{serverConf}, []string{clientConf})
		framework.NewRequestExpect(f).
			RequestModify(func(r *request.Request) {
				r.Timeout(time.Second)
			}).
			PortName(bindPortName).
			Ensure()
	})
})
