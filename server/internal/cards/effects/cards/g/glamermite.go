package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glamermite — Creature — Faerie Rogue {2}{U}, 2/2 (Lorwyn Eclipsed,
// #4458):
//
//	"Flash
//	 Flying
//	 When this creature enters, choose one —
//	 • Tap target creature.
//	 • Untap target creature."
//
// Divining Duelist's modal enters trigger without its loot bullet (CR
// 603.3c): the mode and its target are chosen as the trigger goes on the
// stack, and either bullet targets any creature on the battlefield, the
// Glamermite included.
//
// No simplification.
func init() {
	mite := effects.WhenThisEnters("Glamermite — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	mite.Modes = effects.ChooseOne(
		effects.ModeDoing("Tap target creature.",
			effects.TargetCreature("target creature"),
			effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error { return effects.TapTarget{Target: id}.Apply(ctx) })),
		effects.ModeDoing("Untap target creature.",
			effects.TargetCreature("target creature"),
			effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error { return effects.UntapTarget{Target: id}.Apply(ctx) })),
	)
	effects.Register(effects.Spec{
		OracleID:        "a7b0d897-38c0-4f90-81e8-3c5b8d31a607",
		Name:            "Glamermite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered:       []game.TriggeredAbility{mite},
	})
}
