package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dragonsApproachOracle = "47e191f7-6314-4875-b5ee-57e5daf089c4"

func dragonsApproachCopies(p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dragon's Approach", TypeLine: "Sorcery",
			OracleID: dragonsApproachOracle, Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

// The spell deals 3 damage to each opponent, and with too few copies in the
// graveyard it asks nothing.
func TestDragonsApproachDamagesEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		before[p.ID] = lifeOf(g, p.ID)
	}
	mine := lifeOf(g, me.ID)
	dragonsApproachCopies(me, 3)
	spell := castCatalogSpell(t, g, "Dragon's Approach", "Sorcery", dragonsApproachOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if got := lifeOf(g, p.ID); got != before[p.ID]-3 {
			t.Errorf("%s life = %d, want %d", p.Name, got, before[p.ID]-3)
		}
	}
	if lifeOf(g, me.ID) != mine {
		t.Error("the caster lost life")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			t.Fatalf("asked with only three other copies: %+v", c)
		}
	}
	if !inGraveyard(me, spell) {
		t.Error("the spell did not go to the graveyard")
	}
}

// Four other copies: say yes, five cards leave for exile and a Dragon enters.
func TestDragonsApproachExilesFiveAndFetchesADragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	others := dragonsApproachCopies(me, 4)
	dragon := pushLibraryCardForTest(me, game.Card{Name: "Test Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}", Power: 5, Toughness: 5})
	bear := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	spell := castCatalogSpell(t, g, "Dragon's Approach", "Sorcery", dragonsApproachOracle, nil)
	passPriorityAroundTable(t, g)
	answerConfirmFor(t, g, me, true)
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchByID(t, g, me.ID, dragon)
	}
	for _, id := range append(others, spell) {
		if !inExile(g, id) {
			t.Errorf("card %s not exiled", id)
		}
	}
	if !onBattlefield(g, dragon) {
		t.Error("the Dragon did not enter")
	}
	if onBattlefield(g, bear) {
		t.Error("a non-Dragon entered")
	}
}

// Declining leaves the graveyard alone and searches for nothing.
func TestDragonsApproachDeclineKeepsTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	others := dragonsApproachCopies(me, 4)
	dragon := pushLibraryCardForTest(me, game.Card{Name: "Test Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}", Power: 5, Toughness: 5})
	spell := castCatalogSpell(t, g, "Dragon's Approach", "Sorcery", dragonsApproachOracle, nil)
	passPriorityAroundTable(t, g)
	answerConfirmFor(t, g, me, false)
	for _, id := range append(others, spell) {
		if !inGraveyard(me, id) {
			t.Errorf("card %s left the graveyard", id)
		}
	}
	if onBattlefield(g, dragon) {
		t.Error("a Dragon entered after a decline")
	}
}
