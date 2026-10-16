package game

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// outside_collection_test.go — ADR 0198: a player whose deck seated no
// sideboard takes "a card you own from outside the game" from the card
// database, and the card is minted as a real card they own.

// fakeCollection is an OutsideCollectionSource over a fixed list, in
// list order.
type fakeCollection struct{ cards []Card }

func (f fakeCollection) Search(q string, limit int) []Card {
	var out []Card
	for _, c := range f.cards {
		if strings.Contains(strings.ToLower(c.Name), strings.ToLower(q)) && len(out) < limit {
			out = append(out, c)
		}
	}
	return out
}

func (f fakeCollection) Named(name string) (Card, bool) {
	for _, c := range f.cards {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return Card{}, false
}

func (f fakeCollection) Suggest(fn func(Card) bool) {
	for _, c := range f.cards {
		if !fn(c) {
			return
		}
	}
}

// collectionCard is a database template: no instance, no owner.
func collectionCard(name, typeLine string) Card {
	c := NewCard(name, uuid.Nil)
	c.InstanceID = uuid.Nil
	c.TypeLine = typeLine
	c.OracleID = "oracle-" + name
	c.ScryfallID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(name)).String()
	return c
}

// withCollection installs a collection for the test.
func withCollection(t *testing.T, cards ...Card) {
	t.Helper()
	prev := OutsideCollection
	OutsideCollection = fakeCollection{cards: cards}
	t.Cleanup(func() { OutsideCollection = prev })
}

func standardCollection(t *testing.T) {
	withCollection(t,
		collectionCard("Lightning Bolt", "Instant"),
		collectionCard("Burning Wish", "Sorcery"),
		collectionCard("Llanowar Elves", "Creature — Elf Druid"),
		collectionCard("Counterspell", "Instant"),
	)
}

func isInstant(c Card) bool { return c.IsInstant() }

func TestThePoolIsTheSideboardOnlyWhenTheDeckSeatedOne(t *testing.T) {
	standardCollection(t)
	g := newSideboardGame(t, sbCard("Lightning Bolt", "Instant"))
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if !g.HasSideboardPool(me.ID) || g.HasSideboardPool(opp.ID) {
			t.Fatalf("pools: me %v, opponent %v", g.HasSideboardPool(me.ID), g.HasSideboardPool(opp.ID))
		}
		// Empty the sideboard: it is still the pool, and an empty pool
		// asks nothing, database or no database.
		bolt := me.SideboardCards()[0]
		if err := g.BringFromOutsideForEffect(me.ID, bolt.InstanceID); err != nil {
			t.Fatal(err)
		}
		if !g.HasSideboardPool(me.ID) {
			t.Error("an emptied sideboard stopped being the pool")
		}
		if g.QueueOutsidePickForEffect(OutsidePickPrompt{Player: me.ID, Match: isInstant}) {
			t.Error("an emptied sideboard fell back to the database")
		}
		// The opponent has no sideboard: the database.
		if !g.QueueOutsidePickForEffect(OutsidePickPrompt{Player: opp.ID, Match: isInstant}) {
			t.Fatal("a seat with no sideboard was offered nothing")
		}
		if c := g.PendingChoices[0]; c.Kind != PendingChoiceOutsideCollection {
			t.Errorf("kind %s", c.Kind)
		}
	})
}

func TestReplaceDeckResetsThePool(t *testing.T) {
	g := NewGame()
	p, err := g.AddPlayer("P1", append(buildTestDeck("Commander"), sbCard("Old", "Sorcery")))
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasSideboard {
		t.Fatal("a seated sideboard did not mark the pool")
	}
	if err := g.ReplaceDeck(p.ID, buildTestDeck("Other")); err != nil {
		t.Fatal(err)
	}
	if p.HasSideboard {
		t.Error("a deck without a sideboard kept the old deck's pool")
	}
}

func TestCollectionPickSuggestsOnlyMatchingCards(t *testing.T) {
	g := newSideboardGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		// No collection installed: nothing outside the game.
		if g.QueueOutsidePickForEffect(OutsidePickPrompt{Player: me.ID, Match: isInstant}) {
			t.Fatal("a prompt went up with no collection installed")
		}
	})
	standardCollection(t)
	g.WithWriteLock(func() {
		if g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Match: func(c Card) bool { return c.IsLand() },
		}) {
			t.Error("a prompt went up with nothing in the database matching")
		}
		if !g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Question: "an instant", Match: isInstant,
		}) {
			t.Fatal("no prompt for an instant")
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d choices", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != PendingChoiceOutsideCollection || c.Chooser != me.ID || c.ChooseMin != 0 || c.ChooseMax != 1 ||
		c.ChooseDestination != ChooseFromOutside || c.Reason != "an instant" {
		t.Errorf("prompt %+v", c)
	}
	if got := strings.Join(c.CollectionSuggestions, ","); got != "Lightning Bolt,Counterspell" {
		t.Errorf("suggestions %q", got)
	}
}

func TestCollectionPickMintsARealCardAndRunsThen(t *testing.T) {
	standardCollection(t)
	g := newSideboardGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var got []uuid.UUID
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Match: isInstant,
			Then: func(g *Game, picked []uuid.UUID) error {
				got = picked
				return g.BringFromOutsideForEffect(me.ID, picked[0])
			},
		})
	})
	choice := g.PendingChoices[0]
	handBefore := me.Hand.Size()
	before := len(g.Events)
	if err := g.ResolveOutsideCollection(choice.ID, me.ID, []string{"counterspell"}); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 0 {
		t.Error("the prompt is still open")
	}
	if len(got) != 1 {
		t.Fatalf("Then got %v", got)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand %d, want %d", me.Hand.Size(), handBefore+1)
	}
	var card *Card
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == got[0] {
			card = &me.Hand.Cards[i]
		}
	}
	if card == nil {
		t.Fatal("the minted card is not in the hand")
	}
	// The card is the database's, owned by the chooser and known only to
	// them (nothing revealed it).
	if card.Name != "Counterspell" || card.OracleID != "oracle-Counterspell" || card.ScryfallID == "" {
		t.Errorf("minted %q oracle %q printing %q", card.Name, card.OracleID, card.ScryfallID)
	}
	if card.Owner != me.ID || card.Controller != me.ID || card.IsSideboard {
		t.Errorf("owner %v controller %v sideboard %v", card.Owner, card.Controller, card.IsSideboard)
	}
	if !card.IsKnownTo(me.ID) || card.IsKnownTo(opp.ID) {
		t.Errorf("known to me %v, to the opponent %v", card.IsKnownTo(me.ID), card.IsKnownTo(opp.ID))
	}
	// The pool is still the database; the container is empty again.
	if me.HasSideboard || len(me.SideboardCards()) != 0 {
		t.Errorf("HasSideboard %v, container %d", me.HasSideboard, len(me.SideboardCards()))
	}
	moved := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventZoneMove && ev.CardID == got[0] && ev.OldZone == ZoneOutside && ev.NewZone == ZoneHand {
			moved = true
		}
	}
	if !moved {
		t.Error("no zone move from outside the game to the hand")
	}
}

func TestCollectionPickRefusesAnswersTheEffectWouldNot(t *testing.T) {
	standardCollection(t)
	g := newSideboardGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ran := false
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Match: isInstant, Min: 1,
			Then: func(*Game, []uuid.UUID) error { ran = true; return nil },
		})
	})
	id := g.PendingChoices[0].ID
	cases := []struct {
		name  string
		who   uuid.UUID
		names []string
		want  error
	}{
		{"the filter refuses a sorcery", me.ID, []string{"Burning Wish"}, ErrInvalidParam},
		{"a card the database does not have", me.ID, []string{"Not A Card"}, ErrCardNotFound},
		{"over the ceiling", me.ID, []string{"Lightning Bolt", "Counterspell"}, ErrInvalidParam},
		{"under a floor of one", me.ID, nil, ErrInvalidParam},
		{"an empty name", me.ID, []string{" "}, ErrInvalidParam},
		{"another seat", opp.ID, []string{"Lightning Bolt"}, ErrNotTheChooser},
	}
	for _, tc := range cases {
		if err := g.ResolveOutsideCollection(id, tc.who, tc.names); !errors.Is(err, tc.want) {
			t.Errorf("%s: err %v, want %v", tc.name, err, tc.want)
		}
	}
	if ran || len(g.PendingChoices) != 1 {
		t.Fatalf("a refused answer ran the effect (%v) or closed the prompt (%d open)", ran, len(g.PendingChoices))
	}
	g.WithWriteLock(func() {
		if g.CollectionPickLegalLocked(g.PendingChoices[0], []string{"Burning Wish"}) {
			t.Error("the enumerator's check accepts what the resolver refuses")
		}
		if !g.CollectionPickLegalLocked(g.PendingChoices[0], []string{"Lightning Bolt"}) {
			t.Error("the enumerator's check refuses a legal answer")
		}
	})
	if err := g.ResolveOutsideCollection(id, me.ID, []string{"Lightning Bolt"}); err != nil || !ran {
		t.Fatalf("a legal answer: err %v, ran %v", err, ran)
	}
}

func TestCollectionPickChecksTheSetRule(t *testing.T) {
	standardCollection(t)
	g := newSideboardGame(t)
	me := g.Seats[0]
	var got []uuid.UUID
	twoNames := func(picked []Card) bool { return len(picked) == 2 && picked[0].Name != picked[1].Name }
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Min: 0, Max: 2, Validate: twoNames,
			Then: func(_ *Game, picked []uuid.UUID) error { got = picked; return nil },
		})
	})
	c := g.PendingChoices[0]
	if len(c.CollectionSuggestions) < 2 {
		t.Fatalf("a two-card pick carries %d suggestions", len(c.CollectionSuggestions))
	}
	for _, bad := range [][]string{{"Lightning Bolt"}, {"Lightning Bolt", "Lightning Bolt"}} {
		if err := g.ResolveOutsideCollection(c.ID, me.ID, bad); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%v: err %v", bad, err)
		}
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt", "Llanowar Elves"}); err != nil {
		t.Fatal(err)
	}
	// Both minted, both still outside the game: the continuation decides.
	if len(got) != 2 || len(me.SideboardCards()) != 2 {
		t.Fatalf("minted %v, container %d", got, len(me.SideboardCards()))
	}
	if me.HasSideboard {
		t.Error("a card left outside made the pool a sideboard")
	}
}

func TestDecliningACollectionPickRunsThenWithNothing(t *testing.T) {
	standardCollection(t)
	g := newSideboardGame(t)
	me := g.Seats[0]
	ran, got := false, []uuid.UUID{uuid.New()}
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Match: isInstant,
			Then: func(_ *Game, picked []uuid.UUID) error { ran, got = true, picked; return nil },
		})
	})
	if err := g.ResolveOutsideCollection(g.PendingChoices[0].ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if !ran || len(got) != 0 || me.Sideboard != nil {
		t.Errorf("ran %v with %v; container %v", ran, got, me.Sideboard)
	}
}

// ADR 0167 §5: the database is closed to a controlled player like the
// sideboard is, both when the prompt would be asked and when it is
// answered.
func TestCollectionIsClosedToAControlledPlayer(t *testing.T) {
	standardCollection(t)
	g := newSideboardGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	control := PlayerControl{Controller: opp.ID, Controlled: me.ID, Until: ControlUntilTurn, Stage: ControlTurn}
	var got []uuid.UUID
	g.WithWriteLock(func() {
		control.TurnSeq = g.Turn.Seq
		g.PlayerControls = append(g.PlayerControls, control)
		if g.QueueOutsidePickForEffect(OutsidePickPrompt{Player: me.ID, Match: isInstant}) {
			t.Error("a controlled player was asked")
		}
		g.PlayerControls = nil
		g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Match: isInstant,
			Then: func(_ *Game, picked []uuid.UUID) error { got = picked; return nil },
		})
		g.PlayerControls = append(g.PlayerControls, control)
	})
	if err := g.ResolveOutsideCollection(g.PendingChoices[0].ID, me.ID, []string{"Lightning Bolt"}); err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 || me.Sideboard != nil {
		t.Errorf("a closed pool minted %v", got)
	}
}

func TestSearchOutsideCollectionAnswersOnlyTheChooser(t *testing.T) {
	standardCollection(t)
	g := newSideboardGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(OutsidePickPrompt{Player: me.ID, Match: isInstant})
	})
	id := g.PendingChoices[0].ID
	found, err := g.SearchOutsideCollection(id, me.ID, "l", 0)
	if err != nil {
		t.Fatal(err)
	}
	// "l" matches Lightning Bolt, Llanowar Elves and Counterspell; the
	// filter keeps the instants.
	var names []string
	for _, c := range found {
		names = append(names, c.Name)
	}
	if strings.Join(names, ",") != "Lightning Bolt,Counterspell" {
		t.Errorf("search found %v", names)
	}
	if _, err := g.SearchOutsideCollection(id, opp.ID, "l", 0); !errors.Is(err, ErrNotACollectionPick) {
		t.Errorf("another seat searched: %v", err)
	}
	if _, err := g.SearchOutsideCollection(uuid.New(), me.ID, "l", 0); !errors.Is(err, ErrNotACollectionPick) {
		t.Errorf("an unknown choice searched: %v", err)
	}
	if found, err := g.SearchOutsideCollection(id, me.ID, "  ", 0); err != nil || found != nil {
		t.Errorf("an empty query found %v (%v)", found, err)
	}
}

func TestHasSideboardSurvivesCloneAndSnapshot(t *testing.T) {
	g := newSideboardGame(t, sbCard("Lightning Bolt", "Instant"))
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.BringFromOutsideForEffect(me.ID, me.SideboardCards()[0].InstanceID); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Clone().Seats[0].HasSideboard {
		t.Error("Clone lost the pool")
	}
	_, restored := roundTrip(t, g)
	if !restored.Seats[0].HasSideboard || restored.Seats[1].HasSideboard {
		t.Errorf("restored pools: %v, %v", restored.Seats[0].HasSideboard, restored.Seats[1].HasSideboard)
	}

	// A file written before the key: a sideboard that still holds a card
	// is the pool.
	g2 := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"))
	raw, err := json.Marshal(g2.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	seat := doc["seats"].([]any)[0].(map[string]any)
	if seat["hasSideboard"] != true {
		t.Fatalf("the key was not written: %v", seat["hasSideboard"])
	}
	delete(seat, "hasSideboard")
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	old, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if !old.Seats[0].HasSideboard {
		t.Error("a pre-key file with a sideboard restored as a database pool")
	}
}
