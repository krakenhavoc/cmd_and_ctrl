package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lantern of Revealing — Artifact {3} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"{T}: Add one mana of any color.
//	 {4}, {T}: Look at the top card of your library. If it's a land card, you
//	 may put it onto the battlefield tapped. If you don't put the card onto
//	 the battlefield, you may put it on the bottom of your library."
//
// The look is private. A land is offered onto the battlefield tapped; if it
// goes there nothing more happens, and if it stays (declined, or not a land)
// you are asked whether to put it on the bottom: the card stays on top when
// you decline.
//
// No simplification.
func init() {
	const name = "Lantern of Revealing"
	effects.Register(effects.Spec{
		OracleID:     "00e25126-b807-4753-ba63-9a533755e5ce",
		Name:         name,
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Look at the top card of your library. If it's a land card, you may put it onto the battlefield tapped. If you don't put the card onto the battlefield, you may put it on the bottom of your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  lanternOfRevealingLook,
		}},
	})
}

const (
	lanternOfRevealingBottom = "Put it on the bottom of your library"
	lanternOfRevealingLeave  = "Leave it on top"
)

func lanternOfRevealingLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	top := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if len(top) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(top[0]); ok && c.IsLand() {
		return effects.PutFromLibraryOntoBattlefield{
			Player:   item.Controller,
			Cards:    top,
			Match:    effects.Land(),
			Max:      1,
			Optional: true,
			Tapped:   true,
			Label:    "Lantern of Revealing — you may put the land onto the battlefield tapped",
			ThenTail: lanternOfRevealingAfterPut,
		}.Apply(ctx)
	}
	return lanternOfRevealingAskBottom(ctx, top)
}

// lanternOfRevealingAfterPut asks about the bottom when the land did not go
// onto the battlefield. An on-disk key: never renamed, never reused.
var lanternOfRevealingAfterPut = effects.PutFromLibraryTail("put-from-library/lantern-of-revealing-after-put", func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
	if len(res.Entered) > 0 {
		return nil
	}
	return lanternOfRevealingAskBottom(ctx, res.Rest)
})

func lanternOfRevealingAskBottom(ctx *effects.Context, rest []uuid.UUID) error {
	if len(rest) == 0 {
		return nil
	}
	return effects.PickOption{
		Question: "Lantern of Revealing — put the card on the bottom of your library?",
		Options:  []game.ChoiceOption{{Label: lanternOfRevealingBottom}, {Label: lanternOfRevealingLeave}},
		ThenKey:  lanternOfRevealingAnswer,
		Carry:    rest[:1],
	}.Apply(ctx)
}

// lanternOfRevealingAnswer carries out the pick. An on-disk key: never
// renamed, never reused.
var lanternOfRevealingAnswer = effects.OptionPickThen("option-pick/lantern-of-revealing-bottom", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Label != lanternOfRevealingBottom || len(r.Carry) == 0 {
		return nil
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, r.Carry[:1])
})
