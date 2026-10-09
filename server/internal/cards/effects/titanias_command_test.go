package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// titanias_command_test.go — #2789, CR 608.2c: a modal spell's later
// bullets wait for an earlier bullet's prompt. Titania's Command's
// counters went out while its land search was still open, so a
// creature land the search found missed them.

// The search bullet and the counters bullet, announced counters
// first. The counters must not land until the search is answered, and
// a creature land the search finds gets them too.
func TestTitaniasCommandCountersWaitForTheLandSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1},
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
	)

	castModal(t, g, "Titania's Command", "Sorcery", b38TitaniasCommandOracle, []int{3, 1}, nil)
	passPriorityAroundTable(t, g)

	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("the search bullet should have opened a search prompt")
	}
	if got := findBattlefieldCardForTest(g, bear).Counters[game.CounterPlusOne]; got != 0 {
		t.Fatalf("the counters ran before the search was answered: Bear has %d +1/+1 counters", got)
	}

	answerSearchNamed(t, g, me.ID, "Dryad Arbor", "Forest")
	g.SettleResolution()

	if got := findBattlefieldCardForTest(g, bear).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("Bear has %d +1/+1 counters after the search, want 2", got)
	}
	var arbor *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Dryad Arbor" {
			arbor = &g.Battlefield.Cards[i]
		}
	}
	if arbor == nil {
		t.Fatal("Dryad Arbor should be on the battlefield")
	}
	if got := arbor.Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("the creature land the search found has %d +1/+1 counters, want 2", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("nothing should be left to answer: %d prompts open", len(g.PendingChoices))
	}
}

// The Bears and the counters still run, in printed order, when the
// search finds nothing.
func TestTitaniasCommandBearsAndCountersAfterAFailedSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	castModal(t, g, "Titania's Command", "Sorcery", b38TitaniasCommandOracle, []int{1, 2}, nil)
	passPriorityAroundTable(t, g)
	if n := titaniasBears(g, me.ID); n != 0 {
		t.Fatalf("the Bears were made before the search was answered: %d", n)
	}
	answerSearchFailToFind(t, g, me.ID)
	g.SettleResolution()
	if n := titaniasBears(g, me.ID); n != 2 {
		t.Errorf("two Bears after the search, got %d", n)
	}
}

func titaniasBears(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.Controller == controller && c.IsToken() {
			n++
		}
	}
	return n
}
