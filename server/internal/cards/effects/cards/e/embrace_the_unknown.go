package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embrace the Unknown — Sorcery {2}{R}:
//
//	"Exile the top two cards of your library. Until the end of your next turn, you
//	 may play those cards.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: `CastableZones` opens the graveyard and
// `Retrace(cost)` prices it (ADR 0066, 2026-10-07 amendment). The spell goes
// back to the graveyard when it resolves, so it can be retraced again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9f36dd20-350c-4046-b57f-e6b5cc9aa999",
		Name:             "Embrace the Unknown",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Retrace("{2}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B19ExileTopTwoUntilEndOfNextTurn(ctx.Game, item)
		},
	})
}
