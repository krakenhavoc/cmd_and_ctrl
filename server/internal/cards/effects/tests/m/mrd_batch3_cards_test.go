package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mrd_batch3_cards_test.go — Mirrodin, batch 3 (#5606, S82): Dross
// Harvester onwards.

const (
	mrdDrossHarvesterOracle   = "5520ada0-f173-4b5e-a0b8-a5938a1c0a4b"
	mrdDrossProwlerOracle     = "a3122c6a-20be-4bb3-ab2f-928143378ea6"
	mrdDrossScorpionOracle    = "132786bb-157c-4e2c-8f82-13df6a667448"
	mrdDuskworkerOracle       = "14d85a83-323b-4e42-88bf-71167a8174e3"
	mrdElectrostaticBoltOracl = "6fc26faa-94a8-433f-8863-126c2d2e73b0"
	mrdElfReplicaOracle       = "2dda5011-aaa1-48d0-afa8-76998278ab75"
	mrdEmpyrialPlateOracle    = "e95e100e-4886-4bde-825c-017e878ac825"
	mrdFangrenHunterOracle    = "c5dc5546-e9e5-4b5b-b812-5716d4bdee0e"
	mrdFarsightMaskOracle     = "c82091b2-f2f1-4a85-80d6-91ab218f0a94"
	mrdFistsOfTheAnvilOracle  = "fc05e582-e760-4fe2-ba43-e9d8e63f3f85"
	mrdFlayedNimOracle        = "857818d4-6cf0-4234-b9e0-d7181ba6b54f"
	mrdFracturedLoyaltyOracle = "ccc168ef-b93b-4bff-89a3-80984d3625de"
	mrdFrogmiteOracle         = "a927fd33-f183-48d9-b582-55d379d70e13"
	mrdGalvanicKeyOracle      = "d8a552ca-2c7b-410e-bd7e-1bb81465277a"
)

func TestMrdBatch3CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Dross Harvester": mrdDrossHarvesterOracle, "Dross Prowler": mrdDrossProwlerOracle,
		"Dross Scorpion": mrdDrossScorpionOracle, "Duskworker": mrdDuskworkerOracle,
		"Electrostatic Bolt": mrdElectrostaticBoltOracl, "Elf Replica": mrdElfReplicaOracle,
		"Empyrial Plate": mrdEmpyrialPlateOracle, "Fangren Hunter": mrdFangrenHunterOracle,
		"Farsight Mask": mrdFarsightMaskOracle, "Fists of the Anvil": mrdFistsOfTheAnvilOracle,
		"Flayed Nim": mrdFlayedNimOracle, "Fractured Loyalty": mrdFracturedLoyaltyOracle,
		"Frogmite": mrdFrogmiteOracle, "Galvanic Key": mrdGalvanicKeyOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: registered as %q, %v; want its name and Full", name, spec.Name, spec.Completeness)
		}
	}
}

// --- Dross Harvester --------------------------------------------------------

func TestMrdDrossHarvesterLosesFourAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Mrd2Catalog(g, me.ID, "Dross Harvester", "Creature — Horror", mrdDrossHarvesterOracle, "{1}{B}{B}", 4, 4)
	life := me.Life
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("life at the end step %d, want %d", me.Life, life-4)
	}
}

func TestMrdDrossHarvesterGainsTwoForEveryCreatureThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harvester := effectstest.Mrd2Catalog(g, me.ID, "Dross Harvester", "Creature — Horror", mrdDrossHarvesterOracle, "{1}{B}{B}", 4, 4)
	theirs := effectstest.SeedBear(g, opp.ID)
	mine := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	life := me.Life

	effectstest.P1Destroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("life %d after two creatures died, want %d", me.Life, life+4)
	}
	effectstest.P1Destroy(t, g, harvester)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d after the Harvester itself died, want %d", me.Life, life+6)
	}
}

// --- Dross Prowler ----------------------------------------------------------

func TestMrdDrossProwlerHasFear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	prowler := effectstest.Mrd2Catalog(g, me.ID, "Dross Prowler", "Creature — Zombie", mrdDrossProwlerOracle, "{2}{B}", 2, 1)
	green := effectstest.MrdPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	black := effectstest.MrdPermanent(g, opp.ID, "Their Zombie", "Creature — Zombie", "{1}{B}", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, prowler)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(green, prowler); err == nil {
		t.Error("a green creature was allowed to block a creature with fear")
	}
	if err := g.DeclareBlocker(black, prowler); err != nil {
		t.Errorf("a black creature could not block: %v", err)
	}
}

// --- Dross Scorpion ---------------------------------------------------------

func TestMrdDrossScorpionUntapsAnArtifactWhenAnArtifactCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Mrd2Catalog(g, me.ID, "Dross Scorpion", "Artifact Creature — Scorpion", mrdDrossScorpionOracle, "{4}", 3, 1)
	rock := effectstest.MrdPermanent(g, me.ID, "My Rock", "Artifact", "{2}", 0, 0)
	effectstest.FindBattlefieldCardForTest(g, rock).Tapped = true
	golem := effectstest.MrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{3}", 3, 3)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToMain(t, g)

	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a nonartifact creature's death triggered the Scorpion")
	}

	effectstest.P1Destroy(t, g, golem)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedForTest(t, g, rock) {
		t.Error("the target artifact should be untapped")
	}
}

func TestMrdDrossScorpionTriggersOnItsOwnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scorpion := effectstest.Mrd2Catalog(g, me.ID, "Dross Scorpion", "Artifact Creature — Scorpion", mrdDrossScorpionOracle, "{4}", 3, 1)
	rock := effectstest.MrdPermanent(g, me.ID, "My Rock", "Artifact", "{2}", 0, 0)
	effectstest.FindBattlefieldCardForTest(g, rock).Tapped = true
	effectstest.AdvanceToMain(t, g)

	effectstest.P1Destroy(t, g, scorpion)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedForTest(t, g, rock) {
		t.Error("the Scorpion's own death should untap the target artifact")
	}
}

// --- Duskworker -------------------------------------------------------------

func TestMrdDuskworkerRegeneratesWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	worker := effectstest.Mrd2Catalog(g, me.ID, "Duskworker", "Artifact Creature — Construct", mrdDuskworkerOracle, "{4}", 2, 2)
	wall := effectstest.MrdPermanent(g, opp.ID, "Wall", "Creature — Wall", "{1}", 0, 4)

	effectstest.Mrd3BlockedBy(t, g, opp.ID, worker, wall)
	if n := effectstest.TriggersOnStackFrom(g, worker); n != 1 {
		t.Fatalf("%d Duskworker triggers on the stack, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, worker)
	if !g.Battlefield.Contains(worker) {
		t.Error("the regeneration shield should have saved the Duskworker")
	}
	effectstest.P1Destroy(t, g, worker)
	if g.Battlefield.Contains(worker) {
		t.Error("one block is one shield; the second destruction should stick")
	}
}

func TestMrdDuskworkerPumpsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	worker := effectstest.Mrd2Catalog(g, me.ID, "Duskworker", "Artifact Creature — Construct", mrdDuskworkerOracle, "{4}", 2, 2)
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, worker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.Mrd2Size(t, g, worker); p != 3 || tt != 2 {
		t.Errorf("Duskworker is %d/%d, want 3/2", p, tt)
	}
}

// --- Electrostatic Bolt -----------------------------------------------------

func TestMrdElectrostaticBoltDealsFourToAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	beast := effectstest.MrdPermanent(g, opp.ID, "Their Beast", "Creature — Beast", "{2}{G}", 3, 5)
	golem := effectstest.MrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{4}", 3, 5)

	effectstest.CastCatalogSpell(t, g, "Electrostatic Bolt", "Instant", mrdElectrostaticBoltOracl, effectstest.CardRefs(beast))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, beast); n != 2 {
		t.Errorf("a nonartifact creature took %d damage, want 2", n)
	}
	effectstest.CastCatalogSpell(t, g, "Electrostatic Bolt", "Instant", mrdElectrostaticBoltOracl, effectstest.CardRefs(golem))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, golem); n != 4 {
		t.Errorf("an artifact creature took %d damage, want 4", n)
	}
}

// --- Elf Replica ------------------------------------------------------------

func TestMrdElfReplicaSacrificesToDestroyAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	replica := effectstest.Mrd2Catalog(g, me.ID, "Elf Replica", "Artifact Creature — Elf", mrdElfReplicaOracle, "{3}", 2, 2)
	ench := effectstest.MrdPermanent(g, opp.ID, "Their Enchantment", "Enchantment", "{1}{W}", 0, 0)

	effectstest.B06AddMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, replica, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ench), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(replica) {
		t.Error("the Replica should be sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || !opp.Graveyard.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
}

// --- Empyrial Plate ---------------------------------------------------------

func TestMrdEmpyrialPlateCountsItsControllersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	plate := effectstest.SeedEquipment(g, me.ID, "Empyrial Plate", mrdEmpyrialPlateOracle)
	effectstest.EquipTo(t, g, me.ID, plate, bear)
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", Owner: opp.ID, Controller: opp.ID})

	hand := me.Hand.Size()
	if p, tt := effectstest.Mrd2Size(t, g, bear); p != 2+hand || tt != 2+hand {
		t.Errorf("with %d cards in hand the bear is %d/%d, want %d/%d", hand, p, tt, 2+hand, 2+hand)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if p, tt := effectstest.Mrd2Size(t, g, bear); p != 3+hand || tt != 3+hand {
		t.Errorf("after a draw the bear is %d/%d, want %d/%d", p, tt, 3+hand, 3+hand)
	}
}

// --- Fangren Hunter ---------------------------------------------------------

func TestMrdFangrenHunterTramplesOverABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hunter := effectstest.Mrd2Catalog(g, me.ID, "Fangren Hunter", "Creature — Beast", mrdFangrenHunterOracle, "{3}{G}{G}", 4, 4)
	chump := effectstest.MrdPermanent(g, opp.ID, "Their Squirrel", "Creature — Squirrel", "{G}", 1, 1)
	life := opp.Life

	effectstest.Mrd3BlockedBy(t, g, opp.ID, hunter, chump)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("defender at %d, want %d (3 tramples over the 1/1)", opp.Life, life-3)
	}
}

// --- Farsight Mask ----------------------------------------------------------

func TestMrdFarsightMaskDrawsWhenAnOpponentsSourceDamagesYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mask := effectstest.Mrd2Catalog(g, me.ID, "Farsight Mask", "Artifact", mrdFarsightMaskOracle, "{5}", 0, 0)
	theirs := effectstest.SeedBear(g, opp.ID)
	mine := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	damage := func(source uuid.UUID) {
		t.Helper()
		g.WithWriteLock(func() {
			if err := g.DealDamageToPlayerForEffect(source, me.ID, 1); err != nil {
				t.Fatalf("deal damage: %v", err)
			}
		})
	}

	hand := me.Hand.Size()
	damage(mine)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("damage from my own source triggered the Mask")
	}
	damage(theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d cards, want 1", got)
	}

	effectstest.FindBattlefieldCardForTest(g, mask).Tapped = true
	damage(theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil || me.Hand.Size()-hand != 1 {
		t.Error("a tapped Mask should not trigger")
	}
}

// --- Fists of the Anvil -----------------------------------------------------

func TestMrdFistsOfTheAnvilGivesPlusFourPlusZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Fists of the Anvil", "Instant", mrdFistsOfTheAnvilOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.Mrd2Size(t, g, bear); p != 6 || tt != 2 {
		t.Errorf("the bear is %d/%d, want 6/2", p, tt)
	}
}

// --- Flayed Nim -------------------------------------------------------------

func TestMrdFlayedNimMakesTheDamagedCreaturesControllerLoseThatMuch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	nim := effectstest.Mrd2Catalog(g, me.ID, "Flayed Nim", "Creature — Skeleton", mrdFlayedNimOracle, "{3}{B}", 2, 2)
	wall := effectstest.MrdPermanent(g, opp.ID, "Wall", "Creature — Wall", "{1}", 0, 4)
	myLife, theirLife := me.Life, opp.Life

	effectstest.Mrd3BlockedBy(t, g, opp.ID, nim, wall)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 {
		t.Errorf("the blocker's controller at %d, want %d", opp.Life, theirLife-2)
	}
	if me.Life != myLife {
		t.Errorf("the Nim's controller at %d, want %d", me.Life, myLife)
	}
}

func TestMrdFlayedNimRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	nim := effectstest.Mrd2Catalog(g, me.ID, "Flayed Nim", "Creature — Skeleton", mrdFlayedNimOracle, "{3}{B}", 2, 2)
	effectstest.B06AddMana(me, "C", "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, nim, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, nim)
	if !g.Battlefield.Contains(nim) || !effectstest.TappedForTest(t, g, nim) {
		t.Error("the regenerated Nim should stay on the battlefield, tapped")
	}
}

// --- Fractured Loyalty ------------------------------------------------------

func TestMrdFracturedLoyaltyHandsTheCreatureToWhoeverTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, opp.ID)
	aura := effectstest.AuraCast(t, g, "Fractured Loyalty", mrdFracturedLoyaltyOracle, bear)
	if got := effectstest.ControllerOf(t, g, bear); got != opp.ID {
		t.Fatalf("enchanting alone moved control to %s", got)
	}

	effectstest.CastCatalogSpell(t, g, "Fists of the Anvil", "Instant", mrdFistsOfTheAnvilOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Fatalf("after I targeted it the creature is controlled by %s, want me", got)
	}

	effectstest.P1Destroy(t, g, aura)
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Error("the control change lasts indefinitely; it should outlive the Aura")
	}
}

// --- Frogmite ---------------------------------------------------------------

func TestMrdFrogmiteHasAffinityForArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	if got := effectstest.SelfPricedMV(t, g, me, mrdFrogmiteOracle, "Artifact Creature — Frog", "{4}"); got != 2 {
		t.Errorf("with two artifacts it costs mana value %d, want 2", got)
	}
	for i := 0; i < 3; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	if got := effectstest.SelfPricedMV(t, g, me, mrdFrogmiteOracle, "Artifact Creature — Frog", "{4}"); got != 0 {
		t.Errorf("with five artifacts it costs mana value %d, want 0", got)
	}
}

// --- Galvanic Key -----------------------------------------------------------

func TestMrdGalvanicKeyUntapsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	key := effectstest.Mrd2Catalog(g, me.ID, "Galvanic Key", "Artifact", mrdGalvanicKeyOracle, "{2}", 0, 0)
	rock := effectstest.MrdPermanent(g, me.ID, "My Rock", "Artifact", "{2}", 0, 0)
	effectstest.FindBattlefieldCardForTest(g, rock).Tapped = true

	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, key, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedForTest(t, g, rock) {
		t.Error("the target artifact should be untapped")
	}
	if !effectstest.TappedForTest(t, g, key) {
		t.Error("the Key taps as part of the cost")
	}
}
