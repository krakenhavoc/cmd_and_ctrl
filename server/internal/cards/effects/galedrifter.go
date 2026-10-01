package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         galedrifterOracleID,
		Name:             "Galedrifter",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{4}{U}")},
	})
	Register(Spec{
		OracleID:        galedrifterOracleID + "#1",
		Name:            "Waildrifter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{DisturbedExile("Waildrifter")},
	})
}

const galedrifterOracleID = "ba88575a-4b9a-40cd-abbc-4539912c9455"
