package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b323_cards_test.go — Magic Origins, batch 323 (S82 older sets,
// #5549). Every test plays the real catalog entry through the engine.

const (
	oldB323Helm      = "52804094-47b9-40b9-999d-7c050d220e63"
	oldB323Knife     = "c3973121-ed59-48b3-8efa-754a1b2c704d"
	oldB323Wolf      = "c9375dca-2b32-4185-b88a-4bd5dde9a5d6"
	oldB323Ramroller = "d670d937-aaa7-4a96-9c6a-916ff7a6b935"
	oldB323Acolyte   = "35bf67e0-8141-4664-b762-2c613353e1eb"
	oldB323Trapper   = "d3cc22ea-2ccf-409b-b471-51adb63c7fda"
	oldB323Smash     = "e0583f29-c344-4b56-b0dc-bef1089a95c7"
	oldB323Bear      = "Creature — Bear"
)

func TestOldB323EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB323Helm, oldB323Knife, oldB323Wolf, oldB323Ramroller, oldB323Acolyte, oldB323Trapper, oldB323Smash} {
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

func TestOldB323HelmOfTheGodsCountsYourEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	helm := effectstest.SeedEquipment(g, me.ID, "Helm of the Gods", oldB323Helm)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", oldB323Bear, 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Fatalf("unequipped it is %d/%d", p, tough)
	}
	effectstest.EquipTo(t, g, me.ID, helm, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("with no enchantment it is %d/%d, want 2/2", p, tough)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Mine A", "Enchantment", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Mine B", "Enchantment", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Theirs", "Enchantment", "", false)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("with two of my enchantments it is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB323ThrowingKnifeGivesPlusTwoAndSacrificesToShoot(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		effectstest.AdvanceToMain(t, g)
		knife := effectstest.SeedEquipment(g, me.ID, "Throwing Knife", oldB323Knife)
		bear := effectstest.B12Creature(g, me.ID, "My Bear", oldB323Bear, 2, 2)
		effectstest.EquipTo(t, g, me.ID, knife, bear)
		if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 4 {
			t.Fatalf("equipped it has %d power, want 4", p)
		}
		victim := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB323Bear, 2, 2)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bear, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PickCard(t, g, me.ID, victim)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		gone := effectstest.ZoneOf(g, knife) != game.ZoneBattlefield
		dead := effectstest.ZoneOf(g, victim) != game.ZoneBattlefield
		if gone != yes || dead != yes {
			t.Errorf("yes=%v: knife gone %v, target dead %v", yes, gone, dead)
		}
	}
}

func TestOldB323TimberpackWolfCountsYourOtherWolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	a := effectstest.B12Push(g, me.ID, "Timberpack Wolf", "Creature — Wolf", oldB323Wolf, 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, a); p != 2 {
		t.Fatalf("alone it has %d power, want 2", p)
	}
	b := effectstest.B12Push(g, me.ID, "Timberpack Wolf", "Creature — Wolf", oldB323Wolf, 2, 2)
	effectstest.B12Push(g, opp.ID, "Timberpack Wolf", "Creature — Wolf", oldB323Wolf, 2, 2)
	effectstest.B12Creature(g, me.ID, "Other Wolf", "Creature — Wolf", 2, 2)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("with one other of mine it is %d/%d, want 3/3", p, tough)
		}
	}
	effectstest.B12Push(g, me.ID, "Timberpack Wolf", "Creature — Wolf", oldB323Wolf, 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, a); p != 4 {
		t.Errorf("with two others of mine it has %d power, want 4", p)
	}
}

func TestOldB323RamrollerNeedsAnotherArtifactAndMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	rammer := effectstest.B12Push(g, me.ID, "Ramroller", "Artifact Creature — Juggernaut", oldB323Ramroller, 2, 3)
	if p, _ := effectstest.EffectivePTOf(t, g, rammer); p != 2 {
		t.Errorf("alone it has %d power, want 2 (it is not another artifact)", p)
	}
	if len(effectstest.E2Card(t, g, rammer).Effective().AttackRequirements) == 0 {
		t.Error("it has no attack requirement")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Spare Artifact", "Artifact", "", false)
	if p, tough := effectstest.EffectivePTOf(t, g, rammer); p != 4 || tough != 3 {
		t.Errorf("with another artifact it is %d/%d, want 4/3", p, tough)
	}
}

func TestOldB323RamrollerCantSkipTheAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.B12Push(g, me.ID, "Ramroller", "Artifact Creature — Juggernaut", oldB323Ramroller, 2, 3)
	if !effectstest.FaMustAttack(t, g) {
		t.Error("the active player could pass the declare attackers step without attacking")
	}
}

func TestOldB323AcolyteDamagesEachBlockerAndGrowsWithRenown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	acolyte := effectstest.B12Push(g, me.ID, "Acolyte of the Inferno", "Creature — Human Monk", oldB323Acolyte, 3, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big Blocker", oldB323Bear, 0, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(acolyte, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, acolyte); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if dmg := effectstest.DamageOn(t, g, big); dmg != 2 {
		t.Errorf("the blocker has %d damage, want 2 from the trigger", dmg)
	}
}

func TestOldB323AcolyteRenownGivesACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	acolyte := effectstest.B12Push(g, me.ID, "Acolyte of the Inferno", "Creature — Human Monk", oldB323Acolyte, 3, 1)
	effectstest.AttackWithTrigger(t, g, acolyte, opp.ID, uuid.Nil)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, acolyte, game.CounterPlusOne); n != 1 {
		t.Errorf("the Acolyte has %d +1/+1 counters after connecting, want 1", n)
	}
}

func TestOldB323GnarlrootTrapperManaIsForElfCreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	trapper := effectstest.PushCatalogPermanent(g, me.ID, "Gnarlroot Trapper", "Creature — Elf Druid", oldB323Trapper, false)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, trapper, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d after paying 1", me.Life, life-1)
	}
	if me.ManaPool.CountColor("G") != 1 {
		t.Fatalf("pool %v, want one G", me.ManaPool.Units())
	}
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the restricted mana paid for a non-Elf creature spell")
	}
	elf := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: elf, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the restricted mana could not pay for an Elf creature spell: %v", err)
	}
}

func TestOldB323GnarlrootTrapperGivesAnAttackingElfDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	trapper := effectstest.B12Push(g, me.ID, "Gnarlroot Trapper", "Creature — Elf Druid", oldB323Trapper, 1, 1)
	elf := effectstest.B12Creature(g, me.ID, "Attacking Elf", "Creature — Elf Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Attacking Bear", oldB323Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, elf, bear)
	if err := g.ActivateCatalogAbility(me.ID, trapper, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a Bear was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, trapper, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elf}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, elf), "deathtouch") {
		t.Error("the Elf did not gain deathtouch")
	}
}

func TestOldB323SmashToSmithereensDestroysAnArtifactAndHitsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	artifact := effectstest.PushCatalogPermanent(g, opp.ID, "Their Artifact", "Artifact", "", false)
	life, mine := opp.Life, me.Life
	effectstest.CastCatalogSpell(t, g, "Smash to Smithereens", "Instant", oldB323Smash, []game.TargetRef{{Kind: game.TargetCard, ID: artifact}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, artifact) == game.ZoneBattlefield {
		t.Error("the artifact was not destroyed")
	}
	if opp.Life != life-3 {
		t.Errorf("its controller is at %d, want %d", opp.Life, life-3)
	}
	if me.Life != mine {
		t.Error("the caster lost life")
	}
}

func TestOldB323SmashToSmithereensStillHitsWhenTheArtifactSurvives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	artifact := effectstest.PushCatalogPermanent(g, me.ID, "My Artifact", "Artifact", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == artifact {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "indestructible")
			}
		}
	})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Smash to Smithereens", "Instant", oldB323Smash, []game.TargetRef{{Kind: game.TargetCard, ID: artifact}})
	effectstest.PassPriorityAroundTable(t, g)
	_ = opp
	if me.Life != life-3 {
		t.Errorf("the artifact's controller is at %d, want %d", me.Life, life-3)
	}
}
