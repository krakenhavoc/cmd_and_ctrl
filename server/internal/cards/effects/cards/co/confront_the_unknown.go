package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Confront the Unknown — Instant {G} (Shadows over Innistrad, #5543):
//
//	"Investigate, then target creature gets +1/+1 until end of turn for
//	 each Clue you control."
//
// The Clue is made first, so it counts itself. The number of Clues is read
// as the spell resolves (CR 608.2h), after the investigate, and the boost
// goes to the target if it is still legal (CR 608.2b); a target that is
// illegal makes the whole spell do nothing, the Clue included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "221240c5-3c97-438c-b906-2508eacf190f",
		Name:         "Confront the Unknown",
		Completeness: effects.CompletenessFull,
		Purpose:      confrontPurpose(),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Investigate{N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype("Clue") })
				return effects.BoostUntilEOT{Target: t, Power: n, Toughness: n, Label: "Confront the Unknown — +1/+1 for each Clue"}.Apply(ctx)
			}),
		),
	})
}

// confrontPurpose declares the benefit on the one target (a pump that grows
// with the Clues, so no fixed size) and the Clue the spell makes.
func confrontPurpose() game.Purpose {
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))
	p.Tokens = 1
	return p
}
