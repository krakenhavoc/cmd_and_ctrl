package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Masked Vandal — Creature — Shapeshifter {1}{G}, 1/3:
//
//	"Changeling (This card is every creature type.)
//	 When this creature enters, you may exile a creature card from your
//	 graveyard. If you do, exile target artifact or enchantment an opponent
//	 controls."
//
// The target is chosen as the trigger goes on the stack, whether or not the
// card is exiled. As it resolves the controller may pick one creature card
// from their graveyard (picking none is declining); only when that card has
// reached exile (CR 608.2c) is the target exiled, if it is still a legal
// target (CR 608.2b). Skywarp Skaab's shape for the exile-then-reward.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "210940f1-3c11-4877-bbfc-2429c03f98ee",
		Name:            "Masked Vandal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Masked Vandal — you may exile a creature card from your graveyard; if you do, exile target artifact or enchantment an opponent controls",
				maskedVandalOffer),
				effects.TargetPermanent("target artifact or enchantment an opponent controls",
					effects.And(effects.Or(effects.Artifact(), effects.Enchantment()), effects.OpponentControls()))),
		},
	})
}

func maskedVandalOffer(g *game.Game, item *game.StackItem) error {
	ids := effects.B42CardsInGraveyardOf(g, item.Controller, func(c game.Card) bool { return c.IsCreature() })
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Masked Vandal — you may exile a creature card from your graveyard (if you do, exile target artifact or enchantment an opponent controls)",
		Cards:    ids,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  maskedVandalExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// maskedVandalExile exiles the card picked. An on-disk key.
var maskedVandalExile = effects.AskedPick("choose-cards/masked-vandal-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, ctx.ItemTail(maskedVandalThen))
})

// maskedVandalThen is "if you do, exile target artifact or enchantment an
// opponent controls", told the card that reached exile. An on-disk key.
var maskedVandalThen = effects.CardTail("exile-then/masked-vandal-target", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return effects.ExileTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
})
