package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// U.S.S. Prometheus, Prototype — {5} Legendary Artifact — Spacecraft 6/6
// (Star Trek Commander, #4974):
//
//	"Station (… It's an artifact creature at 4+.)
//	 4+ | Flying
//	 8+ | Multivector Assault Mode — Whenever U.S.S. Prometheus deals
//	 combat damage to an opponent, it deals that much damage to each
//	 other opponent."
//
// A 6/6 flier from four charge counters (CR 721.2b); the trigger is in
// the 8+ striation, so it exists only from eight counters on (CR 721.2a).
// The trigger is Hydra Omnivore's: how much landed and on whom are read
// off the damage event, and every other opponent takes that much from
// the Prometheus in one damage instance. Multivector Assault Mode is an
// ability word with no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	t := whenThisHitsAnOpponentThatMuchToEachOther("U.S.S. Prometheus — Multivector Assault Mode: that much damage to each other opponent")
	t.ActiveWhen = AtChargeCounters(8)
	Register(Spec{
		OracleID:     "d19ebe9f-c24d-405c-b8e8-a602274076c3",
		Name:         "U.S.S. Prometheus, Prototype",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(4, 6, 6),
			ThresholdKeywords(4, "flying"),
		),
	})
}
