package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jolrael, Mwonvuli Recluse — Legendary Creature — Human Druid {1}{G},
// 1/2 (Murders at Karlov Manor Commander, #5368):
//
//	"Whenever you draw your second card each turn, create a 2/2 green
//	 Cat creature token.
//	 {4}{G}{G}: Until end of turn, creatures you control have base power
//	 and toughness X/X, where X is the number of cards in your hand."
//
// The trigger is Wolfbat's second-draw condition. The ability is Mirror
// Entity's layer 7b set over the creatures you control as it resolves
// (CR 611.2c), with X read then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8640899-2458-40f2-98a3-35c402ee9a7a",
		Name:         "Jolrael, Mwonvuli Recluse",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Jolrael — create a 2/2 green Cat creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Cat"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{G}: Until end of turn, creatures you control have base power and toughness X/X, where X is the number of cards in your hand.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{4}{G}{G}"),
			Effect:  effects.Cmdr2S22CreaturesBecomeXByXForCardsInHand,
		}},
	})
}
