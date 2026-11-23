package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rattleblaze Scarecrow — Artifact Creature — Scarecrow {6}, 5/3:
//
//	"This creature has persist as long as you control a black creature.
//	 This creature has haste as long as you control a red creature."
//
// Both are layer-6 self grants read off the board as it is. A Scarecrow
// that dies while you control a black creature had persist as it last
// existed, and returns (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "313130ab-13db-4d0a-b3f3-b60e060b303c",
		Name:         "Rattleblaze Scarecrow",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("B"))), game.KeywordPersist),
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("R"))), "haste"),
		},
	})
}
