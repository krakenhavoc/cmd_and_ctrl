package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghoulcaller Gisa — Legendary Creature — Human Wizard {3}{B}{B}, 3/4:
//
//	"{B}, {T}, Sacrifice another creature: Create X 2/2 black Zombie
//	 creature tokens, where X is the sacrificed creature's power."
//
// X is the sacrificed creature's power as it last existed on the
// battlefield (Context.SacrificedPower, CR 608.2h), counters and
// anthems included and floored at zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0dd894cb-1968-471c-af08-ea7ec5ce8428",
		Name:         "Ghoulcaller Gisa",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, {T}, Sacrifice another creature: Create X 2/2 black Zombie creature tokens, where X is the sacrificed creature's power.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{B}"), TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CreateToken{Template: TokenCard("2/2 black Zombie"), N: ctx.SacrificedPower()}.Apply(ctx)
			},
		}},
	})
}
