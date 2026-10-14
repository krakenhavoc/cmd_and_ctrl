package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// compleated.go — compleated (CR 702.150a, ADR 0183, #2946):
//
//	"If this permanent would enter with one or more loyalty counters on
//	 it and the player who cast it chose to pay life for any part of its
//	 cost represented by Phyrexian mana symbols, it instead enters the
//	 battlefield with that many loyalty counters minus two for each of
//	 those mana symbols."
//
// The permanent's own replacement on its starting-loyalty placement
// (game.ReplacementEvent.StartingLoyalty, the CR 306.5b counters the
// engine puts on as it enters). It reads how many printed Phyrexian
// symbols its cast paid with life (Card.PhyrexianLifePaid, carried
// from PaidCost.PhyrexianLife), so life paid for anything else does not
// count, and a planeswalker that was not cast enters with its printed
// loyalty. Being an ordinary replacement, it is ordered against a
// Doubling Season by the planeswalker's controller (CR 616.1e; the
// ruling: "other replacement effects ... apply as normal").

// compleatedLifePerSymbol is CR 702.150a's "minus two for each".
const compleatedLifePerSymbol = 2

// Compleated is the keyword's replacement, declared on the card's
// Replacements. Add "compleated" to PrintedKeywords beside it.
func Compleated() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventCounterPlaced},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventCounter || !ev.StartingLoyalty ||
				ev.CounterTarget != src.InstanceID || ev.CounterName != game.CounterLoyalty || ev.CounterDelta <= 0 {
				return false
			}
			return src.PhyrexianLifePaid() > 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
			ev.CounterDelta -= compleatedLifePerSymbol * src.PhyrexianLifePaid()
			if ev.CounterDelta <= 0 {
				// It enters with no loyalty counters and dies to
				// CR 704.5i, as printed.
				ev.CounterDelta = 0
				ev.Canceled = true
			}
			return nil
		},
		Label:    "Compleated: two fewer loyalty counters for each Phyrexian symbol paid with life",
		RowLabel: "Compleated: enters with two fewer loyalty counters for each Phyrexian symbol paid with life.",
	}
}
