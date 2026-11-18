package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcane Endeavor — Sorcery {5}{U}{U} (Adventures in the Forgotten Realms):
//
//	"Roll two d8 and choose one result. Draw cards equal to that result.
//	 Then you may cast an instant or sorcery spell with mana value less than
//	 or equal to the other result from your hand without paying its mana
//	 cost."
//
// The shared two-dice shape is afrEndeavorTail; the free cast is the
// catalog's hand offer (free_hand_cast.go), made after the draw.
func init() {
	Register(Spec{
		OracleID: "901ee0e3-1011-40e1-815a-09246112e8dc", Name: "Arcane Endeavor", Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return RollDice{Sides: 8, N: 2, Then: arcaneEndeavorRolled}.Apply(ctx)
		},
	})
}

var arcaneEndeavorFreeCast = registerFreeHandCastOffer("choose-cards/arcane-endeavor-cast", "Arcane Endeavor")

var arcaneEndeavorRolled = afrEndeavorTail("dice/arcane-endeavor", "Arcane Endeavor",
	func(chosen, other int) string {
		return "Draw " + strconv.Itoa(chosen) + ", free cast up to mana value " + strconv.Itoa(other)
	},
	func(ctx *Context, chosen, other int) error {
		if err := (DrawCards{N: chosen}).Apply(ctx); err != nil {
			return err
		}
		offerFreeHandCast(ctx, ctx.Controller(), arcaneEndeavorFreeCast,
			"an instant or sorcery spell with mana value "+strconv.Itoa(other)+" or less",
			freeInstantOrSorcery().withManaValueAtMost(other))
		return nil
	})
