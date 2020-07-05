package lr1parser

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/svkirillov/translator-labs/pkg/grammar"
	"github.com/svkirillov/translator-labs/pkg/helpers"
)

const (
	accept = iota
	shift
	reduce
	err
)

type LR1Parser struct {
	grammar     *grammar.Grammar
	input       string
	stateStack  []int
	production  []int
	actionTable []map[string]state
	gotoTable   []map[string]int
	inputIter   int

	printer *tablewriter.Table
}

type item struct {
	Rule      grammar.Rule
	RuleNum   int
	Position  int
	Lookahead string
}

type state struct {
	action int
	st     int
}

func NewLR1Parser(gr grammar.Grammar, in string) LR1Parser {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(true)

	return LR1Parser{
		grammar:     &gr,
		input:       in,
		stateStack:  []int{0},
		production:  nil,
		actionTable: nil,
		gotoTable:   nil,
		inputIter:   0,
		printer:     table,
	}
}

func (lr1p *LR1Parser) stackPush(t int) {
	old := lr1p.stateStack
	lr1p.stateStack = make([]int, 1)
	lr1p.stateStack[0] = t
	lr1p.stateStack = append(lr1p.stateStack, old...)
}

func (lr1p *LR1Parser) stackPop(n int) {
	lr1p.stateStack = lr1p.stateStack[n:]
}

func (lr1p *LR1Parser) first_(token string) []string {
	if token == "$" {
		return []string{token}
	}
	if lr1p.grammar.TokenType(token) == grammar.Term {
		return []string{token}
	}

	f := make([]string, 0)

	for i := 0; i < len(lr1p.grammar.Rules); i++ {
		if lr1p.grammar.Rules[i].LSymbol == token && lr1p.grammar.Rules[i].RSymbol[0:1] != token {
			fn := lr1p.first_(lr1p.grammar.Rules[i].RSymbol[0:1])
			f = append(f, fn...)
		}
	}

	return helpers.Unique(f)
}

func (lr1p *LR1Parser) first(tokens string) []string {
	f := make([]string, 0)

	if tokens == "$" {
		return []string{tokens}
	}

	if string(tokens[len(tokens)-1]) == "$" {
		tokens = tokens[:len(tokens)-1]
	}

	for i := 0; i < len(tokens); i++ {
		fn := lr1p.first_(tokens[i : i+1])
		if len(fn) != 0 {
			f = append(f, fn...)
		}
	}

	return helpers.Unique(f)
}

func (lr1p *LR1Parser) closure(items []item) []item {
	it := items[:]
	currentLen := len(it)
	oldLen := 0

	for {
		for i := oldLen; i < currentLen; i++ {
			position := it[i].Position
			if position >= len(it[i].Rule.RSymbol) {
				continue
			}

			token := it[i].Rule.RSymbol[position : position+1]
			if lr1p.grammar.TokenType(token) == grammar.Term {
				continue
			}

			tokenIndex := lr1p.grammar.FindNToken(token)

			tail := ""
			if position+1 >= len(it[i].Rule.RSymbol) {
				tail = it[i].Lookahead
			} else {
				tail = it[i].Rule.RSymbol[position+1:] + it[i].Lookahead
			}
			f := lr1p.first(tail)

			for j := range lr1p.grammar.NTokens[tokenIndex].Alt {
				ruleNum := lr1p.grammar.NTokens[tokenIndex].Alt[j]
				for k := range f {
					item := item{
						Rule:      lr1p.grammar.Rules[ruleNum],
						RuleNum:   ruleNum,
						Position:  0,
						Lookahead: f[k],
					}
					if !checkItemIn(&item, it) {
						it = append(
							it,
							item,
						)
					}
				}
			}
		}

		if currentLen == len(it) {
			break
		} else {
			oldLen = currentLen
			currentLen = len(it)
		}
	}

	return it
}

func (lr1p *LR1Parser) goTo(items []item, token string) []item {
	j := make([]item, 0)

	for i := range items {
		position := items[i].Position

		if position+1 > len(items[i].Rule.RSymbol) {
			continue
		}

		if items[i].Rule.RSymbol[position:position+1] == token {
			j = append(
				j,
				item{
					Rule:      items[i].Rule,
					RuleNum:   items[i].RuleNum,
					Position:  items[i].Position + 1,
					Lookahead: items[i].Lookahead,
				},
			)
		}
	}

	return lr1p.closure(j)
}

func setsEqual(s1 []item, s2 []item) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		flag := false

		for j := 0; j < len(s2); j++ {
			if itemEqual(&s1[i], &s2[j]) {
				flag = true
				break
			}
		}

		if !flag {
			return false
		}
	}

	return true
}

func itemEqual(i1 *item, i2 *item) bool {
	return i1.RuleNum == i2.RuleNum && i1.Position == i2.Position && i1.Lookahead == i2.Lookahead
}

func checkItemIn(it *item, items []item) bool {
	for i := range items {
		if itemEqual(it, &items[i]) {
			return true
		}
	}
	return false
}

func (lr1p *LR1Parser) items() [][]item {
	closures := [][]item{
		lr1p.closure(
			[]item{
				{
					Rule:      lr1p.grammar.Rules[0],
					RuleNum:   0,
					Position:  0,
					Lookahead: "$",
				},
			},
		),
	}

	var allSymbols []string
	for i := range lr1p.grammar.TTokens {
		allSymbols = append(allSymbols, lr1p.grammar.TTokens[i].TSymbol)
	}
	for i := range lr1p.grammar.NTokens {
		allSymbols = append(allSymbols, lr1p.grammar.NTokens[i].NTSymbol)
	}

	currentLen := len(closures)
	oldLen := 0

	for {
		for i := oldLen; i < currentLen; i++ {
		l1:
			for j := range allSymbols {
				gt := lr1p.goTo(closures[i], allSymbols[j])

				if len(gt) == 0 {
					continue
				}

				for k := range closures {
					if setsEqual(gt, closures[k]) {
						continue l1
					}
				}

				closures = append(closures, gt)
			}
		}

		if currentLen == len(closures) {
			break
		} else {
			oldLen = currentLen
			currentLen = len(closures)
		}
	}

	return closures
}

func (lr1p *LR1Parser) buildTable() {
	closures := lr1p.items()
	tTokens := lr1p.grammar.TTokens
	ntTokens := lr1p.grammar.NTokens

	lr1p.actionTable = make([]map[string]state, len(closures))
	lr1p.gotoTable = make([]map[string]int, len(closures))

	for i := range closures {
		items := closures[i]

		lr1p.actionTable[i] = make(map[string]state)
		lr1p.gotoTable[i] = make(map[string]int)

		for j := range tTokens {
			lr1p.actionTable[i][tTokens[j].TSymbol] = state{
				action: err,
			}
		}

		for j := range items {
			position := items[j].Position

			if position == len(items[j].Rule.RSymbol) && items[j].Rule.LSymbol == lr1p.grammar.Root && items[j].Lookahead == "$" {
				lr1p.actionTable[i][items[j].Lookahead] = state{
					action: accept,
				}
				continue
			}

			if position == len(items[j].Rule.RSymbol) && items[j].Rule.LSymbol != lr1p.grammar.Root {
				lr1p.actionTable[i][items[j].Lookahead] = state{
					action: reduce,
					st:     items[j].RuleNum,
				}
				continue
			}

			symbol := items[j].Rule.RSymbol[position : position+1]

			if lr1p.grammar.TokenType(symbol) == grammar.Term {
				for k := range closures {
					if setsEqual(lr1p.goTo(items, symbol), closures[k]) {
						lr1p.actionTable[i][symbol] = state{
							action: shift,
							st:     k,
						}
					}
				}
			}
		}

		for j := range ntTokens {
			nts := ntTokens[j].NTSymbol
			for k := range closures {
				if setsEqual(lr1p.goTo(items, nts), closures[k]) {
					lr1p.gotoTable[i][nts] = k
					break
				} else {
					lr1p.gotoTable[i][nts] = -1
				}
			}
		}
	}

	data := make([][]string, 1+len(closures))
	data[0] = make([]string, 1+len(ntTokens)+len(tTokens))
	data[0][0] = "State"
	for i := 1; i < 1+len(closures); i++ {
		data[i] = make([]string, 1+len(ntTokens)+len(tTokens))
		data[i][0] = fmt.Sprintf("%d", i)
	}
	for i := range lr1p.grammar.TTokens {
		for j := 0; j < len(closures); j++ {
			action := lr1p.actionTable[j][lr1p.grammar.TTokens[i].TSymbol].action
			state := lr1p.actionTable[j][lr1p.grammar.TTokens[i].TSymbol].st
			var str string
			switch action {
			case accept:
				str = fmt.Sprintf("\033[1;32m\u2714\033[0m")
			case shift:
				str = fmt.Sprintf("\033[1;33ms%d\033[0m", state)
			case reduce:
				str = fmt.Sprintf("\033[1;34mr%d\033[0m", state)
			default:
				str = ""
			}
			data[1+j][1+i] = str
		}
		data[0][1+i] = lr1p.grammar.TTokens[i].TSymbol
	}
	for i := range lr1p.grammar.NTokens {
		for j := 0; j < len(closures); j++ {
			state := lr1p.gotoTable[j][lr1p.grammar.NTokens[i].NTSymbol]
			if state >= 0 {
				data[1+j][1+len(lr1p.grammar.TTokens)+i] = fmt.Sprintf("%d", state)
			}
		}
		data[0][1+len(lr1p.grammar.TTokens)+i] = lr1p.grammar.NTokens[i].NTSymbol
	}

	lr1p.printer.AppendBulk(data)
	lr1p.printer.Render()
}

func (lr1p *LR1Parser) Parse() error {
	lr1p.production = make([]int, 0)

	lr1p.buildTable()

l1:
	for {
		s := lr1p.stateStack[0]
		a := lr1p.input[lr1p.inputIter : lr1p.inputIter+1]
		act, ok := lr1p.actionTable[s][a]
		if !ok {
			return fmt.Errorf("error")
		}

		switch act.action {
		case shift:
			lr1p.stackPush(act.st)
			lr1p.inputIter++
		case reduce:
			rule := lr1p.grammar.Rules[act.st]
			lr1p.stackPop(len(rule.RSymbol))
			s = lr1p.stateStack[0]
			lr1p.stackPush(lr1p.gotoTable[s][rule.LSymbol])
			lr1p.production = append(lr1p.production, act.st)
		case accept:
			break l1
		default:
			return fmt.Errorf("error")
		}
	}

	fmt.Println(lr1p.production)

	return nil
}
