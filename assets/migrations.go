package assets

import "embed"

// Migrations contains the SQL migration files.
//
//go:embed migrations/*.sql
var Migrations embed.FS
