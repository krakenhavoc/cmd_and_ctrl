package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Call the Skybreaker — Sorcery {5}{U/R}{U/R}:
//
//	"Create a 5/5 blue and red Elemental creature token with flying.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: see Spitting Image. The hybrid symbols
// are paid at their printed cost in either colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a7213bef-0e9c-44e1-97cb-d1da0ef370bf",
		Name:             "Call the Skybreaker",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{5}{U/R}{U/R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("5/5 blue and red Elemental with flying"), N: 1}.Apply(ctx)
		},
	})
}
