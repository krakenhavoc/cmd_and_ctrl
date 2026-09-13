package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stern Lesson — Instant {2}{U}:
//
//	"Draw two cards, then discard a card. Create a tapped Powerstone
//	 token."
//
// The first card to make a tapped token through the structured
// route, which is the reason CreateTokenAdvanced exists: "tapped" is
// a property of THIS card's creation, not of the Powerstone — a
// Powerstone made by another card enters untapped, and forking the
// template in tokens.go per variant is how a token file rots. (The
// older idiom is a private helper that mutates the template before
// handing it to CreateToken — mary_read_and_anne_bonny.go's
// tappedTreasure, hashaton_scarabs_fist.go's Zombie. Both still
// work; new cards should prefer the spec.)
//
// Order matters and is printed order: draw two, THEN queue the
// discard choice (so both drawn cards are legal discards, as in
// paper), then make the token. The token is created while the
// discard prompt is still outstanding — the prompt is a choice owed,
// not a pause in the effect — which is invisible here because
// nothing in the card reads the board between the two.
//
// Inherited gap: the Powerstone's "this mana can't be spent to cast
// a nonartifact spell" restriction is not modelled (see
// PowerstoneToken in tokens.go), so the token taps for unrestricted
// {C}. That makes this card stronger than printed, not weaker.
func init() {
	Register(Spec{
		OracleID: "8315aa34-08b8-403e-a2e6-796a2d6978ad",
		Name:     "Stern Lesson",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.DiscardChoiceForEffect(ctx.Controller(), 1)
			return CreateTokenAdvanced{
				Controller: ctx.Controller(),
				Spec:       Token(PowerstoneToken()).EntersTapped(),
				N:          1,
			}.Apply(ctx)
		},
	})
}
