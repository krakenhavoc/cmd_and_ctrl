package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b515_cards_test.go — Shards of Alara, batch 515 (S82 older sets,
// #5588). Every test plays the real catalog entry through the engine.

const (
	oldB515CreeperOracle    = "0f8e0b2c-f1a5-40ad-b8d8-7c06a012e909"
	oldB515DuelistOracle    = "2b6aaa8b-b90d-459f-95d4-9e00ee134f8e"
	oldB515RhoxOracle       = "5c05c350-1b11-4dda-9de6-01c7e0f07ffc"
	oldB515CrasherOracle    = "5d45fa1f-110c-4bc8-a0fc-ed04556cf639"
	oldB515AmbusherOracle   = "cbd106fb-c9c5-4c5e-8801-619b7c8ff5a6"
	oldB515ArcherOracle     = "19055292-1a79-413e-8056-7d4aec9839ed"
	oldB515PlatoonOracle    = "3bc82914-1524-4d57-8b38-9788867a4046"
	oldB515BloodpyreOracle  = "7c683855-f32e-42f9-8681-c3335d5eaf79"
	oldB515ShadowsOracle    = "9707ac4f-6493-47fe-9c31-4daf469b5c6e"
	oldB515FirestokerOracle = "f8c0c555-c750-4842-8924-d634206bbb81"
)

func TestOldB515EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB515CreeperOracle:    "Kederekt Creeper",
		oldB515DuelistOracle:    "Deft Duelist",
		oldB515RhoxOracle:       "Rhox War Monk",
		oldB515CrasherOracle:    "Rip-Clan Crasher",
		oldB515AmbusherOracle:   "Qasali Ambusher",
		oldB515ArcherOracle:     "Dawnray Archer",
		oldB515PlatoonOracle:    "Rockcaster Platoon",
		oldB515BloodpyreOracle:  "Bloodpyre Elemental",
		oldB515ShadowsOracle:    "Vicious Shadows",
		oldB515FirestokerOracle: "Exuberant Firestoker",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB515KeywordCreaturesHaveTheirKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		kws                    []string
	}{
		{"Kederekt Creeper", "Creature — Horror", oldB515CreeperOracle, []string{"menace", "deathtouch"}},
		{"Deft Duelist", "Creature — Human Rogue", oldB515DuelistOracle, []string{"first strike", "shroud"}},
		{"Rhox War Monk", "Creature — Rhino Monk", oldB515RhoxOracle, []string{"lifelink"}},
		{"Rip-Clan Crasher", "Creature — Human Warrior", oldB515CrasherOracle, []string{"haste"}},
		{"Qasali Ambusher", "Creature — Cat Warrior", oldB515AmbusherOracle, []string{"reach"}},
		{"Dawnray Archer", "Creature — Human Archer", oldB515ArcherOracle, []string{"exalted"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.kws {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func TestOldB515DawnrayArcherPingsOnlyAnAttackingOrBlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	archer := effectstest.B12Push(g, me.ID, "Dawnray Archer", "Creature — Human Archer", oldB515ArcherOracle, 1, 1)
	attacker := effectstest.B12Creature(g, me.ID, "My Attacker", "Creature — Bear", 1, 1)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	// Neither attacking nor blocking: refused.
	if err := g.ActivateCatalogAbility(me.ID, archer, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}},
	}); err == nil {
		t.Fatal("the Archer pinged a creature that is neither attacking nor blocking")
	}
	if err := g.ActivateCatalogAbility(me.ID, archer, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}},
	}); err != nil {
		t.Fatalf("activate on an attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, attacker); ok {
		t.Error("1 damage did not kill the attacking 1/1")
	}
}

func TestOldB515BloodpyreElementalIsSorcerySpeedAndSacrificesForFourDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elemental := effectstest.B12Push(g, me.ID, "Bloodpyre Elemental", "Creature — Elemental", oldB515BloodpyreOracle, 4, 1)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Giant", 4, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, elemental, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, elemental); ok {
		t.Error("the Elemental is still on the battlefield after paying its sacrifice cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("4 damage did not destroy the 4/4")
	}
}

func TestOldB515BloodpyreElementalRefusesToActivateOutsideAMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	elemental := effectstest.B12Push(g, me.ID, "Bloodpyre Elemental", "Creature — Elemental", oldB515BloodpyreOracle, 4, 1)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Giant", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, elemental, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err == nil {
		t.Error("activated at instant speed; the ability is sorcery-speed only")
	}
	if _, ok := effectstest.BattlefieldCard(g, elemental); !ok {
		t.Error("the refused activation still sacrificed the Elemental")
	}
}

func TestOldB515ViciousShadowsDealsDamageEqualToTheTargetsHand(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Vicious Shadows", "Enchantment", oldB515ShadowsOracle, false)
		for i := 0; i < 3; i++ {
			effectstest.PushCatalogHandCard(opp, "Spare", "Sorcery", "")
		}
		handSize := opp.Hand.Size()
		bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		before := effectstest.LifeOf(g, opp.ID)
		effectstest.KillCreature(t, g, opp.ID, bear)
		effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OwedMays(g, me.ID) == 0 {
			t.Fatal("a creature dying must offer the damage")
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := before
		if yes {
			want -= opp.Hand.Size()
		}
		if got := effectstest.LifeOf(g, opp.ID); got != want {
			t.Errorf("answer %v: life %d -> %d, want %d", yes, before, got, want)
		}
		if yes && before-effectstest.LifeOf(g, opp.ID) != handSize {
			t.Errorf("damage = %d, want %d (cards in hand)", before-effectstest.LifeOf(g, opp.ID), handSize)
		}
	}
}

func TestOldB515ExuberantFirestokerNeedsAPowerFiveCreatureAtTheEndStep(t *testing.T) {
	for _, power := range []int{4, 5} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Exuberant Firestoker", "Creature — Human Druid Shaman", oldB515FirestokerOracle, 1, 1)
		effectstest.B12Creature(g, me.ID, "Big Thing", "Creature — Beast", power, 5)
		before := effectstest.LifeOf(g, opp.ID)
		effectstest.AdvanceTo(t, g, game.StepEnd)
		if power == 5 {
			effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.OwedMays(g, me.ID) == 0 {
				t.Fatal("power 5: the end step trigger did not offer the damage")
			}
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if got := before - effectstest.LifeOf(g, opp.ID); got != 2 {
				t.Errorf("damage = %d, want 2", got)
			}
		} else if effectstest.OwedMays(g, me.ID) != 0 || effectstest.LifeOf(g, opp.ID) != before {
			t.Error("power 4: the trigger fired without a power 5 creature")
		}
	}
}

func TestOldB515ExuberantFirestokerTapsForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Exuberant Firestoker", "Creature — Human Druid Shaman", oldB515FirestokerOracle, 1, 1)
	effectstest.TapForMana(t, g, me.ID, id)
	if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool = %v, want one colorless", got)
	}
}

func TestOldB515RockcasterPlatoonHitsFliersAndEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	platoon := effectstest.B12Push(g, me.ID, "Rockcaster Platoon", "Creature — Rhino Soldier", oldB515PlatoonOracle, 5, 7)
	flier := effectstest.B12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	walker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 5; i++ {
		effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.AdvanceToMain(t, g)
	lifeMe, lifeOpp := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, platoon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, flier); ok {
		t.Error("the flying 1/2 survived 2 damage")
	}
	if got := effectstest.DamageOn(t, g, walker); got != 0 {
		t.Errorf("the ground creature took %d, want 0", got)
	}
	if effectstest.LifeOf(g, me.ID) != lifeMe-2 || effectstest.LifeOf(g, opp.ID) != lifeOpp-2 {
		t.Errorf("each player should lose 2: me %d->%d, opp %d->%d", lifeMe, effectstest.LifeOf(g, me.ID), lifeOpp, effectstest.LifeOf(g, opp.ID))
	}
}

func TestOldB515QasaliAmbusherIsFreeAndFlashOnlyUnderItsCondition(t *testing.T) {
	spec, ok := effects.Lookup(oldB515AmbusherOracle)
	if !ok || len(spec.AlternativeCosts) != 1 {
		t.Fatalf("Qasali Ambusher: ok=%v alternative costs=%d, want 1", ok, len(spec.AlternativeCosts))
	}
	alt := spec.AlternativeCosts[0]
	if alt.ManaCost != "" || !alt.AsThoughFlash || alt.Condition == nil {
		t.Fatalf("alternative cost = %+v, want free, as-though-flash and conditional", alt)
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	holds := func() (h bool) {
		g.ReadSnapshot(func() { h = alt.Condition(g, me.ID) })
		return h
	}
	if holds() {
		t.Error("the offer holds with an empty board")
	}
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	attacker := effectstest.B12Creature(g, opp.ID, "Their Attacker", "Creature — Bear", 2, 2)
	setAttack := func(target uuid.UUID) {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == attacker {
					g.Battlefield.Cards[i].AttackingTarget = target
				}
			}
		})
	}
	setAttack(me.ID)
	if holds() {
		t.Error("the offer holds with a Forest but no Plains")
	}
	effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	if !holds() {
		t.Error("the offer is closed although a creature attacks me and I control a Forest and a Plains")
	}
	setAttack(opp.ID)
	if holds() {
		t.Error("the offer holds for a creature attacking someone else")
	}
	setAttack(uuid.Nil)
	if holds() {
		t.Error("the offer holds with no attacker")
	}
	_ = forest
}
