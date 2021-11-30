package main

import (
	"fmt"
	"time"

	"github.com/streamingfast/logging"
	"github.com/streamingfast/logging/example/lib1"
	"github.com/streamingfast/logging/example/pkg1"
	"github.com/streamingfast/logging/example/pkg2"

	"go.uber.org/zap"
)

var zlog = zap.NewNop()
var tracer = logging.ApplicationLogger("example", "github.com/streamingfast/logging/example", &zlog,
	// By default active only when a production environment is detected (e.g. when file '/.dockerenv' file exists).
	// But for the sake of the example here, with forcecully activate it.
	logging.WithSwitcherServerAutoStart(),
)

func main() {
	fmt.Println("This demo show case some features of this logging library.")
	fmt.Println("It is simulating an application using multiple sub-packages")
	fmt.Println("and consuming an external library that also uses this logging")
	fmt.Println("library to configure its logger(s).")
	fmt.Println("")
	fmt.Println("What you will see by default")
	fmt.Println("")
	fmt.Println("Info level for all logger that share the same 'shortName' as the")
	fmt.Println("application logger ('example' in our case).")
	fmt.Println("")
	fmt.Println("Environment variable override")
	fmt.Println("")
	fmt.Println("If you restart the demo using:")
	fmt.Println(" - DEBUG=example ...                                                 You will activate DEBUG logging for all loggers that have the 'shortName' example (so main, pkg1 and pkg2)")
	fmt.Println(" - DEBUG=(example|lib) ...                                           You will activate DEBUG logging for loggers whose 'shortName match the regex (so main, pkg1, pkg2 and lib)")
	fmt.Println(" - DEBUG=github.com/streamingfast/example/logging/pkg2 ...           You will activate DEBUG logging for specific logger (so pkg2)")
	fmt.Println(" - DEBUG=github.com/streamingfast/example/logging/(pkg1|pkg2) ...    You will activate DEBUG logging for loggers whose id match the regex (so pkg1 and pkg2)")
	fmt.Println("")
	fmt.Println("On the fly switcher")
	fmt.Println("")
	fmt.Println("The automatic level switcher is active in this demo, you can use it to")
	fmt.Println("change level on the fly.")
	fmt.Println("")
	fmt.Println(`  curl -d '{"inputs":"example","level":"debug"}' http://locahost:1065`)
	fmt.Println("")
	fmt.Println("Accepted 'inputs' is the same as the environment variable above.")

	for {
		zlog.Info("Info level from 'main'")

		time.Sleep(2 * time.Second)
		zlog.Debug("Debug level from 'main'")

		if tracer.Enabled() {
			zlog.Debug("Trace level from 'main'")
		}

		go pkg1.Invoke()
		go pkg2.Invoke()
		go lib1.Invoke()

		time.Sleep(4 * time.Second)
	}
}
