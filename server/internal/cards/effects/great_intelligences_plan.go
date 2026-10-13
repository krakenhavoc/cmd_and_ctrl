package effects

import (
	"github.com/google/uuid"

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
// cards in their hand with a floor of zero, then a one-shot {0} cast
// permission over the card picked, which is the shape Rishkar's and
// Kari Zev's Expertise use for the same sentence and carries the same
// timing (the card is cast through the ordinary cast action, with its
// own restrictions). The controller's hand is hidden, so the hint a bot
// sees on that option is a fixed guess and not a read of it.
//
// No simplifications.
var greatIntelligencesPlanChoice = NewVillainousChoice("great-intelligences-plan",
	"Great Intelligence's Plan — discard three cards, or the controller may cast a spell for free?",
	VillainousOption{
		Label:    "Discard three cards",
		Harm:     3,
		Possible: func(g *game.Game, facing uuid.UUID) bool { return villainousHandSize(g, facing) > 0 },
		Body:     villainousDiscard(3),
	},
	VillainousOption{
		Label: "The controller of Great Intelligence's Plan may cast a spell from their hand without paying its mana cost",
		Harm:  3,
		Body: func(ctx *Context, _ Facing) error {
			controller := ctx.Controller()
			var candidates []uuid.UUID
			for _, id := range allHandCardIDs(ctx.Game, controller) {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
					candidates = append(candidates, id)
				}
			}
			if len(candidates) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  controller,
				Source:   ctx.Source(),
				Question: "Great Intelligence's Plan — cast a spell from your hand without paying its mana cost?",
				Cards:    candidates,
				Min:      0,
				Max:      1,
				Zone:     game.ZoneHand,
				Then: func(g *game.Game, picked []uuid.UUID) error {
					if len(picked) == 0 {
						return nil
					}
					g.GrantCastPermissionOverCardForEffect(picked[0], game.CastPermission{
						Player:   controller,
						Cost:     "{0}",
						CastOnly: true,
						Label:    "Great Intelligence's Plan — cast it without paying its mana cost",
					})
					return nil
				},
			})
			return nil
		},
	})

func init() {
	Register(Spec{
		OracleID:     "cb9abc41-07f8-4a06-b7b9-9dd6be526e42",
		Name:         "Great Intelligence's Plan",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{N: 3}).Apply(ctx); err != nil {
				return err
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return FaceVillainousChoice{
						Players: []uuid.UUID{t.ID},
						Choice:  greatIntelligencesPlanChoice,
					}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
