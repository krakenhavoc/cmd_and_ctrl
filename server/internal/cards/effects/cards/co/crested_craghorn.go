package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crested Craghorn — Creature — Goat Beast, {4}{R}, 4/1:
//
//	"Haste
//	 Provoke (Whenever this creature attacks, you may have target
//	 creature defending player controls untap and block it if able.)"
//
// #1684 leftover: haste rides PrintedKeywords, Provoke rides the
// Provoke() trigger constructor built for Goblin Grappler. No new
// machinery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "597c390e-deeb-4452-aeda-7cf9312a8b01",
		Name:            "Crested Craghorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{effects.Provoke()},
	})
}
