package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ao_the_dawn_sky_test.go — the card #998 was filed for, and the proof
// that the one-field gap it named was the whole gap. The primitive's
// own tests are in library_pick_validate_test.go; the bot's are in
// legal/choose_cards_library_validate_test.go.

const aoTheDawnSkyOracle = "697ac261-263b-4879-bf45-6d10d23312ae"

func TestAoTheDawnSkyIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(aoTheDawnSkyOracle)
	if !ok {
		t.Fatal("Ao, the Dawn Sky is not registered")
	}
	if spec.Name != "Ao, the Dawn Sky" || spec.Completeness != CompletenessFull {
		t.Errorf("registered as %q / %v, want %q / full", spec.Name, spec.Completeness, "Ao, the Dawn Sky")
	}
}

// pushAoAndKillIt puts Ao onto the battlefield and destroys it, which
// is the whole of "when Ao dies". Returns the open mode_pick prompt.
func pushAoAndKillIt(t *testing.T, g *game.Game, owner uuid.UUID) *game.PendingChoice {
	t.Helper()
	ao := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ao, the Dawn Sky",
		TypeLine: "Legendary Creature — Dragon Spirit", ManaCost: "{3}{W}{W}",
		OracleID: aoTheDawnSkyOracle, Power: 5, Toughness: 4,
		Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(ao); err != nil {
			t.Fatalf("destroy Ao: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, owner)
	if c == nil {
		t.Fatal("the dies trigger asks which mode")
	}
	if len(c.ModeOptionIndex) != 2 || c.ModeMin != 1 || c.ModeMax != 1 {
		t.Fatalf("two bullets, choose one: %+v", c)
	}
	return c
}

// The card #998 was filed for: "put any number of nonland permanent
// cards with total mana value 4 or less from among them onto the
// battlefield". The over-budget set is refused and the legal one is
// performed, with the rest on the bottom.
func TestAoTheDawnSkyLibraryModeEnforcesTheTotalManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Seven cards, top first once looked at: five, land, four, three,
	// two, one, zero.
	zero := plTop(me, "Zero Drop", "Artifact", "")
	one := plTop(me, "One Drop", "Creature — Bear", "{1}")
	two := plTop(me, "Two Drop", "Creature — Bear", "{2}")
	three := plTop(me, "Three Drop", "Creature — Bear", "{3}")
	four := plTop(me, "Four Drop", "Creature — Bear", "{4}")
	land := plTop(me, "Wastes", "Basic Land — Wastes", "")
	five := plTop(me, "Five Drop", "Creature — Bear", "{5}")
	libBefore := me.Library.Size()

	mode := pushAoAndKillIt(t, g, me.ID)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)

	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the library mode asks which cards to put onto the battlefield")
	}
	// The land is not a candidate: "nonland permanent cards".
	for _, id := range pick.ChooseCards {
		if id == land {
			t.Error("a land was offered")
		}
	}
	if len(pick.ChooseCards) != 6 {
		t.Fatalf("offered %d candidates, want 6 (seven looked at, one land)", len(pick.ChooseCards))
	}

	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{three, two}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("3 + 2 = 5 is over budget: got %v, want ErrChoiceSetRejected", err)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{four, five}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("4 + 5 = 9 is over budget: got %v, want ErrChoiceSetRejected", err)
	}
	if still := latestChooseCardsFor(g, me.ID); still == nil || still.ID != pick.ID {
		t.Fatal("two refusals wedged the prompt")
	}

	// 0 + 1 + 3 = 4, on the nose.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{zero, one, three}); err != nil {
		t.Fatalf("the legal set: %v", err)
	}
	for _, id := range []uuid.UUID{zero, one, three} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not enter", id)
		}
	}
	for _, id := range []uuid.UUID{two, four, five, land} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s entered but was not chosen", id)
		}
	}
	// Seven were looked at, three entered, four went to the bottom.
	if got := me.Library.Size(); got != libBefore-3 {
		t.Errorf("library size %d, want %d", got, libBefore-3)
	}
	bottom := plBottomIDs(me, 4)
	for _, id := range []uuid.UUID{two, four, five, land} {
		if !bottom[id] {
			t.Errorf("%s is not among the bottom four", id)
		}
	}
}

// The other bullet, so the mode choice is a real one: two +1/+1
// counters on each creature and each Vehicle its controller controls,
// and nothing on an opponent's board or on a plain artifact.
func TestAoTheDawnSkyCounterModeCountsCreaturesAndVehicles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	truck := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Truck", TypeLine: "Artifact — Vehicle",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})

	mode := pushAoAndKillIt(t, g, me.ID)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)

	if n := counterCount(g, bear, "+1/+1"); n != 2 {
		t.Errorf("bear counters %d, want 2", n)
	}
	if n := counterCount(g, truck, "+1/+1"); n != 2 {
		t.Errorf("Vehicle counters %d, want 2", n)
	}
	if n := counterCount(g, rock, "+1/+1"); n != 0 {
		t.Errorf("a plain artifact got %d counters, want 0", n)
	}
	if n := counterCount(g, theirs, "+1/+1"); n != 0 {
		t.Errorf("an opponent's creature got %d counters, want 0", n)
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("the counter mode asks nothing")
	}
}
