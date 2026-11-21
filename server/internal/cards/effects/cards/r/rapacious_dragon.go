package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rapacious Dragon — Creature — Dragon {4}{R}, 3/3 (EDHREC rank
// 3620):
//
//	"Flying
//	 When this creature enters, create two Treasure tokens. (They're
//	 artifacts with "{T}, Sacrifice this token: Add one mana of any
//	 color.")"
//
// The five-drop that pays itself back. Flying rides PrintedKeywords;
// the Treasures are the real token, with its own mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0944ec2e-1dd9-459f-8f1d-667242cf52fe",
		Name:            "Rapacious Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rapacious Dragon — create two Treasures", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 2})),
		},
	})
}
