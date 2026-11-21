package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const neoOrochiMergeKeeperGrant = "orochi-merge-keeper/tap-for-two-green"

// Orochi Merge-Keeper — Creature — Snake Druid {1}{G}, 1/1 (Kamigawa: Neon
// Dynasty):
//
//	"{T}: Add {G}.
//	 As long as this creature is modified, it has "{T}: Add {G}{G}."
//	 (Equipment, Auras you control, and counters are modifications.)"
//
// The second ability is a granted bundle on itself while modified
// (Tazri's shape), so it really is a second mana ability and goes when the
// last modification does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53e39011-3d77-4883-9c8e-4707f6fc5c47",
		Name:         "Orochi Merge-Keeper",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Grants: []effects.AbilityGrant{effects.TapForManaGrant(neoOrochiMergeKeeperGrant, "{G}{G}", "Add {G}{G}", "{T}: Add {G}{G}.")},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.B07IsModified(g, *target)
			}, neoOrochiMergeKeeperGrant),
				"As long as this creature is modified, it has \"{T}: Add {G}{G}.\""),
		},
	})
}
