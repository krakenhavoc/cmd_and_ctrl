package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orzhov Locket — Artifact {3}:
//
//	"{T}: Add {W} or {B}.
//	 {W/B}{W/B}{W/B}{W/B}, {T}, Sacrifice this artifact: Draw two cards."
//
// Haunted Screen's {W|B} mana ability and Blitzball's tap-and-sacrifice
// draw; each hybrid symbol is paid with {W} or {B}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8aaf248e-66cf-447b-b062-9aa2ab1ec864",
		Name:         "Orzhov Locket",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "{T}: Add {W} or {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{W/B}{W/B}{W/B}{W/B}, {T}, Sacrifice this artifact: Draw two cards.",
			Cost:    Plus(ManaCost("{W/B}{W/B}{W/B}{W/B}"), TapCost(), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
