// Package version carries build identity for the aido binary.
package version

// Version is the semantic version of this build. It is overridden at link time
// by the Makefile (-X). Keep the fallback in sync with the latest release tag.
var Version = "0.1.0-dev"

// Commit is the git revision this binary was built from, set at link time.
var Commit = "unknown"

// BuildDate is an RFC3339 timestamp, set at link time.
var BuildDate = "unknown"
