package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const quickbeamLabel = "Quickbeam, Upstart Ent — up to two target creatures each get +2/+2 and gain trample until end of turn"

// Quickbeam, Upstart Ent — Legendary Creature — Treefolk {4}{G}{G}, 5/6
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever Quickbeam or another Treefolk you control enters, up to two
//	 target creatures each get +2/+2 and gain trample until end of turn."
//
// Karfell Kennel-Master's up-to-two target pump on a "this or another
// Treefolk" enters watch. Each target still legal as the trigger resolves
// gets both; none chosen is a legal answer.
//
// No simplification.
func init() {
	trigger := effects.On(game.EventETB, effects.GapB125ThisOrAnotherTreefolkEntered, quickbeamLabel,
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := effects.GapB125PumpTrample(ctx, t.ID, 2, 2, "Quickbeam, Upstart Ent"); err != nil {
					return err
				}
			}
			return nil
		})
	trigger.Targets = effects.TargetCreature("up to two target creatures").WithCount(0, 2)
	effects.Register(effects.Spec{
		OracleID:     "9ad903c5-1528-44d6-b4b6-f05a2399d91f",
		Name:         "Quickbeam, Upstart Ent",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
