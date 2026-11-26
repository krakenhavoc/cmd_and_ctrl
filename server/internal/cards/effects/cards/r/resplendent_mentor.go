package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resplendent Mentor — Creature — Kithkin Cleric {4}{W}, 2/2 (Shadowmoor, S82
// OLD b529, #5591):
//
//	"White creatures you control have "{T}: You gain 1 life.""
//
// A layer-6 ability grant (ADR 0093, Arcus Acolyte's shape) over the white
// creatures you control, the Mentor included, read on post-layer colour. The
// {T} is the host's own, so a summoning-sick creature can't use it (CR 302.6).
//
// No simplification.
func init() {
	const grant = "resplendent-mentor/tap-gain-1"
	const text = "{T}: You gain 1 life."
	effects.Register(effects.Spec{
		OracleID:     "5b60a257-e4f8-412e-b146-b7983a384753",
		Name:         "Resplendent Mentor",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: text,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
				},
			}},
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.HasColorWord("W")
			}, grant), "White creatures you control have \""+text+"\""),
		},
	})
}
