package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elgaud Inquisitor — Creature — Human Cleric {3}{W}, 2/2 (Dark Ascension, #5571):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)
//	 When this creature dies, create a 1/1 white Spirit creature
//	 token with flying."
//
// Doomed Traveler's dies trigger on a lifelink body. Lifelink is printed
// data the damage path reads (CR 702.15).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e84164cf-c3bf-45ae-9a0a-9dfeacd36768",
		Name:            "Elgaud Inquisitor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Elgaud Inquisitor — create a 1/1 Spirit", effects.Do(effects.CreateToken{
				Template: effects.TokenCard("1/1 white Spirit with flying"),
				N:        1,
			})),
		},
	})
}
