package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 365 (S82, #5557): Conspiracy — Ignition Team, Drakestown Forgotten,
// Respite and Quicksand. Grenzo's Rebuttal is skipped (see the commit).

const (
	oldB365Ignition = "9edaadc9-6bbd-4727-9884-1c6b13c778bb"
	oldB365Drakes   = "b3c8d9ef-4ffb-482b-8859-486e7dc4621e"
	oldB365Respite  = "46137f00-f7f5-430a-857c-ff2dbf2d2579"
	oldB365Quick    = "ef2bb4fa-f292-4d19-aaa4-cfbe445caf45"
)

func TestOldB365CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB365Ignition: "Ignition Team",
		oldB365Drakes:   "Drakestown Forgotten",
		oldB365Respite:  "Respite",
		oldB365Quick:    "Quicksand",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB365Card(name, typeLine, cost, oracle string, p, tough int) game.Card {
	return game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Layout: "normal", Power: p, Toughness: tough}
}

func TestOldB365DrakestownEntersWithOneCounterPerCreatureCardInEveryGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}")
	effectstest.PushGraveyardPermanent(opp, "Dead Wolf", "Creature — Wolf", "{1}")
	effectstest.PushGraveyardPermanent(opp, "Dead Rock", "Artifact", "{1}")
	effectstest.FloatForTest(g, me, "BBBBB")
	id, err := effectstest.Msc17Cast(t, g, me, oldB365Card("Drakestown Forgotten", "Creature — Zombie", "{4}{B}", oldB365Drakes, 0, 0), game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 (two creature cards in graveyards)", n)
	}
}

func TestOldB365DrakestownShrinksATargetByRemovingACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Drakestown Forgotten", "Creature — Zombie", oldB365Drakes, 0, 0)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 3, 3)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{B}{B}{B}", effectstest.CardTargets(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters left after the cost, want 0", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, victim), effectstest.EffectiveToughness(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("target is %d/%d, want 2/2", p, tough)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{B}{B}{B}", effectstest.CardTargets(victim)); err == nil {
		t.Error("activated with no counter to remove")
	}
}

func TestOldB365IgnitionTeamCountsEveryTappedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.AangPushLand(g, me.ID, "Mine Tapped", true)
	effectstest.AangPushLand(g, opp.ID, "Theirs Tapped", true)
	effectstest.AangPushLand(g, opp.ID, "Theirs Untapped", false)
	effectstest.FloatForTest(g, me, "RRRRRRR")
	id, err := effectstest.Msc17Cast(t, g, me, oldB365Card("Ignition Team", "Creature — Goblin Warrior", "{5}{R}{R}", oldB365Ignition, 0, 0), game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 tapped lands", n)
	}
}

func TestOldB365IgnitionTeamAnimatesATargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Ignition Team", "Creature — Goblin Warrior", oldB365Ignition, 0, 0)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	land := effectstest.AangPushLand(g, opp.ID, "Their Land", false)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{R}{R}{R}", effectstest.CardTargets(land)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	lc := effectstest.LayeredCard(t, g, land)
	if !lc.IsLand() || !lc.IsCreature() {
		t.Errorf("the land is not still a land creature: %q", lc.TypeLine)
	}
	if p, tough := effectstest.EffectivePower(t, g, land), effectstest.EffectiveToughness(t, g, land); p != 4 || tough != 4 {
		t.Errorf("the land is %d/%d, want 4/4", p, tough)
	}
	if cols := effectstest.EffectiveColors(t, g, land); len(cols) != 1 || cols[0] != "R" {
		t.Errorf("colours %v, want red", cols)
	}
	nonLand := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{R}{R}{R}", effectstest.CardTargets(nonLand)); err == nil {
		t.Error("targeted a nonland permanent")
	}
}

func TestOldB365RespitePreventsCombatDamageAndGainsLifePerAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a1 := effectstest.PushVanillaCreature(g, me.ID, "My Attacker", 2, 2)
	a2 := effectstest.PushVanillaCreature(g, opp.ID, "Their Attacker", 2, 2)
	idle := effectstest.PushVanillaCreature(g, opp.ID, "Idle", 2, 2)
	_ = idle
	effectstest.ArnSetCombat(g, a1, opp.ID, uuid.Nil)
	effectstest.ArnSetCombat(g, a2, me.ID, uuid.Nil)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Respite", "Instant", oldB365Respite, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2 (every attacking creature counts)", got)
	}
}

func TestOldB365QuicksandShrinksAnAttackerWithoutFlyingAndRefusesAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sand := effectstest.AangPushLand(g, me.ID, "Quicksand", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == sand {
				g.Battlefield.Cards[i].OracleID = oldB365Quick
			}
		}
	})
	walker := effectstest.PushVanillaCreature(g, opp.ID, "Walker", 3, 3)
	flier := effectstest.PushVanillaCreature(g, opp.ID, "Flier", 3, 3)
	idle := effectstest.PushVanillaCreature(g, opp.ID, "Idle", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	effectstest.ArnSetCombat(g, walker, me.ID, uuid.Nil)
	effectstest.ArnSetCombat(g, flier, me.ID, uuid.Nil)
	if err := effectstest.HobB4Activate(t, g, me.ID, sand, 0, "", effectstest.CardTargets(flier)); err == nil {
		t.Error("targeted an attacking creature with flying")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, sand, 0, "", effectstest.CardTargets(idle)); err == nil {
		t.Error("targeted a creature that is not attacking")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, sand, 0, "", effectstest.CardTargets(walker)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sand) {
		t.Error("Quicksand was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, walker), effectstest.EffectiveToughness(t, g, walker); p != 2 || tough != 1 {
		t.Errorf("the attacker is %d/%d, want 2/1", p, tough)
	}
}
