//go:build prod

package web

import "embed"

// DistFS embeds the built frontend static files.
// The dist folder is created by running `bun run generate` in the web directory.
//
//go:embed dist
var DistFS embed.FS
