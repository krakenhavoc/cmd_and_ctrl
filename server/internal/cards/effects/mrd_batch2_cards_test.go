package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mrd_batch2_cards_test.go — Mirrodin, batch 2 (#5606, S82): the A–C
// cards batch 1 left for later, then D onwards.

const (
	mrdAltarOfShadowsOracle    = "4b34a324-d939-453a-97e3-a21c584a0c74"
	mrdAuriokSteelshaperOracle = "e3ee8b0e-9bdc-4b9a-b4f9-59189dd6a80d"
	mrdBansheesBladeOracle     = "81a770e9-adea-4799-9fed-f10e97e7e901"
	mrdBlinkmothUrnOracle      = "08497e8f-3521-48e4-a1e7-c192b485d891"
	mrdCloudpostOracle         = "f705c0eb-9c6c-4315-a860-208ed0c5d93e"
	mrdContaminatedBondOracle  = "56aeaee8-e014-485e-b7ad-e26ade663bcf"
	mrdCopperhoofVorracOracle  = "a5fed574-6ac3-4318-b0a6-27082e07f3bd"
	mrdCrystalShardOracle      = "cd8ccb10-a2c8-4a84-aa41-6ce6018f9bb7"
	mrdCullingScalesOracle     = "ec4f2266-3e8e-4194-8f06-d95ebc543661"
	mrdDampingMatrixOracle     = "10403a8c-2534-421e-8809-5fc019623452"
	mrdDeadIronSledgeOracle    = "e1ee79de-e4c3-48f9-87a7-18f72e2f6116"
	mrdDeconstructOracle       = "36a8ceb1-148b-41e0-a7bf-ceb879bf08e7"
	mrdDisarmOracle            = "44d54814-2eeb-4df8-b8d3-789f0bcb307a"
	mrdDomineerOracle          = "ca283839-864f-4252-8452-480e3ba29b20"
)

func TestMrdBatch2CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Altar of Shadows": mrdAltarOfShadowsOracle, "Auriok Steelshaper": mrdAuriokSteelshaperOracle,
		"Banshee's Blade": mrdBansheesBladeOracle, "Blinkmoth Urn": mrdBlinkmothUrnOracle,
		"Cloudpost": mrdCloudpostOracle, "Contaminated Bond": mrdContaminatedBondOracle,
		"Copperhoof Vorrac": mrdCopperhoofVorracOracle, "Crystal Shard": mrdCrystalShardOracle,
		"Culling Scales": mrdCullingScalesOracle, "Damping Matrix": mrdDampingMatrixOracle,
		"Dead-Iron Sledge": mrdDeadIronSledgeOracle, "Deconstruct": mrdDeconstructOracle,
		"Disarm": mrdDisarmOracle, "Domineer": mrdDomineerOracle,
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

// mrd2Catalog puts a catalog permanent with a mana cost and a size on the
// battlefield, able to act this turn.
func mrd2Catalog(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func mrd2Size(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := layeredCard(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// --- Altar of Shadows -------------------------------------------------------

func TestMrdAltarOfShadowsDestroysThenBanksBlackMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	altar := mrd2Catalog(g, me.ID, "Altar of Shadows", "Artifact", mrdAltarOfShadowsOracle, "{7}", 0, 0)
	bear := seedBear(g, opp.ID)
	advanceToMain(t, g)

	b06AddMana(me, "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, altar, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the target creature was not destroyed")
	}
	if n := counterCount(g, altar, game.CounterCharge); n != 1 {
		t.Fatalf("%d charge counters, want 1", n)
	}

	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := me.ManaPool.CountColor("B"); n != 1 {
		t.Errorf("{B} in pool at the first main phase = %d, want 1 (one charge counter)", n)
	}
}

// --- Auriok Steelshaper -----------------------------------------------------

func TestMrdAuriokSteelshaperDiscountsEquipAndPumpsSoldiersAndKnightsWhileEquipped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	shaper := mrd2Catalog(g, me.ID, "Auriok Steelshaper", "Creature — Human Soldier", mrdAuriokSteelshaperOracle, "{1}{W}", 1, 1)
	knight := mrdPermanent(g, me.ID, "Squire", "Creature — Human Knight", "{1}{W}", 2, 2)
	bear := seedBear(g, me.ID)
	theirSoldier := mrdPermanent(g, opp.ID, "Their Soldier", "Creature — Human Soldier", "{W}", 1, 1)
	blade := seedEquipment(g, me.ID, "Banshee's Blade", mrdBansheesBladeOracle)

	if got := equipPrice(t, g, me.ID, blade, bear); got != "{1}" {
		t.Errorf("equip costs %s, want {1} (Equip {2}, one less)", got)
	}
	if p, tt := mrd2Size(t, g, knight); p != 2 || tt != 2 {
		t.Errorf("unequipped Steelshaper: the Knight is %d/%d, want 2/2", p, tt)
	}

	equipTo(t, g, me.ID, blade, shaper)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		p, t int
	}{
		{"the Steelshaper itself (a Soldier)", shaper, 2, 2},
		{"a Knight you control", knight, 3, 3},
		{"a Bear you control", bear, 2, 2},
		{"an opponent's Soldier", theirSoldier, 1, 1},
	} {
		if p, tt := mrd2Size(t, g, c.id); p != c.p || tt != c.t {
			t.Errorf("equipped Steelshaper: %s is %d/%d, want %d/%d", c.name, p, tt, c.p, c.t)
		}
	}

	equipTo(t, g, me.ID, blade, bear)
	if p, tt := mrd2Size(t, g, knight); p != 2 || tt != 2 {
		t.Errorf("the Equipment moved away: the Knight is %d/%d, want 2/2", p, tt)
	}
}

// --- Banshee's Blade --------------------------------------------------------

func TestMrdBansheesBladeBanksAChargeCounterOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Banshee's Blade", mrdBansheesBladeOracle)
	equipTo(t, g, me.ID, blade, bear)
	if p, tt := mrd2Size(t, g, bear); p != 2 || tt != 2 {
		t.Fatalf("with no counters the bear is %d/%d, want 2/2", p, tt)
	}

	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, blade, game.CounterCharge); n != 1 {
		t.Fatalf("%d charge counters after the hit, want 1", n)
	}
	if p, tt := mrd2Size(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("with one counter the bear is %d/%d, want 3/3", p, tt)
	}
}

func TestMrdBansheesBladeCountsDamageToABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Banshee's Blade", mrdBansheesBladeOracle)
	equipTo(t, g, me.ID, blade, bear)
	wall := mrdPermanent(g, opp.ID, "Wall", "Creature — Wall", "{1}", 0, 4)

	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, blade, game.CounterCharge); n != 1 {
		t.Errorf("%d charge counters after damaging a blocker, want 1", n)
	}
}

// --- Blinkmoth Urn ----------------------------------------------------------

func TestMrdBlinkmothUrnGivesThatPlayerColorlessPerArtifactWhileUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, second, third := g.Seats[0], g.Seats[1], g.Seats[2]
	urn := mrd2Catalog(g, me.ID, "Blinkmoth Urn", "Artifact", mrdBlinkmothUrnOracle, "{5}", 0, 0)
	for i := 0; i < 2; i++ {
		mrdPermanent(g, second.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	}
	mrdPermanent(g, third.ID, "Their Rock", "Artifact", "{2}", 0, 0)

	advanceToPrecombatMainOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := second.ManaPool.CountColor("C"); n != 2 {
		t.Errorf("the second seat's first main phase: {C} = %d, want 2 (its two artifacts)", n)
	}
	if n := me.ManaPool.CountColor("C"); n != 0 {
		t.Errorf("the Urn's controller got %d {C}; the mana is that player's", n)
	}

	findBattlefieldCardForTest(g, urn).Tapped = true
	advanceToPrecombatMainOf(t, g, 2)
	passPriorityAroundTable(t, g)
	if n := third.ManaPool.CountColor("C"); n != 0 {
		t.Errorf("a tapped Urn gave %d {C}, want none", n)
	}
}

// --- Cloudpost --------------------------------------------------------------

func TestMrdCloudpostEntersTappedAndCountsEveryLocus(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	post := mrdEnter(t, g, me, "Cloudpost", "Land — Locus", mrdCloudpostOracle)
	if !tappedForTest(t, g, post) {
		t.Error("Cloudpost should enter tapped")
	}
	mrdPermanent(g, opp.ID, "Their Locus", "Land — Locus", "", 0, 0)
	mrdPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	findBattlefieldCardForTest(g, post).Tapped = false
	advanceToMain(t, g)

	if err := g.ActivateManaAbility(me.ID, post, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if n := me.ManaPool.CountColor("C"); n != 2 {
		t.Errorf("{C} = %d, want 2 (itself and the opponent's Locus)", n)
	}
}

// --- Contaminated Bond ------------------------------------------------------

func TestMrdContaminatedBondDrainsTheEnchantedCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)
	auraCast(t, g, "Contaminated Bond", mrdContaminatedBondOracle, mine)
	auraCast(t, g, "Contaminated Bond", mrdContaminatedBondOracle, theirs)
	myLife, theirLife := me.Life, opp.Life

	declareAttack(t, g, opp.ID, mine)
	passPriorityAroundTable(t, g)
	if me.Life != myLife-3 {
		t.Errorf("attacking: my life %d, want %d", me.Life, myLife-3)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirs, mine); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife-3 {
		t.Errorf("blocking: the blocker's controller at %d, want %d (the Aura's controller is not the loser)", opp.Life, theirLife-3)
	}
	if me.Life != myLife-3 {
		t.Errorf("my life %d, want %d: only the block triggered the second time", me.Life, myLife-3)
	}
}

// --- Copperhoof Vorrac ------------------------------------------------------

func TestMrdCopperhoofVorracCountsUntappedPermanentsOpponentsControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	vorrac := mrd2Catalog(g, me.ID, "Copperhoof Vorrac", "Creature — Boar Beast", mrdCopperhoofVorracOracle, "{3}{G}{G}", 2, 2)
	mrdPermanent(g, me.ID, "My Forest", "Basic Land — Forest", "", 0, 0)
	land := mrdPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	mrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	tapped := mrdPermanent(g, third.ID, "Their Tapped Rock", "Artifact", "{2}", 0, 0)
	findBattlefieldCardForTest(g, tapped).Tapped = true
	mrdPermanent(g, third.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)

	if p, tt := mrd2Size(t, g, vorrac); p != 5 || tt != 5 {
		t.Errorf("Vorrac is %d/%d, want 5/5 (three untapped permanents across two opponents)", p, tt)
	}
	g.WithWriteLock(func() {
		if err := g.TapTargetForEffect(land); err != nil {
			t.Fatal(err)
		}
	})
	if p, tt := mrd2Size(t, g, vorrac); p != 4 || tt != 4 {
		t.Errorf("after an opponent's land taps, Vorrac is %d/%d, want 4/4", p, tt)
	}
}

// --- Crystal Shard ----------------------------------------------------------

func TestMrdCrystalShardBouncesUnlessTheControllerPays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shard := mrd2Catalog(g, me.ID, "Crystal Shard", "Artifact", mrdCrystalShardOracle, "{3}", 0, 0)
	bear := seedBear(g, opp.ID)
	other := seedBear(g, opp.ID)
	advanceToMain(t, g)

	b06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, shard, 1, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate the {U} row: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, opp.ID, false)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("declining: the creature should be in its owner's hand")
	}

	findBattlefieldCardForTest(g, shard).Tapped = false
	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shard, 0, game.ActivateAbilityParams{Targets: cardRefs(other), Strict: true}); err != nil {
		t.Fatalf("activate the {3} row: %v", err)
	}
	passPriorityAroundTable(t, g)
	b06AddMana(opp, "C")
	answerPayUnless(t, g, opp.ID, true)
	if !g.Battlefield.Contains(other) {
		t.Error("paying {1}: the creature should stay on the battlefield")
	}
}

// --- Culling Scales ---------------------------------------------------------

func TestMrdCullingScalesDestroysTheCheapestNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	owner, opp := g.Seats[1], g.Seats[2]
	mrd2Catalog(g, owner.ID, "Culling Scales", "Artifact", mrdCullingScalesOracle, "{3}", 0, 0)
	cheap := mrdPermanent(g, opp.ID, "Their Trinket", "Artifact", "{1}", 0, 0)
	tied := mrdPermanent(g, owner.ID, "My Trinket", "Artifact", "{1}", 0, 0)
	bear := mrdPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	mrdPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)

	advanceToUpkeepOf(t, g, 1)
	pick := latestPickTarget(g, owner.ID)
	if pick == nil {
		t.Fatal("no target prompt for the upkeep trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, owner.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a permanent with a higher mana value was accepted as the target")
	}
	pickCard(t, g, owner.ID, cheap)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) || !opp.Graveyard.Contains(cheap) {
		t.Error("the mana value 1 permanent was not destroyed")
	}
	if !g.Battlefield.Contains(tied) || !g.Battlefield.Contains(bear) {
		t.Error("only the target is destroyed")
	}
}

// --- Damping Matrix ---------------------------------------------------------

func TestMrdDampingMatrixStopsArtifactAndCreatureAbilitiesButNotManaAbilities(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Damping Matrix", "Artifact", mrdDampingMatrixOracle, 0, 0)
	krenko := actPushKrenko(g, opp.ID)
	golem := b21Push(g, me.ID, "Cobalt Golem", "Artifact Creature — Golem", mrdCobaltGolemOracle, 2, 3)
	solRing := b21Push(g, me.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)
	birds := b21Push(g, opp.ID, "Birds of Paradise", "Creature — Bird", actBirdsOracle, 0, 1, "G")
	well := b21Push(g, me.ID, "Blinkmoth Well", "Land", mrdBlinkmothWellOracle, 0, 0)
	rock := mrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)

	if err := g.ActivateCatalogAbility(opp.ID, krenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("a creature's {T} ability: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the Matrix's controller's own artifact creature: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateManaAbility(me.ID, solRing, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("an artifact's mana ability: err = %v, want nil", err)
	}
	if err := g.ActivateManaAbility(opp.ID, birds, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("a creature's mana ability: err = %v, want nil", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, well, 0, game.ActivateAbilityParams{Targets: cardRefs(rock)}); err != nil {
		t.Errorf("a land's non-mana ability: err = %v, want nil", err)
	}
}

// --- Dead-Iron Sledge -------------------------------------------------------

func TestMrdDeadIronSledgeDestroysBothCreaturesOfTheBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	sledge := seedEquipment(g, me.ID, "Dead-Iron Sledge", mrdDeadIronSledgeOracle)
	equipTo(t, g, me.ID, sledge, bear)
	wall := mrdPermanent(g, opp.ID, "Wall", "Creature — Wall", "{1}", 0, 4)
	bystander := seedBear(g, opp.ID)

	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	if n := triggersOnStackFrom(g, sledge); n != 1 {
		t.Fatalf("%d Sledge triggers on the stack, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("the equipped creature should be destroyed")
	}
	if g.Battlefield.Contains(wall) || !opp.Graveyard.Contains(wall) {
		t.Error("the blocker should be destroyed")
	}
	if !g.Battlefield.Contains(bystander) || !g.Battlefield.Contains(sledge) {
		t.Error("only the two creatures of the block are destroyed; the Sledge stays")
	}
}

// --- Deconstruct ------------------------------------------------------------

func TestMrdDeconstructDestroysAnArtifactAndAddsThreeGreen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := mrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	bear := seedBear(g, opp.ID)

	if err := castCatalogSpellErr(t, g, "Deconstruct", "Sorcery", mrdDeconstructOracle, cardRefs(bear)); err == nil {
		t.Error("a nonartifact creature was accepted as the target")
	}
	castCatalogSpell(t, g, "Deconstruct", "Sorcery", mrdDeconstructOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if n := me.ManaPool.CountColor("G"); n != 3 {
		t.Errorf("{G} in pool = %d, want 3", n)
	}
}

// --- Disarm -----------------------------------------------------------------

func TestMrdDisarmUnattachesEveryEquipmentFromTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, opp.ID)
	mine := seedBear(g, me.ID)
	blade := seedEquipment(g, opp.ID, "Banshee's Blade", mrdBansheesBladeOracle)
	sledge := seedEquipment(g, opp.ID, "Dead-Iron Sledge", mrdDeadIronSledgeOracle)
	myBlade := seedEquipment(g, me.ID, "Banshee's Blade", mrdBansheesBladeOracle)
	for _, eq := range []uuid.UUID{blade, sledge} {
		findBattlefieldCardForTest(g, eq).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
	}
	equipTo(t, g, me.ID, myBlade, mine)

	castCatalogSpell(t, g, "Disarm", "Instant", mrdDisarmOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	for name, eq := range map[string]uuid.UUID{"Banshee's Blade": blade, "Dead-Iron Sledge": sledge} {
		if !g.Battlefield.Contains(eq) {
			t.Errorf("%s left the battlefield; it should only come off", name)
			continue
		}
		if host := attachmentHostOf(t, g, eq); host.ID != uuid.Nil {
			t.Errorf("%s is still attached to %s", name, host.ID)
		}
	}
	if host := attachmentHostOf(t, g, myBlade); host.ID != mine {
		t.Error("Equipment on another creature should stay attached")
	}
}

// --- Domineer ---------------------------------------------------------------

func TestMrdDomineerStealsAnArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	golem := mrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{4}", 3, 3)
	bear := seedBear(g, opp.ID)

	if err := castCatalogSpellErr(t, g, "Domineer", auraTypeLine, mrdDomineerOracle, cardRefs(bear)); err == nil {
		t.Error("a nonartifact creature was accepted as the enchant target")
	}
	auraCast(t, g, "Domineer", mrdDomineerOracle, golem)
	if got := controllerOf(t, g, golem); got != me.ID {
		t.Errorf("the enchanted artifact creature is controlled by %s, want the Aura's controller", got)
	}
}
