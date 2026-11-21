package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		before[p.ID] = effectstest.LifeOf(g, p.ID)
	}
	mine := effectstest.LifeOf(g, me.ID)
	dragonsApproachCopies(me, 3)
	spell := effectstest.CastCatalogSpell(t, g, "Dragon's Approach", "Sorcery", dragonsApproachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if got := effectstest.LifeOf(g, p.ID); got != before[p.ID]-3 {
			t.Errorf("%s life = %d, want %d", p.Name, got, before[p.ID]-3)
		}
	}
	if effectstest.LifeOf(g, me.ID) != mine {
		t.Error("the caster lost life")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			t.Fatalf("asked with only three other copies: %+v", c)
		}
	}
	if !effectstest.InGraveyard(me, spell) {
		t.Error("the spell did not go to the graveyard")
	}
}

// Four other copies: say yes, five cards leave for exile and a Dragon enters.
func TestDragonsApproachExilesFiveAndFetchesADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	others := dragonsApproachCopies(me, 4)
	dragon := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Test Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}", Power: 5, Toughness: 5})
	bear := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	spell := effectstest.CastCatalogSpell(t, g, "Dragon's Approach", "Sorcery", dragonsApproachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, dragon)
	}
	for _, id := range append(others, spell) {
		if !effectstest.InExile(g, id) {
			t.Errorf("card %s not exiled", id)
		}
	}
	if !effects.OnBattlefield(g, dragon) {
		t.Error("the Dragon did not enter")
	}
	if effects.OnBattlefield(g, bear) {
		t.Error("a non-Dragon entered")
	}
}

// Declining leaves the graveyard alone and searches for nothing.
func TestDragonsApproachDeclineKeepsTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	others := dragonsApproachCopies(me, 4)
	dragon := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Test Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}", Power: 5, Toughness: 5})
	spell := effectstest.CastCatalogSpell(t, g, "Dragon's Approach", "Sorcery", dragonsApproachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, false)
	for _, id := range append(others, spell) {
		if !effectstest.InGraveyard(me, id) {
			t.Errorf("card %s left the graveyard", id)
		}
	}
	if effects.OnBattlefield(g, dragon) {
		t.Error("a Dragon entered after a decline")
	}
}
