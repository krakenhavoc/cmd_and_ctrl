package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_replacements_test.go — #1222 from the card side: the CR 614
// window on the mana PRODUCED, watched through the ordinary
// Spec.Replacements slot.
//
// Two printed cards, deliberately, so the kind is not shaped around
// one of them: Mana Reflection's ×2 and Nyxbloom Ancient's ×3 are the
// same sentence with different arithmetic, and the board with both on
// it is ×6.

const (
	ManaReflectionOracle  = "a3a8a044-283d-443e-bc40-c2f826d70c22"
	NyxbloomAncientOracle = "8b610f8f-c8dd-4eeb-bc6e-3bc706d5f63e"
)

// pushLandFor seeds a basic land under `owner`; the synthetic
// CR 305.6 mana ability comes off the type line.
func PushLandFor(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
	return id
}
