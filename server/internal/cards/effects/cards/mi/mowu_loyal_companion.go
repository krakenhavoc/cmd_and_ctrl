package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mowu, Loyal Companion — Legendary Creature — Dog {3}{G}, 3/3
// (Foundations Jumpstart):
//
//	"Vigilance, trample
//	 If one or more +1/+1 counters would be put on Mowu, that many plus
//	 one +1/+1 counters are put on it instead."
//
// Hardened Scales narrowed to the source itself: the replacement applies
// only when the counters' recipient is this object (CR 614.1, CR 122.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "34ceb733-0329-4c1b-8d75-25c34fd4400f",
		Name:            "Mowu, Loyal Companion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCounter && ev.CounterName == "+1/+1" &&
					ev.CounterDelta > 0 && src != nil && ev.CounterTarget == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Mowu, Loyal Companion: +1 +1/+1 counter",
		}},
	})
}
