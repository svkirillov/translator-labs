package main

import (
	"fmt"
	"os"

	"github.com/svkirillov/translator-labs/pkg/grammar"
	"github.com/svkirillov/translator-labs/pkg/lrparser"
)

func main() {
	grSettings := grammar.GrammarSettings{
		Root: "B",
		TSymbols: []string{
			"+",
			"*",
			"a",
			"b",
		},
		NTSymbols: []string{
			"B",
			"T",
			"M",
		},
		Rules: []grammar.Rule{
			{"B", "T+B"},
			{"B", "T"},
			{"T", "M"},
			{"T", "M*T"},
			{"M", "a"},
			{"M", "b"},
			{"M", "(B)"},
		},
	}

	gr, err := grammar.New(grSettings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	gr.Print()

	lrParser := lrparser.NewLRParser(*gr, "a+b")
	if err := lrParser.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
