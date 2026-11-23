package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spinebiter — Creature — Phyrexian Beast {4}{G}{G}, 3/4 (New Phyrexia, #5575):
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
//	 You may have this creature assign its combat damage as though it weren't blocked."
//
// Infect (CR 702.90) and Pride of Lions' static (CR 510.1c): when it is
// blocked, its controller may assign all its combat damage to the player
// or planeswalker it attacks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9b844a0c-590f-46c7-b53e-c99398e0d8c0",
		Name:            "Spinebiter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "You may have this creature assign its combat damage as though it weren't blocked.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.AssignsAsThoughUnblocked = true
			},
		}},
	})
}
