package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glyph Keeper — {3}{U}{U} Creature — Sphinx 5/3 (Amonkhet, S82 old batch 247):
//
//	"Flying
//	 Whenever this creature becomes the target of a spell or ability for the
//	 first time each turn, counter that spell or ability.
//	 Embalm {5}{U}{U}"
//
// Kira, Great Glass-Spinner's trigger on the Keeper alone: it counts the
// triggers it has had this turn (CR 603.4's "first time"), and counters the
// stack item that targeted it. The embalmed token is a new object with its own
// count (CR 400.7).
//
// No simplification.
func init() {
	const label = "Glyph Keeper — counter that spell or ability"
	effects.Register(effects.Spec{
		OracleID:        "d22b7d4a-0ac7-4221-80fd-a47edad03b84",
		Name:            "Glyph Keeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Embalm("{5}{U}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenFirstTargetedEachTurnCounterIt(label),
		},
	})
}
