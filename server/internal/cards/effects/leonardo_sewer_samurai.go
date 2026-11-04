package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonardo, Sewer Samurai — Legendary Creature — Mutant Ninja Turtle
// Samurai {3}{W}, 3/3 (TMT, #4457, #4964):
//
//	"Sneak {2}{W}{W}
//	 Double strike
//	 During your turn, you may cast creature spells with power or
//	 toughness 1 or less from your graveyard. If you cast a spell this
//	 way, that creature enters with a finality counter on it."
//
// Sneak is the keyword's alternative cost (CR 702.190, Sneak) and double
// strike a printed keyword. The cast is a STANDING graveyard permission
// (ADR 0066) derived off the battlefield, as Gisa and Geralf's is, but
// with no budget: any number of casts on your turn. TimingYourTurnOnly
// closes it off your turn and leaves each creature's own timing in force.
// "Creature spells with power or toughness 1 or less" is the filter's
// power-or-toughness clause (#4964), judged on the face being cast and
// with a `*` card's own characteristic-defining ability counted where
// the card is (CR 604.3), so a 2/1 and a 1/2 qualify, a 2/2 doesn't, and
// a Tarmogoyf qualifies only while its ability makes it one. "That
// creature enters with a finality counter" is the permission's rider
// (CastPermission.EntersWithCounterKind, #4679), put on through the
// CR 614 pipeline as the creature enters; CR 122.1h's finality counter
// then exiles it instead of letting it die. A creature cast from your
// graveyard by its own text (flashback, escape) or by another permission
// is not cast "this way" and gets no counter.
//
// No simplification.
func init() {
	const name = "Leonardo, Sewer Samurai"
	Register(Spec{
		OracleID:         "74e3243b-6707-4769-8767-7cf2ef71009a",
		Name:             name,
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"double strike"},
		AlternativeCosts: []game.AlternativeCost{Sneak("{2}{W}{W}")},
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneGraveyard,
			CastOnly: true,
			Filter: game.PermissionFilter{
				CreatureOnly:           true,
				PowerOrToughnessCapped: true,
				PowerOrToughnessAtMost: 1,
			},
			Timing:                game.TimingYourTurnOnly,
			EntersWithCounterKind: game.CounterFinality,
			EntersWithCounters:    1,
			Label:                 "During your turn, cast a creature spell with power or toughness 1 or less from your graveyard; it enters with a finality counter (" + name + ")",
		}},
	})
}
