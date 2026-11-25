package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Court Cleric — Creature — Human Cleric {W}, 1/1 (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"Lifelink
//	 This creature gets +1/+1 as long as you control an Ajani planeswalker."
//
// Vivien's Crocodile's conditional self pump over a planeswalker with the
// Ajani subtype, and the lifelink keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dec9edbb-677b-4773-8bc4-b388d7592276",
		Name:            "Court Cleric",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+1 as long as you control an Ajani planeswalker.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Planeswalker(), effects.Subtype("Ajani"))),
			Apply:     effects.PumpBy(1, 1),
		}},
	})
}
