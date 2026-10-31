package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Sentry, Golden Guardian — Legendary Creature — Human Hero {3}{W},
// 5/5 (Marvel Super Heroes, #4455):
//
//	"Flying, vigilance, indestructible
//	 When The Sentry enters, target opponent creates The Void, a legendary
//	 5/5 black Horror Villain creature token with flying, indestructible,
//	 and "The Void attacks each combat if able.""
//
// The drawback targets (so hexproof and shroud on the player apply) and
// the TARGET OPPONENT is the token's controller, as in Dowsing Dagger. The
// Void is a catalog token template (mshA8VoidToken) whose attack
// requirement travels with the token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1fe549c3-79be-40ef-b57c-de59f24df5ab",
		Name:            "The Sentry, Golden Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "indestructible"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("The Sentry — target opponent creates The Void, a legendary 5/5 black Horror Villain creature token",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						targets := ctx.LegalTargets()
						if len(targets) == 0 {
							return nil
						}
						return CreateToken{Controller: targets[0].ID, Template: mshA8VoidToken(), N: 1}.Apply(ctx)
					}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// mshA8VoidToken is The Void, the legendary 5/5 black Horror Villain with
// flying, indestructible and "The Void attacks each combat if able."
func mshA8VoidToken() game.Card { return tokenFromCatalog(printedMshA8VoidToken) }

// printedMshA8VoidToken is The Void as PRINTED.
func printedMshA8VoidToken() tokenTemplate {
	return tokenTemplate{
		Slug: "the-void-msh",
		Card: game.Card{
			Name:      "The Void",
			TypeLine:  "Token Legendary Creature — Horror Villain",
			Power:     5,
			Toughness: 5,
			Colors:    []string{"B"},
			Keywords:  []string{"flying", "indestructible"},
		},
		Static: []game.StaticAbility{AttacksEachCombat()},
		Text:   "Flying, indestructible\nThe Void attacks each combat if able.",
	}
}
