package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulcoil Viper — Creature — Snake {2}{B}, 2/3 (The Lost Caverns of
// Ixalan, #4470):
//
//	"{B}, {T}, Sacrifice this creature: Return target creature card from
//	 your graveyard to the battlefield with a finality counter on it.
//	 Activate only as a sorcery."
//
// Grim Reaper's Scythe's return (grimReapersScytheReturn) with the Viper
// as its own cost. Targets are chosen before costs are paid
// (CR 601.2c, CR 602.2b), so the Viper can't target itself. The {T}
// makes it subject to summoning sickness (CR 302.6), and sorcery timing
// is CR 602.5d. A creature with a finality counter that would die is
// exiled instead (CR 122.1h), which the engine applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24c3c505-ddaa-4adb-983e-7c78d9933cc7",
		Name:         "Soulcoil Viper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{B}, {T}, Sacrifice this creature: Return target creature card from your graveyard to the battlefield with a finality counter on it. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.Reanimates(0)),
			Cost:         effects.Plus(effects.ManaCost("{B}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetCreatureInYourGraveyard(),
			Effect:       effects.GrimReapersScytheReturn,
		}},
	})
}
