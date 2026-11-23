package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nahiri's Warcrafting — Sorcery {1}{R}{R} (March of the Machine):
//
//	"Nahiri's Warcrafting deals 5 damage to target creature,
//	 planeswalker, or battle. Look at the top X cards of your library,
//	 where X is the excess damage dealt this way. You may exile one of
//	 those cards. Put the rest on the bottom of your library in a random
//	 order. You may play the exiled card this turn."
//
// Excess damage is CR 120.10's: beyond lethal damage for a creature,
// beyond loyalty for a planeswalker, beyond defense for a battle, read
// after the damage lands (s82MomB12DealDamageWithExcess). The exile is a
// choice (up to one); the rest go to the bottom at random whether or not a
// card was exiled. With no excess nothing is looked at.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48838239-bec4-4a57-9858-425f0a937dc9",
		Name:         "Nahiri's Warcrafting",
		Completeness: effects.CompletenessFull,
		Targets:      effects.S82MomB12CreaturePlaneswalkerOrBattle("target creature, planeswalker, or battle"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				excess, err := effects.S82MomB12DealDamageWithExcess(ctx, t.ID, 5)
				if err != nil || excess <= 0 {
					return err
				}
				cards := ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, excess)
				if len(cards) == 0 {
					return nil
				}
				ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Nahiri's Warcrafting — you may exile one of those cards (you may play it this turn); the rest go on the bottom in a random order",
					Cards:    cards,
					Min:      0,
					Max:      1,
					LiveZone: game.ZoneLibrary,
					ThenKey:  s82MomB12WarcraftingExile,
					Carry:    cards,
				})
				return nil
			}
			return nil
		},
	})
}

// s82MomB12WarcraftingExile is the answer: the picked card, if any, is
// exiled with a permission to play it this turn, and the rest still in the
// library go to the bottom in a random order. An on-disk key: never
// renamed, never reused.
var s82MomB12WarcraftingExile = game.RegisterChooseCardsThen("choose-cards/s82-mom-b12-warcrafting-exile", func(g *game.Game, r game.CardsChosen) error {
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	var rest []uuid.UUID
	for _, id := range effects.CardsStillInALibrary(g, r.Carry) {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if picked != uuid.Nil {
		if err := g.ExileCardWithPermissionForEffect(picked, game.CastPermission{
			Player: r.Chooser,
			Label:  "Nahiri's Warcrafting — play the exiled card this turn",
		}); err != nil {
			return err
		}
	}
	return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, rest)
})
