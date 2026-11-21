package testsex

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestExploreTheVastlandsWholeCard is the card at a four-player table:
// two predicates picked, one picked, none picked, and a player with
// nothing to pick — then the life.
func TestExploreTheVastlandsWholeCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seats := effectstest.ApnapSeats(g)
	tops := make([][]uuid.UUID, len(seats))
	for i, p := range seats[:3] {
		tops[i] = effectstest.StackTop(p, effectstest.VastFive)
	}
	// The fourth player's five has nothing either slot takes.
	tops[3] = effectstest.StackTop(seats[3], []effectstest.VastCard{
		{Name: "Bear A", TypeLine: "Creature — Bear"}, {Name: "Bear B", TypeLine: "Creature — Bear"}, {Name: "Sol Ring", TypeLine: "Artifact"},
		{Name: "Bear C", TypeLine: "Creature — Bear"}, {Name: "Bear D", TypeLine: "Creature — Bear"},
	})
	life := make([]int, len(seats))
	hand := make([]int, len(seats))
	for i, p := range seats {
		life[i], hand[i] = p.Life, p.Hand.Size()
	}

	// castExplore deals the card into the hand it is cast from, so the
	// caster's baseline above does not include it.
	spell := effectstest.CastExplore(t, g)

	// Every player with a candidate is asked, about their own five,
	// up to one land and one instant or sorcery: two cards at most.
	for i, p := range seats[:3] {
		c := effectstest.LatestChooseCards(g, p.ID)
		if c == nil {
			t.Fatalf("seat %d was not asked", i)
		}
		want := []uuid.UUID{tops[i][0], tops[i][1], tops[i][3]}
		if !effectstest.SameSet(c.ChooseCards, want) {
			t.Errorf("seat %d offered %v, want its land, instant and sorcery", i, c.ChooseCards)
		}
		if c.ChooseMin != 0 || c.ChooseMax != 2 {
			t.Errorf("seat %d bounds [%d,%d], want [0,2]", i, c.ChooseMin, c.ChooseMax)
		}
	}
	if effectstest.LatestChooseCards(g, seats[3].ID) != nil {
		t.Error("a player with no land, instant or sorcery in their five was asked")
	}

	// Answered out of APNAP order. Nothing moves, nothing is revealed
	// and nobody gains life until the last answer is in.
	answer := func(i int, ids ...uuid.UUID) {
		t.Helper()
		c := effectstest.LatestChooseCards(g, seats[i].ID)
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
	if seats[0].Hand.Contains(tops[0][0]) || effectstest.VastRevealed(g, tops[0][0]) {
		t.Error("an answer was acted on before every player had answered")
	}
	answer(1, tops[1][3])

	taken := [][]uuid.UUID{{tops[0][0], tops[0][1]}, {tops[1][3]}, nil, nil}
	for i, p := range seats {
		for _, id := range taken[i] {
			if !p.Hand.Contains(id) {
				t.Errorf("seat %d: a revealed card is not in hand", i)
			}
			if !effectstest.VastRevealed(g, id) {
				t.Errorf("seat %d: a card taken into hand was not revealed", i)
			}
		}
		if got := p.Hand.Size(); got != hand[i]+len(taken[i]) {
			t.Errorf("seat %d hand %d, want %d", i, got, hand[i]+len(taken[i]))
		}
		rest := effectstest.Without(tops[i], taken[i]...)
		for _, id := range rest {
			if effectstest.VastRevealed(g, id) {
				t.Errorf("seat %d: a card that was only looked at was revealed", i)
			}
		}
		if !effectstest.SameSet(effectstest.LibraryBottom(p, len(rest)), rest) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := effectstest.StackTop(me, []effectstest.VastCard{
		{Name: "Forest", TypeLine: "Basic Land — Forest"}, {Name: "Island", TypeLine: "Basic Land — Island"}, {Name: "Opt", TypeLine: "Instant"},
		{Name: "Grizzly Bears", TypeLine: "Creature — Bear"}, {Name: "Preordain", TypeLine: "Sorcery"},
	})
	forest, island, opt, bears, preordain := top[0], top[1], top[2], top[3], top[4]
	effectstest.CastExplore(t, g)
	c := effectstest.LatestChooseCards(g, me.ID)
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
		if effectstest.LatestChooseCards(g, me.ID) == nil {
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
	take := func(t *testing.T, cards []effectstest.VastCard) (*game.Game, *game.Player, []uuid.UUID, *game.PendingChoice) {
		t.Helper()
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		top := effectstest.StackTop(me, cards)
		g.WithWriteLock(func() {
			ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
			err := effects.TakeFromLibraryToHand{
				Player:   me.ID,
				Cards:    g.LookAtTopOfLibraryForEffect(me.ID, len(cards)),
				Optional: true,
				Label:    "test — you may reveal a creature card and/or an artifact card",
				Key:      effectstest.TestSlotsTake,
			}.Apply(ctx)
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}
		})
		return g, me, top, effectstest.LatestChooseCards(g, me.ID)
	}

	// Alone, the artifact creature is one card in one slot: the
	// ceiling is one, not two.
	_, _, _, c := take(t, []effectstest.VastCard{{Name: "Myr", TypeLine: "Artifact Creature — Myr"}, {Name: "Forest", TypeLine: "Basic Land — Forest"}})
	if c == nil || c.ChooseMax != 1 {
		t.Fatalf("one artifact creature: prompt %v, want a ceiling of 1", c)
	}

	// With two plain creatures beside it: two creatures cannot both be
	// taken (one creature slot), but the artifact creature plus either
	// one can — it moves to the artifact slot.
	g, me, top, c := take(t, []effectstest.VastCard{
		{Name: "Myr", TypeLine: "Artifact Creature — Myr"}, {Name: "Bear A", TypeLine: "Creature — Bear"}, {Name: "Bear B", TypeLine: "Creature — Bear"},
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
	if !effectstest.SameSet(effectstest.LibraryBottom(me, 1), []uuid.UUID{bearB}) {
		t.Error("the untaken creature is not on the bottom")
	}
}

// TestExploreTheVastlandsShortAndEmptyLibraries: a short library is
// looked at as far as it goes, an empty one is a look at nothing, and
// neither is an error.
func TestExploreTheVastlandsShortAndEmptyLibraries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seats := effectstest.ApnapSeats(g)
	g.WithWriteLock(func() {
		seats[1].Library.Cards = nil
		seats[2].Library.Cards = nil
	})
	short := effectstest.StackTop(seats[1], []effectstest.VastCard{{Name: "Forest", TypeLine: "Basic Land — Forest"}, {Name: "Grizzly Bears", TypeLine: "Creature — Bear"}})
	for _, p := range []*game.Player{seats[0], seats[3]} {
		g.WithWriteLock(func() { p.Library.Cards = nil })
		effectstest.StackTop(p, []effectstest.VastCard{{Name: "Grizzly Bears", TypeLine: "Creature — Bear"}})
	}
	lifeBefore := seats[2].Life

	effectstest.CastExplore(t, g)

	c := effectstest.LatestChooseCards(g, seats[1].ID)
	if c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != short[0] || c.ChooseMax != 1 {
		t.Fatalf("two-card library: prompt %v, want the Forest alone, up to one", c)
	}
	for _, p := range []*game.Player{seats[0], seats[2], seats[3]} {
		if effectstest.LatestChooseCards(g, p.ID) != nil {
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
	g := effectstest.NewCatalogGame(t)
	seats := effectstest.ApnapSeats(g)
	tops := make([][]uuid.UUID, len(seats))
	for i, p := range seats {
		tops[i] = effectstest.StackTop(p, effectstest.VastFive)
	}
	effectstest.CastExplore(t, g)

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
		c := effectstest.LatestChooseCards(g, p.ID)
		if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{tops[i][0]}); err != nil {
			t.Fatalf("seat %d: %v", i, err)
		}
	}
	for i, p := range seats {
		if !effectstest.VastRevealed(g, tops[i][0]) {
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
		g := effectstest.NewCatalogGame(t)
		var key [32]byte
		key[0] = byte(k)
		g.SetRNGKeyForTest(key)
		seats := effectstest.ApnapSeats(g)
		me := seats[0]
		top := effectstest.StackTop(me, effectstest.VastFive)
		for _, p := range seats[1:] {
			g.WithWriteLock(func() { p.Library.Cards = nil })
		}
		effectstest.CastExplore(t, g)
		c := effectstest.LatestChooseCards(g, me.ID)
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{top[0]}); err != nil {
			t.Fatalf("key %d: %v", k, err)
		}
		rest := effectstest.Without(top, top[0])
		bottom := effectstest.LibraryBottom(me, len(rest))
		if !effectstest.SameSet(bottom, rest) {
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
		{effectstest.WanderingArchaicOracle, "Wandering Archaic"},
		{effectstest.WanderingArchaicOracle + "#1", "Explore the Vastlands"},
	} {
		spec, ok := effects.Lookup(tc.key)
		if !ok || spec.Name != tc.name {
			t.Fatalf("%s: lookup = %v %q, want %q", tc.key, ok, spec.Name, tc.name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want full with none", tc.name, spec.Completeness, spec.Caveats)
		}
	}
}
