package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Racers' Scoreboard — Artifact {4}:
//
//	"Start your engines!
//	 When this artifact enters, draw two cards, then discard a card.
//	 Max speed — Spells you cast cost {1} less to cast."
//
// ADR 0138 (#2122). The discount is an ordinary board cost modifier
// that applies only while the Scoreboard's controller has max speed
// (CR 702.178a); generic only, so it never reduces a coloured pip
// (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "54c2b3b3-e247-4146-8045-d49480b795f2",
		Name:            "Racers' Scoreboard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Racers' Scoreboard — draw two cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return b16DrawThenDiscard(g, item, 2, 1)
				}),
		},
		CostModifiers: []game.CostModifier{
			MaxSpeedCostModifier(CostsLess(1, "Max speed — Spells you cast cost {1} less to cast.", YourSpell())),
		},
	})
}
