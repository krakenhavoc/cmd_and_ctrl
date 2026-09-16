package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// pending_kinds_test.go — the last two PendingChoice kinds #499 found
// enumerated NOTHING: may_cast and choose_creature_type.
//
// A seat owing a choice is offered only that choice's answers, and the
// engine refuses pass_priority while it is open. So a kind with no
// answers is not a small gap: it is an empty move list, a bot asleep,
// and every other seat stuck behind it. The catalog soak (#601) wedged
// real tables on Door of Destinies and Cavern of Souls exactly that way.
//
// Each test ends in dispatchAll, which replays every offered move
// against a clone: the enumerator's promise is that the engine accepts
// all of them.

func allAlwaysLegal(moves []legal.Move) bool {
	for _, m := range moves {
		if !m.AlwaysLegal {
			return false
		}
	}
	return true
}

func typedCreature(name, types string) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — " + types, Power: 1, Toughness: 1}
}

func queueCreatureTypeChoice(g *game.Game, chooser uuid.UUID) {
	g.WithWriteLock(func() {
		g.QueueCreatureTypeChoiceForEffect(chooser, uuid.New(), "Door of Destinies — choose a creature type")
	})
}

func TestMayCastOffersBothAnswersAndDeclineIsTheWayOut(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	hit := handCard(me, creature("Cascade Hit", "{3}{G}", 3, 3))
	var accepted, declined bool
	g.WithWriteLock(func() {
		if err := g.QueueMayCastForEffect(me.ID, uuid.New(), hit, "Bloodbraid Elf — cast Cascade Hit for free?",
			func(*game.Game) error { accepted = true; return nil },
			func(*game.Game) error { declined = true; return nil },
		); err != nil {
			t.Fatalf("QueueMayCastForEffect: %v", err)
		}
	})

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 2 {
		t.Fatalf("enumerated %d answers, want 2: %v", len(moves), labels(moves))
	}
	var sawDecline bool
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("a seat owing may_cast was offered %q as well", m.Label)
		}
		if strings.HasSuffix(m.Label, ": don't cast it") {
			sawDecline = true
			if !m.AlwaysLegal {
				t.Error("the decline commits the seat to nothing and must be its always-legal way out")
			}
		} else if m.AlwaysLegal {
			t.Errorf("%q is marked always legal, but the accept branch is the one that can go wrong", m.Label)
		}
	}
	if !sawDecline {
		t.Errorf("no decline answer: %v", labels(moves))
	}
	if accepted || declined {
		t.Fatal("setup: a branch ran before any answer was dispatched")
	}
	dispatchAll(t, g, me.ID, moves)

	// dispatchAll replays each answer on a clone, and a clone shares
	// the prompt's branch closures. So both flags flipping proves each
	// answer reached its own branch, not merely that it was accepted.
	if !accepted {
		t.Error("the accept answer never reached the accept branch")
	}
	if !declined {
		t.Error("the decline answer never reached the decline branch")
	}
}

func TestCreatureTypeOffersTheBoardsTribesMostCommonFirst(t *testing.T) {
	g := newTable(t)
	me, them := g.Seats[0], g.Seats[1]
	battlefieldCard(g, me, typedCreature("Elvish Mystic", "Elf Druid"))
	battlefieldCard(g, me, typedCreature("Elvish Warrior", "Elf Warrior"))
	battlefieldCard(g, me, typedCreature("Goblin Guide", "Goblin Scout"))
	// Someone else's tribe is not this seat's answer.
	battlefieldCard(g, them, typedCreature("Gravecrawler", "Zombie"))
	queueCreatureTypeChoice(g, me.ID)

	moves := legal.EnumerateFor(g, me.ID)
	got := make([]string, 0, len(moves))
	for _, m := range moves {
		got = append(got, m.Label[strings.LastIndex(m.Label, ": ")+2:])
	}
	want := []string{"Elf", "Druid", "Goblin", "Scout", "Warrior"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("answers = %v, want %v (most common first, then alphabetical, capped at 5)", got, want)
	}
	if !allAlwaysLegal(moves) {
		t.Error("every creature type is accepted by the resolver, so every answer is always legal")
	}
	dispatchAll(t, g, me.ID, moves)
}

func TestCreatureTypeOnAnEmptyBoardStillAnswers(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	queueCreatureTypeChoice(g, me.ID)

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 1 {
		t.Fatalf("enumerated %d answers on a creatureless board, want exactly the fallback: %v", len(moves), labels(moves))
	}
	if !strings.HasSuffix(moves[0].Label, ": Human") {
		t.Errorf("fallback answer = %q, want Human", moves[0].Label)
	}
	dispatchAll(t, g, me.ID, moves)
}

func TestCreatureTypeOffersAreCapped(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	for _, tp := range []string{"Angel", "Bird", "Cat", "Dragon", "Elf", "Faerie", "Goblin"} {
		battlefieldCard(g, me, typedCreature("A "+tp, tp))
	}
	queueCreatureTypeChoice(g, me.ID)

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 5 {
		t.Fatalf("enumerated %d answers, want the cap of 5: %v", len(moves), labels(moves))
	}
	dispatchAll(t, g, me.ID, moves)
}
