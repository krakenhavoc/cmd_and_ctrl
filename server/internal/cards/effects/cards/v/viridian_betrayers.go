package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Viridian Betrayers — Creature — Phyrexian Elf Warrior {1}{G}{G}, 3/1 (New
// Phyrexia, #5575):
//
//	"This creature has infect as long as an opponent is poisoned. (It deals
//	 damage to creatures in the form of -1/-1 counters and to players in the
//	 form of poison counters.)"
//
// A layer 6 keyword grant on the Betrayers alone (CR 613.1f), re-answered on
// every recompute: an opponent is poisoned with one or more poison counters
// (CR 122.1f), so it gains infect when any opponent first gets one and loses it
// if they all lose theirs.
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(effects.WhileYouHave(func(g *game.Game, you uuid.UUID) bool {
		return effects.S82OneB04OpponentPoison(g, you) > 0
	}, effects.SelfCreature), "infect")
	grant.Label = "This creature has infect as long as an opponent is poisoned."
	effects.Register(effects.Spec{
		OracleID:     "1d5ebb4b-0248-4255-8e1a-bcba4d5eaf7f",
		Name:         "Viridian Betrayers",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
