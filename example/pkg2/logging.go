package pkg2

import (
	"github.com/streamingfast/logging"
)

var zlog, tracer = logging.PackageLogger("example", "github.com/streamingfast/example/logging/pkg2")
