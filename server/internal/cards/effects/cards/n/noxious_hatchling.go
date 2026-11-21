package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noxious Hatchling — Creature — Elemental {3}{B/G}, 6/6 (Eventide, #5590):
//
//	"This creature enters with four -1/-1 counters on it.
//	 Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	 Whenever you cast a black spell, remove a -1/-1 counter from this
//	 creature.
//	 Whenever you cast a green spell, remove a -1/-1 counter from this
//	 creature."
//
// The four counters are an entry replacement (Moonshadow's shape), and each
// colour is its own trigger: a spell of both colours removes two. A
// creature with no -1/-1 counter left removes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "36072943-7823-434d-b898-f0a2c4e00af2",
		Name:            "Noxious Hatchling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither},
		Replacements:    []game.ReplacementEffect{effects.EveB1HatchlingCounters("Noxious Hatchling")},
		Triggered: []game.TriggeredAbility{
			effects.EveB1RemoveCounterOnCast("Noxious Hatchling", "B"),
			effects.EveB1RemoveCounterOnCast("Noxious Hatchling", "G"),
		},
	})
}
