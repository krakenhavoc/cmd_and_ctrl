package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serpent-Blade Assailant — Creature — Elf Warrior {2}{G}, 2/1 (March of the Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 Deathtouch"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// deathtouch until end of turn; the card has it itself, declared in PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "66a131ae-dbb2-4e7b-b6ce-7695f47c8f87",
		Name:            "Serpent-Blade Assailant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Serpent-Blade Assailant", 1, "", "deathtouch"),
		},
	})
}
