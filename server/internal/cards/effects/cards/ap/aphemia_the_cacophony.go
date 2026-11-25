package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aphemia, the Cacophony — Legendary Enchantment Creature — Harpy {1}{B}, 2/1
// (Theros Beyond Death, #5514):
//
//	"Flying
//	 At the beginning of your end step, you may exile an enchantment card from your
//	 graveyard. If you do, create a 2/2 black Zombie creature token."
//
// AphemiaEndStep (s82_old_b136_helpers.go): one keyed optional pick over the
// enchantment cards in your graveyard, and the Zombie only if a card was exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "820dab0f-3b6b-437c-ba4b-2311437709c3",
		Name:            "Aphemia, the Cacophony",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AphemiaEndStep("Aphemia, the Cacophony — you may exile an enchantment card from your graveyard; if you do, create a 2/2 black Zombie"),
		},
	})
}
