package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Engineered Explosives — Artifact {X}:
//
//	"Sunburst (This artifact enters with a charge counter on it for
//	 each color of mana spent to cast it.)
//	 {2}, Sacrifice this artifact: Destroy each nonland permanent with
//	 mana value equal to the number of charge counters on this
//	 artifact."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). X is part of the cost and does nothing else, so
// the spec does not declare XMatters: X=0 is a real play, an Explosives
// with no counters. The counters are the colours spent, so X=2 paid
// with {U}{R} is two
// counters and X=2 paid with two colourless is none. The sacrifice is a
// cost, so the number is the Explosives' last-known information
// (CR 608.2h). A token's mana value is 0, so Explosives with no counters
// destroys every token.
func init() {
	Register(Spec{
		OracleID:            "95fd897e-9086-42c4-8e0d-61bb02333c5f",
		Name:                "Engineered Explosives",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []ActivatedAbility{{
			Label: "{2}, Sacrifice this artifact: Destroy each nonland permanent with mana value equal to the number of charge counters on this artifact.",
			Cost:  Plus(ManaCost("{2}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok {
					return nil
				}
				return DestroyAllMatching{Match: And(Nonland(), b03ManaValueIs(info.Counters[game.CounterCharge]))}.Apply(ctx)
			},
		}},
	})
}
