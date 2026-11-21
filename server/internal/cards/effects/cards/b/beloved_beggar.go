package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beloved Beggar // Generous Soul (#1855, ADR 0107 §4) — a disturb
// card.
//
// Front face, Creature — Human Peasant {1}{W}, 0/4:
//
//	"Disturb {4}{W}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit, 4/4:
//
//	"Flying, vigilance
//	 If Generous Soul would be put into a graveyard from anywhere,
//	 exile it instead."
//
// Disturb is effects.Disturb (CR 702.146a): the spell and the
// permanent are the back face, with the front face's mana value (CR
// 712.8c, 712.8e). The exile clause is the back face's own replacement,
// DisturbedExile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         BelovedBeggarOracleID,
		Name:             "Beloved Beggar",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Disturb("{4}{W}{W}")},
	})
	effects.Register(effects.Spec{
		OracleID:        BelovedBeggarOracleID + "#1",
		Name:            "Generous Soul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Replacements:    []game.ReplacementEffect{effects.DisturbedExile("Generous Soul")},
	})
}

const BelovedBeggarOracleID = "74c9cc13-c03f-4322-82af-b7bce1f2a0d8"
