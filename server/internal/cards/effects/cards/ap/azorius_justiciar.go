package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cba0eae3-73ea-4ea9-87bf-1015e490da90",
		Name:         "Azorius Justiciar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Azorius Justiciar — detain up to two target creatures your opponents control",
				effects.DetainEachLegalTarget("Azorius Justiciar — detain")),
				effects.TargetCreature("up to two target creatures your opponents control", effects.OpponentControls()).WithCount(0, 2)),
		},
	})
}
