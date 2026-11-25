package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sporeweb Weaver — Creature — Spider {2}{G}, 1/4 (Core Set 2021, #5511):
//
//	"Reach, hexproof from blue
//	 Whenever this creature is dealt damage, you gain 1 life and create a
//	 1/1 green Saproling creature token."
//
// Reach and hexproof from blue are printed keywords (CR 702.11d). The
// trigger is Enrage's shape: one per batch of damage the Weaver is dealt
// (CR 120.3), never for 0 damage (CR 120.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e7c3881-6fa6-4fd2-ae86-9e3a0fcecdbd",
		Name:            "Sporeweb Weaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", game.HexproofFromColor("U")},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Sporeweb Weaver — you gain 1 life and create a 1/1 green Saproling creature token",
				effects.Do(effects.GainLife{Amount: 1}, effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1})),
		},
	})
}
