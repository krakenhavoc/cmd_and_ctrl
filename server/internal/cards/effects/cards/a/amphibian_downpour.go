package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amphibian Downpour — Enchantment — Aura {2}{U}:
//
//	"Flash
//	 Storm (When you cast this spell, copy it for each spell cast
//	 before it this turn. You may choose new targets for the copies.
//	 Copies become tokens.)
//	 Enchant creature
//	 Enchanted creature loses all abilities and is a blue Frog
//	 creature with base power and toughness 1/1."
//
// Darksteel Mutation's shape with a different body: a layer-4 type
// REPLACEMENT (a creature and a Frog, nothing else), a layer-5 colour
// set, the layer-6 ability removal and a layer-7b base P/T. Flash is
// a printed keyword and storm is the ordinary CR 702.40 trigger; each
// copy is an Aura spell that targets its own creature and becomes a
// token as it resolves (CR 608.3f), so a storm count of two puts the
// Frog effect on two creatures, or twice on one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a4280cea-386b-450d-a7e9-29c7aa58ce5a",
		Name:            "Amphibian Downpour",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Triggered:       []game.TriggeredAbility{effects.Storm()},
		Static: []game.StaticAbility{
			effects.SetAttachedTypes([]string{"Creature"}, []string{"Frog"}),
			effects.SetAttachedColors("U"),
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(1, 1),
		},
	})
}
