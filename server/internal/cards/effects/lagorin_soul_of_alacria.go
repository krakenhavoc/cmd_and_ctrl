package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lagorin, Soul of Alacria — Legendary Creature — Beast Mount {G}{W}:
//
//	"Flying
//	 Whenever Lagorin attacks while saddled, put a +1/+1 counter on each
//	 of up to two target Mounts and/or Vehicles.
//	 Saddle 1"
//
// A multi-target clause: the engine drops a target that left the
// battlefield before resolution (CR 608.2b), and the counters go on
// each one still legal.
//
// No simplification.
func init() {
	attack := AttacksWhileSaddled("Lagorin, Soul of Alacria — +1/+1 counter on each of up to two target Mounts and/or Vehicles", putPlusOneCounterOnEachLegalTarget)
	attack.Targets = TargetPermanent("up to two target Mounts and/or Vehicles",
		Or(OfSubtype("Mount"), OfSubtype("Vehicle"))).WithCount(0, 2)
	Register(Spec{
		OracleID:        "f7b6f5d4-b122-4ab4-a2c2-449ac781359a",
		Name:            "Lagorin, Soul of Alacria",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{Saddle(1)},
		Triggered:       []game.TriggeredAbility{attack},
	})
}
