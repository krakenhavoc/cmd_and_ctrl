package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	t := WhenThisEnters("Ravenous Sailback — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("This creature gains haste until end of turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if !sourceIsStillThisPermanent(ctx.Game, item) {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"haste"},
					Label:    "Ravenous Sailback — haste until end of turn",
				}.Apply(ctx)
			}),
		ModeDoing("Destroy target artifact or enchantment.",
			TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			DestroyTheModesTarget),
	)
	Register(Spec{
		OracleID:     "17e74c4b-435c-469e-a9a4-32bf0cb7634c",
		Name:         "Ravenous Sailback",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
