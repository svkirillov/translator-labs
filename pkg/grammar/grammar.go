package grammar

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/svkirillov/translator-labs/pkg/helpers"
)

// Type of token
const (
	Term = iota
	NTerm
)

// Grammar's rule
type Rule struct {
	LSymbol string
	RSymbol string
}

// Terminal symbol
type TToken struct {
	TSymbol string
}

// Non terminal symbol
type NToken struct {
	NTSymbol string
	Alt      []int
	AltCount int
}

type Grammar struct {
	Root    string
	TTokens []TToken
	NTokens []NToken
	Rules   []Rule
}

type GrammarSettings struct {
	Root      string
	TSymbols  []string
	NTSymbols []string
	Rules     []Rule
}

func New(gs GrammarSettings) (*Grammar, error) {
	newGrammar := Grammar{}

	// set root symbol
	newGrammar.Root = gs.Root

	// set rules
	for i := range gs.Rules {
		ls := gs.Rules[i].LSymbol
		rs := gs.Rules[i].RSymbol

		if ls == rs {
			return nil, fmt.Errorf("wrong rule: '%s -> %s'", ls, rs)
		}

		newGrammar.Rules = append(
			newGrammar.Rules,
			Rule{
				LSymbol: ls,
				RSymbol: rs,
			},
		)
	}

	// set terminal symbols
	ts := helpers.Unique(gs.TSymbols)
	for i := range ts {
		newGrammar.TTokens = append(
			newGrammar.TTokens,
			TToken{TSymbol: ts[i]},
		)
	}

	// set non terminal symbols
	nts := helpers.Unique(gs.NTSymbols)
	for i := range nts {
		var alt []int
		ntsym := nts[i]

		for j, r := range newGrammar.Rules {
			if r.LSymbol == ntsym {
				alt = append(alt, j)
			}
		}

		if len(alt) == 0 {
			return nil, fmt.Errorf("this token is not in the rules: %s", ntsym)
		}

		newGrammar.NTokens = append(
			newGrammar.NTokens,
			NToken{
				NTSymbol: ntsym,
				Alt:      alt,
				AltCount: len(alt),
			},
		)
	}

	return &newGrammar, nil
}

func (gr *Grammar) FindNToken(token string) int {
	for i, nt := range gr.NTokens {
		if nt.NTSymbol == token {
			return i
		}
	}

	return -1
}

func (gr *Grammar) TokenType(symbol string) int {
	if gr.FindNToken(symbol) < 0 {
		return Term
	}

	return NTerm
}

func (gr *Grammar) Print() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)

	table.SetHeader([]string{"#", "Rule"})

	data := make([][]string, 0)
	for i, r := range gr.Rules {
		data = append(
			data,
			[]string{
				fmt.Sprintf("%d", i),
				fmt.Sprintf("%s -> %s", r.LSymbol, r.RSymbol),
			},
		)
	}

	table.AppendBulk(data)

	fmt.Println("\033[1mRules:\033[0m")
	table.Render()

	fmt.Printf("\033[1mStart symbol:\033[0m %s\n", gr.Root)

	fmt.Printf("\033[1mTerminal symbols:\033[0m")
	for _, tt := range gr.TTokens {
		fmt.Printf(" %s", tt.TSymbol)
	}
	fmt.Println()

	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)

	table.SetHeader([]string{"Symbol", "Qty of alts", "Alternatives"})

	data = make([][]string, 0)
	for _, nt := range gr.NTokens {
		data = append(
			data,
			[]string{
				nt.NTSymbol,
				fmt.Sprintf("%d", nt.AltCount),
				fmt.Sprintf("%v", nt.Alt),
			},
		)
	}

	table.AppendBulk(data)

	fmt.Println("\033[1mNon terminal symbols:\033[0m")
	table.Render()
}
