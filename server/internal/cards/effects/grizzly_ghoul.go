package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grizzly Ghoul — Creature — Zombie Bear {2}{B}{G}, 4/3 (Midnight Hunt,
// #5497):
//
//	"Trample
//	 This creature enters with a +1/+1 counter on it for each creature that
//	 died this turn."
//
// Bone Devourer's CR 614.1c self-replacement, counted off the table-wide tally
// of creatures that died this turn (b11CreaturesDiedThisTurn, anyone's,
// tokens included).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86783f8d-26c9-4b8c-a697-ab6facafeb67",
		Name:            "Grizzly Ghoul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return b11CreaturesDiedThisTurn(g)
			}, "Grizzly Ghoul enters with a +1/+1 counter on it for each creature that died this turn"),
		},
	})
}
