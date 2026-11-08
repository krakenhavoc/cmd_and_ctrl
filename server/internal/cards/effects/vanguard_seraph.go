package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanguard Seraph — Creature — Angel Warrior {3}{W}, 3/3 (Foundations,
// #4465):
//
//	"Flying
//	 Whenever you gain life for the first time each turn, surveil 1."
//
// Leech Collector's first-gain-of-the-turn reader (any turn, yours or an
// opponent's), read off the turn's log, so a gain before the Seraph
// arrived still uses up the turn's first. Flying rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3250466f-82a4-45e1-9722-ff247ddf8c15",
		Name:            "Vanguard Seraph",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return socB28FirstLifeGainThisTurn(ev, source, g)
			}, "Vanguard Seraph — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
