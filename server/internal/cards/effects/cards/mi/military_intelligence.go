package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Military Intelligence — Enchantment {1}{U} (Magic 2015, #5556):
//
//	"Whenever you attack with two or more creatures, draw a card."
//
// Chivalric Alliance's trigger on its own: one draw per declaration of
// two or more attackers (OncePerBatch), never one per attacker.
//
// No simplification.
func init() {
	const label = "Military Intelligence — you attacked with two or more creatures: draw a card"
	effects.Register(effects.Spec{
		OracleID:     "464bfcdb-a82b-4573-92a4-c5e463724fbe",
		Name:         "Military Intelligence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.B16PlayerAttackedWithAtLeast(ev, source, g, 2, label)
			}, label, effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
