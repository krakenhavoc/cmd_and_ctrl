package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anchovy & Banana Pizza — {2}{B}{B} Artifact — Food (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"When this artifact enters, destroy target creature.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life."
//
// The enters trigger destroys its one target if it is still legal as the
// trigger resolves (CR 608.2b). The second line is the printed Food ability
// (CR 111.10b), the one every nontoken Food in the catalog shares.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f4834c7-cd63-420b-8492-ef127ffd4f45",
		Name:         "Anchovy & Banana Pizza",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Anchovy & Banana Pizza — destroy target creature", destroyTheFirstLegalTarget),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{samiFoodLifeAbility()},
	})
}
