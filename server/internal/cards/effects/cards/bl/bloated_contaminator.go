package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloated Contaminator — Creature — Phyrexian Beast {2}{G}, 4/4:
//
//	"Trample
//	 Toxic 1
//	 Whenever this creature deals combat damage to a player,
//	 proliferate."
//
// Toxic is the engine's damage tail (ADR 0056, #748): combat damage to
// a player also gives them one poison counter, in the same step as the
// damage and before this trigger resolves — so the proliferate finds
// the poison already there and adds another.
//
// The controller chooses what to proliferate (#2525), as with every
// catalogued proliferate. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "090018e0-4dcb-4b3c-b4e0-7ba62de0484d",
		Name:            "Bloated Contaminator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "toxic 1"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Bloated Contaminator — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
