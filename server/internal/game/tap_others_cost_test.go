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
// The board-level behaviour (activating a real ability with this
// cost) waits on AbilityCost.TapOthers and the activation path in
// activated.go; the skipped test at the foot of this file names that
// seam so it is not quietly forgotten.

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

// tapOthersProbe records the EventTapCard the payment emits.
type tapOthersProbe struct{ tapped []uuid.UUID }

func (p *tapOthersProbe) OnEvent(_ *Game, ev Event) {
	if ev.Kind == EventTapCard {
		p.tapped = append(p.tapped, ev.CardID)
	}
}

// The other half of #758: AbilityCost.TapOthers itself, and the
// validate/pay calls in the activation path. Both live in
// server/internal/game/activated.go, which this PR does not touch —
// so an ability with this cost cannot yet be activated, and the
// engine-level behaviour (the cost paid at announce, refused
// activations leaving the board alone, the interaction with
// AbilityCost.Tap on the same ability) has nowhere to be tested.
//
// The component under it is complete and pinned above; what is owed
// is two lines on the struct and one call per half in the
// activation path.
func TestTapOthersOnAnActivatedAbility(t *testing.T) {
	t.Skip("blocked: AbilityCost.TapOthers and its validate/pay calls live in server/internal/game/activated.go, not touched by this PR — see #758")
}

// The mana-ability half, for Springleaf Drum, Jaspera Sentinel and
// Holdout Settlement. ManaAbilityShape is in effect_hooks.go, but
// the validation and payment it drives are in
// server/internal/game/mutations.go (ActivateManaAbility,
// ManaAbilityParams), which this PR does not touch.
func TestTapOthersOnAManaAbility(t *testing.T) {
	t.Skip("blocked: the mana-ability half is validated and paid in server/internal/game/mutations.go, not touched by this PR — see #758")
}
