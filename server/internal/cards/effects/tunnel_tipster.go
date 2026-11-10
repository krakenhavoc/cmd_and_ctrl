package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tunnel Tipster — Creature — Mole Scout {1}{G}, 1/1 (Murders at Karlov
// Manor, #4469):
//
//	"At the beginning of your end step, if a face-down creature entered
//	 the battlefield under your control this turn, put a +1/+1 counter on
//	 this creature.
//	 {T}: Add {G}."
//
// An intervening if (CR 603.4), checked as the end step begins and again
// on resolution, read off the turn's face-down-entry tally: it counts a
// creature that has since turned face up or left (the ruling) and not one
// that entered face up and turned face down later.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d0bf685-e521-4d72-acc6-86794f685ce2",
		Name:         "Tunnel Tipster",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondFaceDownEnteredThisTurn{},
				On(game.EventBeginEndStep, AllOf(ByYou, YouHadAFaceDownEntryThisTurn),
					"Tunnel Tipster — put a +1/+1 counter on this creature",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
