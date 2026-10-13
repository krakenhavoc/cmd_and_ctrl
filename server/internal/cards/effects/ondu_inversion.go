package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ondu Inversion // Ondu Skyruins — modal double-faced card. This file
// is the FRONT face, Sorcery {6}{W}{W}:
//
//	"Destroy all nonland permanents."
//
// Planar Cleansing's text at eight mana, with the land on the back.
// The back face, Ondu Skyruins ("enters tapped, {T}: Add {W}"), was
// registered with the MDFC land cycle in mdfc_lands.go and nothing
// registered this face, so the spell half did nothing when cast.
// Everything goes as one event (CR 700.4), the caster's board
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15fc4e74-300e-4c2d-8ed7-004553b2f7c2",
		Name:         "Ondu Inversion",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Nonland()}.Apply(ctx)
		},
	})
}
