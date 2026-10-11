package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// charge_storage_lands_test.go — #2951: Black Mana Battery, Bottomless
// Vault and Blast Zone through the real catalog. The engine halves
// (variable counter costs, X costs, the untap opt-out) are pinned in
// game/; this file pins what each card does with them.

const (
	blackManaBatteryOracle = "dc3b2bab-755d-4b2f-97c0-70a74d721a79"
	bottomlessVaultOracle  = "e43413e4-be17-49af-978a-26210d05f52a"
	blastZoneOracle        = "393a254f-be31-431a-9341-a51286f8cbce"
)

func poolColours(p *game.Player) map[string]int {
	out := map[string]int{}
	for _, tok := range p.ManaPool {
		out[tok.Color]++
	}
	return out
}

func addCountersForTest(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, kind, n) })
}

// --- Black Mana Battery -------------------------------------------

func TestBlackManaBatteryBanksAChargeCounterForTwoAndTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bat := pushCatalogPermanent(g, me.ID, "Black Mana Battery", "Artifact", blackManaBatteryOracle, false)
	advanceToMain(t, g)
	fillPool(me, 2)

	if err := g.ActivateCatalogAbility(me.ID, bat, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("bank: %v", err)
	}
	if got := len(me.ManaPool); got != 0 {
		t.Errorf("pool after {2}: %d, want 0", got)
	}
	if got := counterCount(g, bat, game.CounterCharge); got != 0 {
		t.Errorf("the counter arrives on resolution, found %d on the stack", got)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bat, game.CounterCharge); got != 1 {
		t.Errorf("charge counters %d, want 1", got)
	}
	if !tappedForTest(t, g, bat) {
		t.Error("banking taps the Battery")
	}
}

func TestBlackManaBatteryPaysOneBlackPlusOnePerCounterRemoved(t *testing.T) {
	for _, tc := range []struct {
		name         string
		held, remove int
		wantBlack    int
		wantLeft     int
	}{
		{"none removed", 4, 0, 1, 4},
		{"partial", 4, 3, 4, 1},
		{"all", 4, 4, 5, 0},
		{"none held", 0, 0, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bat := b12Push(g, me.ID, "Black Mana Battery", "Artifact", blackManaBatteryOracle, 0, 0)
			advanceToMain(t, g)
			if tc.held > 0 {
				addCountersForTest(g, bat, game.CounterCharge, tc.held)
			}
			activateManaFor(t, g, me.ID, bat, 0, game.ManaAbilityParams{CounterCounts: []int{tc.remove}})
			if got := poolColours(me); got["B"] != tc.wantBlack || len(me.ManaPool) != tc.wantBlack {
				t.Errorf("pool %v, want %d {B} and nothing else", got, tc.wantBlack)
			}
			if got := counterCount(g, bat, game.CounterCharge); got != tc.wantLeft {
				t.Errorf("charge counters left %d, want %d", got, tc.wantLeft)
			}
			if !tappedForTest(t, g, bat) {
				t.Error("the payout taps the Battery")
			}
		})
	}
}

func TestBlackManaBatteryRefusesToRemoveMoreThanItHolds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bat := b12Push(g, me.ID, "Black Mana Battery", "Artifact", blackManaBatteryOracle, 0, 0)
	advanceToMain(t, g)
	addCountersForTest(g, bat, game.CounterCharge, 2)
	if err := g.ActivateManaAbility(me.ID, bat, 0, game.ManaAbilityParams{CounterCounts: []int{3}}); err == nil {
		t.Fatal("removing three of two counters was allowed")
	}
	if len(me.ManaPool) != 0 || counterCount(g, bat, game.CounterCharge) != 2 || tappedForTest(t, g, bat) {
		t.Error("a refused payout changed the board")
	}
}

// --- Bottomless Vault ---------------------------------------------

func TestBottomlessVaultEntersTappedWithNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	land := b12PlayFromHand(t, g, "Bottomless Vault", "Land", bottomlessVaultOracle, game.CastSpellParams{})
	if !b12Card(t, g, land).Tapped {
		t.Error("Bottomless Vault did not enter tapped")
	}
	if got := counterCount(g, land, game.CounterStorage); got != 0 {
		t.Errorf("storage counters %d, want 0", got)
	}
}

func TestBottomlessVaultPaysOneBlackPerCounterActuallyRemoved(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := b12Push(g, me.ID, "Bottomless Vault", "Land", bottomlessVaultOracle, 0, 0)
	advanceToMain(t, g)
	addCountersForTest(g, land, game.CounterStorage, 4)

	// Zero: "any number" includes none. The land taps for nothing.
	activateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{CounterCounts: []int{0}})
	if len(me.ManaPool) != 0 || counterCount(g, land, game.CounterStorage) != 4 || !tappedForTest(t, g, land) {
		t.Fatalf("zero removed: pool %v, counters %d, tapped %v",
			poolColours(me), counterCount(g, land, game.CounterStorage), tappedForTest(t, g, land))
	}

	// Partial.
	untapForTest(g, land)
	activateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{CounterCounts: []int{3}})
	if got := poolColours(me); got["B"] != 3 || len(me.ManaPool) != 3 {
		t.Fatalf("three removed: pool %v, want {B}x3", got)
	}
	if got := counterCount(g, land, game.CounterStorage); got != 1 {
		t.Errorf("counters left %d, want 1", got)
	}

	// All of what is left.
	untapForTest(g, land)
	activateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{CounterCounts: []int{1}})
	if got := poolColours(me); got["B"] != 4 || len(me.ManaPool) != 4 {
		t.Errorf("all removed: pool %v, want {B}x4 in all", got)
	}
	if got := counterCount(g, land, game.CounterStorage); got != 0 {
		t.Errorf("counters left %d, want 0", got)
	}

	// More than it holds is refused.
	untapForTest(g, land)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{CounterCounts: []int{1}}); err == nil {
		t.Error("removing a counter it does not hold was allowed")
	}
}

// Declining to untap keeps it tapped, and the upkeep trigger banks a
// counter; accepting the untap leaves the trigger nothing to count.
func TestBottomlessVaultUntapOptOutAndUpkeepCounter(t *testing.T) {
	t.Run("declined untap banks a counter", func(t *testing.T) {
		g := newCatalogGame(t)
		owner := g.Seats[1]
		land := pushTappedForTest(g, owner.ID, "Bottomless Vault", bottomlessVaultOracle, "Land")

		choice := advanceToUntapChoiceOf(t, g, 1)
		if choice.ChooseMin != 0 {
			t.Fatalf("opt-out prompt floor %d, want 0", choice.ChooseMin)
		}
		if err := g.ResolveUntapChoice(choice.ID, owner.ID, nil); err != nil {
			t.Fatalf("decline: %v", err)
		}
		advanceToUpkeepOf(t, g, 1)
		passPriorityAroundTable(t, g)
		if !tappedForTest(t, g, land) {
			t.Error("the Vault untapped though its controller declined")
		}
		if got := counterCount(g, land, game.CounterStorage); got != 1 {
			t.Errorf("storage counters %d, want 1", got)
		}
	})

	t.Run("accepted untap banks nothing", func(t *testing.T) {
		g := newCatalogGame(t)
		owner := g.Seats[1]
		land := pushTappedForTest(g, owner.ID, "Bottomless Vault", bottomlessVaultOracle, "Land")

		choice := advanceToUntapChoiceOf(t, g, 1)
		if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{land}); err != nil {
			t.Fatalf("accept: %v", err)
		}
		advanceToUpkeepOf(t, g, 1)
		passPriorityAroundTable(t, g)
		if got := counterCount(g, land, game.CounterStorage); got != 0 {
			t.Errorf("an untapped Vault banked %d counters at upkeep", got)
		}
	})

	t.Run("an opponent's upkeep banks nothing", func(t *testing.T) {
		g := newCatalogGame(t)
		land := pushTappedForTest(g, g.Seats[0].ID, "Bottomless Vault", bottomlessVaultOracle, "Land")
		advanceToUpkeepOf(t, g, 1)
		passPriorityAroundTable(t, g)
		if got := counterCount(g, land, game.CounterStorage); got != 0 {
			t.Errorf("banked %d counters on another player's upkeep", got)
		}
	})
}

// The intervening "if" is read again on resolution (CR 603.4). The
// engine settles this upkeep trigger before a test can act in
// response, so the resolution half is driven through the registered
// row: an untapped Vault puts nothing on itself, a tapped one does.
func TestBottomlessVaultInterveningIfIsRecheckedOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Bottomless Vault", "Land", bottomlessVaultOracle, false)
	spec, ok := Lookup(bottomlessVaultOracle)
	if !ok || len(spec.Triggered) != 1 || spec.Triggered[0].Effect == nil {
		t.Fatal("Bottomless Vault does not declare its upkeep trigger")
	}
	resolve := func() {
		item := &game.StackItem{SourceCardID: land, Controller: me.ID}
		g.WithWriteLock(func() {
			if err := spec.Triggered[0].Effect(g, item); err != nil {
				t.Fatalf("resolve: %v", err)
			}
		})
	}

	resolve()
	if got := counterCount(g, land, game.CounterStorage); got != 0 {
		t.Fatalf("an untapped Vault banked %d counters on resolution", got)
	}
	b16Tap(g, land)
	resolve()
	if got := counterCount(g, land, game.CounterStorage); got != 1 {
		t.Errorf("a tapped Vault banked %d counters on resolution, want 1", got)
	}
}

// --- Blast Zone ---------------------------------------------------

func TestBlastZoneEntersWithOneChargeCounterAndTapsForColourless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := b12PlayFromHand(t, g, "Blast Zone", "Land", blastZoneOracle, game.CastSpellParams{})
	if b12Card(t, g, land).Tapped {
		t.Error("Blast Zone enters untapped")
	}
	if got := counterCount(g, land, game.CounterCharge); got != 1 {
		t.Fatalf("charge counters %d, want 1", got)
	}
	activateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
	if got := poolColours(me); got["C"] != 1 || len(me.ManaPool) != 1 {
		t.Errorf("pool %v, want one {C}", got)
	}
}

func TestBlastZoneChargesBothXSlotsAndBanksXCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zone := b12PlayFromHand(t, g, "Blast Zone", "Land", blastZoneOracle, game.CastSpellParams{})
	fillPool(me, 7) // {X}{X} at X=3 is six, so exactly one stays behind

	if err := g.ActivateCatalogAbility(me.ID, zone, 0, game.ActivateAbilityParams{XValue: 3, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("pool after {X}{X} at X=3: %d, want 1", got)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, zone, game.CounterCharge); got != 4 {
		t.Errorf("charge counters %d, want 1 + 3", got)
	}
}

func TestBlastZoneRefusesAnXItCannotPayAndAllowsXZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zone := b12PlayFromHand(t, g, "Blast Zone", "Land", blastZoneOracle, game.CastSpellParams{})
	fillPool(me, 5)
	if err := g.ActivateCatalogAbility(me.ID, zone, 0, game.ActivateAbilityParams{XValue: 3, Strict: true}); err == nil {
		t.Fatal("X=3 on five mana was allowed")
	}
	if tappedForTest(t, g, zone) || len(me.ManaPool) != 5 {
		t.Error("a refused activation changed the board")
	}
	if err := g.ActivateCatalogAbility(me.ID, zone, 0, game.ActivateAbilityParams{XValue: 0, Strict: true}); err != nil {
		t.Fatalf("X=0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, zone, game.CounterCharge); got != 1 {
		t.Errorf("charge counters %d after X=0, want the 1 it entered with", got)
	}
}

// blastZoneBoard is Blast Zone and a spread of other permanents.
type blastZoneBoard struct {
	zone, one, two, twoOpp, twoIndestructible, token, zero, landTwo uuid.UUID
}

func setUpBlastZone(g *game.Game, counters int) blastZoneBoard {
	me, opp := g.Seats[0], g.Seats[1]
	b := blastZoneBoard{}
	b.zone = pushCatalogPermanent(g, me.ID, "Blast Zone", "Land", blastZoneOracle, false)
	if counters > 0 {
		addCountersForTest(g, b.zone, game.CounterCharge, counters)
	}
	push := func(owner uuid.UUID, c game.Card) uuid.UUID {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), owner, owner
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	b.one = push(me.ID, game.Card{Name: "One Drop", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1})
	b.two = push(me.ID, game.Card{Name: "Two Drop", TypeLine: "Artifact", ManaCost: "{2}"})
	b.twoOpp = push(opp.ID, game.Card{Name: "Opposing Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	b.twoIndestructible = push(opp.ID, game.Card{Name: "Sturdy Two Drop", TypeLine: "Artifact Creature — Golem", ManaCost: "{2}", Power: 2, Toughness: 2, Keywords: []string{"indestructible"}})
	b.token = push(opp.ID, game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1})
	b.zero = push(me.ID, game.Card{Name: "Zero Drop", TypeLine: "Artifact", ManaCost: "{0}"})
	b.landTwo = push(opp.ID, game.Card{Name: "Odd Land", TypeLine: "Land", ManaCost: "{2}"})
	return b
}

func TestBlastZoneDestroysNonlandsWithManaValueEqualToItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b := setUpBlastZone(g, 2)
	advanceToMain(t, g)
	fillPool(me, 3)

	if err := g.ActivateCatalogAbility(me.ID, b.zone, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The sacrifice is a cost: the land is gone and its graveyard card
	// holds no counters, so only last-known information can say "2".
	if g.Battlefield.Contains(b.zone) {
		t.Fatal("the sacrifice cost is paid at announce")
	}
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"your mana value 2": b.two, "opposing mana value 2": b.twoOpp} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{
		"mana value 1":         b.one,
		"mana value 0":         b.zero,
		"indestructible":       b.twoIndestructible,
		"token (mana value 0)": b.token,
		"a land of mana value": b.landTwo,
	} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
}

// Zero counters destroys every nonland permanent with mana value 0,
// tokens included, and still spares lands and everything costed.
func TestBlastZoneWithNoCountersDestroysTokensAndZeroValuePermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b := setUpBlastZone(g, 0)
	advanceToMain(t, g)
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, b.zone, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(b.token) || g.Battlefield.Contains(b.zero) {
		t.Error("a mana value 0 nonland permanent survived")
	}
	for name, id := range map[string]uuid.UUID{"mana value 1": b.one, "mana value 2": b.two, "land": b.landTwo} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed at zero counters", name)
		}
	}
}

// Banked counters count: one from entry, one banked through the real
// X ability, one added by hand, for three.
func TestBlastZoneBankedCountersSetTheSweep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zone := b12PlayFromHand(t, g, "Blast Zone", "Land", blastZoneOracle, game.CastSpellParams{})
	three := b12Push(g, me.ID, "Three Drop", "Artifact", "", 0, 0)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == three {
			g.Battlefield.Cards[i].ManaCost = "{3}"
		}
	}
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, zone, 0, game.ActivateAbilityParams{XValue: 1, Strict: true}); err != nil {
		t.Fatalf("bank: %v", err)
	}
	passPriorityAroundTable(t, g)
	addCountersForTest(g, zone, game.CounterCharge, 1)
	untapForTest(g, zone)
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, zone, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(three) {
		t.Error("the mana value 3 artifact survived a three-counter Blast Zone")
	}
}
