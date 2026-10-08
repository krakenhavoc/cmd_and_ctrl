package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Formless Genesis — Kindred Sorcery — Shapeshifter {2}{G}:
//
//	"Changeling (This card is every creature type.)
//	 Create an X/X colorless Shapeshifter creature token with changeling
//	 and deathtouch, where X is the number of land cards in your
//	 graveyard.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: see Spitting Image. X is read at
// resolution and INCLUDES the land that paid for retrace, which was
// discarded as a cost while the spell sat on the stack. With no land in
// the graveyard the token is 0/0 and dies to CR 704.5f, as printed; a
// retrace cast always has at least the one it discarded.
//
// The card's own changeling is printed keyword data and needs nothing
// here; the TOKEN declares it on its template's Keywords, since a token
// has no oracle ID. The template is built by hand because its size is
// decided at resolution, as Corpse Cobble's Zombie is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b978bdfe-a73e-4e99-90d7-4f99182c45ee",
		Name:             "Formless Genesis",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{2}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := b17LandCardsInGraveyard(ctx.Game, item.Controller)
			return CreateToken{Template: formlessGenesisShapeshifter(x), N: 1}.Apply(ctx)
		},
	})
}

// formlessGenesisShapeshifter is the X/X colorless Shapeshifter with
// changeling and deathtouch. PrintedPTKnown, because 0/0 is a real size.
func formlessGenesisShapeshifter(x int) game.Card {
	return game.Card{
		Name:           "Shapeshifter",
		TypeLine:       "Token Creature — Shapeshifter",
		Power:          x,
		Toughness:      x,
		PrintedPTKnown: true,
		Keywords:       []string{"changeling", "deathtouch"},
	}
}
