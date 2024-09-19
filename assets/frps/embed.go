package frpc

import (
	"embed"

	"fxp/assets"
)

//go:embed static/*
var content embed.FS

func init() {
	assets.Register(content)
}
