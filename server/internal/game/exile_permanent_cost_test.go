package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exile_permanent_cost_test.go — #1600: ExilePermanentsCost, "Exile a
// creature you control" as a cost component, on both of its owners: a
// CR 602 ability (The Soul Stone's harness, Altar of Bhaal) and a
// CR 605 mana ability (Food Chain). The catalog half is
// cards/effects/exile_permanent_cost_cards_test.go, which also pins the
// trigger behaviour with real cards; the CR 903.9 rows are in
// cost_commander_choice_test.go's table.
//
// What is pinned here is the ENGINE contract: paid at announce and in
// full, refused when the payment is wrong with nothing moved (CR 118.3),
// the move is an exile and not a sacrifice, and the auto-tapper never
// spends a mana ability that carries the component.

// creatureExileCost is "Exile a creature you control" — the in-package
// stand-in for effects.ExileACreatureYouControl.
func creatureExileCost() *ExilePermanentsCost {
	return &ExilePermanentsCost{Count: 1, CardType: "creature", Label: "a creature you control"}
}

// pushExileCostSource seats an artifact whose one ability costs `cost`
// and stamps a marker counter on itself when it resolves.
func pushExileCostSource(g *Game, owner *Player, cost AbilityCost) uuid.UUID {
	c := NewCard("Exile Cost Source", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "exile a creature: mark",
		Cost:  cost,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// pushExileCostCreature seats a vanilla creature `controller` controls.
func pushExileCostCreature(g *Game, controller *Player, name string) uuid.UUID {
	c := NewCard(name, controller.ID)
	c.TypeLine = "Creature — Bear"
	c.Power, c.Toughness = 2, 2
	c.Controller = controller.ID
	c.ManaCost = "{1}{G}"
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// foodChainShape is Food Chain's mana ability with a fixed output, so a
// test reads the pool without a colour pick.
func foodChainShape() ManaAbilityShape {
	return ManaAbilityShape{
		ExilePermanents: creatureExileCost(),
		Produced:        "{G}{G}",
		Label:           "Exile a creature you control: Add {G}{G}",
	}
}

// The headline: the named creature is in exile at announce, it was not
// sacrificed, the ability is on the stack, it resolves afterwards, and
// the paid-cost record names the creature.
func TestExilePermanentCostPaysAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushExileCostSource(g, me, AbilityCost{ExilePermanents: creatureExileCost()})
	bear := pushExileCostCreature(g, me, "Bear")
	keep := pushExileCostCreature(g, me, "Other Bear")

	before := len(g.Events)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ExilePermanentIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the named creature is not in exile — the cost is paid at announce")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("the unnamed creature was exiled too")
	}
	ltb, sac := 0, 0
	for _, ev := range g.Events[before:] {
		if ev.CardID != bear {
			continue
		}
		switch ev.Kind {
		case EventLTB:
			ltb++
		case EventSacrifice:
			sac++
		}
	}
	if ltb != 1 {
		t.Errorf("%d EventLTB for the exiled creature, want exactly one", ltb)
	}
	if sac != 0 {
		t.Errorf("%d EventSacrifice for the exiled creature — exiling is not sacrificing", sac)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta has %d items, want 1", len(g.StackMeta))
	}
	for _, item := range g.StackMeta {
		if len(item.Paid.Exiled) != 1 || item.Paid.Exiled[0] != bear {
			t.Errorf("Paid.Exiled = %v, want the exiled creature", item.Paid.Exiled)
		}
	}
	if counterOf(g, src, "effect-ran") != 0 {
		t.Error("the effect ran at announce")
	}
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability did not resolve")
	}
}

// Every way the component is unpayable, and that none of them moves or
// taps anything. CR 118.3: a cost is not partially payable.
func TestExilePermanentCostRejectsWhatCannotPay(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	// {T} as well, so a refused activation that tapped anything shows.
	src := pushExileCostSource(g, me, AbilityCost{Tap: true, ExilePermanents: creatureExileCost()})
	mine := pushExileCostCreature(g, me, "My Bear")
	other := pushExileCostCreature(g, me, "My Other Bear")
	theirs := pushExileCostCreature(g, opp, "Their Bear")
	land := NewCard("Forest", me.ID)
	land.TypeLine = "Basic Land — Forest"
	land.Controller = me.ID
	g.Battlefield.PushTop(land)

	cases := []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"nothing named", nil, ErrInvalidParam},
		{"two named for a one-permanent clause", []uuid.UUID{mine, other}, ErrInvalidParam},
		{"the same creature named twice", []uuid.UUID{mine, mine}, ErrInvalidParam},
		{"a card that does not exist", []uuid.UUID{uuid.New()}, ErrCardNotFound},
		{"an opponent's creature", []uuid.UUID{theirs}, ErrCardCallerMismatch},
		{"a permanent the clause does not admit", []uuid.UUID{land.InstanceID}, ErrIllegalTarget},
	}
	for _, tc := range cases {
		err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ExilePermanentIDs: tc.ids})
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	for _, id := range []uuid.UUID{mine, other, theirs, land.InstanceID} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a refused activation moved %v", id)
		}
	}
	if c := findBattlefieldCard(g, src); c == nil || c.Tapped {
		t.Error("a refused activation tapped the source")
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("StackMeta has %d items, want 0 — every activation was refused", len(g.StackMeta))
	}
}

// IDs for an ability with no exile-a-permanent component are refused
// rather than ignored, as an unexpected sacrifice_ids is.
func TestExilePermanentCostRefusesIDsForAnAbilityWithoutOne(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushExileCostSource(g, me, AbilityCost{})
	bear := pushExileCostCreature(g, me, "Bear")

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ExilePermanentIDs: []uuid.UUID{bear}}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a refused activation exiled the creature anyway")
	}
}

// CR 118.3: one permanent pays one component. A creature named to the
// sacrifice cannot also be the one exiled, and the refusal leaves both
// on the battlefield.
func TestExilePermanentCostRefusesAPermanentTheSacrificeAlsoNames(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushExileCostSource(g, me, AbilityCost{SacrificeOther: creatureCostSpec(), ExilePermanents: creatureExileCost()})
	bear := pushExileCostCreature(g, me, "Bear")
	other := pushExileCostCreature(g, me, "Other Bear")

	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		SacrificeIDs:      []uuid.UUID{bear},
		ExilePermanentIDs: []uuid.UUID{bear},
	})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam", err)
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(other) {
		t.Fatal("a refused activation moved a creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		SacrificeIDs:      []uuid.UUID{bear},
		ExilePermanentIDs: []uuid.UUID{other},
	}); err != nil {
		t.Fatalf("two different creatures: %v", err)
	}
	if !me.Graveyard.Contains(bear) || !g.Exile.Contains(other) {
		t.Error("the sacrifice and the exile did not each take their own creature")
	}
}

// The shared walk is the engine's: the options list and the payability
// predicate see only the activator's creatures, and with none the
// ability is not activatable at all (#544).
func TestExilePermanentCostOptionsAreTheActivatorsMatches(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushExileCostSource(g, me, AbilityCost{ExilePermanents: creatureExileCost()})
	pushExileCostCreature(g, opp, "Their Bear")

	g.mu.RLock()
	opts := g.ExilePermanentsOptionsForEffect(me.ID, src, creatureExileCost())
	payable := g.ExilePermanentsPayable(me.ID, src, creatureExileCost())
	g.mu.RUnlock()
	if len(opts) != 0 || payable {
		t.Fatalf("options %v, payable %v with no creature of mine — want none and false", opts, payable)
	}
	mine := pushExileCostCreature(g, me, "My Bear")
	g.mu.RLock()
	opts = g.ExilePermanentsOptionsForEffect(me.ID, src, creatureExileCost())
	payable = g.ExilePermanentsPayable(me.ID, src, creatureExileCost())
	g.mu.RUnlock()
	if len(opts) != 1 || opts[0] != mine || !payable {
		t.Errorf("options %v, payable %v — want exactly my creature, payable", opts, payable)
	}
}

// The mana-ability owner: the creature is exiled, the mana is in the
// pool, nothing was sacrificed — and the exiled creature is on the
// paid-cost record ProducedForPaid reads.
func TestExilePermanentCostOnAManaAbility(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var seen PaidCost
	ab := foodChainShape()
	ab.ProducedForPaid = func(_ *Game, _, _ uuid.UUID, paid PaidCost) string {
		seen = paid
		return "{G}{G}"
	}
	src := manaSource(g, me, ab)
	bear := pushExileCostCreature(g, me, "Bear")

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{ExilePermanentIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !g.Exile.Contains(bear) {
		t.Error("the creature was not exiled")
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %d, want 2", len(me.ManaPool))
	}
	if hasEvent(g, EventSacrifice, bear) {
		t.Error("EventSacrifice for an exiled creature")
	}
	if len(seen.Exiled) != 1 || seen.Exiled[0] != bear {
		t.Errorf("ProducedForPaid saw Exiled = %v, want the exiled creature", seen.Exiled)
	}
	g.mu.RLock()
	info, ok := g.LastKnownPermanentForEffect(bear)
	g.mu.RUnlock()
	if !ok || info.ManaValue != 2 {
		t.Errorf("last-known mana value = %d (ok %v), want 2 — {1}{G}", info.ManaValue, ok)
	}
}

// A refused mana activation pays nothing: no mana, no exile.
func TestExilePermanentCostOnAManaAbilityRefusesWithNothingPaid(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := manaSource(g, me, foodChainShape())
	theirs := pushExileCostCreature(g, opp, "Their Bear")

	for _, ids := range [][]uuid.UUID{nil, {theirs}, {uuid.New()}} {
		if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{ExilePermanentIDs: ids}); err == nil {
			t.Errorf("ids %v: accepted, want refused", ids)
		}
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %d after refusals, want 0", len(me.ManaPool))
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("a refused activation exiled the opponent's creature")
	}
}

// The auto-tapper never spends a mana ability with this component:
// which creature to exile is a decision, and a cast paid by eating the
// player's board would be the planner deciding it. The shape carries a
// {T} as well, because a mana ability with neither {T} nor a sacrifice
// (Food Chain itself) is outside the planner's reach for that reason
// alone, and the exile clause has to be the thing refused.
func TestAutoTapNeverExilesAPermanentForMana(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	tapAndExile := foodChainShape()
	tapAndExile.TapCost = true
	src := manaSource(g, me, tapAndExile)
	bear := pushExileCostCreature(g, me, "Bear")

	g.mu.RLock()
	card := findBattlefieldCard(g, src)
	accepted := g.autoTapAbilityAccepts(me.ID, *card, tapAndExile)
	g.mu.RUnlock()
	if accepted {
		t.Fatal("autoTapAbilityAccepts takes an exile-a-permanent mana ability")
	}
	shape := foodChainShape()
	shape.OncePerTurn = true
	if autoTapFreeOncePerTurn(shape) {
		t.Error("the costless-once-per-turn tier takes an exile-a-permanent mana ability")
	}

	spell := NewCard("Grizzly Bears", me.ID)
	spell.TypeLine = "Creature — Bear"
	spell.ManaCost = "{G}"
	me.Hand.PushTop(spell)
	advanceTo(t, g, StepPrecombatMain)
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		t.Error("the cast was paid — the auto-tapper exiled a creature for mana")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the auto-tapper exiled the creature")
	}
}
