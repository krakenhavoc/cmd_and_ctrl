package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warmonger Hellkite — Creature — Dragon {4}{R}{R}, 5/5 (Commander 2014,
// #5553):
//
//	"Flying
//	 All creatures attack each combat if able.
//	 {1}{R}: Attacking creatures get +1/+0 until end of turn."
//
// The attack requirement is Grand Melee's (AttacksEachCombatWhere over
// every creature, the Hellkite's own and its controller's included); it is
// a requirement of the source, so CR 508.1d's "if able" lets a tapped or
// summoning-sick creature stay home. The pump is Contested War Zone's: every
// attacking creature, whoever controls it, as the ability resolves, so a
// creature that attacks later in the turn is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "746d4445-f28d-41af-827a-2b613b9b84ee",
		Name:            "Warmonger Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			AttacksEachCombatWhere(func(_ *game.Card, _ *game.Game, _ *game.Card) bool { return true }),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: Attacking creatures get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Match: AttackingCreature(), Power: 1, Label: "Warmonger Hellkite — attacking creatures get +1/+0"}.Apply(NewContext(g, item))
			},
		}},
	})
}
