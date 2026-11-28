package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b853_cards_test.go — Fallen Empires, counters, batch 853 (#5637,
// S82 OLD). Every test plays the real catalog entry through the engine.

const (
	oldB853EbonPraetor = "0501af1f-c67c-4cb1-aced-4be426c83316"
	oldB853ThornT      = "1239d3bf-06e8-4732-b1e2-f6eb145e5894"
	oldB853FeralT      = "3953fd17-03f4-4431-99b5-87dcac5a279d"
	oldB853Thallid     = "46abaabe-7015-4d82-a5aa-2706b1f51bed"
	oldB853Farmer      = "91fe1075-e170-4575-ae28-8d692f051637"
	oldB853Moneychgr   = "9b81aa32-e0f7-493f-92b8-85c2c46aef91"
	oldB853Devourer    = "b7326fd9-7648-4db3-b506-61a7de8a4094"
	oldB853SporeFlower = "f352ec62-bf09-45f2-b5b7-75931177b15f"
	oldB853Bloom       = "fa76ee16-ff48-418b-819a-69f1139b36f3"
)

func TestOldB853CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB853EbonPraetor: "Ebon Praetor", oldB853ThornT: "Thorn Thallid", oldB853FeralT: "Feral Thallid",
		oldB853Thallid: "Thallid", oldB853Farmer: "Elvish Farmer", oldB853Moneychgr: "Icatian Moneychanger",
		oldB853Devourer: "Thallid Devourer", oldB853SporeFlower: "Spore Flower", oldB853Bloom: "Fungal Bloom",
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

func oldB853Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB853ThallidsGrowASporeCounterEachOfYourUpkeeps(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Thallid", "Creature — Fungus", oldB853Thallid},
		{"Elvish Farmer", "Creature — Elf", oldB853Farmer},
		{"Thallid Devourer", "Creature — Fungus", oldB853Devourer},
		{"Thorn Thallid", "Creature — Fungus", oldB853ThornT},
		{"Feral Thallid", "Creature — Fungus", oldB853FeralT},
		{"Spore Flower", "Creature — Fungus", oldB853SporeFlower},
	} {
		g, me, opp := oldB853Table(t)
		id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		effectstest.BeginUpkeep(t, g, opp.ID)
		if n := effectstest.Pr8Counters(g, id, "spore"); n != 0 {
			t.Errorf("%s: %d spore counters in an opponent's upkeep, want 0", tc.name, n)
		}
		effectstest.BeginUpkeep(t, g, me.ID)
		if n := effectstest.Pr8Counters(g, id, "spore"); n != 1 {
			t.Errorf("%s: %d spore counters after your upkeep, want 1", tc.name, n)
		}
	}
}

func TestOldB853ThallidMakesASaprolingForThreeSpores(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Thallid", "Creature — Fungus", oldB853Thallid},
		{"Elvish Farmer", "Creature — Elf", oldB853Farmer},
		{"Thallid Devourer", "Creature — Fungus", oldB853Devourer},
	} {
		g, me, _ := oldB853Table(t)
		id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		effectstest.PfAddCounters(g, id, "spore", 2)
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Fatalf("%s: activated with two spore counters", tc.name)
		}
		effectstest.PfAddCounters(g, id, "spore", 1)
		effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true})
		if n := effectstest.Pr8Counters(g, id, "spore"); n != 0 {
			t.Errorf("%s: %d spore counters left, want 0", tc.name, n)
		}
		if n := effectstest.CountBattlefieldByName(g, "Saproling"); n != 1 {
			t.Fatalf("%s: %d Saprolings, want 1", tc.name, n)
		}
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Saproling" && (c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "G" || !c.IsCreature()) {
				t.Errorf("%s: the token is %d/%d colors %v, want a 1/1 green creature", tc.name, c.Power, c.Toughness, c.Colors)
			}
		}
	}
}

func TestOldB853ElvishFarmerSacrificesASaprolingForTwoLife(t *testing.T) {
	g, me, _ := oldB853Table(t)
	farmer := effectstest.PushCatalogPermanent(g, me.ID, "Elvish Farmer", "Creature — Elf", oldB853Farmer, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, farmer, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a non-Saproling was sacrificed")
	}
	sap := effectstest.PushCatalogPermanent(g, me.ID, "Saproling", "Token Creature — Saproling", "", false)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.P7Activate(t, g, me, farmer, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{sap}})
	if g.Battlefield.Contains(sap) {
		t.Error("the Saproling was not sacrificed")
	}
	if got := effectstest.LifeOf(g, me.ID); got != before+2 {
		t.Errorf("life %d, want %d", got, before+2)
	}
}

func TestOldB853ThallidDevourerPumpsBySacrificingASaproling(t *testing.T) {
	g, me, _ := oldB853Table(t)
	dev := effectstest.PushCatalogPermanent(g, me.ID, "Thallid Devourer", "Creature — Fungus", oldB853Devourer, false)
	sap := effectstest.PushCatalogPermanent(g, me.ID, "Saproling", "Token Creature — Saproling", "", false)
	bp, bt := effectstest.EffectivePTOf(t, g, dev)
	effectstest.P7Activate(t, g, me, dev, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{sap}})
	if p, tough := effectstest.EffectivePTOf(t, g, dev); p != bp+1 || tough != bt+2 {
		t.Errorf("the Devourer is %d/%d, want %d/%d", p, tough, bp+1, bt+2)
	}
}

func TestOldB853ThornThallidPingsAnyTarget(t *testing.T) {
	g, me, opp := oldB853Table(t)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Thorn Thallid", "Creature — Fungus", oldB853ThornT, false)
	effectstest.PfAddCounters(g, id, "spore", 3)
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Errorf("opponent life %d, want %d", got, before-1)
	}
	if n := effectstest.Pr8Counters(g, id, "spore"); n != 0 {
		t.Errorf("%d spore counters left, want 0", n)
	}
}

func TestOldB853FeralThallidRegeneratesForThreeSpores(t *testing.T) {
	g, me, _ := oldB853Table(t)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Feral Thallid", "Creature — Fungus", oldB853FeralT, false)
	effectstest.PfAddCounters(g, id, "spore", 3)
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.DestroyIt(t, g, id)
	if !g.Battlefield.Contains(id) {
		t.Error("the Thallid was destroyed despite its regeneration shield")
	}
}

func TestOldB853SporeFlowerPreventsCombatDamage(t *testing.T) {
	g, me, opp := oldB853Table(t)
	flower := effectstest.PushCatalogPermanent(g, me.ID, "Spore Flower", "Creature — Fungus", oldB853SporeFlower, false)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	defender := effectstest.B12Creature(g, me.ID, "Defender", "Creature — Bear", 2, 2)
	effectstest.PfAddCounters(g, flower, "spore", 3)
	effectstest.P7Activate(t, g, me, flower, 0, game.ActivateAbilityParams{Strict: true})
	if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := effectstest.DamageMarkedOn(g, defender); got != 0 {
		t.Errorf("combat damage marked = %d, want 0", got)
	}
	if err := g.MarkDamage(defender, 1); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	if got := effectstest.DamageMarkedOn(g, defender); got != 1 {
		t.Errorf("non-combat damage marked = %d, want 1", got)
	}
}

func TestOldB853FungalBloomPutsASporeOnAFungusOnly(t *testing.T) {
	g, me, _ := oldB853Table(t)
	bloom := effectstest.PushCatalogPermanent(g, me.ID, "Fungal Bloom", "Enchantment", oldB853Bloom, false)
	fungus := effectstest.B12Creature(g, me.ID, "A Fungus", "Creature — Fungus", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 1, 1)
	effectstest.B06AddMana(me, "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, bloom, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("a non-Fungus was targeted")
	}
	effectstest.P7Activate(t, g, me, bloom, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(fungus)})
	if n := effectstest.Pr8Counters(g, fungus, "spore"); n != 1 {
		t.Errorf("%d spore counters on the Fungus, want 1", n)
	}
}

func TestOldB853IcatianMoneychangerEntersWithCreditsAndHurtsYou(t *testing.T) {
	g, me, _ := oldB853Table(t)
	before := effectstest.LifeOf(g, me.ID)
	cardID := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cardID, Name: "Icatian Moneychanger", TypeLine: "Creature — Human", OracleID: oldB853Moneychgr, Power: 0, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, cardID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Icatian Moneychanger" {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("the Moneychanger did not enter")
	}
	if n := effectstest.Pr8Counters(g, id, "credit"); n != 3 {
		t.Errorf("%d credit counters, want 3", n)
	}
	if got := effectstest.LifeOf(g, me.ID); got != before-3 {
		t.Errorf("life %d, want %d", got, before-3)
	}
}

func TestOldB853IcatianMoneychangerSacrificesForLifeOnlyInYourUpkeep(t *testing.T) {
	g, me, _ := oldB853Table(t)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Icatian Moneychanger", "Creature — Human", oldB853Moneychgr, false)
	effectstest.PfAddCounters(g, id, "credit", 4)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("sacrificed in the main phase")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	id2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Icatian Moneychanger", "Creature — Human", oldB853Moneychgr, false)
	effectstest.PfAddCounters(g2, id2, "credit", 4)
	effectstest.AdvanceToUpkeepOfSeat(t, g2, 0)
	before := effectstest.LifeOf(g2, me2.ID)
	effectstest.P7Activate(t, g2, me2, id2, 0, game.ActivateAbilityParams{Strict: true})
	if g2.Battlefield.Contains(id2) {
		t.Error("the Moneychanger was not sacrificed")
	}
	if got := effectstest.LifeOf(g2, me2.ID); got < before+4 {
		t.Errorf("life %d, want at least %d", got, before+4)
	}
}

func TestOldB853EbonPraetorShrinksEachUpkeepAndFeedsOnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	praetor := effectstest.B12Push(g, me.ID, "Ebon Praetor", "Creature — Avatar Praetor", oldB853EbonPraetor, 5, 5)
	thrull := effectstest.B12Creature(g, me.ID, "A Thrull", "Creature — Thrull", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 1, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.Pr8Counters(g, praetor, "-2/-2"); n != 1 {
		t.Fatalf("%d -2/-2 counters after upkeep, want 1", n)
	}
	// Sacrificing a Bear removes the counter and adds nothing.
	effectstest.P7Activate(t, g, me, praetor, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}})
	if n := effectstest.Pr8Counters(g, praetor, "-2/-2"); n != 0 {
		t.Errorf("%d -2/-2 counters, want 0", n)
	}
	if n := effectstest.Pr8Counters(g, praetor, "+1/+0"); n != 0 {
		t.Errorf("%d +1/+0 counters after a Bear, want 0", n)
	}
	// Only once each turn.
	if err := g.ActivateCatalogAbility(me.ID, praetor, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{thrull}}); err == nil {
		t.Fatal("activated twice in one turn")
	}
}

func TestOldB853EbonPraetorGetsAPlusOneCounterFromAThrull(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	praetor := effectstest.B12Push(g, me.ID, "Ebon Praetor", "Creature — Avatar Praetor", oldB853EbonPraetor, 5, 5)
	thrull := effectstest.B12Creature(g, me.ID, "A Thrull", "Creature — Thrull", 1, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P7Activate(t, g, me, praetor, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{thrull}})
	if n := effectstest.Pr8Counters(g, praetor, "+1/+0"); n != 1 {
		t.Errorf("%d +1/+0 counters, want 1", n)
	}
	if n := effectstest.Pr8Counters(g, praetor, "-2/-2"); n != 0 {
		t.Errorf("%d -2/-2 counters, want 0", n)
	}
}

func TestOldB853EbonPraetorCannotFeedOutsideYourUpkeep(t *testing.T) {
	g, me, _ := oldB853Table(t)
	praetor := effectstest.B12Push(g, me.ID, "Ebon Praetor", "Creature — Avatar Praetor", oldB853EbonPraetor, 5, 5)
	bear := effectstest.B12Creature(g, me.ID, "A Bear", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, praetor, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("activated in the main phase")
	}
}
