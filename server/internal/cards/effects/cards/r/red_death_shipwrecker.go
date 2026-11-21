package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Red Death, Shipwrecker — Legendary Creature — Crab Mutant {U}{R}, 1/3
// (Fallout, #5948):
//
//	"Alluring Eyes — {T}: Goad target creature an opponent controls. That
//	 player draws a card. You add {R}. (Until your next turn, that creature
//	 attacks each combat if able and attacks a player other than you if
//	 able.)"
//
// A targeted activated ability, not a mana ability (CR 605.1a: it has a
// target), so the {R} is added as it resolves and empties at the end of the
// step (CR 106.4). The goad is the shared one (GoadTarget); the draw is by the
// creature's controller as it is now, a creature that has left or changed
// hands being read as it resolves, and a target that is no longer legal does
// nothing at all, the mana included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d40b2c3-9f39-45c6-8b25-12353f300613",
		Name:         "Red Death, Shipwrecker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Goad target creature an opponent controls. That player draws a card. You add {R}.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				if err := effects.GoadTarget(g, item); err != nil {
					return err
				}
				if err := (effects.DrawCards{Player: c.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.AddMana{Produced: "{R}"}.Apply(ctx)
			},
		}},
	})
}
