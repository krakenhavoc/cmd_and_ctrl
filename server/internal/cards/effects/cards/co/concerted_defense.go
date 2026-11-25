package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Concerted Defense — Instant {U} (Zendikar Rising, #5508):
//
//	"Counter target noncreature spell unless its controller pays {1} plus an
//	 additional {1} for each creature in your party. (Your party consists of
//	 up to one each of Cleric, Rogue, Warrior, and Wizard.)"
//
// The tax is {1} plus the party (CR 700.8) counted as the spell resolves, so
// it is {1} to {5}: Spell Pierce's pay-or-be-countered prompt over that total.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77f02ae6-2411-4378-b5e0-cc1cf5964458",
		Name:         "Concerted Defense",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target noncreature spell", effects.Noncreature()),
		OnResolve:    effects.OldB096CounterUnlessPaysParty("Concerted Defense"),
	})
}
