package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bastion of Remembrance — Enchantment {2}{B}:
//
//	"When this enchantment enters, create a 1/1 white Human Soldier
//	creature token."
//	"Whenever a creature you control dies, each opponent loses 1 life
//	and you gain 1 life."
//
// A Zulaport Cutthroat that can't be killed by creature removal —
// which in an aristocrats deck is the point: the drain survives the
// board wipe that fuels it. It even brings its own first body.
//
// Two independent halves, both triggered abilities on the stack: the
// ETB token (moved off the direct AsEnters hook in #578, so it can be
// responded to) and a dies-trigger for the drain, identical in shape to
// the Cutthroat's.

func init() {
	Register(Spec{
		OracleID:     "c7f33cea-2ec8-4081-9208-a5b1d86721b3",
		Name:         "Bastion of Remembrance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bastion of Remembrance — create a 1/1 Human Soldier", Do(CreateToken{
				Template: TokenCard("1/1 white Human Soldier"),
				N:        1,
			})),
			WheneverACreatureYouControlDies("Bastion of Remembrance — each opponent loses 1", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, opp := range ctx.Opponents() {
					if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -1); err != nil {
						return err
					}
				}
				return GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
			}),
		},
	})
}
