package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// tap_others_cost_test.go — #758: TapOthersCost, the "tap N untapped
// [permanents] you control" component of an ability's cost
// (CR 602.2, CR 118.3), whose field ADR 0071 §2 names.
//
// What is pinned here is the component itself: the one candidate
// walk, the payability predicate, the validator and the payer. The
// rules that cost this seam its bugs are the ones that are NOT the
// {T} symbol — no summoning sickness, no targeting, "another" as a
// printed word rather than an engine rule — so those get a test
// each rather than a line in a table.
//
// The board-level behaviour — a real ability with this cost,
// activated, refused, and paid with the ability already on the stack
// — is at the foot of this file, on both halves: AbilityCost.TapOthers
// for CR 602 activated abilities and ManaAbilityShape.TapOthers for
// mana abilities (CR 605.3b).

// tapOthersCreatureFilter is "a creature you control" as a cost
// predicate — the in-package stand-in for the effects constructor
// ADR 0071 calls TapAnother(Creature()).
func tapOthersCreatureFilter() *TargetSpec {
	return &TargetSpec{
		Mode:  "permanent",
		Label: "an untapped creature you control",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
}

// pushTapOthersPermanent seats a permanent of the given type line
// under `controller`, tapped or not, and returns its instance ID.
func pushTapOthersPermanent(g *Game, controller *Player, name, typeLine string, tapped bool) uuid.UUID {
	c := NewCard(name, controller.ID)
	c.TypeLine = typeLine
	c.Controller = controller.ID
	c.Power, c.Toughness = 2, 2
	c.Tapped = tapped
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// tapOthersStationCost is the station shape (CR 702.184a): one
// ANOTHER untapped creature you control.
func tapOthersStationCost() *TapOthersCost {
	return &TapOthersCost{
		Count:         1,
		Filter:        tapOthersCreatureFilter(),
		ExcludeSource: true,
		Label:         "Tap another untapped creature you control",
	}
}

// tapOthersDruidCost is Heritage Druid's shape: three, and the
// source may be one of them.
func tapOthersDruidCost() *TapOthersCost {
	return &TapOthersCost{
		Count:  3,
		Filter: tapOthersCreatureFilter(),
		Label:  "Tap three untapped creatures you control",
	}
}

// The nil and zero costs demand nothing and must not be asked for a
// guard by every caller. A cost with a count but no filter is not a
// printed clause either.
func TestTapOthersCostIsInertWhenEmpty(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushTapOthersPermanent(g, me, "Bear", "Creature — Bear", false)

	var nilCost *TapOthersCost
	for name, tc := range map[string]*TapOthersCost{
		"nil":            nilCost,
		"zero":           {},
		"count, no spec": {Count: 1},
		"spec, no count": {Filter: tapOthersCreatureFilter()},
	} {
		if !tc.Empty() {
			t.Errorf("%s cost is not empty", name)
		}
		g.mu.Lock()
		if opts := g.TapOthersOptionsForEffect(me.ID, bear, tc); len(opts) != 0 {
			t.Errorf("%s cost offered %d options", name, len(opts))
		}
		if !g.TapOthersPayable(me.ID, bear, tc) {
			t.Errorf("%s cost is unpayable", name)
		}
		// IDs for an ability with no such component are a confused
		// client, and refused rather than ignored.
		if err := g.validateTapOthersCostLocked(me.ID, bear, tc, []uuid.UUID{bear}); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s cost accepted stray IDs: err = %v, want ErrInvalidParam", name, err)
		}
		if err := g.validateTapOthersCostLocked(me.ID, bear, tc, nil); err != nil {
			t.Errorf("%s cost rejected an empty payment: %v", name, err)
		}
		g.mu.Unlock()
	}
}

// The candidate walk offers only permanents that could actually pay:
// mine, untapped, and matching the filter.
func TestTapOthersOptionsOfferOnlyWhatCouldPay(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTapOthersPermanent(g, me, "Spacecraft", "Artifact — Spacecraft", false)
	ready := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)
	pushTapOthersPermanent(g, me, "Tired Bear", "Creature — Bear", true)
	pushTapOthersPermanent(g, opp, "Their Bear", "Creature — Bear", false)
	pushTapOthersPermanent(g, me, "Signet", "Artifact", false)

	g.mu.RLock()
	opts := g.TapOthersOptionsForEffect(me.ID, src, tapOthersStationCost())
	g.mu.RUnlock()
	if len(opts) != 1 || opts[0] != ready {
		t.Fatalf("options = %v, want only the untapped creature I control (%v)", opts, ready)
	}
}

// "Another" is a property of the printed cost. Station excludes its
// source; Azami, who taps a Wizard you control and is herself a
// Wizard, does not.
func TestTapOthersExcludeSourceIsThePrintedWordAnother(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	azami := pushTapOthersPermanent(g, me, "Azami", "Creature — Human Wizard", false)

	another := tapOthersStationCost()
	herself := tapOthersStationCost()
	herself.ExcludeSource = false

	g.mu.RLock()
	withAnother := g.TapOthersOptionsForEffect(me.ID, azami, another)
	withSelf := g.TapOthersOptionsForEffect(me.ID, azami, herself)
	g.mu.RUnlock()
	if len(withAnother) != 0 {
		t.Errorf("\"another\" offered the source itself: %v", withAnother)
	}
	if len(withSelf) != 1 || withSelf[0] != azami {
		t.Errorf("a cost without \"another\" refused the source: %v", withSelf)
	}

	g.mu.Lock()
	errAnother := g.validateTapOthersCostLocked(me.ID, azami, another, []uuid.UUID{azami})
	errSelf := g.validateTapOthersCostLocked(me.ID, azami, herself, []uuid.UUID{azami})
	g.mu.Unlock()
	if !errors.Is(errAnother, ErrInvalidParam) {
		t.Errorf("\"another\" accepted the source: err = %v, want ErrInvalidParam", errAnother)
	}
	if errSelf != nil {
		t.Errorf("a cost without \"another\" refused the source: %v", errSelf)
	}
}

// CR 118.3: a cost is payable or it is not. Two untapped Elves
// cannot pay Heritage Druid's three.
func TestTapOthersPayableNeedsCountUntappedPermanents(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushTapOthersPermanent(g, me, "Heritage Druid", "Creature — Elf Druid", false)
	cost := tapOthersDruidCost()

	// The source counts toward the three: this cost has no "another".
	pushTapOthersPermanent(g, me, "Elf One", "Creature — Elf", false)
	g.mu.RLock()
	twoOnBoard := g.TapOthersPayable(me.ID, src, cost)
	g.mu.RUnlock()
	if twoOnBoard {
		t.Error("two untapped creatures paid a cost of three")
	}

	pushTapOthersPermanent(g, me, "Elf Two", "Creature — Elf", false)
	g.mu.RLock()
	threeOnBoard := g.TapOthersPayable(me.ID, src, cost)
	g.mu.RUnlock()
	if !threeOnBoard {
		t.Error("three untapped creatures did not pay a cost of three")
	}
}

// Every way a naming is refused, and that none of them taps
// anything — ADR 0020 §3, validate everything then pay everything.
func TestTapOthersValidationRefusesAndTapsNothing(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTapOthersPermanent(g, me, "Spacecraft", "Artifact — Spacecraft", false)
	ready := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)
	spare := pushTapOthersPermanent(g, me, "Spare Bear", "Creature — Bear", false)
	tired := pushTapOthersPermanent(g, me, "Tired Bear", "Creature — Bear", true)
	theirs := pushTapOthersPermanent(g, opp, "Their Bear", "Creature — Bear", false)
	signet := pushTapOthersPermanent(g, me, "Signet", "Artifact", false)

	cases := []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"nothing named", nil, ErrInvalidParam},
		{"two named for a cost of one", []uuid.UUID{ready, spare}, ErrInvalidParam},
		{"the source, for an \"another\" cost", []uuid.UUID{src}, ErrInvalidParam},
		{"a card that does not exist", []uuid.UUID{uuid.New()}, ErrCardNotFound},
		{"an opponent's creature", []uuid.UUID{theirs}, ErrCardCallerMismatch},
		{"an already-tapped creature", []uuid.UUID{tired}, ErrAlreadyTapped},
		{"a permanent the filter does not match", []uuid.UUID{signet}, ErrIllegalTarget},
	}
	for _, tc := range cases {
		g.mu.Lock()
		err := g.validateTapOthersCostLocked(me.ID, src, tapOthersStationCost(), tc.ids)
		g.mu.Unlock()
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}

	// A cost of three refuses a duplicate: one creature cannot pay
	// two of the three.
	g.mu.Lock()
	dup := g.validateTapOthersCostLocked(me.ID, src, tapOthersDruidCost(), []uuid.UUID{ready, ready, spare})
	g.mu.Unlock()
	if !errors.Is(dup, ErrInvalidParam) {
		t.Errorf("a duplicate naming: err = %v, want ErrInvalidParam", dup)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, id := range []uuid.UUID{src, ready, spare, theirs, signet} {
		if c := findBattlefieldCard(g, id); c == nil || c.Tapped {
			t.Errorf("a refused validation tapped %v", id)
		}
	}
}

// The rule this component exists to get right: tapping to pay this
// cost is NOT the {T} symbol, so CR 302.6 / 602.5a do not apply and
// a creature that arrived this turn may pay.
func TestTapOthersIgnoresSummoningSickness(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushTapOthersPermanent(g, me, "Spacecraft", "Artifact — Spacecraft", false)
	fresh := pushTapOthersPermanent(g, me, "Fresh Bear", "Creature — Bear", false)

	g.mu.Lock()
	if c := findBattlefieldCard(g, fresh); c != nil {
		c.SummonedThisTurn = true
	}
	opts := g.TapOthersOptionsForEffect(me.ID, src, tapOthersStationCost())
	err := g.validateTapOthersCostLocked(me.ID, src, tapOthersStationCost(), []uuid.UUID{fresh})
	g.mu.Unlock()

	if len(opts) != 1 || opts[0] != fresh {
		t.Errorf("a summoning-sick creature was left out of the options: %v", opts)
	}
	if err != nil {
		t.Errorf("a summoning-sick creature was refused: %v", err)
	}
}

// Choosing a permanent to pay a cost does not target it (CR 601.2h /
// 602.2b), so shroud — the keyword that stops even its controller
// targeting it — is no obstacle. The same rule as sacrifice, crew
// and the counter cost.
func TestTapOthersIgnoresShroud(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushTapOthersPermanent(g, me, "Spacecraft", "Artifact — Spacecraft", false)
	shrouded := pushTapOthersPermanent(g, me, "Shrouded Bear", "Creature — Bear", false)

	g.mu.Lock()
	if c := findBattlefieldCard(g, shrouded); c != nil {
		c.Keywords = []string{"shroud"}
	}
	err := g.validateTapOthersCostLocked(me.ID, src, tapOthersStationCost(), []uuid.UUID{shrouded})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("a shrouded creature of mine was refused: %v", err)
	}

	g.mu.RLock()
	c := findBattlefieldCard(g, shrouded)
	targetable := c != nil && CanBeTargetedBy(c, ZoneBattlefield, SourceChooser(me.ID))
	g.mu.RUnlock()
	if targetable {
		t.Fatal("test setup: the creature should be untargetable, or this test proves nothing")
	}
}

// Payment taps every named permanent, emits one EventTapCard each so
// a "becomes tapped" payoff sees them all, and hands back the cards
// AS TAPPED — the snapshot a station ability reads its charge count
// off, so the creature dying afterwards changes nothing.
func TestTapOthersPaymentTapsEachAndReturnsTheSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushTapOthersPermanent(g, me, "Heritage Druid", "Creature — Elf Druid", false)
	one := pushTapOthersPermanent(g, me, "Elf One", "Creature — Elf", false)
	two := pushTapOthersPermanent(g, me, "Elf Two", "Creature — Elf", false)

	probe := &tapOthersProbe{}
	g.RegisterListener(probe)

	g.mu.Lock()
	if err := g.validateTapOthersCostLocked(me.ID, src, tapOthersDruidCost(), []uuid.UUID{src, one, two}); err != nil {
		g.mu.Unlock()
		t.Fatalf("validate three untapped Elves: %v", err)
	}
	paid := g.payTapOthersCostLocked(me.ID, []uuid.UUID{src, one, two})
	g.mu.Unlock()

	if len(paid) != 3 {
		t.Fatalf("payment returned %d cards, want 3", len(paid))
	}
	for _, c := range paid {
		if !c.Tapped {
			t.Errorf("%s came back untapped in the payment snapshot", c.Name)
		}
		if c.Power != 2 {
			t.Errorf("%s: snapshot power %d, want 2 — the number a station ability banks", c.Name, c.Power)
		}
	}
	if len(probe.tapped) != 3 {
		t.Errorf("EventTapCard fired %d times, want 3 (one per permanent)", len(probe.tapped))
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, id := range []uuid.UUID{src, one, two} {
		if c := findBattlefieldCard(g, id); c == nil || !c.Tapped {
			t.Errorf("%v was not tapped by the payment", id)
		}
	}
}

// A permanent that has already left cannot be tapped twice, and is
// skipped rather than panicking.
func TestTapOthersPaymentSkipsWhatItCannotTap(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ready := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)
	tired := pushTapOthersPermanent(g, me, "Tired Bear", "Creature — Bear", true)

	g.mu.Lock()
	paid := g.payTapOthersCostLocked(me.ID, []uuid.UUID{ready, tired, uuid.New()})
	g.mu.Unlock()
	if len(paid) != 1 || paid[0].InstanceID != ready {
		t.Fatalf("payment returned %d cards, want only the untapped one", len(paid))
	}
}

// tapOthersProbe records the EventTapCard the payment emits, and —
// for the ordering test — which stack item was already there when the
// first of them fired.
type tapOthersProbe struct {
	tapped         []uuid.UUID
	abilityOnStack uuid.UUID
}

func (p *tapOthersProbe) OnEvent(g *Game, ev Event) {
	if ev.Kind != EventTapCard {
		return
	}
	if len(p.tapped) == 0 {
		for id, item := range g.StackMeta {
			if item.Kind == StackItemActivated {
				p.abilityOnStack = id
			}
		}
	}
	p.tapped = append(p.tapped, ev.CardID)
}

// --- the activated-ability half (CR 602) ---------------------------

// pushTapOthersAbilitySource seats an artifact whose one activated
// ability costs `cost` and, on resolution, stamps a marker counter on
// itself — so a test can tell "on the stack" from "resolved".
func pushTapOthersAbilitySource(g *Game, owner *Player, cost AbilityCost) uuid.UUID {
	c := NewCard("Tap Others Source", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "tap others: mark",
		Cost:  cost,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// The headline: the named permanent taps at announce, the ability
// goes on the stack, and it resolves.
func TestTapOthersOnAnActivatedAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTapOthersAbilitySource(g, me, AbilityCost{TapOthers: tapOthersStationCost()})
	named := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)
	spare := pushTapOthersPermanent(g, me, "Spare Bear", "Creature — Bear", false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{named}}); err != nil {
		t.Fatalf("activate: %v", err)
	}

	g.mu.RLock()
	tappedNamed := findBattlefieldCard(g, named).Tapped
	tappedSpare := findBattlefieldCard(g, spare).Tapped
	tappedSource := findBattlefieldCard(g, src).Tapped
	g.mu.RUnlock()
	if !tappedNamed {
		t.Error("the named permanent was not tapped")
	}
	if tappedSpare {
		t.Error("a permanent that was not named was tapped")
	}
	// CR 702.122b's rule one component over: the cost taps OTHERS.
	// This ability prints no {T}, so its source stays untapped.
	if tappedSource {
		t.Error("the source tapped itself for a cost that does not print {T}")
	}

	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta has %d items, want 1 — the cost is paid at announce, the effect waits", len(g.StackMeta))
	}
	if counterOf(g, src, "effect-ran") != 0 {
		t.Error("the effect ran at announce")
	}
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability did not resolve")
	}
}

// Every way an activation with this cost is refused, and that none of
// them taps anything — ADR 0020 §3 through the real activation path,
// not just through the validator.
func TestTapOthersActivationRefusesAndTapsNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTapOthersAbilitySource(g, me, AbilityCost{TapOthers: tapOthersStationCost()})
	ready := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)
	spare := pushTapOthersPermanent(g, me, "Spare Bear", "Creature — Bear", false)
	tired := pushTapOthersPermanent(g, me, "Tired Bear", "Creature — Bear", true)
	theirs := pushTapOthersPermanent(g, opp, "Their Bear", "Creature — Bear", false)
	signet := pushTapOthersPermanent(g, me, "Signet", "Artifact", false)

	cases := []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"nothing named", nil, ErrInvalidParam},
		{"two named for a cost of one", []uuid.UUID{ready, spare}, ErrInvalidParam},
		{"the source, for an \"another\" cost", []uuid.UUID{src}, ErrInvalidParam},
		{"a card that does not exist", []uuid.UUID{uuid.New()}, ErrCardNotFound},
		{"an opponent's creature", []uuid.UUID{theirs}, ErrCardCallerMismatch},
		{"an already-tapped creature", []uuid.UUID{tired}, ErrAlreadyTapped},
		{"a permanent the filter does not match", []uuid.UUID{signet}, ErrIllegalTarget},
	}
	for _, tc := range cases {
		err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: tc.ids})
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("a refused activation put %d items on the stack", len(g.StackMeta))
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, id := range []uuid.UUID{src, ready, spare, theirs, signet} {
		if c := findBattlefieldCard(g, id); c == nil || c.Tapped {
			t.Errorf("a refused activation tapped %v", id)
		}
	}
}

// CR 118.3: a board that cannot produce Count untapped matches cannot
// activate the ability at all, and the predicate the enumerator and
// the client's greyed row read says so before it is tried.
func TestTapOthersActivationNeedsEnoughLegalPermanents(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTapOthersAbilitySource(g, me, AbilityCost{TapOthers: tapOthersDruidCost()})
	one := pushTapOthersPermanent(g, me, "Elf One", "Creature — Elf", false)
	two := pushTapOthersPermanent(g, me, "Elf Two", "Creature — Elf", false)

	g.mu.RLock()
	payable := g.TapOthersPayable(me.ID, src, tapOthersDruidCost())
	g.mu.RUnlock()
	if payable {
		t.Error("two untapped creatures reported a cost of three as payable")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{one, two}}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a short payment: err = %v, want ErrInvalidParam", err)
	}

	three := pushTapOthersPermanent(g, me, "Elf Three", "Creature — Elf", false)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{one, two, three}}); err != nil {
		t.Fatalf("three untapped Elves: %v", err)
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, id := range []uuid.UUID{one, two, three} {
		if c := findBattlefieldCard(g, id); c == nil || !c.Tapped {
			t.Errorf("%v was not tapped by the payment", id)
		}
	}
}

// CR 118.3 where two components of one cost meet: an ability that
// prints {T} as well has already spent its source, so the source may
// not also pay one of the N — even when the clause says nothing about
// "another" and the filter would admit it.
func TestTapOthersCannotPayWithASourceTheTapCostAlreadySpent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	// A creature source, so the filter genuinely matches it.
	src := pushTapOthersPermanent(g, me, "Jaspera Sentinel", "Creature — Elf", false)
	g.mu.Lock()
	findBattlefieldCard(g, src).ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "tap: mark",
		Cost: AbilityCost{
			Tap: true,
			TapOthers: &TapOthersCost{
				Count:  1,
				Filter: tapOthersCreatureFilter(),
				Label:  "Tap an untapped creature you control",
			},
		},
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	findBattlefieldCard(g, src).SummonedThisTurn = false
	g.mu.Unlock()
	friend := pushTapOthersPermanent(g, me, "Friend", "Creature — Elf", false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{src}}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("the source paid its own tap-others cost: err = %v, want ErrInvalidParam", err)
	}
	g.mu.RLock()
	srcTapped := findBattlefieldCard(g, src).Tapped
	g.mu.RUnlock()
	if srcTapped {
		t.Fatal("a refused activation tapped the source")
	}

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{friend}}); err != nil {
		t.Fatalf("activate with another creature: %v", err)
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, id := range []uuid.UUID{src, friend} {
		if c := findBattlefieldCard(g, id); c == nil || !c.Tapped {
			t.Errorf("%v was not tapped: the {T} and the tap-others half both have to be paid", id)
		}
	}
}

// ADR 0020 §4 / CR 603.3b: the cost is paid with the ability already
// on the stack, so a "becomes tapped" payoff goes ABOVE it. This is
// Opposition's whole archetype.
func TestTapOthersPaymentTriggersSitAboveTheAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTapOthersAbilitySource(g, me, AbilityCost{TapOthers: tapOthersStationCost()})
	bear := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)

	probe := &tapOthersProbe{}
	g.RegisterListener(probe)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(probe.tapped) != 1 || probe.tapped[0] != bear {
		t.Fatalf("EventTapCard: %v, want exactly the named permanent", probe.tapped)
	}
	// The ability was already on the stack when the tap fired, which
	// is what puts anything watching the tap above it.
	item, ok := g.StackMeta[probe.abilityOnStack]
	if !ok || item.SourceCardID != src {
		t.Error("the tap was paid before the ability reached the stack")
	}
}

// --- the mana-ability half (CR 605.3b) -----------------------------

// springleafDrumAbility is Springleaf Drum's shape: "{T}, Tap an
// untapped creature you control: Add one mana of any color."
func springleafDrumAbility() []ManaAbilityShape {
	return []ManaAbilityShape{{
		TapCost: true,
		TapOthers: &TapOthersCost{
			Count:  1,
			Filter: tapOthersCreatureFilter(),
			Label:  "Tap an untapped creature you control",
		},
		Produced: "{G}",
		Label:    "{T}, Tap an untapped creature you control: Add {G}",
	}}
}

// The mana-ability half end to end: both halves of the cost are paid
// and the mana lands (CR 605.3b — no stack, it resolves immediately).
func TestTapOthersOnAManaAbility(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	drum := pushIntrinsicPermanent(g, me, "Springleaf Drum", "Artifact", springleafDrumAbility(), nil)
	bear := pushTapOthersPermanent(g, me, "Ready Bear", "Creature — Bear", false)
	spare := pushTapOthersPermanent(g, me, "Spare Bear", "Creature — Bear", false)

	if err := g.ActivateManaAbility(me.ID, drum, 0, ManaAbilityParams{TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("mana pool = %d, want 1", len(me.ManaPool))
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	if c := findBattlefieldCard(g, drum); c == nil || !c.Tapped {
		t.Error("the {T} half of the cost was not paid")
	}
	if c := findBattlefieldCard(g, bear); c == nil || !c.Tapped {
		t.Error("the tap-others half of the cost was not paid")
	}
	if c := findBattlefieldCard(g, spare); c == nil || c.Tapped {
		t.Error("a permanent that was not named was tapped")
	}
}

// The refusals on the mana side, and that none of them taps the
// source — the same validate-all-then-pay discipline, through
// ActivateManaAbility this time.
func TestTapOthersManaAbilityRefusesAndTapsNothing(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drum := pushIntrinsicPermanent(g, me, "Springleaf Drum", "Artifact", springleafDrumAbility(), nil)
	tired := pushTapOthersPermanent(g, me, "Tired Bear", "Creature — Bear", true)
	theirs := pushTapOthersPermanent(g, opp, "Their Bear", "Creature — Bear", false)
	signet := pushTapOthersPermanent(g, me, "Signet", "Artifact", false)

	cases := []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"nothing named", nil, ErrInvalidParam},
		{"a card that does not exist", []uuid.UUID{uuid.New()}, ErrCardNotFound},
		{"an opponent's creature", []uuid.UUID{theirs}, ErrCardCallerMismatch},
		{"an already-tapped creature", []uuid.UUID{tired}, ErrAlreadyTapped},
		{"a permanent the filter does not match", []uuid.UUID{signet}, ErrIllegalTarget},
	}
	for _, tc := range cases {
		err := g.ActivateManaAbility(me.ID, drum, 0, ManaAbilityParams{TapIDs: tc.ids})
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("a refused activation produced %d mana", len(me.ManaPool))
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, id := range []uuid.UUID{drum, theirs, signet} {
		if c := findBattlefieldCard(g, id); c == nil || c.Tapped {
			t.Errorf("a refused activation tapped %v", id)
		}
	}
}

// The auto-tapper must never plan a mana source whose cost taps other
// permanents: it would spend a blocker the player never offered, and
// the planner has no way to weigh that (the same bar a life cost
// fails).
func TestTapOthersManaAbilityIsNeverAutoTapped(t *testing.T) {
	// #1183: the picker is a method now — one of its exclusions is a
	// fact about the game (a spent exhaust ability), not about the
	// shape. Nothing here activates anything, so the record is empty
	// and every exclusion under test is still the shape's.
	g := newActiveGame(t)
	g.mu.Lock()
	defer g.mu.Unlock()
	if a := g.autoTapAbilityFor(uuid.New(), Card{InstanceID: uuid.New()}, springleafDrumAbility()); a != nil {
		t.Errorf("the auto-tapper planned a tap-others mana ability: %+v", a)
	}
	plain := []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "{T}: Add {G}"}}
	if a := g.autoTapAbilityFor(uuid.New(), Card{InstanceID: uuid.New()}, plain); a == nil {
		t.Error("test setup: an ordinary {T} mana ability should still be plannable")
	}
}
