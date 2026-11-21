package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tributary Vaulter — Creature — Merfolk Warrior {2}{W}, 1/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Flying
//	 Whenever this creature becomes tapped, another target Merfolk you
//	 control gets +2/+0 until end of turn."
//
// Hemosymbic Mite's trigger: "becomes tapped" is the tap event and the
// attack that taps it (ThisBecameTapped), so a vigilance attack is not
// one and any other tap counts. The target is chosen as the trigger goes
// on the stack and must be another Merfolk (a changeling is one) its
// controller controls; with none the trigger is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "991f4eec-cf3e-4554-ba88-c44bb0d255f0",
		Name:            "Tributary Vaulter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Tributary Vaulter — another target Merfolk you control gets +2/+0 until end of turn",
				effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.BoostUntilEOT{Target: id, Power: 2, Label: "Tributary Vaulter — +2/+0"}.Apply(ctx)
				}))),
				effects.Another(effects.TargetCreature("another target Merfolk you control", effects.YouControl(), effects.HasSubtype("Merfolk")))),
		},
	})
}
