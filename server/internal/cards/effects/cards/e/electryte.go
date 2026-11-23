package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Electryte — Creature — Trilobite Beast {3}{R}{R}, 3/3 (Urza's Saga):
//
//	"Whenever this creature deals combat damage to defending player, it
//	 deals damage equal to its power to each blocking creature."
//
// Cadira's "deals combat damage to a player" condition: combat damage
// to a player can only be dealt to the player it attacks. "Each
// blocking creature" is every creature blocking as the trigger
// resolves, not only those blocking Electryte. Its power is read then
// too, or, if it has left the battlefield, as it last existed
// (CR 608.2h), and it is still the source of the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ff0c25f-4593-460b-9d7f-d9c37f135a18",
		Name:         "Electryte",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B36SelfDealtCombatDamageToPlayer(ev, source, g)
			}, "Electryte — damage equal to its power to each blocking creature", func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.BlockingCreature(), effects.B43PowerNowOrLastKnown(g, item.SourceCardID))
			}),
		},
	})
}
