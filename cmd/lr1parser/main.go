package main

import (
	"fmt"
	"os"

	"github.com/svkirillov/translator-labs/pkg/grammar"
	"github.com/svkirillov/translator-labs/pkg/lr1parser"
)

func main() {
	grSettings := grammar.GrammarSettings{
		Root: "S",
		TSymbols: []string{
			"+",
			"*",
			"a",
			"(",
			")",
			"$",
		},
		NTSymbols: []string{
			"E",
			"T",
			"F",
			"S",
		},
		Rules: []grammar.Rule{
			{"S", "E"},
			{"E", "E+T"},
			{"E", "T"},
			{"T", "T*F"},
			{"T", "F"},
			{"F", "(E)"},
			{"F", "a"},
		},
	}

	// grSettings := grammar.GrammarSettings{
	// 	Root: "S",
	// 	TSymbols: []string{
	// 		"c",
	// 		"d",
	// 	},
	// 	NTSymbols: []string{
	// 		"S",
	// 		"E",
	// 		"C",
	// 	},
	// 	Rules: []grammar.Rule{
	// 		{"S", "E"},
	// 		{"E", "CC"},
	// 		{"C", "cC"},
	// 		{"C", "d"},
	// 	},
	// }

	gr, err := grammar.New(grSettings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	gr.Print()

	lr1Parser := lr1parser.NewLR1Parser(*gr, "(a+a)*a*a$")
	if err := lr1Parser.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
