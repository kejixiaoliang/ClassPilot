package webassets

import "embed"

// Dist contains the production Web bundle created by Vite.
//
//go:embed dist/*
var Dist embed.FS
