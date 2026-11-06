package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hardened Escort — Creature — Human Soldier {2}{W}, 2/4 (Duskmourn:
// House of Horror, #4466):
//
//	"Whenever this creature attacks, another target creature you control
//	 gets +1/+0 and gains indestructible until end of turn. (Damage and
//	 effects that say "destroy" don't destroy it.)"
//
// Foot Elite's attack trigger word for word: the shared
// pumpAndGrantFirstTarget body gives the still-legal target (CR 608.2b)
// +1/+0 and indestructible, nothing if it left. The target need not be
// attacking.
//
// No simplification.
func init() {
	const label = "Hardened Escort — +1/+0 and indestructible until end of turn"
	Register(Spec{
		OracleID:     "2c6ef9e8-6b60-436c-8391-ae0baabb3a1b",
		Name:         "Hardened Escort",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(label, pumpAndGrantFirstTarget(1, 0, label, "indestructible")),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
