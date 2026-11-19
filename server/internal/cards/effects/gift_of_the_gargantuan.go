package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gift of the Gargantuan — Sorcery {2}{G} (Foundations Jumpstart):
//
//	"Look at the top four cards of your library. You may reveal a
//	 creature card and/or a land card from among them and put the
//	 revealed cards into your hand. Put the rest on the bottom of your
//	 library in any order."
//
// In the Presence of Ages' two slots (a creature card and a land card,
// a card that is both filling only one) with the rest going to the
// bottom in an order the player chooses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34f9c2a9-ca7c-4a02-aec5-e8f407d09580",
		Name:         "Gift of the Gargantuan",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Optional: true,
				Label:    "Gift of the Gargantuan — reveal a creature card and/or a land card and put them into your hand",
				Key:      giftOfTheGargantuanTake,
			}.Apply(ctx)
		},
	})
}

// giftOfTheGargantuanTake is the look's registered rule and rest: a
// creature slot and a land slot, revealed, then the rest to the bottom
// in any order. An on-disk key: never renamed, never reused.
var giftOfTheGargantuanTake = RegisterTake("choose-cards/gift-of-the-gargantuan-take",
	TakeDef{
		Reveal: true,
		Slots: []TakeSlot{
			{Label: "a creature card", Match: Creature(), Max: 1},
			{Label: "a land card", Match: Land(), Max: 1},
		},
		Then: TakeRestOnBottomInAnyOrder,
	})
