package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b538_test.go — Shadowmoor, batch 538 (#5591).

const (
	oldB538Gloomlance = "2aa040d2-35a1-4b1d-9542-395e1b8a9fd8"
	oldB538Arsonists  = "54831951-2cd3-41bb-8326-6cfce48feb34"
	oldB538StripBare  = "957e8ded-90e7-4bc6-98c1-eb6af1fd0ce5"
	oldB538PoisonWell = "a47ae84f-dcbc-4a60-9266-c1a02f92e7f8"
	oldB538Fulminator = "bd4c46a3-b723-4b35-9061-9a1dee7cc9d8"
	oldB538Hexhunter  = "f8f450c1-f405-4ccc-aef3-f0186d1f5056"
)

func TestOldB538EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB538Gloomlance: "Gloomlance", oldB538Arsonists: "Boggart Arsonists", oldB538StripBare: "Strip Bare",
		oldB538PoisonWell: "Poison the Well", oldB538Fulminator: "Fulminator Mage", oldB538Hexhunter: "Elvish Hexhunter",
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

func oldB538Land(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return effectstest.B12Permanent(g, owner, name, typeLine)
}

func TestOldB538FulminatorMageDestroysANonbasicLandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Fulminator Mage", "Creature — Elemental Shaman", oldB538Fulminator, true)
	basic := oldB538Land(g, opp.ID, "Forest", "Basic Land — Forest")
	nonbasic := oldB538Land(g, opp.ID, "Wooded Hollow", "Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(basic)}); err == nil {
		t.Fatal("a basic land was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(nonbasic)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.OnBattlefieldByID(g, mage) {
		t.Error("the Mage was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, nonbasic) || !effectstest.OnBattlefieldByID(g, basic) {
		t.Error("the nonbasic land should be destroyed and the basic one left")
	}
}

func TestOldB538ElvishHexhunterNeedsManaAndATapAndDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Elvish Hexhunter", "Creature — Elf Shaman", oldB538Hexhunter, true)
	ench := effectstest.B12Permanent(g, opp.ID, "Pacifism-ish", "Enchantment")
	art := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, sick, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(ench)}); err == nil {
		t.Fatal("a summoning-sick Hexhunter paid its {T} cost")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, sick).SummonedThisTurn = false })
	if err := g.ActivateCatalogAbility(me.ID, sick, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(art)}); err == nil {
		t.Fatal("an artifact was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, sick, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(ench)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, ench) || effectstest.OnBattlefieldByID(g, sick) {
		t.Error("the enchantment and the sacrificed Hexhunter should both be gone")
	}
}

func TestOldB538BoggartArsonistsHasPlainswalkAndDestroysAScarecrowOrAPlains(t *testing.T) {
	spec, _ := effects.Lookup(oldB538Arsonists)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "plainswalk" {
		t.Errorf("printed keywords %v, want plainswalk", spec.PrintedKeywords)
	}
	for _, tc := range []struct {
		name, typeLine string
		ok             bool
	}{
		{"Plains", "Basic Land — Plains", true},
		{"Nonbasic Plains", "Land — Plains", true},
		{"Scarecrow", "Artifact Creature — Scarecrow", true},
		{"Forest", "Basic Land — Forest", false},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		src := effectstest.PushCatalogPermanent(g, me.ID, "Boggart Arsonists", "Creature — Goblin Rogue", oldB538Arsonists, true)
		victim := effectstest.B12Permanent(g, opp.ID, tc.name, tc.typeLine)
		effectstest.FloatMana(t, g, me, "{C}{C}{R}")
		err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)})
		if (err == nil) != tc.ok {
			t.Errorf("%s: activation error %v, want legal=%v", tc.name, err, tc.ok)
			continue
		}
		if !tc.ok {
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, victim) || effectstest.OnBattlefieldByID(g, src) {
			t.Errorf("%s: the target and the sacrificed Arsonists should both be gone", tc.name)
		}
	}
}

func TestOldB538StripBareDestroysAurasAndEquipmentOnTheTargetOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	aura := effectstest.B12Permanent(g, me.ID, "Some Aura", "Enchantment — Aura")
	equip := effectstest.B12Permanent(g, opp.ID, "Some Blade", "Artifact — Equipment")
	elsewhere := effectstest.B12Permanent(g, opp.ID, "Other Aura", "Enchantment — Aura")
	for _, a := range []uuid.UUID{aura, equip} {
		effectstest.AttachForTest(t, g, a, host)
	}
	effectstest.AttachForTest(t, g, elsewhere, other)
	effectstest.CastCatalogSpell(t, g, "Strip Bare", "Instant", oldB538StripBare, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, aura) || effectstest.OnBattlefieldByID(g, equip) {
		t.Error("an Aura or Equipment attached to the target survived")
	}
	if !effectstest.OnBattlefieldByID(g, host) {
		t.Error("the creature itself was destroyed")
	}
	if !effectstest.OnBattlefieldByID(g, elsewhere) {
		t.Error("an Aura on a different creature was destroyed")
	}
}

func TestOldB538GloomlanceDestroysAndOnlyGreenOrWhiteDiscard(t *testing.T) {
	for _, tc := range []struct {
		color   string
		discard bool
	}{{"G", true}, {"W", true}, {"B", false}, {"R", false}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Victim", TypeLine: "Creature — Bear", Colors: []string{tc.color},
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		})
		hand := opp.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Gloomlance", "Sorcery", oldB538Gloomlance, effectstest.CardTarget(victim))
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, victim) {
			t.Errorf("%s: the creature was not destroyed", tc.color)
		}
		owed := effectstest.DiscardOwed(g, opp.ID)
		if tc.discard {
			if owed != 1 {
				t.Fatalf("%s: owed %d discards, want 1", tc.color, owed)
			}
			effectstest.DiscardFromHand(t, g, opp.ID)
			if opp.Hand.Size() != hand-1 {
				t.Errorf("%s: hand %d, want %d", tc.color, opp.Hand.Size(), hand-1)
			}
		} else if owed != 0 {
			t.Errorf("%s: owed %d discards, want none", tc.color, owed)
		}
	}
}

func TestOldB538PoisonTheWellDestroysTheLandAndHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.PushLand(g, opp.ID, "Island")
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Poison the Well", "Sorcery", oldB538PoisonWell, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, land) {
		t.Error("the land was not destroyed")
	}
	if opp.Life != theirs-2 {
		t.Errorf("the land's controller is at %d, want %d", opp.Life, theirs-2)
	}
	if me.Life != mine {
		t.Errorf("the caster's life changed: %d → %d", mine, me.Life)
	}
}
