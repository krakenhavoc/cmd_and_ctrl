package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// devour.go — Devour N (CR 702.82a): "As this creature enters, you may
// sacrifice any number of creatures. This creature enters with N
// +1/+1 counters on it for each creature sacrificed this way."
//
// A CR 614.12 as-enters replacement with a choice inside it, so it is
// the sacrifice EntryCardChoice (ADR 0098) in its "any number" form
// (AnyNumber). Four properties follow from that and are the point:
//
//   - The sacrifice happens as the creature enters, not off a trigger:
//     SacrificeAllThenForEffect runs before the creature is on the
//     battlefield, so EventSacrifice and the dies triggers of the
//     devoured creatures fire, and nothing can respond in between.
//   - The counters ride the entry event (ev.AddCounterAtETB), so
//     Doubling Season and Hardened Scales apply (CR 614.1c).
//   - Zero is always legal: an empty answer is the decline, and
//     the creature enters with no counters.
//   - The number is recorded on the permanent (Card.Devoured) for CR
//     702.82b's "each creature it devoured".
//
// The entering creature is never a candidate (CR 614.13a, handled by
// entryChoiceCandidatesLocked).

// Devour is "Devour N" over creatures.
func Devour(name string, n int) game.ReplacementEffect {
	return devourOver(name, n, "creatures", func(c game.Card) bool { return c.IsCreature() })
}

// devourOver is the shared body: sacrifice any number of permanents
// matching `matches` ("Devour land 3", "Devour Food 3").
func devourOver(name string, n int, plural string, matches func(game.Card) bool) game.ReplacementEffect {
	r := entersOnlyIf(name,
		fmt.Sprintf("%s — devour %d: you may sacrifice any number of %s as it enters.", name, n, plural),
		&game.EntryCardChoice{
			Action:    game.EntryCardSacrifice,
			Matches:   matches,
			AnyNumber: true,
			Devour:    n,
		})
	// Declining is not a graveyard trip: the creature enters either way.
	r.Replace = nil
	return r
}

// DevourPaying is Devour with the creature's own "for each creature it
// devoured" ability declared as data: cards drawn and life gained per
// creature. The trigger still reads Card.Devoured; this is only what
// the prompt tells a bot (#2419).
func DevourPaying(name string, n, drawPer, lifePer int) game.ReplacementEffect {
	r := Devour(name, n)
	r.EntryCardChoice.DevourDraw = drawPer
	r.EntryCardChoice.DevourLife = lifePer
	return r
}
