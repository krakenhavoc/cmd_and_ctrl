package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faithless Looting — Sorcery for {R}:
//
//	"Draw two cards, then discard two cards."
//
// The archetypal loot. Printed order matters and the engine gets it
// right for free: DrawCards resolves synchronously, then the discard
// choice is queued against the post-draw hand, so the cards just
// drawn are legal discards exactly as in paper.
//
// Flashback {2}{R} (S29). The card is the mechanic's simplest
// possible case and a good place to state what the two declarations
// buy: CastableZones opens the graveyard as a cast source, and the
// Flashback offer is bound to it — so the {R} in the corner is
// payable only out of hand, and the {2}{R} only out of the
// graveyard. The exile-on-leaving-the-stack half rides the
// constructor, which is what stops the loot from repeating every
// turn for one red mana.
func init() {
	Register(Spec{
		OracleID:         "3d6fa57a-aa53-4b5c-b8af-a7612c823117",
		Name:             "Faithless Looting",
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return lootOne(ctx.Game, item, 2)
		},
	})
}
