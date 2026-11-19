package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Justiciar — Creature — Human Wizard {2}{W}{W}, 2/2 (Ravnica:
// Clue Edition):
//
//	"When this creature enters, detain up to two target creatures your
//	 opponents control. (Until your next turn, those creatures can't attack
//	 or block and their activated abilities can't be activated.)"
//
// Azorius Arrester's detain over up to two targets: each target still
// legal at resolution is detained until the controller's next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cba0eae3-73ea-4ea9-87bf-1015e490da90",
		Name:         "Azorius Justiciar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Azorius Justiciar — detain up to two target creatures your opponents control",
				detainEachLegalTarget("Azorius Justiciar — detain")),
				TargetCreature("up to two target creatures your opponents control", OpponentControls()).WithCount(0, 2)),
		},
	})
}
