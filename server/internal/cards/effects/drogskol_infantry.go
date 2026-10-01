package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         drogskolInfantryOracleID,
		Name:             "Drogskol Infantry",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{3}{W}")},
	})
	Register(Spec{
		OracleID:     drogskolInfantryOracleID + "#1",
		Name:         "Drogskol Armaments",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Replacements: []game.ReplacementEffect{DisturbedExile("Drogskol Armaments")},
	})
}

const drogskolInfantryOracleID = "389bcb9f-4e66-4704-9968-a1c1574ec2c8"
