package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reprobation — Enchantment — Aura {1}{W} (Modern Horizons):
//
//	"Enchant creature
//	 Enchanted creature loses all abilities and is a Coward creature with base
//	 power and toughness 0/1. (It keeps all supertypes but loses all other
//	 types and creature types.)"
//
// Witness Protection's layered effect without the colour and the name (CR
// 613.6): layer 4, its card types become creature alone and its subtypes
// Coward alone, so an artifact, Equipment or Vehicle subtype goes too while
// its supertypes stay (CR 205.1a); layer 6, it loses all abilities; layer 7b,
// base power and toughness 0/1, so counters and pumps still apply on top. Its
// colours and name are untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "445dfef5-358e-49b1-880c-b30fc45f0a9a",
		Name:         "Reprobation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.SetAttachedTypes([]string{"Creature"}, []string{"Coward"}),
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(0, 1),
		},
	})
}
