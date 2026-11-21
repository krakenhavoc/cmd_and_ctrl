package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Denethor, Stone Seer — Legendary Creature — Human Noble {1}{U}, 1/3
// (Tales of Middle-earth Commander):
//
//	"When Denethor enters, scry 2.
//	 {3}{R}, {T}, Sacrifice Denethor: Target player becomes the monarch.
//	 Denethor deals 3 damage to any target."
//
// Survey Mechan's two-slot activation: "target player" is slot 0 and
// "any target" slot 1, each judged on its own at resolution (CR 608.2b).
// Denethor was sacrificed as a cost, so the damage is dealt by him as he
// last existed on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6669a19e-7944-44c9-ac37-a92016783375",
		Name:         "Denethor, Stone Seer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Denethor, Stone Seer — scry 2", Do(Scry{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}, {T}, Sacrifice Denethor: Target player becomes the monarch. Denethor deals 3 damage to any target.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{3}{R}"), TapCost(), SacrificeThis()),
			Targets: TargetPlayer("target player").Then(TargetAny()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if t, ok := ctx.ClauseTarget(0); ok {
					if err := (BecomeTheMonarch{Player: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				if t, ok := ctx.ClauseTarget(1); ok {
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
