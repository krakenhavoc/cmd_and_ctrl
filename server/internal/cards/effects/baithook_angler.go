package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baithook Angler // Hook-Haunt Drifter (#1855, ADR 0107 §4) — a
// disturb card.
//
// Front face, Creature — Human Peasant {1}{U}, 2/1:
//
//	"Disturb {1}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit, 1/2:
//
//	"Flying
//	 If Hook-Haunt Drifter would be put into a graveyard from anywhere,
//	 exile it instead."
//
// Disturb is an alternative cost bound to the graveyard that casts the
// back face (effects.Disturb, CR 702.146a); the spell and the permanent
// are the back face (CR 712.11a, 702.146b), with the front face's mana
// value (CR 712.8c, 712.8e). The back face's exile clause is its own
// replacement (DisturbedExile), read only while that face is up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         baithookAnglerOracleID,
		Name:             "Baithook Angler",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{1}{U}")},
	})
	Register(Spec{
		OracleID:        baithookAnglerOracleID + "#1",
		Name:            "Hook-Haunt Drifter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{DisturbedExile("Hook-Haunt Drifter")},
	})
}

const baithookAnglerOracleID = "c6bb4b41-8dae-429a-b928-ae9d39c74711"
