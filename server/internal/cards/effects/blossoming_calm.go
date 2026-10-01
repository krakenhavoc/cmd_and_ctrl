package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blossoming Calm — Instant {W}:
//
//	"You gain hexproof until your next turn. You gain 2 life.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Player hexproof is the player-keyword grant (player_keywords.go),
// stamped with CR 611.2's "until your next turn", as Teferi's
// Protection does. Cast again from rebound in your upkeep, it lasts
// until your following turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "489a60f1-83f8-465b-918f-7d63d4f76d14",
		Name:            "Blossoming Calm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			if err := (GainPlayerKeyword{
				Player:   me,
				Keyword:  KeywordHexproof,
				Label:    "Blossoming Calm — you gain hexproof",
				Duration: DurationUntilYourNextTurn(ctx, me),
			}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: me, Amount: 2}.Apply(ctx)
		},
	})
}
