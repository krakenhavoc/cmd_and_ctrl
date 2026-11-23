package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sparring Dummy — Artifact Creature — Scarecrow {1}{G}, 1/3 (Avatar: The
// Last Airbender, #4459):
//
//	"Defender
//	 {T}: Mill a card. You may put a land card milled this way into your
//	 hand. You gain 2 life if a Lesson card is milled this way."
//
// "Milled this way" is MillToZone.ThenTail's list: only the cards that
// really reached the graveyard. A land card among them is offered (up to
// one, so declining is an answer) and goes to hand if picked; then the
// life is gained if any of the milled cards is a Lesson card, wherever it
// is now. An empty library mills nothing and does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "036b7b89-e49a-4560-9de9-1d4d71f3f666",
		Name:            "Sparring Dummy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Mill a card. You may put a land card milled this way into your hand. You gain 2 life if a Lesson card is milled this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MillToZone{N: 1, ThenTail: ctx.ItemTail(sparringDummyMilled)}.Apply(ctx)
			},
		}},
	})
}

// sparringDummyMilled is the clause after the mill: offer the milled land
// cards still in the graveyard, then the Lesson check.
var sparringDummyMilled = effects.CardTail("sparring-dummy/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	var lands []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
			lands = append(lands, id)
		}
	}
	if len(lands) == 0 || ctx.Item == nil {
		return sparringDummyGain(ctx, o.IDs)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Sparring Dummy — put a land card milled this way into your hand?",
		Cards:    lands,
		Min:      0,
		Max:      1,
		ThenKey:  sparringDummyLandPick,
		Carry:    o.IDs,
		LiveZone: game.ZoneGraveyard,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// sparringDummyLandPick is the land chosen (or none), keyed (#4253); the
// milled cards ride the prompt as Carry for the Lesson check after it.
var sparringDummyLandPick = effects.AskedPick("choose-cards/sparring-dummy-milled-land", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return sparringDummyGain(ctx, r.Carry)
})

// sparringDummyGain is "You gain 2 life if a Lesson card is milled this
// way."
func sparringDummyGain(ctx *effects.Context, milled []uuid.UUID) error {
	for _, id := range milled {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && effects.IsLessonCard(c) {
			return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		}
	}
	return nil
}
