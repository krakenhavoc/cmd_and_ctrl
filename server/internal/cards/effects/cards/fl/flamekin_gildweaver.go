package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamekin Gildweaver — Creature — Elemental Sorcerer {3}{R}, 4/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Trample
//	 When this creature enters, create a Treasure token. (It's an artifact
//	 with "{T}, Sacrifice this token: Add one mana of any color.")"
//
// Professional Wrestler's entry trigger on a trampler: the shared
// TreasureToken template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bc8427fb-e5ee-4b44-9a4f-6d9f01b4b896",
		Name:            "Flamekin Gildweaver",
		Purpose:         game.Purpose{Tokens: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Flamekin Gildweaver — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
