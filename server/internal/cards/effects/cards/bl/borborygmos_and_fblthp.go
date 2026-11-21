package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Borborygmos and Fblthp — Legendary Creature — Cyclops Homunculus
// {2}{G}{U}{R}, 6/5 (March of the Machine Commander):
//
//	"Whenever Borborygmos and Fblthp enters or attacks, draw a card, then you
//	 may discard any number of land cards. When you discard one or more cards
//	 this way, Borborygmos and Fblthp deals twice that much damage to target
//	 creature.
//	 {1}{U}: Put Borborygmos and Fblthp into its owner's library third from
//	 the top."
//
// The trigger is a walk of two steps: the draw, then (once any prompt the draw
// asked is settled, CR 608.2c) the optional discard of land cards from the
// hand. The discard goes through the one discard path, and the cards that
// really reached the graveyard (a replacement can send one elsewhere) are
// "that much": a reflexive trigger (CR 603.12) is raised only when there were
// any, chooses its own target creature as it goes on the stack, and deals twice
// the count as filed with it, from Borborygmos and Fblthp as it last stood
// (CR 608.2h) if it has left.
//
// The activated ability puts the source third from the top of its owner's
// library, and does nothing once it is no longer the same permanent on the
// battlefield (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22d9eef5-0f86-4d1a-bde0-65b990d4be75",
		Name:         "Borborygmos and Fblthp",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Borborygmos and Fblthp — draw a card, then you may discard any number of land cards; when you discard one or more cards this way, it deals twice that much damage to target creature",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 1}.Apply(ctx) }),
					effects.Instruction(borborygmosFblthpOfferDiscard),
				)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Put Borborygmos and Fblthp into its owner's library third from the top.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.B09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return effects.B22TuckThirdFromTop(g, item.SourceCardID)
			},
		}},
	})
}

// borborygmosFblthpOfferDiscard asks for any number of the land cards in the
// controller's hand.
func borborygmosFblthpOfferDiscard(ctx *effects.Context) error {
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil || p.Hand == nil {
		return nil
	}
	var lands []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsLand() {
			lands = append(lands, c.InstanceID)
		}
	}
	if len(lands) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Item.SourceCardID,
		Question: "Borborygmos and Fblthp — you may discard any number of land cards",
		Cards:    lands,
		Min:      0,
		Max:      len(lands),
		Zone:     game.ZoneHand,
		LiveZone: game.ZoneHand,
		ThenKey:  borborygmosFblthpPicked,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// borborygmosFblthpPicked is the answer: discard the chosen lands. An on-disk
// key: never renamed, never reused.
var borborygmosFblthpPicked = effects.AskedPick("choose-cards/borborygmos-and-fblthp-discard", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.DiscardCardsKeyedForEffect(r.Chooser, r.Picked, r.Source, ctx.ItemTail(borborygmosFblthpDiscarded))
})

// borborygmosFblthpDiscarded raises "when you discard one or more cards this
// way", carrying twice the number that really went. An on-disk key.
var borborygmosFblthpDiscarded = effects.CardTail("then/borborygmos-and-fblthp-discarded", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label:  "Borborygmos and Fblthp — deal twice that much damage to target creature",
		Body:   borborygmosFblthpDamageBody,
		Params: game.EffectParams{Amount: 2 * len(o.IDs)},
	}.Apply(ctx)
})

// borborygmosFblthpDamageBody is the reflexive trigger: Params.Amount damage
// to its target creature. An on-disk key.
var borborygmosFblthpDamageBody = game.ReflexiveBody("borborygmos-and-fblthp/damage", borborygmosFblthpDamage,
	effects.ConstTargets(func() *game.TargetSpec { return effects.Harmful(effects.TargetCreature("target creature")) }))

func borborygmosFblthpDamage(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: p.Amount}.Apply(ctx)
	}
	return nil
}
