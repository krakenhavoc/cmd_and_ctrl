package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smashing Success — Instant {3}{R}:
//
//	"Destroy target artifact or land. If an artifact is destroyed this way,
//	 create a Treasure token. (It's an artifact with "{T}, Sacrifice this
//	 token: Add one mana of any color.")"
//
// Wreak Havoc's destroy over an artifact or a land. The Treasure is made only
// when the target was an artifact as it was destroyed and it actually left: an
// indestructible or regenerated artifact stays and makes nothing, and a land
// that is not an artifact (a Treasure-less land) never does. A target gone by
// resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb19d2f7-0511-4c66-ba4c-f38516920c20",
		Name:         "Smashing Success",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets:      effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !effects.OnBattlefield(ctx.Game, id) {
				return nil
			}
			// "If an artifact is destroyed this way": the destroyed list the tail is
			// told (CR 608.2c) says it really left; N says it was an artifact as it was.
			n := 0
			if c.IsArtifact() {
				n = 1
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, ctx.TailN(smashingSuccessDestroyed, n))
		},
	})
}

// smashingSuccessDestroyed is "if an artifact is destroyed this way, create a
// Treasure token", told the permanents that were destroyed (N is 1 when the
// target was an artifact). An on-disk key: never renamed, never reused.
var smashingSuccessDestroyed = effects.CardTail("destroy-then/smashing-success-treasure", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N != 1 || len(o.IDs) != 1 {
		return nil
	}
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TreasureToken(), N: 1}.Apply(ctx)
})
