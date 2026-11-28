package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b764_cards_test.go — Tempest Auras, batch 764 (S82 older sets, #5628).
// Every test plays the real catalog entry through the engine.

const (
	oldB764HerosResolve  = "1854a99d-f8c7-45b3-83a2-98ae1c5b5b09"
	oldB764SadisticGlee  = "2af22962-8f66-4868-a175-ad5d7f4c9d58"
	oldB764Enfeeblement  = "42b2db4c-4a1d-436f-9eeb-53a04db46c58"
	oldB764StealEnchant  = "567a3fa7-5ca9-4812-b81d-4953c401c83d"
	oldB764TahngarthRage = "5b0cb089-087d-4b91-9e63-8186dc737a00"
	oldB764SpinalGraft   = "70b5eca0-c61b-4066-ae5f-90948912c04e"
	oldB764FrogTongue    = "861c5374-a8e7-4684-9a1b-e6ce3b98e4a2"
	oldB764EndlessScream = "cde3e29c-31fd-4dce-a008-283bbe90bfc6"
	oldB764GiantStrength = "db85ba13-f00d-4cdd-99e1-22a4d39c8837"
)

func TestOldB764EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB764HerosResolve, oldB764SadisticGlee, oldB764Enfeeblement, oldB764StealEnchant,
		oldB764TahngarthRage, oldB764SpinalGraft, oldB764FrogTongue, oldB764EndlessScream, oldB764GiantStrength} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB764CastAura casts the named Aura from the active seat's hand at `target`
// and settles the stack. It returns the Aura's permanent id (zero when it did not
// resolve onto the battlefield) and the cast error.
func oldB764CastAura(t *testing.T, g *game.Game, name, oracle, cost string, x int, target uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Enchantment — Aura", ManaCost: cost, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	if cost != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}{B}"); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		return uuid.Nil, err
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		return uuid.Nil, nil
	}
	return id, nil
}

func TestOldB764PumpAurasChangeThePowerAndToughnessOfTheirHost(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Hero's Resolve", oldB764HerosResolve, 3, 7},
		{"Giant Strength", oldB764GiantStrength, 4, 4},
		{"Enfeeblement", oldB764Enfeeblement, 0, 0}, // see the next test for the dying 2/2
	} {
		if tc.name == "Enfeeblement" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			opp := g.Seats[1]
			bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
			if _, err := oldB764CastAura(t, g, tc.name, tc.oracle, "", 0, land); err == nil {
				t.Fatal("enchanted a land")
			}
			aura, err := oldB764CastAura(t, g, tc.name, tc.oracle, "", 0, bear)
			if err != nil || aura == uuid.Nil {
				t.Fatalf("cast: %v", err)
			}
			if p, th := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != tc.power || th != tc.tough {
				t.Errorf("the creature is %d/%d, want %d/%d", p, th, tc.power, tc.tough)
			}
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
			if p := effectstest.EffectivePower(t, g, bear); p != 2 {
				t.Errorf("power %d after the Aura left, want 2", p)
			}
		})
	}
}

func TestOldB764EnfeeblementShrinksAndKills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Their Big", "Creature — Ogre", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Their Small", "Creature — Bear", 2, 2)
	if aura, err := oldB764CastAura(t, g, "Enfeeblement", oldB764Enfeeblement, "", 0, big); err != nil || aura == uuid.Nil {
		t.Fatalf("cast on the big one: %v", err)
	}
	if p, th := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 2 || th != 2 {
		t.Errorf("the 4/4 is %d/%d, want 2/2", p, th)
	}
	if _, err := oldB764CastAura(t, g, "Enfeeblement", oldB764Enfeeblement, "", 0, small); err != nil {
		t.Fatalf("cast on the small one: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived -2/-2")
	}
}

func TestOldB764StealEnchantmentTakesControlWhileItIsAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := oldB764CastAura(t, g, "Steal Enchantment", oldB764StealEnchant, "", 0, bear); err == nil {
		t.Fatal("enchanted a creature")
	}
	aura, err := oldB764CastAura(t, g, "Steal Enchantment", oldB764StealEnchant, "", 0, ench)
	if err != nil || aura == uuid.Nil {
		t.Fatalf("cast: %v", err)
	}
	if c, ok := effectstest.BattlefieldCard(g, ench); !ok || c.Controller != me.ID {
		t.Fatalf("you do not control the enchantment: %v %v", ok, c.Controller)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {}) // the layers are recomputed on a read
	if c, ok := effectstest.BattlefieldCard(g, ench); !ok || c.Controller != opp.ID {
		t.Error("control did not go back when the Aura left")
	}
}

func TestOldB764FrogTongueDrawsAndGivesReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "L1", "L2", "L3")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "reach") {
		t.Fatal("the Bear has reach already")
	}
	aura, err := oldB764CastAura(t, g, "Frog Tongue", oldB764FrogTongue, "", 0, bear)
	if err != nil || aura == uuid.Nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d→%d, want a card drawn (the Aura left the hand to be cast)", hand, me.Hand.Size())
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "reach") {
		t.Error("the enchanted creature has no reach")
	}
}

func TestOldB764SadisticGleePutsACounterOnTheHostWheneverACreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "My Host", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Victim", "Creature — Bear", 1, 1)
	aura, err := oldB764CastAura(t, g, "Sadistic Glee", oldB764SadisticGlee, "", 0, host)
	if err != nil || aura == uuid.Nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.D5681Dies(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, host, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d +1/+1 counters on the host after one death, want 1", n)
	}
	other := effectstest.B12Creature(g, me.ID, "My Other", "Creature — Bear", 1, 1)
	effectstest.D5681Dies(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, host, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters after a second death, want 2 (your own creatures count)", n)
	}
}

func TestOldB764TahngarthsRageIsBigWhileAttackingAndSmallOtherwise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "My Host", "Creature — Bear", 3, 3)
	aura, err := oldB764CastAura(t, g, "Tahngarth's Rage", oldB764TahngarthRage, "", 0, host)
	if err != nil || aura == uuid.Nil {
		t.Fatalf("cast: %v", err)
	}
	if p, th := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 1 || th != 2 {
		t.Fatalf("not attacking: %d/%d, want 1/2", p, th)
	}
	effectstest.BrAttack(t, g, host)
	if p, th := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 6 || th != 3 {
		t.Errorf("attacking: %d/%d, want 6/3", p, th)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if p, th := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 1 || th != 2 {
		t.Errorf("after combat: %d/%d, want 1/2", p, th)
	}
}

func TestOldB764SpinalGraftPumpsAndDestroysTheCreatureThatBecomesATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	aura, err := oldB764CastAura(t, g, "Spinal Graft", oldB764SpinalGraft, "", 0, host)
	if err != nil || aura == uuid.Nil {
		t.Fatalf("cast: %v", err)
	}
	// Casting the Aura itself targeted the creature before the Aura was on the battlefield.
	if !g.Battlefield.Contains(host) {
		t.Fatal("the creature died when the Aura was cast on it")
	}
	if p, th := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 7 || th != 7 {
		t.Fatalf("the creature is %d/%d, want 7/7", p, th)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == host {
				g.Battlefield.Cards[i].RegenerationShields = 1
			}
		}
	})
	// Even a spell that cannot kill a 7/7 destroys it, shield or no shield.
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: host})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature survived becoming a target")
	}
	_ = me
}

func TestOldB764EndlessScreamEntersWithXCountersAndPumpsPowerPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "My Host", "Creature — Bear", 2, 2)
	aura, err := oldB764CastAura(t, g, "Endless Scream", oldB764EndlessScream, "{X}{B}", 3, host)
	if err != nil || aura == uuid.Nil {
		t.Fatalf("cast: %v", err)
	}
	if n := effectstest.B12Counter(t, g, aura, "scream"); n != 3 {
		t.Fatalf("%d scream counters, want 3", n)
	}
	if p, th := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 5 || th != 2 {
		t.Errorf("the creature is %d/%d, want 5/2", p, th)
	}
	// The pump follows the Aura's counters.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(aura, "scream", -2) })
	if p := effectstest.EffectivePower(t, g, host); p != 3 {
		t.Errorf("power %d after two counters came off, want 3", p)
	}
}
