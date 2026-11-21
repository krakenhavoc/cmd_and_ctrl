package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wellgabber Apothecary — Creature — Merfolk Cleric {4}{W}, 2/3:
//
//	"{1}{W}: Prevent all damage that would be dealt to target tapped
//	 Merfolk or Kithkin creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the not-one-use shield pinned to
// the target. "Tapped" and the creature types are checked as the target
// is chosen and again as the ability resolves (CR 608.2b); once made,
// the shield lasts the turn even if the creature untaps.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10cdb0af-a206-4eec-ba9b-634f6ec583d8",
		Name:         "Wellgabber Apothecary",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.ShieldTargetRow(
			"{1}{W}: Prevent all damage that would be dealt to target tapped Merfolk or Kithkin creature this turn.",
			effects.ManaCost("{1}{W}"),
			effects.TargetCreature("target tapped Merfolk or Kithkin creature", effects.Not(effects.Untapped()), effects.AnySubtype("Merfolk", "Kithkin")))},
	})
}
