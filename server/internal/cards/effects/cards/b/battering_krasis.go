package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battering Krasis — Creature — Shark Beast {2}{G}, 2/1:
//
//	"Trample
//	 Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)"
//
// Trample is read at combat damage (CR 702.19b); evolve is the engine's
// keyword trigger (game/evolve.go, #1805), as on Elusive Krasis.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3d0c1301-4fa7-4245-a13f-f613188eb638",
		Name:            "Battering Krasis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordEvolve},
	})
}
