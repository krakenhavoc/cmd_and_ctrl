package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Psychic Membrane — Creature — Wall {2}{U}, 0/3 (Mirrodin, S82 OLD,
// #5606):
//
//	"Defender
//	 Whenever this creature blocks, you may draw a card."
//
// "Whenever this creature blocks" triggers once however many attackers it
// blocks (CR 509.3a, SelfBlocksOnce: High Ground's shape), and the "you
// may" is the CR 603.5 optional prompt, asked as the trigger resolves. A
// player a draw limit stops is not asked (CR 121.2b).
//
// No simplification.
func init() {
	const label = "Psychic Membrane — draw a card"
	effects.Register(effects.Spec{
		OracleID:        "5a2deb96-01f4-48d5-bec8-44e4f141f551",
		Name:            "Psychic Membrane",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, label, effects.Do(effects.DrawCards{N: 1})), "Psychic Membrane — draw a card?", 1),
		},
	})
}
