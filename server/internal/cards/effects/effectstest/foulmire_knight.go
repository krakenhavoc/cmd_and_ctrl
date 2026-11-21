package effectstest

import (
	"github.com/google/uuid"

	cardsfl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/fl"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// foulmire_knight_test.go — the CARD half of CR 715 (#719). The
// engine half, against a fixture with no catalog entry, is
// server/internal/game/adventure_test.go; if this fails and that
// passes, the bug is in this file's Spec rather than in the rules.
//
// What a real adventure card adds over the fixture is the composite
// catalog key: the Adventure half's Spec is registered under
// "<oracle_id>#1", and it is only reached because CastSpell
// materialises face 1 before CatalogKey is read (ADR 0034 §5). A
// wiring mistake there shows up here as a card that draws nothing.

// foulmireKnight builds the printed card the way deck.toGameCard
// does: per-face data, then SetFace(0).
func FoulmireKnight(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   cardsfl.FoulmireKnightOracleID,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Foulmire Knight",
				TypeLine:  "Creature — Zombie Knight",
				ManaCost:  "{B}",
				Colors:    []string{"B"},
				Power:     1,
				Toughness: 1,
			},
			{
				Name:     "Profane Insight",
				TypeLine: "Instant — Adventure",
				ManaCost: "{2}{B}",
				Colors:   []string{"B"},
			},
		},
	}
	c.SetFace(0)
	return c
}
