package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lich Lord of Unx — Creature — Zombie Wizard {1}{U}{B}, 2/2:
//
//	"{U}{B}, {T}: Create a 1/1 blue and black Zombie Wizard creature
//	 token.
//	 {U}{U}{B}{B}: Target player loses X life and mills X cards, where X
//	 is the number of Zombies you control."
//
// X is counted as the second ability resolves (CR 608.2h), over every
// Zombie permanent its controller controls, the Lich Lord included.
// The life is lost first, then the cards are milled, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6643fc8e-548e-48db-9f67-ffbca9e57016",
		Name:         "Lich Lord of Unx",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{U}{B}, {T}: Create a 1/1 blue and black Zombie Wizard creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    Plus(ManaCost("{U}{B}"), TapCost()),
				Effect:  Do(CreateToken{Template: TokenCard("1/1 blue and black Zombie Wizard"), N: 1}),
			},
			{
				Label:   "{U}{U}{B}{B}: Target player loses X life and mills X cards, where X is the number of Zombies you control.",
				Cost:    ManaCost("{U}{U}{B}{B}"),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						x := zombiesYouControl(g, item.Controller)
						if x == 0 {
							return nil
						}
						if err := g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -x); err != nil {
							return err
						}
						return MillCards{Player: t.ID, N: x}.Apply(ctx)
					}
					return nil
				},
			},
		},
	})
}
