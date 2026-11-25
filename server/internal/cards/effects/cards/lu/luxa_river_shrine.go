package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Luxa River Shrine — Artifact {3} (Amonkhet, S82 OLD batch 242, #5535):
//
//	"{1}, {T}: You gain 1 life. Put a brick counter on this artifact.
//	 {T}: You gain 2 life. Activate only if there are three or more brick
//	 counters on this artifact."
//
// Oracle's Vault's two rows: the first puts the brick counter on as it resolves,
// the second is gated on the Shrine's own brick counters, read as it is
// activated.
//
// No simplification.
func init() {
	const brick = "brick"
	effects.Register(effects.Spec{
		OracleID:     "1eb64899-8865-48cc-bd58-529b1294ec99",
		Name:         "Luxa River Shrine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: You gain 1 life. Put a brick counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GainLife{Player: ctx.Controller(), Amount: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.AddCounter{Target: ctx.Source(), Kind: brick, N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "{T}: You gain 2 life. Activate only if there are three or more brick counters on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Condition: func(g *game.Game, _, source uuid.UUID) bool {
					c, ok := g.LookupCardForEffect(source)
					return ok && c.Counters[brick] >= 3
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
				},
			},
		},
	})
}
