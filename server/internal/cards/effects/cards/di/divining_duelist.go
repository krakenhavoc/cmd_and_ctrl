package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	duelist := effects.WhenThisEnters("Divining Duelist — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	duelist.Modes = effects.ChooseOne(
		effects.ModeDoing("Tap target creature.",
			effects.TargetCreature("target creature"),
			effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error { return effects.TapTarget{Target: id}.Apply(ctx) })),
		effects.ModeDoing("Untap target creature.",
			effects.TargetCreature("target creature"),
			effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error { return effects.UntapTarget{Target: id}.Apply(ctx) })),
		effects.ModeDoing("Draw a card, then discard a card.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.LootOne(ctx.Game, item, 1)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "d0ce03fa-2284-4217-85b2-4db6b8a88c94",
		Name:            "Divining Duelist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{duelist},
	})
}
