package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chaplain of Alms // Chapel Shieldgeist (#1855, ADR 0107 §4) — a
// disturb card.
//
// Front face, Creature — Human Cleric {W}, 1/1:
//
//	"First strike
//	 Ward {1}
//	 Disturb {3}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit Cleric, 2/1:
//
//	"Flying, first strike
//	 Each creature you control has ward {1}.
//	 If Chapel Shieldgeist would be put into a graveyard from anywhere,
//	 exile it instead."
//
// Ward is a triggered ability (CR 702.21a), so the front face's is the
// shared Ward trigger and the back face's grant is WardGranted over
// every creature its controller controls — the Shieldgeist itself
// included, since "each creature you control" names it too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         chaplainOfAlmsOracleID,
		Name:             "Chaplain of Alms",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"first strike"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{3}{W}")},
		Triggered:        []game.TriggeredAbility{Ward(WardMana("{1}"), "Chaplain of Alms — ward {1}")},
	})
	Register(Spec{
		OracleID:        chaplainOfAlmsOracleID + "#1",
		Name:            "Chapel Shieldgeist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{1}"), "Chapel Shieldgeist — each creature you control has ward {1}",
				TribeFilter{YoursOnly: true}.Matches),
		},
		Replacements: []game.ReplacementEffect{DisturbedExile("Chapel Shieldgeist")},
	})
}

const chaplainOfAlmsOracleID = "c21d1ca3-3d19-4b4d-bbfa-07b5b7bcea4b"
