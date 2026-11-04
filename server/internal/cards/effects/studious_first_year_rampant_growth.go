package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Studious First-Year // Rampant Growth — Creature — Bear Wizard {G}, 1/1 //
// Sorcery {1}{G} (preparation card, CR 722; Secrets of Strixhaven, #4456):
//
//	"This creature enters prepared."
//
//	Rampant Growth — "Search your library for a basic land card, put that
//	 card onto the battlefield tapped, then shuffle."
//
// The spell is the printed Rampant Growth, so it shares that card's search
// (rampantGrowthSearch, soc_b36_helpers.go) rather than restating it.
//
// No simplification.
func init() {
	const id = "58b0c737-0a84-4f9a-b3b7-300c5de43874"
	Register(Spec{
		OracleID:     id,
		Name:         "Studious First-Year",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Rampant Growth",
		Purpose:      game.Purpose{Lands: 1},
		Completeness: CompletenessFull,
		OnResolve:    rampantGrowthSearch(),
	})
}
