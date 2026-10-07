package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "9f36dd20-350c-4046-b57f-e6b5cc9aa999",
		Name:             "Embrace the Unknown",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{2}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b19ExileTopTwoUntilEndOfNextTurn(ctx.Game, item)
		},
	})
}
