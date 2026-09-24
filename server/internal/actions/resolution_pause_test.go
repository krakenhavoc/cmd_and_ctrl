package actions

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_pause_test.go — #1289 at the wire seam. A resolution
// paused on one of its own prompts holds state-based actions until it
// finishes, and not every answer path ends by running them. Dispatch
// settles the resolution after every action, so the answer that
// finishes a resolution is always followed by its CR 704.3 boundary.
//
// The colour pick is the answer path that proves it:
// ResolveManaChoice drops the colour into the pool and returns without
// running the state checks, which was harmless while the checks had
// already run at the end of the resolution.

func TestTheAnswerThatFinishesAResolutionIsFollowedByStateBasedActions(t *testing.T) {
	g := newGame(t)
	me := g.Seats[0]
	bear := game.NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Power, bear.Toughness = 2, 2
	bear.PrintedPTKnown = true
	bear.Controller = me.ID
	bear.DamageMarked = 2
	g.Battlefield.PushTop(bear)

	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		id := uuid.New()
		g.StackMeta[id] = &game.StackItem{
			ID: id, Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID,
			Label: "add one mana of any colour",
			Effect: func(g *game.Game, it *game.StackItem) error {
				return g.AddManaForEffect(it.Controller, uuid.Nil, "{W|U}")
			},
		}
	})
	for i := 0; i < len(g.Seats) && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceMana {
		t.Fatalf("pending = %v, want the resolution's colour pick", g.PendingChoices)
	}
	if !onBattlefield(g, bear.InstanceID) {
		t.Fatal("the sweep ran while the resolution was waiting on its colour pick")
	}

	params, _ := json.Marshal(map[string]string{"choice_id": g.PendingChoices[0].ID.String(), "color": "W"})
	if err := Dispatch(g, Action{Type: TypeResolveChoice, Player: me.ID, Caller: me.ID, Params: params}); err != nil {
		t.Fatalf("Dispatch resolve_choice: %v", err)
	}

	if g.ResolutionPaused() {
		t.Error("the resolution is still paused after its last prompt was answered")
	}
	if onBattlefield(g, bear.InstanceID) {
		t.Error("the lethally damaged bear survived the answer that finished the resolution")
	}
}

func onBattlefield(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}
