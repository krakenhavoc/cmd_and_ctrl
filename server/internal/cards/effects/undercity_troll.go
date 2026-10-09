package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undercity Troll — Creature — Troll {1}{G}, 2/2:
//
//	"Renown 1 (When this creature deals combat damage to a player, if
//	 it isn't renowned, put a +1/+1 counter on it and it becomes
//	 renowned.)
//	 {2}{G}: Regenerate this creature."
//
// #2049: renown is the engine's keyword trigger (game/renown.go);
// regeneration is Regenerate on the ability's own source (CR 701.19),
// Albino Troll's ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b7851b17-faa2-4767-9716-86997b882690",
		Name:            "Undercity Troll",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"renown 1"},
		Activated: []ActivatedAbility{{
			Label: "{2}{G}: Regenerate this creature.",
			Cost:  ManaCost("{2}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
