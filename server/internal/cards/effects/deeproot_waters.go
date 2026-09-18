package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deeproot Waters — Enchantment for {2}{U} (EDHREC rank 4367):
//
//	"Whenever you cast a Merfolk spell, create a 1/1 blue Merfolk
//	 creature token with hexproof."
//
// The Merfolk deck's token engine, and the tokens are Merfolk too, so
// every lord on the board makes them real threats. Roadmap batch 42
// (#449) filed it under protection / hexproof (#95); hexproof has
// been a canonical keyword the targeting layer honours since S23, so
// the card is writable today — the token is an ordinary table entry
// with the keyword on it.
//
// Triggers on the CAST, not on resolution: a Merfolk that gets
// countered still leaves you the token. It is one of the reasons the
// enchantment plays as well as it does against a blue table.
//
// "A MERFOLK SPELL" is the subtype on the spell, read through the
// effective subtypes — so a changeling spell is a Merfolk spell, and
// a non-creature spell with the Merfolk type (there are none printed,
// but the rule does not care) would count too. Deeproot Waters
// itself is not a Merfolk and does not trigger on its own cast.
//
// HEXPROOF on the tokens is the half that matters in Commander: the
// board a Merfolk deck builds is immune to targeted removal, so only
// a sweeper answers it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e7b31eb-7a91-4992-b24a-d81173e1dbc8",
		Name:         "Deeproot Waters",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Subtype("Merfolk"),
				"Deeproot Waters — create a 1/1 blue Merfolk with hexproof",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   b42BlueMerfolkHexproofToken(),
						N:          1,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
