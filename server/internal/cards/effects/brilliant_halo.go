package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brilliant Halo — Enchantment — Aura {1}{W} (Urza's Saga):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+2.
//	 When this Aura is put into a graveyard from the battlefield, return
//	 it to its owner's hand."
//
// Rancor's return on a +1/+2 Aura. The trigger watches the Aura's own
// move from the battlefield to a graveyard, however it got there (its
// creature died, it was destroyed, or it fell off as a state-based
// action). A Halo that has left the graveyard before the trigger
// resolves stays where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f68f6d5-1953-4ccd-8af0-7d0a69139556",
		Name:         "Brilliant Halo",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Brilliant Halo — return it to its owner's hand", returnThisCardFromYourGraveyardToHand),
		},
	})
}
