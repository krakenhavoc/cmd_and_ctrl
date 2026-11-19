package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b10_helpers.go — shared pieces for the S82 GAP batch 10 cards
// (Innistrad Remastered creatures, #5948).

// SelfEntersWithCountersIfMorbid is "Morbid — this creature enters with N
// <kind> counters on it if a creature died this turn" (Festerhide Boar):
// Cackling Slasher's self-replacement (CR 614.1c), whose condition is the
// table-wide tally of creatures that died this turn (any creature,
// anyone's, tokens included, read as the creature enters), with the
// counter count and kind left to the card.
func SelfEntersWithCountersIfMorbid(kind string, n int) game.ReplacementEffect {
	r := cacklingSlasherEntry()
	r.RowLabel = "Morbid — Enters with " + counterCountPhrase(n, kind) + " on it if a creature died this turn."
	r.Replace = func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
		ev.AddCounterAtETB(kind, n)
		return nil
	}
	return r
}
