package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Sailback — Creature — Dinosaur {4}{G}, 3/4 (March of the Machine,
// S82 MOM batch 13):
//
//	"When this creature enters, choose one —
//	 • This creature gains haste until end of turn.
//	 • Destroy target artifact or enchantment."
//
// One modal trigger, Elven Raft-Steerer's shape: the mode is chosen as the
// trigger goes on the stack and only the destroy bullet takes a target
// (CR 700.2b). The haste bullet gives this creature haste until end of
// turn while it is still the same permanent (CR 400.7).
//
// No simplification.
func init() {
	t := effects.WhenThisEnters("Ravenous Sailback — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("This creature gains haste until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if !effects.SourceIsStillThisPermanent(ctx.Game, item) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"haste"},
					Label:    "Ravenous Sailback — haste until end of turn",
				}.Apply(ctx)
			}),
		effects.ModeDoing("Destroy target artifact or enchantment.",
			effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			effects.DestroyTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:     "17e74c4b-435c-469e-a9a4-32bf0cb7634c",
		Name:         "Ravenous Sailback",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
