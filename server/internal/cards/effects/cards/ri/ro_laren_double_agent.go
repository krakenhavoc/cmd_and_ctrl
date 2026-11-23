package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ro Laren, Double Agent — Legendary Creature — Bajoran Rogue Officer
// {2}{R}, 3/2 (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, Ro Laren
//	 deals 2 damage to each opponent.
//	 {1}, {T}, Discard a card: Draw a card."
//
// Seek out is an ability word. The trigger is Bard the Bowman's "your
// second card each turn" (b41DrewYourSecondCardThisTurn), any turn, the
// draw step's card counted. The loot is Charging Strifeknight's with a
// {1}: the discard is a cost (CR 602.2b), so the draw happens only when a
// card was really discarded, and that draw can be the second card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f10f6bf-f889-4c3d-a6cf-04cf74f1b4bd",
		Name:         "Ro Laren, Double Agent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Ro Laren, Double Agent — Seek out: 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Discard a card: Draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
