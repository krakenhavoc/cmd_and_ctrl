package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost, Spectral Saboteur — {2}{U/B} Legendary Creature — Human Rogue
// Villain 2/2:
//
//	"Flash
//	 Intangibility — Ghost can't be blocked."
//
// The hybrid symbol is part of the mana cost the catalog reads from the
// card data; the evasion is the same self restriction Blighted Agent has.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "88f34d15-c338-40a7-b56c-c5b650847430",
		Name:            "Ghost, Spectral Saboteur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
