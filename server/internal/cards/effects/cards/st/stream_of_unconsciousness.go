package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stream of Unconsciousness — Kindred Instant — Wizard {U} (Morningtide,
// #5592):
//
//	"Target creature gets -4/-0 until end of turn. If you control a Wizard,
//	 draw a card."
//
// Code of Constraint's -4/-0 and draw, with the draw gated on a Wizard you
// control as the spell resolves (Mana Sculpt's ControlsA, which reads
// effective subtypes, so a changeling counts). With its only target gone
// the spell does nothing at all (CR 608.2b), so no card either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7da64f81-64d0-4019-a208-9ee8b6782d85",
		Name:         "Stream of Unconsciousness",
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.HarmUntilEOT(0, -4, 0))},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -4, Label: "Stream of Unconsciousness — -4/-0"}.Apply(ctx)
			}),
			effects.Instruction(effects.MorDrawIfYouControlAWizard),
		),
	})
}
