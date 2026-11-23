package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scathe Zombies — Creature — Zombie {2}{B}, 2/2
// (Limited Edition Alpha, #5642):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0fefaf0-da20-4d58-8db7-019dba16c780",
		Name:         "Scathe Zombies",
		Completeness: effects.CompletenessFull,
	})
}
