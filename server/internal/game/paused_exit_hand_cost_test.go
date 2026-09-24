package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// paused_exit_hand_cost_test.go — #1445, the hand and graveyard legs.
// The battlefield leg's reasoning one zone over: a commander an effect
// is exiling out of its owner's hand or graveyard is asked about the
// command zone, and sits where it was while its owner decides. Every
// cost that spends a card from those piles — a Force of Will pitch,
// cycling, a discard, an exile-cards cost, a Spirit Guide, scavenge,
// a delve-style alternative cost — must refuse it until the answer is
// in.

// pushCardCostSource seats an artifact whose one ability costs `cost`
// and marks itself when it resolves.
func pushCardCostSource(g *Game, owner *Player, cost AbilityCost) uuid.UUID {
	c := NewCard("Card Cost Source", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "card cost: mark",
		Cost:  cost,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// editZoneCard applies `fn` to the card `id` in `z`.
func editZoneCard(z *Zone, id uuid.UUID, fn func(*Card)) {
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			fn(&z.Cards[i])
		}
	}
}

// exileCommanderPaused exiles `id` out of `zone` by effect and asserts
// the premise: its owner is asked, and it is still in `zone`.
func exileCommanderPaused(t *testing.T, g *Game, owner *Player, zone *Zone, id uuid.UUID) *PendingChoice {
	t.Helper()
	if err := g.ExileCardForEffect(id); err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if !zone.Contains(id) {
		t.Fatalf("the commander left the %s before its owner answered — the premise of #1445 is gone", zone.Kind)
	}
	return prompt
}

// The hand leg. A blue commander in hand, with cycling and a Spirit
// Guide's mana ability of its own, is being exiled by an effect; while
// its owner decides, no payment may spend it. A different blue card
// still pitches, and the commander lands where the answer says.
func TestHandCostsCannotSpendACommanderWhoseExitIsPaused(t *testing.T) {
	const force, thrill = "test-force", "test-thrill"
	withCatalogAlternativeCosts(t, altCostFor(force, AlternativeCost{
		Key: "pitch", Label: "Pay 1 life, exile a blue card", Life: 1,
		ExileFromHand: blueCardSpec(),
	}))
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id == thrill {
			return &AdditionalCost{DiscardCards: 1, Label: "Discard a card"}
		}
		return nil
	})

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	me.Hand.Cards = nil
	cmdr := seatCommander(t, me.Hand, me)
	editZoneCard(me.Hand, cmdr, func(c *Card) {
		c.Colors = []string{"U"}
		c.ActivatedAbilities = []ActivatedAbilityShape{cyclingAbility("{0}")}
		c.ManaAbilities = spiritGuideAbility("{U}")
	})
	inHand := func(name, oracle string, colors ...string) uuid.UUID {
		c := NewCard(name, me.ID)
		c.TypeLine = "Instant"
		c.OracleID = oracle
		c.Colors = colors
		me.Hand.PushTop(c)
		return c.InstanceID
	}
	discarder := pushCardCostSource(g, me, AbilityCost{DiscardCards: &DiscardCost{N: 1, Label: "a card"}})
	exiler := pushCardCostSource(g, me, AbilityCost{ExileCards: &ExileCost{N: 1, Label: "a card", From: ZoneHand}})
	forceID := inHand("Test Force of Will", force, "U")
	thrillID := inHand("Thrill of Possibility", thrill)

	prompt := exileCommanderPaused(t, g, me, me.Hand, cmdr)

	for _, tc := range []struct {
		name string
		try  func() error
	}{
		{"a Force of Will pitch", func() error {
			return g.CastSpell(me.ID, forceID, CastSpellParams{AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{cmdr}})
		}},
		{"cycling the commander itself", func() error {
			return g.ActivateCatalogAbility(me.ID, cmdr, 0, ActivateAbilityParams{})
		}},
		{"its own exile-from-hand mana ability", func() error {
			return g.ActivateManaAbility(me.ID, cmdr, 0, ManaAbilityParams{})
		}},
		{"an activated discard cost", func() error {
			return g.ActivateCatalogAbility(me.ID, discarder, 0, ActivateAbilityParams{DiscardIDs: []uuid.UUID{cmdr}})
		}},
		{"an activated exile-from-hand cost", func() error {
			return g.ActivateCatalogAbility(me.ID, exiler, 0, ActivateAbilityParams{ExileIDs: []uuid.UUID{cmdr}})
		}},
		{"a spell's additional discard", func() error {
			return g.CastSpell(me.ID, thrillID, CastSpellParams{DiscardIDs: []uuid.UUID{cmdr}})
		}},
	} {
		if err := tc.try(); !errors.Is(err, ErrChoicePending) {
			t.Errorf("%s naming the paused commander: err = %v, want ErrChoicePending", tc.name, err)
		}
		if !me.Hand.Contains(cmdr) {
			t.Fatalf("%s moved the paused commander", tc.name)
		}
		if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != prompt.ID {
			t.Fatalf("%s disturbed the open prompt (%d pending)", tc.name, len(g.PendingChoices))
		}
	}
	if len(g.StackMeta) != 0 || len(me.ManaPool) != 0 {
		t.Fatalf("stack %d, pool %v — a refused payment paid something", len(g.StackMeta), me.ManaPool)
	}
	// The auto-tapper plans a Spirit Guide in hand; not this one.
	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{U}"), 0); ok {
		t.Errorf("auto-tap planned %v with the paused commander as the only source", plan)
	}
	// …and a stale plan naming it is dropped by the executor.
	g.WithWriteLock(func() { g.materializePlanLocked(me, tapPlan{{CardID: cmdr}}, costFor(t, "{U}")) })
	if len(me.ManaPool) != 0 || !me.Hand.Contains(cmdr) || len(g.PendingChoices) != 1 {
		t.Fatalf("a stale plan spent the paused commander (pool %v)", me.ManaPool)
	}

	// Another blue card still pays.
	spare := altCostHandCard(me, "Test Brainstorm", "U")
	if err := g.CastSpell(me.ID, forceID, CastSpellParams{AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("Force of Will pitching an unrelated blue card: %v", err)
	}

	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdr, g.Exile, me.Hand, me.Command, me.Graveyard)
}

// The graveyard leg: a commander Bojuka Bog is exiling out of its
// owner's graveyard cannot pay an exile-from-graveyard cost, a
// delve-style alternative cost, or its own scavenge-shaped exile-this.
// After the answer the same alternative cost pays with another card.
func TestGraveyardCostsCannotSpendACommanderWhoseExitIsPaused(t *testing.T) {
	const delve = "test-delve"
	withCatalogAlternativeCosts(t, altCostFor(delve, AlternativeCost{
		Key: "delve", Label: "Exile a card from your graveyard",
		ExileFromGraveyard: &TargetSpec{Zones: []ZoneKind{ZoneGraveyard}, Min: 1, Max: 1},
	}))

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cmdr := seatCommander(t, me.Graveyard, me)
	editZoneCard(me.Graveyard, cmdr, func(c *Card) { c.ActivatedAbilities = []ActivatedAbilityShape{exileSelfAbility("{0}")} })
	filler := NewCard("Filler", me.ID)
	filler.TypeLine = "Sorcery"
	me.Graveyard.PushTop(filler)
	exiler := pushCardCostSource(g, me, AbilityCost{ExileCards: &ExileCost{N: 1, Label: "a card", From: ZoneGraveyard}})
	spell := NewCard("Test Delve", me.ID)
	spell.TypeLine = "Sorcery"
	spell.OracleID = delve
	me.Hand.PushTop(spell)

	prompt := exileCommanderPaused(t, g, me, me.Graveyard, cmdr)

	for _, tc := range []struct {
		name string
		try  func() error
	}{
		{"an activated exile-from-graveyard cost", func() error {
			return g.ActivateCatalogAbility(me.ID, exiler, 0, ActivateAbilityParams{ExileIDs: []uuid.UUID{cmdr}})
		}},
		{"a delve-style alternative cost", func() error {
			return g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{AlternativeCost: "delve", AltCostIDs: []uuid.UUID{cmdr}})
		}},
		{"its own exile-this ability", func() error {
			return g.ActivateCatalogAbility(me.ID, cmdr, 0, ActivateAbilityParams{})
		}},
	} {
		if err := tc.try(); !errors.Is(err, ErrChoicePending) {
			t.Errorf("%s naming the paused commander: err = %v, want ErrChoicePending", tc.name, err)
		}
		if !me.Graveyard.Contains(cmdr) || len(g.PendingChoices) != 1 {
			t.Fatalf("%s moved the commander or disturbed its prompt", tc.name)
		}
	}

	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdr, me.Command, me.Graveyard, g.Exile)
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{AlternativeCost: "delve", AltCostIDs: []uuid.UUID{filler.InstanceID}}); err != nil {
		t.Fatalf("the alternative cost after the answer: %v", err)
	}
}
