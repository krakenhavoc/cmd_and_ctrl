package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cinderheart Giant — Creature — Giant Berserker {5}{R}{R}, 7/6 (Kaldheim):
//
//	"Trample
//	 When this creature dies, it deals 7 damage to a creature an opponent
//	 controls chosen at random."
//
// Not a target: the creature is drawn at random as the trigger resolves,
// through the keyed random stream (RandomPick), from the creatures the
// controller's opponents control then. The damage is the Giant's, from its
// last-known self (CR 608.2h); with no such creature nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13efba9f-aedd-4ec0-a387-60c09b887341",
		Name:            "Cinderheart Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Cinderheart Giant — 7 damage to a creature an opponent controls chosen at random",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.Controller != item.Controller {
							ids = append(ids, c.InstanceID)
						}
					}
					pick := effects.RandomPick(ctx, ids, 1)
					if len(pick) == 0 {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: pick[0], Amount: 7}.Apply(ctx)
				}),
		},
	})
}
