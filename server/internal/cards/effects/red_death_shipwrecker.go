package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0d40b2c3-9f39-45c6-8b25-12353f300613",
		Name:         "Red Death, Shipwrecker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Goad target creature an opponent controls. That player draws a card. You add {R}.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature an opponent controls", OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				if err := GoadTarget(g, item); err != nil {
					return err
				}
				if err := (DrawCards{Player: c.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return AddMana{Produced: "{R}"}.Apply(ctx)
			},
		}},
	})
}
