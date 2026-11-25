package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Elderspell — Sorcery {B}{B} (War of the Spark):
//
//	"Destroy any number of target planeswalkers. Choose a planeswalker you control.
//	 Put two loyalty counters on it for each planeswalker destroyed this way."
//
// The planeswalkers are destroyed together as one event, and the count is of
// those really destroyed (an indestructible one stays, CR 701.8b), told to the
// continuation. The choice is not a target (CR 115.10): the caster picks one of
// the planeswalkers they control then, the destroyed ones included if they are
// gone, none of them if they are all gone, and with none destroyed or none left to
// choose nothing is chosen. The loyalty goes on through the counter pipeline.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eff0a1ad-ecc6-416f-a4ae-96b26cd8a905",
		Name:         "The Elderspell",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("any number of target planeswalkers", effects.Planeswalker()).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard && t.ID != uuid.Nil {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect(ids, ctx.ItemTail(theElderspellDestroyed))
		},
	})
}

// theElderspellDestroyed asks for the planeswalker once the destruction is
// settled. An on-disk key: never renamed, never reused.
var theElderspellDestroyed = effects.CardTail("then/s82-old-b167-the-elderspell-destroyed", func(ctx *effects.Context, o game.TailOutcome) error {
	n := len(o.IDs)
	if n == 0 {
		return nil
	}
	return effects.ChoosePermanents{
		Question: "The Elderspell — choose a planeswalker you control",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := effects.PermanentsControlledByMatching(g, of, effects.Planeswalker())
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 1, 1
		},
		Keyed: theElderspellPicked,
		Nums:  []int{n},
	}.Apply(ctx)
})

var theElderspellPicked = effects.RegisterPermanentsPick("prompt-run/s82-old-b167-the-elderspell-loyalty", effects.PermanentsDef{
	ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
		cards := picked.Cards()
		if len(cards) == 0 || len(r.Nums) == 0 {
			return nil
		}
		return effects.AddCounter{Target: cards[0], Kind: game.CounterLoyalty, N: 2 * r.Nums[0]}.Apply(ctx)
	},
})
