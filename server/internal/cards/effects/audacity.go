package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Audacity — Enchantment — Aura {G} (The Brothers' War):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0 and has trample.
//	 When this Aura is put into a graveyard from the battlefield, draw a card."
//
// The draw is the Aura's own trigger (it leaves the battlefield whether the
// creature died or the Aura was destroyed), so it follows the Aura and not
// the creature (Demonic Ruckus' shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4eb8d877-ee46-4bbe-99da-fee9e238ca58",
		Name:         "Audacity",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Audacity — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
