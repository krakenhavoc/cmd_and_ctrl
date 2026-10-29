package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent Phil Coulson — Legendary Creature — Human Spy Hero {1}{W}, 2/2
// (Marvel Super Heroes):
//
//	"Vigilance
//	 {T}: Put a +1/+1 counter on each other Hero you control."
//
// Vigilance is a printed keyword, so tapping for the ability is not an
// attack's cost. The ability is untargeted: it finds every other Hero you
// control as it resolves, snapshots the set, and gives each one its own
// counter placement. "Hero" is read through the source's text and the
// creature's effective subtypes, so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dae20bcd-dc24-4f9d-9400-a6f013b49f14",
		Name:            "Agent Phil Coulson",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on each other Hero you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    TapCost(),
			Effect:  mshB1CounterOnEachOtherYouControl(And(Creature(), Subtype("Hero"))),
		}},
	})
}
