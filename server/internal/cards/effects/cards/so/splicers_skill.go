package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splicer's Skill — Sorcery {2}{W}:
//
//	"Create a 3/3 colorless Phyrexian Golem artifact creature token.
//	Splice onto instant or sorcery {3}{W}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96fe9bab-0ffa-4d8a-81b7-a48fd532d3f2",
		Name:         "Splicer's Skill",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoInstantOrSorcery("{3}{W}"),
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), N: 1, Template: game.Card{
				Name: "Phyrexian Golem", TypeLine: "Token Artifact Creature — Phyrexian Golem", Power: 3, Toughness: 3,
			}}.Apply(ctx)
		},
	})
}
