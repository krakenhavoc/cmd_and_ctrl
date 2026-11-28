package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reviving Vapors — Instant {2}{W}{U} (Invasion, #5616):
//
//	"Reveal the top three cards of your library and put one of them into your
//	 hand. You gain life equal to that card's mana value. Put all other cards
//	 revealed this way into your graveyard."
//
// The three cards are revealed (public), the controller picks one, and the
// picked card goes to hand while the others go to the graveyard. The life is
// the picked card's mana value, read before it moves (CR 608.2h). A library of
// fewer than three reveals what is there; with nothing to reveal nothing
// happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57104a79-cfb6-4330-8ea1-7726534ada56",
		Name:         "Reviving Vapors",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 3,
				"Reviving Vapors — reveal the top three cards of your library")
			if len(revealed) == 0 {
				return nil
			}
			return effects.RevealPick{
				Player:   ctx.Controller(),
				Owner:    ctx.Controller(),
				Question: "Reviving Vapors — choose one of these cards to put into your hand; the others go to your graveyard",
				Cards:    revealed,
				Min:      1,
				Max:      1,
				Keyed:    revivingVaporsPicked,
			}.Apply(ctx)
		},
	})
}

// revivingVaporsPicked puts the picked card into hand; once it has reached a
// hand, revivingVaporsInHand gains the life and puts the rest into the
// graveyard. An on-disk key: never renamed, never reused.
var revivingVaporsPicked = effects.RegisterRevealPick("choose-cards/reviving-vapors-settle", func(ctx *effects.Context, picked, left []uuid.UUID, _ []int) error {
	if len(picked) == 0 {
		return revivingVaporsBin(ctx, left)
	}
	return ctx.Game.BounceCardsToHandKeyedForEffect(picked, effects.WithCarry(ctx.ItemTail(revivingVaporsInHand), left...))
})

// revivingVaporsInHand gains life equal to the mana value of the card that
// reached the hand (an exit the CR 614 window cancelled gains nothing), then
// puts the rest of the revealed cards into the graveyard. An on-disk key:
// never renamed, never reused.
var revivingVaporsInHand = effects.CardTail("then/reviving-vapors-in-hand", func(ctx *effects.Context, o game.TailOutcome) error {
	life := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			life += c.ManaValue()
		}
	}
	if err := (effects.GainLife{Amount: life}).Apply(ctx); err != nil {
		return err
	}
	return revivingVaporsBin(ctx, o.Tail.Carry)
})

func revivingVaporsBin(ctx *effects.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		if ctx.Game.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return nil
}
