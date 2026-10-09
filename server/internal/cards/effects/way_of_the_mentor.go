package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Way of the Mentor — Legendary Enchantment {2}{W} (Reality Fracture,
// tracker #2795):
//
//	"When Way of the Mentor enters, empower Jace 5.
//	 Whenever you gain life, put a loyalty counter on each planeswalker
//	 you control."
//
// Empower Jace is the keyword action (ADR 0139). The lifegain trigger is
// one per life-gain event, as printed, and puts a counter on each
// planeswalker you control as it resolves; the counters are an effect's,
// so each walker is a placement that loyalty-counter watchers see.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce924285-f4c5-44e6-909b-f5f7d302ebaf",
		Name:         "Way of the Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Way of the Mentor — empower Jace 5", Do(EmpowerJace{N: 5})),
			WheneverYouGainLife("Way of the Mentor — a loyalty counter on each planeswalker you control",
				frPutLoyaltyCounterOnEachPlaneswalkerYouControl),
		},
	})
}
