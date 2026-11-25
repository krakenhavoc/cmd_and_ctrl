package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood for Bones — Sorcery {3}{B} (Core Set 2020):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Return a creature card from your graveyard to the battlefield, then
//	 return another creature card from your graveyard to your hand."
//
// The sacrifice is Altar's Reap's additional cost (CR 118.8), paid before the
// spell resolves, so the sacrificed creature is itself a candidate for either
// return. Nothing targets: as it resolves its controller chooses a creature
// card (CR 608.2d) and it enters the battlefield, and only then is "another"
// creature card chosen from what the graveyard holds now and put into hand
// (the first is no longer there). With no creature card to return the second
// choice is skipped, and with one the spell returns it to the battlefield and
// stops.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "fc7439ec-32e3-47db-9ff4-6eb84cb33d89",
		Name:           "Blood for Bones",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Purpose:        game.Purpose{Answers: game.AnswerSacOutlet},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			bloodForBonesAsk(ctx.Game, item.Controller, item.SourceCardID, bloodForBonesFirst,
				"Blood for Bones — return a creature card from your graveyard to the battlefield")
			return nil
		},
	})
}

// bloodForBonesAsk asks `player` to choose one creature card in their
// graveyard, if there is one.
func bloodForBonesAsk(g *game.Game, player, source uuid.UUID, then game.ChooseCardsThen, question string) {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: question,
		Cards:    ids,
		Min:      1,
		Max:      1,
		ThenKey:  then,
	})
}

// bloodForBonesFirst puts the chosen creature card onto the battlefield, then
// asks for the second. An on-disk key: never renamed, never reused.
var bloodForBonesFirst = game.RegisterChooseCardsThen("choose-cards/blood-for-bones-first", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(r.Picked[0], r.Chooser, false,
		game.KeyedTail{Ref: bloodForBonesAfterFirst, Player: r.Chooser, Source: r.Source})
})

// bloodForBonesAfterFirst is "then return another creature card from your
// graveyard to your hand". An on-disk key: never renamed, never reused.
var bloodForBonesAfterFirst = effects.CardTail("tail/blood-for-bones-after-first", func(ctx *effects.Context, _ game.TailOutcome) error {
	bloodForBonesAsk(ctx.Game, ctx.Controller(), ctx.Source(), bloodForBonesSecond,
		"Blood for Bones — return another creature card from your graveyard to your hand")
	return nil
})

// bloodForBonesSecond puts the second chosen creature card into its owner's
// hand. An on-disk key: never renamed, never reused.
var bloodForBonesSecond = game.RegisterChooseCardsThen("choose-cards/blood-for-bones-second", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.ReturnFromGraveyardForEffect(r.Picked[0], game.ZoneHand)
})
