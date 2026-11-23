package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b19_cards_test.go — Mystery Booster 2 creatures, GAP batch 19 (#5948).

const (
	gapMarduOutrider = "9bc5ed59-ae13-40df-94b9-c56c32673857"
	gapSengirVampire = "749141aa-f6c4-4ad8-b146-406e68ae9b0b"
	gapStreetWraith  = "bd7867f4-4f58-49c2-9c24-a124dfdb506d"
)

func TestS82GapB19EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapMarduOutrider: "Mardu Outrider", gapSengirVampire: "Sengir Vampire", gapStreetWraith: "Street Wraith",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82GapB19MarduOutriderNeedsAnotherCardToDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "B")
	// Alone in hand: nothing to discard, so a strict cast is refused.
	me.Hand.Cards = nil
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Mardu Outrider", "Creature — Orc Warrior", gapMarduOutrider, nil); err == nil {
		t.Fatal("cast with no card to discard")
	}
}

func TestS82GapB19MarduOutriderDiscardsACardAsItIsCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "B")
	fodder := effectstest.PushChromeMoxHandCard(me, "Fodder", "Instant", []string{"R"})
	outrider := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: outrider, Name: "Mardu Outrider", TypeLine: "Creature — Orc Warrior",
		OracleID: gapMarduOutrider, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, outrider, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast without paying the discard")
	}
	if err := g.CastSpell(me.ID, outrider, game.CastSpellParams{Strict: true, DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast with the discard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the additional cost should have discarded the other card")
	}
	if !g.Battlefield.Contains(outrider) {
		t.Error("the Outrider should have resolved")
	}
}

func TestS82GapB19SengirVampireGrowsWhenAWoundedCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sv := effectstest.B12Push(g, me.ID, "Sengir Vampire", "Creature — Vampire", gapSengirVampire, 4, 4)
	victim := effectstest.B16Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2, "G")
	bystander := effectstest.B16Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(sv, victim, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	effectstest.KillCreature(t, g, opp.ID, bystander)
	effectstest.PassPriorityAroundTable(t, g)
	if got := r33GapCountersOf(t, g, sv); got != 0 {
		t.Errorf("a creature this never damaged died: counters = %d, want 0", got)
	}
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := r33GapCountersOf(t, g, sv); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
}

func r33GapCountersOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s left the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}

func TestS82GapB19StreetWraithCyclesForTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Street Wraith", TypeLine: "Creature — Wraith", OracleID: gapStreetWraith,
		ManaCost: "{3}{B}{B}", Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceToMain(t, g)
	life, hand := me.Life, effectstest.HandSize(me)
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the Wraith should be discarded")
	}
	if got := effectstest.HandSize(me); got != hand {
		t.Errorf("hand %d -> %d, want the same (discarded one, drew one)", hand, got)
	}
}
