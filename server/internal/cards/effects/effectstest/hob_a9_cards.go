package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hobA9NextChapter walks to the active seat's next precombat main, where the
// Saga's lore counter arrives, and settles the chapter that reaches.
func HobA9NextChapter(t *testing.T, g *game.Game) {
	t.Helper()
	AdvanceToPrecombatMainOf(t, g, 0)
	PassPriorityAroundTable(t, g)
}

// hobA9SagaRef is the object reference of a battlefield Saga.
func HobA9SagaRef(t *testing.T, g *game.Game, id uuid.UUID) game.ObjectRef {
	t.Helper()
	var ref game.ObjectRef
	var ok bool
	g.ReadSnapshot(func() { ref, ok = g.PermanentRefForEffect(id) })
	if !ok {
		t.Fatal("the Saga is not on the battlefield")
	}
	return ref
}
