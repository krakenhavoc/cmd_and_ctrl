package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// waterbend_cost_test.go — #1310 / #1311: waterbend (CR 701.67) as an
// activated ability's cost and as a pay-or-counter prompt's payment.
//
// The component is tap_cost.go's TapPermanentsCost; what is pinned
// here is the two new owners — the CR 602 activation path and the
// CR 118.12 pay-unless prompt — and the rules each has to get right:
// a tap pays exactly {1} of the waterbend's generic and nothing else
// (CR 701.67b), a refused payment taps nothing (ADR 0020 §3), a
// tapped permanent is out of the auto-tapper's reach (CR 118.3), and
// tapping to waterbend is not the {T} symbol (CR 302.6).

// waterbendTestCost is effects.Waterbend's clause without the effects
// import: "an untapped artifact or creature you control" with `extra`
// as the waterbend cost.
func waterbendTestCost(extra string) *TapPermanentsCost {
	return &TapPermanentsCost{
		Key:   "waterbend",
		Label: "Waterbend " + extra,
		Extra: extra,
		Spec: &TargetSpec{
			Mode:  "permanent",
			Label: "an untapped artifact or creature you control",
			Zones: []ZoneKind{ZoneBattlefield},
			CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
				return c.IsArtifact() || c.IsCreature()
			},
		},
	}
}

// waterbendAbilityCost is effects.WaterbendCost: the mana in the mana
// component, the clause beside it.
func waterbendAbilityCost(extra string) AbilityCost {
	return AbilityCost{Mana: extra, Waterbend: waterbendTestCost(extra)}
}

func tappedIn(g *Game, id uuid.UUID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	c := findBattlefieldCard(g, id)
	return c != nil && c.Tapped
}

func TestWaterbendBudgetIsTheClauseCappedByWhatIsOwed(t *testing.T) {
	wb := waterbendTestCost("{8}")
	eight, _ := ParseCost("{8}")
	if got := WaterbendBudget(wb, eight, 0); got != 8 {
		t.Errorf("budget = %d, want 8", got)
	}
	// A CR 601.2f discount that took {2} off leaves six symbols for a
	// tap to pay. Tapping a seventh creature would tap it for nothing.
	six, _ := ParseCost("{6}")
	if got := WaterbendBudget(wb, six, 0); got != 6 {
		t.Errorf("discounted budget = %d, want 6", got)
	}
	// CR 701.67b: "{1}{U}, Waterbend {2}" — only the waterbend's own
	// generic, never the rest of the total's.
	total, _ := ParseCost("{1}{U}{2}")
	if got := WaterbendBudget(waterbendTestCost("{2}"), total, 0); got != 2 {
		t.Errorf("composite budget = %d, want 2", got)
	}
	// Waterbend {X} is sized by the announced X.
	x, _ := ParseCost("{X}")
	if got := WaterbendBudget(waterbendTestCost("{X}"), x, 4); got != 4 {
		t.Errorf("X budget = %d, want 4", got)
	}
	if got := WaterbendBudget(nil, eight, 0); got != 0 {
		t.Errorf("nil clause budget = %d, want 0", got)
	}

	reduced := WaterbendReduced(x, 4, 3)
	if reduced.Generic != 1 || reduced.XSlots != 0 {
		t.Errorf("reduced = %+v, want {1} with X folded", reduced)
	}
	colored, _ := ParseCost("{1}{U}{2}")
	if r := WaterbendReduced(colored, 0, 2); r.Generic != 1 || len(r.Required) != 1 {
		t.Errorf("a tap paid a coloured symbol: %+v", r)
	}
}

// The headline: every generic symbol paid by tapping, in strict mode
// with an empty pool, and the source helps pay for its own ability —
// it prints no {T}, and it arrived this turn (CR 302.6 does not apply).
func TestWaterbendAbilityPaidEntirelyByTapping(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTapOthersAbilitySource(g, me, waterbendAbilityCost("{3}"))
	g.mu.Lock()
	s := findBattlefieldCard(g, src)
	s.TypeLine = "Artifact Creature — Construct"
	s.Power, s.Toughness = 2, 2
	s.SummonedThisTurn = true
	g.mu.Unlock()
	bear := pushTapOthersPermanent(g, me, "Bear", "Creature — Bear", false)
	rock := pushTapOthersPermanent(g, me, "Rock", "Artifact", false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		WaterbendIDs: []uuid.UUID{src, bear, rock},
		Strict:       true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range []uuid.UUID{src, bear, rock} {
		if !tappedIn(g, id) {
			t.Errorf("%v was not tapped to waterbend", id)
		}
	}
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability did not resolve")
	}
}

// Taps and mana together: two taps, one {C} from the pool.
func TestWaterbendAbilitySplitsBetweenTapsAndMana(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTapOthersAbilitySource(g, me, waterbendAbilityCost("{3}"))
	a := pushTapOthersPermanent(g, me, "Bear A", "Creature — Bear", false)
	b := pushTapOthersPermanent(g, me, "Bear B", "Creature — Bear", false)
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		WaterbendIDs: []uuid.UUID{a, b},
		Strict:       true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := len(me.ManaPool); n != 0 {
		t.Errorf("pool has %d mana left, want the {C} spent", n)
	}
	if !tappedIn(g, a) || !tappedIn(g, b) {
		t.Error("the named permanents were not tapped")
	}
}

// Short on mana after the taps: refused, and NOTHING tapped — the mana
// half is settled before the taps are paid.
func TestWaterbendAbilityShortOnManaTapsNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTapOthersAbilitySource(g, me, waterbendAbilityCost("{3}"))
	a := pushTapOthersPermanent(g, me, "Bear A", "Creature — Bear", false)
	b := pushTapOthersPermanent(g, me, "Bear B", "Creature — Bear", false)

	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		WaterbendIDs: []uuid.UUID{a, b},
		Strict:       true,
	})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("err = %v, want insufficient mana", err)
	}
	if tappedIn(g, a) || tappedIn(g, b) {
		t.Error("a refused activation tapped a waterbender")
	}
	if len(g.StackMeta) != 0 {
		t.Error("a refused activation reached the stack")
	}
}

// Every malformed tap list is refused and taps nothing.
func TestWaterbendAbilityRefusalsTapNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTapOthersAbilitySource(g, me, waterbendAbilityCost("{2}"))
	tapSrc := pushTapOthersAbilitySource(g, me, withTapSymbol(waterbendAbilityCost("{2}"), AbilityCost{Tap: true}))
	plain := pushTapOthersAbilitySource(g, me, AbilityCost{Mana: "{2}"})
	a := pushTapOthersPermanent(g, me, "Bear A", "Creature — Bear", false)
	b := pushTapOthersPermanent(g, me, "Bear B", "Creature — Bear", false)
	c := pushTapOthersPermanent(g, me, "Bear C", "Creature — Bear", false)
	tired := pushTapOthersPermanent(g, me, "Tired Bear", "Creature — Bear", true)
	theirs := pushTapOthersPermanent(g, opp, "Their Bear", "Creature — Bear", false)
	land := pushTapOthersPermanent(g, me, "Wastes", "Land", false)
	for i := 0; i < 4; i++ {
		me.ManaPool.AddMana(ManaToken{Color: "C"})
	}

	cases := []struct {
		name string
		src  uuid.UUID
		ids  []uuid.UUID
		want error
	}{
		{"more taps than the waterbend's generic", src, []uuid.UUID{a, b, c}, ErrInvalidParam},
		{"one permanent named twice", src, []uuid.UUID{a, a}, ErrInvalidParam},
		{"an opponent's creature", src, []uuid.UUID{theirs}, ErrInvalidParam},
		{"an already-tapped creature", src, []uuid.UUID{tired}, ErrInvalidParam},
		{"a land, which is neither artifact nor creature", src, []uuid.UUID{land}, ErrInvalidParam},
		{"a card that does not exist", src, []uuid.UUID{uuid.New()}, ErrCardNotFound},
		{"the source, when the cost also prints {T}", tapSrc, []uuid.UUID{tapSrc}, ErrInvalidParam},
		{"waterbend_ids on an ability with no waterbend", plain, []uuid.UUID{a}, ErrInvalidParam},
	}
	for _, tc := range cases {
		err := g.ActivateCatalogAbility(me.ID, tc.src, 0, ActivateAbilityParams{WaterbendIDs: tc.ids, Strict: true})
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	for _, id := range []uuid.UUID{src, tapSrc, a, b, c, theirs, land} {
		if tappedIn(g, id) {
			t.Errorf("a refused activation tapped %v", id)
		}
	}
	if n := len(me.ManaPool); n != 4 {
		t.Errorf("a refused activation spent mana: pool %d, want 4", n)
	}
}

// CR 118.3 through the auto-tapper: a mana creature named to the
// waterbend may not ALSO tap for the rest of the cost. Here it is the
// only mana source on the board, so the only honest answer is "short".
// Without the exclusion the planner taps it for {G}, the waterbend
// payer then skips it as already tapped, and one creature pays {2}.
func TestWaterbendTapperIsOutOfTheAutoTappersReach(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTapOthersAbilitySource(g, me, waterbendAbilityCost("{2}"))
	birds := pushTapOthersPermanent(g, me, "Birds", "Creature — Bird", false)
	g.mu.Lock()
	b := findBattlefieldCard(g, birds)
	b.ManaAbilities = []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}
	b.SummonedThisTurn = false
	g.mu.Unlock()

	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		WaterbendIDs: []uuid.UUID{birds},
		Strict:       true,
		AutoTap:      true,
	})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("err = %v, want insufficient mana — the waterbender paid twice", err)
	}
	if tappedIn(g, birds) {
		t.Error("a refused activation tapped the waterbender")
	}
}

// Waterbend {X} with a floor: X is announced, and the taps are capped
// by it.
func TestWaterbendXAbilityIsSizedByTheAnnouncedX(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cost := waterbendAbilityCost("{X}")
	cost.MinX = 1
	src := pushTapOthersAbilitySource(g, me, cost)
	a := pushTapOthersPermanent(g, me, "Bear A", "Creature — Bear", false)
	b := pushTapOthersPermanent(g, me, "Bear B", "Creature — Bear", false)
	c := pushTapOthersPermanent(g, me, "Bear C", "Creature — Bear", false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		XValue: 2, WaterbendIDs: []uuid.UUID{a, b, c}, Strict: true,
	}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("three taps at X=2: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		XValue: 0, Strict: true,
	}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("X=0 under a floor of 1: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		XValue: 2, WaterbendIDs: []uuid.UUID{a, b}, Strict: true,
	}); err != nil {
		t.Fatalf("two taps at X=2: %v", err)
	}
	if tappedIn(g, c) {
		t.Error("a permanent that was not named was tapped")
	}
	for id, item := range g.StackMeta {
		if item.XValue != 2 {
			t.Errorf("stack item %v carries X=%d, want 2", id, item.XValue)
		}
	}
}

// The taps are paid with the ability already on the stack, so a
// "becomes tapped" payoff sits above it — the ordering the cast path's
// convoke taps and #758's tap-another taps keep.
func TestWaterbendTapsArePaidWithTheAbilityOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTapOthersAbilitySource(g, me, waterbendAbilityCost("{1}"))
	bear := pushTapOthersPermanent(g, me, "Bear", "Creature — Bear", false)
	probe := &tapOthersProbe{}
	g.RegisterListener(probe)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		WaterbendIDs: []uuid.UUID{bear}, Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(probe.tapped) != 1 || probe.tapped[0] != bear {
		t.Fatalf("tap events = %v, want exactly the bear", probe.tapped)
	}
	if probe.abilityOnStack == uuid.Nil {
		t.Error("the waterbend tap fired before the ability was on the stack")
	}
}

// withTapSymbol composes two in-package costs the way effects.Plus does for
// the two fields these tests need.
func withTapSymbol(a, b AbilityCost) AbilityCost {
	out := a
	out.Tap = out.Tap || b.Tap
	return out
}

// --- the pay-or-counter half (#1311, "Ward—Waterbend {4}") ----------

// queueWaterbendWardTax is the ward leg a "Ward—Waterbend {N}" permanent
// queues, straight at the engine.
func queueWaterbendWardTax(t *testing.T, g *Game, spell, chooser uuid.UUID, cost string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueueCounterUnlessPaidForEffect(CounterUnlessPaidPrompt{
			StackItem: spell,
			Chooser:   chooser,
			Source:    uuid.New(),
			Cost:      cost,
			Question:  "Ward — waterbend " + cost + " or the spell is countered",
			Waterbend: waterbendTestCost(cost),
		}); err != nil {
			t.Fatalf("QueueCounterUnlessPaidForEffect: %v", err)
		}
	})
}

func spellStillOnStack(g *Game, id uuid.UUID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.StackItemForEffect(id) != nil
}

// The ward paid entirely by tapping: four taps, no mana, the spell
// survives, and the prompt carries the clause the view projects.
func TestWaterbendWardPaidByTapping(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	var ids []uuid.UUID
	for _, tl := range []string{"Creature — Bear", "Creature — Bear", "Artifact", "Artifact Creature — Golem"} {
		ids = append(ids, pushTapOthersPermanent(g, caster, "Helper", tl, false))
	}
	queueWaterbendWardTax(t, g, spell, caster.ID, "{4}")
	prompt := onlyPendingChoice(t, g)
	if tc := prompt.PayTapCost(); tc == nil || tc.Key != "waterbend" {
		t.Fatalf("PayTapCost = %+v, want the waterbend clause", tc)
	}

	if err := g.ResolvePayUnlessWithTaps(prompt.ID, caster.ID, true, ids); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if !spellStillOnStack(g, spell) {
		t.Error("a paid waterbend ward countered the spell")
	}
	for _, id := range ids {
		if !tappedIn(g, id) {
			t.Errorf("%v was not tapped to pay the ward", id)
		}
	}
}

// Taps short of the whole cost, and no mana for the rest: the answer
// degrades to a decline, the spell is countered, and nothing is
// tapped — the mana half is settled first.
func TestWaterbendWardUnfundedRemainderTapsNothingAndCounters(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	a := pushTapOthersPermanent(g, caster, "Bear A", "Creature — Bear", false)
	b := pushTapOthersPermanent(g, caster, "Bear B", "Creature — Bear", false)
	queueWaterbendWardTax(t, g, spell, caster.ID, "{4}")
	prompt := onlyPendingChoice(t, g)

	if err := g.ResolvePayUnlessWithTaps(prompt.ID, caster.ID, true, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if spellStillOnStack(g, spell) {
		t.Error("an unfunded ward payment left the spell on the stack")
	}
	if tappedIn(g, a) || tappedIn(g, b) {
		t.Error("an unfunded ward payment tapped the waterbenders")
	}
}

// Two taps and {2} of mana in the pool pay the {4}.
func TestWaterbendWardSplitsBetweenTapsAndMana(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	a := pushTapOthersPermanent(g, caster, "Bear A", "Creature — Bear", false)
	b := pushTapOthersPermanent(g, caster, "Bear B", "Creature — Bear", false)
	caster.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	queueWaterbendWardTax(t, g, spell, caster.ID, "{4}")
	prompt := onlyPendingChoice(t, g)

	if err := g.ResolvePayUnlessWithTaps(prompt.ID, caster.ID, true, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if !spellStillOnStack(g, spell) {
		t.Error("a paid ward countered the spell")
	}
	if len(caster.ManaPool) != 0 {
		t.Errorf("pool has %d left, want the {2} spent", len(caster.ManaPool))
	}
}

// A malformed tap list is REFUSED with the prompt left in place — a
// client bug must not cost the payer their spell.
func TestWaterbendWardRefusesAMalformedTapList(t *testing.T) {
	g := newActiveGame(t)
	caster, warder := g.Seats[1], g.Seats[0]
	spell := guardedSpellOnStack(t, g, caster)
	mine := pushTapOthersPermanent(g, caster, "Bear", "Creature — Bear", false)
	theirs := pushTapOthersPermanent(g, warder, "Their Bear", "Creature — Bear", false)
	land := pushTapOthersPermanent(g, caster, "Wastes", "Land", false)
	var five []uuid.UUID
	for i := 0; i < 5; i++ {
		five = append(five, pushTapOthersPermanent(g, caster, "Rock", "Artifact", false))
	}
	queueWaterbendWardTax(t, g, spell, caster.ID, "{4}")
	prompt := onlyPendingChoice(t, g)

	for _, tc := range []struct {
		name  string
		apply bool
		ids   []uuid.UUID
	}{
		{"the warder's creature", true, []uuid.UUID{theirs}},
		{"a land", true, []uuid.UUID{land}},
		{"more than the generic", true, five},
		{"taps on a decline", false, []uuid.UUID{mine}},
	} {
		if err := g.ResolvePayUnlessWithTaps(prompt.ID, caster.ID, tc.apply, tc.ids); err == nil {
			t.Errorf("%s: accepted", tc.name)
		}
	}
	if len(g.PendingChoices) != 1 {
		t.Fatal("a refused answer consumed the prompt")
	}
	if !spellStillOnStack(g, spell) {
		t.Error("a refused answer countered the spell")
	}
	for _, id := range append([]uuid.UUID{mine, theirs, land}, five...) {
		if tappedIn(g, id) {
			t.Errorf("a refused answer tapped %v", id)
		}
	}
}

// CR 118.3 on the ward side: a mana creature tapped to waterbend may
// not also be auto-tapped for the rest. With it as the payer's only
// mana source, a {2} ward "paid" with it alone is a decline.
func TestWaterbendWardTapperIsOutOfTheAutoTappersReach(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	birds := pushTapOthersPermanent(g, caster, "Birds", "Creature — Bird", false)
	g.mu.Lock()
	b := findBattlefieldCard(g, birds)
	b.ManaAbilities = []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}
	g.mu.Unlock()
	queueWaterbendWardTax(t, g, spell, caster.ID, "{2}")
	prompt := onlyPendingChoice(t, g)

	if err := g.ResolvePayUnlessWithTaps(prompt.ID, caster.ID, true, []uuid.UUID{birds}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if spellStillOnStack(g, spell) {
		t.Error("one creature paid a {2} ward — tapped once to waterbend and again for mana")
	}
}

// An ordinary mana ward refuses tap_ids outright — it has no clause.
func TestManaWardRefusesTapIDs(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	bear := pushTapOthersPermanent(g, caster, "Bear", "Creature — Bear", false)
	queueWardTax(t, g, spell, caster.ID)
	prompt := onlyPendingChoice(t, g)
	if prompt.PayTapCost() != nil {
		t.Fatal("a mana ward carries a waterbend clause")
	}
	if err := g.ResolvePayUnlessWithTaps(prompt.ID, caster.ID, true, []uuid.UUID{bear}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
}
