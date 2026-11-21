package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noggle the Mind — Enchantment — Aura {1}{U} (Lorwyn Eclipsed, #4458):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature loses all abilities and is a colorless Noggle with
//	 base power and toughness 1/1. (It loses all colors and all other
//	 creature types.)"
//
// Darksteel Mutation's four layers in printed order: its creature types
// become Noggle alone and it keeps its card types (layer 4, the Fog on the
// Barrow-Downs' SetAttachedCreatureTypes), it is colorless (layer 5), it
// loses all abilities (layer 6) and its base power and toughness are 1/1
// (layer 7b, so counters and anthems still apply, and a later set overwrites
// it). "Noggle" is a word of the Aura's text. Flash rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b49efbe8-e15b-4dc7-82de-970cbce92b79",
		Name:            "Noggle the Mind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.SetAttachedCreatureTypes("Noggle"),
			effects.SetAttachedColors(),
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(1, 1),
		},
	})
}
