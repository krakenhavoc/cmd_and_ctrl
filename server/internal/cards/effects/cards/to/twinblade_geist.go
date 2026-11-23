package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twinblade Geist // Twinblade Invocation (#1855, ADR 0107 §4) — a
// disturb card whose back face is an Aura.
//
// Front face, Creature — Spirit Warrior {1}{W}, 1/1:
//
//	"Double strike
//	 Disturb {2}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature has double strike.
//	 If Twinblade Invocation would be put into a graveyard from
//	 anywhere, exile it instead."
//
// Disturbed, the spell is the Aura (CR 712.8c) and targets through the
// back face's EnchantCreature clause (CR 303.4a). The double strike
// grant is the shared attachment static. The exile clause is
// DisturbedExile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         TwinbladeGeistOracleID,
		Name:             "Twinblade Geist",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"double strike"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Disturb("{2}{W}")},
	})
	effects.Register(effects.Spec{
		OracleID:     TwinbladeGeistOracleID + "#1",
		Name:         "Twinblade Invocation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.GrantToAttached("double strike")},
		Replacements: []game.ReplacementEffect{effects.DisturbedExile("Twinblade Invocation")},
	})
}

const TwinbladeGeistOracleID = "4db96d32-b4c2-44e9-a73f-aca7dad279b6"
