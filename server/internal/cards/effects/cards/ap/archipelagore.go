package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archipelagore — Creature — Leviathan {5}{U}{U}, 7/7 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {5}{U} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Whenever this creature mutates, tap up to X target creatures, where X is
//	 the number of times this creature has mutated. Those creatures don't untap
//	 during their controller's next untap step."
//
// Mutate is the engine's (ADR 0223) and the "up to X" is Midnight Arsonist's:
// built as the trigger goes on the stack (CR 601.2c), X being the merged
// permanent's count of mutations then (S82OldB112TimesMutated). Every target
// still legal as it resolves is tapped and held through its controller's next
// untap step (Dragon Turtle's TapAndFreeze).
//
// No simplification.
func init() {
	mutates := effects.On(game.EventMutate, effects.ThisCreatureMutates,
		"Archipelagore — tap up to X target creatures; they don't untap during their controller's next untap step",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return effects.TapAndFreeze{Targets: ids, Label: "Archipelagore — it doesn't untap during its controller's next untap step"}.Apply(ctx)
		})
	mutates.TargetsFrom = func(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		x := effects.S82OldB112TimesMutated(g, source.InstanceID)
		return effects.Harmful(effects.TargetCreature("up to X target creatures").WithCount(0, x))
	}
	effects.Register(effects.Spec{
		OracleID:         "ae1df2ad-f34c-4569-92f2-98be776c612d",
		Name:             "Archipelagore",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRestrict},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{5}{U}")},
		Triggered:        []game.TriggeredAbility{mutates},
	})
}
