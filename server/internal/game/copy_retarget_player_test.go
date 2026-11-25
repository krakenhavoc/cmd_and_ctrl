package game

import (
	"testing"

	"github.com/google/uuid"
)

// copy_retarget_player_test.go — #6159, CR 707.10d: a copy "for each other
// permanent or player it could target" may move a player target onto a
// permanent and a permanent target onto a player (Radiant Performer). The
// copy's ref takes the kind of what it now names.

func TestAnAbilityCopyMovesAPlayerTargetOntoAPermanentAndAPlayer(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := bearNamed(me.ID, "Bear")
	g.WithWriteLock(func() { g.Battlefield.PushTop(bear) })
	item := pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "Fixture ping", func(*Game, *StackItem) error { return nil })
	g.WithWriteLock(func() {
		item.Targets = []TargetRef{{Kind: TargetPlayer, ID: them.ID}}
		if n := g.CopyAbilityRetargetedEachForEffect(item.ID, me.ID, them.ID, []uuid.UUID{bear.InstanceID, me.ID}); n != 2 {
			t.Fatalf("%d copies, want 2", n)
		}
	})
	var kinds []TargetRefKind
	for _, it := range g.StackMeta {
		if it.ID == item.ID || len(it.Targets) != 1 {
			continue
		}
		switch it.Targets[0].ID {
		case bear.InstanceID:
			kinds = append(kinds, it.Targets[0].Kind)
			if it.Targets[0].Kind != TargetCard {
				t.Errorf("the copy at the Bear has kind %s, want card", it.Targets[0].Kind)
			}
		case me.ID:
			kinds = append(kinds, it.Targets[0].Kind)
			if it.Targets[0].Kind != TargetPlayer {
				t.Errorf("the copy at me has kind %s, want player", it.Targets[0].Kind)
			}
		}
	}
	if len(kinds) != 2 {
		t.Fatalf("found copies at %v, want one at the Bear and one at me", kinds)
	}
}
