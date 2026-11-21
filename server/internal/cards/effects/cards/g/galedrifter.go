package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galedrifter // Waildrifter (#1855, ADR 0107 §4) — a disturb card.
//
// Front face, Creature — Hippogriff {3}{U}, 3/2:
//
//	"Flying
//	 Disturb {4}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Hippogriff Spirit, 2/2:
//
//	"Flying
//	 If Waildrifter would be put into a graveyard from anywhere, exile
//	 it instead."
//
// Disturb is effects.Disturb (CR 702.146a): the spell and the
// permanent are the back face, with the front face's mana value (CR
// 712.8c, 712.8e). The exile clause is the back face's own replacement,
// DisturbedExile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         GaledrifterOracleID,
		Name:             "Galedrifter",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Disturb("{4}{U}")},
	})
	effects.Register(effects.Spec{
		OracleID:        GaledrifterOracleID + "#1",
		Name:            "Waildrifter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.DisturbedExile("Waildrifter")},
	})
}

const GaledrifterOracleID = "ba88575a-4b9a-40cd-abbc-4539912c9455"
