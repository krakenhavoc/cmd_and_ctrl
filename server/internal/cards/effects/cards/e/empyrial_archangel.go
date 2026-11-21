package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Empyrial Archangel — Creature — Angel {4}{G}{W}{W}{U}, 5/8:
//
//	"Flying
//	 Shroud
//	 All damage that would be dealt to you is dealt to this creature
//	 instead."
//
// ADR 0108 §9 decision 4 (#1905): a static redirection (CR 614.9) to the
// Archangel itself.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a105959-dfce-4202-b37f-ae635dfcb30b",
		Name:         "Empyrial Archangel",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.RedirectYourDamageToThis("Empyrial Archangel — damage to you is dealt to it instead"),
		},
	})
}
