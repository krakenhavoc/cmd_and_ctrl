package testsm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMrdBatch2CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Altar of Shadows": effectstest.MrdAltarOfShadowsOracle, "Auriok Steelshaper": effectstest.MrdAuriokSteelshaperOracle,
		"Banshee's Blade": effectstest.MrdBansheesBladeOracle, "Blinkmoth Urn": effectstest.MrdBlinkmothUrnOracle,
		"Cloudpost": effectstest.MrdCloudpostOracle, "Contaminated Bond": effectstest.MrdContaminatedBondOracle,
		"Copperhoof Vorrac": effectstest.MrdCopperhoofVorracOracle, "Crystal Shard": effectstest.MrdCrystalShardOracle,
		"Culling Scales": effectstest.MrdCullingScalesOracle, "Damping Matrix": effectstest.MrdDampingMatrixOracle,
		"Dead-Iron Sledge": effectstest.MrdDeadIronSledgeOracle, "Deconstruct": effectstest.MrdDeconstructOracle,
		"Disarm": effectstest.MrdDisarmOracle, "Domineer": effectstest.MrdDomineerOracle,
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

// --- Altar of Shadows -------------------------------------------------------

func TestMrdAltarOfShadowsDestroysThenBanksBlackMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	altar := effectstest.Mrd2Catalog(g, me.ID, "Altar of Shadows", "Artifact", effectstest.MrdAltarOfShadowsOracle, "{7}", 0, 0)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToMain(t, g)

	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, altar, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the target creature was not destroyed")
	}
	if n := effectstest.CounterCount(g, altar, game.CounterCharge); n != 1 {
		t.Fatalf("%d charge counters, want 1", n)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := me.ManaPool.CountColor("B"); n != 1 {
		t.Errorf("{B} in pool at the first main phase = %d, want 1 (one charge counter)", n)
	}
}

// --- Auriok Steelshaper -----------------------------------------------------

func TestMrdAuriokSteelshaperDiscountsEquipAndPumpsSoldiersAndKnightsWhileEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shaper := effectstest.Mrd2Catalog(g, me.ID, "Auriok Steelshaper", "Creature — Human Soldier", effectstest.MrdAuriokSteelshaperOracle, "{1}{W}", 1, 1)
	knight := effectstest.MrdPermanent(g, me.ID, "Squire", "Creature — Human Knight", "{1}{W}", 2, 2)
	bear := effectstest.SeedBear(g, me.ID)
	theirSoldier := effectstest.MrdPermanent(g, opp.ID, "Their Soldier", "Creature — Human Soldier", "{W}", 1, 1)
	blade := effectstest.SeedEquipment(g, me.ID, "Banshee's Blade", effectstest.MrdBansheesBladeOracle)

	if got := effectstest.EquipPrice(t, g, me.ID, blade, bear); got != "{1}" {
		t.Errorf("equip costs %s, want {1} (Equip {2}, one less)", got)
	}
	if p, tt := effectstest.Mrd2Size(t, g, knight); p != 2 || tt != 2 {
		t.Errorf("unequipped Steelshaper: the Knight is %d/%d, want 2/2", p, tt)
	}

	effectstest.EquipTo(t, g, me.ID, blade, shaper)
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
		if p, tt := effectstest.Mrd2Size(t, g, c.id); p != c.p || tt != c.t {
			t.Errorf("equipped Steelshaper: %s is %d/%d, want %d/%d", c.name, p, tt, c.p, c.t)
		}
	}

	effectstest.EquipTo(t, g, me.ID, blade, bear)
	if p, tt := effectstest.Mrd2Size(t, g, knight); p != 2 || tt != 2 {
		t.Errorf("the Equipment moved away: the Knight is %d/%d, want 2/2", p, tt)
	}
}

// --- Banshee's Blade --------------------------------------------------------

func TestMrdBansheesBladeBanksAChargeCounterOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Banshee's Blade", effectstest.MrdBansheesBladeOracle)
	effectstest.EquipTo(t, g, me.ID, blade, bear)
	if p, tt := effectstest.Mrd2Size(t, g, bear); p != 2 || tt != 2 {
		t.Fatalf("with no counters the bear is %d/%d, want 2/2", p, tt)
	}

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, blade, game.CounterCharge); n != 1 {
		t.Fatalf("%d charge counters after the hit, want 1", n)
	}
	if p, tt := effectstest.Mrd2Size(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("with one counter the bear is %d/%d, want 3/3", p, tt)
	}
}

func TestMrdBansheesBladeCountsDamageToABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Banshee's Blade", effectstest.MrdBansheesBladeOracle)
	effectstest.EquipTo(t, g, me.ID, blade, bear)
	wall := effectstest.MrdPermanent(g, opp.ID, "Wall", "Creature — Wall", "{1}", 0, 4)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, blade, game.CounterCharge); n != 1 {
		t.Errorf("%d charge counters after damaging a blocker, want 1", n)
	}
}

// --- Blinkmoth Urn ----------------------------------------------------------

func TestMrdBlinkmothUrnGivesThatPlayerColorlessPerArtifactWhileUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, second, third := g.Seats[0], g.Seats[1], g.Seats[2]
	urn := effectstest.Mrd2Catalog(g, me.ID, "Blinkmoth Urn", "Artifact", effectstest.MrdBlinkmothUrnOracle, "{5}", 0, 0)
	for i := 0; i < 2; i++ {
		effectstest.MrdPermanent(g, second.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	}
	effectstest.MrdPermanent(g, third.ID, "Their Rock", "Artifact", "{2}", 0, 0)

	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := second.ManaPool.CountColor("C"); n != 2 {
		t.Errorf("the second seat's first main phase: {C} = %d, want 2 (its two artifacts)", n)
	}
	if n := me.ManaPool.CountColor("C"); n != 0 {
		t.Errorf("the Urn's controller got %d {C}; the mana is that player's", n)
	}

	effectstest.FindBattlefieldCardForTest(g, urn).Tapped = true
	effectstest.AdvanceToPrecombatMainOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := third.ManaPool.CountColor("C"); n != 0 {
		t.Errorf("a tapped Urn gave %d {C}, want none", n)
	}
}

// --- Cloudpost --------------------------------------------------------------

func TestMrdCloudpostEntersTappedAndCountsEveryLocus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	post := effectstest.MrdEnter(t, g, me, "Cloudpost", "Land — Locus", effectstest.MrdCloudpostOracle)
	if !effectstest.TappedForTest(t, g, post) {
		t.Error("Cloudpost should enter tapped")
	}
	effectstest.MrdPermanent(g, opp.ID, "Their Locus", "Land — Locus", "", 0, 0)
	effectstest.MrdPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.FindBattlefieldCardForTest(g, post).Tapped = false
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateManaAbility(me.ID, post, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if n := me.ManaPool.CountColor("C"); n != 2 {
		t.Errorf("{C} = %d, want 2 (itself and the opponent's Locus)", n)
	}
}

// --- Contaminated Bond ------------------------------------------------------

func TestMrdContaminatedBondDrainsTheEnchantedCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Contaminated Bond", effectstest.MrdContaminatedBondOracle, mine)
	effectstest.AuraCast(t, g, "Contaminated Bond", effectstest.MrdContaminatedBondOracle, theirs)
	myLife, theirLife := me.Life, opp.Life

	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-3 {
		t.Errorf("attacking: my life %d, want %d", me.Life, myLife-3)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirs, mine); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-3 {
		t.Errorf("blocking: the blocker's controller at %d, want %d (the Aura's controller is not the loser)", opp.Life, theirLife-3)
	}
	if me.Life != myLife-3 {
		t.Errorf("my life %d, want %d: only the block triggered the second time", me.Life, myLife-3)
	}
}

// --- Copperhoof Vorrac ------------------------------------------------------

func TestMrdCopperhoofVorracCountsUntappedPermanentsOpponentsControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	vorrac := effectstest.Mrd2Catalog(g, me.ID, "Copperhoof Vorrac", "Creature — Boar Beast", effectstest.MrdCopperhoofVorracOracle, "{3}{G}{G}", 2, 2)
	effectstest.MrdPermanent(g, me.ID, "My Forest", "Basic Land — Forest", "", 0, 0)
	land := effectstest.MrdPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	tapped := effectstest.MrdPermanent(g, third.ID, "Their Tapped Rock", "Artifact", "{2}", 0, 0)
	effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true
	effectstest.MrdPermanent(g, third.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)

	if p, tt := effectstest.Mrd2Size(t, g, vorrac); p != 5 || tt != 5 {
		t.Errorf("Vorrac is %d/%d, want 5/5 (three untapped permanents across two opponents)", p, tt)
	}
	g.WithWriteLock(func() {
		if err := g.TapTargetForEffect(land); err != nil {
			t.Fatal(err)
		}
	})
	if p, tt := effectstest.Mrd2Size(t, g, vorrac); p != 4 || tt != 4 {
		t.Errorf("after an opponent's land taps, Vorrac is %d/%d, want 4/4", p, tt)
	}
}

// --- Crystal Shard ----------------------------------------------------------

func TestMrdCrystalShardBouncesUnlessTheControllerPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shard := effectstest.Mrd2Catalog(g, me.ID, "Crystal Shard", "Artifact", effectstest.MrdCrystalShardOracle, "{3}", 0, 0)
	bear := effectstest.SeedBear(g, opp.ID)
	other := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToMain(t, g)

	effectstest.B06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, shard, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate the {U} row: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("declining: the creature should be in its owner's hand")
	}

	effectstest.FindBattlefieldCardForTest(g, shard).Tapped = false
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shard, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other), Strict: true}); err != nil {
		t.Fatalf("activate the {3} row: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(opp, "C")
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	if !g.Battlefield.Contains(other) {
		t.Error("paying {1}: the creature should stay on the battlefield")
	}
}

// --- Culling Scales ---------------------------------------------------------

func TestMrdCullingScalesDestroysTheCheapestNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, opp := g.Seats[1], g.Seats[2]
	effectstest.Mrd2Catalog(g, owner.ID, "Culling Scales", "Artifact", effectstest.MrdCullingScalesOracle, "{3}", 0, 0)
	cheap := effectstest.MrdPermanent(g, opp.ID, "Their Trinket", "Artifact", "{1}", 0, 0)
	tied := effectstest.MrdPermanent(g, owner.ID, "My Trinket", "Artifact", "{1}", 0, 0)
	bear := effectstest.MrdPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	effectstest.MrdPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	pick := effectstest.LatestPickTarget(g, owner.ID)
	if pick == nil {
		t.Fatal("no target prompt for the upkeep trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, owner.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a permanent with a higher mana value was accepted as the target")
	}
	effectstest.PickCard(t, g, owner.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) || !opp.Graveyard.Contains(cheap) {
		t.Error("the mana value 1 permanent was not destroyed")
	}
	if !g.Battlefield.Contains(tied) || !g.Battlefield.Contains(bear) {
		t.Error("only the target is destroyed")
	}
}

// --- Damping Matrix ---------------------------------------------------------

func TestMrdDampingMatrixStopsArtifactAndCreatureAbilitiesButNotManaAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Damping Matrix", "Artifact", effectstest.MrdDampingMatrixOracle, 0, 0)
	krenko := effectstest.ActPushKrenko(g, opp.ID)
	golem := effectstest.B21Push(g, me.ID, "Cobalt Golem", "Artifact Creature — Golem", effectstest.MrdCobaltGolemOracle, 2, 3)
	solRing := effectstest.B21Push(g, me.ID, "Sol Ring", "Artifact", effectstest.ActSolRingOracle, 0, 0)
	birds := effectstest.B21Push(g, opp.ID, "Birds of Paradise", "Creature — Bird", effectstest.ActBirdsOracle, 0, 1, "G")
	well := effectstest.B21Push(g, me.ID, "Blinkmoth Well", "Land", effectstest.MrdBlinkmothWellOracle, 0, 0)
	rock := effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)

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
	if err := g.ActivateCatalogAbility(me.ID, well, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock)}); err != nil {
		t.Errorf("a land's non-mana ability: err = %v, want nil", err)
	}
}

// --- Dead-Iron Sledge -------------------------------------------------------

func TestMrdDeadIronSledgeDestroysBothCreaturesOfTheBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	sledge := effectstest.SeedEquipment(g, me.ID, "Dead-Iron Sledge", effectstest.MrdDeadIronSledgeOracle)
	effectstest.EquipTo(t, g, me.ID, sledge, bear)
	wall := effectstest.MrdPermanent(g, opp.ID, "Wall", "Creature — Wall", "{1}", 0, 4)
	bystander := effectstest.SeedBear(g, opp.ID)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if n := effectstest.TriggersOnStackFrom(g, sledge); n != 1 {
		t.Fatalf("%d Sledge triggers on the stack, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	bear := effectstest.SeedBear(g, opp.ID)

	if err := effectstest.CastCatalogSpellErr(t, g, "Deconstruct", "Sorcery", effectstest.MrdDeconstructOracle, effectstest.CardRefs(bear)); err == nil {
		t.Error("a nonartifact creature was accepted as the target")
	}
	effectstest.CastCatalogSpell(t, g, "Deconstruct", "Sorcery", effectstest.MrdDeconstructOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if n := me.ManaPool.CountColor("G"); n != 3 {
		t.Errorf("{G} in pool = %d, want 3", n)
	}
}

// --- Disarm -----------------------------------------------------------------

func TestMrdDisarmUnattachesEveryEquipmentFromTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, opp.ID)
	mine := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, opp.ID, "Banshee's Blade", effectstest.MrdBansheesBladeOracle)
	sledge := effectstest.SeedEquipment(g, opp.ID, "Dead-Iron Sledge", effectstest.MrdDeadIronSledgeOracle)
	myBlade := effectstest.SeedEquipment(g, me.ID, "Banshee's Blade", effectstest.MrdBansheesBladeOracle)
	for _, eq := range []uuid.UUID{blade, sledge} {
		effectstest.FindBattlefieldCardForTest(g, eq).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
	}
	effectstest.EquipTo(t, g, me.ID, myBlade, mine)

	effectstest.CastCatalogSpell(t, g, "Disarm", "Instant", effectstest.MrdDisarmOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	for name, eq := range map[string]uuid.UUID{"Banshee's Blade": blade, "Dead-Iron Sledge": sledge} {
		if !g.Battlefield.Contains(eq) {
			t.Errorf("%s left the battlefield; it should only come off", name)
			continue
		}
		if host := effectstest.AttachmentHostOf(t, g, eq); host.ID != uuid.Nil {
			t.Errorf("%s is still attached to %s", name, host.ID)
		}
	}
	if host := effectstest.AttachmentHostOf(t, g, myBlade); host.ID != mine {
		t.Error("Equipment on another creature should stay attached")
	}
}

// --- Domineer ---------------------------------------------------------------

func TestMrdDomineerStealsAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	golem := effectstest.MrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{4}", 3, 3)
	bear := effectstest.SeedBear(g, opp.ID)

	if err := effectstest.CastCatalogSpellErr(t, g, "Domineer", effectstest.AuraTypeLine, effectstest.MrdDomineerOracle, effectstest.CardRefs(bear)); err == nil {
		t.Error("a nonartifact creature was accepted as the enchant target")
	}
	effectstest.AuraCast(t, g, "Domineer", effectstest.MrdDomineerOracle, golem)
	if got := effectstest.ControllerOf(t, g, golem); got != me.ID {
		t.Errorf("the enchanted artifact creature is controlled by %s, want the Aura's controller", got)
	}
}
