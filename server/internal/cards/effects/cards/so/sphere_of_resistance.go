package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphere of Resistance — Artifact {2}:
//
//	"Spells cost {1} more to cast."
//
// The simplest cost modifier there is, and the one worth reading
// first: no predicate at all, so it taxes every spell every player
// casts, including its own controller's. That symmetry is the card —
// a Stax player builds around paying the tax they impose.
//
// "Spells", so land drops are untouched: playing a land is a special
// action (CR 116.2a), not a cast, and the engine's land branch
// returns before any cost is priced.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09c96077-3804-4f12-a613-5bebc5e0413f",
		Name:         "Sphere of Resistance",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "Spells cost {1} more to cast."),
		},
	})
}
