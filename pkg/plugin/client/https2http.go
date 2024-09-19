// Copyright 2019 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !frps

package plugin

import (
	"crypto/tls"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/fatedier/golib/pool"

	v1 "fxp/pkg/config/v1"
	"fxp/pkg/transport"
	"fxp/pkg/util/log"
	netpkg "fxp/pkg/util/net"
)

func init() {
	Register(v1.PluginHTTPS2HTTP, NewHTTPS2HTTPPlugin)
}

type HTTPS2HTTPPlugin struct {
	opts *v1.HTTPS2HTTPPluginOptions

	l *Listener
	s *http.Server
}

func NewHTTPS2HTTPPlugin(options v1.ClientPluginOptions) (Plugin, error) {
	opts := options.(*v1.HTTPS2HTTPPluginOptions)
	listener := NewProxyListener()

	p := &HTTPS2HTTPPlugin{
		opts: opts,
		l:    listener,
	}

	rp := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
			r.SetXForwarded()
			req := r.Out
			req.URL.Scheme = "http"
			req.URL.Host = p.opts.LocalAddr
			if p.opts.HostHeaderRewrite != "" {
				req.Host = p.opts.HostHeaderRewrite
			}
			for k, v := range p.opts.RequestHeaders.Set {
				req.Header.Set(k, v)
			}
		},
		BufferPool: pool.NewBuffer(32 * 1024),
		ErrorLog:   stdlog.New(log.NewWriteLogger(log.WarnLevel, 2), "", 0),
	}

	var (
		tlsConfig *tls.Config
		err       error
	)
	if opts.CrtPath != "" || opts.KeyPath != "" {
		tlsConfig, err = p.genTLSConfig()
	} else {
		tlsConfig, err = transport.NewServerTLSConfig("", "", "")
		tlsConfig.InsecureSkipVerify = true
	}
	if err != nil {
		return nil, fmt.Errorf("gen TLS config error: %v", err)
	}

	p.s = &http.Server{
		Handler:           rp,
		ReadHeaderTimeout: 60 * time.Second,
		TLSConfig:         tlsConfig,
	}

	go func() {
		_ = p.s.ServeTLS(listener, "", "")
	}()
	return p, nil
}

func (p *HTTPS2HTTPPlugin) genTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(p.opts.CrtPath, p.opts.KeyPath)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	return config, nil
}

func (p *HTTPS2HTTPPlugin) Handle(conn io.ReadWriteCloser, realConn net.Conn, extra *ExtraInfo) {
	wrapConn := netpkg.WrapReadWriteCloserToConn(conn, realConn)
	if extra.SrcAddr != nil {
		wrapConn.SetRemoteAddr(extra.SrcAddr)
	}
	_ = p.l.PutConn(wrapConn)
}

func (p *HTTPS2HTTPPlugin) Name() string {
	return v1.PluginHTTPS2HTTP
}

func (p *HTTPS2HTTPPlugin) Close() error {
	return p.s.Close()
}
