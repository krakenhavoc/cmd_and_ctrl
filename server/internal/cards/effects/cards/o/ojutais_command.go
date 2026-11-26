package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ojutai's Command — Instant {2}{W}{U} (Dragons of Tarkir):
//
//	"Choose two —
//	 • Return target creature card with mana value 2 or less from your graveyard
//	   to the battlefield.
//	 • You gain 4 life.
//	 • Counter target creature spell.
//	 • Draw a card."
//
// The chosen modes run in printed order (CR 700.2, 608.2c): chosen with the
// first and the last, the creature returns before the draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1fafe93-8fd2-44c1-a4a0-5c9ecee550cc",
		Name:         "Ojutai's Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Return target creature card with mana value 2 or less from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target creature card with mana value 2 or less from your graveyard", effects.YouOwn(), effects.And(effects.Creature(), effects.ManaValueLE(2))),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				})),
			effects.ModeWithPurpose(effects.ModeDoing("You gain 4 life.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.GainLife{Amount: 4}.Apply(ctx)
				}),
				game.Purpose{LifeGain: 4}),
			effects.ModeDoing("Counter target creature spell.",
				effects.TargetSpell("target creature spell", effects.Creature()), effects.CounterTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Draw a card.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error { return effects.DrawCards{N: 1}.Apply(ctx) }),
				game.Purpose{Draws: 1}),
		),
	})
}
