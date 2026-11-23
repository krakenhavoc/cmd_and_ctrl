package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shaper Apprentice — Creature — Merfolk Wizard {1}{U}, 2/1 (Jumpstart 2022, #5948):
//
//	"This creature has flying as long as you control another Merfolk."
//
// Barrow Naughty's "has <keyword> as long as you control another <type>"
// static, with flying; read through the layers each time they are
// recomputed, so the keyword comes and goes with the other Merfolk.
//
// No simplification.
func init() {
	flying := effects.KeywordGrant(effects.HobA9WhileYouControlAnother("Merfolk"), "flying")
	flying.Label = "This creature has flying as long as you control another Merfolk."
	effects.Register(effects.Spec{
		OracleID:     "ee6b0bb4-9c40-435b-8a5d-fd89d7e29ae1",
		Name:         "Shaper Apprentice",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{flying},
	})
}
