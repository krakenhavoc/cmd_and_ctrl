package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leyline Dowser — Artifact {2}:
//
//	"{1}, {T}: Mill a card. You may put an instant or sorcery card milled
//	 this way into your hand.
//	 Tap an untapped legendary creature you control: Untap this
//	 artifact."
//
// The mill is Sparring Dummy's: "milled this way" is MillToZone's list,
// so only a card that really reached the graveyard is offered, and only
// an instant or sorcery card (up to one, so declining is an answer).
// The untap is Basalt Monolith's whole-effect untap with a cost of
// tapping an untapped legendary creature. Tapping it is not the {T}
// symbol (CR 302.6), so a creature that arrived this turn may pay it,
// and it does not target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ea2ac09-690d-4cd2-8e3a-fc01ecd7c132",
		Name:         "Leyline Dowser",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Mill a card. You may put an instant or sorcery card milled this way into your hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 1, ThenTail: ctx.ItemTail(leylineDowserMilled)}.Apply(ctx)
				},
			},
			{
				Label:     "Tap an untapped legendary creature you control: Untap this artifact.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				UntapSelf: true,
				Cost:      TapAnUntapped("an untapped legendary creature you control", Creature(), Legendary()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// leylineDowserMilled offers the milled instant and sorcery cards still in
// the graveyard.
var leylineDowserMilled = CardTail("leyline-dowser/after-mill", func(ctx *Context, o game.TailOutcome) error {
	var eligible []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && (c.HasCardType("instant") || c.HasCardType("sorcery")) {
			eligible = append(eligible, id)
		}
	}
	if len(eligible) == 0 || ctx.Item == nil {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Leyline Dowser — put an instant or sorcery card milled this way into your hand?",
		Cards:    eligible,
		Min:      0,
		Max:      1,
		ThenKey:  leylineDowserPick,
		LiveZone: game.ZoneGraveyard,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

var leylineDowserPick = askedPick("choose-cards/leyline-dowser-milled", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
