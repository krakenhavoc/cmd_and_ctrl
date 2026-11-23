package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Latchkey Faerie — Creature — Faerie Rogue {3}{U}, 3/1 (Morningtide,
// #5592):
//
//	"Flying
//	 Prowl {2}{U} (You may cast this for its prowl cost if you dealt combat
//	 damage to a player this turn with a Faerie or Rogue.)
//	 When this creature enters, if its prowl cost was paid, draw a card."
//
// Prowl is Enigma Thief's gated alternative cost (prowl_keyword.go,
// CR 702.76a). "If its prowl cost was paid" is an intervening if
// (CR 603.4) over the permanent's cast provenance (CR 400.7d), Spiders-Man,
// Heroic Horde's shape: a hard-cast Faerie, or one put onto the
// battlefield without being cast, draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "7bc4a0a0-f06e-4074-95a8-67a5d8fdfef9",
		Name:             "Latchkey Faerie",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Prowl("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.MorCastWithProwl),
				"Latchkey Faerie — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
