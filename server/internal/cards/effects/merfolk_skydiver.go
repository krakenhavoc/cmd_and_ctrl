package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Skydiver — Creature — Merfolk Mutant {G}{U}, 1/1 (Lost Caverns of
// Ixalan Commander, #5368):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on target creature you
//	 control.
//	 {3}{G}{U}: Proliferate."
//
// The enters trigger targets, so it is dropped with no legal target
// (CR 603.3d); the Skydiver itself is a legal target. Proliferate is the
// shared primitive (Contagion Clasp).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d2ccaa4-4815-4987-889f-d08c4d1e7a8e",
		Name:            "Merfolk Skydiver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Merfolk Skydiver — put a +1/+1 counter on target creature you control", PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}{U}: Proliferate",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{G}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Proliferate{}.Apply(NewContext(g, item))
			},
		}},
	})
}
