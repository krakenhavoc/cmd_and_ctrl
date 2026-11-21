package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wingrattle Scarecrow — Artifact Creature — Scarecrow {3}, 2/2:
//
//	"This creature has flying as long as you control a blue creature.
//	 This creature has persist as long as you control a black creature."
//
// Rattleblaze Scarecrow's shape (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "034264da-a404-4037-b215-b095bba77490",
		Name:         "Wingrattle Scarecrow",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("U"))), "flying"),
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("B"))), game.KeywordPersist),
		},
	})
}
