package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Artifist Acumen — Sorcery {R}:
//
//	"Creatures you control gain first strike until end of turn.
//	 Draw a card."
//
// The set of creatures is fixed as the spell resolves (CR 611.2c), so a
// creature that enters later this turn does not get first strike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "129b41cf-8327-42f0-b186-1111b42c08ab",
		Name:         "Artifist Acumen",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"first strike"},
				Label:    "Artifist Acumen — first strike",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
