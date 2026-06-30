package webs

import _ "embed"

//go:embed nav-rain-grid-web.zip
var staticFile []byte

func Static() []byte {
	return staticFile
}
