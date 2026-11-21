package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drogskol Infantry // Drogskol Armaments (#1855, ADR 0107 §4) — a
// disturb card whose back face is an Aura.
//
// Front face, Creature — Spirit Soldier {1}{W}, 2/2:
//
//	"Disturb {3}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 If Drogskol Armaments would be put into a graveyard from anywhere,
//	 exile it instead."
//
// Disturbed, the spell is the Aura (CR 712.8c), so it targets a
// creature as it is cast (CR 303.4a) through the back face's own
// EnchantCreature clause. An Aura that falls off, and an Aura spell
// whose target is gone, would go to the graveyard and are exiled
// instead (DisturbedExile).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         DrogskolInfantryOracleID,
		Name:             "Drogskol Infantry",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Disturb("{3}{W}")},
	})
	effects.Register(effects.Spec{
		OracleID:     DrogskolInfantryOracleID + "#1",
		Name:         "Drogskol Armaments",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Replacements: []game.ReplacementEffect{effects.DisturbedExile("Drogskol Armaments")},
	})
}

const DrogskolInfantryOracleID = "389bcb9f-4e66-4704-9968-a1c1574ec2c8"
