package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mask of the Mimic — Instant {U} (Stronghold, S82 OLD b751, #5627):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Search your library for a card with the same name as target nontoken
//	 creature, put that card onto the battlefield, then shuffle."
//
// Bifurcate's search with Scapegoat's cost: the sacrifice is paid after the
// target is chosen (CR 601.2c, 601.2h), so sacrificing the target makes it
// illegal and the spell does nothing (CR 608.2b); a creature that was the target
// and is still there has its name read as the spell resolves. A face-down
// creature has no name (CR 708.2), and nothing matches it. The search may find
// nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "f1a16f6d-ea11-485e-8c5b-7f919a731fe0",
		Name:           "Mask of the Mimic",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Tutors: 1},
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetCreature("target nontoken creature", effects.Not(effects.IsTokenPredicate())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || c.FaceDown || c.Name == "" {
					return nil
				}
				return effects.OldB751SearchForTheNameOf(ctx, "Mask of the Mimic", c.Name)
			}
			return nil
		},
	})
}
