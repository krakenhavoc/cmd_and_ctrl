package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cosmic Rebirth — Instant {1}{G}{W} (March of the Machine: The Aftermath,
// #5476):
//
//	"Choose target permanent card in your graveyard. If it has mana value 3
//	 or less, you may put it onto the battlefield. If you don't put it onto
//	 the battlefield, put it into your hand.
//	 You gain 3 life."
//
// The target is a permanent card in your graveyard, checked again as the
// spell resolves (CR 608.2b): if it left, the spell does nothing, the life
// included. With a mana value of 3 or less the controller is asked
// (MayChoice, CR 608.2d): yes puts it onto the battlefield under their
// control, no puts it into their hand. A card of mana value 4 or more
// goes to the hand with no question. The 3 life is the last instruction,
// gained after the card has moved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "add60aad-2532-4f59-bfb9-f0168c2adda6",
		Name:         "Cosmic Rebirth",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target permanent card in your graveyard", Permanent(), YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			if card.ManaValue() > 3 {
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
				return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
			}
			return MayChoice{
				Question: "Cosmic Rebirth — put " + card.Name + " onto the battlefield? (If you don't, put it into your hand.)",
				YesLabel: "Put it onto the battlefield",
				NoLabel:  "Put it into your hand",
				Then:     cosmicRebirthAnswer,
				Carry:    []uuid.UUID{id},
			}.Apply(ctx)
		},
	})
}

// cosmicRebirthAnswer moves the card to the battlefield on a yes and to the
// hand on a no, then gains the 3 life. Keyed on disk (#4291).
var cosmicRebirthAnswer = MayYesNo("confirm/cosmic-rebirth-battlefield-or-hand",
	func(ctx *Context, r game.ConfirmAnswer) error {
		if err := (ReturnFromGraveyard{Target: carryAt(r, 0), Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
			return err
		}
		return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
	},
	func(ctx *Context, r game.ConfirmAnswer) error {
		if err := (ReturnFromGraveyard{Target: carryAt(r, 0), Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
		return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
	})
