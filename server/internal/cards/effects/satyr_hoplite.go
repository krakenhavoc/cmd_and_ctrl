package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Satyr Hoplite — Creature — Satyr Soldier {R}, 1/1 (Foundations
// Jumpstart, S82 BNS):
//
//	"Heroic — Whenever you cast a spell that targets this creature, put a
//	 +1/+1 counter on this creature."
//
// Heroic (CR 207.2c) is the cast trigger of Favored Hoplite, read off the
// spell's stack item as it is cast; the counter goes on Satyr Hoplite only
// while it is still the same object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee7dc820-10ee-429a-a1ef-ecd6a11f5608",
		Name:         "Satyr Hoplite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			heroic("Satyr Hoplite — a +1/+1 counter on it (heroic)", thisGetsAPlusOneCounter),
		},
	})
}
