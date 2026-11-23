package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "dae20bcd-dc24-4f9d-9400-a6f013b49f14",
		Name:            "Agent Phil Coulson",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on each other Hero you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapCost(),
			Effect:  effects.MshB1CounterOnEachOtherYouControl(effects.And(effects.Creature(), effects.Subtype("Hero"))),
		}},
	})
}
