package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cindering Cutthroat — Creature — Lizard Assassin {2}{B/R}, 3/2
// (Bloomburrow, #4467):
//
//	"This creature enters with a +1/+1 counter on it if an opponent lost
//	 life this turn.
//	 {1}{B/R}: This creature gains menace until end of turn. (It can't be
//	 blocked except by two or more creatures.)"
//
// The counter is Blacksnag Buzzard's CR 614.1c self-replacement
// (b19EntersWithCountersCounted), with the count read as the Cutthroat
// enters: one if any opponent of its controller lost life this turn
// (opponentLifeLostThisTurn, Stromkirk Bloodthief's read). It asks
// whether an opponent lost life, not how their total moved, as the
// ruling says. The menace row is Endling's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46e8bec0-1a09-4e6e-b78a-7ba161686ca9",
		Name:         "Cindering Cutthroat",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, oneIfAnOpponentLostLifeThisTurn,
				"Cindering Cutthroat enters with a +1/+1 counter if an opponent lost life this turn"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B/R}: This creature gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{B/R}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("menace", "Cindering Cutthroat — menace until end of turn"),
		}},
	})
}

// oneIfAnOpponentLostLifeThisTurn is the count of the Lizards' "enters
// with a +1/+1 counter on it if an opponent lost life this turn"
// (Cindering Cutthroat, Frilled Sparkshooter), for
// b19EntersWithCountersCounted: one if any opponent of the entering
// permanent's controller lost life this turn, read as it enters.
func oneIfAnOpponentLostLifeThisTurn(g *game.Game, src *game.Card) int {
	if opponentLifeLostThisTurn(g, src.Controller) > 0 {
		return 1
	}
	return 0
}
