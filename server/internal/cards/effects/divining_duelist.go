package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Divining Duelist — Creature — Merfolk Wizard {2}{U}, 3/2:
//
//	"Flash
//	 When this creature enters, choose one —
//	 • Tap target creature.
//	 • Untap target creature.
//	 • Draw a card, then discard a card."
//
// A modal enters trigger (CR 603.3c, Charming Scoundrel's shape): the
// mode and its target are chosen as the trigger goes on the stack, so
// the loot bullet asks for no target. The two creature bullets target
// ANY creature on the battlefield, yours or not. The loot draws before
// it discards, through the shared drawThenDiscard.
//
// No simplification.
func init() {
	duelist := WhenThisEnters("Divining Duelist — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	duelist.Modes = ChooseOne(
		ModeDoing("Tap target creature.",
			TargetCreature("target creature"),
			rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error { return TapTarget{Target: id}.Apply(ctx) })),
		ModeDoing("Untap target creature.",
			TargetCreature("target creature"),
			rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error { return UntapTarget{Target: id}.Apply(ctx) })),
		ModeDoing("Draw a card, then discard a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return lootOne(ctx.Game, item, 1)
			}),
	)
	Register(Spec{
		OracleID:        "d0ce03fa-2284-4217-85b2-4db6b8a88c94",
		Name:            "Divining Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{duelist},
	})
}
