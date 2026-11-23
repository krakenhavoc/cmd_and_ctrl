package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demon of Catastrophes — Creature — Demon {2}{B}{B}, 6/6 (Jumpstart 2022,
// #5948):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Flying, trample"
//
// Rite of Consumption's additional cost, paid as the spell is cast
// (CR 601.2b, 601.2h); the sacrificed creature may be any creature the
// caster controls. Flying and trample are the printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e0b838d-c858-4aa5-999e-c4ef9d29571b",
		Name:            "Demon of Catastrophes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		AdditionalCost:  effects.SacrificeCost("a creature", effects.Creature()),
	})
}
