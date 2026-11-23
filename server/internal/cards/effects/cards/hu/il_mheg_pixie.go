package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Il Mheg Pixie — Creature — Faerie {1}{U}, 2/1 (Final Fantasy, #4462):
//
//	"Flying
//	 Whenever this creature attacks, surveil 1."
//
// Boulderborn Dragon's attack trigger over the footprinted Surveil
// primitive (CR 701.25a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c2c47e13-b581-41a8-a4af-542557a6249a",
		Name:            "Il Mheg Pixie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Il Mheg Pixie — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
