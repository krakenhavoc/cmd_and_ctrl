package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Welder — {R}, 1/1 Goblin Artificer.
// {T}: Choose target artifact a player controls and target artifact card
// in that player's graveyard. If both targets are still legal as this
// ability resolves, that player simultaneously sacrifices the artifact
// and returns the artifact card to the battlefield.
func init() {
	Register(Spec{
		OracleID:     "9fc28e1f-f4ea-4bf8-9442-a36dba3b3b29",
		Name:         "Goblin Welder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: Exchange an artifact a player controls with an artifact card in their graveyard.",
			Cost:  TapCost(),
			Targets: Clauses(
				TargetPermanent("target artifact a player controls", Artifact()),
				InGraveyardOfTargetsController(TargetCardInGraveyard("target artifact card in that player's graveyard", Artifact()), 0),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				a, aOK := ctx.ClauseTarget(0)
				b, bOK := ctx.ClauseTarget(1)
				if !aOK || !bOK || a.Kind != game.TargetCard || b.Kind != game.TargetCard || !ctx.IsTargetLegal(a) || !ctx.IsTargetLegal(b) {
					return nil
				}
				// #1997: "that player … returns" the card, so they put
				// it (CR 110.2a), which Arboria reads.
				return g.PutByEnteringPlayerForEffect(func() error {
					return g.ExchangeArtifactWithGraveyardForEffect(item.SourceCardID, a.ID, b.ID)
				})
			},
		}},
	})
}
