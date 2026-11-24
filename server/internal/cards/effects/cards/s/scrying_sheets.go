package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrying Sheets — Snow Land (Secret Lair Drop, #5462):
//
//	"{T}: Add {C}.
//	 {1}{S}, {T}: Look at the top card of your library. If that card is snow,
//	 you may reveal it and put it into your hand."
//
// The look is private. A snow card (read as it sits in the library, CR 205.4g)
// is offered into the hand; a card that is not snow, or one the player leaves,
// stays on top. The {S} is paid by mana a snow source made (CR 107.4h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "0f98e055-ab61-4314-aae8-9d3c19f66acf",
		Name:          "Scrying Sheets",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.ColorlessTapRow()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{S}, {T}: Look at the top card of your library. If that card is snow, you may reveal it and put it into your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{S}"), effects.TapCost()),
			Effect:  scryingSheetsLook,
		}},
	})
}

const (
	scryingSheetsTake  = "Reveal it and put it into your hand"
	scryingSheetsLeave = "Leave it on top"
)

func scryingSheetsLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	top := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	if len(top) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(top[0])
	if !ok || c.IsToken() || !c.HasSupertype("snow") {
		return nil
	}
	return effects.PickOption{
		Question: "Scrying Sheets — reveal the snow card on top of your library and put it into your hand?",
		Options:  []game.ChoiceOption{{Label: scryingSheetsTake}, {Label: scryingSheetsLeave}},
		ThenKey:  scryingSheetsAnswer,
		Carry:    []uuid.UUID{top[0]},
	}.Apply(ctx)
}

// scryingSheetsAnswer carries out the pick. An on-disk key: never renamed,
// never reused.
var scryingSheetsAnswer = effects.OptionPickThen("option-pick/scrying-sheets-top-card", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 || r.Option.Label != scryingSheetsTake {
		return nil
	}
	g, id := ctx.Game, r.Carry[0]
	g.RevealForEffect(game.RevealSpec{Player: r.Chooser, Source: r.Source, Reason: "Scrying Sheets — reveal the snow card", Cards: []uuid.UUID{id}})
	g.TakeFromLibraryToHandForEffect(r.Chooser, []uuid.UUID{id})
	return nil
})
