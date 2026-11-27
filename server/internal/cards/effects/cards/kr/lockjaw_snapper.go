package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lockjaw Snapper — Artifact Creature — Scarecrow {4}, 2/2 (Shadowmoor,
// #5591):
//
//	"Wither
//	 When this creature dies, put a -1/-1 counter on each creature with a
//	 -1/-1 counter on it."
//
// The creatures are chosen as the trigger resolves (the Snapper is already
// in the graveyard and is not one of them), and the counters go on one at a
// time so a replacement ordering prompt can pause the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fadb0d89-0345-4edd-9546-dd4664e751f7",
		Name:            "Lockjaw Snapper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"wither"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Lockjaw Snapper — put a -1/-1 counter on each creature with a -1/-1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && effects.HasCounterOfKind(c, game.CounterMinusOne) {
							ids = append(ids, c.InstanceID)
						}
					}
					return effects.OldB541MinusCounterOnEach(effects.NewContext(g, item), ids)
				}),
		},
	})
}
