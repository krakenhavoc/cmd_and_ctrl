// Package rules is ADR 0033 §5's Layer A: the free, sub-millisecond
// filter that resolves a decision window outright when the rules of
// the game — not judgement about the board — already settle it.
//
// # Why this is the load-bearing layer
//
// The three-layer funnel only works economically if most windows
// never reach a model. ADR 0033 puts a number on it: "if Layer A is
// not absorbing >80% of windows, the funnel is broken and should be
// fixed before reaching for a cheaper model." That is the failure
// mode this package exists to prevent — not bad play, but a bot that
// spends a frontier call deciding whether to pass priority during its
// own untap step.
//
// # Layer A inherits nothing
//
// S13.6 shipped a smart auto-pass, and it is easy to assume this is
// the server half of it. It is not: `priority.ts` and `settings.ts`
// are CLIENT code, a convenience that suppresses prompts in one
// browser. The server hands a bot every window a human would be
// offered, so Layer A does the entire job from scratch.
//
// # What counts as "the rules already settled it"
//
// Every rule here has to be defensible as a fact about the game
// rather than an opinion about the position. The bar is that a
// competent player would not consider the window a decision at all:
//
//   - one legal move — there is nothing to decide
//   - the only non-pass moves are bare mana activations — casts carry
//     auto_tap, so floating mana buys nothing and empties at the next
//     step boundary (CR 106.4)
//   - the only non-pass moves are land drops for copies of the same
//     card — the land is free, once a turn, and the copies are
//     interchangeable
//
// Anything that needs a view on what the board is worth — which
// removal spell, whether to block, whether to hold up the counter —
// is NOT settled by the rules and escalates. Layer A being small is
// the point: it is the layer that must never be wrong.
//
// # The type gate
//
// Like every policy package under aiseat/, this one may not import
// internal/game (ADR 0033 §3). It reads protocol.GameView and
// []legal.Move and nothing else. TestPolicyPackagesDoNotImportGame in
// aiseat/heuristic enforces it across the whole subtree.
package rules

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// Outcome is what Layer A did with a window.
type Outcome uint8

const (
	// Escalate means Layer A declined to answer: the window is a real
	// decision and belongs to Layer B or Layer C.
	Escalate Outcome = iota
	// Take means Layer A resolved the window; Verdict.Index is the
	// move to make.
	Take
)

// Rule names, stable enough to aggregate on in a log or a metric.
const (
	// RuleNone is the rule attributed to an escalation.
	RuleNone = "escalate"
	// RuleNoMoves is the degenerate window: nothing was offered. The
	// runner never produces one, but a policy is a public interface.
	RuleNoMoves = "no-moves"
	// RuleForced is a window with exactly one legal move.
	RuleForced = "forced"
	// RuleManaOnly is a window whose only alternative to passing is
	// floating mana.
	RuleManaOnly = "mana-only"
	// RuleSameLand is a window whose only alternative to passing is
	// playing one of several copies of the same land.
	RuleSameLand = "same-land"
)

// Verdict is Layer A's answer for one window.
type Verdict struct {
	Outcome Outcome
	// Index is the chosen move, valid only when Outcome is Take.
	Index int
	// Rule names the rule that fired, for instrumentation. Set on an
	// escalation too, where it is RuleNone.
	Rule string
	// Reason is the human-readable line that reaches the decision log
	// and, behind a setting, the table.
	Reason string
}

// Absorbed reports whether the verdict resolved the window.
func (v Verdict) Absorbed() bool { return v.Outcome == Take }

// Resolve runs the filter. It allocates nothing and is a handful of
// passes over the move list — call it on every window, including in
// the heuristic tier, where it costs nothing and saves the scorer a
// walk.
func Resolve(in aiseat.Input) Verdict {
	moves := in.Moves
	switch {
	case len(moves) == 0:
		return escalate(RuleNoMoves)

	case len(moves) == 1:
		// Includes the pass-only window, which is the single most
		// common thing that happens at a four-player table: three
		// seats in a row are offered priority with nothing to
		// respond with.
		//
		// One edge deserves naming rather than discovering. A window
		// holding exactly one BLOCK is not, strictly, forced — the
		// defender could decline and take the damage instead, and
		// aiseat.Decline exists so a policy can say so. This rule
		// takes the block anyway, because the heuristic's own
		// `len(moves) == 1` short-circuit does exactly the same thing
		// and Layer A's contract is that it never changes how the
		// tier plays (TestLayerAAgreesWithTheHeuristicOnEveryWindow-
		// ItAbsorbs holds that over thousands of windows). If that
		// judgement is ever wrong it is wrong in aiseat/heuristic,
		// and it should be fixed there, with this rule following.
		return Verdict{Outcome: Take, Index: 0, Rule: RuleForced, Reason: "only legal move: " + moves[0].Label}
	}

	pass := aiseat.PassIndex(moves)
	if pass < 0 {
		// No pass on offer means this seat does not hold priority —
		// a combat declaration window. Those are decisions.
		return escalate(RuleNone)
	}

	// Classify the alternatives to passing in one pass.
	var mana, lands, other int
	firstLand, landLabel := -1, ""
	sameLand := true
	for i := range moves {
		switch moves[i].Kind {
		case legal.KindPass:
			// The enumerator emits exactly one, but do not rely on it.
		case legal.KindMana:
			mana++
		case legal.KindLand:
			lands++
			if firstLand < 0 {
				firstLand, landLabel = i, moves[i].Label
			} else if moves[i].Label != landLabel {
				sameLand = false
			}
		default:
			other++
		}
	}
	if other > 0 {
		return escalate(RuleNone)
	}

	switch {
	case lands == 0 && mana > 0:
		// Casts auto-tap for their costs, so a bare mana activation
		// adds nothing the bot could not have had on demand, and the
		// pool empties at the next step boundary (CR 106.4). There is
		// no position in which this is the play.
		return Verdict{Outcome: Take, Index: pass, Rule: RuleManaOnly,
			Reason: "nothing on offer but floating mana"}

	case lands > 0 && sameLand:
		// The land drop is free and once a turn; copies of one land
		// are interchangeable, so which instance goes down is not a
		// decision. Two DIFFERENT lands is a colour decision and
		// escalates.
		return Verdict{Outcome: Take, Index: firstLand, Rule: RuleSameLand,
			Reason: "free land drop: " + landLabel}
	}
	return escalate(RuleNone)
}

func escalate(rule string) Verdict {
	return Verdict{Outcome: Escalate, Rule: rule}
}
