package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s30_indestructible_wipes_test.go — #470 / #446 at the catalog
// level.
//
// The engine-side cases live in game/indestructible_test.go. These
// are the ones a player would file the bug from: a real Wrath of God,
// cast off a real Avacyn board, resolved through the real priority
// loop. They exist because the gap that shipped was not a missing
// rule — IsIndestructible had been correct for four sprints — it was
// one of two entry points not asking the question, and only an
// end-to-end cast goes through the one that was wrong.
//
// The second and third cases are the half of the fix that is easy to
// forget. Filtering the destruction is not enough on its own: "for
// each creature destroyed this way" reads a count, and Deadly
// Tempest's per-controller tally reads the swept CARDS, so a survivor
// left in either would pay out for a creature still standing.

// pushIndestructibleWipeCreature seeds a creature whose indestructible
// is PRINTED on it, via game.Card.Keywords — the same road
// Darksteel Citadel takes. No oracle ID needed, and it survives a
// layer recompute because printedCharacteristic merges the slice in.
func PushIndestructibleWipeCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Golem",
		Power:      2,
		Toughness:  2,
		Keywords:   []string{"indestructible"},
		Owner:      owner,
		Controller: owner,
	})
	return id
}
