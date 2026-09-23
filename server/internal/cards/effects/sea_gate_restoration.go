package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sea Gate Restoration // Sea Gate, Reborn — modal double-faced card.
// This file is the FRONT face, Sorcery {4}{U}{U}{U}:
//
//	"Draw cards equal to the number of cards in your hand plus one.
//	 You have no maximum hand size for the rest of the game."
//
// The back face, Sea Gate, Reborn, is registered with the MDFC land
// cycle in mdfc_lands.go under "<oracle>#1", as Hagra Mauling's is.
//
// The count is read as the spell RESOLVES (CR 608.2h), when the spell
// itself is on the stack and no longer in the hand, so a hand of four
// draws five. It is read once, before any card is drawn: "equal to the
// number of cards in your hand plus one" fixes the number, and the
// draws do not feed back into it.
//
// "No maximum hand size for the rest of the game" is Finale of
// Revelation's clause and uses the same setter,
// Game.SetMaxHandSizeForEffect: a player-level grant with no
// battlefield dependency, which is what "for the rest of the game"
// needs. Spec.NoMaxHandSize (Reliquary Tower) is the wrong tool — it
// is derived from a permanent and lapses when the permanent does, and
// this spell is in the graveyard a moment after it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a8d41fe-e04d-484b-a7d1-19be311e6ca7",
		Name:         "Sea Gate Restoration",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			n := 1
			if p := ctx.PlayerByID(controller); p != nil && p.Hand != nil {
				n += p.Hand.Size()
			}
			if err := (DrawCards{Player: controller, N: n}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.SetMaxHandSizeForEffect(controller, game.NoMaxHandSize)
		},
	})
}
