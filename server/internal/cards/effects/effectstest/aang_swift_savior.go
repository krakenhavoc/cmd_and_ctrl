package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aang_swift_savior_test.go — issues #325 and #343.
//
// #325 is the whole of this file: "when I cast my commander Aang
// Swift Savior he entered the battlefield but his enter-the-
// battlefield trigger never triggered to airbend a creature or
// spell." It never triggered because no spec existed for the card,
// and no spec was worth writing while the card imported as a
// costless colourless 0/0 — which is #343's other symptom and what
// the face model fixes.

// aangSwiftSaviorCard is the card as deck.toGameCard now imports it:
// per-face printed data with face 0 materialised. The point of
// building it this way rather than stamping the flat fields is that
// the pre-ADR-0034 import produced ManaCost "", Power 0, Toughness 0
// and TypeLine "Legendary Creature — Human Avatar Ally // Legendary
// Creature — Avatar Spirit Ally", and every assertion below would
// have been meaningless against that.
func AangSwiftSaviorCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   effects.AangSwiftSaviorOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Aang, Swift Savior",
				TypeLine:  "Legendary Creature — Human Avatar Ally",
				ManaCost:  "{1}{W}{U}",
				Colors:    []string{"U", "W"},
				Power:     2,
				Toughness: 3,
			},
			{
				Name:      "Aang and La, Ocean's Fury",
				TypeLine:  "Legendary Creature — Avatar Spirit Ally",
				Colors:    []string{"U", "W"},
				Power:     5,
				Toughness: 5,
			},
		},
	}
	c.SetFace(0)
	return c
}
