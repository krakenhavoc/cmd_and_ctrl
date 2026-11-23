package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gudul Lurker — Creature — Salamander {U}, 1/1 (Dragons of Tarkir,
// #5551):
//
//	"This creature can't be blocked.
//	 Megamorph {U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)"
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// "Can't be blocked" is the self restriction Deep-Sea Kraken uses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "bde976c6-afed-44bb-8e3d-4103289dff5e",
		Name:             "Gudul Lurker",
		Completeness:     effects.CompletenessFull,
		Static:           []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{U}")},
	})
}
