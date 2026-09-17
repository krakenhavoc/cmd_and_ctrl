package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ureni of the Unwritten — Legendary Creature — Spirit Dragon
// {4}{G}{U}{R}, 7/7:
//
//	"Flying, trample
//	 Whenever Ureni enters or attacks, look at the top eight cards of
//	 your library. You may put a Dragon creature card from among them
//	 onto the battlefield. Put the rest on the bottom of your library in
//	 a random order."
//
// The Dragon commander that deploys Dragons. The shared
// LookAtTopThenMayPutOntoBattlefield sentence (#745): the eight cards
// are LOOKED at, so only Ureni's controller sees them — the
// choose_cards prompt is withheld from every other seat, count
// included — the pick enters through the CR 614 pipeline without
// being cast, and the rest go to the bottom in an order drawn from the
// game's seeded RNG.
//
// "Dragon creature card" reads the card's subtypes, so a changeling
// creature card qualifies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "99c2d3ef-e5b4-48cd-b3f5-de9b02c7c36a",
		Name:            "Ureni of the Unwritten",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Ureni of the Unwritten — look at the top eight cards; you may put a Dragon creature card onto the battlefield",
				LookAtTopThenMayPutOntoBattlefield(8, OfCreatureType("Dragon"), 1,
					"Ureni of the Unwritten — you may put a Dragon creature card from among them onto the battlefield")),
		},
	})
}
