package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resuscitate — Instant {1}{G} (Exodus, S82 OLD batch 742, #5626):
//
//	"Until end of turn, creatures you control gain "{1}: Regenerate this
//	 creature.""
//
// Arm with Aether's mass grant: the creatures are fixed as the spell resolves
// (CR 611.2c) and each carries the ability for the turn, with its own source,
// so each regenerates itself (CR 701.19a) for {1}.
//
// No simplification.
func init() {
	const grant = "resuscitate/regenerate"
	const text = "{1}: Regenerate this creature."
	effects.Register(effects.Spec{
		OracleID:     "4390e251-e154-4ef8-9929-0c78aa2c8f10",
		Name:         "Resuscitate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		Grants: []effects.AbilityGrant{{
			Key:       grant,
			Activated: []effects.ActivatedAbility{effects.RegenerateThisRow("{1}")},
			Text:      text,
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match: effects.And(effects.Creature(), effects.YouControl()),
				Keys:  []string{grant},
				Label: "Resuscitate — creatures you control gain \"" + text + "\"",
			}.Apply(ctx)
		},
	})
}
