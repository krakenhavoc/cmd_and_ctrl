package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonsilver Key — Artifact {2} (EDHREC rank 1666):
//
//	"{1}, {T}, Sacrifice this artifact: Search your library for an
//	 artifact card with a mana ability or a basic land card, reveal
//	 it, put it into your hand, then shuffle."
//
// Expedition Map for mana rocks: a two-mana artifact that becomes
// the Sol Ring or the basic you needed. Same three-component cost,
// same search to hand; the clause is
// b15ArtifactWithManaAbilityOrBasicLand. The searcher chooses (the
// S22 chooser), so with more than one match the prompt opens.
//
// "An artifact card with a mana ability" is about what the card prints
// (#3943). The deck importer reads every card's printed text once for a
// mana ability as CR 605.1a/b defines one (game.ManaAbilityInText) and
// stamps the answer on the card, so an artifact the catalog does not
// run is found when its text has one (Sol Ring, Mind Stone), and one
// whose only mana is not a mana ability is not: Su-Chi's dies trigger
// (CR 605.5a), Chromatic Sphere's draw and Millikin's mill (each moves
// a card from a library, CR 605.1a). Scryfall's produced_mana would
// find all three. The catalog answers only for a card whose text was
// never read.
func init() {
	Register(Spec{
		OracleID:     "373395b2-f04b-46c9-b9f4-4f6815b71009",
		Name:         "Moonsilver Key",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Search your library for an artifact card with a mana ability or a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: b15ArtifactWithManaAbilityOrBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Moonsilver Key — an artifact with a mana ability or a basic land, to your hand",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
