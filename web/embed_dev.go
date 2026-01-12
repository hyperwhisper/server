//go:build !prod

package web

import "embed"

// DistFS is empty in dev mode - frontend is served via Nuxt dev server
var DistFS embed.FS
