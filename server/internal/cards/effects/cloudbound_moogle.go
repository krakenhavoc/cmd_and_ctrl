package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloudbound Moogle — Creature — Moogle {3}{W}{W}, 2/3 (Final Fantasy,
// #4462):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on target creature.
//	 Plainscycling {2} ({2}, Discard this card: Search your library for a
//	 Plains card, reveal it, put it into your hand, then shuffle.)"
//
// Dominion Supervisor's enters trigger on a flier: any creature, the
// Moogle itself included, and nothing if the target is gone (CR 608.2b).
// Plainscycling is Angel of the Ruins' Typecycling with "a Plains card":
// any land with the Plains type, not only a basic one (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f718b520-d79d-4b91-9d85-de1f1c90c176",
		Name:            "Cloudbound Moogle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Cloudbound Moogle — put a +1/+1 counter on target creature", PlusOneCounterOnChosenTarget),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
		},
	})
}
