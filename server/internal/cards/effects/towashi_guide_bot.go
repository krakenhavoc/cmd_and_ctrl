package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Towashi Guide-Bot — Artifact Creature — Construct {4}, 2/1 (Kamigawa: Neon
// Dynasty):
//
//	"When this creature enters, put a +1/+1 counter on target creature you
//	 control.
//	 {4}, {T}: Draw a card. This ability costs {1} less to activate for each
//	 modified creature you control. (Equipment, Auras you control, and
//	 counters are modifications.)"
//
// The reduction is the ability's own cost modifier, counting live as the
// cost is determined (CR 602.2b). The ETB may target Guide-Bot itself.
//
// No simplification.
func init() {
	counter := Targeting(
		On(game.EventETB, Self, "Towashi Guide-Bot — put a +1/+1 counter on target creature you control",
			PlusOneCounterOnTheTarget),
		TargetCreature("target creature you control", YouControl()))
	Register(Spec{
		OracleID:     "28bb9a58-f45a-4452-a1f4-c7609373085f",
		Name:         "Towashi Guide-Bot",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{counter},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Draw a card. This ability costs {1} less to activate for each modified creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			CostModifiers: []game.CostModifier{
				ActivationCostsLessEach(neoB10ModifiedCount,
					"This ability costs {1} less to activate for each modified creature you control."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
