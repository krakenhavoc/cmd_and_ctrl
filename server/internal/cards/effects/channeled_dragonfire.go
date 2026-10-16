package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Channeled Dragonfire — Sorcery {R} (#3931):
//
//	"Channeled Dragonfire deals 2 damage to any target.
//	 Harmonize {5}{R}{R}"
//
// Shock at sorcery speed, and harmonize (CR 702.180) from the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "91179937-6f55-4d1b-8f06-2efe1c8fea7a",
		Name:             "Channeled Dragonfire",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{5}{R}{R}")},
		Targets:          TargetAny(),
		Purpose:          ForTargets(DamageToTarget(0, 2)),
		OnResolve:        damageToFirstTarget(2),
	})
}
