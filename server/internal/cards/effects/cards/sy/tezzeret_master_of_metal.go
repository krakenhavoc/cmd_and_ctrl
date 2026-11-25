package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret, Master of Metal — Legendary Planeswalker — Tezzeret {4}{U}{B},
// loyalty 5 (Aether Revolt, S82 older sets, #5536):
//
//	"+1: Reveal cards from the top of your library until you reveal an
//	 artifact card. Put that card into your hand and the rest on the bottom
//	 of your library in a random order.
//	 −3: Target opponent loses life equal to the number of artifacts you
//	 control.
//	 −8: Gain control of all artifacts and creatures target opponent
//	 controls."
//
// The −3 counts artifacts as it resolves (CR 608.2h); the −8 takes every
// artifact and creature the opponent controls as it resolves, and the control
// change has no duration (CR 611.2a), so it outlasts Tezzeret. A target
// opponent who has left stops both (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3fba19a9-44c6-4771-9587-9d2c4009e996",
		Name:            "Tezzeret, Master of Metal",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Reveal cards from the top of your library until you reveal an artifact card. Put that card into your hand and the rest on the bottom of your library in a random order.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool { return c.IsArtifact() },
						"Tezzeret, Master of Metal — revealed until an artifact card")
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  run,
						Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
						All:    true,
						Label:  "Tezzeret, Master of Metal — put the artifact card into your hand",
						Key:    effects.TakeRestBottomRandom,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Target opponent loses life equal to the number of artifacts you control.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.CountControlled(g, ctx.Controller(), effects.MatchArtifact)
					if n <= 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -n)
						}
					}
					return nil
				},
			},
			{
				Label:   "−8: Gain control of all artifacts and creatures target opponent controls.",
				Cost:    effects.LoyaltyCost(-8),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item).AsGroupMember()
					var opp uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							opp = t.ID
						}
					}
					if opp == uuid.Nil {
						return nil
					}
					var taken []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == opp && (c.IsArtifact() || c.IsCreature()) {
							taken = append(taken, c.InstanceID)
						}
					}
					for _, id := range taken {
						if err := (effects.GainControl{
							Target:   id,
							Duration: game.IndefiniteDuration(),
							Label:    "Tezzeret, Master of Metal — gain control",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
