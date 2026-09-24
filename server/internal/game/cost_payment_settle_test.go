package game

import (
	"testing"

	"github.com/google/uuid"
)

// costMoveBecomes is one mandatory replacement for a particular card's
// requested destination. Registering two different instances makes the
// affected player order them under CR 616.1 unless the move is part of the
// indivisible cost-payment step (CR 601.2h / 602.2b).
func costMoveBecomes(card uuid.UUID, from, to ZoneKind, owner uuid.UUID, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == card && ev.NewZone == from
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.NewZone = to
			ev.NewZoneOwner = owner
			return nil
		},
		Label: label,
	}
}

func TestSacrificeCostSettlesReplacementOrderBeforeActivation(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	source := abilitySource(g, me, AbilityCost{SacrificeOther: creatureCostSpec()})
	fodder := seedCostCard(g.Battlefield, me.ID, me.ID, false, nil)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(costMoveBecomes(fodder, ZoneGraveyard, ZoneExile,
			uuid.Nil, "exile it instead"))
		g.RegisterReplacementForTest(costMoveBecomes(fodder, ZoneGraveyard, ZoneLibrary,
			me.ID, "put it into its owner's library instead"))
	})

	params := ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}
	if err := g.ActivateCatalogAbility(me.ID, source, 0, params); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("cost payment queued %d choice(s); CR 602.2b must settle it inline", len(g.PendingChoices))
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("stack items = %d, want the paid-for ability", len(g.StackMeta))
	}
	if !g.Exile.Contains(fodder) {
		t.Error("the first gathered replacement did not settle the sacrificed card into exile")
	}
	if err := g.ActivateCatalogAbility(me.ID, source, 0, params); err == nil {
		t.Error("the sacrificed card paid the same cost twice")
	}
	if len(g.StackMeta) != 1 {
		t.Errorf("stack items after the refused re-use = %d, want 1", len(g.StackMeta))
	}
}

func TestNonCostSacrificeStillOffersReplacementOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fodder := seedCostCard(g.Battlefield, me.ID, me.ID, false, nil)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(costMoveBecomes(fodder, ZoneGraveyard, ZoneExile,
			uuid.Nil, "exile it instead"))
		g.RegisterReplacementForTest(costMoveBecomes(fodder, ZoneGraveyard, ZoneLibrary,
			me.ID, "put it into its owner's library instead"))
	})

	if err := g.SacrificePermanent(me.ID, fodder); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("non-cost sacrifice queued %d choices, want its CR 616 ordering prompt", len(g.PendingChoices))
	}
	if g.PendingChoices[0].Kind != PendingChoiceReplacementOrder {
		t.Errorf("choice kind = %q, want %q", g.PendingChoices[0].Kind, PendingChoiceReplacementOrder)
	}
	if findBattlefieldCard(g, fodder) == nil {
		t.Error("the permanent moved before its replacement-order choice was answered")
	}
}

func TestAlternativeCostSettlesReplacementOrderBeforeCast(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	const oracle = "test-1420-pitch"
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "pitch", Label: "Exile a blue card", ExileFromHand: blueCardSpec(),
	}))

	spell := NewCard("Test Force", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	me.Hand.PushTop(spell)
	pitch := seedCostCard(me.Hand, me.ID, me.ID, false, func(c *Card) {
		c.Colors = []string{"U"}
	})

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(costMoveBecomes(pitch, ZoneExile, ZoneGraveyard,
			me.ID, "put it into its owner's graveyard instead"))
		g.RegisterReplacementForTest(costMoveBecomes(pitch, ZoneExile, ZoneLibrary,
			me.ID, "put it into its owner's library instead"))
	})

	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		AlternativeCost: "pitch",
		AltCostIDs:      []uuid.UUID{pitch},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("alternative-cost payment queued %d choice(s); CR 601.2h must settle it inline", len(g.PendingChoices))
	}
	if !g.Stack.Contains(spell.InstanceID) {
		t.Error("the spell did not reach the stack after its alternative cost settled")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the first gathered replacement did not settle the pitched card into the graveyard")
	}
}
