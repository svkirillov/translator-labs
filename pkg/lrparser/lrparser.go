package lrparser

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/svkirillov/translator-labs/pkg/grammar"
)

type LRParser struct {
	grammar    *grammar.Grammar
	input      string
	l1Stack    []l1StackNode
	l2Stack    []l2StackNode
	state      int
	production []int
	inputIter  int

	printer *tablewriter.Table
}

type l1StackNode struct {
	token     string
	tokenType int
	altCount  int // number of alternative in rules
	altNum    int // current number of alternative in rules
}

type l2StackNode struct {
	token     string
	tokenType int
}

// State const
const (
	normal = iota
	ret
	end
)

func (lrp *LRParser) pushL1Stack(l1Token l1StackNode) {
	newL1Stack := make([]l1StackNode, 1)
	newL1Stack[0] = l1Token
	newL1Stack = append(newL1Stack, lrp.l1Stack...)
	lrp.l1Stack = newL1Stack
}

func (lrp *LRParser) pushL2Stack(l2Token l2StackNode) {
	newL2Stack := make([]l2StackNode, len(l2Token.token))

	for i, s := range l2Token.token {
		newL2Stack[i].token = string(s)
		newL2Stack[i].tokenType = lrp.grammar.IsNTerm(string(s))
	}

	newL2Stack = append(newL2Stack, lrp.l2Stack...)
	lrp.l2Stack = newL2Stack
}

func NewLRParser(gr *grammar.Grammar, in string) LRParser {
	var l2Stack []l2StackNode
	for _, nt := range gr.NTokens {
		if nt.NTSymbol == gr.Root {
			l2Stack = append(
				l2Stack,
				l2StackNode{
					token:     nt.NTSymbol,
					tokenType: grammar.NTerm,
				},
			)
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetHeader([]string{"State", "L1", "L2", "Input"})

	return LRParser{
		grammar:    gr,
		input:      in,
		l1Stack:    nil,
		l2Stack:    l2Stack,
		state:      normal,
		production: nil,
		inputIter:  0,
		printer:    table,
	}
}

func (lrp *LRParser) expandTree() {
	symbol := lrp.l2Stack[0].token
	nToken := lrp.grammar.NTokens[lrp.grammar.FindNToken(symbol)]
	l1Token := l1StackNode{
		token:     nToken.NTSymbol,
		tokenType: grammar.NTerm,
		altCount:  nToken.AltCount,
		altNum:    1,
	}

	lrp.pushL1Stack(l1Token)

	numRule := nToken.Alt[l1Token.altNum-1]
	ruleRSymbol := lrp.grammar.Rules[numRule].RSymbol
	l2token := l2StackNode{
		token:     ruleRSymbol,
		tokenType: lrp.grammar.IsNTerm(ruleRSymbol),
	}

	lrp.l2Stack = lrp.l2Stack[1:]
	lrp.pushL2Stack(l2token)
}

func (lrp *LRParser) pushL2NodeToL1Stack() {
	lrp.inputIter++

	l1Token := l1StackNode{
		token:     lrp.l2Stack[0].token,
		tokenType: grammar.Term,
		altCount:  0,
		altNum:    1,
	}

	lrp.pushL1Stack(l1Token)

	lrp.l2Stack = lrp.l2Stack[1:]
}

func (lrp *LRParser) pushL1NodeToL2Stack() {
	lrp.inputIter--

	l2Token := l2StackNode{
		token:     lrp.l1Stack[0].token,
		tokenType: grammar.Term,
	}

	lrp.pushL2Stack(l2Token)

	lrp.l1Stack = lrp.l1Stack[1:]
}

func (lrp *LRParser) successfulCompletion() {
	lrp.state = end

	for _, l1Token := range lrp.l1Stack {
		if l1Token.tokenType == grammar.Term {
			continue
		}

		it := lrp.grammar.FindNToken(l1Token.token)
		r := lrp.grammar.NTokens[it].Alt[l1Token.altNum-1]

		lrp.production = append(lrp.production, r)
	}

	for i := len(lrp.production)/2 - 1; i >= 0; i-- {
		j := len(lrp.production) - 1 - i
		lrp.production[i], lrp.production[j] = lrp.production[j], lrp.production[i]
	}
}

func (lrp *LRParser) testAlternative() {
	lrp.state = normal

	lrp.l1Stack[0].altNum++

	tokenIndex := lrp.grammar.FindNToken(lrp.l1Stack[0].token)
	ruleNum := lrp.grammar.NTokens[tokenIndex].Alt[lrp.l1Stack[0].altNum-1]

	ruleRSymbol := lrp.grammar.Rules[ruleNum].RSymbol
	orRule := lrp.grammar.Rules[ruleNum-1].RSymbol
	lrp.l2Stack = lrp.l2Stack[len(orRule):]
	lrp.pushL2Stack(
		l2StackNode{
			token:     ruleRSymbol,
			tokenType: -1,
		},
	)
}

func (lrp *LRParser) returnNonTerm() {
	tokenIndex := lrp.grammar.FindNToken(lrp.l1Stack[0].token)
	ruleNum := lrp.grammar.NTokens[tokenIndex].Alt[lrp.l1Stack[0].altNum-1]
	ruleRSymbol := lrp.grammar.Rules[ruleNum].RSymbol
	ruleLSymbol := lrp.grammar.Rules[ruleNum].LSymbol
	lrp.l2Stack = lrp.l2Stack[len(ruleRSymbol):]
	lrp.pushL2Stack(
		l2StackNode{
			token:     ruleLSymbol,
			tokenType: -1,
		},
	)

	lrp.l1Stack = lrp.l1Stack[1:]
}

func (lrp *LRParser) updateTable() {
	var state string
	switch lrp.state {
	case normal:
		state = "\033[32mnormal\033[0m"
	case ret:
		state = "\033[31mret\033[0m"
	case end:
		state = "end"
	}

	var l1Stack string
	for i := range lrp.l1Stack {
		var index string
		if lrp.grammar.IsNTerm(lrp.l1Stack[i].token) == grammar.NTerm {
			index = getIndex(lrp.l1Stack[i].altNum)
		} else {
			index = ""
		}
		l1Stack = fmt.Sprintf("%s%s", lrp.l1Stack[i].token+index, l1Stack)
	}

	var l2Stack string

	for i := range lrp.l2Stack {
		l2Stack += lrp.l2Stack[i].token
	}
	// if len(lrp.l2Stack) > 0 {
	// 	l2Stack = lrp.l2Stack[0].token
	// } else {
	// 	l2Stack = "e"
	// }

	var pointer string
	pointer = lrp.input[lrp.inputIter:]

	lrp.printer.AppendBulk([][]string{{state, l1Stack, l2Stack, pointer}})
}

func getIndex(num int) string {
	var index string
	for {
		switch num % 10 {
		case 0:
			index = "₀" + index
		case 1:
			index = "₁" + index
		case 2:
			index = "₂" + index
		case 3:
			index = "₃" + index
		case 4:
			index = "₄" + index
		case 5:
			index = "₅" + index
		case 6:
			index = "₆" + index
		case 7:
			index = "₇" + index
		case 8:
			index = "₈" + index
		case 9:
			index = "₉" + index
		}

		num /= 10

		if num == 0 {
			break
		}
	}

	return index
}

func (lrp *LRParser) StartParse() error {
	if len(lrp.input) == 0 {
		return fmt.Errorf("input srtring is empty")
	}

	lrp.updateTable()

	for {
		switch lrp.state {
		case normal:
			switch {
			case lrp.l2Stack[0].tokenType == grammar.NTerm:
				lrp.expandTree()
				lrp.updateTable()
				continue

			case lrp.l2Stack[0].tokenType == grammar.Term && lrp.l2Stack[0].token != string(lrp.input[lrp.inputIter]):
				lrp.state = ret
				lrp.updateTable()
				continue

			case lrp.l2Stack[0].tokenType == grammar.Term && lrp.l2Stack[0].token == string(lrp.input[lrp.inputIter]):
				lrp.pushL2NodeToL1Stack()
				lrp.updateTable()
				if lrp.inputIter == len(lrp.input) {
					switch len(lrp.l2Stack) {
					case 0:
						lrp.successfulCompletion()
						lrp.updateTable()
						continue
					default:
						lrp.state = ret
						lrp.updateTable()
						continue
					}
				} else {
					switch len(lrp.l2Stack) {
					case 0:
						lrp.state = ret
						lrp.updateTable()
						continue
					default:
						continue
					}
				}
			}

		case ret:
			switch {
			case lrp.l1Stack[0].tokenType == grammar.Term:
				lrp.pushL1NodeToL2Stack()
				lrp.updateTable()
				continue
			case lrp.l1Stack[0].tokenType == grammar.NTerm && lrp.l1Stack[0].altNum < lrp.l1Stack[0].altCount:
				lrp.testAlternative()
				lrp.updateTable()
				continue
			case lrp.l1Stack[0].tokenType == grammar.NTerm && lrp.l1Stack[0].altNum >= lrp.l1Stack[0].altCount:
				if lrp.l1Stack[0].token == lrp.grammar.Root && lrp.inputIter == 0 {
					return fmt.Errorf("the input string does not belong to the grammar")
				} else {
					lrp.returnNonTerm()
					lrp.updateTable()
					continue
				}
			}

		case end:
			fmt.Println("\033[1mSteps:\033[0m")
			lrp.printer.Render()
			fmt.Printf("\033[1mLeft out:\033[0m %d\n", lrp.production)
			return nil
		}
	}
}
