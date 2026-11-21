package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jacob Hauken, Inspector // Hauken's Insight — {1}{U} Legendary Creature —
// Human Advisor 0/2 // Legendary Enchantment (Innistrad: Crimson Vow,
// #5495):
//
//	Front: "{T}: Draw a card, then exile a card from your hand face down.
//	        You may look at that card for as long as it remains exiled. You
//	        may pay {4}{U}{U}. If you do, transform Jacob Hauken."
//	Back:  "At the beginning of your upkeep, exile the top card of your
//	        library face down. You may look at that card for as long as it
//	        remains exiled.
//	        Once during each of your turns, you may play a land or cast a
//	        spell from among the cards exiled with this permanent without
//	        paying its mana cost."
//
// Both exiles are LINKED to the permanent as an object (CR 607.2a, Kayla's
// Music Box's hidden exile: the permanent's controller may look at the card,
// nobody else may, CR 406.3). The play permission reads "exiled with this
// permanent", not with one ability, so it covers the cards the front face's
// ability exiled as well: the link is to the object, which survives the
// transform (CR 712.18). It is a standing permission derived off the
// battlefield, in your turn only, with a budget of one use however spent
// (Serra Paragon's PermissionTypePermanent: a land or any spell), free of
// mana cost; a Hauken that leaves strands the cards (CR 400.7).
//
// The front ability is one resolution: the draw, the exile from hand (the
// card chosen is the controller's, and may be the one just drawn; an empty
// hand exiles nothing), then the optional {4}{U}{U}, offered even when the
// hand was empty, since "you may pay" does not depend on the exile.
// Ashling's paid branch transforms, unless Hauken has turned over since the
// ability went on the stack (CR 701.27f).
//
// No simplification.
func init() {
	const oracle = "16d2a56a-7181-4ef3-866b-10dd8615d9e9"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Jacob Hauken, Inspector",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card, then exile a card from your hand face down. You may look at that card for as long as it remains exiled. You may pay {4}{U}{U}. If you do, transform Jacob Hauken.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  jacobHaukenResolve,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Hauken's Insight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Hauken's Insight — exile the top card of your library face down; you may look at it", func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTopFaceDownLinked{
					Source:   effects.HiddenRefOfActivation(item),
					Name:     "Hauken's Insight",
					Required: true,
				}.Apply(effects.NewContext(g, item))
			}),
		},
		CastPermissions: []game.CastPermission{{
			Zone:    game.ZoneExile,
			Filter:  game.PermissionFilter{ExiledWithSource: true},
			PerType: []string{game.PermissionTypePermanent},
			Timing:  game.TimingYourTurnOnly,
			Cost:    "{0}",
			Label:   "Once during each of your turns, play a land or cast a spell from among the cards exiled with this permanent without paying its mana cost (Hauken's Insight)",
		}},
	})
}

func jacobHaukenResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	var hand []uuid.UUID
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
		}
	}
	if len(hand) == 0 {
		return jacobHaukenMayTransform(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   item.Controller,
		Source:    item.SourceCardID,
		Question:  "Jacob Hauken — exile a card from your hand face down",
		Cards:     hand,
		Min:       1,
		Max:       1,
		LiveZone:  game.ZoneHand,
		ThenKey:   jacobHaukenExiled,
		Carry:     []uuid.UUID{item.SourceCardID},
		CarryNums: []int{item.SourceEpoch},
		Asked:     game.NewAskedItem(item),
	})
	return nil
}

// jacobHaukenExiled exiles the chosen card face down, linked to Jacob Hauken
// as the object the ability came from, then offers the payment. An on-disk
// key: never renamed, never reused.
var jacobHaukenExiled = effects.AskedPick("choose-cards/s82-vow-b15-jacob-hauken-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) > 0 && len(r.Carry) > 0 && len(r.Nums) > 0 {
		src := game.PermissionCardRef{ID: r.Carry[0], Epoch: r.Nums[0]}
		if _, err := ctx.Game.ExileHiddenLinkedForEffect(src, r.Chooser, r.Picked[0]); err != nil {
			return err
		}
	}
	return jacobHaukenMayTransform(ctx)
})

func jacobHaukenMayTransform(ctx *effects.Context) error {
	return effects.MayPay{
		Chooser:  ctx.Controller(),
		Cost:     "{4}{U}{U}",
		Question: "Jacob Hauken — pay {4}{U}{U} to transform Jacob Hauken?",
		Then:     effects.AshlingTransformPaid,
	}.Apply(ctx)
}
