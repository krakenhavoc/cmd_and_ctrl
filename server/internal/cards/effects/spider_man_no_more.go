package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Man No More — Enchantment — Aura {1}{U} (Marvel's Spider-Man, #4460):
//
//	"Enchant creature
//	 Enchanted creature is a Citizen with base power and toughness 1/1. It has
//	 defender and loses all other abilities. (It also loses all other creature
//	 types.)"
//
// Darksteel Mutation's layers with a creature-type-only replacement (the card
// keeps its other card types, Fog on the Barrow-Downs' shape): layer 4 sets
// the creature types to Citizen, layer 6 removes every ability but the granted
// defender in one effect, layer 7b sets base 1/1 so counters and anthems still
// apply on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "163f1fce-70c8-44e3-a679-f97d52cd1e9a",
		Name:         "Spider-Man No More",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			SetAttachedCreatureTypes("Citizen"),
			LoseAllAbilities("defender"),
			SetAttachedBasePT(1, 1),
		},
	})
}
