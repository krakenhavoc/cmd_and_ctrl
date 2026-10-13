package game

import (
	"fmt"
	"reflect"
	"strings"
)

// ability_row_labels.go — #2410: the wording of the CardDef slots whose
// entries are structured data rather than a sentence — a hand-size
// static, a block rule, an attack limit, a game-end gate. Each reads the
// entry's own fields, so every card that declares one gains wording at
// once. A sentence a player reads, in the Caveats tone
// (docs/adding-cards.md): no engine vocabulary.

func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// handSizeRows words CardDef.HandSize.
func handSizeRows(v reflect.Value) []string {
	statics, _ := v.Interface().([]HandSizeStatic)
	out := make([]string, 0, len(statics))
	for _, s := range statics {
		out = append(out, s.rowLabel())
	}
	return out
}

func (s HandSizeStatic) rowLabel() string {
	var subject, has, possessive string
	switch s.Players {
	case HandSizeEachOpponent:
		subject, has, possessive = "Each opponent", "has", "Each opponent's"
	case HandSizeEachPlayer:
		subject, has, possessive = "Players", "have", "Players'"
	case HandSizeChosenPlayer:
		subject, has, possessive = "The chosen player", "has", "The chosen player's"
	default:
		subject, has, possessive = "You", "have", "Your"
	}
	if s.Dynamic != nil && s.Kind != HandSizeNoMaximum {
		return possessive + " maximum hand size depends on the board."
	}
	switch s.Kind {
	case HandSizeNoMaximum:
		return fmt.Sprintf("%s %s no maximum hand size.", subject, has)
	case HandSizeSet:
		return fmt.Sprintf("%s maximum hand size is %d.", possessive, s.N)
	case HandSizeModify:
		if s.N < 0 {
			return fmt.Sprintf("%s maximum hand size is reduced by %d.", possessive, -s.N)
		}
		return fmt.Sprintf("%s maximum hand size is increased by %d.", possessive, s.N)
	}
	return "Changes maximum hand size."
}

// attackLimitRows words CardDef.AttackLimits.
func attackLimitRows(v reflect.Value) []string {
	limits, _ := v.Interface().([]AttackLimit)
	out := make([]string, 0, len(limits))
	for _, l := range limits {
		creatures := "creatures"
		if l.Max == 1 {
			creatures = "creature"
		}
		var what string
		switch l.Scope {
		case AttackLimitEachCombat:
			what = fmt.Sprintf("No more than %d %s can attack each combat.", l.Max, creatures)
		case AttackLimitAttackingThis:
			what = fmt.Sprintf("No more than %d %s can attack it each combat.", l.Max, creatures)
		default:
			what = fmt.Sprintf("No more than %d %s can attack you each combat.", l.Max, creatures)
		}
		if l.While != nil {
			what = strings.TrimSuffix(what, ".") + " while a condition holds."
		}
		out = append(out, what)
	}
	return out
}

// gameEndGateRows words CardDef.GameEndGates.
func gameEndGateRows(v reflect.Value) []string {
	gates, _ := v.Interface().([]GameEndGate)
	out := make([]string, 0, len(gates))
	for _, g := range gates {
		var who string
		switch g.Scope {
		case GateYou:
			who = "You"
		case GateOpponents:
			who = "Your opponents"
		case GateEachPlayer:
			who = "Players"
		}
		var what string
		switch {
		case who == "" || len(g.Causes) > 0:
			out = append(out, "Changes who can win or lose.")
			continue
		case g.CantLose && g.CantWin:
			what = "win or lose"
		case g.CantLose:
			what = "lose"
		case g.CantWin:
			what = "win"
		default:
			out = append(out, "Changes who can win or lose.")
			continue
		}
		out = append(out, fmt.Sprintf("%s can't %s the game.", who, what))
	}
	return out
}

// blockRuleRows words CardDef.BlockRules. The Label of a rule is the
// card's own parameter ("Walls", "creatures with toughness 3 or
// greater"), so the sentence is built around it. The rule's scope (which
// creature it protects) is a closure the row cannot read, so the wording
// leaves the subject to the card, as an Aura's chip does.
func blockRuleRows(v reflect.Value) []string {
	rules, _ := v.Interface().([]BlockRule)
	out := make([]string, 0, len(rules))
	for _, r := range rules {
		l := strings.TrimSpace(r.Label)
		switch {
		case r.Count != nil || r.Limit != nil:
			if l != "" {
				out = append(out, capital(l)+".")
			} else {
				out = append(out, "Limits how many creatures can block.")
			}
		case l == "":
			out = append(out, "Blocking restriction")
		case r.Reason == BlockReasonCantBeBlockedExceptBy:
			out = append(out, "Can't be blocked except by "+l+".")
		case r.Reason == BlockReasonCantBlockAttacker:
			out = append(out, capital(strings.TrimSuffix(l, "."))+".")
		default:
			out = append(out, "Can't be blocked by "+l+".")
		}
	}
	return out
}

// entersWithCountersRows words CardDef.EntersWithCountersFromCast.
func entersWithCountersRows(v reflect.Value) []string {
	entries, _ := v.Interface().([]EntryCountersFromCast)
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, "Enters with "+e.Kind+" counters, based on how it was cast.")
	}
	return out
}
