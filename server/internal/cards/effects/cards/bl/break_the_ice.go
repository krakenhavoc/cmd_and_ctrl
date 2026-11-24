package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Break the Ice — Sorcery {B}{B} (Modern Horizons 2):
//
//	"Destroy target land that is snow or could produce {C}.
//	 Overload {4}{B}{B} (You may cast this spell for its overload cost. If you do,
//	 change "target" in its text to "each.")"
//
// "Could produce {C}" is CR 106.7's reader (the mana the land's abilities could
// make right now, so a Mutavault or a Wastes does, and a Forest does not unless
// it is snow). The overloaded spell has no target and destroys every such
// land on the battlefield as it resolves, those of every player.
//
// No simplification.
func init() {
	pred := effects.OldB035SnowOrCouldProduceColorless()
	effects.Register(effects.Spec{
		OracleID:     "22ec98d2-0ac5-48a2-b48b-c6a98f35ecf0",
		Name:         "Break the Ice",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target land that is snow or could produce {C}", pred),
		AlternativeCosts: []game.AlternativeCost{
			effects.Overload("{4}{B}{B}"),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.PaidAltCost("overload") {
				return effects.DestroyAllMatching{Match: pred}.Apply(ctx)
			}
			return effects.DestroyTheTargetPermanent(item, ctx)
		},
	})
}
