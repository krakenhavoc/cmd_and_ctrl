package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// put_from_library_test.go — #745: the library-to-battlefield move and
// the public random-order bottom.
//
// The move is pinned the way put_from_hand_test.go pins the hand move
// (#728): an ORDINARY entry — the CR 614 pipeline, the ETB event, the
// catalog hook — that is NOT a land play. On top of that it must not
// be a search (no EventSearchLibrary, no shuffle), and the batch form
// must run every pipeline before anything enters.

// topOfLibraryFor pushes a card on top of a player's library and
// returns its ID.
func topOfLibraryFor(p *Player, name, typeLine string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Library.PushTop(c)
	return c.InstanceID
}

func putFromLibrary(t *testing.T, g *Game, id uuid.UUID, opts LibraryEntryOptions) (uuid.UUID, error) {
	t.Helper()
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		entered, err = g.PutFromLibraryOntoBattlefieldForEffect(id, opts)
	})
	return entered, err
}

// The base case: the land leaves the library for the battlefield under
// its owner's control, untapped and public — and it was neither a land
// play (CR 305.4) nor a search.
func TestPutFromLibraryEntersAndIsNeitherALandDropNorASearch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := topOfLibraryFor(me, "Forest", "Basic Land — Forest")
	below := libraryIDs(me)[:len(me.Library.Cards)-1]
	before := len(g.Events)

	entered, err := putFromLibrary(t, g, land, LibraryEntryOptions{})
	if err != nil {
		t.Fatalf("PutFromLibraryOntoBattlefieldForEffect: %v", err)
	}
	if entered != land {
		t.Errorf("entered as %s, want the library instance %s", entered, land)
	}
	if me.Library.Contains(land) {
		t.Error("the land is still in the library")
	}
	c := battlefieldCardForTest(t, g, land)
	if c.Controller != me.ID || c.Tapped {
		t.Errorf("controller %s tapped %v, want the owner and untapped", c.Controller, c.Tapped)
	}
	for _, seat := range g.Seats {
		if !c.IsKnownTo(seat.ID) {
			t.Errorf("seat %s does not know the land: the battlefield is public", seat.Name)
		}
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d, want 0 (CR 305.4)", n)
	}
	var sawMove, sawETB bool
	for _, ev := range g.Events[before:] {
		switch ev.Kind {
		case EventSearchLibrary:
			t.Error("a reveal-and-put is not a search: EventSearchLibrary fired")
		case EventZoneMove:
			if ev.CardID == land && ev.OldZone == ZoneLibrary && ev.NewZone == ZoneBattlefield {
				sawMove = true
			}
		case EventETB:
			if ev.CardID == land {
				sawETB = true
			}
		case EventTapCard:
			t.Error("nothing was tapped")
		}
	}
	if !sawMove || !sawETB {
		t.Errorf("zone move %v, ETB %v: both are owed by an entry", sawMove, sawETB)
	}
	got := libraryIDs(me)
	if len(got) != len(below) {
		t.Fatalf("library %d cards, want %d", len(got), len(below))
	}
	for i := range below {
		if got[i] != below[i] {
			t.Fatal("the library was reordered: nothing here shuffles")
		}
	}
}

// A card's own "enters tapped" replacement and the effect's own
// "tapped" clause both reach the permanent, which ARRIVES tapped.
func TestPutFromLibraryRunsThePipelineAndTheTappedOption(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hollow := topOfLibraryFor(me, "Jungle Hollow", "Land")
	forest := topOfLibraryFor(me, "Forest", "Basic Land — Forest")
	before := len(g.Events)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield && ev.CardID == hollow
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.EntersTapped = true
				ev.EntersWithCounters = map[string]int{"+1/+1": 1}
				return nil
			},
			Label: "Test: this land enters tapped with a counter",
		})
	})
	if _, err := putFromLibrary(t, g, hollow, LibraryEntryOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := putFromLibrary(t, g, forest, LibraryEntryOptions{Tapped: true}); err != nil {
		t.Fatal(err)
	}
	h := battlefieldCardForTest(t, g, hollow)
	if !h.Tapped || h.Counters["+1/+1"] != 1 {
		t.Errorf("tapped %v counters %d: the entry replacement was not applied", h.Tapped, h.Counters["+1/+1"])
	}
	if !battlefieldCardForTest(t, g, forest).Tapped {
		t.Error("the effect said tapped and the land arrived untapped")
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventTapCard {
			t.Error("a permanent put onto the battlefield tapped must not emit a tap event")
		}
	}
}

// "Under your control" from another player's library (Lonis): the
// controller is stamped BEFORE the pipeline runs.
func TestPutFromLibraryHonoursAnExplicitController(t *testing.T) {
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	card := topOfLibraryFor(owner, "Bear", "Creature — Bear")

	var sawController uuid.UUID
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, gg *Game, _ *Card) bool {
				if ev.Kind == RepEventMove && ev.CardID == card {
					if c, ok := gg.LookupCardForEffect(card); ok {
						sawController = c.Controller
					}
				}
				return false
			},
			Label: "Test: observe the entering permanent's controller",
		})
	})
	if _, err := putFromLibrary(t, g, card, LibraryEntryOptions{Controller: thief.ID}); err != nil {
		t.Fatal(err)
	}
	if sawController != thief.ID {
		t.Errorf("the pipeline saw controller %s, want %s", sawController, thief.ID)
	}
	c := battlefieldCardForTest(t, g, card)
	if c.Controller != thief.ID || c.Owner != owner.ID {
		t.Errorf("controller %s owner %s: control changes, ownership does not", c.Controller, c.Owner)
	}
}

// A canceled entry leaves the card in the library and is not an error.
func TestPutFromLibraryCanceledLeavesTheCardInTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := topOfLibraryFor(me, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == land
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Canceled = true
				return nil
			},
			Label: "Test: this entry is replaced with nothing",
		})
	})
	var prior uuid.UUID
	g.WithWriteLock(func() {
		c, _ := g.LookupCardForEffect(land)
		prior = c.Controller
	})
	// Put under another player's control (the Lonis shape), so the
	// controller stamped for the entry differs from what the card had.
	entered, err := putFromLibrary(t, g, land, LibraryEntryOptions{Controller: g.Seats[1].ID})
	if err != nil || entered != uuid.Nil {
		t.Errorf("entered %s err %v, want uuid.Nil and nil", entered, err)
	}
	if !me.Library.Contains(land) {
		t.Error("the canceled land left the library anyway")
	}
	g.WithWriteLock(func() {
		if c, _ := g.LookupCardForEffect(land); c.Controller != prior {
			t.Errorf("the canceled card kept the entry's controller %s, want %s back", c.Controller, prior)
		}
	})
}

// It is a LIBRARY move: a card in a hand, on the battlefield or nowhere
// is refused, and so is a nonpermanent card (CR 110.4). The hand move
// refuses a library card in turn — one body, two sources, no crossing.
func TestPutFromLibraryRefusesTheWrongZoneAndNonpermanents(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	inHand := handCardFor(me, "Forest", "Basic Land — Forest")
	onBattlefield := permanentFor(g, me, "Bear", "Creature — Bear", "{1}{G}")
	bolt := topOfLibraryFor(me, "Lightning Bolt", "Instant")
	land := topOfLibraryFor(me, "Island", "Basic Land — Island")
	// #773 review: a token tucked into a library earlier (the engine has
	// no CR 704.5d sweep) is not a card (CR 108.2) and can't come back
	// onto the battlefield (CR 111.8).
	token := topOfLibraryFor(me, "Goblin", "Token Creature — Goblin")

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want error
	}{
		{"in a hand", inHand, ErrCardNotFound},
		{"on the battlefield", onBattlefield, ErrCardNotFound},
		{"nowhere at all", uuid.New(), ErrCardNotFound},
		{"an instant", bolt, ErrInvalidParam},
		{"a token", token, ErrInvalidParam},
	} {
		if _, err := putFromLibrary(t, g, tc.id, LibraryEntryOptions{}); !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	var err error
	g.WithWriteLock(func() {
		_, err = g.PutFromHandOntoBattlefieldForEffect(land, HandEntryOptions{})
	})
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("the hand move took a library card: err = %v", err)
	}
	if !me.Hand.Contains(inHand) || !me.Library.Contains(bolt) || !me.Library.Contains(land) || !me.Library.Contains(token) {
		t.Error("a refused put moved something anyway")
	}
}

// The batch runs every CR 614 pipeline against the board as it stood
// before any card entered. A check land ("enters tapped unless you
// control a Forest") put alongside a Forest does not see that Forest —
// entering them one at a time would let it, which is stronger than
// printed.
func TestPutCardsFromLibraryEnterSimultaneously(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := topOfLibraryFor(me, "Forest", "Basic Land — Forest")
	check := topOfLibraryFor(me, "Hinterland Harbor", "Land")

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, gg *Game, _ *Card) bool {
				if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.CardID != check {
					return false
				}
				for _, c := range gg.Battlefield.Cards {
					if c.Controller == me.ID && c.HasSubtype("Forest") {
						return false
					}
				}
				return true
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.EntersTapped = true
				return nil
			},
			Label: "Test: enters tapped unless you control a Forest",
		})
	})
	var entered []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		// The Forest first, which is the order that would leak it.
		entered, err = g.PutCardsFromLibraryOntoBattlefieldForEffect([]uuid.UUID{forest, check, forest}, LibraryEntryOptions{})
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entered) != 2 {
		t.Fatalf("entered %d, want 2 (a repeated ID is taken once)", len(entered))
	}
	if !battlefieldCardForTest(t, g, check).Tapped {
		t.Error("the check land saw a Forest that entered alongside it")
	}
}

// One bad ID refuses the whole batch before anything moves.
func TestPutCardsFromLibraryRefusesTheWholeBatch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := topOfLibraryFor(me, "Forest", "Basic Land — Forest")
	bolt := topOfLibraryFor(me, "Lightning Bolt", "Instant")
	var err error
	g.WithWriteLock(func() {
		_, err = g.PutCardsFromLibraryOntoBattlefieldForEffect([]uuid.UUID{land, bolt}, LibraryEntryOptions{})
	})
	if !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if !me.Library.Contains(land) {
		t.Error("the land entered although its batch was refused")
	}
}

// --- random-order bottom --------------------------------------------

// From exile the move is a real zone change through the exit path: the
// cards land on the bottom, forget their impulse grant and their
// knowers, and emit a zone move each.
func TestPutOnBottomInRandomOrderFromExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	top := topOfLibraryFor(me, "Top", "Land")
	var pile []uuid.UUID
	for i := 0; i < 3; i++ {
		c := NewCard("Exiled", me.ID)
		c.TypeLine = "Instant"
		c.ExilePlay = ExilePlayPermission{Player: me.ID, UntilTurn: 99}
		c.AddKnowersAll([]uuid.UUID{g.Seats[0].ID, g.Seats[1].ID})
		g.Exile.PushTop(c)
		pile = append(pile, c.InstanceID)
	}
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.PutOnBottomInRandomOrderForEffect(me.ID, ZoneExile, pile); err != nil {
			t.Fatal(err)
		}
	})
	if len(g.Exile.Cards) != 0 {
		t.Fatalf("%d cards stayed in exile", len(g.Exile.Cards))
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:3] {
		bottom[c.InstanceID] = true
		if c.ExilePlay.Player != uuid.Nil {
			t.Error("a card kept its exile grant in the library (CR 400.7)")
		}
		if c.IsKnownTo(g.Seats[1].ID) || c.IsKnownTo(me.ID) {
			t.Error("a card on the bottom of a library is still known")
		}
	}
	for _, id := range pile {
		if !bottom[id] {
			t.Errorf("card %s is not among the bottom three", id)
		}
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != top {
		t.Error("the top card moved")
	}
	moves := 0
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventZoneMove && ev.OldZone == ZoneExile && ev.NewZone == ZoneLibrary {
			moves++
		}
	}
	if moves != 3 {
		t.Errorf("%d zone moves, want 3", moves)
	}
}

// The exit path opens the CR 614 window, which the old private cascade
// helper never did: a replacement on "would be put into a library"
// sees the move.
func TestPutOnBottomInRandomOrderOpensTheReplacementWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := NewCard("Exiled", me.ID)
	c.TypeLine = "Instant"
	g.Exile.PushTop(c)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == c.InstanceID && ev.NewZone == ZoneLibrary
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Canceled = true
				return nil
			},
			Label: "Test: it stays in exile instead",
		})
		if err := g.PutOnBottomInRandomOrderForEffect(me.ID, ZoneExile, []uuid.UUID{c.InstanceID}); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("the move skipped the replacement window")
	}
}

// Cards still in their owner's library (revealed and passed over) are
// REORDERED to the bottom: no zone change, no zone-move event, and the
// table forgets them.
func TestPutOnBottomInRandomOrderReordersLibraryCards(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := topOfLibraryFor(me, "A", "Instant")
	b := topOfLibraryFor(me, "B", "Instant")
	g.WithWriteLock(func() {
		g.RevealForEffect(RevealSpec{Player: me.ID, Cards: []uuid.UUID{a, b}})
	})
	size := me.Library.Size()
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.PutOnBottomInRandomOrderForEffect(me.ID, ZoneLibrary, []uuid.UUID{a, b}); err != nil {
			t.Fatal(err)
		}
	})
	if me.Library.Size() != size {
		t.Fatalf("library %d → %d", size, me.Library.Size())
	}
	for _, c := range me.Library.Cards[:2] {
		if c.InstanceID != a && c.InstanceID != b {
			t.Fatalf("%s is on the bottom, want the two revealed cards", c.Name)
		}
		if c.IsKnownTo(g.Seats[1].ID) {
			t.Error("a revealed card is still known after going to the bottom in a random order")
		}
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventZoneMove {
			t.Error("a reorder within a library is not a zone change")
		}
	}
}

// The order comes from the game's seeded source: two games from the
// same seed bottom the same pile in the same order. Compared by name
// position, since instance IDs are minted fresh per game.
func TestPutOnBottomInRandomOrderIsReproducible(t *testing.T) {
	names := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	order := func() []string {
		g := newActiveGame(t)
		me := g.Seats[0]
		ids := make([]uuid.UUID, 0, len(names))
		for _, n := range names {
			ids = append(ids, topOfLibraryFor(me, n, "Instant"))
		}
		g.WithWriteLock(func() {
			if err := g.PutOnBottomInRandomOrderForEffect(me.ID, ZoneLibrary, ids); err != nil {
				t.Fatal(err)
			}
		})
		out := make([]string, 0, len(names))
		for _, c := range me.Library.Cards[:len(names)] {
			out = append(out, c.Name)
		}
		return out
	}
	a, b := order(), order()
	identity := true
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("same seed, different order: %v vs %v", a, b)
		}
		// Bottom-first vs pushed top-first: the untouched order would
		// read H..A from the bottom.
		if a[i] != names[len(names)-1-i] {
			identity = false
		}
	}
	if identity {
		t.Error("eight cards came out in exactly their original order; the shuffle did not run")
	}
}

// --- look at the top N, and the graveyard put -----------------------

func TestLookAtTopOfLibraryTellsOnlyTheLooker(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := topOfLibraryFor(me, "A", "Instant")
	b := topOfLibraryFor(me, "B", "Instant")
	var ids []uuid.UUID
	g.WithWriteLock(func() { ids = g.LookAtTopOfLibraryForEffect(me.ID, 2) })
	if len(ids) != 2 || ids[0] != b || ids[1] != a {
		t.Fatalf("looked at %v, want [B A] top first", ids)
	}
	for _, c := range me.Library.Cards[len(me.Library.Cards)-2:] {
		if !c.IsKnownTo(me.ID) {
			t.Error("the looker does not know a card they looked at")
		}
		if c.IsKnownTo(opp.ID) {
			t.Error("an opponent learned a card from a look (CR 701.20 is reveal, not look)")
		}
	}
}

func TestPutIntoGraveyardIsNotAMillAndRefusesPermanents(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := topOfLibraryFor(me, "A", "Instant")
	bear := permanentFor(g, me, "Bear", "Creature — Bear", "{1}{G}")
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.PutIntoGraveyardForEffect(card); err != nil {
			t.Fatal(err)
		}
		if err := g.PutIntoGraveyardForEffect(bear); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("battlefield permanent: err = %v, want ErrInvalidParam", err)
		}
	})
	if !me.Graveyard.Contains(card) || !g.Battlefield.Contains(bear) {
		t.Fatal("wrong card moved")
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventMill {
			t.Error("putting a revealed card into a graveyard is not a mill")
		}
	}
}

// `from` is where the effect left the cards. A saved ID whose card has
// since moved on is skipped, not pulled back out of wherever it went.
func TestPutOnBottomInRandomOrderSkipsCardsThatLeftTheZone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	exiled := NewCard("Still Exiled", me.ID)
	exiled.TypeLine = "Instant"
	g.Exile.PushTop(exiled)
	inHand := NewCard("Cast Meanwhile", me.ID)
	inHand.TypeLine = "Instant"
	me.Hand.PushTop(inHand)
	g.WithWriteLock(func() {
		if err := g.PutOnBottomInRandomOrderForEffect(me.ID, ZoneExile, []uuid.UUID{exiled.InstanceID, inHand.InstanceID}); err != nil {
			t.Fatal(err)
		}
	})
	if !me.Library.Contains(exiled.InstanceID) {
		t.Error("the card still in exile did not go to the bottom")
	}
	if !me.Hand.Contains(inHand.InstanceID) || me.Library.Contains(inHand.InstanceID) {
		t.Error("a card no longer in exile was pulled out of its hand")
	}
}
