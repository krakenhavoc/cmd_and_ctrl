package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const jackdawSaviorOracle = "9e31ef54-b2f2-4f3f-9779-151b74470ce4"

func TestJackdawSaviorIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(jackdawSaviorOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Jackdaw Savior: %+v %v", spec, ok)
	}
}

func jsGrave(me *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: cost, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	return id
}

func jsFlyer(g *game.Game, owner uuid.UUID, name, cost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird", ManaCost: cost, Keywords: []string{"flying"},
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func TestJackdawSaviorReturnsACreatureWithLesserManaValueThanTheFlyerThatDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dskB1Creature(g, me.ID, "Jackdaw Savior", "Creature — Bird Cleric", jackdawSaviorOracle, 3, 1)
	flyer := jsFlyer(g, me.ID, "Big Flyer", "{3}{W}") // mana value 4
	cheap := jsGrave(me, "Two Drop", "{1}{W}")
	three := jsGrave(me, "Three Drop", "{2}{W}")
	four := jsGrave(me, "Four Drop", "{3}{W}")
	five := jsGrave(me, "Five Drop", "{4}{W}")
	advanceToMain(t, g)
	killCreature(t, g, me.ID, flyer)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{four, five, flyer} {
		for _, c := range pick.PickTargetCards {
			if c == bad {
				t.Errorf("offered a creature card with mana value not lesser than 4")
			}
		}
	}
	if len(pick.PickTargetCards) != 2 {
		t.Fatalf("offered %d cards, want the two with mana value 3 or less", len(pick.PickTargetCards))
	}
	answerPickTarget(t, g, three)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, three); !ok {
		t.Error("the chosen creature did not return")
	}
	if !me.Graveyard.Contains(cheap) {
		t.Error("the other card left the graveyard")
	}
}

func TestJackdawSaviorIgnoresANonflyerAndAnOpponentsFlyer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dskB1Creature(g, me.ID, "Jackdaw Savior", "Creature — Bird Cleric", jackdawSaviorOracle, 3, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := jsFlyer(g, opp.ID, "Their Flyer", "{3}{W}")
	jsGrave(me, "Two Drop", "{1}{W}")
	advanceToMain(t, g)
	killCreature(t, g, me.ID, bear)
	killCreature(t, g, opp.ID, theirs)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a nonflyer or an opponent's flyer triggered the Savior")
	}
}

func TestJackdawSaviorTriggersOnItsOwnDeathWithNothingCheaperDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jackdaw := dskB1Creature(g, me.ID, "Jackdaw Savior", "Creature — Bird Cleric", jackdawSaviorOracle, 3, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == jackdaw {
				g.Battlefield.Cards[i].ManaCost = "{2}{W}"
			}
		}
	})
	four := jsGrave(me, "Four Drop", "{3}{W}")
	advanceToMain(t, g)
	killCreature(t, g, me.ID, jackdaw)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the Savior triggered with no creature card of lesser mana value")
	}
	_ = four
}
