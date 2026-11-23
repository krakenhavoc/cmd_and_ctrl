package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Great Intelligence's Plan — Sorcery {4}{U}{B}:
//
//	"Draw three cards. Then target opponent faces a villainous choice —
//	 They discard three cards, or you may cast a spell from your hand
//	 without paying its mana cost."
//
// A villainous choice (CR 701.55, #3140), faced by the targeted opponent
// after the controller has drawn. The discard is the facing player's own
// pick (three, or the whole hand when it holds fewer). The second option
// is the CONTROLLER's "you may": a choose-cards pick over the nonland
// cards in their hand with a floor of zero, then the free cast every
// "cast a spell from your hand without paying its mana cost" offers
// (offerFreeHandCast, #3270): the client opens the cast chain at once,
// timing is ignored (CR 608.2g), and passing priority ends the chance.
// The controller's hand is hidden, so the hint a bot sees on that option
// is a fixed guess and not a read of it.
//
// No simplifications.
var greatIntelligencesPlanChoice = effects.NewVillainousChoice("great-intelligences-plan",
	"Great Intelligence's Plan — discard three cards, or the controller may cast a spell for free?",
	effects.VillainousOption{
		Label:    "Discard three cards",
		Harm:     3,
		Possible: func(g *game.Game, facing uuid.UUID) bool { return effects.VillainousHandSize(g, facing) > 0 },
		Body:     effects.VillainousDiscard(3),
	},
	effects.VillainousOption{
		Label: "The controller of Great Intelligence's Plan may cast a spell from their hand without paying its mana cost",
		Harm:  3,
		Body: func(ctx *effects.Context, _ effects.Facing) error {
			effects.OfferFreeHandCast(ctx, ctx.Controller(), greatIntelligencesPlanFreeCast, "a spell", effects.AnySpell)
			return nil
		},
	})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb9abc41-07f8-4a06-b7b9-9dd6be526e42",
		Name:         "Great Intelligence's Plan",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		// The choice waits for a prompt the draws ask (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 3},
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.FaceVillainousChoice{
							Players: []uuid.UUID{t.ID},
							Choice:  greatIntelligencesPlanChoice,
						}.Apply(ctx)
					}
				}
				return nil
			}),
		),
	})
}

// greatIntelligencesPlanFreeCast is the offer's keyed continuation (#4246).
var greatIntelligencesPlanFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/great-intelligences-plan-cast", "Great Intelligence's Plan")
