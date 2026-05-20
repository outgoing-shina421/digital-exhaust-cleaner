// Package ui embeds compiled frontend assets produced by webpack.
// Run `npm run build` inside the web/ directory to (re)generate these files.
package ui

import "embed"

// StaticFS holds the compiled CSS and JS served under /static/.
//
//go:embed static
var StaticFS embed.FS
