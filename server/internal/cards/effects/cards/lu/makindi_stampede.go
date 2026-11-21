package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Makindi Stampede // Makindi Mesas — Modal DFC. Front: Sorcery
// {3}{W}{W}: "Creatures you control get +2/+2 until end of turn." Back:
// Land (mdfc_lands.go).
//
// The set is the creatures you control as it resolves, pinned there
// (CR 611.2c): one that enters later this turn does not get the bonus.
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "342e08f9-d4d0-4408-8621-66e087058616",
		Name:         "Makindi Stampede",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match:     effects.And(effects.Creature(), effects.YouControl()),
				Power:     2,
				Toughness: 2,
				Label:     "Makindi Stampede — +2/+2",
			}.Apply(ctx)
		},
	})
}
