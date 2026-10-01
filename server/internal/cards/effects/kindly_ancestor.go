package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kindly Ancestor // Ancestor's Embrace (#1855, ADR 0107 §4) — a
// disturb card whose back face is an Aura.
//
// Front face, Creature — Spirit {2}{W}, 2/3:
//
//	"Lifelink
//	 Disturb {1}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Enchanted creature has lifelink.
//	 If Ancestor's Embrace would be put into a graveyard from anywhere,
//	 exile it instead."
//
// Disturbed, the spell is the Aura (CR 712.8c) and targets through the
// back face's EnchantCreature clause (CR 303.4a). The lifelink grant
// is the shared attachment static, which follows the Aura. The exile
// clause is DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         kindlyAncestorOracleID,
		Name:             "Kindly Ancestor",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{1}{W}")},
	})
	Register(Spec{
		OracleID:     kindlyAncestorOracleID + "#1",
		Name:         "Ancestor's Embrace",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{GrantToAttached("lifelink")},
		Replacements: []game.ReplacementEffect{DisturbedExile("Ancestor's Embrace")},
	})
}

const kindlyAncestorOracleID = "ee049bf3-b31c-4dcc-996f-bb076848432b"
