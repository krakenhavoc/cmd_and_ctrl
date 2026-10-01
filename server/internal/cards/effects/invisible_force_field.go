package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invisible Force Field — Instant {1}{W}:
//
//	"Up to four target permanents you control gain indestructible until
//	 end of turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Each target still legal as the spell resolves gains indestructible
// (CR 608.2b); one that left in response is skipped. Rebound is the
// engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c96d67fb-359c-4439-aa1b-598d9bca2880",
		Name:            "Invisible Force Field",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetPermanent("up to four target permanents you control", YouControl()).WithCount(0, 4),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"indestructible"},
					Label:    "Invisible Force Field — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
