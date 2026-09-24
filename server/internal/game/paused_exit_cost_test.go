package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// paused_exit_cost_test.go — #1445 (and #1415), the battlefield leg.
//
// Two halves:
//
//  1. The case as reported: a commander with "{1}, Sacrifice this"
//     activated twice while the first activation's CR 903.9 prompt was
//     open, paying both with one card. #1423 (#1397) fixed that by
//     asking BEFORE paying; the regression test below pins it in that
//     model.
//  2. #1445, what #1423 leaves: a commander whose exit an EFFECT has paused
//     (destroyed, with its owner still deciding) sits on the
//     battlefield, and no cost may spend it in that window.
//     refusePausedCostCardsLocked (cost_commander_choice.go) is the
//     gate; the hand and graveyard legs are in
//     paused_exit_hand_cost_test.go and the proof cards in
//     cards/effects/paused_exit_cost_test.go.

// sacrificeSelfDrawAbility is "{1}, Sacrifice this: draw a card".
func sacrificeSelfDrawAbility() ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label: "{1}, Sacrifice this: draw a card",
		Cost:  AbilityCost{Mana: "{1}", SacrificeSelf: true},
		Effect: func(g *Game, item *StackItem) error {
			return g.DrawNForEffect(item.Controller, 1)
		},
	}
}

// editBattlefieldCard applies `fn` to the battlefield card `id`.
func editBattlefieldCard(g *Game, id uuid.UUID, fn func(*Card)) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			fn(&g.Battlefield.Cards[i])
		}
	}
}

// pushSacrificeOutlet seats an artifact with "Sacrifice a creature:
// mark" — a stack-using sacrifice-another outlet.
func pushSacrificeOutlet(g *Game, owner *Player) uuid.UUID {
	c := NewCard("Sacrifice Outlet", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "Sacrifice a creature: mark",
		Cost:  AbilityCost{SacrificeOther: creatureCostSpec()},
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// destroyCommanderPaused destroys `id` by effect and asserts the
// premise every test here rests on: its owner is asked, and it is
// still on the battlefield while they decide.
func destroyCommanderPaused(t *testing.T, g *Game, owner *Player, id uuid.UUID) *PendingChoice {
	t.Helper()
	if err := g.DestroyPermanentForEffect(id); err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the destroyed commander left before its owner answered — the premise of #1445 is gone")
	}
	return prompt
}

// #1415 as reported, in #1423's ask-first model. Two activations of
// "{1}, Sacrifice this" each park a question and pay nothing; the
// first answer pays once and puts one ability on the stack; the second
// answer finds the commander gone and pays nothing more. Both answers
// to the first question are checked.
func TestSacrificeSelfCommanderPaysOnceHoweverOftenItIsActivated(t *testing.T) {
	for _, takeCommandZone := range []bool{true, false} {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		me := g.Seats[0]
		id := seatCommander(t, g.Battlefield, me)
		editBattlefieldCard(g, id, func(c *Card) { c.ActivatedAbilities = []ActivatedAbilityShape{sacrificeSelfDrawAbility()} })
		me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})

		for i := 0; i < 2; i++ {
			if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{}); err != nil {
				t.Fatalf("activation %d: %v", i+1, err)
			}
		}
		if len(g.StackMeta) != 0 || len(me.ManaPool) != 2 {
			t.Fatalf("stack %d, pool %d before any answer — nothing is paid before the owner answers", len(g.StackMeta), len(me.ManaPool))
		}
		if len(g.PendingChoices) != 2 {
			t.Fatalf("%d pending choices, want one parked question per activation", len(g.PendingChoices))
		}
		first, second := g.PendingChoices[0].ID, g.PendingChoices[1].ID
		if err := g.ResolveOptionalReplacement(first, me.ID, takeCommandZone); err != nil {
			t.Fatalf("first answer: %v", err)
		}
		// The payer is the one answering, so the refusal of the stale
		// re-run comes back to them; what matters is that it paid
		// nothing.
		_ = g.ResolveOptionalReplacement(second, me.ID, takeCommandZone)

		if len(g.StackMeta) != 1 {
			t.Errorf("%d stack items, want the ONE ability the one card paid for", len(g.StackMeta))
		}
		if len(me.ManaPool) != 1 {
			t.Errorf("pool = %v, want one {1} spent", me.ManaPool)
		}
		if takeCommandZone {
			assertOnlyIn(t, id, me.Command, g.Battlefield, me.Graveyard)
		} else {
			assertOnlyIn(t, id, me.Graveyard, g.Battlefield, me.Command)
		}
	}
}

// A destroyed commander, its owner still deciding, cannot pay any
// battlefield cost: not its own sacrifice-this or exile-this, not a
// sacrifice outlet,
// not Ashnod's Altar, not its own sacrifice-this mana ability, not a
// return-to-hand cost. Each refusal pays nothing and leaves the
// destroy's prompt alone. An unrelated creature still pays while the
// prompt is open, and after the answer the commander has gone where
// the answer sent it — the destroy, not a cost, took it.
func TestADestroyedCommanderCannotPayABattlefieldCostWhileAsked(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	outlet := pushSacrificeOutlet(g, me)
	altar := pushIntrinsicPermanent(g, me, "Ashnod's Altar", "Artifact", altarAbility(), nil)
	bounce := pushReturnCostSource(g, me, AbilityCost{ReturnToHand: &ReturnToHandCost{
		Count: 1, Filter: creatureCostSpec(), Label: "a creature you control",
	}})
	cmdr := seatCommander(t, g.Battlefield, me)
	editBattlefieldCard(g, cmdr, func(c *Card) {
		c.ActivatedAbilities = []ActivatedAbilityShape{
			sacrificeSelfDrawAbility(),
			// #1404's "Exile this <permanent>:" — no Zones, so it
			// functions from the battlefield.
			{Label: "Exile this: draw", Cost: AbilityCost{ExileSelf: true}, Effect: func(g *Game, item *StackItem) error {
				return g.DrawNForEffect(item.Controller, 1)
			}},
		}
		c.ManaAbilities = spawnAbility()
	})
	bear := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)
	wolf := pushIntrinsicPermanent(g, me, "Wolf", "Creature — Wolf", nil, nil)
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	prompt := destroyCommanderPaused(t, g, me, cmdr)

	for _, tc := range []struct {
		name string
		try  func() error
	}{
		{"its own sacrifice-this ability", func() error {
			return g.ActivateCatalogAbility(me.ID, cmdr, 0, ActivateAbilityParams{})
		}},
		{"its own exile-this ability", func() error {
			return g.ActivateCatalogAbility(me.ID, cmdr, 1, ActivateAbilityParams{})
		}},
		{"a sacrifice outlet", func() error {
			return g.ActivateCatalogAbility(me.ID, outlet, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{cmdr}})
		}},
		{"Ashnod's Altar", func() error {
			return g.ActivateManaAbility(me.ID, altar, 0, ManaAbilityParams{SacrificeIDs: []uuid.UUID{cmdr}})
		}},
		{"its own sacrifice-this mana ability", func() error {
			return g.ActivateManaAbility(me.ID, cmdr, 0, ManaAbilityParams{})
		}},
		{"a return-to-hand cost", func() error {
			return g.ActivateCatalogAbility(me.ID, bounce, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{cmdr}})
		}},
	} {
		if err := tc.try(); !errors.Is(err, ErrChoicePending) {
			t.Errorf("%s naming the destroyed commander: err = %v, want ErrChoicePending", tc.name, err)
		}
		if !g.Battlefield.Contains(cmdr) {
			t.Fatalf("%s moved the destroyed commander", tc.name)
		}
		if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != prompt.ID {
			t.Fatalf("%s disturbed the destroy's prompt (%d pending)", tc.name, len(g.PendingChoices))
		}
	}
	if len(g.StackMeta) != 0 || len(me.ManaPool) != 1 {
		t.Fatalf("stack %d, pool %v — a refused payment paid something", len(g.StackMeta), me.ManaPool)
	}

	// The guard is about the card, not the table.
	if err := g.ActivateCatalogAbility(me.ID, outlet, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("the outlet eating an unrelated creature while the prompt is open: %v", err)
	}

	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdr, me.Graveyard, g.Battlefield, me.Command, me.Hand)
	if hasEvent(g, EventSacrifice, cmdr) {
		t.Error("the destroyed commander was also sacrificed")
	}
	if err := g.ActivateManaAbility(me.ID, altar, 0, ManaAbilityParams{SacrificeIDs: []uuid.UUID{wolf}}); err != nil {
		t.Fatalf("the Altar after the answer: %v", err)
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("pool = %d mana, want the {C} plus the Wolf's {C}{C}", len(me.ManaPool))
	}
}

// A non-commander is untouched: destroying it never pauses, so there is
// nothing to refuse, and its own sacrifice-this ability works as it
// always did.
func TestSacrificeCostsOnANonCommanderAreUnaffected(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	doomed := pushIntrinsicPermanent(g, me, "Doomed Bear", "Creature — Bear", nil, nil)
	if err := g.DestroyPermanentForEffect(doomed); err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
	if n := len(g.PendingChoices); n != 0 {
		t.Fatalf("%d pending choices after destroying a non-commander, want none", n)
	}
	assertOnlyIn(t, doomed, me.Graveyard, g.Battlefield)

	id := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)
	editBattlefieldCard(g, id, func(c *Card) { c.ActivatedAbilities = []ActivatedAbilityShape{sacrificeSelfDrawAbility()} })
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(g.PendingChoices) != 0 || len(g.StackMeta) != 1 {
		t.Fatalf("pending %d, stack %d — want no prompt and the one ability", len(g.PendingChoices), len(g.StackMeta))
	}
	assertOnlyIn(t, id, me.Graveyard, g.Battlefield)
}

// The auto-tapper never activates through ActivateManaAbility, so the
// announcement gate does not reach it: it plans a destroyed commander's
// "Sacrifice this: Add {C}" unless it asks the same question itself.
// Checked through the planner and through a real auto-tapped
// activation, which must fail short rather than sacrifice the card.
func TestAutoTapSkipsADestroyedCommandersSacrificeAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cmdr := seatCommander(t, g.Battlefield, me)
	editBattlefieldCard(g, cmdr, func(c *Card) { c.ManaAbilities = spawnAbility() })
	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0); !ok || !containsID(plan, cmdr) {
		t.Fatalf("premise: the commander's sacrifice ability should pay {1} before anything happens (plan %v, ok %v)", plan, ok)
	}

	prompt := destroyCommanderPaused(t, g, me, cmdr)

	// A plan made before the destroy arrives stale; the executor must
	// drop the source rather than sacrifice it.
	g.WithWriteLock(func() { g.materializePlanLocked(me, tapPlan{{CardID: cmdr}}, costFor(t, "{1}")) })
	if len(me.ManaPool) != 0 || hasEvent(g, EventSacrifice, cmdr) {
		t.Fatalf("a stale plan spent the destroyed commander (pool %v)", me.ManaPool)
	}

	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0); ok {
		t.Errorf("auto-tap planned %v with the destroyed commander as the only source", plan)
	}
	src := pushReturnCostSource(g, me, AbilityCost{Mana: "{1}"})
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{AutoTap: true}); err == nil {
		t.Error("an auto-tapped {1} activation was paid with the destroyed commander")
	}
	if !g.Battlefield.Contains(cmdr) || len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != prompt.ID {
		t.Fatal("the auto-tapper moved the commander or disturbed its prompt")
	}
}
