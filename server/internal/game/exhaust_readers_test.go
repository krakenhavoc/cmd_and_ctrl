package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exhaust_readers_test.go — #1184, the three seams that read the
// exhaust record from OUTSIDE the ability that owns it.
//
// #1181 built the record and one reader; these are the three
// mechanisms four printed cards need on top of it, and they are
// pinned here in the engine rather than only on the cards so that a
// card file being deleted cannot quietly delete the rule:
//
//  1. the ACTIVATION EVENT — one event, carrying the ability's
//     identity and the exhaust bit, that a trigger can watch;
//  2. the PERMISSION — the gate answered per asking player;
//  3. the COST MODIFIER — the CR 601.2f pass run over an ability's
//     mana component, with a predicate that can see the ability.

// --- 1. the activation event ----------------------------------------

// activationEventsOf collects the activation-shaped events in the log.
func activationEventsOf(g *Game, kind EventKind) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}

// TestActivatingAnAbilityEmitsAnEventNamingIt is the whole of seam 1:
// before #1184 the announce emitted EventTrigger with no ability on
// it, so nothing could tell one activation from another — or an
// exhaust one from an ordinary one.
func TestActivatingAnAbilityEmitsAnEventNamingIt(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-event",
		exhaustProbe(exhaustLabelA, true),
		exhaustProbe(plainLabel, false),
	)
	src := pushExhaustSource(g, me, "probe-event")

	if err := fireExhaust(t, g, me, src, 1); err != nil {
		t.Fatalf("the plain activation: %v", err)
	}
	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("the exhaust activation: %v", err)
	}

	evs := activationEventsOf(g, EventActivateAbility)
	if len(evs) != 2 {
		t.Fatalf("EventActivateAbility count = %d, want 2", len(evs))
	}
	if evs[0].Label != plainLabel || evs[0].Exhaust {
		t.Errorf("the plain activation's event = {label:%q exhaust:%v}, want the plain label and no exhaust bit",
			evs[0].Label, evs[0].Exhaust)
	}
	if evs[1].Label != exhaustLabelA || !evs[1].Exhaust {
		t.Errorf("the exhaust activation's event = {label:%q exhaust:%v}, want %q and the exhaust bit",
			evs[1].Label, evs[1].Exhaust, exhaustLabelA)
	}
	for i, ev := range evs {
		if ev.Actor != me.ID {
			t.Errorf("event %d: actor = %s, want the activator %s", i, ev.Actor, me.ID)
		}
		if ev.Source != src || ev.CardID != src {
			t.Errorf("event %d: source/card = %s/%s, want the ability's source %s", i, ev.Source, ev.CardID, src)
		}
	}
}

// TestTheActivationEventIsHarvestable is the property the EventTrigger
// breadcrumb could never have: triggerHarvester.OnEvent returns
// immediately on EventTrigger, so a trigger watching it sees nothing.
// The new kind walks the harvest like any other event.
func TestTheActivationEventIsHarvestable(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-harvest",
		exhaustProbe(exhaustLabelA, true),
		exhaustProbe(plainLabel, false),
	)
	src := pushExhaustSource(g, me, "probe-harvest")

	// A watcher whose whole condition is the two stamps: your
	// activation, and an exhaust one.
	var fired int
	watcherCatalog(t, "probe-watcher", TriggeredAbility{
		Watches: []EventKind{EventActivateAbility},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.Actor == source.Controller && ev.Exhaust
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return newTriggeredItemForTest(source, "watcher — count one", func(*Game, *StackItem) error {
				fired++
				return nil
			})
		},
	})
	pushExhaustSource(g, me, "probe-watcher")

	if err := fireExhaust(t, g, me, src, 1); err != nil { // not an exhaust ability
		t.Fatalf("plain activation: %v", err)
	}
	passBothForTest(g)
	if fired != 0 {
		t.Fatalf("the watcher fired %d times on an ordinary activation, want 0", fired)
	}

	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("exhaust activation: %v", err)
	}
	passBothForTest(g)
	if fired != 1 {
		t.Errorf("the watcher fired %d times on an exhaust activation, want 1 — the announce's EventTrigger is not harvestable and this event is", fired)
	}
}

// watcherCatalog installs one triggered ability on `oracle` for the
// length of the test.
func watcherCatalog(t *testing.T, oracle string, triggers ...TriggeredAbility) {
	t.Helper()
	prev := CatalogTriggers
	CatalogTriggers = func(key string) []TriggeredAbility {
		if key == oracle {
			return triggers
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogTriggers = prev })
}

// TestAMananAbilityActivationCarriesTheSameTwoStamps — CR 605.1a: a
// mana ability is an activated ability, so "whenever you activate an
// exhaust ability" has to see Loot, the Pathfinder's. The mana path
// keeps its own event kind and gained the same two fields.
func TestAMananAbilityActivationCarriesTheSameTwoStamps(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	const label = "Exhaust — {T}: Add {G}{G}{G}"
	manaExhaustCatalog(t, "probe-mana", ManaAbilityShape{
		Label:    label,
		Exhaust:  true,
		TapCost:  true,
		Produced: "{G}{G}{G}",
	})
	src := pushExhaustSource(g, me, "probe-mana")

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	evs := activationEventsOf(g, EventManaAbilityActivated)
	if len(evs) != 1 {
		t.Fatalf("EventManaAbilityActivated count = %d, want 1", len(evs))
	}
	if evs[0].Label != label || !evs[0].Exhaust {
		t.Errorf("mana activation event = {label:%q exhaust:%v}, want %q and the exhaust bit",
			evs[0].Label, evs[0].Exhaust, label)
	}
}

// manaExhaustCatalog installs a mana-ability catalog entry for the
// length of the test, mirroring exhaustCatalog.
func manaExhaustCatalog(t *testing.T, oracle string, abilities ...ManaAbilityShape) {
	t.Helper()
	prev := CatalogManaAbilities
	CatalogManaAbilities = func(key string) []ManaAbilityShape {
		if key == oracle {
			return abilities
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogManaAbilities = prev })
}

// --- 2. the permission ----------------------------------------------

// permissionCatalog installs one exhaust permission on `oracle` for
// the length of the test.
func permissionCatalog(t *testing.T, oracle string, applies func(g *Game, asker uuid.UUID, source Card) bool) {
	t.Helper()
	prev := CatalogExhaustPermissions
	CatalogExhaustPermissions = func(key string) []ExhaustPermission {
		if key == oracle {
			return []ExhaustPermission{{Label: "test permission", Applies: applies}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogExhaustPermissions = prev })
}

// TestAPermissionSuspendsTheGateForItsOwnerOnly is the shape of seam
// 2: the record still says the ability was activated — nothing was
// cleared — and one player is allowed to act as though it did not.
func TestAPermissionSuspendsTheGateForItsOwnerOnly(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	exhaustCatalog(t, "probe-perm", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-perm")

	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	passBothForTest(g)

	shape := exhaustProbe(exhaustLabelA, true)
	var spentForMe, spentForThem bool
	g.WithWriteLock(func() {
		spentForMe = g.AbilityExhausted(me.ID, src, shape)
		spentForThem = g.AbilityExhausted(them.ID, src, shape)
	})
	if !spentForMe || !spentForThem {
		t.Fatalf("with no permission on the board the ability reads spent for everyone; got me=%v them=%v", spentForMe, spentForThem)
	}

	// The permission arrives, granted to `me` and nobody else.
	permissionCatalog(t, "probe-grant", func(_ *Game, asker uuid.UUID, source Card) bool {
		return asker == source.Controller
	})
	grant := pushExhaustSource(g, me, "probe-grant")
	_ = grant

	g.WithWriteLock(func() {
		spentForMe = g.AbilityExhausted(me.ID, src, shape)
		spentForThem = g.AbilityExhausted(them.ID, src, shape)
	})
	if spentForMe {
		t.Error("the permission's controller still reads the ability as spent")
	}
	if !spentForThem {
		t.Error("the permission reached a player it was not granted to")
	}

	// The RECORD is untouched — "as though" changes no fact (CR 609.4).
	if n := exhaustedThisGame(g, src, exhaustLabelA); n != 1 {
		t.Errorf("Activations.Ever = %d, want the original 1 — a permission clears nothing", n)
	}
}

// TestAPermissionLetsTheAbilityBeActivatedAgainAndIsThenSpent — the
// end-to-end of seam 2, and the reason the record is not cleared: the
// second activation WRITES the record again, so a card whose
// permission is conditioned on "you haven't activated an exhaust
// ability this turn" turns itself off.
func TestAPermissionLetsTheAbilityBeActivatedAgainAndIsThenSpent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-again", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-again")
	permissionCatalog(t, "probe-again-grant", func(g *Game, asker uuid.UUID, source Card) bool {
		return asker == source.Controller && g.ExhaustAbilitiesActivatedThisTurn(asker) < 2
	})
	pushExhaustSource(g, me, "probe-again-grant")

	for i := 0; i < 2; i++ {
		if err := fireExhaust(t, g, me, src, 0); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		passBothForTest(g)
	}
	if got := counterOf(g, src, "ran-"+exhaustLabelA); got != 2 {
		t.Fatalf("the ability resolved %d times, want 2 — the permission bought one more", got)
	}
	if n := exhaustedThisGame(g, src, exhaustLabelA); n != 2 {
		t.Errorf("Activations.Ever = %d, want 2 — the second activation is recorded like the first", n)
	}
	// The permission's own condition has now gone false, so the third
	// is refused with the ordinary error.
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); !errors.Is(err, ErrAbilityExhausted) {
		t.Fatalf("third activation: err = %v, want ErrAbilityExhausted", err)
	}
}

// TestExhaustActivationsThisTurnCountsBothKindsAndOnlyExhaust is the
// counter the printed condition reads.
func TestExhaustActivationsThisTurnCountsBothKindsAndOnlyExhaust(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	exhaustCatalog(t, "probe-count",
		exhaustProbe(exhaustLabelA, true),
		exhaustProbe(plainLabel, false),
	)
	src := pushExhaustSource(g, me, "probe-count")

	if err := fireExhaust(t, g, me, src, 1); err != nil { // not an exhaust ability
		t.Fatalf("plain activation: %v", err)
	}
	passBothForTest(g)
	if n := exhaustActivationsFor(g, me.ID); n != 0 {
		t.Fatalf("after a non-exhaust activation the count = %d, want 0", n)
	}
	if err := fireExhaust(t, g, me, src, 0); err != nil {
		t.Fatalf("exhaust activation: %v", err)
	}
	passBothForTest(g)
	if n := exhaustActivationsFor(g, me.ID); n != 1 {
		t.Errorf("after an exhaust activation the count = %d, want 1", n)
	}
	if n := exhaustActivationsFor(g, them.ID); n != 0 {
		t.Errorf("the opponent's count = %d, want 0 — it is a per-player number", n)
	}
}

func exhaustActivationsFor(g *Game, p uuid.UUID) int {
	var n int
	g.WithWriteLock(func() { n = g.ExhaustAbilitiesActivatedThisTurn(p) })
	return n
}

// --- 3. the cost modifier -------------------------------------------

// costModifierCatalog installs one cost modifier on `oracle` for the
// length of the test.
func costModifierCatalog(t *testing.T, oracle string, mods ...CostModifier) {
	t.Helper()
	prev := CatalogCostModifiers
	CatalogCostModifiers = func(key string) []CostModifier {
		if key == oracle {
			return mods
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogCostModifiers = prev })
}

// TestAnActivationCostModifierReadsTheAbility is seam 3: the query
// carries the ability, so a modifier can say "the exhaust ones".
func TestAnActivationCostModifierReadsTheAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-price",
		exhaustProbe(exhaustLabelA, true),
		exhaustProbe(plainLabel, false),
	)
	src := pushExhaustSource(g, me, "probe-price")
	costModifierCatalog(t, "probe-discount", CostModifier{
		Kind:        CostReduction,
		Activations: true,
		Label:       "Exhaust abilities of other permanents you control cost {2} less to activate.",
		AppliesTo: func(q CostQuery) bool {
			return q.Ability != nil && q.Ability.Exhaust &&
				q.Card.Controller == q.Source.Controller &&
				q.Card.InstanceID != q.Source.InstanceID
		},
		Amount: func(CostQuery) int { return 2 },
	})

	var exhaustCost, plainCost ParsedCost
	g.WithWriteLock(func() {
		card := *g.findCardByIDLocked(src)
		abilities := ActivatedAbilitiesForCard(card)
		exhaustCost, _ = g.AbilityManaCostForEffect(me.ID, card, ZoneBattlefield, abilities[0])
		plainCost, _ = g.AbilityManaCostForEffect(me.ID, card, ZoneBattlefield, abilities[1])
	})
	if exhaustCost.Generic != 1 || plainCost.Generic != 1 {
		t.Fatalf("with no discounter on the board: exhaust {%d}, plain {%d}, want {1} and {1}",
			exhaustCost.Generic, plainCost.Generic)
	}

	pushExhaustSource(g, me, "probe-discount")
	g.WithWriteLock(func() {
		card := *g.findCardByIDLocked(src)
		abilities := ActivatedAbilitiesForCard(card)
		exhaustCost, _ = g.AbilityManaCostForEffect(me.ID, card, ZoneBattlefield, abilities[0])
		plainCost, _ = g.AbilityManaCostForEffect(me.ID, card, ZoneBattlefield, abilities[1])
	})
	// CR 601.2f: a reduction floors at zero generic.
	if exhaustCost.Generic != 0 {
		t.Errorf("the exhaust ability costs {%d}, want {0} — {1} reduced by {2} floors at zero", exhaustCost.Generic)
	}
	if plainCost.Generic != 1 {
		t.Errorf("the ordinary ability costs {%d}, want the printed {1} — the clause says exhaust abilities", plainCost.Generic)
	}
}

// TestACastModifierDoesNotPriceAnActivation is the partition, and it
// is the half that would have broken every existing card: Sphere of
// Resistance's AppliesTo is nil, meaning "every spell", and without
// the Activations bit it would have started taxing every {T} ability
// in the game.
func TestACastModifierDoesNotPriceAnActivation(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-tax", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-tax")
	costModifierCatalog(t, "probe-sphere", CostModifier{
		Kind:   CostIncrease,
		Label:  "Spells cost {1} more to cast.",
		Amount: func(CostQuery) int { return 1 },
	})
	pushExhaustSource(g, me, "probe-sphere")

	var cost ParsedCost
	g.WithWriteLock(func() {
		card := *g.findCardByIDLocked(src)
		cost, _ = g.AbilityManaCostForEffect(me.ID, card, ZoneBattlefield, ActivatedAbilitiesForCard(card)[0])
	})
	if cost.Generic != 1 {
		t.Errorf("the ability costs {%d}, want the printed {1} — a cast modifier prices casts", cost.Generic)
	}
}

// TestTheDiscountIsWhatTheActivationActuallyPays — the pass has to be
// on the PAYING path, not only in an accessor, or the engine and the
// enumerator would agree with each other and both be wrong.
func TestTheDiscountIsWhatTheActivationActuallyPays(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "probe-pay", exhaustProbe(exhaustLabelA, true))
	src := pushExhaustSource(g, me, "probe-pay")
	costModifierCatalog(t, "probe-pay-discount", CostModifier{
		Kind:        CostReduction,
		Activations: true,
		Label:       "Exhaust abilities of other permanents you control cost {2} less to activate.",
		AppliesTo: func(q CostQuery) bool {
			return q.Ability != nil && q.Ability.Exhaust && q.Card.InstanceID != q.Source.InstanceID
		},
		Amount: func(CostQuery) int { return 2 },
	})
	pushExhaustSource(g, me, "probe-pay-discount")

	// STRICT mode, which is the whole point of the assertion:
	// permissive mode waives an unpayable charge and would pass
	// whatever the pass computed.
	var insufficient *InsufficientManaError
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("strict activation with an empty pool: %v (want no error — the discount took the printed {1} to {0}); "+
			"insufficient-mana shape: %v", err, errors.As(err, &insufficient))
	}
	passBothForTest(g)
	if got := counterOf(g, src, "ran-"+exhaustLabelA); got != 1 {
		t.Errorf("the ability resolved %d times, want 1", got)
	}
}
