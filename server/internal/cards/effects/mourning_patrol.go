package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mourning Patrol // Morning Apparition (#1855, ADR 0107 §4) — a
// disturb card.
//
// Front face, Creature — Human Soldier {2}{W}, 2/3:
//
//	"Vigilance
//	 Disturb {3}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit Soldier, 2/1:
//
//	"Flying, vigilance
//	 If Morning Apparition would be put into a graveyard from anywhere,
//	 exile it instead."
//
// Disturb is effects.Disturb (CR 702.146a): the spell and the
// permanent are the back face, with the front face's mana value (CR
// 712.8c, 712.8e). The exile clause is the back face's own replacement,
// DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         mourningPatrolOracleID,
		Name:             "Mourning Patrol",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"vigilance"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{3}{W}")},
	})
	Register(Spec{
		OracleID:        mourningPatrolOracleID + "#1",
		Name:            "Morning Apparition",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Replacements:    []game.ReplacementEffect{DisturbedExile("Morning Apparition")},
	})
}

const mourningPatrolOracleID = "600db821-4210-4996-a3f7-e05a143e50c2"
