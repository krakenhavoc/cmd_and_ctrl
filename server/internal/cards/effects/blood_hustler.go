package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Hustler — Creature — Vampire Rogue {1}{B}, 1/1 (Outlaws of
// Thunder Junction, #4468):
//
//	"Whenever you commit a crime, put a +1/+1 counter on this creature.
//	 This ability triggers only once each turn. (Targeting opponents,
//	 anything they control, and/or cards in their graveyards is a
//	 crime.)
//	 {3}{B}: Target opponent loses 1 life and you gain 1 life."
//
// The crime trigger is Magda, the Hoardmaster's: the once-per-announce
// CrimeCommitted annotation (CR 700.13, YouCommittedCrime), held to once
// each turn by the per-turn trigger tally (b11TriggeredThisTurn). The
// counter goes on this object only; one that left and came back is a new
// object and gets none (putACounterOnThis).
//
// The drain is Duskwielder's body (drainTargetOpponentOne). Targeting an
// opponent with it is itself a crime, so it can feed the trigger.
//
// No simplification.
const otjB1BloodHustlerLabel = "Blood Hustler — put a +1/+1 counter on this creature"

func init() {
	Register(Spec{
		OracleID:     "bb02e64c-0c4a-41ae-972e-ce83e6fc4cb9",
		Name:         "Blood Hustler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return YouCommittedCrime(ev, source) &&
					!b11TriggeredThisTurn(g, source.InstanceID, otjB1BloodHustlerLabel)
			}, otjB1BloodHustlerLabel, putACounterOnThis(game.CounterPlusOne)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}: Target opponent loses 1 life and you gain 1 life.",
			Cost:    ManaCost("{3}{B}"),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  drainTargetOpponentOne,
		}},
	})
}
