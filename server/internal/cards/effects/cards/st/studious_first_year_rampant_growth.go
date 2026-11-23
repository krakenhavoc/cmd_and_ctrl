package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Studious First-Year",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Rampant Growth",
		Purpose:      game.Purpose{Lands: 1},
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.RampantGrowthSearch(),
	})
}
