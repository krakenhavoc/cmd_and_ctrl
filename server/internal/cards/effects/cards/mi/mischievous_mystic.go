package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mischievous Mystic — Creature — Human Wizard {1}{U}, 2/1 (Foundations,
// #4465):
//
//	"Flying
//	 Whenever you draw your second card each turn, create a 1/1 blue
//	 Faerie creature token with flying."
//
// Construct a Cosmic Cube's draw trigger (b41DrewYourSecondCardThisTurn):
// it fires on the draw that makes your count for the turn two, on any
// player's turn, and on no other draw. Flying rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "542831c6-6279-4977-9b29-0fb9c0ebf04d",
		Name:            "Mischievous Mystic",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Mischievous Mystic — create a 1/1 blue Faerie with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Faerie with flying"), N: 1})),
		},
	})
}
