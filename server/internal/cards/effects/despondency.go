package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Despondency — Enchantment — Aura {1}{B} (Urza's Saga):
//
//	"Enchant creature
//	 Enchanted creature gets -2/-0.
//	 When this Aura is put into a graveyard from the battlefield, return
//	 it to its owner's hand."
//
// Brilliant Halo's return on a -2/-0 Aura. The trigger watches the
// Aura's own move from the battlefield to a graveyard, however it got
// there (its creature died, it was destroyed, or it fell off as a
// state-based action). A Despondency that has left the graveyard before
// the trigger resolves stays where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a68ac1ba-f51e-446c-abf0-7a2bd4d65a8c",
		Name:         "Despondency",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			PumpAttached(-2, 0),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Despondency — return it to its owner's hand", returnThisCardFromYourGraveyardToHand),
		},
	})
}
