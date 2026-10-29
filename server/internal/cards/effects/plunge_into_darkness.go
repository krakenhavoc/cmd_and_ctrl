package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plunge into Darkness — Instant {1}{B}:
//
//	"Choose one —
//	 • Sacrifice any number of creatures. You gain 3 life for each
//	   creature sacrificed this way.
//	 • Pay any amount of life, then look at that many cards from the top
//	   of your library. Put one of those cards into your hand and exile
//	   the rest.
//	 Entwine {B} (Choose both if you pay the entwine cost.)"
//
// Both bullets are choices made as the spell resolves, not costs (CR
// 608.2): the sacrifice is ChoosePermanents' "any number" pick over your
// own creatures (God-Eternal Bontu's shape), the life is PayLifeAmount
// (Necrodominance's, CR 118.12, capped at your life total). The look is
// Impulse's take of exactly one, with the rest exiled together instead
// of put on the bottom. Paying 0 life looks at nothing.
//
// Entwined (CR 702.42, #3866) the sacrifice and its life come first, so
// the life gained can be paid in the second bullet (CR 702.42b); the
// walk holds the second bullet until the sacrifice is answered (#2789).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "07a39452-f9d5-41ba-acbc-fc5e7f698f64",
		Name:          "Plunge into Darkness",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{B}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Sacrifice any number of creatures. You gain 3 life for each creature sacrificed this way.", nil,
				plungeIntoDarknessSacrifice),
			ModeDoing("Pay any amount of life, then look at that many cards from the top of your library. Put one of those cards into your hand and exile the rest.", nil,
				plungeIntoDarknessLook),
		)),
	})
}

func plungeIntoDarknessSacrifice(_ *game.StackItem, ctx *Context, _ int) error {
	return ChoosePermanents{
		Question: "Plunge into Darkness — sacrifice any number of creatures (you gain 3 life for each)",
		Keyed:    plungeIntoDarknessSacrificed,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsCreature() {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 0 // "any number": no floor, no ceiling
		},
	}.Apply(ctx)
}

// plungeIntoDarknessSacrificed is the chosen creatures sacrificed and 3
// life gained for each that really left, keyed (#4290, #4294).
var plungeIntoDarknessSacrificed = RegisterPermanentsPick("prompt-run/plunge-into-darkness-sacrifice", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		return GainLife{Player: ctx.Controller(), Amount: 3 * picked.Count()}.Apply(ctx)
	},
})

func plungeIntoDarknessLook(_ *game.StackItem, ctx *Context, _ int) error {
	return PayLifeAmount{
		Question: "Plunge into Darkness — pay any amount of life, then look at that many cards from the top of your library",
		Unit:     game.PayAmountCards,
		Then:     plungeIntoDarknessLooks,
	}.Apply(ctx)
}

// plungeIntoDarknessLooks is the life paid's branch (#4294): look at that
// many cards from the top of the controller's library.
var plungeIntoDarknessLooks = NumberThen("number/plunge-into-darkness-look-at-that-many",
	func(ctx *Context, r game.NumberAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		controller := ctx.Controller()
		return TakeFromLibraryToHand{
			Player: controller,
			Cards:  ctx.Game.LookAtTopOfLibraryForEffect(controller, r.Number),
			Max:    1,
			Label:  "Plunge into Darkness — put one of those cards into your hand (the rest are exiled)",
			Key:    plungeIntoDarknessTake,
		}.Apply(ctx)
	})

// plungeIntoDarknessTake is the look's pick, keyed (#4294): one of the
// cards into your hand and the rest exiled.
var plungeIntoDarknessTake = RegisterTake("choose-cards/plunge-into-darkness-take",
	TakeDef{Then: plungeIntoDarknessExileRest})

// plungeIntoDarknessExileRest is "and exile the rest": every looked-at
// card still in the library, exiled together.
func plungeIntoDarknessExileRest(g *game.Game, res TakeFromLibraryResult) error {
	if len(res.Rest) == 0 {
		return nil
	}
	g.ExileCardsForEffect(res.Rest)
	return nil
}
