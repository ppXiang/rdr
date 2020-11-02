// Copyright 2017 XUEQIU.COM
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/urfave/cli"

	"fmt"

	"github.com/xueqiu/rdr/decoder"
	"github.com/xueqiu/rdr/dump"

	"strconv"
)

//go:generate go-bindata -prefix "static/" -o=static/static.go -pkg=static -ignore static.go static/...
//go:generate go-bindata -prefix "views/" -o=views/views.go -pkg=views -ignore views.go views/...

// keys is function for command `keys`
// output all keys in rdbfile(s) get from args
func keys(c *cli.Context) {
	if c.NArg() < 1 {
		fmt.Fprintln(c.App.ErrWriter, "keys requires at least 1 argument")
		cli.ShowCommandHelp(c, "keys")
		return
	}
	for _, filepath := range c.Args() {
		decoder := decoder.NewDecoder()
		go dump.Decode(c, decoder, filepath)
		for e := range decoder.Entries {
			fmt.Fprintf(c.App.Writer, "%v\n", e.Key)
		}
	}
}

func topkey(c *cli.Context) {
	var (
		keysNums      int    = 50
		keysDelimiter string = "[:DMR:]"
		minSize       uint64
		filepath      string
	)
	if c.NArg() == 1 {
		filepath = c.Args()[0]
	} else if c.NArg() == 2 {
		filepath = c.Args()[0]
		keysNums, _ = strconv.Atoi(c.Args()[1])
	} else if c.NArg() == 3 {
		filepath = c.Args()[0]
		keysNums, _ = strconv.Atoi(c.Args()[1])
		minSize, _ = strconv.ParseUint(c.Args()[2], 10, 64)
	} else if c.NArg() == 4 {
		keysDelimiter = c.Args()[3]
		minSize, _ = strconv.ParseUint(c.Args()[2], 10, 64)
		keysNums, _ = strconv.Atoi(c.Args()[1])
		filepath = c.Args()[0]
	} else if c.NArg() == 5 {
		logfile, err := os.OpenFile(c.Args()[4], os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			fmt.Fprintln(c.App.ErrWriter, "%v\n", err)
			cli.ShowCommandHelp(c, "topkeys")
			return
		}
		c.App.Writer = logfile
		keysDelimiter = c.Args()[3]
		minSize, _ = strconv.ParseUint(c.Args()[2], 10, 64)
		keysNums, _ = strconv.Atoi(c.Args()[1])
		filepath = c.Args()[0]
	} else {
		fmt.Fprintln(c.App.ErrWriter, "keys requires at least 1 argument")
		cli.ShowCommandHelp(c, "topkey")
		return
	}

	decoder := decoder.NewDecoder()
	go dump.Decode(c, decoder, filepath)
	//init Counter
	counter := dump.NewCounter()
	counter.Count(decoder.Entries)

	topKeyList := counter.GetLargestEntries(keysNums)
	for i := 0; i < len(topKeyList); i++ {
		if topKeyList[i].Bytes >= minSize {
			fmt.Fprintf(c.App.Writer, "%v%s%v%s%v\n", topKeyList[i].Key, keysDelimiter, topKeyList[i].Bytes, keysDelimiter, topKeyList[i].Type)
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "rdr"
	app.Usage = "a tool to parse redis rdbfile"
	app.Version = "v0.0.1"
	app.Writer = os.Stdout
	app.ErrWriter = os.Stderr
	app.Commands = []cli.Command{
		cli.Command{
			Name:      "dump",
			Usage:     "dump statistical information of rdbfile to STDOUT",
			ArgsUsage: "FILE1 [FILE2] [FILE3]...",
			Action:    dump.ToCliWriter,
		},
		cli.Command{
			Name:      "show",
			Usage:     "show statistical information of rdbfile by webpage",
			ArgsUsage: "DIR1 [DIR2] [DIR3] or FILE1 [FILE2] [FILE3]...",
			Flags: []cli.Flag{
				cli.UintFlag{
					Name:  "port, p",
					Value: 8080,
					Usage: "Port for rdr to listen",
				},
			},
			Action: dump.Show,
		},
		cli.Command{
			Name:      "keys",
			Usage:     "get all keys from rdbfile",
			ArgsUsage: "FILE1 [FILE2] [FILE3]...",
			Action:    keys,
		},
		cli.Command{
			Name:      "topkey",
			Usage:     "get the top list of key size from rdbfile",
			ArgsUsage: "RDBFILE | RDBFILE TOPNUMS | RDBFILE TOPNUMS MINSIZE | RDBFILE DELIMITER TOPNUMS STDOUTFILE",
			Action:    topkey,
		},
	}
	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.ErrWriter, "command %q can not be found.\n", command)
		cli.ShowAppHelp(c)
	}
	app.Run(os.Args)
}
