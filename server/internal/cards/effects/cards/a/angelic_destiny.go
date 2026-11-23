package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8ba36878-0816-44c7-b543-8ebfc28c2ecd",
		Name:         "Angelic Destiny",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(4, 4),
			effects.GrantToAttached("flying", "first strike"),
			effects.AttachedIsAlso("Angel"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Angelic Destiny — return this card to its owner's hand", effects.EndlessEvilReturn),
		},
	})
}
