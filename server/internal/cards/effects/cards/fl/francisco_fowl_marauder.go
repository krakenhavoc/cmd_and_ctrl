package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Francisco, Fowl Marauder — Legendary Creature — Bird Pirate {1}{B}, 0/1:
//
//	"Flying
//	 Francisco can't block.
//	 Whenever one or more Pirates you control deal damage to a player,
//	 Francisco explores.
//	 Partner (You can have two commanders if both have partner.)"
//
// Any damage counts, not only combat damage, and a player includes the
// Pirates' own controller; Francisco itself is a Pirate, so its own damage
// explores. One trigger per player damaged per batch (CR 603.2c). Partner
// is a deck-construction rule (CR 702.124) and needs nothing on the card.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1dbe4eb4-8313-459a-89ce-982442cd0cd0",
		Name:            "Francisco, Fowl Marauder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerPlayer(effects.On(game.EventDealDamage, effects.Cmdr2B05PirateDealtDamageToAPlayer,
				"Francisco — explores", func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Explores{Explorer: ctx.Source()}.Apply(ctx)
				})),
		},
	})
}
