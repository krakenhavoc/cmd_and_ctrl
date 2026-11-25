package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Helix — Sorcery {3}{G}{U} (Strixhaven: School of Mages, #5502):
//
//	"Return target permanent to its owner's hand. Return target permanent card
//	 from your graveyard to your hand."
//
// Two targets of two clauses, each re-checked as the spell resolves (CR
// 608.2b): the other half still happens if one target has become illegal, and
// the spell does nothing only if both have. The permanent returns to its
// OWNER's hand whoever controls it; the graveyard card is a permanent card
// (not an instant or sorcery) in your graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c83979a-b354-4e20-9638-85e8f354da81",
		Name:         "Aether Helix",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetPermanent("target permanent"),
			effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			tail := ctx.ItemTail(aetherHelixReturnGraveyardCard)
			first, ok := ctx.ClauseTarget(0)
			if !ok || first.Kind != game.TargetCard {
				return tail.Run(ctx.Game, nil)
			}
			return effects.BounceToHand{Target: first.ID, ThenTail: tail}.Apply(ctx)
		},
	})
}

// aetherHelixReturnGraveyardCard is the second sentence, once the first
// target's bounce has settled: the graveyard card goes to your hand if it is
// still a legal target (CR 608.2b). An on-disk key.
var aetherHelixReturnGraveyardCard = effects.CardTail("old-b044/aether-helix-return", func(ctx *effects.Context, _ game.TailOutcome) error {
	t, ok := ctx.ClauseTarget(1)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
})
