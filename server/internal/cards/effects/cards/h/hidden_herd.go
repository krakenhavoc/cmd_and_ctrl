package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Herd — Enchantment {G}:
//
//	"When an opponent plays a nonbasic land, if this permanent is an
//	 enchantment, it becomes a 3/3 Beast creature."
//
// "Plays" is the land drop (CR 305.1): a nonbasic land put onto the battlefield by
// an effect does not count. An intervening "if" (CR 603.4).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e7c5649-41c2-4a0c-b873-e6f6c5908567",
		Name:         "Hidden Herd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return source.HasCardType("enchantment") && effects.OldB720AnOpponentPlayedALand(effects.NonbasicLand())(ev, source, ch, g)
			}, "Hidden Herd — it becomes a 3/3 Beast creature",
				effects.OldB717BecomesACreature([]string{"Beast"}, 3, 3, "Hidden Herd — a 3/3 Beast creature")),
		},
	})
}
