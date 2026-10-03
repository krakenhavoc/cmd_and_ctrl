package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clearwater Goblet — Artifact {5}:
//
//	"Sunburst (This artifact enters with a charge counter on it for
//	 each color of mana spent to cast it.)
//	 At the beginning of your upkeep, you may gain life equal to the
//	 number of charge counters on this artifact."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). The upkeep trigger is optional, and reads the
// counters as it resolves.
func init() {
	Register(Spec{
		OracleID:            "ffecd3e8-aa28-41a1-8154-a54ae3ce8016",
		Name:                "Clearwater Goblet",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Clearwater Goblet — gain life equal to its charge counters",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok || info.Counters[game.CounterCharge] <= 0 {
						return nil
					}
					return GainLife{Player: item.Controller, Amount: info.Counters[game.CounterCharge]}.Apply(ctx)
				}), "Gain life equal to the charge counters on Clearwater Goblet?"),
		},
	})
}
