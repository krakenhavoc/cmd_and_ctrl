package game

import (
	"testing"

	"github.com/google/uuid"
)

// ability_own_cost_modifier_test.go — #1296, ADR 0020 amendment
// 2026-09-24: ActivatedAbilityShape.CostModifiers, the ability's OWN
// cost clause, and the targets it may read. The catalog cards that use
// it (Dragonfire Blade, Ghostfire Blade, Warrior's Blades, the channel
// lands) are pinned in cards/effects; these are the engine rules.

// ownClauseAbility is a "{4}: …" ability carrying `mods` as its own
// cost clause.
func ownClauseAbility(mods ...CostModifier) ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label:         "{4}: Probe.",
		Cost:          AbilityCost{Mana: "{4}"},
		CostModifiers: mods,
	}
}

// TestAbilityOwnClauseSeesTargetsOnlyWhenItSaysSo is ADR 0048 §13 on
// the new slot: the announced targets reach a modifier that declares
// ReadsTargets and nobody else, so a clause that forgot to declare it
// gets the same wrong answer everywhere rather than a price the
// enumerator and the engine disagree on.
func TestAbilityOwnClauseSeesTargetsOnlyWhenItSaysSo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	target := []TargetRef{{Kind: TargetCard, ID: uuid.New()}}
	var sawDeclared, sawUndeclared int
	ab := ownClauseAbility(
		CostModifier{
			Kind: CostReduction, Label: "declared", ReadsTargets: true,
			Amount: func(q CostQuery) int { sawDeclared = len(q.Targets); return len(q.Targets) },
		},
		CostModifier{
			Kind: CostReduction, Label: "undeclared",
			Amount: func(q CostQuery) int { sawUndeclared = len(q.Targets); return 0 },
		},
	)
	var cost ParsedCost
	var err error
	g.WithWriteLock(func() {
		cost, err = g.AbilityManaCostForTargetsForEffect(me.ID, NewCard("Probe", me.ID), ZoneBattlefield, ab, target)
	})
	if err != nil {
		t.Fatal(err)
	}
	if sawDeclared != 1 || sawUndeclared != 0 {
		t.Errorf("targets seen: declared %d, undeclared %d — want 1 and 0", sawDeclared, sawUndeclared)
	}
	if cost.Generic != 3 {
		t.Errorf("{4} less one per target = {%d}, want {3}", cost.Generic)
	}
}

// TestAbilityOwnClauseIsBoundToTheActivator: the clause is the
// ability's, so its source is the ability's source with the ACTIVATOR
// as its controller (CR 602.2) — a hand card, which has no controller
// of its own, still reads "you" as the player activating it.
func TestAbilityOwnClauseIsBoundToTheActivator(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	handCard := NewCard("Channel Probe", me.ID)
	handCard.Controller = uuid.Nil
	var sourceController uuid.UUID
	ab := ownClauseAbility(CostModifier{
		Kind: CostReduction, Label: "probe",
		Amount: func(q CostQuery) int { sourceController = q.Source.Controller; return 1 },
	})
	g.WithWriteLock(func() {
		_, _ = g.AbilityManaCostForEffect(me.ID, handCard, ZoneHand, ab)
	})
	if sourceController != me.ID {
		t.Errorf("the clause's source controller = %s, want the activator %s", sourceController, me.ID)
	}
}

// TestAbilityOwnClausePricesOnlyItsOwnAbility: two abilities on one
// source, one with a clause. The other is untouched — the slot, not a
// predicate, is what scopes it.
func TestAbilityOwnClausePricesOnlyItsOwnAbility(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	discounted := ownClauseAbility(CostModifier{
		Kind: CostReduction, Label: "{2} less", Amount: func(CostQuery) int { return 2 },
	})
	plain := ownClauseAbility()
	src := NewCard("Probe", me.ID)
	var a, b ParsedCost
	g.WithWriteLock(func() {
		a, _ = g.AbilityManaCostForEffect(me.ID, src, ZoneBattlefield, discounted)
		b, _ = g.AbilityManaCostForEffect(me.ID, src, ZoneBattlefield, plain)
	})
	if a.Generic != 2 || b.Generic != 4 {
		t.Errorf("discounted {%d}, plain {%d}; want {2} and {4}", a.Generic, b.Generic)
	}
}

// TestAbilityPriceReadsTargets is the enumerator's and the view's
// question: true for an own clause that reads targets, true for an
// activation-scoped board modifier that does, false otherwise — and
// false for a cost with no mana, which the pass never prices.
func TestAbilityPriceReadsTargets(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	reads := ownClauseAbility(CostModifier{Kind: CostReduction, ReadsTargets: true})
	plain := ownClauseAbility()
	noMana := ActivatedAbilityShape{
		Label: "{T}: Probe.", Cost: AbilityCost{Tap: true},
		CostModifiers: []CostModifier{{Kind: CostReduction, ReadsTargets: true}},
	}
	var gotReads, gotPlain, gotNoMana bool
	g.WithWriteLock(func() {
		gotReads = g.AbilityPriceReadsTargetsForEffect(reads)
		gotPlain = g.AbilityPriceReadsTargetsForEffect(plain)
		gotNoMana = g.AbilityPriceReadsTargetsForEffect(noMana)
	})
	if !gotReads || gotPlain || gotNoMana {
		t.Errorf("reads=%v plain=%v noMana=%v, want true false false", gotReads, gotPlain, gotNoMana)
	}

	// A board modifier that prices activations by target makes every
	// activated ability with a mana cost target-priced.
	costModifierCatalog(t, "probe-board", CostModifier{
		Kind: CostReduction, Activations: true, ReadsTargets: true,
	})
	pushExhaustSource(g, me, "probe-board")
	g.WithWriteLock(func() { gotPlain = g.AbilityPriceReadsTargetsForEffect(plain) })
	if !gotPlain {
		t.Error("an activation-scoped board modifier that reads targets was not seen")
	}
}

// TestActivationPaysThePriceOfTheTargetItNamed drives
// ActivateCatalogAbility end to end on an instance-carried ability:
// the clause takes {1} off per target colour, the target is two
// colours, and exactly {2} leaves the pool.
func TestActivationPaysThePriceOfTheTargetItNamed(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	host := NewCard("Host", me.ID)
	host.TypeLine = "Creature — Bear"
	host.Colors = []string{"U", "R"}
	host.Controller = me.ID
	g.Battlefield.PushTop(host)

	src := NewCard("Probe Blade", me.ID)
	src.TypeLine = "Artifact"
	src.Controller = me.ID
	src.ActivatedAbilities = []ActivatedAbilityShape{{
		Label:   "{4}: Probe.",
		Cost:    AbilityCost{Mana: "{4}"},
		Targets: enchantCreatureSpec(),
		CostModifiers: []CostModifier{{
			Kind: CostReduction, Label: "{1} less per colour of the target", ReadsTargets: true,
			Amount: func(q CostQuery) int {
				for _, tr := range q.Targets {
					if c, ok := q.Game.LookupCardForEffect(tr.ID); ok {
						return len(c.EffectiveColors())
					}
				}
				return 0
			},
		}},
		Effect: func(*Game, *StackItem) error { return nil },
	}}
	g.Battlefield.PushTop(src)

	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"}, ManaToken{Color: "C"})
	err := g.ActivateCatalogAbility(me.ID, src.InstanceID, 0, ActivateAbilityParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: host.InstanceID}},
		Strict:  true,
	})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("pool after = %d, want 1 — the activation charges {4} − 2 colours = {2}", got)
	}
}
