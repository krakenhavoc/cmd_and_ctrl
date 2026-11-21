package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gilded Ghoda — Creature — Horse Mount {1}{R}:
//
//	"Whenever this creature attacks while saddled, create a Treasure
//	 token.
//	 Saddle 1"
//
// Saddle is Saddle(1) (CR 702.171, ADR 0071 amendment 2026-10-08); the
// attack trigger reads the saddled designation when the attack is
// declared. The Treasure is the catalog token, so it carries its mana
// ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f4939dc-6e87-47fa-8287-9da60d4b5db2",
		Name:         "Gilded Ghoda",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(1)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Gilded Ghoda — create a Treasure", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
