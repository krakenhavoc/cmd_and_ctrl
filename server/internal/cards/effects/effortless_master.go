package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Effortless Master — Creature — Orc Monk {2}{U}{R}, 4/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Vigilance
//	 Menace
//	 This creature enters with two +1/+1 counters on it if you've cast two
//	 or more spells this turn."
//
// An enters-with-counters self-replacement (Diregraf Colossus's counted
// form) whose count is read as it enters: two when its controller's
// spell tally for the turn is two or more, otherwise none. The tally
// holds every spell they cast this turn, the Master itself when it was
// cast (the ruling), whether or not the others resolved or were
// countered. A Master put onto the battlefield without being cast reads
// the same tally, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "185e0892-194f-43ac-853f-1e142bf8476e",
		Name:            "Effortless Master",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				if g.SpellsCastThisTurn[src.Controller].Total >= 2 {
					return 2
				}
				return 0
			}, "Effortless Master: enters with two +1/+1 counters if you've cast two or more spells this turn"),
		},
	})
}
