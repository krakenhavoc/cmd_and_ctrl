package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Virulent Swipe — Instant {B}:
//
//	"Target creature gets +2/+0 and gains deathtouch until end of turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "77f53f6a-36ab-43d1-b547-4578fd724864",
		Name:            "Virulent Swipe",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (BoostUntilEOT{Target: target, Power: 2, Label: "Virulent Swipe — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: target, Keywords: []string{"deathtouch"}, Label: "Virulent Swipe — deathtouch"}.Apply(ctx)
		},
	})
}
