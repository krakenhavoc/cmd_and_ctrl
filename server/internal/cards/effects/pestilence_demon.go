package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pestilence Demon — Creature — Demon {5}{B}{B}{B}, 7/6 (Rise of the
// Eldrazi, #5581):
//
//	"Flying
//	 {B}: This creature deals 1 damage to each creature and each player."
//
// Pestilence's activated ability on a creature, without Pestilence's
// end-step sacrifice. The Demon damages itself too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2572add6-c0b1-44bd-8713-1e94e3fe69d6",
		Name:            "Pestilence Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{B}: This creature deals 1 damage to each creature and each player.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
			Cost:    ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b23DamageEachCreatureAndEachPlayer(NewContext(g, item), 1)
			},
		}},
	})
}
