package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intruding Soulrager — Creature — Spirit {U}{R}, 2/2:
//
//	"Vigilance
//	 {T}, Sacrifice a Room: This creature deals 2 damage to each
//	 opponent. Draw a card."
//
// The Room is sacrificed as a cost, named at announce (SacrificeN over
// the Room subtype), so the activation is refused without one. The
// damage is dealt by the Soulrager, to each opponent; the draw follows.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7673d0db-07d6-4b40-a32e-f2c98d7ea7c1",
		Name:            "Intruding Soulrager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice a Room: This creature deals 2 damage to each opponent. Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), SacrificeN(1, "a Room", HasSubtype("Room"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := ctx.Game.DamageInstanceForEffect(func() error {
					for _, opp := range ctx.Opponents() {
						if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: 2}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					return err
				}
				return DrawCards{N: 1}.Apply(ctx)
			},
		}},
	})
}
