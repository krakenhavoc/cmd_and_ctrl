package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thornwatch Scarecrow — Artifact Creature — Scarecrow {6}, 4/4
// (Shadowmoor, #5591):
//
//	"This creature has wither as long as you control a green creature.
//	 (It deals damage to creatures in the form of -1/-1 counters.)
//	 This creature has vigilance as long as you control a white
//	 creature."
//
// Rattleblaze Scarecrow's two self grants (layer 6, CR 613.1f), each read
// off the board as it is: the colourless Scarecrow never satisfies its own
// condition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe01e07c-dbca-4598-a944-1b194df44260",
		Name:         "Thornwatch Scarecrow",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("G"))), "wither"),
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("W"))), "vigilance"),
		},
	})
}
