package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// discard_x_cost_test.go — #6145: "As an additional cost to cast this
// spell, discard X cards" (game.AdditionalCost.DiscardX). Scorched Earth,
// Firestorm and Nahiri's Wrath belong to batches not yet built, so their
// shapes are registered here under test oracle ids.

const (
	scorchedShapeOracle  = "test-6145-scorched-earth"
	firestormShapeOracle = "test-6145-firestorm"
	wrathShapeOracle     = "test-6145-nahiris-wrath"
	tideShapeOracle      = "test-6145-aether-tide"
)

func discardXMain(t *testing.T, g *game.Game, me *game.Player, name, oracle, typeLine string) uuid.UUID {
	t.Helper()
	me.Hand.Cards = nil
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{R}",
		OracleID: oracle, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.FloatMana(t, g, me, "{R}")
	return id
}

// Scorched Earth: "discard X land cards. Destroy X target lands." X=2
// takes exactly two land discards and two land targets, and destroys both.
func TestDiscardXLandCardsDestroysXTargetLands(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: scorchedShapeOracle, Name: "Scorched Shape", Completeness: effects.CompletenessFull,
		AdditionalCost: effects.DiscardXLandCost(),
		Targets:        effects.VsXTarget("X target lands", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.VsDestroyTheTargets(ctx)
		},
	})
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spell := discardXMain(t, g, me, "Scorched Shape", scorchedShapeOracle, "Sorcery")
	lands := []uuid.UUID{
		effectstest.HandCard(me, "Mountain", "Basic Land — Mountain"),
		effectstest.HandCard(me, "Forest", "Basic Land — Forest"),
	}
	nonland := effectstest.HandCard(me, "Shock", "Instant")
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: effectstest.AcPermanent(g, opp.ID, "Island", "Basic Land — Island", 0, 0)},
		{Kind: game.TargetCard, ID: effectstest.AcPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp", 0, 0)},
	}

	// A nonland card does not pay it, and the target count must be X.
	for name, params := range map[string]game.CastSpellParams{
		"a nonland discard":  {XValue: 2, DiscardIDs: []uuid.UUID{lands[0], nonland}, Targets: targets},
		"one target at X=2":  {XValue: 2, DiscardIDs: lands, Targets: targets[:1]},
		"one discard at X=2": {XValue: 2, DiscardIDs: lands[:1], Targets: targets},
	} {
		params.Strict = true
		if err := g.CastSpell(me.ID, spell, params); err == nil {
			t.Fatalf("%s: the cast was accepted", name)
		}
		if g.Stack.Contains(spell) {
			t.Fatalf("%s: a refused cast reached the stack", name)
		}
	}

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: lands, Targets: targets}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, tr := range targets {
		if !opp.Graveyard.Contains(tr.ID) {
			t.Errorf("target land %s was not destroyed", tr.ID)
		}
	}
	if !me.Hand.Contains(nonland) {
		t.Error("the nonland card left the hand")
	}
}

// Firestorm at X=0: castable with no discard and no targets (CR 107.3a: 0
// is a legal announcement), and it does nothing.
func TestDiscardXAtZeroIsCastableWithNoTargets(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: firestormShapeOracle, Name: "Firestorm Shape", Completeness: effects.CompletenessFull,
		AdditionalCost: effects.DiscardXCost(),
		Targets:        effects.VsXTarget("X target creatures", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, tr := range ctx.LegalTargets() {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: tr.ID, Amount: ctx.X()}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := discardXMain(t, g, me, "Firestorm Shape", firestormShapeOracle, "Instant")
	kept := effectstest.HandCard(me, "Filler", "Instant")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell at X=0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) || !me.Hand.Contains(kept) {
		t.Error("X=0 should resolve and discard nothing")
	}
}

// Nahiri's Wrath: "damage equal to the total mana value of the discarded
// cards" reads the discards from PaidCost.Discarded after they are in the
// graveyard (CR 608.2h).
func TestDiscardXRecordsTheDiscardedCardsForResolution(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: wrathShapeOracle, Name: "Wrath Shape", Completeness: effects.CompletenessFull,
		AdditionalCost: effects.DiscardXCost(),
		Targets:        effects.VsXTarget("X target creatures", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			total := 0
			for _, id := range ctx.Discarded() {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok {
					continue
				}
				if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
					continue
				}
				if mv, ok := ctx.Game.ManaValueForEffect(c); ok {
					total += mv
				}
			}
			for _, tr := range ctx.LegalTargets() {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: tr.ID, Amount: total}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spell := discardXMain(t, g, me, "Wrath Shape", wrathShapeOracle, "Sorcery")
	var discards []uuid.UUID
	for _, cost := range []string{"{2}{R}", "{1}"} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Pitched", TypeLine: "Instant", ManaCost: cost, Owner: me.ID, Controller: me.ID})
		discards = append(discards, id)
	}
	giant := effectstest.AcPermanent(g, opp.ID, "Giant", "Creature — Giant", 5, 9)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, XValue: 2, DiscardIDs: discards,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: giant},
			{Kind: game.TargetCard, ID: effectstest.AcPermanent(g, opp.ID, "Bear", "Creature — Bear", 2, 2)},
		},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// {2}{R} and {1}: a total mana value of 4.
	if got := effectstest.DamageMarkedOn(g, giant); got != 4 {
		t.Errorf("damage = %d, want the discarded cards' total mana value 4", got)
	}
}

// Aether Tide: "discard X creature cards" — only creature cards pay it,
// and the view offers exactly those, with the ceiling X may reach.
func TestDiscardXCreatureCardsOffersOnlyCreatures(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: tideShapeOracle, Name: "Tide Shape", Completeness: effects.CompletenessFull,
		AdditionalCost: effects.DiscardXCreatureCost(),
		Targets:        effects.VsXTarget("X target creatures", effects.Creature()),
		OnResolve:      func(*game.StackItem, *effects.Context) error { return nil },
	})
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := discardXMain(t, g, me, "Tide Shape", tideShapeOracle, "Instant")
	bear := effectstest.HandCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.HandCardOf(g, me, "Shock", "Instant", "{R}", "")

	v := protocol.ViewOfGameFor(g, me.ID.String())
	var ac *protocol.AdditionalCostView
	for _, s := range v.Seats {
		for _, c := range s.Hand.Cards {
			if c.InstanceID == spell.String() {
				ac = c.AdditionalCost
			}
		}
	}
	if ac == nil || !ac.DiscardX || !ac.DemandsX || ac.DiscardXMax != 1 {
		t.Fatalf("view additional cost = %+v, want discard_x, demands_x, max 1", ac)
	}
	if ac.DiscardOptions == nil || len(ac.DiscardOptions.Cards) != 1 || ac.DiscardOptions.Cards[0] != bear.String() {
		t.Fatalf("discard options = %+v, want only the creature card", ac.DiscardOptions)
	}
}
