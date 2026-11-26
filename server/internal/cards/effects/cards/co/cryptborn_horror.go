package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryptborn Horror — Creature — Horror {1}{B/R}{B/R}, 0/0 (Return to Ravnica,
// #5566):
//
//	"Trample
//	 This creature enters with X +1/+1 counters on it, where X is the total life
//	 lost by your opponents this turn."
//
// A CR 614.1c "enters with" clause counted as the creature enters, from the life
// each opponent has lost this turn (the turn tally, as Cindering Cutthroat and
// Rakdos, Lord of Riots read it). With none lost it is a 0/0 and dies (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f5d62baa-648d-4030-87f9-e795ed665084",
		Name:            "Cryptborn Horror",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return effects.OpponentLifeLostThisTurn(g, src.Controller)
			}, "Cryptborn Horror enters with X +1/+1 counters, X the total life your opponents lost this turn"),
		},
	})
}
