package effectstest

import (
	"github.com/google/uuid"

	cardsbl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bl"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bonecrusher_giant_test.go — the first TARGETED adventure card
// (#992). Foulmire Knight proved the CR 715 lifecycle with a half that
// announces nothing; this one proves the half that announces a target,
// which is what the per-face announce data on the wire exists for.
//
// The wire half is server/internal/protocol/per_face_announce_view_test.go
// and the client half is client/src/lib/adventureCast.test.ts. If those
// pass and this fails, the bug is in this file's Spec rather than in
// the seam.

// bonecrusherGiant builds the printed card the way deck.toGameCard
// does: per-face data, then SetFace(0).
func BonecrusherGiant(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   cardsbl.BonecrusherGiantOracleID,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Bonecrusher Giant",
				TypeLine:  "Creature — Giant",
				ManaCost:  "{2}{R}",
				Colors:    []string{"R"},
				Power:     4,
				Toughness: 3,
			},
			{
				Name:     "Stomp",
				TypeLine: "Instant — Adventure",
				ManaCost: "{1}{R}",
				Colors:   []string{"R"},
			},
		},
	}
	c.SetFace(0)
	return c
}
