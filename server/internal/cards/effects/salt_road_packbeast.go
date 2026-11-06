package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Salt Road Packbeast — Creature — Beast {5}{W}, 4/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Affinity for creatures (This spell costs {1} less to cast for each
//	 creature you control.)
//	 When this creature enters, draw a card."
//
// Affinity is AffinityFor (CR 702.41a), Leonardo, Worldly Warrior's row:
// it takes only generic mana off, so the {W} is always paid. The draw is
// an ordinary enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50261b9d-48f2-4bad-a35b-14f13b3cdd32",
		Name:         "Salt Road Packbeast",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for creatures", Creature()),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Salt Road Packbeast — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
