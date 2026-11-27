package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Matopi Golem — Artifact Creature — Golem {5}, 3/3 (Visions, #5631):
//
//	"{1}: Regenerate this creature. When it regenerates this way, put a
//	 -1/-1 counter on it."
//
// The ability makes a regeneration shield marked as the one that carries
// the rider (RegenerateThisWithRider, #5631). When that shield replaces a
// destruction the engine says so on the EventRegenerated it emits
// (Event.RegenerationRider), and the Golem's trigger puts a -1/-1 counter on
// it, on the stack above whatever follows. A regeneration from any other
// source never sets it, and a plain shield is spent before a rider one
// (the controller's choice among replacements, CR 616.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0308b9f2-11da-41d0-8d18-18e8c0c600ab",
		Name:         "Matopi Golem",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Regenerate this creature. When it regenerates this way, put a -1/-1 counter on it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}"),
			Effect:  effects.RegenerateThisWithRider,
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisRegeneratesThisWay("Matopi Golem — put a -1/-1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterMinusOne, N: 1})),
		},
	})
}
