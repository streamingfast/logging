package pkg2

import (
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var zlog = zap.NewNop()
var tracer = logging.PackageLogger("example", "github.com/streamingfast/example/logging/pkg2", zlog)
