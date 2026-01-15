//go:build prod

package web

import "embed"

// DistFS embeds the built frontend static files.
// The dist folder is created by running `bun run generate` in the web directory.
// Using all: prefix to include _nuxt directory (go:embed skips _ prefixed dirs by default)
//
//go:embed all:dist
var DistFS embed.FS
