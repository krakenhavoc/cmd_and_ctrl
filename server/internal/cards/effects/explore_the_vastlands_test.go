package effects

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// explore_the_vastlands_test.go — #1743 item 3. Explore the Vastlands
// is the back face of Wandering Archaic: "Each player looks at the top
// five cards of their library and may reveal a land card and/or an
// instant or sorcery card from among them. Each player puts the cards
// they revealed this way into their hand and the rest on the bottom of
// their library in a random order. Each player gains 3 life."
//
// Two shapes are under test: TakeFromLibraryToHand.Slots (two
// predicates, one look, one card per slot) and
// EachPlayerTakesFromLibrary (that question asked of every player as
// one instruction). The card is the proof of both.

type vastCard struct{ name, typeLine string }

// vastFive: a land, an instant, a sorcery and two creatures — two of
// the five fit the land slot or the spell slot each way round.
var vastFive = []vastCard{
	{"Forest", "Basic Land — Forest"},
	{"Opt", "Instant"},
	{"Grizzly Bears", "Creature — Bear"},
	{"Preordain", "Sorcery"},
	{"Llanowar Elves", "Creature — Elf Druid"},
}

// stackTop puts `cards` on top of p's library, the first one on top,
// and returns their IDs in that (top-first) order.
func stackTop(p *game.Player, cards []vastCard) []uuid.UUID {
	pairs := make([]struct{ name, typeLine string }, len(cards))
	for i, c := range cards {
		pairs[i] = struct{ name, typeLine string }{c.name, c.typeLine}
	}
	return hornTopFive(p, pairs)
}

// castExplore puts Wandering Archaic // Explore the Vastlands into the
// active player's hand, casts its BACK face, and resolves it.
func castExplore(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := mdfcCard(me.ID, wanderingArchaicOracle, game.LayoutModalDFC,
		game.Face{Name: "Wandering Archaic", TypeLine: "Creature — Avatar", ManaCost: "{5}", Power: 4, Toughness: 4},
		game.Face{Name: "Explore the Vastlands", TypeLine: "Sorcery", ManaCost: "{3}"},
	)
	me.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Explore the Vastlands: %v", err)
	}
	passPriorityAroundTable(t, g)
	return card.InstanceID
}

// apnapSeats is the table from the active player round.
func apnapSeats(g *game.Game) []*game.Player {
	out := make([]*game.Player, 0, len(g.Seats))
	for i := range g.Seats {
		out = append(out, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)])
	}
	return out
}

func vastRevealed(g *game.Game, id uuid.UUID) bool { return hornRevealed(g, id) }

// libraryBottom returns the bottom n cards of p's library, bottom card
// first (Zone.Cards is bottom-first).
func libraryBottom(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for _, c := range p.Library.Cards[:n] {
		out = append(out, c.InstanceID)
	}
	return out
}

func sameSet(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for _, id := range a {
		if !slices.Contains(b, id) {
			return false
		}
	}
	return true
}

func without(ids []uuid.UUID, drop ...uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		if !slices.Contains(drop, id) {
			out = append(out, id)
		}
	}
	return out
}

// TestExploreTheVastlandsWholeCard is the card at a four-player table:
// two predicates picked, one picked, none picked, and a player with
// nothing to pick — then the life.
func TestExploreTheVastlandsWholeCard(t *testing.T) {
	g := newCatalogGame(t)
	seats := apnapSeats(g)
	tops := make([][]uuid.UUID, len(seats))
	for i, p := range seats[:3] {
		tops[i] = stackTop(p, vastFive)
	}
	// The fourth player's five has nothing either slot takes.
	tops[3] = stackTop(seats[3], []vastCard{
		{"Bear A", "Creature — Bear"}, {"Bear B", "Creature — Bear"}, {"Sol Ring", "Artifact"},
		{"Bear C", "Creature — Bear"}, {"Bear D", "Creature — Bear"},
	})
	life := make([]int, len(seats))
	hand := make([]int, len(seats))
	for i, p := range seats {
		life[i], hand[i] = p.Life, p.Hand.Size()
	}

	// castExplore deals the card into the hand it is cast from, so the
	// caster's baseline above does not include it.
	spell := castExplore(t, g)

	// Every player with a candidate is asked, about their own five,
	// up to one land and one instant or sorcery: two cards at most.
	for i, p := range seats[:3] {
		c := latestChooseCards(g, p.ID)
		if c == nil {
			t.Fatalf("seat %d was not asked", i)
		}
		want := []uuid.UUID{tops[i][0], tops[i][1], tops[i][3]}
		if !sameSet(c.ChooseCards, want) {
			t.Errorf("seat %d offered %v, want its land, instant and sorcery", i, c.ChooseCards)
		}
		if c.ChooseMin != 0 || c.ChooseMax != 2 {
			t.Errorf("seat %d bounds [%d,%d], want [0,2]", i, c.ChooseMin, c.ChooseMax)
		}
	}
	if latestChooseCards(g, seats[3].ID) != nil {
		t.Error("a player with no land, instant or sorcery in their five was asked")
	}

	// Answered out of APNAP order. Nothing moves, nothing is revealed
	// and nobody gains life until the last answer is in.
	answer := func(i int, ids ...uuid.UUID) {
		t.Helper()
		c := latestChooseCards(g, seats[i].ID)
		if err := g.ResolveChooseCards(c.ID, seats[i].ID, ids); err != nil {
			t.Fatalf("seat %d answer: %v", i, err)
		}
	}
	answer(2) // declines
	answer(0, tops[0][0], tops[0][1])
	for i, p := range seats {
		if p.Life != life[i] {
			t.Errorf("seat %d gained life with a prompt still open", i)
		}
	}
	if seats[0].Hand.Contains(tops[0][0]) || vastRevealed(g, tops[0][0]) {
		t.Error("an answer was acted on before every player had answered")
	}
	answer(1, tops[1][3])

	taken := [][]uuid.UUID{{tops[0][0], tops[0][1]}, {tops[1][3]}, nil, nil}
	for i, p := range seats {
		for _, id := range taken[i] {
			if !p.Hand.Contains(id) {
				t.Errorf("seat %d: a revealed card is not in hand", i)
			}
			if !vastRevealed(g, id) {
				t.Errorf("seat %d: a card taken into hand was not revealed", i)
			}
		}
		if got := p.Hand.Size(); got != hand[i]+len(taken[i]) {
			t.Errorf("seat %d hand %d, want %d", i, got, hand[i]+len(taken[i]))
		}
		rest := without(tops[i], taken[i]...)
		for _, id := range rest {
			if vastRevealed(g, id) {
				t.Errorf("seat %d: a card that was only looked at was revealed", i)
			}
		}
		if !sameSet(libraryBottom(p, len(rest)), rest) {
			t.Errorf("seat %d: the rest of the look is not on the bottom", i)
		}
		if p.Life != life[i]+3 {
			t.Errorf("seat %d life %d, want %d", i, p.Life, life[i]+3)
		}
	}
	if !seats[0].Graveyard.Contains(spell) {
		t.Error("the sorcery is not in its owner's graveyard")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left open", len(g.PendingChoices))
	}
}

// TestExploreTheVastlandsRefusesAnIllegalPick: two of one kind, a card
// neither slot takes, and three cards are all refused, and the prompt
// stays open for another try.
func TestExploreTheVastlandsRefusesAnIllegalPick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := stackTop(me, []vastCard{
		{"Forest", "Basic Land — Forest"}, {"Island", "Basic Land — Island"}, {"Opt", "Instant"},
		{"Grizzly Bears", "Creature — Bear"}, {"Preordain", "Sorcery"},
	})
	forest, island, opt, bears, preordain := top[0], top[1], top[2], top[3], top[4]
	castExplore(t, g)
	c := latestChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("not asked")
	}

	for _, bad := range []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"two lands", []uuid.UUID{forest, island}, game.ErrChoiceSetRejected},
		{"two spells", []uuid.UUID{opt, preordain}, game.ErrChoiceSetRejected},
		{"a creature", []uuid.UUID{bears}, game.ErrInvalidParam},
		{"three cards", []uuid.UUID{forest, opt, preordain}, game.ErrInvalidParam},
	} {
		if err := g.ResolveChooseCards(c.ID, me.ID, bad.ids); !errors.Is(err, bad.want) {
			t.Errorf("%s: err = %v, want %v", bad.name, err, bad.want)
		}
		if latestChooseCards(g, me.ID) == nil {
			t.Fatalf("%s: a refused answer took the prompt down", bad.name)
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{island, preordain}); err != nil {
		t.Fatalf("a land and a sorcery: %v", err)
	}
	if !me.Hand.Contains(island) || !me.Hand.Contains(preordain) {
		t.Error("the legal pair did not reach the hand")
	}
}

// TestTakeSlotsACardMatchingBothFillsOneSlot is the primitive, alone,
// on the case the two Vastlands predicates never print together but
// real cards do: an artifact creature fits "a creature card" AND "an
// artifact card", and fills one of them, not both.
func TestTakeSlotsACardMatchingBothFillsOneSlot(t *testing.T) {
	slots := []TakeSlot{
		{Label: "a creature card", Match: Creature(), Max: 1},
		{Label: "an artifact card", Match: Artifact(), Max: 1},
	}
	take := func(t *testing.T, cards []vastCard) (*game.Game, *game.Player, []uuid.UUID, *game.PendingChoice) {
		t.Helper()
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		top := stackTop(me, cards)
		g.WithWriteLock(func() {
			ctx := NewContext(g, &game.StackItem{Controller: me.ID})
			err := TakeFromLibraryToHand{
				Player:   me.ID,
				Cards:    g.LookAtTopOfLibraryForEffect(me.ID, len(cards)),
				Slots:    slots,
				Optional: true,
				Reveal:   true,
				Label:    "test — you may reveal a creature card and/or an artifact card",
				Then:     TakeRestOnBottomInRandomOrder,
			}.Apply(ctx)
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}
		})
		return g, me, top, latestChooseCards(g, me.ID)
	}

	// Alone, the artifact creature is one card in one slot: the
	// ceiling is one, not two.
	_, _, _, c := take(t, []vastCard{{"Myr", "Artifact Creature — Myr"}, {"Forest", "Basic Land — Forest"}})
	if c == nil || c.ChooseMax != 1 {
		t.Fatalf("one artifact creature: prompt %v, want a ceiling of 1", c)
	}

	// With two plain creatures beside it: two creatures cannot both be
	// taken (one creature slot), but the artifact creature plus either
	// one can — it moves to the artifact slot.
	g, me, top, c := take(t, []vastCard{
		{"Myr", "Artifact Creature — Myr"}, {"Bear A", "Creature — Bear"}, {"Bear B", "Creature — Bear"},
	})
	myr, bearA, bearB := top[0], top[1], top[2]
	if c == nil || c.ChooseMax != 2 {
		t.Fatalf("prompt %v, want a ceiling of 2", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bearA, bearB}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("two plain creatures: err = %v, want ErrChoiceSetRejected", err)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bearA, myr}); err != nil {
		t.Fatalf("a creature and the artifact creature: %v", err)
	}
	if !me.Hand.Contains(myr) || !me.Hand.Contains(bearA) || me.Hand.Contains(bearB) {
		t.Error("the wrong cards reached the hand")
	}
	if !sameSet(libraryBottom(me, 1), []uuid.UUID{bearB}) {
		t.Error("the untaken creature is not on the bottom")
	}
}

// TestExploreTheVastlandsShortAndEmptyLibraries: a short library is
// looked at as far as it goes, an empty one is a look at nothing, and
// neither is an error.
func TestExploreTheVastlandsShortAndEmptyLibraries(t *testing.T) {
	g := newCatalogGame(t)
	seats := apnapSeats(g)
	g.WithWriteLock(func() {
		seats[1].Library.Cards = nil
		seats[2].Library.Cards = nil
	})
	short := stackTop(seats[1], []vastCard{{"Forest", "Basic Land — Forest"}, {"Grizzly Bears", "Creature — Bear"}})
	for _, p := range []*game.Player{seats[0], seats[3]} {
		g.WithWriteLock(func() { p.Library.Cards = nil })
		stackTop(p, []vastCard{{"Grizzly Bears", "Creature — Bear"}})
	}
	lifeBefore := seats[2].Life

	castExplore(t, g)

	c := latestChooseCards(g, seats[1].ID)
	if c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != short[0] || c.ChooseMax != 1 {
		t.Fatalf("two-card library: prompt %v, want the Forest alone, up to one", c)
	}
	for _, p := range []*game.Player{seats[0], seats[2], seats[3]} {
		if latestChooseCards(g, p.ID) != nil {
			t.Errorf("seat %s with nothing to take was asked", p.Name)
		}
	}
	if err := g.ResolveChooseCards(c.ID, seats[1].ID, []uuid.UUID{short[0]}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if !seats[1].Hand.Contains(short[0]) {
		t.Error("the Forest is not in hand")
	}
	if seats[1].Library.Size() != 1 || seats[1].Library.Cards[0].InstanceID != short[1] {
		t.Error("the Bears should be the whole library")
	}
	if seats[2].Life != lifeBefore+3 {
		t.Error("a player with an empty library still gains 3 life")
	}
	if seats[2].AttemptedEmptyDraw {
		t.Error("looking at an empty library is not a draw")
	}
}

// TestExploreTheVastlandsIsPrivateUntilRevealed: each player is the
// only knower of their own five, no seat sees another seat's prompt,
// only the taken cards are revealed, and the rest go to the bottom
// with nobody — the looker included — knowing where.
func TestExploreTheVastlandsIsPrivateUntilRevealed(t *testing.T) {
	g := newCatalogGame(t)
	seats := apnapSeats(g)
	tops := make([][]uuid.UUID, len(seats))
	for i, p := range seats {
		tops[i] = stackTop(p, vastFive)
	}
	castExplore(t, g)

	for i, p := range seats {
		for _, id := range tops[i] {
			card, _ := g.LookupCardForEffect(id)
			for _, viewer := range seats {
				if got := card.IsKnownTo(viewer.ID); got != (viewer.ID == p.ID) {
					t.Errorf("seat %d's card %q known to %s = %v", i, card.Name, viewer.Name, got)
				}
			}
		}
	}
	for _, viewer := range seats {
		view := protocol.ViewOfGameFor(g, viewer.ID.String())
		for _, c := range view.PendingChoices {
			if c.Kind != string(game.PendingChoiceChooseCards) {
				continue
			}
			mine := c.Chooser == viewer.ID.String()
			if mine && (len(c.Options) != 3 || c.ChooseMax != 2) {
				t.Errorf("%s's own prompt shows %d options, max %d; want 3 and 2", viewer.Name, len(c.Options), c.ChooseMax)
			}
			if !mine && (len(c.Options) != 0 || c.ChooseMax != 0) {
				t.Errorf("%s sees %d options of %s's prompt (max %d); want nothing", viewer.Name, len(c.Options), c.Chooser, c.ChooseMax)
			}
		}
	}

	for i, p := range seats {
		c := latestChooseCards(g, p.ID)
		if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{tops[i][0]}); err != nil {
			t.Fatalf("seat %d: %v", i, err)
		}
	}
	for i, p := range seats {
		if !vastRevealed(g, tops[i][0]) {
			t.Errorf("seat %d's taken land was not revealed", i)
		}
		for _, c := range p.Library.Cards[:4] {
			if len(c.KnownBy) != 0 {
				t.Errorf("seat %d: %q went to the bottom with knowers %v — library order leaked", i, c.Name, c.KnownBy)
			}
		}
	}
}

// TestExploreTheVastlandsBottomsTheRestInARandomOrder: the rest is the
// game's keyed random_order shuffle, not the look's order. Under six
// RNG keys the four cards always land as the same SET and not always
// in the same ORDER.
func TestExploreTheVastlandsBottomsTheRestInARandomOrder(t *testing.T) {
	orders := map[string]bool{}
	for k := 1; k <= 6; k++ {
		g := newCatalogGame(t)
		var key [32]byte
		key[0] = byte(k)
		g.SetRNGKeyForTest(key)
		seats := apnapSeats(g)
		me := seats[0]
		top := stackTop(me, vastFive)
		for _, p := range seats[1:] {
			g.WithWriteLock(func() { p.Library.Cards = nil })
		}
		castExplore(t, g)
		c := latestChooseCards(g, me.ID)
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{top[0]}); err != nil {
			t.Fatalf("key %d: %v", k, err)
		}
		rest := without(top, top[0])
		bottom := libraryBottom(me, len(rest))
		if !sameSet(bottom, rest) {
			t.Fatalf("key %d: bottom %v is not the rest %v", k, bottom, rest)
		}
		orders[fmt.Sprint(bottom)] = true
	}
	if len(orders) < 2 {
		t.Error("six RNG keys bottomed the rest in one order — it is not being shuffled")
	}
}

// TestWanderingArchaicBothFacesAreCatalogued pins the registry: the
// creature under the bare oracle ID, the sorcery under "#1", both full.
func TestWanderingArchaicBothFacesAreCatalogued(t *testing.T) {
	for _, tc := range []struct{ key, name string }{
		{wanderingArchaicOracle, "Wandering Archaic"},
		{wanderingArchaicOracle + "#1", "Explore the Vastlands"},
	} {
		spec, ok := Lookup(tc.key)
		if !ok || spec.Name != tc.name {
			t.Fatalf("%s: lookup = %v %q, want %q", tc.key, ok, spec.Name, tc.name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want full with none", tc.name, spec.Completeness, spec.Caveats)
		}
	}
}
