package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ballad of the Black Flag — Enchantment — Saga {1}{U}{U}:
//
//	"I, II, III — Mill three cards. You may put a historic card from
//	 among them into your hand. (Artifacts, legendaries, and Sagas are
//	 historic.)
//	 IV — Historic spells you cast this turn cost {2} less to cast."
//
// Chapters I-III are Ainok Wayfarer's mill-then-take shape on a historic
// card (CR 700.6): only a card this mill put into the graveyard is
// offered, and the pick may be none. Chapter IV is a promise about every
// historic spell its controller casts for the rest of the turn
// (NextSpellPromise with EveryMatchingSpell, as Urza, Planeswalker's
// reduction), priced as a CR 601.2f reduction; it outlives the Saga.
func init() {
	const name = "Ballad of the Black Flag"
	mill := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(balladOfTheBlackFlagMilled)}.Apply(ctx)
	}
	label := "mill three cards; you may put a historic card from among them into your hand"
	effects.Register(effects.Spec{
		OracleID:     "e8e15ab4-e4c7-4faf-b469-288451ddee02",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, label), mill),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, label), mill),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, label), mill),
			effects.ChapterTrigger(4, effects.SagaChapterLabel(name, 4, "historic spells you cast this turn cost {2} less to cast"),
				effects.Do(effects.GrantNextSpellPromise{From: name, Promise: game.NextSpellPromise{
					Filter:             game.PermissionFilter{HistoricOnly: true},
					Reduce:             2,
					EveryMatchingSpell: true,
					Text:               "Historic spells you cast this turn cost {2} less to cast.",
				}})),
		},
	})
}

// balladOfTheBlackFlagMilled offers a historic card from among the milled
// cards. An on-disk key.
var balladOfTheBlackFlagMilled = effects.CardTail("ballad-of-the-black-flag/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	player := ctx.Controller()
	var historic []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !effects.B09IsHistoric(c) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		historic = append(historic, id)
	}
	if len(historic) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Ballad of the Black Flag — you may put a historic card from among them into your hand",
		Cards:    historic,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  balladOfTheBlackFlagPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// balladOfTheBlackFlagPick is the historic card taken. An on-disk key.
var balladOfTheBlackFlagPick = effects.AskedPick("choose-cards/ballad-of-the-black-flag-historic", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
})
