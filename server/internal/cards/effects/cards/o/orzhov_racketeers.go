package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orzhov Racketeers — Creature — Human Rogue {4}{B}, 3/2 (Ravnica: Clue Edition, #5929):
//
//	"Whenever this creature deals combat damage to a player, that player
//	 discards a card.
//	 Afterlife 2 (When this creature dies, create two 1/1 white and black
//	 Spirit creature tokens with flying.)"
//
// The discard is Blazing Specter's (the damaged player picks the card, an empty hand
// discards nothing). Afterlife has no helper: it is the dies trigger it reminds
// (CR 702.135a), Tithe Taker's shape with N: 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f6c9cee-5af7-4b00-a2a4-2770dab0a2dc",
		Name:         "Orzhov Racketeers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Orzhov Racketeers — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DiscardOne(g, item, item.Trigger.Event.Target)
				}),
			effects.WhenThisDies("Orzhov Racketeers — afterlife 2: create two 1/1 white and black Spirit creature tokens with flying",
				func(g *game.Game, item *game.StackItem) error {
					// Looked up as it resolves: the Spirit row is added by a later file's init.
					return effects.CreateToken{Template: effects.TokenCard("1/1 white and black Spirit with flying"), N: 2}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
