package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swarm Surge — Sorcery {2}{B} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Creatures you control get +2/+0 until end of turn. Colorless creatures you
//	 control also gain first strike until end of turn."
//
// Both sets of creatures are fixed as the spell resolves (CR 611.2c): the
// creatures you control then get +2/+0, and those of them that are colorless
// (through the layers) gain first strike. A creature that enters afterwards
// gets neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "77b98632-e103-44fa-a940-2c3b2897b60f",
		Name:            "Swarm Surge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.BoostUntilEOT{Match: mine, Power: 2, Label: "Swarm Surge — +2/+0 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(mine, effects.Colorless()),
				Keywords: []string{"first strike"},
				Label:    "Swarm Surge — colorless creatures gain first strike until end of turn",
			}.Apply(ctx)
		},
	})
}
