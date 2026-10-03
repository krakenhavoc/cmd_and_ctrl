package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Unlife — Enchantment {2}{W}:
//
//	"You don't lose the game for having 0 or less life.
//	 As long as you have 0 or less life, all damage is dealt to you as
//	 though its source had infect. (Damage is dealt to you in the form of
//	 poison counters.)"
//
// Two battlefield statics:
//
//   - A "can't lose" gate narrowed to one cause, CR 704.5a's 0 or less
//     life (ADR 0057, LossLife), Transcendence's shape. Its ruling: "You
//     can still lose the game for other reasons, including having ten or
//     more poison counters or drawing a card from a library with no cards
//     in it." If it leaves the battlefield while you are at 0 or less, the
//     next state-based action check takes you out, because the gate is
//     read at every check.
//   - ADR 0108 §10's "dealt to you as though its source had infect" (CR
//     120.3b, 702.90b, 609.4): damage dealt to its controller is poison
//     counters instead of life loss. The condition is read once per damage
//     instance, against the life total you had as it began (owner decision
//     4), which is its ruling: "Phyrexian Unlife won't affect damage that
//     reduces your life total from a positive number to 0 or less. … The
//     next time you're dealt damage, it will be dealt as though its source
//     had infect."
//
// Its third ruling, "If you're at 0 or less life, you can't pay any
// amount of life except 0", is CR 119.4, which CanPayLifeLocked already
// enforces.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "6598d988-d60e-4441-88b6-8d0995c4675a",
		Name:           "Phyrexian Unlife",
		Completeness:   CompletenessFull,
		GameEndGates:   []game.GameEndGate{{Scope: game.GateYou, CantLose: true, Causes: []game.LossCause{game.LossLife}}},
		DamageAsThough: DamageToYouAsThoughInfectAtOrBelowZeroLife(),
	})
}
