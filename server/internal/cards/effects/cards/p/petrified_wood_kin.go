package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Petrified Wood-Kin — Creature — Elemental Warrior {6}{G}, 3/3 (Guildpact):
//
//	"This spell can't be countered.
//	 Bloodthirst X (This creature enters with X +1/+1 counters on it, where X
//	 is the damage dealt to your opponents this turn.)
//	 Protection from instants"
//
// "Can't be countered" is the spec's CantBeCountered (Balustrade Wurm). Bloodthirst X (CR 702.54b) is
// the shared entry replacement of this batch's helpers: the total damage all
// the controller's opponents were dealt this turn. Protection from instants is
// a printed protection keyword (Emrakul, the Promised End's), so an instant
// can't damage, enchant, block or target it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3bc9649c-4350-4c6a-8441-dc76a5f833af",
		Name:            "Petrified Wood-Kin",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"protection from instants"},
		Replacements:    []game.ReplacementEffect{effects.OldB571BloodthirstX("Petrified Wood-Kin")},
	})
}
