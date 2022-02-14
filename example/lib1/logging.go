package lib1

import (
	"github.com/streamingfast/logging"
)

// We simulate here a file that would be actually pulled as a library inside a go project.
// Since it's a different project, we use a completely new "shortName" value for the
// `PackageLogger` invocation.
var zlog, tracer = logging.PackageLogger("lib", "github.com/streamingfast/lib")
