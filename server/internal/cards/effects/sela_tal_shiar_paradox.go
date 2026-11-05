package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sela, Tal Shiar Paradox — Legendary Creature — Romulan Human Rogue
// {3}{B}, 3/3 (Star Trek, #4973):
//
//	"Whenever Sela attacks, exile up to one target card from a graveyard.
//	 Whenever one or more cards leave your graveyard, create a tapped 2/2
//	 black Romulan creature token."
//
// The attack trigger's target is chosen as it goes on the stack and
// re-checked as it resolves (CR 608.2b). The second trigger fires once per
// batch of cards leaving your graveyard together (CR 603.2c), whatever
// moved them and wherever they went, Sela's own exile included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0cbf0f7a-7f5e-4819-9a9f-f655f92fd5f1",
		Name:         "Sela, Tal Shiar Paradox",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Sela, Tal Shiar Paradox — exile up to one target card from a graveyard",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						targets := ctx.LegalTargets()
						if len(targets) == 0 {
							return nil
						}
						return ExileTarget{Target: targets[0].ID}.Apply(ctx)
					}),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
			OncePerBatch(OnAny(leftGraveyardEvents, whenCardsLeaveYourGraveyard,
				"Sela, Tal Shiar Paradox — create a tapped 2/2 black Romulan",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenAdvanced{
						Controller: item.Controller,
						Spec:       Token(TokenCard("2/2 black Romulan")).EntersTapped(),
						N:          1,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
