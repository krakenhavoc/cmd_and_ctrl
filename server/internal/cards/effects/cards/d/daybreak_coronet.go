package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daybreak Coronet — Enchantment — Aura {W}{W} (Wilds of Eldraine
// Commander, #5368):
//
//	"Enchant creature with another Aura attached to it
//	 Enchanted creature gets +3/+3 and has first strike, vigilance, and
//	 lifelink."
//
// The enchant clause is a predicate on the host (cmdr2B19HasAnotherAuraAttached),
// so it is the one the CR 704.5m check reruns: when the other Aura leaves,
// the Coronet is put into its owner's graveyard at the next state-based
// check. The Aura may be anyone's, a Role included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.Cmdr2B19DaybreakCoronetOracle,
		Name:         "Daybreak Coronet",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("enchant creature with another Aura attached to it", effects.Cmdr2B19HasAnotherAuraAttached()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("first strike", "vigilance", "lifelink"),
		},
	})
}
