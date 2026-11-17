package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venerated Rotpriest — Creature — Phyrexian Druid {G}, 1/2:
//
//	"Toxic 1
//	 Whenever a creature you control becomes the target of a spell,
//	 target opponent gets a poison counter."
//
// The trigger is Gargos's condition (b40CreatureYouControlTargetedByASpell):
// "of a spell", not "of a spell or ability", read off the stack, and the
// Rotpriest counts as a creature its controller controls. It fires per
// target instance (CR 115.3), so a spell that targets two of your
// creatures gives two counters. The counter is placed by the trigger's
// controller on the opponent still targeted at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eafd7ba5-691d-4f8d-94bb-70a5612245a1",
		Name:            "Venerated Rotpriest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b40CreatureYouControlTargetedByASpell(ev, source, g)
				}, "Venerated Rotpriest — target opponent gets a poison counter",
					func(g *game.Game, item *game.StackItem) error {
						for _, t := range NewContext(g, item).LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return g.AddPlayerCounterByForEffect(item.Controller, t.ID, game.CounterPoison, 1)
							}
						}
						return nil
					}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
