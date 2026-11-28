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
// The ability makes a regeneration shield that carries the rider
// (RegenerateThisWithRider, on the Skeleton Scavengers seam). When that
// shield replaces a destruction the engine announces the regeneration with
// Amount 1, and the Golem's trigger puts a -1/-1 counter on it, on the stack
// above whatever follows. A regeneration from any other source never does.
//
// Simplification: with a shield from another source up as well, the engine
// spends the rider shield first (the order Skeleton Scavengers wants), so the
// controller cannot choose the plain one (CR 616.1).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0308b9f2-11da-41d0-8d18-18e8c0c600ab",
		Name:         "Matopi Golem",
		Completeness: effects.CompletenessCaveats,
		Caveats:      []string{"If it also has a regeneration shield from another source, it still gets the -1/-1 counter when it regenerates."},
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
