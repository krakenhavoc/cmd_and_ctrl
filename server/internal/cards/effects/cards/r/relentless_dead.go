package cardsr

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Relentless Dead — Creature — Zombie {B}{B}, 2/2 (Secret Lair Drop):
//
//	"Menace
//	 When this creature dies, you may pay {B}. If you do, return it to its
//	 owner's hand.
//	 When this creature dies, you may pay {X}. If you do, return another target
//	 Zombie creature card with mana value X from your graveyard to the
//	 battlefield."
//
// Two death triggers. The first is the shared pay-to-return-to-hand
// (MayPayForGraveyardReturnToHand), checked against the same graveyard object.
// The second targets the Zombie as it goes on the stack; the X is the payment
// that makes the target's mana value legal, so the offer is {N} for a target
// of mana value N, and it returns the card only if it is still a Zombie
// creature card of that mana value in the graveyard (CR 608.2b).
func init() {
	toHand := effects.WhenThisDies("Relentless Dead — you may pay {B}; if you do, return it to its owner's hand",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{B}",
				Question: "Relentless Dead — pay {B} to return it to its owner's hand?",
				Then:     relentlessDeadToHand,
			}.Apply(effects.NewContext(g, item))
		})
	zombie := effects.Targeting(effects.WhenThisDies("Relentless Dead — you may pay {X}; if you do, return another target Zombie creature card with mana value X from your graveyard to the battlefield",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			ids := effects.LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			c, ok := g.LookupCardForEffect(ids[0])
			if !ok {
				return nil
			}
			mv := c.ManaValue()
			cost := "{" + strconv.Itoa(mv) + "}"
			if mv == 0 {
				cost = "{0}"
			}
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     cost,
				Question: "Relentless Dead — pay " + cost + " to return " + c.Name + " to the battlefield?",
				Then:     relentlessDeadPaid,
				Carry:    []uuid.UUID{ids[0]},
			}.Apply(ctx)
		}),
		func() *game.TargetSpec {
			spec := effects.TargetCardInGraveyard("another target Zombie creature card with mana value X from your graveyard",
				effects.YouOwn(), effects.Creature(), effects.HasSubtype("Zombie"))
			spec.ExcludeSource = true
			return spec
		}())
	effects.Register(effects.Spec{
		OracleID:        "7404c078-228b-4296-bf1f-62f57bf832d9",
		Name:            "Relentless Dead",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Purpose:         game.Purpose{Answers: game.AnswerGraveyard},
		Triggered:       []game.TriggeredAbility{toHand, zombie},
	})
}

// relentlessDeadPaid is the payment's branch: the Zombie comes back if it is
// still a Zombie creature card in its owner's graveyard. On-disk key.
var relentlessDeadPaid = effects.PayUnlessBranches("pay-unless/relentless-dead-zombie",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		for _, id := range r.Carry {
			z := ctx.Game.FindCardZoneForEffect(id)
			if z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !c.IsCreature() || !c.HasSubtype(ctx.TextSubtype("Zombie")) {
				continue
			}
			return effects.ReturnFromGraveyardTogether{Targets: []uuid.UUID{id}, Controller: ctx.Controller()}.Apply(ctx)
		}
		return nil
	}, nil)

// relentlessDeadToHand is the {B} payment's branch: the card that died goes
// from the graveyard to its owner's hand, unless it has left the graveyard
// since it died (a card that did is a new object, CR 400.7). On-disk key.
var relentlessDeadToHand = effects.PayUnlessBranches("pay-unless/relentless-dead-to-hand",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		self := ctx.Source()
		if z := ctx.Game.FindCardZoneForEffect(self); z == nil || z.Kind != game.ZoneGraveyard || ctx.IsNewSourceObjectAsThis(self) {
			return nil
		}
		return effects.ReturnThisCardFromYourGraveyardToHand(ctx.Game, ctx.Item)
	}, nil)
