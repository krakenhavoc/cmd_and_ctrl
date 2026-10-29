package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fog on the Barrow-Downs — Enchantment — Aura {2}{W}:
//
//	"Enchant creature
//	 Enchanted creature is a Spirit and can't attack or block. (It
//	 loses all other creature types.)"
//
// Pacifism's restriction and a layer 4 replacement of the creature types
// alone (SetAttachedCreatureTypes, CR 205.1a): the host keeps its card
// types, its other abilities and everything else it is, and stops being
// whatever it was for a lord, a tribal payoff or a Rite of Replication
// count. The restriction is the Aura's own effect, not an ability the
// creature has, so it holds whatever the creature is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66bee79a-7e47-4fc4-ba07-323cd2de6261",
		Name:         "Fog on the Barrow-Downs",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			SetAttachedCreatureTypes("Spirit"),
			RestrictAttached(game.CantAttackOrBlock),
		},
	})
}
