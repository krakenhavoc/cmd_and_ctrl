package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geyser Leaper — Creature — Human Warrior Ally {4}{U}, 4/3 (Avatar: The
// Last Airbender, #4459):
//
//	"Flying
//	 Waterbend {4}: Draw a card, then discard a card. (While paying a
//	 waterbend cost, you can tap your artifacts and creatures to help.
//	 Each one pays for {1}.)"
//
// WaterbendCost("{4}") is the whole cost (CR 701.67a), and the loot is
// the catalog's one-draw, one-discard body, the discard chosen as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dfdcb451-a597-46fe-893c-49e6dbda5f71",
		Name:            "Geyser Leaper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Waterbend {4}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.WaterbendCost("{4}"),
			Effect:  effects.B35LootOne,
		}},
	})
}
