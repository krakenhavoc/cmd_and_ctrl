package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Primal Surge — {8}{G}{G} Sorcery: "Exile the top card of your library. If it's
// a permanent card, you may put it onto the battlefield. If you do, repeat this
// process."
//
// Each pass exiles one card, asks about it if it is a permanent card, and only a
// "yes" that really puts it onto the battlefield starts the next pass; a
// nonpermanent card or a "no" ends it, the card staying in exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44b12ccc-81cb-4ec1-891e-4bf706e53aa4",
		Name:         "Primal Surge",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return primalSurgeRound(ctx)
		},
	})
}

// primalSurgeRound is one pass of Primal Surge: exile the top card of your
// library; if it is a permanent card you may put it onto the battlefield, and
// if you do, repeat this process.
func primalSurgeRound(ctx *effects.Context) error {
	return ctx.Game.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), 1, ctx.ItemTail(primalSurgeExiled))
}

var primalSurgeExiled game.TailRef
var primalSurgeAnswer game.OptionPickThen

func init() {
	primalSurgeExiled = effects.CardTail("then/s82-old-b435-primal-surge-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
		for _, id := range o.IDs {
			c, ok := ctx.Game.LookupCardForEffect(id)
			z := ctx.Game.FindCardZoneForEffect(id)
			if !ok || z == nil || z.Kind != game.ZoneExile || !c.IsPermanent() {
				continue
			}
			return effects.PickOption{
				Question: "Primal Surge — put " + c.Name + " onto the battlefield?",
				Options: []game.ChoiceOption{
					{Label: "Put it onto the battlefield", Cards: []uuid.UUID{id}},
					{Label: "Leave it in exile", Cards: []uuid.UUID{id}},
				},
				ThenKey:  primalSurgeAnswer,
				Carry:    []uuid.UUID{id},
				WithItem: true,
			}.Apply(ctx)
		}
		return nil
	})
	primalSurgeAnswer = effects.OptionPickThen("option-pick/s82-old-b435-primal-surge", func(ctx *effects.Context, r game.OptionPicked) error {
		if r.Index != 0 || len(r.Carry) == 0 {
			return nil
		}
		id := r.Carry[0]
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			return nil
		}
		entered, err := ctx.Game.ReturnFromExileToBattlefieldForEffect(id, ctx.Controller(), false)
		if err != nil || entered == uuid.Nil {
			return err
		}
		return primalSurgeRound(ctx)
	})
}
