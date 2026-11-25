package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tizerus Charger — {2}{B} Creature — Pegasus 3/2 (Theros Beyond Death,
// S82 OLD b031, #6154):
//
//	"Escape—{4}{B}, Exile five other cards from your graveyard.
//	 This creature escapes with your choice of a +1/+1 counter or a flying
//	 counter on it."
//
// Escape is the ordinary graveyard alternative cost (CR 702.138a). The
// counter choice is CR 702.138c's "if it escaped, it enters with" read
// through EscapesWithChoiceOfCounters: asked as the escaped creature
// arrives, and not at all when it is cast from hand or put onto the
// battlefield another way. A flying counter gives it flying (CR 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "8f94754d-fb2a-4634-a1de-049c82fa6f30",
		Name:             "Tizerus Charger",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{4}{B}", 5)},
		AsEnters: effects.EscapesWithChoiceOfCounters("Tizerus Charger",
			[]string{game.CounterPlusOne}, []string{game.CounterFlying}),
	})
}
