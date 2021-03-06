package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	flaghelper "github.com/hashicorp/nomad/helper/flag-helpers"
	"github.com/jrasell/levant/levant"
	"github.com/jrasell/levant/logging"
)

// DispatchCommand is the command implementation that allows users to
// dispatch a Nomad job.
type DispatchCommand struct {
	args []string
	Meta
}

// Help provides the help information for the dispatch command.
func (c *DispatchCommand) Help() string {
	helpText := `
Usage: levant dispatch [options] <parameterized job> [input source]

  Dispatch creates an instance of a parameterized job. A data payload to the
  dispatched instance can be provided via stdin by using "-" or by specifying a
  path to a file. Metadata can be supplied by using the meta flag one or more
  times. 

General Options:

  -address=<http_address>
    The Nomad HTTP API address including port which Levant will use to make
    calls.

  -log-level=<level>
    Specify the verbosity level of Levant's logs. Valid values include DEBUG,
    INFO, and WARN, in decreasing order of verbosity. The default is INFO.

Dispatch Options:

  -meta <key>=<value>
    Meta takes a key/value pair separated by "=". The metadata key will be
    merged into the job's metadata. The job may define a default value for the
    key which is overridden when dispatching. The flag can be provided more 
    than once to inject multiple metadata key/value pairs. Arbitrary keys are
    not allowed. The parameterized job must allow the key to be merged.
`
	return strings.TrimSpace(helpText)
}

// Synopsis is provides a brief summary of the dispatch command.
func (c *DispatchCommand) Synopsis() string {
	return "Dispatch an instance of a parameterized job"
}

// Run triggers a run of the Levant dispatch functions.
func (c *DispatchCommand) Run(args []string) int {

	var meta []string
	var addr, logLevel string

	flags := c.Meta.FlagSet("dispatch", FlagSetVars)
	flags.Usage = func() { c.UI.Output(c.Help()) }
	flags.Var((*flaghelper.StringFlag)(&meta), "meta", "")
	flags.StringVar(&addr, "address", "", "")
	flags.StringVar(&logLevel, "log-level", "INFO", "")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if l := len(args); l < 1 || l > 2 {
		c.UI.Error(c.Help())
		return 1
	}

	logging.SetLevel(logLevel)

	job := args[0]
	var payload []byte
	var readErr error

	if len(args) == 2 {
		switch args[1] {
		case "-":
			payload, readErr = ioutil.ReadAll(os.Stdin)
		default:
			payload, readErr = ioutil.ReadFile(args[1])
		}
		if readErr != nil {
			c.UI.Error(fmt.Sprintf("Error reading input data: %v", readErr))
			return 1
		}
	}

	metaMap := make(map[string]string, len(meta))
	for _, m := range meta {
		split := strings.SplitN(m, "=", 2)
		if len(split) != 2 {
			c.UI.Error(fmt.Sprintf("Error parsing meta value: %v", m))
			return 1
		}
		metaMap[split[0]] = split[1]
	}

	success := levant.TriggerDispatch(job, metaMap, payload, addr)
	if !success {
		return 1
	}

	return 0
}
