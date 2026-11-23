package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demon's Grasp — Sorcery {4}{B} (Battle for Zendikar, #5547):
//
//	"Target creature gets -5/-5 until end of turn."
//
// Pull Under's body at sorcery speed: the shrink is a layer-7c
// modification until end of turn, so a creature with 5 or less
// toughness dies to the state-based check (CR 704.5f), and a target
// that has left by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13228a22-d63a-4952-a482-4f84da696e88",
		Name:         "Demon's Grasp",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -5, -5)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			return effects.BoostUntilEOT{Target: id, Power: -5, Toughness: -5, Label: "Demon's Grasp — -5/-5"}.Apply(ctx)
		})),
	})
}
