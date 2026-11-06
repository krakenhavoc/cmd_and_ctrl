package effects

import (
	"github.com/google/uuid"

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
	mite := WhenThisEnters("Glamermite — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	mite.Modes = ChooseOne(
		ModeDoing("Tap target creature.",
			TargetCreature("target creature"),
			rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error { return TapTarget{Target: id}.Apply(ctx) })),
		ModeDoing("Untap target creature.",
			TargetCreature("target creature"),
			rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error { return UntapTarget{Target: id}.Apply(ctx) })),
	)
	Register(Spec{
		OracleID:        "a7b0d897-38c0-4f90-81e8-3c5b8d31a607",
		Name:            "Glamermite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered:       []game.TriggeredAbility{mite},
	})
}
