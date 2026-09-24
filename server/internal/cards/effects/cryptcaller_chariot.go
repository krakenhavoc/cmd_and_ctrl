package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cryptcaller Chariot — Artifact — Vehicle {3}{B}, 5/5 (EDHREC rank
// 4527):
//
//	"Menace
//	 Whenever you discard one or more cards, create that many tapped
//	 2/2 black Zombie creature tokens.
//	 Crew 2"
//
// A four-mana 5/5 with menace that pays a Zombie for every card a
// madness, cycling or looting deck throws away. The Vehicle shell is
// what makes it safe: it is not a creature until it is crewed, so a
// sweeper misses it and the Zombies it made can crew it afterwards.
//
// The tokens arrive TAPPED, which is the card's own drawback — they
// cannot crew the Chariot or block the turn they appear, only attack
// on a later one (or be sacrificed immediately, which is what the
// aristocrats deck actually does with them).
//
// Crew 2 is the standard crew cost: tap any number of untapped
// creatures you control with total power 2 or more, and the Chariot
// becomes an artifact creature until end of turn (CR 702.122).
//
// DECLARED SIMPLIFICATION (batching): the card reads "one
// or more cards … that many", one trigger for the whole batch. The
// engine emits one EventDiscardCard per card, so discarding three
// cards to a single Windfall fires three triggers of one Zombie each
// rather than one trigger of three. Same total Zombies, three log
// lines — and three separate response windows, which is the
// observable difference.
func init() {
	Register(Spec{
		OracleID:     "2bdd0bcb-6cfb-48e7-972e-355ff46621a4",
		Name:         "Cryptcaller Chariot",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Discarding several cards at once makes the Zombies one trigger at a time rather than all at once.",
		},
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Cryptcaller Chariot — create a tapped 2/2 black Zombie", func(g *game.Game, item *game.StackItem) error {
				tmpl := TokenCard("2/2 black Zombie")
				tmpl.Tapped = true
				return CreateToken{Controller: item.Controller, Template: tmpl, N: 1}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:  "Crew 2",
			Cost:   CrewCost(2),
			Effect: CrewEffect("Cryptcaller Chariot"),
		}},
	})
}
