package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voracious Hatchling — Creature — Elemental {3}{W/B}, 6/6 (Eventide, #5590):
//
//	"Lifelink
//	 This creature enters with four -1/-1 counters on it.
//	 Whenever you cast a white spell, remove a -1/-1 counter from this
//	 creature.
//	 Whenever you cast a black spell, remove a -1/-1 counter from this
//	 creature."
//
// The four counters are an entry replacement (Moonshadow's shape), and each
// colour is its own trigger: a spell of both colours removes two. A
// creature with no -1/-1 counter left removes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eab5eb29-bed6-456b-8929-5b9b820aa5a3",
		Name:            "Voracious Hatchling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Replacements:    []game.ReplacementEffect{effects.EveB1HatchlingCounters("Voracious Hatchling")},
		Triggered: []game.TriggeredAbility{
			effects.EveB1RemoveCounterOnCast("Voracious Hatchling", "W"),
			effects.EveB1RemoveCounterOnCast("Voracious Hatchling", "B"),
		},
	})
}
