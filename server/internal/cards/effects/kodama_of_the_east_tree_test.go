package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kodama_of_the_east_tree_test.go — #3933, CR 603.4: "Whenever another
// permanent you control enters, if it wasn't put onto the battlefield
// with this ability, you may put a permanent card with equal or lesser
// mana value from your hand onto the battlefield."

const kodamaOracle = "cb1afab6-b9c8-4b9c-840b-24106523a783"

// kodamaTable is a main phase on seat 0's turn with Kodama out and an
// empty hand.
func kodamaTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	k := pushCatalogPermanent(g, me.ID, "Kodama of the East Tree", "Legendary Creature — Spirit", kodamaOracle, false)
	return g, me, k
}

// kodamaPrompt passes priority until a put prompt is open for `me`, or
// the stack empties. Nil when there is none.
func kodamaPrompt(t *testing.T, g *game.Game, me uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		if c := latestPutFromHandChoice(g, me); c != nil {
			return c
		}
		var order *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == me {
				order = c
			}
		}
		if order == nil {
			return nil
		}
		if err := g.ResolveTriggerOrder(order.ID, me, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	return nil
}

// kodamaHandCard puts a card with no catalog entry in the hand.
func kodamaHandCard(g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	return handCardOf(g, p, name, typeLine, cost, "")
}

func TestKodamaIsFull(t *testing.T) {
	spec, ok := Lookup(kodamaOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Triggered) != 1 || !spec.Triggered[0].NotPutByThisAbility {
		t.Fatalf("Kodama of the East Tree: %+v, want Full with the not-put-by-this-ability trigger", spec)
	}
}

// A land played normally triggers it; the land it then puts does not
// trigger it again (the chain stops at one), and a card with greater
// mana value is never offered.
func TestKodamaLandPlayTriggersAndItsOwnPutDoesNot(t *testing.T) {
	g, me, _ := kodamaTable(t)
	second := kodamaHandCard(g, me, "Second Forest", "Basic Land — Forest", "")
	bear := kodamaHandCard(g, me, "Bear", "Creature — Bear", "{1}{G}")
	playLandFromHand(t, g, "Forest", "")
	c := kodamaPrompt(t, g, me.ID)
	if c == nil {
		t.Fatal("a land played normally did not trigger Kodama")
	}
	for _, id := range c.ChooseCards {
		if id == bear {
			t.Fatal("a two-drop was offered for a land's mana value of 0")
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{second}); err != nil {
		t.Fatalf("put the second Forest: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(second) {
		t.Fatal("the second Forest did not enter")
	}
	if c := latestPutFromHandChoice(g, me.ID); c != nil || g.Stack.Size() != 0 {
		t.Fatal("the permanent Kodama put triggered Kodama again")
	}
}

// No infinite chain: a creature cast triggers it, it puts a two-drop,
// and that two-drop does not trigger it, though more cards remain.
func TestKodamaDoesNotChainOffItself(t *testing.T) {
	g, me, _ := kodamaTable(t)
	cheap := []uuid.UUID{
		kodamaHandCard(g, me, "Cub A", "Creature — Cat", "{1}{G}"),
		kodamaHandCard(g, me, "Cub B", "Creature — Cat", "{1}{G}"),
		kodamaHandCard(g, me, "Cub C", "Creature — Cat", "{1}{G}"),
	}
	bear := kodamaHandCard(g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the bear: %v", err)
	}
	c := kodamaPrompt(t, g, me.ID)
	if c == nil {
		t.Fatal("a creature that resolved did not trigger Kodama")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{cheap[0]}); err != nil {
		t.Fatalf("put a cub: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap[0]) {
		t.Fatal("the cub did not enter")
	}
	if c := latestPutFromHandChoice(g, me.ID); c != nil {
		t.Fatal("Kodama triggered off its own put")
	}
	if !me.Hand.Contains(cheap[1]) || !me.Hand.Contains(cheap[2]) {
		t.Error("more cards left the hand than the one Kodama put")
	}
}

// Two Kodamas feed each other: what one puts triggers the other (it
// was not put with THAT ability), and each stops at its own put. The
// second is a nonlegendary copy (a Mirror Gallery table, or a token copy
// that is not legendary), so the legend rule leaves both.
func TestTwoKodamasTriggerEachOther(t *testing.T) {
	g, me, _ := kodamaTable(t)
	pushCatalogPermanent(g, me.ID, "Kodama of the East Tree", "Creature — Spirit", kodamaOracle, false)
	a := kodamaHandCard(g, me, "Land A", "Basic Land — Forest", "")
	b := kodamaHandCard(g, me, "Land B", "Basic Land — Forest", "")
	c := kodamaHandCard(g, me, "Land C", "Basic Land — Forest", "")
	playLandFromHand(t, g, "Forest", "")
	puts := 0
	for i := 0; i < 8; i++ {
		p := kodamaPrompt(t, g, me.ID)
		if p == nil {
			break
		}
		var pick []uuid.UUID
		for _, id := range []uuid.UUID{a, b, c} {
			if me.Hand.Contains(id) {
				pick = []uuid.UUID{id}
				break
			}
		}
		if err := g.ResolveChooseCards(p.ID, me.ID, pick); err != nil {
			t.Fatalf("put: %v", err)
		}
		if len(pick) > 0 {
			puts++
		}
	}
	// The land play triggers both (two puts); each put triggers only the
	// other Kodama (two more); that is every land in hand and more.
	if puts != 3 || me.Hand.Contains(a) || me.Hand.Contains(b) || me.Hand.Contains(c) {
		t.Errorf("%d puts, want all three lands put by the two Kodamas", puts)
	}
}

// "Another": Kodama entering does not trigger itself.
func TestKodamaDoesNotTriggerOnItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	kodamaHandCard(g, me, "Cub", "Creature — Cat", "{1}{G}")
	castCatalogSpell(t, g, "Kodama of the East Tree", "Legendary Creature — Spirit", kodamaOracle, nil)
	if c := kodamaPrompt(t, g, me.ID); c != nil {
		t.Fatal("Kodama triggered on its own entry")
	}
}

// The entry's EventETB names the ability that made it, and only that:
// a land play names none.
func TestEntryEventNamesTheAbilityThatPutIt(t *testing.T) {
	g, me, k := kodamaTable(t)
	forest := kodamaHandCard(g, me, "Second Forest", "Basic Land — Forest", "")
	played := playLandFromHand(t, g, "Forest", "")
	c := kodamaPrompt(t, g, me.ID)
	if c == nil {
		t.Fatal("no Kodama prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("put: %v", err)
	}
	passPriorityAroundTable(t, g)
	var landETB, putETB *game.Event
	for i := range g.Events {
		ev := &g.Events[i]
		if ev.Kind != game.EventETB {
			continue
		}
		switch ev.CardID {
		case played:
			landETB = ev
		case forest:
			putETB = ev
		}
	}
	if landETB == nil || putETB == nil {
		t.Fatal("missing an EventETB")
	}
	if landETB.CauseSource != nil || landETB.CauseAbility != nil {
		t.Errorf("a land play names an ability: %+v / %+v", landETB.CauseSource, landETB.CauseAbility)
	}
	if putETB.CauseSource == nil || putETB.CauseSource.ID != k || putETB.CauseAbility == nil ||
		putETB.CauseAbility.Slot != game.AbilitySlotTriggered {
		t.Errorf("Kodama's put names %+v / %+v, want Kodama's triggered row", putETB.CauseSource, putETB.CauseAbility)
	}
}

func TestRegisterRefusesNotPutByThisAbilityOffAnEntryTrigger(t *testing.T) {
	row := On(game.EventDiscardCard, func(game.Event, *game.Card, game.Characteristic, *game.Game) bool { return true },
		"Test — discard", func(*game.Game, *game.StackItem) error { return nil })
	row.NotPutByThisAbility = true
	msg := registerPanics(Spec{OracleID: "test-3933-not-put", Name: "Test not put", Triggered: []game.TriggeredAbility{row}})
	if !strings.Contains(msg, "#3933") {
		t.Errorf("panic %q, want the #3933 guard", msg)
	}
}
