package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gnarlid Colony — Creature — Beast {1}{G}, 2/2 (Foundations, #4465):
//
//	"Kicker {2}{G} (You may pay an additional {2}{G} as you cast this
//	 spell.)
//	 If this creature was kicked, it enters with two +1/+1 counters on it.
//	 Each creature you control with a +1/+1 counter on it has trample."
//
// The kicker is the optional cost, and the counters are read off the
// cast (Vodalian Serpent's CountersPerKick: a single kicker is paid at
// most once). The trample grant is Pridemalkin's layer 6 grant over each
// creature you control with a +1/+1 counter on it, the Colony included,
// so it gains and loses trample with its counters like any other.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "e5481431-c952-4ad9-94fe-355076ef632b",
		Name:                       "Gnarlid Colony",
		Completeness:               CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{Kicker("{2}{G}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 2)},
		Static:                     []game.StaticAbility{b16GrantKeywords(b42CreatureYouControlWithAPlusOneCounter, "trample")},
	})
}
