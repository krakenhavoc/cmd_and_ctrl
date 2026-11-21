package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cryptic Gateway — Artifact {5} (Dominaria Remastered):
//
//	"Tap two untapped creatures you control: You may put a creature card
//	 from your hand that shares a creature type with each creature tapped
//	 this way onto the battlefield."
//
// The cost is a TapOthers component of two untapped creatures you control,
// which is not the {T} symbol, so summoning sickness does not stop one
// paying (CR 302.6). The two that paid ride the ability's payment record
// (ctx.TappedToPay). On resolution the controller may put a creature card
// from their hand that shares a creature type with EACH of them
// (CR 205.3m; game.SharesCreatureType, so a changeling shares with any
// creature that has a type). A creature that left the battlefield in
// response is read as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abdc1c92-1434-47d3-abaa-debe047903b9",
		Name:         "Cryptic Gateway",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Label:   "Tap two untapped creatures you control: You may put a creature card from your hand that shares a creature type with each creature tapped this way onto the battlefield.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetCreature("two untapped creatures you control"),
				Label:  "two untapped creatures you control",
			}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				tapped := ctx.TappedToPay()
				if len(tapped) == 0 {
					return nil
				}
				return effects.PutFromHandOntoBattlefield{
					Match: func(g *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsCreature() && crypticGatewaySharesWithEach(g, tapped, c)
					},
					Optional: true,
					Label:    "Cryptic Gateway — you may put a creature card that shares a creature type with each tapped creature onto the battlefield",
				}.Apply(ctx)
			},
		}},
	})
}

// crypticGatewaySharesWithEach reports whether `c` shares a creature type
// with every creature that was tapped to pay, each as it is now or, once it
// has left the battlefield, as it last existed.
func crypticGatewaySharesWithEach(g *game.Game, tapped []game.ObjectRef, c game.Card) bool {
	for _, ref := range tapped {
		if live, ok := g.LookupCardForEffect(ref.ID); ok && effects.OnBattlefield(g, ref.ID) && g.ObjectEpochForEffect(ref.ID) == ref.Epoch {
			if !game.SharesCreatureType(&live, &c) {
				return false
			}
			continue
		}
		info, ok := g.LastKnownPermanentForEffect(ref.ID)
		if !ok || !effects.SharesACreatureTypeWith(info, c) {
			return false
		}
	}
	return true
}
