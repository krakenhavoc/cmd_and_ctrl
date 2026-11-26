package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Everflame Eidolon — Enchantment Creature — Spirit {1}{R}, 1/1 (Born of the
// Gods, S82 OLD b377, #5559):
//
//	"Bestow {2}{R}
//	 {R}: This creature gets +1/+0 until end of turn. If it's an Aura,
//	 enchanted creature gets +1/+0 until end of turn instead.
//	 Enchanted creature gets +1/+1."
//
// Bestow is the shared alternative cost (ADR 0141). The firebreathing goes to
// the source itself, or to the host while the source is a bestowed Aura
// (S82OldB377PumpSelfOrEnchanted, read as the ability resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2c6b2f4a-4e0e-4dc1-92a8-c07e5cbba5f4",
		Name:             "Everflame Eidolon",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{2}{R}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn. If it's an Aura, enchanted creature gets +1/+0 until end of turn instead.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.S82OldB377PumpSelfOrEnchanted("Everflame Eidolon — +1/+0", 1, 0),
		}},
		Static: []game.StaticAbility{effects.PumpAttached(1, 1)},
	})
}
