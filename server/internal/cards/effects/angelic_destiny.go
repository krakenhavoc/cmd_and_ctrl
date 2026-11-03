package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Destiny — Enchantment — Aura {2}{W}{W}:
//
//	"Enchant creature
//	 Enchanted creature gets +4/+4, has flying and first strike, and is an
//	 Angel in addition to its other types.
//	 When enchanted creature dies, return this card to its owner's hand."
//
// The static is three layered pieces: the +4/+4 (layer 7c), the keywords
// (layer 6) and the Angel creature type added in layer 4 (it keeps the
// creature's other types). The dies trigger is harvested while the Aura is
// still attached and returns this card only if it is still the Aura the
// state-based action put into the graveyard (CR 400.7), Endless Evil's
// body, so one exiled in response stays exiled. A creature that dies while
// it is not enchanted by this Aura triggers nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ba36878-0816-44c7-b543-8ebfc28c2ecd",
		Name:         "Angelic Destiny",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(4, 4),
			GrantToAttached("flying", "first strike"),
			AttachedIsAlso("Angel"),
		},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Angelic Destiny — return this card to its owner's hand", endlessEvilReturn),
		},
	})
}
