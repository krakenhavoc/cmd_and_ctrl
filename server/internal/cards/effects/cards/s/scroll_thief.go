package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scroll Thief — 1/3 Creature — Merfolk Rogue for {2}{U}:
//
//	"Whenever Scroll Thief deals combat damage to a player, draw a
//	card."
//
// S19 sub-PR 7: the self-only shape — the source of the damage must
// be the Thief itself (ev.Source == source.InstanceID), not just any
// creature its controller has. Mandatory, no prompt.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "637c5583-4683-4ae4-8b4e-f5da42a772c7",
		Name:         "Scroll Thief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Scroll Thief — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
