package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// March of Progress — Sorcery {2}{U} (The Brothers' War Commander):
//
//	"Choose target artifact creature you control. For each creature chosen
//	 this way, create a token that's a copy of it.
//	 Overload {6}{U} (You may cast this spell for its overload cost. If you
//	 do, change its text by replacing all instances of "target" with
//	 "each.")"
//
// The copy is of the copiable values as the spell resolves (CR 707.2), and
// every copy is made at the same moment (CR 608.2f): one
// CreateTokenCopiesTogether, so "one or more tokens" triggers and token
// doublers see a single event. Overloaded, the set is every artifact
// creature the caster controls as the spell resolves, and the cast has no
// target (CR 702.96a). A target that is no longer legal copies nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "dd45c284-c34c-4085-8066-dbeee223fb6e",
		Name:             "March of Progress",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target artifact creature you control", effects.Artifact(), effects.YouControl()),
		AlternativeCosts: []game.AlternativeCost{effects.Overload("{6}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			if ctx.PaidAltCost("overload") {
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsArtifact() && c.IsCreature() {
						ids = append(ids, c.InstanceID)
					}
				}
			} else if id := effects.FirstLegalBattlefieldTarget(ctx); id != uuid.Nil {
				ids = append(ids, id)
			}
			if len(ids) == 0 {
				return nil
			}
			return effects.CreateTokenCopiesTogether{
				Controller: item.Controller,
				Copies:     ids,
			}.Apply(ctx)
		},
	})
}
