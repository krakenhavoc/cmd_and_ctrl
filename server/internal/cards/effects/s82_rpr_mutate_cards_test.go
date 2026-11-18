package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_mutate_cards_test.go — #5908: Illuna, Apex of Wishes and
// Nethroi, Apex of Death, the mutate cards skipped in RPR batches 17 and
// 18, on the ADR 0223 seam. Each mutate trigger is driven through the
// engine both over and under (CR 702.140d, 702.140e).

const (
	illunaOracle  = "08daa30f-7ce0-4a72-b539-fbb0c585fd73"
	nethroiOracle = "66bef5a7-79b7-4a9c-98dc-8670f867f04c"
)

func TestS82RprMutateCardsAreFullWithTheirMutateOffer(t *testing.T) {
	for _, tc := range []struct {
		oracle, name, cost string
		keywords           []string
	}{
		{illunaOracle, "Illuna, Apex of Wishes", "{3}{R/G}{U}{U}", []string{"flying", "trample"}},
		{nethroiOracle, "Nethroi, Apex of Death", "{4}{G/W}{B}{B}", []string{"deathtouch", "lifelink"}},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok || spec.Name != tc.name {
			t.Fatalf("%s is not registered: %+v", tc.name, spec)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 || len(spec.Triggered) != 1 {
			t.Errorf("%s: want Full, no caveat, one trigger; got %v %q %d", tc.name, spec.Completeness, spec.Caveats, len(spec.Triggered))
		}
		if len(spec.PrintedKeywords) != len(tc.keywords) || spec.PrintedKeywords[0] != tc.keywords[0] || spec.PrintedKeywords[1] != tc.keywords[1] {
			t.Errorf("%s: printed keywords %v, want %v", tc.name, spec.PrintedKeywords, tc.keywords)
		}
		alt := game.AlternativeCostByKey(tc.oracle, game.MutateKey)
		if alt == nil || alt.ManaCost != tc.cost || alt.Targets == nil || !alt.Targets.OwnedBySourceOwner {
			t.Errorf("%s: mutate offer %+v, want %s targeting a creature with the spell's owner", tc.name, alt, tc.cost)
		}
	}
}

// rprMutatorInHand puts a catalog mutate card in `owner`'s hand and
// walks to a main phase.
func rprMutatorInHand(t *testing.T, g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	g.PlayerByID(owner).Hand.PushTop(c)
	advanceToMainForTest(t, g)
	return c.InstanceID
}

func illunaCard() game.Card {
	return game.Card{
		Name: "Illuna, Apex of Wishes", TypeLine: "Legendary Creature — Beast Elemental Dinosaur",
		OracleID: illunaOracle, ManaCost: "{2}{G}{U}{R}", Colors: []string{"G", "U", "R"},
		Keywords: []string{"flying", "mutate", "trample"}, Power: 6, Toughness: 6,
	}
}

func nethroiCard() game.Card {
	return game.Card{
		Name: "Nethroi, Apex of Death", TypeLine: "Legendary Creature — Cat Nightmare Beast",
		OracleID: nethroiOracle, ManaCost: "{2}{W}{B}{G}", Colors: []string{"W", "B", "G"},
		Keywords: []string{"deathtouch", "lifelink", "mutate"}, Power: 5, Toughness: 5,
	}
}

// illunaLibrary stacks the top of `p`'s library: a land, an instant,
// then a creature card (the first nonland permanent card), then a card
// below it. Returns the IDs top first.
func illunaLibrary(p *game.Player) (land, instant, bears, below uuid.UUID) {
	below = pushLibraryTyped(p, "Below", "Creature — Bear")
	bears = uuid.New()
	p.Library.PushTop(game.Card{InstanceID: bears, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	instant = pushLibraryTyped(p, "Shock", "Instant")
	land = pushLibraryTyped(p, "Forest", "Basic Land — Forest")
	return
}

// Over: the Elves become Illuna (flying, trample, and the Elves' mana
// ability from under it), and the trigger exiles down to the first
// nonland permanent card — past a land and an instant — and puts it onto
// the battlefield.
func TestIllunaMutatesOverAndPutsTheCardOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := llanowarOn(g, me.ID)
	illuna := rprMutatorInHand(t, g, me.ID, illunaCard())
	land, instant, bears, below := illunaLibrary(me)
	if err := castMutating(g, me.ID, illuna, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)

	f := readMerged(t, g, elves)
	if f.name != "Illuna, Apex of Wishes" || f.power != 6 || !f.flying || !f.tapsForMana || len(f.components) != 2 {
		t.Fatalf("over: Illuna on top with the Elves' mana ability: %+v", f)
	}
	if c := findBattlefieldCardByID(g, elves); c == nil || !game.HasKeyword(c, "trample") {
		t.Errorf("over: Illuna's trample")
	}
	if !zoneHolds(g.Exile, land) || !zoneHolds(g.Exile, instant) || !zoneHolds(g.Exile, bears) {
		t.Fatalf("the run exiles the land, the instant and the creature: exile %+v", g.Exile.Cards)
	}
	if !zoneHolds(me.Library, below) {
		t.Errorf("the run stops at the first nonland permanent card")
	}
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatalf("no battlefield-or-hand question")
	}
	answerConfirmFor(t, g, me, true)
	var entered *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Grizzly Bears" {
			entered = &g.Battlefield.Cards[i]
		}
	}
	if entered == nil || entered.Controller != me.ID || zoneHolds(g.Exile, bears) || zoneHolds(me.Hand, bears) {
		t.Fatalf("Grizzly Bears enters under your control: %+v", entered)
	}
	if !zoneHolds(g.Exile, land) || !zoneHolds(g.Exile, instant) {
		t.Errorf("the other exiled cards stay in exile")
	}
}

// Under: the creature stays Llanowar Elves, has Illuna's flying and
// trample and its trigger, and the trigger fires; the card goes to hand.
func TestIllunaMutatesUnderAndPutsTheCardIntoYourHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := llanowarOn(g, me.ID)
	illuna := rprMutatorInHand(t, g, me.ID, illunaCard())
	land, instant, bears, _ := illunaLibrary(me)
	if err := castMutating(g, me.ID, illuna, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)

	f := readMerged(t, g, elves)
	if f.name != "Llanowar Elves" || f.power != 1 || !f.flying || len(f.components) != 2 || f.components[1] != illuna {
		t.Fatalf("under: the Elves on top with Illuna's abilities: %+v", f)
	}
	if !zoneHolds(g.Exile, bears) {
		t.Fatalf("the trigger of the card underneath fires and exiles down to the creature: %+v", g.Exile.Cards)
	}
	answerConfirmFor(t, g, me, false)
	if !zoneHolds(me.Hand, bears) || findBattlefieldCardByID(g, bears) != nil {
		t.Fatalf("Grizzly Bears goes into your hand")
	}
	if !zoneHolds(g.Exile, land) || !zoneHolds(g.Exile, instant) {
		t.Errorf("the other exiled cards stay in exile")
	}
}

// Cast for its mana cost, Illuna does not mutate and nothing triggers.
func TestIllunaCastNormallyDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	illuna := rprMutatorInHand(t, g, me.ID, illunaCard())
	land, _, _, _ := illunaLibrary(me)
	if err := g.CastSpell(me.ID, illuna, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, illuna) == nil {
		t.Fatalf("Illuna enters as a creature of its own")
	}
	if zoneHolds(g.Exile, land) || latestConfirmFor(g, me.ID) != nil {
		t.Errorf("nothing mutated, so nothing is exiled")
	}
}

// nethroiGrave puts a creature card of power `power` in p's graveyard.
func nethroiGrave(p *game.Player, name string, power int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Zombie", ManaCost: "{1}{B}",
		Power: power, Toughness: 1, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// castNethroiOnto mutates Nethroi onto `host`, over or under, and
// returns the trigger's target prompt.
func castNethroiOnto(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID, under bool) (uuid.UUID, *game.PendingChoice) {
	t.Helper()
	nethroi := rprMutatorInHand(t, g, me.ID, nethroiCard())
	if err := castMutating(g, me.ID, nethroi, host, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	passPriorityAroundTable(t, g)
	idx := 0
	if under {
		idx = 1
	}
	answerOptionPick(t, g, me.ID, idx)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no target prompt for Nethroi's mutate trigger")
	}
	return nethroi, pick
}

// Over: the Elves become Nethroi (deathtouch, lifelink, the Elves' mana
// ability), and the trigger returns creature cards of total power 10 or
// less; a set totalling 11 is refused, and an opponent's card is not a
// candidate.
func TestNethroiMutatesOverAndReturnsTotalPowerTenOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	six := nethroiGrave(me, "Six", 6)
	four := nethroiGrave(me, "Four", 4)
	five := nethroiGrave(me, "Five", 5)
	theirs := nethroiGrave(opp, "Theirs", 1)
	elves := llanowarOn(g, me.ID)
	_, pick := castNethroiOnto(t, g, me, elves, false)
	if pick.PickTargetMin != 0 {
		t.Errorf("any number: min %d", pick.PickTargetMin)
	}
	if hasID(pick.PickTargetCards, theirs) {
		t.Errorf("only your graveyard")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, cardRefs(six, five)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("total power 11: %v, want ErrIllegalTarget", err)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, cardRefs(six, four)); err != nil {
		t.Fatalf("total power 10: %v", err)
	}
	passPriorityAroundTable(t, g)

	f := readMerged(t, g, elves)
	if f.name != "Nethroi, Apex of Death" || f.power != 5 || !f.tapsForMana || len(f.components) != 2 {
		t.Fatalf("over: Nethroi on top with the Elves' mana ability: %+v", f)
	}
	if c := findBattlefieldCardByID(g, elves); c == nil || !game.HasKeyword(c, "deathtouch") || !game.HasKeyword(c, "lifelink") {
		t.Errorf("over: Nethroi's deathtouch and lifelink")
	}
	for _, id := range []uuid.UUID{six, four} {
		if c := findBattlefieldCardByID(g, id); c == nil || c.Controller != me.ID {
			t.Errorf("%v returns under your control", id)
		}
	}
	if !zoneHolds(me.Graveyard, five) {
		t.Errorf("an untargeted card stays")
	}
}

// Under: the creature stays Llanowar Elves, has deathtouch and lifelink
// from Nethroi beneath it, and the trigger of the card underneath fires.
// A target that left the graveyard in response is skipped (CR 608.2b).
func TestNethroiMutatesUnderAndItsTriggerStillFires(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	three := nethroiGrave(me, "Three", 3)
	two := nethroiGrave(me, "Two", 2)
	elves := llanowarOn(g, me.ID)
	nethroi, pick := castNethroiOnto(t, g, me, elves, true)
	if err := g.ResolvePickTargets(pick.ID, me.ID, cardRefs(three, two)); err != nil {
		t.Fatalf("total power 5: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(two); err != nil {
			t.Fatalf("exile a target: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	f := readMerged(t, g, elves)
	if f.name != "Llanowar Elves" || f.power != 1 || len(f.components) != 2 || f.components[1] != nethroi {
		t.Fatalf("under: the Elves on top: %+v", f)
	}
	if c := findBattlefieldCardByID(g, elves); c == nil || !game.HasKeyword(c, "deathtouch") || !game.HasKeyword(c, "lifelink") {
		t.Errorf("under: deathtouch and lifelink from Nethroi")
	}
	if findBattlefieldCardByID(g, three) == nil {
		t.Errorf("the still-legal target returns")
	}
	if !zoneHolds(g.Exile, two) {
		t.Errorf("the target exiled in response stays in exile")
	}
}
