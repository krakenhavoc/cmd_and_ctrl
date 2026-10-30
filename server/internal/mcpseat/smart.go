package mcpseat

import (
	"regexp"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Step wire values the smart pass reads (game.Step; this package may not
// import internal/game, imports_test.go). stepDeclareAttackers is in
// attackall.go.
const (
	stepPrecombatMain  = "precombat_main"
	stepPostcombatMain = "postcombat_main"
	stepBeginCombat    = "begin_combat"
	stepDeclareBlock   = "declare_blockers"
	stepEnd            = "end"
)

// smartPassApplies is pass_until "smart" (#3603): the browser's smart
// auto-pass at its default settings, decided from the same flags the
// server puts on each move. client/src/lib/autopassDecision.ts owns the
// original; this is its port to the move list the agent sees, and
// responseWindow.ts's classifyMove is what counts as a response below.
//
// It passes a priority window when nothing is owed (passGuards: no choice,
// declaration, mulligan, loop notice or trigger still queuing) and
//
//   - with something on the stack, every item on it is the seat's own
//     (autoPassOwnStack, on by default), or the seat holds no response to
//     it; or
//   - with the stack empty, it is not a window worth a stop. Those are the
//     seat's own main phases (always: that is where its turn is played,
//     and where the browser's engineMayMissMana case lives), a
//     declare-attackers or declare-blockers step with an attack in
//     progress while the seat holds a response, and an opponent's end step
//     while it holds one.
//
// So every other step of the seat's own turn (upkeep, draw, beginning of
// combat, combat damage, end of combat, end) and of an opponent's turn
// passes even when the seat holds an instant-speed move. Unlike
// "my_turn_or_stack" it is not cleared by the start of the seat's own
// turn, and it does not pass for an opponent's stack item the seat could
// answer.
//
// The seat's own main phases stop with an instant in hand, as the
// browser's do: there the instant is a cast in the seat's sorcery window,
// which classifyMove counts as a play on a ticked step (the default grid
// ticks both main phases).
func smartPassApplies(v *protocol.GameView, me string, moves []legal.Move) (int, bool) {
	self, ok := passGuards(v, me, moves)
	if !ok {
		return 0, false
	}
	pass := aiseat.PassIndex(moves)
	if pass < 0 {
		return 0, false
	}
	mine := v.Turn.ActiveSeat == self.Seat
	step := v.Turn.Step
	respond := func() bool { return hasResponseMove(v, me, moves, mine) }

	if len(v.StackItems) > 0 {
		// #4186: a stack that is entirely the seat's own passes, as the
		// browser's autoPassOwnStack (on by default, #323) does: casting
		// in this window already answered the question a stop asks. The
		// server gives the seat priority again after its own item
		// resolves, and an opponent's answer to it stops as usual.
		if ownsEveryStackItem(v, me) {
			return pass, true
		}
		if respond() {
			return 0, false
		}
		return pass, true
	}
	if mine && (step == stepPrecombatMain || step == stepPostcombatMain) {
		return 0, false
	}
	combat := (step == stepDeclareAttackers || step == stepDeclareBlock) && anyAttacking(v)
	oppEnd := step == stepEnd && !mine
	if (combat || oppEnd) && respond() {
		return 0, false
	}
	return pass, true
}

// anyAttacking reports whether any creature is declared as an attacker.
func anyAttacking(v *protocol.GameView) bool {
	for i := range v.Battlefield.Cards {
		if v.Battlefield.Cards[i].AttackingTarget != "" {
			return true
		}
	}
	return false
}

// hasResponseMove is responseWindow.ts's hasResponse at the default
// "Stop for" categories (DEFAULT_RESPONSES): a counter, an instant, a
// targeted or interacting ability and a special action count; a pure
// value ability (a fetch, Mind Stone, cycling), a mana ability, a land and
// a declaration do not, and nor does a value instant (Opt, Brainstorm:
// a cast marked value_only, #2877) while something is on the stack.
func hasResponseMove(v *protocol.GameView, me string, moves []legal.Move, mine bool) bool {
	sorcery := mine && len(v.StackItems) == 0 && (v.Turn.Step == stepPrecombatMain || v.Turn.Step == stepPostcombatMain)
	onStack := len(v.StackItems) > 0
	combat := inCombatWindow(v)
	defending := isDefending(v, me)
	controllers := map[string]string{}
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		if c.Controller != "" {
			controllers[c.InstanceID] = c.Controller
		}
	}
	for _, m := range moves {
		if m.Kind == legal.KindActivate && m.Source != uuid.Nil {
			if ctl, ok := controllers[m.Source.String()]; ok && ctl != me {
				// ADR 0106 §1: "any player may activate" a permanent
				// somebody else controls is never a reason to stop.
				continue
			}
		}
		if moveIsResponse(m, sorcery, onStack, combat, defending) {
			return true
		}
	}
	return false
}

// moveIsResponse is classifyMove's response classes: counter, instant,
// ability and special. The "untargeted" class is off by default, and so
// is "valueInstant", a value_only cast while the stack holds something
// (#2877; the browser's respondValueInstants, off by default). The
// caller has already passed a stack that is entirely the seat's own, so
// onStack means an opponent's item is there.
func moveIsResponse(m legal.Move, sorcery, onStack, combat, defending bool) bool {
	switch m.Kind {
	case legal.KindMana:
		// A sacrifice outlet that makes mana still answers removal.
		return m.Interacts && !sorcery
	case legal.KindCast, legal.KindActivate:
		if sorcery {
			return false
		}
		if m.TargetsStack {
			return true
		}
		if m.Kind == legal.KindCast {
			return !(m.ValueOnly && onStack)
		}
		if m.HasTargets || m.Interacts {
			return true
		}
		return combat && m.CombatInteracts && (!m.CombatDefenderOnly || defending)
	case legal.KindSpecialAction:
		return true
	}
	return false
}

// attackOrBlockLabel names a stack item about an attack or a block.
var attackOrBlockLabel = regexp.MustCompile(`(?i)\b(attack|attacks|attacking|attacked|block|blocks|blocking|blocked)\b`)

// inCombatWindow is responseWindow.ts's inCombatWindow: beginning of
// combat, declare attackers, declare blockers, or an attack or block
// trigger on the stack. A combat ability is a response only there.
func inCombatWindow(v *protocol.GameView) bool {
	switch v.Turn.Step {
	case stepBeginCombat, stepDeclareAttackers, stepDeclareBlock:
		return true
	}
	for _, it := range v.StackItems {
		if it.Kind != "spell" && attackOrBlockLabel.MatchString(it.Label) {
			return true
		}
	}
	return false
}

// isDefending is responseWindow.ts's isDefending: a creature is attacking
// the seat, or a planeswalker or battle it defends.
func isDefending(v *protocol.GameView, me string) bool {
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		if c.AttackingTarget == "" {
			continue
		}
		if c.DefendingPlayer == me {
			return true
		}
		if c.DefendingPlayer == "" && c.AttackingTarget == me && (c.AttackingTargetKind == "" || c.AttackingTargetKind == "player") {
			return true
		}
	}
	return false
}

// ownsEveryStackItem is holdPriority.ts's ownsEveryStackItem: no trigger
// still queuing, something on the stack, and every item and card on it
// controlled by the seat.
func ownsEveryStackItem(v *protocol.GameView, me string) bool {
	if len(v.PendingTriggers) > 0 {
		return false
	}
	if len(v.StackItems) == 0 && len(v.Stack.Cards) == 0 {
		return false
	}
	for _, it := range v.StackItems {
		if it.Controller != me {
			return false
		}
	}
	for i := range v.Stack.Cards {
		if v.Stack.Cards[i].Controller != me {
			return false
		}
	}
	return true
}
