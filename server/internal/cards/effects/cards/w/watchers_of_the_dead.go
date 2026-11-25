package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Watchers of the Dead — {2} Artifact Creature — Cat 2/2 (Amonkhet, #5535):
//
//	"Exile this creature: Each opponent chooses two cards in their graveyard
//	 and exiles the rest."
//
// Each opponent with more than two cards is asked, one at a time (the
// choice is theirs, CR 608.2d), to name the two they keep; the rest of that
// graveyard is exiled as they answer. An opponent with two or fewer cards
// keeps them all and is not asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ac46f46-3099-4ad9-ae58-bf25d7270e87",
		Name:         "Watchers of the Dead",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile this creature: Each opponent chooses two cards in their graveyard and exiles the rest.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ExileThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return watchersAskNext(ctx, ctx.Opponents())
			},
		}},
	})
}

func watchersAskNext(ctx *effects.Context, order []uuid.UUID) error {
	for i, opp := range order {
		cards := effects.GraveyardCardIDs(ctx, opp, func(game.Card) bool { return true })
		if len(cards) <= 2 {
			continue
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    opp,
			FromPlayer: opp,
			Source:     ctx.Item.SourceCardID,
			Question:   "Watchers of the Dead — choose two cards in your graveyard to keep; the rest are exiled",
			Cards:      cards,
			Min:        2,
			Max:        2,
			ThenKey:    watchersKept,
			LiveZone:   game.ZoneGraveyard,
			Carry:      append([]uuid.UUID(nil), order[i+1:]...),
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return nil
}

var watchersKept game.ChooseCardsThen

func init() {
	watchersKept = effects.AskedPick("choose-cards/watchers-of-the-dead-keep", func(ctx *effects.Context, r game.CardsChosen) error {
		keep := map[uuid.UUID]bool{}
		for _, id := range r.Picked {
			keep[id] = true
		}
		for _, id := range effects.GraveyardCardIDs(ctx, r.Chooser, func(game.Card) bool { return true }) {
			if keep[id] {
				continue
			}
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return watchersAskNext(ctx, r.Carry)
	})
}
