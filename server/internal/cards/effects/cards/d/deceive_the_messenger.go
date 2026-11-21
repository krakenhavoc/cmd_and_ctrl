package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deceive the Messenger — Instant {U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Target creature gets -3/-0 until end of turn.
//	 Amass Orcs 1."
//
// The shrink, then the keyword action (CR 701.47) as the next step of the same
// resolution. With the target illegal the spell is countered on resolution
// (CR 608.2b) and does not amass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9720545-37f4-42f1-932f-13ef88c9ea72",
		Name:         "Deceive the Messenger",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Amass: 1, Targets: game.ForTargets(effects.HarmUntilEOT(0, -3, 0))},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: -3, Label: "Deceive the Messenger — -3/-0"}.Apply(ctx)
			}),
			effects.Amass{Subtype: "Orc", N: 1},
		),
	})
}
