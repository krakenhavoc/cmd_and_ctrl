package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guul Draz Specter — Creature — Specter {2}{B}{B}, 2/2 (Conspiracy: Take
// the Crown, S82 older sets batch 270, #5540):
//
//	"Flying
//	 This creature gets +3/+3 as long as an opponent has no cards in hand.
//	 Whenever this creature deals combat damage to a player, that player
//	 discards a card."
//
// The pump is a layer 7c modification of the source alone (CR 613.4c) that
// is on while any opponent still in the game has an empty hand, and it
// declares that it depends on hand sizes so a draw or a discard re-reads it
// at once (Demon's Jester's hellbent). The trigger is Abyssal Specter's:
// the player the Specter hit discards a card of their choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "733cfc16-9a16-41e2-b36e-3d7a546c67f0",
		Name:            "Guul Draz Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{guulDrazSpecterPump()},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Guul Draz Specter — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DiscardOne(g, item, effects.NewContext(g, item).Trigger().Event.Target)
				}),
		},
	})
}

func guulDrazSpecterPump() game.StaticAbility {
	return game.StaticAbility{
		Label:    "This creature gets +3/+3 as long as an opponent has no cards in hand.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if !effects.SelfCreature(target, g, source) || g == nil {
				return false
			}
			for _, p := range g.Seats {
				if p != nil && p.ID != source.Controller && !p.Eliminated && effects.CardsInHand(g, p.ID) == 0 {
					return true
				}
			}
			return false
		},
		Apply:             effects.PumpBy(3, 3),
		DependsOnHandSize: true,
	}
}
