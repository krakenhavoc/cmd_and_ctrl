package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9ac2dcb6-dacc-4fcb-b820-9fc6744ec503",
		Name:         "Nettle Guard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Valiant(pump, thisGetsUntilEndOfTurn(0, 2, pump)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  destroyChosenPermanent,
		}},
	})
}
