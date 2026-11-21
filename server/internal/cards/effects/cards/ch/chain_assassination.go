package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chain Assassination — Instant {2}{B}{B}:
//
//	"Freerunning {1}{B} (You may cast this spell for its freerunning
//	 cost if you dealt combat damage to a player this turn with an
//	 Assassin or commander.)
//	 Destroy target creature. If another creature died this turn, draw
//	 a card."
//
// "Another creature" is a creature other than the one this spell
// destroys, so the creature destroyed is taken off the turn's tally of
// creatures that died once the destroy is done: only another creature
// counts, and the target's own death does not. The tally is the last-known-information one
// Mahadi uses, so a token that died counts. Freerunning is Eagle
// Vision's alternative cost.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "37addb87-234a-402f-b7fb-f88ef3f2a589",
		Name:             "Chain Assassination",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Freerunning("{1}{B}", "Assassin")},
		Targets:          effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, ctx.ItemTail(chainAssassinationDrawTail))
		},
	})
}

// chainAssassinationDrawTail is "If another creature died this turn, draw a
// card", run once the destroy is done (so a destroy a replacement
// cancelled or has not answered yet draws nothing early). The turn's tally
// of creatures that died counts the creature this spell destroyed, so the
// creatures destroyed this way are taken off it: only another creature
// counts. An on-disk key.
var chainAssassinationDrawTail = effects.CardTail("then/chain-assassination-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	if effects.B11CreaturesDiedThisTurn(ctx.Game)-len(o.IDs) <= 0 {
		return nil
	}
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
