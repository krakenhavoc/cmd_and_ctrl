package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch43_test.go — card-level coverage for the card-coverage roadmap's
// batch 43 (#450, `edhrec_rank` 4456–4555). One test per observable
// behaviour, driven through a real cast, activation, attack or land
// play rather than by calling primitives directly.

const (
	b43GigantosaurusOracle      = "e666bae7-dd51-4921-8b89-7e8d423caba0"
	b43IndomitableAncientsOracl = "2764741c-1f3e-459a-a487-54cbac2c84a6"
	b43RaiseTheAlarmOracle      = "5b2364d7-a811-4595-a1b4-224c70555ffa"
	b43DreadboreOracle          = "d685799f-cc1a-40d6-9df6-d05b8b1f5b13"
	b43NyxFleeceRamOracle       = "5a20113c-7ea8-4edf-af9f-148ccd326a88"
	b43BaronOracle              = "cc710da0-5a2e-4bc4-8fdd-d90e7bc1f224"
	b43ThornglintBridgeOracle   = "99720c65-be96-4220-8ed4-720660bf6928"
	b43UrzasFactoryOracle       = "5a620d20-f14e-43d0-8e57-c2a197e2ec51"
	b43NayaPanoramaOracle       = "71e28800-c42c-48c0-95e5-0296be54a4e8"
	b43ShamanOfThePackOracle    = "39c7d2cc-4fd1-4f1a-be1a-a6ae56d2356d"
	b43SeafloorOracleOracle     = "f9f719cb-7778-4109-bbd3-4504ee1b5f49"
	b43NemesisOfReasonOracle    = "b1d70845-869b-45f7-b4bf-1549bac51d58"
	b43PeekOracle               = "873abbf4-c9b5-4b8f-8cd0-613cf3b9b1d5"
	b43CaptainSisayOracle       = "2e7a9ea9-f76c-4c12-950f-c613fa16cfa8"
	b43PlanarBridgeOracle       = "853c6cc0-ed2b-4d65-add1-2449ade8cf68"
	b43NightshadeHarvesterOracl = "875c3db1-2752-48e7-9664-bbb12120c032"
	b43DarienOracle             = "d05336cb-6157-47e7-942f-43becd67a5bf"
	b43ElectrickeryOracle       = "99fd4b51-1698-4168-9366-a6ea2df22361"
	b43AngelicFieldMarshalOracl = "502575f4-7c56-44bc-a77f-ae28d66c8e1f"
	b43BoonOfTheSpiritRealmOrac = "7731217e-cdbb-4c38-a9c1-083c96f51931"
	b43NoxiousGhoulOracle       = "602e421e-afab-4488-8a52-af66b4fdce4b"
	b43DistinguishedConjurerOra = "6df57d67-2fd9-4e7a-b67b-f361fc30e496"
	b43CeruleanWispsOracle      = "a6605c50-558e-417c-8c75-6c45b06d6e13"
	b43TowerWinderOracle        = "c1aecc8a-db3a-4158-b1f1-bfe06d782482"
	b43MooglesValorOracle       = "9627ed9a-5ade-412d-bf8d-f0705b1c9405"
	b43EowynOracle              = "98a2a642-a30a-43e7-b47f-4776c22d1cb0"
	b43DeeprootPilgrimageOracle = "188656d8-4ea8-4e58-88da-b010a16eb6f2"
	b43FinneasOracle            = "69924636-138e-4baa-a378-0fd7df5b847d"
	b43CalderaPyremawOracle     = "3eb5ae68-7f76-4874-80a4-80058b10fc7a"
	b43ZukosExileOracle         = "5cf18398-9552-415e-a61f-2c149e35ec3d"
	b43DragonlordAtarkaOracle   = "daff70a4-a990-47c9-b6ef-2c379272c8a3"
	b43MaelstromColossusOracle  = "06fc3eb1-7a48-42a0-8a5c-d61bd5aacfff"
	b43ZephyrBootsOracle        = "02f3b9c3-f611-4f05-ab3e-b296916efdad"
	b43ShieldOfTheOversoulOracl = "9b598025-80a3-4144-be9a-863165161594"
	b43DesertOracle             = "195107ad-879d-4b02-a44a-a3ba70fedf88"
	b43IfnirDeadlandsOracle     = "af698bd5-5f56-4d2a-9f02-8c3e781210cd"
	b43CryptcallerChariotOracle = "2bdd0bcb-6cfb-48e7-972e-355ff46621a4"
	b43ArchetypeOfCourageOracle = "79b48704-480d-4905-b87a-40b127894670"
	b43VizkopaGuildmageOracle   = "f19e7c5c-67fa-4ae4-89b8-afa0e08a6c48"
	b43RestartSequenceOracle    = "c415bfdd-3d42-4f94-9f93-7e310becbc8b"
	b43FracturedIdentityOracle  = "1515b0c2-1b55-4cd4-ad81-fb6b1f3e8188"
	b43FloodOfTearsOracle       = "3d92f2e6-27df-4329-b552-cf905f7616ba"
	b43AgonasaurRexOracle       = "1ad5766b-9ae3-437b-bd91-c4c988a8b095"
	b43InterplanarBeaconOracle  = "073169f2-da3a-4a93-8c01-b3fd8558d225"
	b43SkemfarShadowsageOracle  = "496ecb73-f508-4ac7-b101-23e03404580f"
	b43GimlisRecklessMightOracl = "824abca7-b5e2-4ddb-ac7d-b7d04758a3c9"
	b43LostJitteOracle          = "a9d72f78-2ab5-4e2e-ab7b-ef875e0a0609"
)

// --- fixtures ------------------------------------------------------

// b43Creature seeds a creature with a real battlefield timestamp, so
// the layer cache and the "entered this turn" tally both see it.
func b43Creature(g *game.Game, controller uuid.UUID, name, typeLine string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Colors: colors,
		Owner: controller, Controller: controller,
	})
}

// b43Catalog seeds a catalog permanent with a real timestamp and a
// real power / toughness — pushCatalogPermanent stamps 1/1 and no
// timestamp, which is wrong for anything the layer engine reads.
func b43Catalog(g *game.Game, controller uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: controller, Controller: controller,
	})
}

// b43Commander seeds a creature flagged as its owner's commander.
func b43Commander(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
	})
}

// b43TokensNamed counts battlefield tokens with the given name under
// one controller.
func b43TokensNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.IsToken() {
			n++
		}
	}
	return n
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Naya Panorama
// is a row in the Panorama cycle table, so a transposed row would be
// invisible until someone played that exact land.
func TestBatch43CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b43GigantosaurusOracle:      "Gigantosaurus",
		b43IndomitableAncientsOracl: "Indomitable Ancients",
		b43RaiseTheAlarmOracle:      "Raise the Alarm",
		b43DreadboreOracle:          "Dreadbore",
		b43NyxFleeceRamOracle:       "Nyx-Fleece Ram",
		b43BaronOracle:              "Baron, Airship Kingdom",
		b43ThornglintBridgeOracle:   "Thornglint Bridge",
		b43UrzasFactoryOracle:       "Urza's Factory",
		b43NayaPanoramaOracle:       "Naya Panorama",
		b43ShamanOfThePackOracle:    "Shaman of the Pack",
		b43SeafloorOracleOracle:     "Seafloor Oracle",
		b43NemesisOfReasonOracle:    "Nemesis of Reason",
		b43PeekOracle:               "Peek",
		b43CaptainSisayOracle:       "Captain Sisay",
		b43PlanarBridgeOracle:       "Planar Bridge",
		b43NightshadeHarvesterOracl: "Nightshade Harvester",
		b43DarienOracle:             "Darien, King of Kjeldor",
		b43ElectrickeryOracle:       "Electrickery",
		b43AngelicFieldMarshalOracl: "Angelic Field Marshal",
		b43BoonOfTheSpiritRealmOrac: "Boon of the Spirit Realm",
		b43NoxiousGhoulOracle:       "Noxious Ghoul",
		b43DistinguishedConjurerOra: "Distinguished Conjurer",
		b43CeruleanWispsOracle:      "Cerulean Wisps",
		b43TowerWinderOracle:        "Tower Winder",
		b43MooglesValorOracle:       "Moogles' Valor",
		b43EowynOracle:              "Éowyn, Shieldmaiden",
		b43DeeprootPilgrimageOracle: "Deeproot Pilgrimage",
		b43FinneasOracle:            "Finneas, Ace Archer",
		b43CalderaPyremawOracle:     "Caldera Pyremaw",
		b43ZukosExileOracle:         "Zuko's Exile",
		b43DragonlordAtarkaOracle:   "Dragonlord Atarka",
		b43MaelstromColossusOracle:  "Maelstrom Colossus",
		b43ZephyrBootsOracle:        "Zephyr Boots",
		b43ShieldOfTheOversoulOracl: "Shield of the Oversoul",
		b43DesertOracle:             "Desert",
		b43IfnirDeadlandsOracle:     "Ifnir Deadlands",
		b43CryptcallerChariotOracle: "Cryptcaller Chariot",
		b43ArchetypeOfCourageOracle: "Archetype of Courage",
		b43VizkopaGuildmageOracle:   "Vizkopa Guildmage",
		b43RestartSequenceOracle:    "Restart Sequence",
		b43FracturedIdentityOracle:  "Fractured Identity",
		b43FloodOfTearsOracle:       "Flood of Tears",
		b43AgonasaurRexOracle:       "Agonasaur Rex",
		b43InterplanarBeaconOracle:  "Interplanar Beacon",
		b43SkemfarShadowsageOracle:  "Skemfar Shadowsage",
		b43GimlisRecklessMightOracl: "Gimli's Reckless Might",
		b43LostJitteOracle:          "Lost Jitte",
	}
	if len(want) != 47 {
		t.Fatalf("the batch registers 47 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// The four vanilla / near-vanilla bodies are registered so they resolve
// automatically, and the two that print keywords declare them.
func TestB43VanillaBodiesDeclareOnlyWhatTheyPrint(t *testing.T) {
	for _, row := range []struct {
		name, oracle string
		keywords     []string
	}{
		{"Gigantosaurus", b43GigantosaurusOracle, nil},
		{"Indomitable Ancients", b43IndomitableAncientsOracl, nil},
		{"Agonasaur Rex", b43AgonasaurRexOracle, []string{"trample"}},
	} {
		spec, ok := Lookup(row.oracle)
		if !ok {
			t.Fatalf("%s is not registered", row.name)
		}
		if len(spec.PrintedKeywords) != len(row.keywords) {
			t.Errorf("%s: printed keywords %v, want %v", row.name, spec.PrintedKeywords, row.keywords)
		}
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
			t.Errorf("%s: a vanilla body declares no abilities", row.name)
		}
	}
}

// --- lands ---------------------------------------------------------

func TestB43BaronAndThornglintBridgeEnterTappedAndTapForEitherColour(t *testing.T) {
	for _, row := range []struct {
		name, typeLine, oracle, colour string
	}{
		{"Baron, Airship Kingdom", "Land — Town", b43BaronOracle, "R"},
		{"Thornglint Bridge", "Artifact Land", b43ThornglintBridgeOracle, "W"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := b12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if !b16Tapped(t, g, land) {
			t.Fatalf("%s enters tapped", row.name)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		b28TapForMana(t, g, me.ID, land, row.colour)
		if got := poolColors(me); len(got) != 1 || got[0] != row.colour {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.colour, got)
		}
	}
}

// Thornglint Bridge is the only indestructible land in the batch, and
// the keyword is what makes it worth playing over a plain tapland.
func TestB43ThornglintBridgeIsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bridge := b43Catalog(g, me.ID, "Thornglint Bridge", "Artifact Land", b43ThornglintBridgeOracle, 0, 0)
	assertKeywords(t, g, bridge, "indestructible")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bridge) })
	if !g.Battlefield.Contains(bridge) {
		t.Error("an indestructible land survives a destroy")
	}
}

func TestB43UrzasFactoryTapsForColourlessAndBuildsAssemblyWorkers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	factory := b43Catalog(g, me.ID, "Urza's Factory", "Land — Urza's", b43UrzasFactoryOracle, 0, 0)
	b28TapForMana(t, g, me.ID, factory, "C")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(factory) })
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C", "C", "C", "C")
	b16Activate(t, g, me.ID, factory, 0, game.ActivateAbilityParams{})
	if n := b43TokensNamed(g, me.ID, "Assembly-Worker"); n != 1 {
		t.Errorf("one 2/2 Assembly-Worker: %d", n)
	}
	if !b16Tapped(t, g, factory) {
		t.Error("the Factory taps to build")
	}
}

func TestB43NayaPanoramaSacrificesItselfToFetchOneOfThreeBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
	)
	panorama := b43Catalog(g, me.ID, "Naya Panorama", "Land", b43NayaPanoramaOracle, 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, panorama, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the fetch opens a search")
	}
	if searchOptionNamed(g, c, "Island") != uuid.Nil {
		t.Error("an Island is none of Mountain, Forest or Plains")
	}
	mountain := searchOptionNamed(g, c, "Mountain")
	if mountain == uuid.Nil || searchOptionNamed(g, c, "Plains") == uuid.Nil {
		t.Fatal("both the Mountain and the Plains are offered")
	}
	answerSearchByID(t, g, me.ID, mountain)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(panorama) {
		t.Error("the Panorama sacrificed itself to pay")
	}
	if !g.Battlefield.Contains(mountain) || !b16Tapped(t, g, mountain) {
		t.Error("the basic arrives on the battlefield tapped")
	}
}

func TestB43DesertPingsAnAttackerOnlyInTheEndOfCombatStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	desert := b43Catalog(g, me.ID, "Desert", "Land — Desert", b43DesertOracle, 0, 0)
	bear := b43Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b28TapForMana(t, g, me.ID, desert, "C")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(desert) })
	advanceToMain(t, g)
	// Not the end of combat step: the condition refuses it.
	if err := g.ActivateCatalogAbility(me.ID, desert, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("the ping is not available in a main phase")
	}
	spec, _ := Lookup(b43DesertOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].SorcerySpeed || spec.Activated[0].Condition == nil {
		t.Fatal("one ability, gated by a step CONDITION rather than sorcery speed")
	}
	// The condition really is the end of combat step and nothing else.
	advanceTo(t, g, game.StepEndCombat)
	var open bool
	g.WithWriteLock(func() { open = spec.Activated[0].Condition(g, me.ID, desert) })
	if !open {
		t.Error("the window opens in the end of combat step")
	}
	advanceTo(t, g, game.StepEnd)
	g.WithWriteLock(func() { open = spec.Activated[0].Condition(g, me.ID, desert) })
	if open {
		t.Error("and closes again after it")
	}
	// The damage half cannot be exercised yet: the engine clears
	// AttackingTarget on ENTERING the end of combat step, where CR 511.3
	// removes creatures from combat as that step ENDS. Declared as a
	// caveat on the card.
}

func TestB43IfnirDeadlandsPaysLifeForBlackAndEatsADesertForTwoMinusCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deadlands := b43Catalog(g, me.ID, "Ifnir Deadlands", "Land — Desert", b43IfnirDeadlandsOracle, 0, 0)
	other := b43Catalog(g, me.ID, "Desert", "Land — Desert", b43DesertOracle, 0, 0)
	bear := b43Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	mine := b43Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, deadlands, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility (pay 1 life: add {B}): %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Fatalf("tapped for {B}: pool %v", got)
	}
	if me.Life != life-1 {
		t.Errorf("the black mana costs 1 life: %d → %d", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(deadlands) })
	advanceToMain(t, g)
	b06AddMana(me, "B", "B", "C", "C")
	// Your own creature is not "a creature an opponent controls".
	if err := g.ActivateCatalogAbility(me.ID, deadlands, 0, game.ActivateAbilityParams{
		Targets: cardRefs(mine), SacrificeIDs: []uuid.UUID{other},
	}); err == nil {
		t.Fatal("the counters only go on a creature an opponent controls")
	}
	b16Activate(t, g, me.ID, deadlands, 0, game.ActivateAbilityParams{
		Targets: cardRefs(bear), SacrificeIDs: []uuid.UUID{other},
	})
	if counterOn(g, bear, game.CounterMinusOne) != 2 {
		t.Errorf("two -1/-1 counters: %d", counterOn(g, bear, game.CounterMinusOne))
	}
	if g.Battlefield.Contains(other) {
		t.Error("the other Desert paid the cost")
	}
	if !g.Battlefield.Contains(deadlands) {
		t.Error("sacrificing a different Desert leaves the Deadlands in play")
	}
}

func TestB43InterplanarBeaconGainsLifeOnAPlaneswalkerCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beacon := b43Catalog(g, me.ID, "Interplanar Beacon", "Land", b43InterplanarBeaconOracle, 0, 0)
	b28TapForMana(t, g, me.ID, beacon, "C")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	_ = beacon
	before := me.Life
	castCatalogSpell(t, g, "Some Planeswalker", "Legendary Planeswalker — Test", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("a planeswalker cast gains 1 life: %d → %d", before, me.Life)
	}
	// A creature spell is not a planeswalker spell.
	again := me.Life
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != again {
		t.Errorf("a creature cast gains nothing: %d → %d", again, me.Life)
	}
}

// --- spells --------------------------------------------------------

func TestB43RaiseTheAlarmMakesTwoSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Raise the Alarm", "Instant", b43RaiseTheAlarmOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("two 1/1 white Soldiers: %d", n)
	}
}

func TestB43DreadboreDestroysACreatureOrAPlaneswalkerAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b43Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Signet", "Artifact")
	advanceToMain(t, g)
	if !b03CastRefused(t, g, "Dreadbore", "Sorcery", b43DreadboreOracle, game.CastSpellParams{Targets: cardRefs(rock)}) {
		t.Error("an artifact is not a creature or a planeswalker")
	}
	castCatalogSpell(t, g, "Dreadbore", "Sorcery", b43DreadboreOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature is destroyed")
	}

	// The planeswalker half, in its own game so the loyalty counters are
	// the only thing keeping it alive.
	g2 := newCatalogGame(t)
	opp2 := g2.Seats[1]
	walker := pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: opp2.ID, Controller: opp2.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	castCatalogSpell(t, g2, "Dreadbore", "Sorcery", b43DreadboreOracle, cardRefs(walker))
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(walker) {
		t.Error("the planeswalker is destroyed too")
	}
}

func TestB43PeekShowsOnlyYouTheHandAndDrawsACard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Peek", "Instant", b43PeekOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Peek replaces itself: hand %d → %d", hand, me.Hand.Size())
	}
	for _, c := range opp.Hand.Cards {
		if !c.KnownBy[me.ID] {
			t.Fatal("the caster knows every card in the hand")
		}
		if c.KnownBy[third.ID] {
			t.Fatal("\"look at\" is not \"reveal\" — nobody else learns the hand")
		}
	}
}

func TestB43ElectrickeryPingsOneOpposingCreatureOrOverloadsOntoAllOfThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b43Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	theirs := b43Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	other := b43Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	advanceToMain(t, g)
	if b03CastRefused(t, g, "Electrickery", "Instant", b43ElectrickeryOracle, game.CastSpellParams{Targets: cardRefs(mine)}) != true {
		t.Error("\"you don't control\" keeps your own creature off the target list")
	}
	castCatalogSpell(t, g, "Electrickery", "Instant", b43ElectrickeryOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("1 damage kills an x/1")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("the untargeted x/1 is untouched")
	}
	// Overloaded, it sweeps every creature you don't control and leaves
	// yours alone.
	survivor := b43Creature(g, opp.ID, "Their Other Bird", "Creature — Bird", 1, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Electrickery", TypeLine: "Instant", OracleID: b43ElectrickeryOracle,
		Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	b06AddMana(me, "R", "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overload cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(survivor) || g.Battlefield.Contains(other) {
		t.Error("the overloaded sweep hits every creature you don't control")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the overloaded sweep is one-sided — your own board survives")
	}
}

func TestB43CeruleanWispsRecoloursUntapsAndCantrips(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		Power: 1, Toughness: 1, Colors: []string{"G"}, Tapped: true,
		Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Cerulean Wisps", "Instant", b43CeruleanWispsOracle, cardRefs(elf))
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, elf) {
		t.Error("the creature untaps")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("the Wisps replace themselves: hand %d → %d", hand, me.Hand.Size())
	}
	var colors []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == elf {
				colors = c.Effective().Colors
			}
		}
	})
	if len(colors) != 1 || colors[0] != "U" {
		t.Errorf("\"becomes blue\" replaces the colours: %v", colors)
	}
}

func TestB43MooglesValorDoublesTheBoardThenSavesAllOfIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b43Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	castCatalogSpell(t, g, "Moogles' Valor", "Instant", b43MooglesValorOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Moogle"); n != 2 {
		t.Fatalf("one Moogle per creature you controlled: %d", n)
	}
	// The tokens were created first, so they are indestructible too.
	var moogle uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Moogle" {
			moogle = c.InstanceID
			break
		}
	}
	for _, id := range []uuid.UUID{a, b, moogle} {
		assertKeywords(t, g, id, "indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(moogle) })
	if !g.Battlefield.Contains(moogle) {
		t.Error("the Moogles it just made are indestructible as well")
	}
}

func TestB43ZukosExileExilesAndGivesTheControllerAClue(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b43Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Zuko's Exile", "Instant — Lesson", b43ZukosExileOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the permanent is exiled")
	}
	if n := b43TokensNamed(g, opp.ID, "Clue"); n != 1 {
		t.Errorf("ITS CONTROLLER gets the Clue, not you: %d under the victim", n)
	}
	if n := b43TokensNamed(g, g.Seats[0].ID, "Clue"); n != 0 {
		t.Errorf("the caster gets no Clue: %d", n)
	}
}

func TestB43FracturedIdentityHandsACopyToEveryOtherPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b43Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Fractured Identity", "Sorcery", b43FracturedIdentityOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the permanent is exiled")
	}
	for _, p := range g.Seats {
		want := 1
		if p.ID == opp.ID {
			want = 0
		}
		if n := b43TokensNamed(g, p.ID, "Grizzly Bears"); n != want {
			t.Errorf("seat %s has %d copies, want %d", p.Name, n, want)
		}
	}
	_ = me
}

func TestB43FloodOfTearsBouncesEverythingAndPaysOffOnFourOfYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 4; i++ {
		b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}
	theirs := b43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	me.Hand.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Big Thing", TypeLine: "Creature — Avatar",
		Power: 9, Toughness: 9, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Flood of Tears", "Sorcery", b43FloodOfTearsOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("all nonland permanents go back, including opponents'")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("lands stay")
	}
	p := latestPutFromHandChoice(g, me.ID)
	if p == nil {
		t.Fatal("four of your own nontoken permanents returned: the payoff prompt opens")
	}
}

// latestPutFromHandChoice finds an open "put a permanent card from
// your hand onto the battlefield" prompt for the chooser.
func latestPutFromHandChoice(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Chooser == chooser && c.Kind == game.PendingChoiceChooseCards {
			return c
		}
	}
	return nil
}

func TestB43RestartSequenceReanimatesFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	castCatalogSpell(t, g, "Restart Sequence", "Sorcery", b43RestartSequenceOracle, cardRefs(dead))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("the creature card returns to the battlefield")
	}
}

// --- creatures with an enters trigger ------------------------------

func TestB43ShamanOfThePackDrainsForEveryElfIncludingItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b43Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	b43Creature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	b43Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 1, 1)
	before := opp.Life
	castCatalogSpell(t, g, "Shaman of the Pack", "Creature — Elf Shaman", b43ShamanOfThePackOracle, nil)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	// Two Elves plus the Shaman itself; the opponent's Elf does not count.
	if opp.Life != before-3 {
		t.Errorf("target opponent loses 3: %d → %d", before, opp.Life)
	}
}

func TestB43TowerWinderFindsCommandTowerAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Command Tower", TypeLine: "Land"},
		game.Card{Name: "Command Tower", TypeLine: "Land"},
		game.Card{Name: "Command Beacon", TypeLine: "Land"},
	)
	castCatalogSpell(t, g, "Tower Winder", "Creature — Snake", b43TowerWinderOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger opens a search")
	}
	if searchOptionNamed(g, c, "Command Beacon") != uuid.Nil {
		t.Error("only a card NAMED Command Tower is offered")
	}
	tower := searchOptionNamed(g, c, "Command Tower")
	if tower == uuid.Nil {
		t.Fatal("Command Tower is offered")
	}
	answerSearchByID(t, g, me.ID, tower)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(tower) {
		t.Error("it goes to your hand")
	}
}

func TestB43DragonlordAtarkaSplitsFiveAmongOpposingCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b43Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := b43Creature(g, opp.ID, "Ox", "Creature — Ox", 4, 4)
	b := b43Creature(g, opp.ID, "Elk", "Creature — Elk", 4, 4)
	castCatalogSpell(t, g, "Dragonlord Atarka", "Legendary Creature — Elder Dragon", b43DragonlordAtarkaOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for targets")
	}
	if hasID(p.PickTargetCards, mine) {
		t.Error("\"your opponents control\" keeps your own creatures off the list")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Divided as evenly as possible, remainder to the earliest pick.
	if dmg := b12Card(t, g, a).DamageMarked; dmg != 3 {
		t.Errorf("first pick takes 3: %d", dmg)
	}
	if dmg := b12Card(t, g, b).DamageMarked; dmg != 2 {
		t.Errorf("second pick takes 2: %d", dmg)
	}
}

// --- statics -------------------------------------------------------

func TestB43AngelicFieldMarshalTurnsOnOnlyWithYourOwnCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marshal := b43Catalog(g, me.ID, "Angelic Field Marshal", "Creature — Angel", b43AngelicFieldMarshalOracl, 3, 3)
	bear := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p := effectivePower(t, g, marshal); p != 3 {
		t.Errorf("with no commander out the Marshal is 3/3: power %d", p)
	}
	// An OPPONENT'S commander is not "your commander".
	b43Commander(g, opp.ID, "Their Commander")
	if p := effectivePower(t, g, marshal); p != 3 {
		t.Errorf("an opponent's commander does not switch it on: power %d", p)
	}
	b43Commander(g, me.ID, "My Commander")
	if p := effectivePower(t, g, marshal); p != 5 {
		t.Errorf("lieutenant: the Marshal is 5/5: power %d", p)
	}
	assertKeywords(t, g, marshal, "flying", "vigilance")
	assertKeywords(t, g, bear, "vigilance")
	if got := effectiveAbilities(t, g, theirs); containsString(got, "vigilance") {
		t.Errorf("only creatures YOU control get vigilance: %v", got)
	}
}

func TestB43BoonOfTheSpiritRealmCountsItselfThenGrowsWithEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	boon := castCatalogSpell(t, g, "Boon of the Spirit Realm", "Enchantment", b43BoonOfTheSpiritRealmOrac, nil)
	passPriorityAroundTable(t, g)
	if counterOn(g, boon, "blessing") != 1 {
		t.Fatalf("\"this enchantment OR another\": its own entry is a counter: %d", counterOn(g, boon, "blessing"))
	}
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("one blessing counter is +1/+1: power %d", p)
	}
	castCatalogSpell(t, g, "Another Enchantment", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	if counterOn(g, boon, "blessing") != 2 {
		t.Fatalf("another enchantment is another counter: %d", counterOn(g, boon, "blessing"))
	}
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("two blessing counters is +2/+2: power %d", p)
	}
}

func TestB43ArchetypeOfCourageGivesYourTeamFirstStrikeAndLeavesOpponentsAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archetype := b43Catalog(g, me.ID, "Archetype of Courage", "Enchantment Creature — Human Soldier", b43ArchetypeOfCourageOracle, 2, 2)
	bear := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, archetype, "first strike")
	assertKeywords(t, g, bear, "first strike")
	if got := effectiveAbilities(t, g, theirs); containsString(got, "first strike") {
		t.Errorf("opponents' creatures get nothing from it: %v", got)
	}
}

func TestB43ShieldOfTheOversoulReadsTheHostsColoursLive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b43Creature(g, me.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	shield := castCatalogSpell(t, g, "Shield of the Oversoul", "Enchantment — Aura", b43ShieldOfTheOversoulOracl, cardRefs(host))
	passPriorityAroundTable(t, g)
	_ = shield
	if p := effectivePower(t, g, host); p != 3 {
		t.Errorf("green only: +1/+1: power %d", p)
	}
	assertKeywords(t, g, host, "indestructible")
	if got := effectiveAbilities(t, g, host); containsString(got, "flying") {
		t.Errorf("a green creature that is not white gets no flying: %v", got)
	}
	// A green-white host gets both halves at once.
	gw := b43Creature(g, me.ID, "Selesnya Bear", "Creature — Bear", 2, 2, "G", "W")
	shield2 := castCatalogSpell(t, g, "Shield of the Oversoul", "Enchantment — Aura", b43ShieldOfTheOversoulOracl, cardRefs(gw))
	passPriorityAroundTable(t, g)
	_ = shield2
	if p := effectivePower(t, g, gw); p != 4 {
		t.Errorf("green AND white: +2/+2: power %d", p)
	}
	assertKeywords(t, g, gw, "indestructible", "flying")
}

// --- triggers ------------------------------------------------------

func TestB43NyxFleeceRamGainsALifeEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b43Catalog(g, me.ID, "Nyx-Fleece Ram", "Enchantment Creature — Sheep", b43NyxFleeceRamOracle, 0, 5)
	before := me.Life
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life <= before {
		t.Errorf("your upkeep gains 1 life: %d → %d", before, me.Life)
	}
}

func TestB43NightshadeHarvesterDrainsTheLandsControllerAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harvester := b43Catalog(g, me.ID, "Nightshade Harvester", "Creature — Elf Shaman", b43NightshadeHarvesterOracl, 2, 2)
	// Your own land drop does nothing.
	before := me.Life
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Life != before || counterOn(g, harvester, game.CounterPlusOne) != 0 {
		t.Fatal("\"a land an OPPONENT controls\" does not see your own")
	}
	oppLife := opp.Life
	advanceToNextSeatsTurn(t, g)
	b12PlayFromHand(t, g, "Swamp", "Basic Land — Swamp", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("the LAND'S controller loses 1: %d → %d", oppLife, opp.Life)
	}
	if counterOn(g, harvester, game.CounterPlusOne) != 1 {
		t.Errorf("and the Harvester grows: %d", counterOn(g, harvester, game.CounterPlusOne))
	}
}

func TestB43DarienMakesOneSoldierPerPointOfDamageAndOnlyIfYouSayYes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b43Catalog(g, me.ID, "Darien, King of Kjeldor", "Legendary Creature — Human Soldier", b43DarienOracle, 3, 3)
	bolt := b43Creature(g, g.Seats[1].ID, "Some Source", "Creature — Elemental", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bolt, me.ID, 3) })
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Soldier"); n != 3 {
		t.Errorf("three damage is three Soldiers: %d", n)
	}
	// Life LOSS is not damage, and Darien does not see it.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(bolt, me.ID, -4) })
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Soldier"); n != 3 {
		t.Errorf("life loss makes no Soldiers: %d", n)
	}
}

func TestB43NoxiousGhoulShrinksEveryNonZombieWhenAnyZombieArrives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := b43Catalog(g, me.ID, "Noxious Ghoul", "Creature — Zombie", b43NoxiousGhoulOracle, 3, 3)
	mine := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirZombie := b43Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	castCatalogSpell(t, g, "Gravecrawler", "Creature — Zombie", "", nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p := effectivePower(t, g, id); p != 1 {
			t.Errorf("every non-Zombie creature gets -1/-1: power %d", p)
		}
	}
	for _, id := range []uuid.UUID{ghoul, theirZombie} {
		if p := effectivePower(t, g, id); p != 3 && p != 2 {
			t.Errorf("Zombies are spared: power %d", p)
		}
	}
	if p := effectivePower(t, g, ghoul); p != 3 {
		t.Errorf("the Ghoul is a Zombie and is spared: power %d", p)
	}
	if p := effectivePower(t, g, theirZombie); p != 2 {
		t.Errorf("an opponent's Zombie is spared too: power %d", p)
	}
}

func TestB43SeafloorOracleDrawsOnlyForMerfolkCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b43Catalog(g, me.ID, "Seafloor Oracle", "Creature — Merfolk Wizard", b43SeafloorOracleOracle, 2, 3)
	merfolk := b43Creature(g, me.ID, "Merfolk Scout", "Creature — Merfolk Scout", 2, 2)
	bear := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, merfolk, bear)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("one Merfolk connected, so one card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB43NemesisOfReasonMillsTheDefenderTenOnDeclaration(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nemesis := b43Catalog(g, me.ID, "Nemesis of Reason", "Creature — Leviathan Horror", b43NemesisOfReasonOracle, 3, 7)
	before := opp.Library.Size()
	attackWith(t, g, opp.ID, nemesis)
	passPriorityAroundTable(t, g)
	want := 10
	if before < want {
		want = before
	}
	if opp.Library.Size() != before-want {
		t.Errorf("the defending player mills ten: library %d → %d", before, opp.Library.Size())
	}
	if !g.Battlefield.Contains(nemesis) {
		t.Error("the mill happens on declaration, before any damage")
	}
}

func TestB43DeeprootPilgrimageMakesOneTokenPerTapBatchAndIgnoresTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b43Catalog(g, me.ID, "Deeproot Pilgrimage", "Enchantment", b43DeeprootPilgrimageOracle, 0, 0)
	a := b43Creature(g, me.ID, "Merfolk One", "Creature — Merfolk", 1, 1)
	b := b43Creature(g, me.ID, "Merfolk Two", "Creature — Merfolk", 1, 1)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Merfolk", TypeLine: "Token Creature — Merfolk",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	attackWith(t, g, opp.ID, a, b, token)
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Merfolk") - 1; n != 1 {
		t.Errorf("\"one or more\" is ONE trigger per batch: %d new tokens", n)
	}
}

func TestB43FinneasGrowsTokensAndRabbitsThenDrawsOnTenPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	finneas := b43Catalog(g, me.ID, "Finneas, Ace Archer", "Legendary Creature — Rabbit Archer", b43FinneasOracle, 2, 2)
	rabbit := b43Creature(g, me.ID, "Regal Bunnicorn", "Creature — Rabbit", 4, 4)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	bear := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, finneas)
	passPriorityAroundTable(t, g)
	if counterOn(g, rabbit, game.CounterPlusOne) != 1 {
		t.Errorf("a Rabbit gets a counter: %d", counterOn(g, rabbit, game.CounterPlusOne))
	}
	if counterOn(g, token, game.CounterPlusOne) != 1 {
		t.Errorf("a token gets a counter: %d", counterOn(g, token, game.CounterPlusOne))
	}
	if counterOn(g, bear, game.CounterPlusOne) != 0 {
		t.Errorf("a nontoken non-Rabbit gets nothing: %d", counterOn(g, bear, game.CounterPlusOne))
	}
	if counterOn(g, finneas, game.CounterPlusOne) != 0 {
		t.Errorf("\"each OTHER creature\" excludes Finneas: %d", counterOn(g, finneas, game.CounterPlusOne))
	}
	// 2 + 5 + 4 + 2 = 13 after the counters, so the draw happens.
	if me.Hand.Size() != hand+1 {
		t.Errorf("total power 10 or more draws a card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB43CalderaPyremawGrowsThenBurnsForItsNewPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pyremaw := b43Catalog(g, me.ID, "Caldera Pyremaw", "Creature — Dragon", b43CalderaPyremawOracle, 3, 3)
	before := opp.Life
	castCatalogSpell(t, g, "Opt", "Instant", "", nil)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if counterOn(g, pyremaw, game.CounterPlusOne) != 1 {
		t.Fatalf("the counter goes on first: %d", counterOn(g, pyremaw, game.CounterPlusOne))
	}
	if opp.Life != before-4 {
		t.Errorf("a 3/3 with a fresh counter deals 4: %d → %d", before, opp.Life)
	}
}

func TestB43EowynNeedsAnotherHumanThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b43Catalog(g, me.ID, "Éowyn, Shieldmaiden", "Legendary Creature — Human Knight", b43EowynOracle, 5, 4)
	// Nothing entered this turn: no trigger.
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Human Knight"); n != 0 {
		t.Fatalf("no Human entered this turn, so nothing happens: %d tokens", n)
	}
	advanceToNextSeatsTurn(t, g)
	advanceToNextSeatsTurn(t, g)
	advanceToNextSeatsTurn(t, g)
	advanceToNextSeatsTurn(t, g)
	castCatalogSpell(t, g, "Human Soldier", "Creature — Human Soldier", "", nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Human Knight"); n != 2 {
		t.Errorf("a Human entered this turn: two 2/2 Knights: %d", n)
	}
	_ = opp
}

func TestB43DistinguishedConjurerGainsLifeAndBlinksAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	conjurer := b43Catalog(g, me.ID, "Distinguished Conjurer", "Creature — Human Wizard", b43DistinguishedConjurerOra, 1, 2)
	bear := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	before := me.Life
	castCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("another creature you control entering gains 1 life: %d → %d", before, me.Life)
	}
	advanceToMain(t, g)
	b06AddMana(me, "W", "C", "C", "C", "C")
	// The Conjurer cannot blink itself — "another".
	if err := g.ActivateCatalogAbility(me.ID, conjurer, 0, game.ActivateAbilityParams{Targets: cardRefs(conjurer)}); err == nil {
		t.Fatal("\"another target creature\" excludes the Conjurer")
	}
	b16Activate(t, g, me.ID, conjurer, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if g.Battlefield.Contains(bear) {
		t.Error("a blinked permanent returns as a NEW object, with a new instance")
	}
	returned, ok := b43BattlefieldByName(g, "Bear")
	if !ok {
		t.Fatal("and it does come back")
	}
	if counterOn(g, returned, game.CounterPlusOne) != 0 {
		t.Error("counters fall off the new object")
	}
}

// b43BattlefieldByName finds the first battlefield permanent with the
// given name — the only handle on a permanent that has been blinked,
// since the return is a new instance.
func b43BattlefieldByName(g *game.Game, name string) (uuid.UUID, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID, true
		}
	}
	return uuid.Nil, false
}

func TestB43ZephyrBootsFlyAndLootOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	boots := b43Catalog(g, me.ID, "Zephyr Boots", "Artifact — Equipment", b43ZephyrBootsOracle, 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, boots, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	assertKeywords(t, g, bear, "flying")
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("the draw happens before the discard prompt: hand %d → %d", hand, me.Hand.Size())
	}
	if !anyPendingChoiceFor(g, me.ID) {
		t.Error("and then it asks you to discard")
	}
}

func TestB43CryptcallerChariotMakesATappedZombiePerDiscardAndCrews(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chariot := b43Catalog(g, me.ID, "Cryptcaller Chariot", "Artifact — Vehicle", b43CryptcallerChariotOracle, 5, 5)
	if isCreatureNow(g, chariot) {
		t.Fatal("a Vehicle is not a creature until it is crewed")
	}
	card := me.Hand.Cards[0].InstanceID
	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	answerDiscard(t, g, me.ID, card)
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Zombie"); n != 1 {
		t.Fatalf("one discard is one Zombie: %d", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && !c.Tapped {
			t.Error("the Zombies arrive tapped")
		}
	}
	crew := b43Creature(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	_ = crew
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, chariot, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crew}})
	if !isCreatureNow(g, chariot) {
		t.Error("crewed, it is an artifact creature")
	}
	assertKeywords(t, g, chariot, "menace")
}

// --- activated abilities -------------------------------------------

func TestB43CaptainSisayTutorsOnlyLegendaryCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Sol Ring", TypeLine: "Legendary Artifact"},
		game.Card{Name: "Kozilek", TypeLine: "Legendary Creature — Eldrazi"},
		game.Card{Name: "Signet", TypeLine: "Artifact"},
	)
	sisay := b43Catalog(g, me.ID, "Captain Sisay", "Legendary Creature — Human Soldier", b43CaptainSisayOracle, 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sisay, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the tap opens a search")
	}
	if searchOptionNamed(g, c, "Signet") != uuid.Nil {
		t.Error("a nonlegendary card is not offered")
	}
	ring := searchOptionNamed(g, c, "Sol Ring")
	if ring == uuid.Nil {
		t.Fatal("a legendary ARTIFACT is a legal find — the clause says \"card\"")
	}
	answerSearchByID(t, g, me.ID, ring)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ring) {
		t.Error("it goes to your hand")
	}
	if !b16Tapped(t, g, sisay) {
		t.Error("Sisay taps to activate")
	}
}

func TestB43PlanarBridgePutsAnyPermanentCardOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Big Eldrazi", TypeLine: "Creature — Eldrazi"},
		game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Some Instant", TypeLine: "Instant"},
	)
	bridge := b43Catalog(g, me.ID, "Planar Bridge", "Legendary Artifact", b43PlanarBridgeOracle, 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bridge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the activation opens a search")
	}
	if searchOptionNamed(g, c, "Some Instant") != uuid.Nil {
		t.Error("an instant is not a permanent card")
	}
	eldrazi := searchOptionNamed(g, c, "Big Eldrazi")
	if eldrazi == uuid.Nil {
		t.Fatal("the creature card is offered")
	}
	answerSearchByID(t, g, me.ID, eldrazi)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(eldrazi) {
		t.Error("it arrives on the battlefield")
	}
	if b16Tapped(t, g, eldrazi) {
		t.Error("untapped — the card says nothing about tapped")
	}
}

func TestB43VizkopaGuildmageGrantsLifelinkUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mage := b43Catalog(g, me.ID, "Vizkopa Guildmage", "Creature — Human Wizard", b43VizkopaGuildmageOracle, 2, 2)
	bear := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spec, _ := Lookup(b43VizkopaGuildmageOracle)
	if len(spec.Activated) != 1 {
		t.Fatalf("only the lifelink half ships (declared caveat): %d abilities", len(spec.Activated))
	}
	advanceToMain(t, g)
	b06AddMana(me, "W", "B", "C")
	b16Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	assertKeywords(t, g, bear, "lifelink")
}

// --- cascade -------------------------------------------------------

func TestB43MaelstromColossusCascadesOnCast(t *testing.T) {
	if got := len(game.CatalogTriggers(b43MaelstromColossusOracle)); got != 1 {
		t.Fatalf("one cascade trigger, got %d", got)
	}
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 3)
	castWithCost(t, g, "Maelstrom Colossus", "Artifact Creature — Golem", "{8}", b43MaelstromColossusOracle)
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers the free cast: %d, want 1", got)
	}
	if n := answerAllMayCast(t, g, me.ID, false); n != 1 {
		t.Errorf("answered %d offers, want 1", n)
	}
}

// --- modal triggers and multi-clause targets (#937) -----------------

// Skemfar Shadowsage's mode is chosen as the trigger goes on the
// stack (CR 603.3c), and X is the biggest single tribe on your board
// — not the sum of your creatures.
func TestB43SkemfarShadowsageChoosesAModeThenCountsTheBiggestTribe(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Three Elves (two of them also Clerics) and a Bear: the biggest
	// single type is Elf at 3, and the Shadowsage itself makes 4.
	b43Creature(g, me.ID, "Elf One", "Creature — Elf Druid", 1, 1)
	b43Creature(g, me.ID, "Elf Two", "Creature — Elf Cleric", 1, 1)
	b43Creature(g, me.ID, "Elf Three", "Creature — Elf Cleric", 1, 1)
	b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b43Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 1, 1)
	before := opp.Life
	castCatalogSpell(t, g, "Skemfar Shadowsage", "Creature — Elf Cleric", b43SkemfarShadowsageOracle, nil)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger asks which mode, as it goes on the stack")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Four Elves of yours; the opponent's Elf does not count, and the
	// Bear is not in any tribe with anything.
	if opp.Life != before-4 {
		t.Errorf("each opponent loses 4: %d → %d", before, opp.Life)
	}
}

func TestB43SkemfarShadowsageSecondModeGainsTheSameAmount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b43Creature(g, me.ID, "Zombie One", "Creature — Zombie", 2, 2)
	b43Creature(g, me.ID, "Zombie Two", "Creature — Zombie", 2, 2)
	before := me.Life
	castCatalogSpell(t, g, "Skemfar Shadowsage", "Creature — Elf Cleric", b43SkemfarShadowsageOracle, nil)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the mode prompt is there")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Two Zombies is the biggest tribe; the Shadowsage is an Elf Cleric
	// and is a tribe of one on its own.
	if me.Life != before+2 {
		t.Errorf("you gain 2: %d → %d", before, me.Life)
	}
}

// Gimli's haste is unconditional; the fight waits for total power 8.
func TestB43GimlisRecklessMightGivesHasteAndFightsOnlyAtEightPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b43Catalog(g, me.ID, "Gimli's Reckless Might", "Enchantment", b43GimlisRecklessMightOracl, 0, 0)
	small := b43Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 2)
	theirs := b43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, small, "haste")
	// Total power 2: formidable is not met, so no trigger at all.
	attackWith(t, g, opp.ID, small)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("total power 2 is not 8: the trigger does not fire")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("and nothing fought")
	}
}

func TestB43GimlisRecklessMightFightsWithTwoDifferentTargetClauses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b43Catalog(g, me.ID, "Gimli's Reckless Might", "Enchantment", b43GimlisRecklessMightOracl, 0, 0)
	big := b43Creature(g, me.ID, "Big Bear", "Creature — Bear", 5, 5)
	other := b43Creature(g, me.ID, "Other Bear", "Creature — Bear", 4, 4)
	theirs := b43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	// The trigger's target prompt opens as the attack declaration is
	// locked in, and it asks CLAUSE BY CLAUSE: the attacker you control
	// first, then the creature it fights.
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(big, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	_, _ = g.AdvanceStep()
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("total power 9: formidable is met and the trigger asks for targets")
	}
	if hasID(p.PickTargetCards, other) {
		t.Error("the first clause is an ATTACKING creature you control — the one at home is not legal")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err != nil {
		t.Fatalf("ResolvePickTarget (the attacker): %v", err)
	}
	p = latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("then the second clause asks for the creature it fights")
	}
	if hasID(p.PickTargetCards, big) {
		t.Error("the second clause is a creature you DON'T control")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatalf("ResolvePickTarget (the victim): %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("a 5/5 fighting a 3/3 kills it")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("and survives")
	}
}

// Lost Jitte banks a counter on ANY combat damage the equipped
// creature deals, and spends it through a modal activated ability
// whose bullets have three different target clauses (#937).
func TestB43LostJitteBanksACounterAndSpendsItOnAMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := b43Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	jitte := b43Catalog(g, me.ID, "Lost Jitte", "Legendary Artifact — Equipment", b43LostJitteOracle, 0, 0)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	advanceToMain(t, g)
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, jitte, 1, game.ActivateAbilityParams{Targets: cardRefs(bear)})

	// Blocked, so the damage never reaches a player — the Jitte still
	// banks a counter, which is what "deals combat damage" means.
	attackIntoBlocks(t, g, opp.ID, bear)
	if err := g.DeclareBlocker(blocker, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if counterOn(g, jitte, "charge") < 1 {
		t.Fatalf("equipped creature dealt combat damage: a charge counter: %d", counterOn(g, jitte, "charge"))
	}

	// Spend it on the untap-a-land bullet.
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(land) })
	advanceToMainOf(t, g, 0)
	if err := g.ActivateCatalogAbility(me.ID, jitte, 0, game.ActivateAbilityParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility (untap a land): %v", err)
	}
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, land) {
		t.Error("the untap bullet untapped the land")
	}
	if counterOn(g, jitte, "charge") != 0 {
		t.Errorf("the counter was the cost and is gone: %d", counterOn(g, jitte, "charge"))
	}
}
