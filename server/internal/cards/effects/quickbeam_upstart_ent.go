package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	trigger := On(game.EventETB, gapB125ThisOrAnotherTreefolkEntered, quickbeamLabel,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := gapB125PumpTrample(ctx, t.ID, 2, 2, "Quickbeam, Upstart Ent"); err != nil {
					return err
				}
			}
			return nil
		})
	trigger.Targets = TargetCreature("up to two target creatures").WithCount(0, 2)
	Register(Spec{
		OracleID:     "9ad903c5-1528-44d6-b4b6-f05a2399d91f",
		Name:         "Quickbeam, Upstart Ent",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
