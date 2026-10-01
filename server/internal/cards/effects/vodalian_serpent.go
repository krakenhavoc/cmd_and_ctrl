package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vodalian Serpent — Creature — Serpent {3}{U}, 2/2:
//
//	"Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	 This creature can't attack unless defending player controls an
//	 Island.
//	 If this creature was kicked, it enters with four +1/+1 counters on
//	 it."
//
// The restriction is ADR 0107 §2's (#1879, CR 508.1c), with the defending
// player worked out per target (CR 508.5, 508.5a).
//
// The counters are a CR 614.1c entry clause read off the cast
// (CountersPerKick): kicker is paid at most once, so a kicked Serpent
// enters with four and an unkicked one with none. A Serpent put onto the
// battlefield any other way was not kicked and enters with none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "c39c1604-3bae-454d-9985-85101e51ec6e",
		Name:                       "Vodalian Serpent",
		Completeness:               CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{Kicker("{2}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 4)},
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
	})
}
