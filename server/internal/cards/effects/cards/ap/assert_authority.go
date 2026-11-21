package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assert Authority — Instant {5}{U}{U} (Mirrodin, #5606):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Counter target spell. If that spell is countered this way, exile it
//	 instead of putting it into its owner's graveyard."
//
// Affinity is Myr Enforcer's self cost modifier (CR 702.41a). The
// counter sends the spell to exile through CounterTarget.Dest, Force of
// Negation's shape: a spell that can't be countered stays on the stack
// and is not exiled either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "085bc7be-bd44-40af-8e8f-a1a8006fe22c",
		Name:         "Assert Authority",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		Targets:   effects.TargetSpell("target spell"),
		OnResolve: effects.CounterTargetSpellAndExileIt,
	})
}
