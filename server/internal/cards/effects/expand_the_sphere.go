package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expand the Sphere — Sorcery {3}{G} (Phyrexia: All Will Be One, #5479):
//
//	"Look at the top six cards of your library. Put up to two land cards from
//	 among them onto the battlefield tapped and the rest on the bottom of
//	 your library in a random order. If you put fewer than two lands onto the
//	 battlefield this way, proliferate a number of times equal to the
//	 difference."
//
// Elvish Rejuvenator's look-and-put with up to two lands. The rider is a
// keyed tail (s82OneB08ExpandTheSphereRest): the rest goes to the bottom,
// then it proliferates 2 minus the lands that actually entered, each pass
// asked once the one before is answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0fac3187-9b86-4854-854b-a4c203526a64",
		Name:         "Expand the Sphere",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, 6),
				Match:    Land(),
				Max:      2,
				Optional: true,
				Tapped:   true,
				Label:    "Expand the Sphere — put up to two land cards from among them onto the battlefield tapped",
				ThenTail: s82OneB08ExpandTheSphereRest,
			}.Apply(ctx)
		},
	})
}
