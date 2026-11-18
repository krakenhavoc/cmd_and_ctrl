package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gavony Silversmith — Creature — Human Soldier {3}{W}, 2/3 (Innistrad: Midnight Hunt, S82):
//
//	"When this creature enters, put a +1/+1 counter on each of up to two target creatures."
//
// Byrke, Long Ear of the Law's entry: one counter on each target still legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0ee4faa-1ab0-4b4f-a564-ee2358500daf",
		Name:         "Gavony Silversmith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Gavony Silversmith — put a +1/+1 counter on each of up to two target creatures",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("up to two target creatures").WithCount(0, 2)),
		},
	})
}
