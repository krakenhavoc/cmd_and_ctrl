package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hare Apparent — 2/2 Creature — Rabbit Noble for {1}{W} (EDHREC rank
// 3993):
//
//	"When this creature enters, create a number of 1/1 white Rabbit
//	 creature tokens equal to the number of other creatures you
//	 control named Hare Apparent.
//	 A deck can have any number of cards named Hare Apparent."
//
// The Relentless Rats of white: each copy is a two-mana 2/2 that
// brings along one Rabbit for every earlier copy already on the
// board, so the fourth one lands with three friends and the deck
// snowballs. It is in the batch as the first card in the catalog
// whose ETB counts permanents BY NAME rather than by type.
//
// # What the count reads
//
// Names, post-layer, so a Clone copying a Hare Apparent counts and a
// Rabbit token does not (the tokens are named "Rabbit"). "Other"
// excludes the copy that just entered — it is already on the
// battlefield when its own trigger resolves, and counting itself
// would make the first copy bring a Rabbit. "You control" is the
// controller only; an opponent running the same deck contributes
// nothing.
//
// The count is taken at RESOLUTION, not when the trigger goes on the
// stack, so a Hare Apparent killed in response lowers the total.
//
// # Declared simplification: the deckbuilding line is not enforced
//
// "A deck can have any number of cards named Hare Apparent" is a
// deckbuilding permission (CR 903.5b's exception), and the deck
// validator applies the Commander singleton rule to every nonbasic
// card with no exception list. So a legal paper deck with thirty
// Hare Apparents is rejected at import, and the card is playable only
// as a singleton — where its trigger makes nothing unless a copy
// effect is involved.
//
// That is strictly WEAKER than printed, which is the direction #259
// allows: nothing here lets a deck do something paper would not. The
// trigger itself is complete and will start mattering the moment the
// validator learns the exception (it is one predicate over the oracle
// text, shared with Relentless Rats, Dragon's Approach, Persistent
// Petitioners, Shadowborn Apostle, Seven Dwarves and the Nazgûl).
func init() {
	Register(Spec{
		OracleID:     "3c1619bd-db5e-4df6-a196-0a9d62374f6d",
		Name:         "Hare Apparent",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Deckbuilding still treats it as a singleton, so you can't run the many copies the card allows — with one copy the enters trigger makes no Rabbits.",
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hare Apparent — create a Rabbit for each other Hare Apparent you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := b38OtherCreaturesYouControlNamed(g, item.Controller, item.SourceCardID, "Hare Apparent")
					if n <= 0 {
						return nil
					}
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("1/1 white Rabbit"),
						N:          n,
					}.Apply(ctx)
				}),
		},
	})
}
