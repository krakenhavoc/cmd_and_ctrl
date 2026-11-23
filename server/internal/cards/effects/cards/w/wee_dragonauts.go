package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wee Dragonauts — Creature — Faerie Wizard {1}{U}{R}, 1/3 (Guildpact,
// #5599):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, this creature gets
//	 +2/+0 until end of turn."
//
// Festival Crasher's cast trigger on a flyer: each instant or sorcery
// its controller casts gives it +2/+0 as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9e56b1d-5ad2-4807-8c4c-1a1a6a25ecbf",
		Name:            "Wee Dragonauts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Wee Dragonauts — gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Wee Dragonauts — +2/+0")),
		},
	})
}
