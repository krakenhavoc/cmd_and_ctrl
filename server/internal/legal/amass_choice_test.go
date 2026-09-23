package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// amass_choice_test.go — the enumerator half of #1236.
//
// Amass queues a prompt the engine did not have a caller for before:
// "choose an Army creature you control" (CR 701.47a), as a
// PendingChoiceChooseCards over the BATTLEFIELD with a floor of one.
// The kind is already enumerated (choices.go), so the claim here is
// not that new code answers it — it is that the floor-of-one shape a
// bot cannot decline is reachable, sound and bounded:
//
//   - every Army is offered, and nothing else on the board is;
//   - every offered move is one the dispatcher accepts;
//   - NO move is marked always-legal, because a floor above zero has
//     no unconditional answer — the enumerator's own rule, and the
//     one that decides whether a seat owing this prompt can move at
//     all.
//
// A seat that cannot answer a mandatory prompt is a wedged table
// (#544), and amass is the busiest new source of one in the catalog:
// Orcish Bowmasters asks on every opponent draw.

// amassArmy parks an Army creature under p, with one +1/+1 counter so
// it survives the state-based check a dispatched answer runs.
func amassArmy(g *game.Game, p *game.Player, name string) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name:           name,
		TypeLine:       "Token Creature — Zombie Army",
		Power:          0,
		Toughness:      0,
		PrintedPTKnown: true,
		Counters:       map[string]int{game.CounterPlusOne: 1},
	})
}

// queueAmass takes the keyword action, which queues the prompt when
// the seat controls more than one Army.
func queueAmass(t *testing.T, g *game.Game, seat uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		err := g.AmassForEffect(seat, uuid.Nil, game.Card{
			Name:           "Zombie Army",
			TypeLine:       "Token Creature — Zombie Army",
			PrintedPTKnown: true,
			Colors:         []string{"B"},
		}, "Zombie", 2, nil)
		if err != nil {
			t.Fatalf("AmassForEffect: %v", err)
		}
	})
}

func TestAmassOffersEveryArmyAndNothingElse(t *testing.T) {
	g := newTable(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := amassArmy(g, me, "My Army")
	alsoMine := amassArmy(g, me, "My Other Army")
	theirs := amassArmy(g, them, "Their Army")
	bear := battlefieldCard(g, me, creature("Grizzly Bears", "{1}{G}", 2, 2))

	queueAmass(t, g, me.ID)
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts queued, want the amass's one", len(g.PendingChoices))
	}

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)

	offered := map[uuid.UUID]bool{}
	for _, m := range moves {
		if m.Kind != legal.KindChoice {
			continue
		}
		if m.AlwaysLegal {
			t.Errorf("move %q is marked always-legal; a floor of one has no unconditional answer", m.Label)
		}
		ids := pickedCardIDs(t, m)
		if len(ids) != 1 {
			t.Errorf("move %q picks %d cards, want exactly one Army", m.Label, len(ids))
			continue
		}
		offered[ids[0]] = true
	}

	if !offered[mine] || !offered[alsoMine] {
		t.Errorf("offered %v, want both of the seat's Armies (%v, %v)", offered, mine, alsoMine)
	}
	if offered[theirs] {
		t.Error("an opponent's Army was offered")
	}
	if offered[bear] {
		t.Error("a non-Army creature was offered")
	}
	if len(offered) != 2 {
		t.Errorf("%d distinct answers, want 2", len(offered))
	}
}

// TestAmassLeavesNoChoiceWhenThereIsNothingToChoose is the other half
// of "the seat can always move": with one Army the verb asks nothing,
// so the prompt queue stays empty and the seat's move list is its
// ordinary one rather than a forced click.
func TestAmassLeavesNoChoiceWhenThereIsNothingToChoose(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	amassArmy(g, me, "My Army")

	queueAmass(t, g, me.ID)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts queued for a single-Army amass, want none", len(g.PendingChoices))
	}
	moves := legal.EnumerateFor(g, me.ID)
	for _, m := range moves {
		if m.Type == legal.TypeResolveChoice {
			t.Errorf("the seat is offered %q with no choice pending", m.Label)
		}
	}
}
