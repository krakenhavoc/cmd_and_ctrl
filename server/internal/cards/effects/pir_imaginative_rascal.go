package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pir, Imaginative Rascal — 1/1 Human for {2}{G}:
//
//	"Partner with Toothy, Imaginary Friend
//	 If one or more counters would be put on a permanent your team
//	 controls, that many plus one of each of those kinds of counters
//	 are put on that permanent instead."
//
// The +1 half is Hardened Scales widened from "a creature you
// control" to "a permanent your team controls", which matters: Pir
// adds a loyalty counter to your planeswalker's +1 and a defense
// counter to your battle, and Hardened Scales does neither.
//
// "Your team" is read as "you". Two-Headed Giant is not a format this
// engine has, so a team is one player and the two readings coincide
// at every table it will ever see. Stated rather than silently
// assumed, because the day a team format arrives this is where the
// assumption is.
//
// Partner with is CR 702.124j's two abilities: the commander pairing
// with Toothy (internal/deck) and the entry search for a card named
// Toothy (PartnerWith). Until #2142 neither existed and Pir shipped
// with a caveat for both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7683c2b2-a06f-4691-9cc5-1968dc032885",
		Name:         "Pir, Imaginative Rascal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Pir, Imaginative Rascal", "Toothy, Imaginary Friend"),
		},
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					// Placement only, and on any PERMANENT you
					// control — not just a creature, which is the
					// whole difference from Hardened Scales, and not
					// a PLAYER, which is the difference from
					// Vorinclex. Pir names no placer and asks nothing
					// about one: "if one or more counters WOULD BE
					// PUT" is passive, like Hardened Scales and
					// Winding Constrictor, so an opponent's effect
					// putting counters on your permanent still gets
					// the +1.
					return counterPlacementOn(ev, g, src.Controller)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.CounterDelta++
					return nil
				},
				Controller: vorinclexController,
				Label:      "Pir: one more counter",
			},
		},
	})
}
