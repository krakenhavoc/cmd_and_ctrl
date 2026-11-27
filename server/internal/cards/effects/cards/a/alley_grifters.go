package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alley Grifters — Creature — Human Mercenary {1}{B}{B}, 2/2 (Mercadian Masques, S82 OLD batch 687, #5619):
//
//	"Whenever this creature becomes blocked, defending player discards a card."
//
// Slate Street Ruffian's trigger: the once-per-attacker "becomes blocked" event names the
// defending player as its actor (CR 509.3c), and the discarding player chooses the card
// (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "399af8ea-7a6f-4f43-bd39-dc29a6e75e06",
		Name:         "Alley Grifters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.ThisBecameBlockedEvent,
				"Alley Grifters — defending player discards a card",
				effects.BlockedDefenderDiscardsACard),
		},
	})
}
