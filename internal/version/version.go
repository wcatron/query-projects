package version

// Version returns the current version of the CLI
func Version() string {
	return cliVersion
}

// cliVersion is set at build time via ldflags
var cliVersion string
