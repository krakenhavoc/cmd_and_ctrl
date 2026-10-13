package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aura_host_view_test.go — #3158's wire half. CR 303.4f's aura_host
// prompt rides the pick_target projection (`pick_target.cards` /
// `.players`, min 1 max 1), so the client's board-click flow — the zone
// browser for a card in a graveyard — answers it with no new picker.
func TestAuraHostPromptRidesThePickTargetProjection(t *testing.T) {
	g := newTwoSeatGame(t)
	me, them := g.Seats[0], g.Seats[1]
	host := uuid.New()
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:              game.PendingChoiceAuraHost,
			Chooser:           me.ID,
			Count:             1,
			Reason:            "Animate Dead — choose what it enchants as it enters",
			PickTargetCards:   []uuid.UUID{host},
			PickTargetPlayers: []uuid.UUID{them.ID},
			PickTargetMin:     1,
			PickTargetMax:     1,
		})
	})
	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	c := choiceFor(t, v, string(game.PendingChoiceAuraHost))
	if c.PickTarget == nil {
		t.Fatal("aura_host carries no pick_target projection")
	}
	if len(c.PickTarget.Cards) != 1 || c.PickTarget.Cards[0] != host.String() ||
		len(c.PickTarget.Players) != 1 || c.PickTarget.Players[0] != them.ID.String() {
		t.Errorf("candidates on the wire: %+v", c.PickTarget)
	}
	if c.PickTarget.Min != 1 || c.PickTarget.Max != 1 {
		t.Errorf("bounds %d..%d, want 1..1", c.PickTarget.Min, c.PickTarget.Max)
	}
}
