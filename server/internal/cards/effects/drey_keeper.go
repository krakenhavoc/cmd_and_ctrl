package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drey Keeper — Creature — Elf Druid {3}{B}{G}, 2/2 (Modern Horizons 2,
// #5501):
//
//	"When this creature enters, create two 1/1 green Squirrel creature
//	 tokens.
//	 {3}{B}: Squirrels you control get +1/+0 and gain menace until end of
//	 turn."
//
// Deep Forest Hermit's Squirrels, and a group pump whose set is chosen
// once as the ability resolves (CR 611.2c): BoostUntilEOT and Along the
// Crooked Way's GrantKeywordUntilEOT over the same predicate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7223ef44-0bfb-4108-810d-134174ff26b0",
		Name:         "Drey Keeper",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Drey Keeper — create two 1/1 green Squirrel creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 green Squirrel"), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}: Squirrels you control get +1/+0 and gain menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				squirrels := And(Creature(), YouControl(), HasSubtype(ctx.TextSubtype("Squirrel")))
				if err := (BoostUntilEOT{Match: squirrels, Power: 1, Label: "Drey Keeper — Squirrels +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: squirrels, Keywords: []string{"menace"}, Label: "Drey Keeper — Squirrels gain menace"}.Apply(ctx)
			},
		}},
	})
}
