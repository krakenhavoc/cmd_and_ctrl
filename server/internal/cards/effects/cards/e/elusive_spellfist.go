package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elusive Spellfist — Creature — Human Monk {1}{U}, 1/3 (Dragons of Tarkir, #5551):
//
//	"Whenever you cast a noncreature spell, this creature gets +1/+0 until end of turn and can't be blocked this turn."
//
// Monastery Mentor's cast trigger. Both halves apply to the Spellfist only if
// it is still the permanent that triggered (CR 400.7).
//
// No simplification.

func init() {
	const label = "Elusive Spellfist — +1/+0 until end of turn and can't be blocked this turn"
	effects.Register(effects.Spec{
		OracleID:     "b9ba00f2-b11c-45d8-9c38-f9ade5d07d0f",
		Name:         "Elusive Spellfist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), label, func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Elusive Spellfist — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "Elusive Spellfist — can't be blocked"}.Apply(ctx)
			}),
		},
	})
}
