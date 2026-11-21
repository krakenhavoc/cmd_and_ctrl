package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plague Witch — Creature — Elf Spellshaper {1}{B}, 1/1
// (Nemesis, #5618):
//
//	"{B}, {T}, Discard a card: Target creature gets -1/-1 until end of turn."
//
// The spellshaper cost (nemSpellshaperCost) on Grandmother Sengir's
// body, boostTargetUntilEOT at -1/-1. A 1/1 target dies to the
// state-based check (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "532ab335-4873-405b-84b3-c447207172b5",
		Name:         "Plague Witch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}, Discard a card: Target creature gets -1/-1 until end of turn.",
			Cost:    effects.NemSpellshaperCost("{B}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
			Effect:  effects.BoostTargetUntilEOT("Plague Witch — -1/-1 until end of turn", -1, -1),
		}},
	})
}
