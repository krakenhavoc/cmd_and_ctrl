package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hopeful Initiate — Creature — Human Warlock {W}, 1/1 (EDHREC rank
// 6721):
//
//	"Training (Whenever this creature attacks with another creature
//	 with greater power, put a +1/+1 counter on this creature.)
//	 {2}{W}, Remove two +1/+1 counters from among creatures you
//	 control: Destroy target artifact or enchantment."
//
// The white one-drop that turns counters into removal. The activated
// ability is live since #789: "from among creatures you control" is a
// counter removal split across permanents, and the payment names a
// count per creature totalling exactly two — one each from two
// creatures, or both from one.
//
// The Initiate itself is a creature you control and a legal part of
// that payment, which is how the card is meant to be played once
// training has grown it.
//
// Training is the shared Training() trigger (training.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "317169b8-9014-48e6-862b-ca21b706846e",
		Name:         "Hopeful Initiate",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Training("Hopeful Initiate — training: put a +1/+1 counter on it")},
		Activated: []ActivatedAbility{{
			Label: "{2}{W}, Remove two +1/+1 counters from among creatures you control: Destroy target artifact or enchantment.",
			Cost: Plus(
				ManaCost("{2}{W}"),
				RemoveCountersAmong(game.CounterPlusOne, 2, "creatures you control", Creature()),
			),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
