/*
Copyright 2019 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package records

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pkg/errors"

	"github.com/fission/fission/pkg/controller/client"
	"github.com/fission/fission/pkg/fission-cli/cliwrapper/cli"
	"github.com/fission/fission/pkg/fission-cli/cmd"
	"github.com/fission/fission/pkg/fission-cli/util"
	redisCache "github.com/fission/fission/pkg/redis/build/gen"
)

type ViewSubCommand struct {
	client *client.Client
}

func View(flags cli.Input) error {
	opts := ViewSubCommand{
		client: cmd.GetServer(flags),
	}
	return opts.do(flags)
}

func (opts *ViewSubCommand) do(flags cli.Input) error {
	return opts.run(flags)
}

func (opts *ViewSubCommand) run(flags cli.Input) error {
	var verbosity int
	if flags.Bool("v") && flags.Bool("vv") {
		return errors.New("conflicting verbosity levels, use either --v or --vv")
	}
	if flags.Bool("v") {
		verbosity = 1
	}
	if flags.Bool("vv") {
		verbosity = 2
	}

	function := flags.String("function")
	trigger := flags.String("trigger")
	from := flags.String("from")
	to := flags.String("to")

	//Refuse multiple filters for now
	if multipleFiltersSpecified(function, trigger, from+to) {
		return errors.New("maximum of one filter is currently supported, either --function, --trigger, or --from,--to")
	}

	if len(function) != 0 {
		return recordsByFunction(function, verbosity, flags)
	}
	if len(trigger) != 0 {
		return recordsByTrigger(trigger, verbosity, flags)
	}
	if len(from) != 0 && len(to) != 0 {
		return recordsByTime(from, to, verbosity, flags)
	}
	err := recordsAll(verbosity, flags)
	if err != nil {
		return errors.Wrap(err, "error viewing records")
	}
	return nil
}

func recordsAll(verbosity int, flags cli.Input) error {
	fc := util.GetApiClient(flags.GlobalString("server"))

	records, err := fc.RecordsAll()
	if err != nil {
		return errors.Wrap(err, "error viewing records")
	}

	showRecords(records, verbosity)

	return nil
}

func recordsByTrigger(trigger string, verbosity int, flags cli.Input) error {
	fc := util.GetApiClient(flags.GlobalString("server"))

	records, err := fc.RecordsByTrigger(trigger)
	if err != nil {
		return errors.Wrap(err, "error viewing records")
	}

	showRecords(records, verbosity)

	return nil
}

// TODO: More accurate function name (function filter)
func recordsByFunction(function string, verbosity int, flags cli.Input) error {
	fc := util.GetApiClient(flags.GlobalString("server"))

	records, err := fc.RecordsByFunction(function)
	if err != nil {
		return errors.Wrap(err, "error viewing records")
	}

	showRecords(records, verbosity)

	return nil
}

func recordsByTime(from string, to string, verbosity int, flags cli.Input) error {
	fc := util.GetApiClient(flags.GlobalString("server"))

	records, err := fc.RecordsByTime(from, to)
	if err != nil {
		return errors.Wrap(err, "error viewing records")
	}

	showRecords(records, verbosity)

	return nil
}

func showRecords(records []*redisCache.RecordedEntry, verbosity int) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	if verbosity == 1 {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n",
			"REQUID", "REQUEST METHOD", "FUNCTION", "RESPONSE STATUS", "TRIGGER")
		for _, record := range records {
			fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n",
				record.ReqUID, record.Req.Method, record.Req.Header["X-Fission-Function-Name"], record.Resp.Status, record.Trigger)
		}
	} else if verbosity == 2 {
		for _, record := range records {
			fmt.Println(record)
		}
	} else {
		fmt.Fprintf(w, "%v\n",
			"REQUID")
		for _, record := range records {
			fmt.Fprintf(w, "%v\n",
				record.ReqUID)
		}
	}
	w.Flush()
}

func multipleFiltersSpecified(entries ...string) bool {
	var specified int
	for _, entry := range entries {
		if len(entry) > 0 {
			specified += 1
		}
	}
	return specified > 1
}