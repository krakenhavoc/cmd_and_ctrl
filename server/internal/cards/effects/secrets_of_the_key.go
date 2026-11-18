package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secrets of the Key — Instant {U} (Midnight Hunt, #5497):
//
//	"Investigate. If this spell was cast from a graveyard, investigate
//	 twice instead.
//	 Flashback {3}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// "Cast from a graveyard" reads the spell's announce record
// (StackItem.CastFromZone), so the flashback cast counts and so does any
// other cast out of a graveyard. "Instead" means two Clues, not three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "55c7ce94-3cd3-42f3-9dd8-0c118fa626c8",
		Name:             "Secrets of the Key",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 1
			if item.CastFromZone == game.ZoneGraveyard {
				n = 2
			}
			return Investigate{N: n}.Apply(ctx)
		},
	})
}
