package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shi'ar Soldier — Creature — Shi'ar Soldier {1}{U}, 1/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 {U}, {T}: Return another target permanent you control to its owner's
//	 hand. Activate only during your turn."
//
// Kelpie Guide's "another target permanent you control" on a {T} ability,
// the turn gate is DuringYourTurn, and the bounce is the simultaneous one
// Repulsor Bots uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f20ec8e2-d6e4-4707-9b86-107dbc09ac63",
		Name:            "Shi'ar Soldier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:     "{U}, {T}: Return another target permanent you control to its owner's hand. Activate only during your turn.",
			Cost:      Plus(ManaCost("{U}"), TapCost()),
			Targets:   Another(TargetPermanent("another target permanent you control", YouControl())),
			Condition: DuringYourTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b35BounceChosenPermanents(NewContext(g, item))
			},
		}},
	})
}
