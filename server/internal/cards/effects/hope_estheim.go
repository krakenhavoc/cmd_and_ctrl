package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hope Estheim — Legendary Creature — Human Wizard {W}{U}, 2/2 (Final
// Fantasy, #4462):
//
//	"Lifelink
//	 At the beginning of your end step, each opponent mills X cards,
//	 where X is the amount of life you gained this turn."
//
// X is the life the controller gained this turn (b15LifeGainedThisTurn,
// lifelink damage included, CR 702.15b), read as the trigger resolves; an
// X of 0 mills nothing. Each opponent mills in turn order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "187a110c-032a-42e1-be2d-3cfaaa9495f8",
		Name:            "Hope Estheim",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Hope Estheim — each opponent mills X cards, where X is the life you gained this turn",
				func(g *game.Game, item *game.StackItem) error {
					x := b15LifeGainedThisTurn(g, item.Controller)
					if x <= 0 {
						return nil
					}
					ctx := NewContext(g, item)
					eachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
						_ = MillCards{Player: opp, N: x}.Apply(ctx)
						return false
					})
					return nil
				}),
		},
	})
}
