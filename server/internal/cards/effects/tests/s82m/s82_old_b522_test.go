package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eventide counters cards (S82 older sets, batch 522): every test plays the
// real catalog entry through the engine.

const (
	oldB522CrumblingAshes = "2e4030eb-c74f-4f81-8d35-8f1d080b15cf"
	oldB522Razorfin       = "38a2ded4-77c5-4d2b-b411-7dca8108aa02"
	oldB522Canker         = "4440e41c-042f-4e13-925a-c1e40b77e010"
	oldB522GilderBairn    = "a4441c7d-e22a-4e19-8365-34d52da9fe86"
	oldB522PyrrhicRevival = "e8daf155-1fc9-4a0e-a7b5-69a50bc23576"
	oldB522HotheadedGiant = "f90b2cc6-9edb-4666-849a-7c1f6d669daa"
)

func TestOldB522EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB522CrumblingAshes: "Crumbling Ashes", oldB522Razorfin: "Razorfin Abolisher",
		oldB522Canker: "Canker Abomination", oldB522GilderBairn: "Gilder Bairn",
		oldB522PyrrhicRevival: "Pyrrhic Revival", oldB522HotheadedGiant: "Hotheaded Giant",
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

func oldB522Creature(g *game.Game, owner uuid.UUID, name string, p, tough int, counters map[string]int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", Power: p, Toughness: tough,
		Owner: owner, Controller: owner, Counters: counters,
	})
}

func TestOldB522CrumblingAshesDestroysACreatureWithAMinusCounterEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	effectstest.PushCatalogPermanent(g, me.ID, "Crumbling Ashes", "Enchantment", oldB522CrumblingAshes, false)
	marked := oldB522Creature(g, opp.ID, "Marked", 3, 3, map[string]int{game.CounterMinusOne: 1})
	plusOne := oldB522Creature(g, opp.ID, "PlusOne", 3, 3, map[string]int{game.CounterPlusOne: 1})
	plain := oldB522Creature(g, opp.ID, "Plain", 3, 3, nil)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PlsB1Settle(t, g, func(c *game.PendingChoice) game.TargetRef {
		if len(c.PickTargetCards) != 1 || c.PickTargetCards[0] != marked {
			t.Errorf("offered %v, want only the creature with a -1/-1 counter", c.PickTargetCards)
		}
		return game.TargetRef{Kind: game.TargetCard, ID: marked}
	}, nil)
	if g.Battlefield.Contains(marked) {
		t.Error("the creature with a -1/-1 counter survived")
	}
	if !g.Battlefield.Contains(plusOne) || !g.Battlefield.Contains(plain) {
		t.Error("a creature without a -1/-1 counter was destroyed")
	}
}

func TestOldB522RazorfinAbolisherBouncesACreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	abolisher := effectstest.PushCatalogPermanent(g, me.ID, "Razorfin Abolisher", "Creature — Merfolk Wizard", oldB522Razorfin, false)
	marked := oldB522Creature(g, opp.ID, "Marked", 3, 3, map[string]int{"shield": 1})
	plain := oldB522Creature(g, opp.ID, "Plain", 3, 3, nil)
	effectstest.ApaMana(me, "U", "U")
	if err := effectstest.InvB1Activate(g, me.ID, abolisher, 0, effectstest.CardTargets(plain)...); err == nil {
		t.Fatal("a creature with no counter was a legal target")
	}
	effectstest.ApaActivate(t, g, me, abolisher, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(marked)})
	if g.Battlefield.Contains(marked) || !opp.Hand.Contains(marked) {
		t.Error("the creature with a counter did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(plain) {
		t.Error("the plain creature left the battlefield")
	}
}

func TestOldB522GilderBairnDoublesEachKindOfCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bairn := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gilder Bairn", TypeLine: "Creature — Ouphe", OracleID: oldB522GilderBairn,
		Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	target := oldB522Creature(g, opp.ID, "Target", 3, 3, map[string]int{game.CounterPlusOne: 2, "shield": 1})
	effectstest.ApaMana(me, "G", "G", "G")
	effectstest.ApaActivate(t, g, me, bairn, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(target)})
	c := effectstest.FindCardAnywhere(t, g, target)
	if c.Counters[game.CounterPlusOne] != 4 || c.Counters["shield"] != 2 {
		t.Errorf("counters %v, want +1/+1 ×4 and shield ×2", c.Counters)
	}
	if effectstest.FindCardAnywhere(t, g, bairn).Tapped {
		t.Error("the Bairn is still tapped: {Q} untaps it")
	}
}

func TestOldB522GilderBairnNeedsToBeTappedForQ(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bairn := oldB522Creature(g, me.ID, "Gilder Bairn", 1, 3, nil)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == bairn {
			g.Battlefield.Cards[i].OracleID = oldB522GilderBairn
		}
	}
	target := oldB522Creature(g, me.ID, "Target", 3, 3, map[string]int{game.CounterPlusOne: 1})
	effectstest.ApaMana(me, "G", "G", "G")
	if err := effectstest.InvB1Activate(g, me.ID, bairn, 0, effectstest.CardTargets(target)...); err == nil {
		t.Error("an untapped Bairn paid {Q}")
	}
}

func TestOldB522CankerAbominationEntersWithACounterPerCreatureTheChosenOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	oldB522Creature(g, a.ID, "A1", 1, 1, nil)
	oldB522Creature(g, a.ID, "A2", 1, 1, nil)
	oldB522Creature(g, a.ID, "A3", 1, 1, nil)
	oldB522Creature(g, b.ID, "B1", 1, 1, nil)
	oldB522Creature(g, me.ID, "Mine", 1, 1, nil)
	canker := effectstest.PlsB1Cast(t, g, "Canker Abomination", "Creature — Treefolk Horror", oldB522Canker, "{2}{B/G}{B/G}", 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if choice == nil {
		t.Fatal("no opponent was asked for")
	}
	idx := -1
	for i, o := range choice.PickOptions {
		if o.Player == me.ID {
			t.Error("the caster was offered as an opponent")
		}
		if o.Player == a.ID {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("the opponent with three creatures was not offered")
	}
	if err := g.ResolveOptionPick(choice.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindCardAnywhere(t, g, canker)
	if c.Counters[game.CounterMinusOne] != 3 {
		t.Errorf("%d -1/-1 counters, want 3", c.Counters[game.CounterMinusOne])
	}
	if p, tough := c.PowerForComparison(), c.CurrentToughness(); p != 3 || tough != 3 {
		t.Errorf("Canker Abomination is %d/%d, want 3/3", p, tough)
	}
}

func oldB522CastRed(t *testing.T, g *game.Game, name string) {
	t.Helper()
	effectstest.PlsB1Cast(t, g, name, "Instant", "", "{R}", 0, 0)
	effectstest.PlsB1Settle(t, g, nil, nil)
}

func TestOldB522HotheadedGiantEntersWithCountersUnlessYouCastAnotherRedSpell(t *testing.T) {
	// Alone: its own cast does not count as "another".
	g := effectstest.NewCatalogGame(t)
	giant := effectstest.PlsB1Cast(t, g, "Hotheaded Giant", "Creature — Giant Warrior", oldB522HotheadedGiant, "{3}{R}", 4, 4)
	effectstest.PlsB1Settle(t, g, nil, nil)
	if n := effectstest.FindCardAnywhere(t, g, giant).Counters[game.CounterMinusOne]; n != 2 {
		t.Errorf("alone it has %d -1/-1 counters, want 2", n)
	}

	// After another red spell: no counters.
	g = effectstest.NewCatalogGame(t)
	oldB522CastRed(t, g, "Red Spell")
	giant = effectstest.PlsB1Cast(t, g, "Hotheaded Giant", "Creature — Giant Warrior", oldB522HotheadedGiant, "{3}{R}", 4, 4)
	effectstest.PlsB1Settle(t, g, nil, nil)
	if n := effectstest.FindCardAnywhere(t, g, giant).Counters[game.CounterMinusOne]; n != 0 {
		t.Errorf("after a red spell it has %d -1/-1 counters, want 0", n)
	}

	// After a spell that is not red: counters.
	g = effectstest.NewCatalogGame(t)
	effectstest.PlsB1Cast(t, g, "Blue Spell", "Instant", "", "{U}", 0, 0)
	effectstest.PlsB1Settle(t, g, nil, nil)
	giant = effectstest.PlsB1Cast(t, g, "Hotheaded Giant", "Creature — Giant Warrior", oldB522HotheadedGiant, "{3}{R}", 4, 4)
	effectstest.PlsB1Settle(t, g, nil, nil)
	if n := effectstest.FindCardAnywhere(t, g, giant).Counters[game.CounterMinusOne]; n != 2 {
		t.Errorf("after a blue spell it has %d -1/-1 counters, want 2", n)
	}
}

func TestOldB522PyrrhicRevivalReturnsEveryPlayersCreaturesWithAMinusCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushGraveyardPermanent(me, "My Bear", "Creature — Bear", "{1}{G}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	relic := effectstest.PushGraveyardPermanent(me, "Relic", "Artifact", "{2}")
	effectstest.CastCatalogSpell(t, g, "Pyrrhic Revival", "Sorcery", oldB522PyrrhicRevival, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for id, owner := range map[uuid.UUID]uuid.UUID{mine: me.ID, theirs: opp.ID} {
		c := effectstest.FindCardAnywhere(t, g, id)
		if !g.Battlefield.Contains(id) || c.Controller != owner || c.Counters[game.CounterMinusOne] != 1 {
			t.Errorf("%s: on battlefield %v, controller %v, counters %v", c.Name, g.Battlefield.Contains(id), c.Controller, c.Counters)
		}
	}
	if g.Battlefield.Contains(relic) {
		t.Error("a noncreature card returned")
	}
}
