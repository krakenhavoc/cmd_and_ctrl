package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kor Cartographer — Creature — Kor Scout {3}{W}, 2/2 (EDHREC rank
// 4104):
//
//	"When this creature enters, you may search your library for a
//	 Plains card, put it onto the battlefield tapped, then shuffle."
//
// White's Solemn Simulacrum without the death draw: a four-mana 2/2
// that ramps, which is the deal white has to take because it does
// not get Rampant Growth. The body matters more than it looks —
// white token decks are happy to have a creature attached to the
// land drop.
//
// "A PLAINS CARD", NOT "A BASIC PLAINS". The filter is the Plains
// SUBTYPE on the printed type line, so it finds Hallowed Fountain,
// Sacred Foundry, a Snow-Covered Plains and Glacial Floodplain as
// readily as the basic — which is a real upgrade over Solemn in a
// deck with shocklands, and it is what the card prints.
//
// "YOU MAY SEARCH" IS THE PROMPT, NOT AN ASSUMPTION: SearchLibrary's
// Optional forces the prompt even when the pick looks free, so the
// searcher can decline the land AND the shuffle. TappedOnEntry is the
// card's own "tapped" clause, OR-ed with whatever the fetched land's
// own entry replacement says — a Glacial Floodplain fetched by the
// Cartographer enters tapped for two independent reasons and that is
// still just tapped.
//
// Sandbox note, not a caveat on this card: which Plains the search
// picks when the prompt is answered by a bot is the catalog-wide
// deterministic search pick, the same for every tutor in the
// catalog.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5f537f6-3254-45a3-a4e7-d4d4fd7a972d",
		Name:         "Kor Cartographer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kor Cartographer — search for a Plains card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     b39IsPlainsCard,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					TappedOnEntry: true,
					Optional:      true,
					Shuffle:       true,
					Source:        item.SourceCardID,
					Reason:        "Kor Cartographer — search for a Plains card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
