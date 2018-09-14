package main

import (
	"etym/cmd/transformer/command"
	"etym/pkg/log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		println("Usage: transformer <combine-etym|combine-trans|stardict|ecdict|db|merge> [path]")
		return
	}

	cmd := strings.TrimSpace(os.Args[1])
	switch cmd {
	case "combine-etym":
		command.CombineEtym()
	case "combine-trans":
		command.CombineTrans()
	case "stardict":
		command.StarDict()
	case "ecdict":
		command.ECDICT()
	case "merge":
		command.Merge()

	case "db":
		command.DBImport(os.Args[2:])
		//	case "wordroot":
		//	case "lemma":

	default:
		log.Fatalf("command: %s not implemented", cmd)
	}
}
