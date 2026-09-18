package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// superfriends_test.go — S27: the two planeswalkers that prove the
// loyalty catalog hook end to end, and the two counter-payoff bridges
// that make the counters they place worth placing.

const (
	elspethSunsChampionOracle = "05e6b243-48a6-4a42-bc5f-413441de9c33"
	wrennAndSixOracle         = "108ae90a-50fa-4cfd-b751-d630e41425fe"
	vorinclexOracle           = "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b"
	pirOracle                 = "7683c2b2-a06f-4691-9cc5-1968dc032885"
)

// pushWalkerForTest puts a planeswalker on the battlefield with its
// starting loyalty already stamped, past summoning sickness.
func pushWalkerForTest(g *game.Game, owner uuid.UUID, name, oracleID string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		OracleID:   oracleID,
		TypeLine:   "Legendary Planeswalker — Test",
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

func loyaltyOf(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[game.CounterLoyalty]
		}
	}
	return -1
}

func countNamed(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// advanceToMainOf walks to the named seat's precombat main phase,
// which is where a sorcery-speed loyalty ability can be activated.
func advanceToMainOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.Step == game.StepPrecombatMain && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached a precombat main phase")
}

func TestElspethPlusOneMakesThreeSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	pw := pushWalkerForTest(g, owner.ID, "Elspeth, Sun's Champion", elspethSunsChampionOracle, 4)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	// The loyalty is paid at ANNOUNCE (CR 606.4), before the ability
	// reaches the stack.
	if got := loyaltyOf(g, pw); got != 5 {
		t.Errorf("loyalty after +1 = %d, want 5", got)
	}
	passPriorityAroundTable(t, g)
	if got := countNamed(g, "Soldier"); got != 3 {
		t.Errorf("Soldiers = %d, want 3", got)
	}
}

func TestElspethMinusThreeSweepsBigCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	pw := pushWalkerForTest(g, owner.ID, "Elspeth, Sun's Champion", elspethSunsChampionOracle, 4)
	small := pushCreatureToBattlefieldForTest(g, owner.ID, "Small")
	big := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: big,
		Name:       "Big",
		TypeLine:   "Creature — Test",
		Power:      5,
		Toughness:  5,
		Owner:      g.Seats[(seat+1)%len(g.Seats)].ID,
		Controller: g.Seats[(seat+1)%len(g.Seats)].ID,
	})
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	if got := loyaltyOf(g, pw); got != 1 {
		t.Errorf("loyalty after −3 = %d, want 1", got)
	}
	passPriorityAroundTable(t, g)

	if countNamed(g, "Big") != 0 {
		t.Error("a 5/5 survived 'destroy all creatures with power 4 or greater'")
	}
	if countNamed(g, "Small") != 1 {
		t.Error("a 2/2 was destroyed by 'power 4 or greater'")
	}
	// It is a symmetric sweep: nothing about it spares your own.
	_ = small
}

// TestElspethsUltimateIsWholeAndDeclaresItself replaces
// TestElspethDeclaresItsOmittedUltimate, which pinned the −7 as
// OMITTED from #621 until emblems landed (#623). Every assertion in
// it has inverted: the card now declares three abilities, the third
// is the −7, the emblem it makes is registered with both halves of
// its printed text, and the catalog page publishes the card as whole
// rather than caveated.
//
// The Caveats check is the half that would have failed silently:
// effects.Register panics on Caveats without CompletenessCaveats, so
// a stale caveat could not survive — but a caveat that became untrue
// while the declaration stayed at CompletenessCaveats would.
func TestElspethsUltimateIsWholeAndDeclaresItself(t *testing.T) {
	spec, ok := Lookup(elspethSunsChampionOracle)
	if !ok {
		t.Fatal("Elspeth, Sun's Champion is not registered")
	}
	if len(spec.Activated) != 3 {
		t.Fatalf("Elspeth declares %d abilities, want 3 (+1, −3 and −7)", len(spec.Activated))
	}
	var ultimate bool
	for i, ab := range spec.Activated {
		if ab.Cost.Loyalty == nil {
			t.Errorf("ability %d (%q) has no loyalty cost", i, ab.Label)
			continue
		}
		if *ab.Cost.Loyalty == -7 {
			ultimate = true
		}
	}
	if !ultimate {
		t.Error("no −7 among Elspeth's abilities")
	}
	if spec.Emblem == nil {
		t.Fatal("Elspeth declares no Emblem — the −7 has nothing to create")
	}
	if spec.Emblem.Label == "" || spec.Emblem.Text == "" {
		t.Errorf("emblem is missing its label or text: %+v", *spec.Emblem)
	}
	// Two statics, because CR 613 applies the ability grant (layer 6)
	// before the P/T modification (layer 7c) and one StaticAbility
	// declares one layer.
	if len(spec.Emblem.Static) != 2 {
		t.Errorf("emblem declares %d statics, want 2 (+2/+2 and flying)", len(spec.Emblem.Static))
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("Completeness = %v, want CompletenessFull", spec.Completeness)
	}
	if len(spec.Caveats) != 0 {
		t.Errorf("Elspeth still lists caveats: %v", spec.Caveats)
	}
}

// TestEveryPlaneswalkerDeclaresItsCompleteness is the catalog-wide
// guard. Omitting an ultimate is still an ordinary shape for a
// planeswalker here (ADR 0032) — Wrenn and Six's −7 waits on retrace
// and #652 even now that emblems exist — so a spec with a loyalty
// ability left at CompletenessUnreviewed is almost certainly an
// omission the catalog page is not telling anyone about.
func TestEveryPlaneswalkerDeclaresItsCompleteness(t *testing.T) {
	for _, spec := range All() {
		for _, ab := range spec.Activated {
			if ab.Cost.Loyalty == nil {
				continue
			}
			if spec.Completeness == CompletenessUnreviewed {
				t.Errorf("%s has loyalty abilities but no Completeness declaration", spec.Name)
			}
			break
		}
	}
}

// TestLoyaltyAbilityIsOncePerTurn is CR 606.3, enforced by the engine
// off the presence of the loyalty component rather than by anything
// the card declares.
func TestLoyaltyAbilityIsOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	pw := pushWalkerForTest(g, owner.ID, "Elspeth, Sun's Champion", elspethSunsChampionOracle, 4)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first +1: %v", err)
	}
	if err := g.ActivateCatalogAbility(owner.ID, pw, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a second loyalty ability was activated in the same turn")
	}
}

// TestWrennMinusOneDealsDamage covers the ability end to end: the
// loyalty payment at announce, the target clause, and the damage.
//
// The target here is a CREATURE. Wrenn's −1 aimed at a PLANESWALKER
// is the case issue #406 makes real — until that fix lands, damage to
// a walker increments a number the lethal-damage SBA never reads —
// and it is tested where the fix lives rather than here, so this
// branch asserts only what this branch can deliver.
func TestWrennMinusOneDealsDamage(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	victimSeat := g.Seats[(seat+1)%len(g.Seats)]
	wrenn := pushWalkerForTest(g, owner.ID, "Wrenn and Six", wrennAndSixOracle, 3)
	victim := pushCreatureToBattlefieldForTest(g, victimSeat.ID, "Victim")
	advanceToMainOf(t, g, seat)

	// Ability index 1 is the −1.
	err := g.ActivateCatalogAbility(owner.ID, wrenn, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if err != nil {
		t.Fatalf("−1: %v", err)
	}
	if got := loyaltyOf(g, wrenn); got != 2 {
		t.Errorf("loyalty after −1 = %d, want 2", got)
	}
	passPriorityAroundTable(t, g)

	marked := -1
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == victim {
			marked = c.DamageMarked
		}
	}
	if marked != 1 {
		t.Errorf("damage marked on the target = %d, want 1", marked)
	}
}

// TestWrennPlusOneIsActivatableWithAnEmptyGraveyard — "up to one
// target" is Min 0, and the loyalty gain is most of why the card is
// played.
func TestWrennPlusOneIsActivatableWithAnEmptyGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	wrenn := pushWalkerForTest(g, owner.ID, "Wrenn and Six", wrennAndSixOracle, 3)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, wrenn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1 with an empty graveyard: %v", err)
	}
	if got := loyaltyOf(g, wrenn); got != 4 {
		t.Errorf("loyalty after +1 = %d, want 4", got)
	}
}

// --- counter-payoff bridges -------------------------------------

func TestVorinclexDoublesYourCountersAndHalvesTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	them := g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Vorinclex, Monstrous Raider",
		OracleID:   vorinclexOracle,
		TypeLine:   "Legendary Creature — Phyrexian Praetor",
		Power:      6,
		Toughness:  6,
		Owner:      me.ID,
		Controller: me.ID,
	})
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	theirs := pushCreatureToBattlefieldForTest(g, them.ID, "Theirs")

	if err := g.AddCounter(mine, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounter mine: %v", err)
	}
	if err := g.AddCounter(theirs, game.CounterPlusOne, 3); err != nil {
		t.Fatalf("AddCounter theirs: %v", err)
	}

	if got := counterCount(g, mine, game.CounterPlusOne); got != 4 {
		t.Errorf("your counters = %d, want 4 (doubled)", got)
	}
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("their counters = %d, want 1 (halved, rounded down)", got)
	}
}

// TestVorinclexDoublesLoyaltyToo — "counters", uncategorised. The
// loyalty stamp on a planeswalker entering is a counter placement
// like any other.
func TestVorinclexDoublesLoyaltyToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Vorinclex, Monstrous Raider",
		OracleID:   vorinclexOracle,
		TypeLine:   "Legendary Creature — Phyrexian Praetor",
		Power:      6,
		Toughness:  6,
		Owner:      me.ID,
		Controller: me.ID,
	})
	pw := pushWalkerForTest(g, me.ID, "Some Walker", "", 0)

	if err := g.AddCounter(pw, game.CounterLoyalty, 4); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := loyaltyOf(g, pw); got != 8 {
		t.Errorf("loyalty = %d, want 8 (doubled)", got)
	}
}

// TestVorinclexLeavesRemovalsAlone — halving a REMOVAL would round
// the wrong way in Go (integer division truncates toward zero), and
// CR 614.1 replacements do not apply to counter removal in any case.
func TestVorinclexLeavesRemovalsAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Vorinclex, Monstrous Raider",
		OracleID:   vorinclexOracle,
		TypeLine:   "Legendary Creature — Phyrexian Praetor",
		Power:      6,
		Toughness:  6,
		Owner:      me.ID,
		Controller: me.ID,
	})
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	if err := g.AddCounter(mine, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	// 2 doubled to 4, then remove 1 — which must stay 1, not become 2.
	if err := g.AddCounter(mine, game.CounterPlusOne, -1); err != nil {
		t.Fatalf("remove counter: %v", err)
	}
	if got := counterCount(g, mine, game.CounterPlusOne); got != 3 {
		t.Errorf("counters after removing one = %d, want 3", got)
	}
}

func TestPirAddsOneToAnyPermanentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	them := g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Pir, Imaginative Rascal",
		OracleID:   pirOracle,
		TypeLine:   "Legendary Creature — Human",
		Power:      1,
		Toughness:  1,
		Owner:      me.ID,
		Controller: me.ID,
	})
	// A PLANESWALKER, not a creature: that is the difference from
	// Hardened Scales, which only ever reads creatures.
	pw := pushWalkerForTest(g, me.ID, "Some Walker", "", 0)
	theirs := pushCreatureToBattlefieldForTest(g, them.ID, "Theirs")

	if err := g.AddCounter(pw, game.CounterLoyalty, 3); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := loyaltyOf(g, pw); got != 4 {
		t.Errorf("loyalty = %d, want 4 (3 plus one)", got)
	}
	if err := g.AddCounter(theirs, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter theirs: %v", err)
	}
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's counters = %d, want 1 (Pir is yours only)", got)
	}
}
