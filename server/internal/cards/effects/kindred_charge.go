package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kindred Charge — Sorcery {4}{R}{R}:
//
//	"Choose a creature type. For each creature you control of the
//	 chosen type, create a token that's a copy of that creature. Those
//	 tokens gain haste. Exile them at the beginning of the next end
//	 step."
//
// Chosen as the spell resolves (#2382). The originals are listed once,
// when the answer arrives, so a token copy that is itself of the type
// is not copied again. Every copy is made by ONE creation, one group per
// original (#4594, CreateTokenCopiesTogether): the tokens are created at
// the same time (CR 608.2f) and enter together, so each newcomer sees
// the others enter (CR 603.6a) and a "one or more" trigger counts one
// event (CR 603.2c). "Those tokens gain haste" is an effect of the
// spell, not an exception to the copy (CR 707.9a), so haste is a grant
// on each token and a copy of one has none. The grant and the exile, a
// CR 603.7 delayed trigger over exactly the tokens this spell made, are
// the creation's keyed tail, so a copy whose entry asks a question
// (Adaptive Automaton's creature type) still gets them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb0b5b5f-54d3-49c6-9ad8-d44a5551acdb",
		Name:         "Kindred Charge",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Kindred Charge — choose a creature type",
				ctx.ItemTail(kindredChargeTypeChosen))
			return nil
		},
	})
}

// kindredChargeTypeChosen copies each creature of the chosen type the caster
// controls. An on-disk key.
var kindredChargeTypeChosen = creatureTypeChosen("then/kindred-charge-type-chosen", func(ctx *Context, t string, _ game.TailOutcome) error {
	controller := ctx.Controller()
	return kindredChargeCopies(ctx, controller, creaturesOfTypeYouControl(ctx.Game, controller, t))
})

func kindredChargeCopies(ctx *Context, controller uuid.UUID, originals []uuid.UUID) error {
	return CreateTokenCopiesTogether{
		Controller: controller, Copies: originals, Then: ctx.ItemTail(kindredChargeCopiesRest),
	}.Apply(ctx)
}

// kindredChargeCopiesRest is "those tokens gain haste. Exile them at the
// beginning of the next end step", told the tokens that landed. An on-disk
// key.
var kindredChargeCopiesRest = CardTail("kindred-charge/haste-and-exile-copies", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: "Kindred Charge — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		Label: "Kindred Charge — exile the tokens",
		Cards: o.IDs,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
})
