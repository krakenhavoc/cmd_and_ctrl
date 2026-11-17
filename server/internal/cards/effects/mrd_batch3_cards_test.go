package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered as %q, %v; want its name and Full", name, spec.Name, spec.Completeness)
		}
	}
}

// mrd3BlockedBy declares `attacker` attacking `defender` and blocked by
// `blocker`, with the blocks locked in.
func mrd3BlockedBy(t *testing.T, g *game.Game, defender, attacker, blocker uuid.UUID) {
	t.Helper()
	declareAttack(t, g, defender, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
}

// --- Dross Harvester --------------------------------------------------------

func TestMrdDrossHarvesterLosesFourAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mrd2Catalog(g, me.ID, "Dross Harvester", "Creature — Horror", mrdDrossHarvesterOracle, "{1}{B}{B}", 4, 4)
	life := me.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("life at the end step %d, want %d", me.Life, life-4)
	}
}

func TestMrdDrossHarvesterGainsTwoForEveryCreatureThatDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harvester := mrd2Catalog(g, me.ID, "Dross Harvester", "Creature — Horror", mrdDrossHarvesterOracle, "{1}{B}{B}", 4, 4)
	theirs := seedBear(g, opp.ID)
	mine := seedBear(g, me.ID)
	advanceToMain(t, g)
	life := me.Life

	p1Destroy(t, g, theirs)
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, mine)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("life %d after two creatures died, want %d", me.Life, life+4)
	}
	p1Destroy(t, g, harvester)
	passPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d after the Harvester itself died, want %d", me.Life, life+6)
	}
}

// --- Dross Prowler ----------------------------------------------------------

func TestMrdDrossProwlerHasFear(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	prowler := mrd2Catalog(g, me.ID, "Dross Prowler", "Creature — Zombie", mrdDrossProwlerOracle, "{2}{B}", 2, 1)
	green := mrdPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	black := mrdPermanent(g, opp.ID, "Their Zombie", "Creature — Zombie", "{1}{B}", 2, 2)

	declareAttack(t, g, opp.ID, prowler)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(green, prowler); err == nil {
		t.Error("a green creature was allowed to block a creature with fear")
	}
	if err := g.DeclareBlocker(black, prowler); err != nil {
		t.Errorf("a black creature could not block: %v", err)
	}
}

// --- Dross Scorpion ---------------------------------------------------------

func TestMrdDrossScorpionUntapsAnArtifactWhenAnArtifactCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mrd2Catalog(g, me.ID, "Dross Scorpion", "Artifact Creature — Scorpion", mrdDrossScorpionOracle, "{4}", 3, 1)
	rock := mrdPermanent(g, me.ID, "My Rock", "Artifact", "{2}", 0, 0)
	findBattlefieldCardForTest(g, rock).Tapped = true
	golem := mrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{3}", 3, 3)
	bear := seedBear(g, opp.ID)
	advanceToMain(t, g)

	p1Destroy(t, g, bear)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a nonartifact creature's death triggered the Scorpion")
	}

	p1Destroy(t, g, golem)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, rock)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if tappedForTest(t, g, rock) {
		t.Error("the target artifact should be untapped")
	}
}

func TestMrdDrossScorpionTriggersOnItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scorpion := mrd2Catalog(g, me.ID, "Dross Scorpion", "Artifact Creature — Scorpion", mrdDrossScorpionOracle, "{4}", 3, 1)
	rock := mrdPermanent(g, me.ID, "My Rock", "Artifact", "{2}", 0, 0)
	findBattlefieldCardForTest(g, rock).Tapped = true
	advanceToMain(t, g)

	p1Destroy(t, g, scorpion)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, rock)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if tappedForTest(t, g, rock) {
		t.Error("the Scorpion's own death should untap the target artifact")
	}
}

// --- Duskworker -------------------------------------------------------------

func TestMrdDuskworkerRegeneratesWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	worker := mrd2Catalog(g, me.ID, "Duskworker", "Artifact Creature — Construct", mrdDuskworkerOracle, "{4}", 2, 2)
	wall := mrdPermanent(g, opp.ID, "Wall", "Creature — Wall", "{1}", 0, 4)

	mrd3BlockedBy(t, g, opp.ID, worker, wall)
	if n := triggersOnStackFrom(g, worker); n != 1 {
		t.Fatalf("%d Duskworker triggers on the stack, want 1", n)
	}
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, worker)
	if !g.Battlefield.Contains(worker) {
		t.Error("the regeneration shield should have saved the Duskworker")
	}
	p1Destroy(t, g, worker)
	if g.Battlefield.Contains(worker) {
		t.Error("one block is one shield; the second destruction should stick")
	}
}

func TestMrdDuskworkerPumpsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	worker := mrd2Catalog(g, me.ID, "Duskworker", "Artifact Creature — Construct", mrdDuskworkerOracle, "{4}", 2, 2)
	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, worker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := mrd2Size(t, g, worker); p != 3 || tt != 2 {
		t.Errorf("Duskworker is %d/%d, want 3/2", p, tt)
	}
}

// --- Electrostatic Bolt -----------------------------------------------------

func TestMrdElectrostaticBoltDealsFourToAnArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	beast := mrdPermanent(g, opp.ID, "Their Beast", "Creature — Beast", "{2}{G}", 3, 5)
	golem := mrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{4}", 3, 5)

	castCatalogSpell(t, g, "Electrostatic Bolt", "Instant", mrdElectrostaticBoltOracl, cardRefs(beast))
	passPriorityAroundTable(t, g)
	if n := damageMarkedOn(g, beast); n != 2 {
		t.Errorf("a nonartifact creature took %d damage, want 2", n)
	}
	castCatalogSpell(t, g, "Electrostatic Bolt", "Instant", mrdElectrostaticBoltOracl, cardRefs(golem))
	passPriorityAroundTable(t, g)
	if n := damageMarkedOn(g, golem); n != 4 {
		t.Errorf("an artifact creature took %d damage, want 4", n)
	}
}

// --- Elf Replica ------------------------------------------------------------

func TestMrdElfReplicaSacrificesToDestroyAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	replica := mrd2Catalog(g, me.ID, "Elf Replica", "Artifact Creature — Elf", mrdElfReplicaOracle, "{3}", 2, 2)
	ench := mrdPermanent(g, opp.ID, "Their Enchantment", "Enchantment", "{1}{W}", 0, 0)

	b06AddMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, replica, 0, game.ActivateAbilityParams{Targets: cardRefs(ench), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(replica) {
		t.Error("the Replica should be sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || !opp.Graveyard.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
}

// --- Empyrial Plate ---------------------------------------------------------

func TestMrdEmpyrialPlateCountsItsControllersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	plate := seedEquipment(g, me.ID, "Empyrial Plate", mrdEmpyrialPlateOracle)
	equipTo(t, g, me.ID, plate, bear)
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", Owner: opp.ID, Controller: opp.ID})

	hand := me.Hand.Size()
	if p, tt := mrd2Size(t, g, bear); p != 2+hand || tt != 2+hand {
		t.Errorf("with %d cards in hand the bear is %d/%d, want %d/%d", hand, p, tt, 2+hand, 2+hand)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if p, tt := mrd2Size(t, g, bear); p != 3+hand || tt != 3+hand {
		t.Errorf("after a draw the bear is %d/%d, want %d/%d", p, tt, 3+hand, 3+hand)
	}
}

// --- Fangren Hunter ---------------------------------------------------------

func TestMrdFangrenHunterTramplesOverABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hunter := mrd2Catalog(g, me.ID, "Fangren Hunter", "Creature — Beast", mrdFangrenHunterOracle, "{3}{G}{G}", 4, 4)
	chump := mrdPermanent(g, opp.ID, "Their Squirrel", "Creature — Squirrel", "{G}", 1, 1)
	life := opp.Life

	mrd3BlockedBy(t, g, opp.ID, hunter, chump)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("defender at %d, want %d (3 tramples over the 1/1)", opp.Life, life-3)
	}
}

// --- Farsight Mask ----------------------------------------------------------

func TestMrdFarsightMaskDrawsWhenAnOpponentsSourceDamagesYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mask := mrd2Catalog(g, me.ID, "Farsight Mask", "Artifact", mrdFarsightMaskOracle, "{5}", 0, 0)
	theirs := seedBear(g, opp.ID)
	mine := seedBear(g, me.ID)
	advanceToMain(t, g)
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
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("damage from my own source triggered the Mask")
	}
	damage(theirs)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d cards, want 1", got)
	}

	findBattlefieldCardForTest(g, mask).Tapped = true
	damage(theirs)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil || me.Hand.Size()-hand != 1 {
		t.Error("a tapped Mask should not trigger")
	}
}

// --- Fists of the Anvil -----------------------------------------------------

func TestMrdFistsOfTheAnvilGivesPlusFourPlusZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedBear(g, me.ID)
	castCatalogSpell(t, g, "Fists of the Anvil", "Instant", mrdFistsOfTheAnvilOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tt := mrd2Size(t, g, bear); p != 6 || tt != 2 {
		t.Errorf("the bear is %d/%d, want 6/2", p, tt)
	}
}

// --- Flayed Nim -------------------------------------------------------------

func TestMrdFlayedNimMakesTheDamagedCreaturesControllerLoseThatMuch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	nim := mrd2Catalog(g, me.ID, "Flayed Nim", "Creature — Skeleton", mrdFlayedNimOracle, "{3}{B}", 2, 2)
	wall := mrdPermanent(g, opp.ID, "Wall", "Creature — Wall", "{1}", 0, 4)
	myLife, theirLife := me.Life, opp.Life

	mrd3BlockedBy(t, g, opp.ID, nim, wall)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 {
		t.Errorf("the blocker's controller at %d, want %d", opp.Life, theirLife-2)
	}
	if me.Life != myLife {
		t.Errorf("the Nim's controller at %d, want %d", me.Life, myLife)
	}
}

func TestMrdFlayedNimRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	nim := mrd2Catalog(g, me.ID, "Flayed Nim", "Creature — Skeleton", mrdFlayedNimOracle, "{3}{B}", 2, 2)
	b06AddMana(me, "C", "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, nim, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, nim)
	if !g.Battlefield.Contains(nim) || !tappedForTest(t, g, nim) {
		t.Error("the regenerated Nim should stay on the battlefield, tapped")
	}
}

// --- Fractured Loyalty ------------------------------------------------------

func TestMrdFracturedLoyaltyHandsTheCreatureToWhoeverTargetsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, opp.ID)
	aura := auraCast(t, g, "Fractured Loyalty", mrdFracturedLoyaltyOracle, bear)
	if got := controllerOf(t, g, bear); got != opp.ID {
		t.Fatalf("enchanting alone moved control to %s", got)
	}

	castCatalogSpell(t, g, "Fists of the Anvil", "Instant", mrdFistsOfTheAnvilOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, bear); got != me.ID {
		t.Fatalf("after I targeted it the creature is controlled by %s, want me", got)
	}

	p1Destroy(t, g, aura)
	if got := controllerOf(t, g, bear); got != me.ID {
		t.Error("the control change lasts indefinitely; it should outlive the Aura")
	}
}

// --- Frogmite ---------------------------------------------------------------

func TestMrdFrogmiteHasAffinityForArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	if got := selfPricedMV(t, g, me, mrdFrogmiteOracle, "Artifact Creature — Frog", "{4}"); got != 2 {
		t.Errorf("with two artifacts it costs mana value %d, want 2", got)
	}
	for i := 0; i < 3; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	if got := selfPricedMV(t, g, me, mrdFrogmiteOracle, "Artifact Creature — Frog", "{4}"); got != 0 {
		t.Errorf("with five artifacts it costs mana value %d, want 0", got)
	}
}

// --- Galvanic Key -----------------------------------------------------------

func TestMrdGalvanicKeyUntapsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	key := mrd2Catalog(g, me.ID, "Galvanic Key", "Artifact", mrdGalvanicKeyOracle, "{2}", 0, 0)
	rock := mrdPermanent(g, me.ID, "My Rock", "Artifact", "{2}", 0, 0)
	findBattlefieldCardForTest(g, rock).Tapped = true

	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, key, 0, game.ActivateAbilityParams{Targets: cardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tappedForTest(t, g, rock) {
		t.Error("the target artifact should be untapped")
	}
	if !tappedForTest(t, g, key) {
		t.Error("the Key taps as part of the cost")
	}
}
