package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Roustabout — Creature — Skeleton Warrior {1}{B}, 1/1:
//
//	"Unleash (You may have this creature enter with a +1/+1 counter on
//	 it. It can't block as long as it has a +1/+1 counter on it.)
//	 {1}{B}: Regenerate this creature."
//
// Unleash is the engine's keyword (ADR 0109 §10). Regeneration is
// Regenerate on the ability's own source (CR 701.19).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1adcc4a8-b2dd-4d9a-bded-188b81c84a10",
		Name:            "Grim Roustabout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUnleash},
		Activated: []ActivatedAbility{{
			Label: "{1}{B}: Regenerate this creature.",
			Cost:  ManaCost("{1}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
