package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b130_cards_test.go — S82 OLD batch 130 (#5514): Theros Beyond
// Death Auras. Every test plays the real catalog card through the engine.

const (
	b130Presence = "03160a74-8b83-4e07-b702-977a91a6311f"
	b130Stars    = "0ab43fc3-53ec-4686-873d-8dbeee2634f4"
	b130Iroas    = "76905904-b3e4-44ee-bfcb-e518f0815af5"
	b130Ichthyo  = "816e13c6-79d8-431f-a04e-15c423baf7b3"
	b130Heliod   = "ac5d5133-6f5c-43b7-82da-24802246e5a4"
	b130Mantle   = "cd69b2c7-2be7-4915-9d17-4fd400ace6b8"
	b130End      = "e1a5c389-6a56-4da7-b7c4-e26fdec224e5"
	b130Aura     = "Enchantment — Aura"
	b130Bear     = "Creature — Bear"
)

func TestB130EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b130Presence: "Commanding Presence", b130Stars: "One with the Stars", b130Iroas: "Iroas's Blessing",
		b130Ichthyo: "Ichthyomorphosis", b130Heliod: "Heliod's Punishment", b130Mantle: "Mantle of the Wolf",
		b130End: "Inevitable End",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func b130Has(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func b130Cast(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, b130Aura, oracle, effectstest.B16TargetCard(target))
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestB130CommandingPresenceBuffsGivesFirstStrikeAndMakesASoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	b130Cast(t, g, "Commanding Presence", b130Presence, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("enchanted Bear is %d/%d, want 4/4", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the Bear lacks first strike")
	}
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("Soldier tokens = %d, want 1", n)
	}
	if n := effectstest.CountTokensControlled(g, me.ID, "Human"); n != 1 {
		t.Errorf("Human tokens = %d, want 1", n)
	}
}

func TestB130OneWithTheStarsMakesACreatureAnEnchantmentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	b130Cast(t, g, "One with the Stars", b130Stars, bear)
	c := effectstest.GaLayered(t, g, bear)
	if c.IsCreature() || !c.IsEnchantment() {
		t.Errorf("enchanted permanent: creature=%v enchantment=%v, want not a creature and an enchantment", c.IsCreature(), c.IsEnchantment())
	}
}

func TestB130OneWithTheStarsEnchantsAnEnchantmentAndKeepsItsSubtype(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ench := effectstest.B12Permanent(g, me.ID, "Some Saga", "Enchantment — Saga")
	b130Cast(t, g, "One with the Stars", b130Stars, ench)
	c := effectstest.GaLayered(t, g, ench)
	eff := c.Effective()
	if !c.IsEnchantment() || !b130Has(eff.Subtypes, "Saga") {
		t.Errorf("an enchantment lost its Saga subtype: %v %v", eff.Types, eff.Subtypes)
	}
	// Neither a land nor an artifact is a legal "enchant creature or enchantment".
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "One with the Stars", b130Aura, b130Stars, effectstest.B16TargetCard(land)); err == nil {
		t.Error("enchanted a land")
	}
}

func TestB130IroassBlessingDealsFourAndPumpsOnlyYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", b130Bear, 2, 4)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Iroas's Blessing", b130Aura, b130Iroas, effectstest.B16TargetCard(theirs)); err == nil {
		t.Error("enchanted an opponent's creature")
	}
	effectstest.CastCatalogSpell(t, g, "Iroas's Blessing", b130Aura, b130Iroas, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, theirs)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 4-toughness creature survived 4 damage")
	}
	if p, tg := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tg != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tg)
	}
}

func TestB130IchthyomorphosisMakesItABlueFishWithNoAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	drake := effectstest.B12Creature(g, me.ID, "Drake", "Creature — Drake", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == drake {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Fatal("setup: the Drake does not fly")
	}
	b130Cast(t, g, "Ichthyomorphosis", b130Ichthyo, drake)
	if p, tg := effectstest.EffectivePower(t, g, drake), effectstest.EffectiveToughness(t, g, drake); p != 0 || tg != 1 {
		t.Errorf("enchanted Drake is %d/%d, want 0/1", p, tg)
	}
	if effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("the Fish still flies")
	}
	c := effectstest.GaLayered(t, g, drake)
	eff := c.Effective()
	if !c.IsCreature() || !b130Has(eff.Subtypes, "Fish") || b130Has(eff.Subtypes, "Drake") {
		t.Errorf("types %v subtypes %v, want a Fish creature and no Drake", eff.Types, eff.Subtypes)
	}
	if len(eff.Colors) != 1 || eff.Colors[0] != "U" {
		t.Errorf("colors %v, want blue only", eff.Colors)
	}
}

func TestB130HeliodsPunishmentCountsDownAndDestroysItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	aura := b130Cast(t, g, "Heliod's Punishment", b130Heliod, bear)
	if n := effectstest.B12Counter(t, g, aura, "task"); n != 4 {
		t.Fatalf("task counters %d, want 4", n)
	}
	for i := 0; i < 4; i++ {
		effectstest.B08Untap(g, bear)
		idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
		if idx < 0 {
			t.Fatalf("round %d: the enchanted creature has no granted ability", i)
		}
		effectstest.B16Activate(t, g, me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		if i < 3 && (!g.Battlefield.Contains(aura) || effectstest.B12Counter(t, g, aura, "task") != 3-i) {
			t.Fatalf("round %d: task counters %d, want %d", i, effectstest.B12Counter(t, g, aura, "task"), 3-i)
		}
	}
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura survived its last task counter")
	}
	_ = opp
}

func TestB130HeliodsPunishmentCantAttackAndLosesAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	b130Cast(t, g, "Heliod's Punishment", b130Heliod, bear)
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature kept flying")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
}

func TestB130MantleOfTheWolfBuffsAndLeavesTwoWolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	mantle := b130Cast(t, g, "Mantle of the Wolf", b130Mantle, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 6 || tg != 6 {
		t.Errorf("enchanted Bear is %d/%d, want 6/6", p, tg)
	}
	if n := effectstest.CountTokensControlled(g, me.ID, "Wolf"); n != 0 {
		t.Fatalf("Wolves before the Aura left = %d", n)
	}
	// The creature dying puts the Aura into the graveyard from the battlefield.
	effectstest.P1Destroy(t, g, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mantle) {
		t.Fatal("the Aura stayed on the battlefield")
	}
	if n := effectstest.CountTokensControlled(g, me.ID, "Wolf"); n != 2 {
		t.Errorf("Wolf tokens = %d, want 2", n)
	}
}

func TestB130InevitableEndMakesTheEnchantedCreaturesControllerSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.PushVanillaCreature(g, me.ID, "Host", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 2, 2)
	b130Cast(t, g, "Inevitable End", b130End, host)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) {
		t.Error("the chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(host) {
		t.Error("the host was sacrificed though another creature was chosen")
	}
}
