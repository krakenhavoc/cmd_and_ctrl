package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abyssal Specter — Creature — Specter {2}{B}{B}, 2/3 (Ice Age, #5636):
//
//	"Flying
//	 Whenever this creature deals damage to a player, that player
//	 discards a card."
//
// Entropic Specter's trigger: any damage, combat or not, to any player,
// its controller included (ThisDealtDamageToAPlayer). "That player" is
// the damage event's target, and they choose the card (CR 701.9b), so
// the body is the shared discardOne; a player with an empty hand
// discards nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51833cff-6519-4806-8d91-0040e9a02189",
		Name:            "Abyssal Specter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtDamageToAPlayer,
				"Abyssal Specter — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return discardOne(g, item, NewContext(g, item).Trigger().Event.Target)
				}),
		},
	})
}
