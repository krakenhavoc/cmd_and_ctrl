package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         KindlyAncestorOracleID,
		Name:             "Kindly Ancestor",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Disturb("{1}{W}")},
	})
	effects.Register(effects.Spec{
		OracleID:     KindlyAncestorOracleID + "#1",
		Name:         "Ancestor's Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.GrantToAttached("lifelink")},
		Replacements: []game.ReplacementEffect{effects.DisturbedExile("Ancestor's Embrace")},
	})
}

const KindlyAncestorOracleID = "ee049bf3-b31c-4dcc-996f-bb076848432b"
