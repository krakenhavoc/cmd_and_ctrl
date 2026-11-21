package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "f718b520-d79d-4b91-9d85-de1f1c90c176",
		Name:            "Cloudbound Moogle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Cloudbound Moogle — put a +1/+1 counter on target creature", effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")),
		},
	})
}
