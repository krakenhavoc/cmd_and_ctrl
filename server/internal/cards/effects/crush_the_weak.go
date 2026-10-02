package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crush the Weak — Sorcery {2}{R}:
//
//	"Crush the Weak deals 2 damage to each creature. If a creature dealt
//	 damage this way would die this turn, exile it instead.
//	 Foretell {R} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell
//	 cost.)"
//
// Anger of the Gods for 2: each creature dealt damage is marked from
// the damage's continuation, so one behind a shield that prevented all
// of it is not (ADR 0108 §1 decision 3). Foretell is the keyword's
// (CR 702.143a).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "4aa119f7-d411-4188-956e-547f7d14e789",
		Name:           "Crush the Weak",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Foretell("{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachExileIfDealtDies(ctx, Creature(), 2, false)
		},
	})
}
