package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sedraxis Specter — Creature — Specter {U}{B}{R}, 3/2:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player
//	 discards a card.
//	 Unearth {1}{B}"
//
// Blazing Specter's discard (the damaged player chooses the card, CR
// 701.9b) and Molten Gatekeeper's Unearth constructor: the return, the
// haste, the end-step exile and the leaves-the-battlefield redirect all
// come with the keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eccc1230-df5a-44a0-859b-9e1e0d176dff",
		Name:            "Sedraxis Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Sedraxis Specter — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DiscardOne(g, item, item.Trigger.Event.Target)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{1}{B}")},
	})
}
