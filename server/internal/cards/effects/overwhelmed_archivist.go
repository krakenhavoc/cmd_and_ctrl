package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overwhelmed Archivist // Archive Haunt (#1855, ADR 0107 §4) — a
// disturb card.
//
// Front face, Creature — Human Wizard {2}{U}, 3/2:
//
//	"When this creature enters, draw a card, then discard a card.
//	 Disturb {3}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit Wizard, 2/1:
//
//	"Flying
//	 Whenever this creature attacks, draw a card, then discard a card.
//	 If Archive Haunt would be put into a graveyard from anywhere, exile
//	 it instead."
//
// Both loots are the shared draw-then-discard body (lootOne): the draw
// first, then a discard prompt built from the hand after it. The front
// face's ETB does not fire for a disturbed Archive Haunt — the
// permanent enters back face up and has only the back face's abilities
// (CR 702.146b, 712.8e). The exile clause is DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         overwhelmedArchivistOracleID,
		Name:             "Overwhelmed Archivist",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{3}{U}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Overwhelmed Archivist — draw a card, then discard a card", b35LootOne),
		},
	})
	Register(Spec{
		OracleID:        overwhelmedArchivistOracleID + "#1",
		Name:            "Archive Haunt",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Archive Haunt — draw a card, then discard a card", b35LootOne),
		},
		Replacements: []game.ReplacementEffect{DisturbedExile("Archive Haunt")},
	})
}

const overwhelmedArchivistOracleID = "13c90d78-cfb1-4d40-a35e-1fd170450b45"
