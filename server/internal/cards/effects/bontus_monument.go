package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bontu's Monument — Legendary Artifact {3}:
//
//	"Black creature spells you cast cost {1} less to cast.
//	 Whenever you cast a creature spell, each opponent loses 1 life
//	 and you gain 1 life."
//
// The drain trigger fires on EVERY creature spell you cast, not just
// black ones — the cost clause and the trigger name different sets on
// purpose, exactly as printed. drainEachOpponent is the shared "each
// opponent loses 1, you gain 1" body (batch07_helpers.go), already
// used by Cauldron of Essence's dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "940ba435-7abc-40f8-a5af-c1c653b3284e",
		Name:         "Bontu's Monument",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Black creature spells you cast cost {1} less to cast.",
				YourSpell(), CreatureSpell(), ColoredSpell("B")),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Creature(),
				"Bontu's Monument — each opponent loses 1 life and you gain 1 life",
				drainEachOpponent),
		},
	})
}
