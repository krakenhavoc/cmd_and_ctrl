package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprCarrionCallOracle     = "8f1e8f25-ee23-40a6-a3f2-369bb960267f"
	rprElspethTirelOracle    = "5f815e69-fdd6-4186-9317-25f5c7e0b08e"
	rprExoskeletonOracle     = "f4350c4e-7e53-4401-b889-14bed27bee9c"
	rprArchonOracle          = "ed8392b8-71e9-4874-a5bd-3c9fa5605780"
	rprHateMirageOracle      = "2db23ceb-448b-45d6-a931-d72e8ccdb541"
	rprIchorclawOracle       = "8462a2fa-944b-4fff-a161-10345690b56f"
	rprJorKadeenOracle       = "7c4dbaa8-e24f-4126-81fe-3f218f520d8c"
	rprMaceOracle            = "881c3dd4-49df-48c3-b514-509990332e25"
	rprMidnightHauntOracle   = "a79e21fc-89bd-4547-8bef-5510554e1431"
	rprMycosynthFiendOracle  = "e7c4bc25-806c-4a75-85bf-a0e66308d5d8"
	rprPhantomGeneralOracle  = "b800c3e5-4ed9-4603-a987-d830e90cf25e"
	rprArtifactTypeLine      = "Artifact"
	rprCreatureTypeLineTest  = "Creature — Test"
	rprInfectKeywordForCheck = "infect"
)

func TestRPRBatch11AreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprCarrionCallOracle: "Carrion Call", rprElspethTirelOracle: "Elspeth Tirel",
		rprExoskeletonOracle: "Grafted Exoskeleton", rprArchonOracle: "Harmonious Archon",
		rprHateMirageOracle: "Hate Mirage", rprIchorclawOracle: "Ichorclaw Myr",
		rprJorKadeenOracle: "Jor Kadeen, the Prevailer", rprMaceOracle: "Mace of the Valiant",
		rprMidnightHauntOracle: "Midnight Haunting", rprMycosynthFiendOracle: "Mycosynth Fiend",
		rprPhantomGeneralOracle: "Phantom General",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registration %+v", name, spec)
		}
	}
}

func TestCarrionCallMakesTwoInfectInsects(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Carrion Call", "Instant", rprCarrionCallOracle, nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Insect" {
			n++
			if !c.IsToken() || c.Power != 1 || c.Toughness != 1 || !hasKeywordOf(g, c.InstanceID, "infect") {
				t.Errorf("Insect = %+v, want a 1/1 token with infect", c)
			}
		}
	}
	if n != 2 {
		t.Fatalf("%d Insects, want 2", n)
	}
}

func TestMidnightHauntingMakesTwoFlyingSpirits(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Midnight Haunting", "Instant", rprMidnightHauntOracle, nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" {
			n++
			if !c.IsToken() || !hasKeywordOf(g, c.InstanceID, "flying") {
				t.Errorf("Spirit = %+v, want a flying token", c)
			}
		}
	}
	if n != 2 {
		t.Fatalf("%d Spirits, want 2", n)
	}
}

func TestPhantomGeneralPumpsOnlyYourCreatureTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	general := pushCatalogPermanent(g, me.ID, "Phantom General", "Creature — Spirit Soldier", rprPhantomGeneralOracle, false)
	g.WithWriteLock(func() {
		g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Power = 2
		g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Toughness = 3
	})
	bear := seedBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 1) })
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 white Spirit with flying"), 1) })
	soldier := findBattlefieldByName(g, "Soldier")
	spirit := findBattlefieldByName(g, "Spirit")
	if p, tough := effectivePTOf(t, g, soldier); p != 2 || tough != 2 {
		t.Errorf("my token = %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, spirit); p != 1 || tough != 1 {
		t.Errorf("their token = %d/%d, want 1/1", p, tough)
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("nontoken Bear power %d, want 2", p)
	}
	if p, tough := effectivePTOf(t, g, general); p != 2 || tough != 3 {
		t.Errorf("the General = %d/%d, want 2/3", p, tough)
	}
}

func TestHarmoniousArchonSetsNonArchonsToThreeAndMakesHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	big := tdCreature(g, opp.ID, "Big", 6, 6)
	bear := seedBear(g, me.ID)
	equipTo(t, g, me.ID, seedEquipment(g, me.ID, "Bonesplitter", bonesplitterOracle), bear)
	id := handCardForTest(me, "Harmonious Archon", "Creature — Archon", rprArchonOracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 4, 5
			}
		}
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, id); p != 4 || tough != 5 {
		t.Errorf("Archon = %d/%d, want 4/5", p, tough)
	}
	if p, tough := effectivePTOf(t, g, big); p != 3 || tough != 3 {
		t.Errorf("their 6/6 = %d/%d, want base 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("my equipped Bear = %d/%d, want base 3/3 plus Bonesplitter's +2/+0", p, tough)
	}
	humans := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Human" && c.IsToken() {
			humans++
			if p, tough := effectivePTOf(t, g, c.InstanceID); p != 3 || tough != 3 {
				t.Errorf("Human = %d/%d, want 3/3 under the Archon", p, tough)
			}
		}
	}
	if humans != 2 {
		t.Errorf("%d Humans, want 2", humans)
	}
}

func TestIchorclawMyrGetsPumpedWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myr := b12Push(g, me.ID, "Ichorclaw Myr", "Artifact Creature — Phyrexian Myr", rprIchorclawOracle, 1, 1)
	blocker := tdCreature(g, opp.ID, "Blocker", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(myr, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if !hasKeywordOf(g, myr, rprInfectKeywordForCheck) {
		t.Error("Ichorclaw Myr lacks infect")
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, myr); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, myr); p != 3 || tough != 3 {
		t.Errorf("blocked Myr = %d/%d, want 3/3", p, tough)
	}
}

func TestJorKadeenPrevailerNeedsMetalcraft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jor := b12Push(g, me.ID, "Jor Kadeen, the Prevailer", "Legendary Creature — Human Warrior", rprJorKadeenOracle, 5, 4)
	bear := seedBear(g, me.ID)
	theirs := tdCreature(g, opp.ID, "Theirs", 2, 2)
	a1 := seedArtifactForRPR(g, me.ID, "Rock A")
	seedArtifactForRPR(g, me.ID, "Rock B")
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("two artifacts: Bear power %d, want 2", p)
	}
	seedArtifactForRPR(g, me.ID, "Rock C")
	if p, _ := effectivePTOf(t, g, bear); p != 5 {
		t.Errorf("metalcraft: Bear power %d, want 5", p)
	}
	if p, _ := effectivePTOf(t, g, jor); p != 8 {
		t.Errorf("metalcraft: Jor Kadeen power %d, want 8", p)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("their creature power %d, want 2", p)
	}
	if !hasKeywordOf(g, jor, "first strike") {
		t.Error("no first strike")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, a1); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("after an artifact left: Bear power %d, want 2", p)
	}
}

func TestMycosynthFiendCountsEveryOpponentsPoison(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	fiend := b12Push(g, me.ID, "Mycosynth Fiend", "Creature — Phyrexian Horror", rprMycosynthFiendOracle, 2, 2)
	if p, _ := effectivePTOf(t, g, fiend); p != 2 {
		t.Fatalf("no poison: power %d, want 2", p)
	}
	for _, s := range []struct {
		id uuid.UUID
		n  int
	}{{opp.ID, 2}, {third.ID, 3}, {me.ID, 4}} {
		if err := g.SetPoison(s.id, s.n); err != nil {
			t.Fatalf("SetPoison: %v", err)
		}
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if p, tough := effectivePTOf(t, g, fiend); p != 7 || tough != 7 {
		t.Errorf("Fiend = %d/%d, want 7/7 (2+3 opponents' poison, not mine)", p, tough)
	}
}

func TestMaceOfTheValiantGrowsPerCreatureAndGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mace := seedEquipment(g, me.ID, "Mace of the Valiant", rprMaceOracle)
	bear := seedBear(g, me.ID)
	equipTo(t, g, me.ID, mace, bear)
	if !hasKeywordOf(g, bear, "vigilance") {
		t.Error("equipped creature lacks vigilance")
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("no counters: power %d, want 2", p)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 2) })
	passPriorityAroundTable(t, g)
	if n := counterCountOf(t, g, mace, "charge"); n != 2 {
		t.Fatalf("%d charge counters, want 2", n)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear = %d/%d, want 4/4", p, tough)
	}
	// An opponent's creature entering adds nothing.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 white Soldier"), 1) })
	passPriorityAroundTable(t, g)
	if n := counterCountOf(t, g, mace, "charge"); n != 2 {
		t.Errorf("%d charge counters after an opposing creature, want 2", n)
	}
}

func TestGraftedExoskeletonPumpsAndGrantsInfect(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	skel := seedEquipment(g, me.ID, "Grafted Exoskeleton", rprExoskeletonOracle)
	bear := seedBear(g, me.ID)
	equipTo(t, g, me.ID, skel, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear = %d/%d, want 4/4", p, tough)
	}
	if !hasKeywordOf(g, bear, "infect") {
		t.Error("equipped creature lacks infect")
	}
	if !onBattlefieldByID(g, bear) {
		t.Error("the host was sacrificed by simply being equipped")
	}
}

func TestGraftedExoskeletonSacrificesTheHostItMovesOffOf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	skel := seedEquipment(g, me.ID, "Grafted Exoskeleton", rprExoskeletonOracle)
	first, second := seedBear(g, me.ID), seedBear(g, me.ID)
	equipTo(t, g, me.ID, skel, first)
	equipTo(t, g, me.ID, skel, second)
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, first) {
		t.Error("the creature the Equipment moved off of was not sacrificed")
	}
	if !onBattlefieldByID(g, second) {
		t.Error("the new host was sacrificed")
	}
}

func TestGraftedExoskeletonSacrificesTheHostWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	skel := seedEquipment(g, me.ID, "Grafted Exoskeleton", rprExoskeletonOracle)
	bear := seedBear(g, me.ID)
	equipTo(t, g, me.ID, skel, bear)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, skel); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, bear) {
		t.Error("the host survived the Equipment leaving")
	}
}

func TestHateMirageCopiesTwoOpposingCreaturesWithHasteAndExilesThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := tdCreature(g, opp.ID, "Alpha", 3, 3)
	b := tdCreature(g, opp.ID, "Beta", 1, 4)
	mine := tdCreature(g, me.ID, "Mine", 2, 2)
	castCatalogSpell(t, g, "Hate Mirage", "Sorcery", rprHateMirageOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if (c.Name == "Alpha" || c.Name == "Beta") && c.InstanceID != a && c.InstanceID != b {
			copies++
			if !c.IsToken() || c.Controller != me.ID || !hasKeywordOf(g, c.InstanceID, "haste") {
				t.Errorf("copy %+v: want a hasty token under my control", c)
			}
		}
	}
	if copies != 2 {
		t.Fatalf("%d copies, want 2", copies)
	}
	// Not a creature the caster controls.
	g2 := newCatalogGame(t)
	own := tdCreature(g2, g2.Seats[0].ID, "Own", 2, 2)
	advanceTo(t, g2, game.StepPrecombatMain)
	id := handCardForTest(g2.Seats[0], "Hate Mirage", "Sorcery", rprHateMirageOracle)
	if err := g2.CastSpell(g2.Seats[0].ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(own)}); err == nil {
		t.Error("Hate Mirage accepted a creature its caster controls")
	}
	_ = mine
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && (c.Name == "Alpha" || c.Name == "Beta") {
			t.Errorf("copy %q survived the end step", c.Name)
		}
	}
	if !onBattlefieldByID(g, a) || !onBattlefieldByID(g, b) {
		t.Error("an original was exiled")
	}
}

func TestElspethTirelPlusTwoGainsLifePerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pw := pushWalkerForTest(g, me.ID, "Elspeth Tirel", rprElspethTirelOracle, 4)
	tdCreature(g, me.ID, "A", 1, 1)
	tdCreature(g, me.ID, "B", 1, 1)
	tdCreature(g, g.Seats[1].ID, "Theirs", 1, 1)
	advanceToMainOf(t, g, 0)
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d, want %d", me.Life, before+2)
	}
	if l := loyaltyOf(g, pw); l != 6 {
		t.Errorf("loyalty %d, want 6", l)
	}
}

func TestElspethTirelMinusTwoMakesThreeSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pw := pushWalkerForTest(g, me.ID, "Elspeth Tirel", rprElspethTirelOracle, 4)
	advanceToMainOf(t, g, 0)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Soldier"); n != 3 {
		t.Errorf("%d Soldiers, want 3", n)
	}
	if l := loyaltyOf(g, pw); l != 2 {
		t.Errorf("loyalty %d, want 2", l)
	}
}

func TestElspethTirelUltimateSparesLandsTokensAndHerself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pw := pushWalkerForTest(g, me.ID, "Elspeth Tirel", rprElspethTirelOracle, 5)
	bear := tdCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := tdCreature(g, me.ID, "My Bear", 2, 2)
	rock := seedPermanentFor(g, opp.ID, "Rock", rprArtifactTypeLine)
	land := seedPermanentFor(g, opp.ID, "Their Land", "Land")
	otherWalker := pushWalkerForTest(g, opp.ID, "Other Walker", "", 3)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 white Soldier"), 1) })
	token := findBattlefieldByName(g, "Soldier")
	advanceToMainOf(t, g, 0)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−5: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"their Bear": bear, "my Bear": mine, "the Rock": rock, "the other walker": otherWalker} {
		if onBattlefieldByID(g, id) {
			t.Errorf("%s survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"the land": land, "the token": token} {
		if !onBattlefieldByID(g, id) {
			t.Errorf("%s was destroyed", name)
		}
	}
	// She spent all five loyalty and goes to the graveyard as a state-based action.
	if onBattlefieldByID(g, pw) {
		t.Error("Elspeth stayed at 0 loyalty")
	}
}

// seedArtifactForRPR puts an artifact on the battlefield the way a permanent
// arrives, so the layer cache is invalidated.
func seedArtifactForRPR(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: rprArtifactTypeLine, Owner: owner, Controller: owner,
	})
}
