package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Tinkerer — Creature — Goblin Artificer {1}{R}, 1/2 (Mirage):
//
//	"{R}, {T}: Destroy target artifact. That artifact deals damage equal to
//	 its mana value to this creature."
//
// The artifact's mana value is read before it is destroyed (CR 608.2h), and it
// deals the damage as the source whether or not the destruction happened (an
// indestructible artifact still deals it). A Tinkerer that has left the
// battlefield is dealt nothing. The artifact must still be a legal target as
// the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0afc2d2f-4fdb-440f-9c35-45bf103c0437",
		Name:         "Goblin Tinkerer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: Destroy target artifact. That artifact deals damage equal to its mana value to this creature.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				card, ok := g.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				mv := card.ManaValue()
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.DealDamage{Source: t.ID, Target: item.SourceCardID, Amount: mv}.Apply(ctx)
			},
		}},
	})
}
