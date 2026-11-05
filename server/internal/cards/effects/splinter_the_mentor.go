package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splinter, the Mentor — {1}{B} Legendary Creature — Mutant Ninja Rat 2/2
// (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Menace
//	 Whenever Splinter or another nontoken creature you control leaves
//	 the battlefield, create a Mutagen token.
//	 Partner—Character select (You can have two commanders if both have
//	 this ability.)"
//
// One trigger on any exit, not only dying: Splinter himself (token or
// not, since "nontoken" qualifies only "another"), or another creature
// that was not a token and that its controller controlled as it left.
// Both are read from the last-known information the exit records, since
// a leaves-the-battlefield ability looks back in time (CR 603.10a), so
// Splinter leaving together with other creatures sees each of them. The
// Mutagen token is the catalog's CR 111.10v template (MutagenToken).
//
// Partner—Character select is a deck-building rule (CR 702.124i) that
// internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fdcb016e-36e7-46b3-9dff-20b9b9cc421d",
		Name:            "Splinter, the Mentor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, splinterTheMentorLeft,
				"Splinter, the Mentor — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
	})
}

// splinterTheMentorLeft is "Splinter or another nontoken creature you
// control leaves the battlefield".
func splinterTheMentorLeft(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	if !aCreatureYouControlLeft(ev, source, lki, g) {
		return false
	}
	left, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !IsToken(left)
}
