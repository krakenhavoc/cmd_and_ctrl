package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exotic Pets — Instant {1}{W}{U} (Streets of New Capenna, #5491):
//
//	"Create two 1/1 blue Fish creature tokens with "This token can't be
//	 blocked." Then for each kind of counter among creatures you control,
//	 put a counter of that kind on either of those tokens."
//
// The Fish arrive first, then the kinds are read off the creatures the
// controller has (CR 122.1; the new Fish have no counters, so they add no
// kind), in name order. Each kind is its own choice of which Fish gets the
// counter, asked in turn, so two kinds may go on one token or one on each.
// A Fish that has died by the time its answer comes is skipped. Doubled
// tokens (Doubling Season) are all "those tokens".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "497cb281-e057-42fc-937d-512855b74159",
		Name:         "Exotic Pets",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: item.Controller,
				Source:     item.SourceCardID,
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(S82SncB07FishToken()), Count: 2}},
			}, ctx.ItemTail(exoticPetsFish))
		},
	})
}

// exoticPetsFish is "then for each kind of counter among creatures you
// control, put a counter of that kind on either of those tokens", as a keyed
// tail. An on-disk key.
var exoticPetsFish = CardTail("exotic-pets/fish", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return exoticPetsAsk(ctx, sncB07KindsAmongCreaturesYouControl(ctx.Game, ctx.Controller()), o.IDs)
})

// exoticPetsAsk asks which Fish gets the first of the kinds, then the next.
func exoticPetsAsk(ctx *Context, kinds []string, fish []uuid.UUID) error {
	var live []uuid.UUID
	for _, id := range fish {
		if onBattlefield(ctx.Game, id) {
			live = append(live, id)
		}
	}
	if len(kinds) == 0 || len(live) == 0 {
		return nil
	}
	if len(live) == 1 {
		for _, kind := range kinds {
			if err := ctx.Game.AddCounterByForEffect(ctx.Controller(), live[0], kind, 1); err != nil {
				return err
			}
		}
		return nil
	}
	options := make([]game.ChoiceOption, len(live))
	for i := range live {
		options[i] = game.ChoiceOption{Label: fmt.Sprintf("Fish %d", i+1)}
	}
	return PickOption{
		Question: "Exotic Pets — put a " + kinds[0] + " counter on which Fish?",
		Options:  options,
		ThenKey:  exoticPetsPicked,
		Carry:    live,
		Words:    kinds,
	}.Apply(ctx)
}

// exoticPetsPicked puts the first kind on the chosen Fish and asks about
// the rest. An on-disk key.
var exoticPetsPicked game.OptionPickThen

func init() {
	exoticPetsPicked = OptionPickThen("option-pick/exotic-pets-fish", func(ctx *Context, r game.OptionPicked) error {
		if r.Option == nil || len(r.Words) == 0 || r.Index < 0 || r.Index >= len(r.Carry) {
			return nil
		}
		if onBattlefield(ctx.Game, r.Carry[r.Index]) {
			if err := ctx.Game.AddCounterByForEffect(ctx.Controller(), r.Carry[r.Index], r.Words[0], 1); err != nil {
				return err
			}
		}
		return exoticPetsAsk(ctx, r.Words[1:], r.Carry)
	})
}
