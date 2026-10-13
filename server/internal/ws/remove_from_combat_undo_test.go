package ws

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestRemoveFromCombatIsOneUndoEntry — #672: remove_from_combat commits
// through the ordinary Apply path, so it is exactly one undo entry, and
// one undo puts the creature back in combat. A refused removal (a
// creature not in combat) pushes none.
func TestRemoveFromCombatIsOneUndoEntry(t *testing.T) {
	room, _ := newBundleRoom(t)
	g := room.Game
	seat := g.Seats[0]
	foe := g.Seats[1]

	c := game.NewCard("Test Creature", seat.ID)
	c.TypeLine = "Creature — Test"
	c.Power, c.Toughness = 2, 2
	g.Battlefield.PushTop(c)
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.DeclareAttacker(c.InstanceID, foe.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	attacking := func() bool {
		var on bool
		g.ReadSnapshot(func() {
			for _, bc := range g.Battlefield.Cards {
				if bc.InstanceID == c.InstanceID {
					on = bc.AttackingTarget != uuid.Nil
				}
			}
		})
		return on
	}
	remove := func() error {
		raw, _ := json.Marshal(map[string]string{"instance_id": c.InstanceID.String()})
		a, err := actions.Decode(string(actions.TypeRemoveFromCombat), "", raw)
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		a.Caller = seat.ID
		_, _, err = room.ApplyAction(ActionTag{Actor: seat.ID, Type: a.Type}, func() error {
			return actions.Dispatch(g, a)
		})
		return err
	}

	if err := remove(); err != nil {
		t.Fatalf("remove_from_combat: %v", err)
	}
	if attacking() {
		t.Fatalf("still attacking after remove_from_combat")
	}
	// A second removal is refused and must not mint an undo entry.
	if err := remove(); err == nil {
		t.Fatalf("removing a creature already out of combat succeeded")
	}
	if _, _, err := room.Undo(seat.ID); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if !attacking() {
		t.Errorf("one undo did not put the creature back in combat")
	}
}
