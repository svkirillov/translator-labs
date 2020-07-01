package grammar

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
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
		newGrammar.Rules = append(
			newGrammar.Rules,
			Rule{
				LSymbol: gs.Rules[i].LSymbol,
				RSymbol: gs.Rules[i].RSymbol,
			},
		)
	}

	// set terminal symbols
	for i := range gs.TSymbols {
		newGrammar.TTokens = append(
			newGrammar.TTokens,
			TToken{TSymbol: gs.TSymbols[i]},
		)
	}

	// set non terminal symbols
	for i := range gs.NTSymbols {
		var alt []int
		nts := gs.NTSymbols[i]

		for j, r := range newGrammar.Rules {
			if r.LSymbol == nts {
				alt = append(alt, j)
			}
		}

		if len(alt) == 0 {
			return nil, fmt.Errorf("this token is not in the rules: %s", nts)
		}

		newGrammar.NTokens = append(
			newGrammar.NTokens,
			NToken{
				NTSymbol: nts,
				Alt:      alt,
				AltCount: len(alt),
			},
		)
	}

	return &newGrammar, nil
}

func (g *Grammar) FindNToken(token string) int {
	for i, nt := range g.NTokens {
		if nt.NTSymbol == token {
			return i
		}
	}

	return -1
}

func (g *Grammar) IsNTerm(symbol string) int {
	if g.FindNToken(symbol) < 0 {
		return Term
	}

	return NTerm
}

func (g *Grammar) Print() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)

	table.SetHeader([]string{"#", "Rule"})

	data := make([][]string, 0)
	for i, r := range g.Rules {
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

	fmt.Printf("\033[1mStart symbol:\033[0m %s\n", g.Root)

	fmt.Printf("\033[1mTerminal symbols:\033[0m")
	for _, tt := range g.TTokens {
		fmt.Printf(" %s", tt.TSymbol)
	}
	fmt.Println()

	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)

	table.SetHeader([]string{"Symbol", "Qty of alts", "Alternatives"})

	data = make([][]string, 0)
	for _, nt := range g.NTokens {
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
