package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abomination of Gudul — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Horror {3}{B}{G}{U}, 3/4: Flying. Whenever this creature deals combat
//	damage to a player, you may draw a card. If you do, discard a card.
//	Morph {2}{B}{G}{U}
//
// The loot is the may-draw-then-discard of Marauding Looter.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "3d98af5f-7a0b-4a5a-b3e4-f3c9d150c993",
		Name:             "Abomination of Gudul",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}{B}{G}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B36SelfDealtCombatDamageToPlayer(ev, source, g)
			}, "Abomination of Gudul — you may draw a card; if you do, discard a card",
				effects.OldB223MayLoot("Abomination of Gudul — draw a card, then discard a card?")),
		},
	})
}
