package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Snake — Creature — Snake {1}{G}{U}{U}, 2/2 (Apocalypse, #5614):
//
//	"Flash
//	 When this creature enters, counter target spell."
//
// Fear of Impostors' enters trigger without the dread: the spell is
// chosen as the trigger goes on the stack (CR 603.3d), and
// counterTheChosenSpell counters it if it is still a legal target as the
// trigger resolves (CR 608.2b). A spell that can't be countered stays on
// the stack. With no spell on the stack the trigger has no target and is
// removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "161b6d06-e6a7-488e-9f9c-d266c1402a3a",
		Name:            "Mystic Snake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Mystic Snake — counter target spell", effects.CounterTheChosenSpell),
				effects.TargetSpell("target spell"),
			),
		},
	})
}
