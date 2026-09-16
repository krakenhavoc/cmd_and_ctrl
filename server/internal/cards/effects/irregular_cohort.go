package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Irregular Cohort — Creature — Shapeshifter, {2}{W}{W}, 2/2:
//
//	"Changeling (This card is every creature type.)
//	 When this creature enters, create a 2/2 colorless Shapeshifter
//	 creature token with changeling."
//
// Two bodies of every creature type for four mana, which is why it is
// the changeling card worth a catalog entry: the VANILLA changelings
// (Woodland Changeling, Universal Automaton) need none at all. Since
// #330 the deck importer stamps Scryfall's keyword array onto every
// card, and S26 made "changeling" one of the keywords the engine
// honours, so a vanilla changeling is every creature type with no
// catalog Spec in sight — and stops being reported as unimplemented.
// There is a test pinning exactly that, because "the card works
// without a file" is the sort of claim that rots quietly.
//
// The token carries the keyword on its template rather than through
// the catalog, for the reason every token ability does: a token has no
// oracle ID for a catalog hook to key on.
func init() {
	Register(Spec{
		OracleID:        "c0636d16-671c-4e80-af8c-67d80d2cd979",
		Name:            "Irregular Cohort",
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Irregular Cohort — create a Shapeshifter", Do(CreateToken{
				Template: ColorlessShapeshifterToken(),
				N:        1,
			})),
		},
	})
}
