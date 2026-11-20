package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nettle Guard — Creature — Mouse Soldier {1}{W}, 3/1 (Bloomburrow, #5230):
//
//	"Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, it gets +0/+2 until
//	 end of turn.
//	 {1}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// The shared valiant trigger (valiant.go) with the until-end-of-turn
// self-pump, which does nothing to a Guard that left and came back (CR
// 400.7). The activated ability is Cathar Commando's: the sacrifice is a
// cost, so the destroy still happens after the Guard is gone.
//
// No simplification.
func init() {
	const pump = "Nettle Guard — Valiant: it gets +0/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "9ac2dcb6-dacc-4fcb-b820-9fc6744ec503",
		Name:         "Nettle Guard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Valiant(pump, effects.ThisGetsUntilEndOfTurn(0, 2, pump)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
