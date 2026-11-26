package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx-Bone Wand — Artifact {7} (Duel Decks: Izzet vs. Golgari):
//
//	"Whenever you cast an instant or sorcery spell, you may put a charge
//	 counter on this artifact. If you do, this artifact deals damage equal to
//	 the number of charge counters on it to any target."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d); the "may"
// is asked as it resolves, and both halves are "if you do": declining puts no
// counter and deals no damage. The damage counts the counters as the ability
// resolves, the new one included, and is dealt by the Wand to the target if it
// is still legal (CR 608.2b).
//
// No simplification.
func init() {
	const charge = "charge"
	effects.Register(effects.Spec{
		OracleID:     "4b9e529c-4fbd-4944-828d-7fbc8f26c991",
		Name:         "Sphinx-Bone Wand",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
					"Sphinx-Bone Wand — put a charge counter on it; it deals damage equal to the number of charge counters on it to any target",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if !effects.SourceIsStillThisPermanent(g, item) {
							return nil
						}
						if err := effects.PutACounterOnThis(charge)(g, item); err != nil {
							return err
						}
						n := effects.CountersOnThisAsItResolves(charge)(ctx)
						if n <= 0 {
							return nil
						}
						for _, t := range ctx.LegalTargets() {
							return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetAny()),
				"Sphinx-Bone Wand — put a charge counter on it?"),
		},
	})
}
