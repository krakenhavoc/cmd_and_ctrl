package mcpseat

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// DefaultAbsorb is the Layer A rules the seat answers for the agent
// (§4): forced moves, mana-only windows, coin calls and the opening
// roll's die (ADR 0121: a d20 to roll has nothing to choose). `same-land`
// is not among them: playing a land is an action whose timing can matter
// (landfall, a discard outlet), and the owner's list ("empty priority
// passes, forced moves, mana steps") did not include it. Neither is
// `opening-choice`: who takes the first turn is a real choice (CR 103.1),
// so the winner's window reaches the model.
var DefaultAbsorb = []string{rules.RuleForced, rules.RuleManaOnly, rules.RuleCoinCall, rules.RuleOpeningRoll}

// absorbable is every rule --absorb may name. A rule can be removed so
// more windows reach the model; nothing outside Layer A can be added,
// and there is no flag that answers anything Layer A would not.
var absorbable = map[string]bool{
	rules.RuleForced:        true,
	rules.RuleManaOnly:      true,
	rules.RuleCoinCall:      true,
	rules.RuleSameLand:      true,
	rules.RuleOpeningRoll:   true,
	rules.RuleOpeningChoice: true,
}

// ParseAbsorb reads --absorb: a comma-separated subset of forced,
// mana-only, coin-call, same-land, opening-roll and opening-choice.
// "none" absorbs nothing, so every window reaches the model.
func ParseAbsorb(s string) (map[string]bool, error) {
	out := map[string]bool{}
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return out, nil
	}
	for _, r := range strings.Split(s, ",") {
		r = strings.TrimSpace(r)
		if !absorbable[r] {
			names := make([]string, 0, len(absorbable))
			for n := range absorbable {
				names = append(names, n)
			}
			sort.Strings(names)
			return nil, fmt.Errorf("--absorb: %q is not a Layer A rule (want a subset of %s, or none)", r, strings.Join(names, ","))
		}
		out[r] = true
	}
	return out, nil
}

// autoAnswer is what the seat sends without asking the model.
type autoAnswer struct {
	index int
	rule  string // a rules.Rule* name, or rulePassUntil
}

// rulePassUntil counts the passes pass_until made (§4).
const rulePassUntil = "pass_until"

// decideAutomatically runs Layer A on a window, then the model's own
// pass_until, and says whether the seat answers it without the model.
// ok is false when the window is a real decision.
//
// The rules are rules.Resolve's, imported rather than written again, so
// the agent's trivial windows are answered exactly as the bot's are
// (decision 4). Two rules come from outside Resolve:
//
//   - while the CR 732 loop notice is up nothing is passed automatically
//     (ADR 0055); the notice asks a person to decide, and here the agent
//     is the person;
//   - a rule the owner left out of --absorb escalates.
func decideAutomatically(v *protocol.GameView, me uuid.UUID, moves []legal.Move, absorb map[string]bool, mode passMode) (autoAnswer, bool) {
	if len(moves) == 0 {
		return autoAnswer{}, false
	}
	verdict := rules.Resolve(aiseat.Input{View: *v, Seat: me, Moves: moves})
	if verdict.Absorbed() && absorb[verdict.Rule] && verdict.Index >= 0 && verdict.Index < len(moves) {
		if !(v.LoopNotice != nil && moves[verdict.Index].Kind == legal.KindPass) {
			return autoAnswer{index: verdict.Index, rule: verdict.Rule}, true
		}
	}
	switch mode {
	case passOthers:
		if i, ok := passUntilApplies(v, me.String(), moves); ok {
			return autoAnswer{index: i, rule: rulePassUntil}, true
		}
	case passSmart:
		if i, ok := smartPassApplies(v, me.String(), moves); ok {
			return autoAnswer{index: i, rule: rulePassUntil}, true
		}
	}
	return autoAnswer{}, false
}

// passMode is what the model last asked wait_for_decision's pass_until
// to do.
type passMode uint8

const (
	// passNone passes nothing beyond Layer A.
	passNone passMode = iota
	// passOthers is "my_turn_or_stack" (§4).
	passOthers
	// passSmart is "smart" (#3603): the browser's smart auto-pass.
	passSmart
)

// Wire values of pass_until.
const (
	passUntilNone   = "none"
	passUntilOthers = "my_turn_or_stack"
	passUntilSmart  = "smart"
)

// parsePassUntil reads wait_for_decision's pass_until.
func parsePassUntil(s string) (passMode, bool) {
	switch s {
	case "", passUntilNone:
		return passNone, true
	case passUntilOthers:
		return passOthers, true
	case passUntilSmart:
		return passSmart, true
	}
	return passNone, false
}

// passUntilApplies is pass_until "my_turn_or_stack" (§4): pass a priority
// window on another player's turn while the stack is empty, nothing is
// owed and nothing is being declared. It stops at once for anything on
// the stack, a pending choice, a combat declaration or the loop notice.
// The start of the seat's own turn clears the setting (seat.go).
func passUntilApplies(v *protocol.GameView, me string, moves []legal.Move) (int, bool) {
	if len(v.StackItems) > 0 {
		return 0, false
	}
	self, ok := passGuards(v, me, moves)
	if !ok || v.Turn.ActiveSeat == self.Seat {
		return 0, false
	}
	pass := aiseat.PassIndex(moves)
	return pass, pass >= 0
}

// passGuards is what stops every pass_until mode: the loop notice,
// triggers still being put on the stack, a choice the seat owes, a combat
// declaration, a mulligan, or a seat the view does not list. On success it
// returns the seat's own row.
func passGuards(v *protocol.GameView, me string, moves []legal.Move) (*protocol.PlayerView, bool) {
	if v.LoopNotice != nil || len(v.PendingTriggers) > 0 {
		return nil, false
	}
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Chooser == me {
			return nil, false
		}
	}
	self := mySeat(v, me)
	if self == nil {
		return nil, false
	}
	for _, m := range moves {
		switch m.Kind {
		case legal.KindAttack, legal.KindBlock, legal.KindFinishBlocks, legal.KindChoice, legal.KindMulligan:
			return nil, false
		}
	}
	return self, true
}
