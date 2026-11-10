package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fang-Druid Summoner — Creature — Ape Druid {3}{G}, 2/4 (Aetherdrift,
// #4464):
//
//	"Reach
//	 When this creature enters, you may search your library and/or
//	 graveyard for a creature card with no abilities, reveal it, and put
//	 it into your hand. If you search your library this way, shuffle."
//
// Captain Kirk's library-and/or-graveyard search (#3217): the player says
// which zones they search before anything is shown, may find nothing, and
// the library is shuffled if and only if it was searched (CR 701.23). "A
// creature card with no abilities" reads the card's printed text through
// game.PrintsNoAbilities (Card.NoAbilitiesText, stamped by the deck
// importer): a vanilla creature, so a card with only a keyword is not one.
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Fang-Druid Summoner — you may search your library and/or graveyard for a creature card with no abilities",
		func(g *game.Game, item *game.StackItem) error {
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     b13CreatureCardWithNoAbilities,
				Dest:          game.ZoneHand,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				Optional:      true,
				AlsoGraveyard: true,
				Reason:        "Fang-Druid Summoner — a creature card with no abilities to put into your hand",
			}.Apply(NewContext(g, item))
		})
	trigger.Purpose = game.Purpose{Tutors: 1}
	Register(Spec{
		OracleID:        "1a9bfe76-6c74-4393-b04e-461422a29ddb",
		Name:            "Fang-Druid Summoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
