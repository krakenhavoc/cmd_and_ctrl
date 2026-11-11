package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Court of Garenbrig — Enchantment {1}{G}{G} (Wilds of Eldraine Commander,
// #5368):
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, distribute two +1/+1 counters among
//	 up to two target creatures. Then if you're the monarch, double the
//	 number of +1/+1 counters on each creature you control."
//
// The division is announced with the targets (CR 601.2d, 603.3d): zero, one
// or two creatures, each chosen creature at least one counter. The doubling
// reads the counts after the counters land, on every creature you control,
// as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f3f6336-6f99-4181-be3c-1814f7b8ef36",
		Name:         "Court of Garenbrig",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Garenbrig"),
			{
				Watches:   []game.EventKind{game.EventBeginUpkeep},
				AppliesTo: ByYou,
				Targets: TargetCreature("up to two target creatures").WithCount(0, 2).
					Dividing(Divide(2)),
				Key:    "Court of Garenbrig — distribute two +1/+1 counters; if you're the monarch, double the +1/+1 counters on each creature you control",
				Effect: cmdr2S48CourtOfGarenbrigUpkeep,
			},
		},
	})
}
