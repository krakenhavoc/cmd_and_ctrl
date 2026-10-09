package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Way of the Necromancer — Legendary Enchantment {1}{B} (Reality
// Fracture, tracker #2795):
//
//	"When Way of the Necromancer enters, empower Jace 2.
//	 Whenever a creature you control dies, put a loyalty counter on each
//	 planeswalker you control."
//
// Empower Jace is the keyword action (ADR 0139). The dies trigger is
// one per creature, tokens included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3087daa-d2ff-4227-abff-3398da6297af",
		Name:         "Way of the Necromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Way of the Necromancer — empower Jace 2", Do(EmpowerJace{N: 2})),
			WheneverACreatureYouControlDies("Way of the Necromancer — a loyalty counter on each planeswalker you control",
				frPutLoyaltyCounterOnEachPlaneswalkerYouControl),
		},
	})
}
