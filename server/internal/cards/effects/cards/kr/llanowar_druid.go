package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Llanowar Druid — Creature — Elf Druid {1}{G}, 1/2 (Weatherlight):
//
//	"{T}, Sacrifice this creature: Untap all Forests."
//
// Every Forest, not only the controller's: each tapped Forest on the
// battlefield untaps, whoever controls it. The sacrifice is part of the cost,
// so the ability resolves even though the Druid is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8cadeee-36b7-4a61-91ff-7d6c7c780464",
		Name:         "Llanowar Druid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Untap all Forests.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Tapped && c.HasSubtype(ctx.TextSubtype("Forest")) {
						if err := (effects.UntapTarget{Target: c.InstanceID}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
				}
				return nil
			},
		}},
	})
}
