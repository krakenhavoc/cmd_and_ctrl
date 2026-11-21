package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blazethorn Scarecrow — Artifact Creature — Scarecrow {5}, 3/3
// (Shadowmoor, #5591):
//
//	"This creature has haste as long as you control a red creature.
//	 This creature has wither as long as you control a green creature.
//	 (It deals damage to creatures in the form of -1/-1 counters.)"
//
// Rattleblaze Scarecrow's two self grants (layer 6, CR 613.1f), each read
// off the board as it is: the colourless Scarecrow never satisfies its own
// condition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f7a4ffb-83f7-46e5-b0a7-c8b068af6969",
		Name:         "Blazethorn Scarecrow",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("R"))), "haste"),
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("G"))), "wither"),
		},
	})
}
