package main

import sybil "github.com/logV/sybil/src/lib"
import cmd "github.com/logV/sybil/src/cmd"
import "os"
import "log"

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Command should be one of: ingest, digest, query, rebuild")
	}

	first_arg := os.Args[1]
	os.Args = os.Args[1:]

	sybil.SetDefaults()

	switch first_arg {
	case "ingest":
		cmd.RunIngestCmdLine()
	case "digest":
		cmd.RunDigestCmdLine()
	case "session":
		cmd.RunSessionizeCmdLine()
	case "trim":
		cmd.RunTrimCmdLine()
	case "query":
		cmd.RunQueryCmdLine()
	case "index":
		cmd.RunIndexCmdLine()
	case "rebuild":
		cmd.RunRebuildCmdLine()
	case "inspect":
		cmd.RunInspectCmdLine()
	default:
		log.Fatal("Unknown command:", os.Args[0])
	}
}
