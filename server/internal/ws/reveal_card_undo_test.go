package ws

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
)

// TestRevealCardIsOneUndoEntry — #671: reveal_card commits through the
// ordinary Apply path, so showing a hand card is exactly one undo
// entry, and one undo takes the knowledge back. A refused show (to the
// owner) pushes none.
func TestRevealCardIsOneUndoEntry(t *testing.T) {
	room, _ := newBundleRoom(t)
	g := room.Game
	owner, foe := g.Seats[0], g.Seats[1]
	var id uuid.UUID
	g.ReadSnapshot(func() { id = owner.Hand.Cards[0].InstanceID })

	known := func() bool {
		var k bool
		g.ReadSnapshot(func() {
			for _, c := range g.Seats[0].Hand.Cards {
				if c.InstanceID == id {
					k = c.KnownBy[foe.ID]
				}
			}
		})
		return k
	}
	show := func(to uuid.UUID) error {
		p := map[string]string{"card_id": id.String()}
		if to != uuid.Nil {
			p["to"] = to.String()
		}
		raw, _ := json.Marshal(p)
		a, err := actions.Decode(string(actions.TypeRevealCard), "", raw)
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		a.Caller = owner.ID
		_, _, err = room.ApplyAction(ActionTag{Actor: owner.ID, Type: a.Type}, func() error {
			return actions.Dispatch(g, a)
		})
		return err
	}

	if err := show(owner.ID); err == nil {
		t.Fatalf("showing a card to its owner succeeded")
	}
	if err := show(foe.ID); err != nil {
		t.Fatalf("reveal_card: %v", err)
	}
	if !known() {
		t.Fatalf("the shown seat is not a knower after reveal_card")
	}
	if _, _, err := room.Undo(owner.ID); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if known() {
		t.Errorf("one undo did not take the knowledge back")
	}
}
