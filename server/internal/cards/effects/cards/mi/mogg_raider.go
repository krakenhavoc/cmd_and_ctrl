package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Raider — Creature — Goblin {R}, 1/1 (Tempest, S82 OLD b755, #5628):
//
//	"Sacrifice a Goblin: Target creature gets +1/+1 until end of turn."
//
// Weirding Shaman's Goblin sacrifice (the Raider may sacrifice itself) paying
// for Battle Rampart's targeted pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "240746f5-caab-46b9-914f-c2c7ecce6ca9",
		Name:         "Mogg Raider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Goblin: Target creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.B25SacrificeAGoblin(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Mogg Raider — +1/+1 until end of turn", 1, 1),
		}},
	})
}
