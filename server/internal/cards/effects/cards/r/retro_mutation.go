package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retro-Mutation — {2}{U} Enchantment — Aura (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature is a Turtle with base power and toughness 0/1. It
//	 can't attack and loses all abilities. (It also loses all other creature
//	 types.)"
//
// Spider-Man No More's layers: the creature types alone are replaced in
// layer 4 (CR 613.1d), so it keeps its card types; the abilities go in layer
// 6; base 0/1 is set in layer 7b (CR 613.4b), so counters and anthems still
// apply. "Can't attack" is Pacifism's restriction without the block half.
// All of it is read live from the Aura and ends when the Aura leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8b5de6d-8654-4550-bccb-79de6cbef4d5",
		Name:            "Retro-Mutation",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.SetAttachedCreatureTypes("Turtle"),
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(0, 1),
			effects.RestrictAttached(game.CantAttack),
		},
	})
}
