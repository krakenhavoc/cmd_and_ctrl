package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dka_5571_b1_cards_test.go — the first batch of Dark Ascension cards
// (#5571). Every test plays the real catalog entry through the engine.

const (
	dkaHollowhengeBeastOracle  = "a13b02c2-7ec1-4032-99cc-c76fe512914e"
	dkaRussetWolvesOracle      = "e788e9dd-449d-495b-adad-6d49e408b003"
	dkaSanctuaryCatOracle      = "c623bc4b-7fc1-4bc1-84e9-71c301dd402f"
	dkaHighbornGhoulOracle     = "1c145f2d-8345-4902-b869-afb70145dec7"
	dkaKessigRecluseOracle     = "61c3b999-4f5d-4018-a9ee-ef894fd1ef6c"
	dkaNephaliaSeakiteOracle   = "379ef166-a29f-4715-8ce4-9d7d567e4e91"
	dkaNearheathStalkerOracle  = "c804157e-2911-4973-83bf-f5befc09065d"
	dkaVengefulVampireOracle   = "1fcf95d7-fd70-474b-be0a-e43502633556"
	dkaStranglerootGeistOracle = "af12758f-4a7b-4156-8942-de4716aa0623"
	dkaVorapedeOracle          = "fd44986c-cf37-4f6e-b8d6-1795ec794b5f"
	dkaSilverclawGriffinOracle = "5a7ecb5c-9b22-4d2c-b851-db92c1d3c745"
	dkaSomberwaldDryadOracle   = "b3e9ae89-3507-4429-ab37-7175f4db4710"
	dkaExecutionersHoodOracle  = "041738f9-7465-4a6c-8046-a897d88cbcdd"
	dkaChantOfTheSkifsang      = "5c902588-b16e-4679-b89e-7b11c928061f"
	dkaBarTheDoorOracle        = "1ecd7248-ff28-4e12-b72a-0cb279ec7561"
	dkaBoneToAshOracle         = "31be5bf4-9950-456c-8365-74b48e132ef5"
	dkaBurningOilOracle        = "c8e9bec0-43e5-4e84-9e0e-5f3db362f1e4"
	dkaChillOfForebodingOracle = "cd1677e4-c356-43cd-9766-5fb8a1e64249"
	dkaCrushingVinesOracle     = "aebe0ba5-70b2-406b-b53b-03b12a95c2c2"
	dkaDeadlyAllureOracle      = "f9cdaecd-9621-4a9a-888e-748ecfe551fd"
	dkaFiresOfUndeathOracle    = "205bcab6-8b6b-4d09-9890-0b7391d59788"
	dkaGrimFloweringOracle     = "65428247-b386-410c-91b7-81dd278c0927"
	dkaGriptideOracle          = "5a6bbc45-6cb8-40ff-9f7c-7b2c5713dd78"
	dkaHarrowingJourneyOracle  = "70f3d545-d114-485d-8d8b-4a93b2393c41"
	dkaMysticRetrievalOracle   = "b7cfbfa1-dd38-47ff-9138-fa6b247afc11"
	dkaReapTheSeagrafOracle    = "95ff3ba5-cbd2-4723-9498-3fc6df5be973"
	dkaSavingGraspOracle       = "f7a4458a-7e8a-46bb-b62d-991c606c7a15"
	dkaSkillfulLungeOracle     = "bc730e60-b9c2-4c49-b97b-2878be105b04"
	dkaWildHungerOracle        = "9d250088-d5f2-40ed-bdb7-a6decce0fb67"
	dkaBriarpackAlphaOracle    = "5604c475-6e01-4e9d-b4e6-8e8d72874581"
	dkaFarbogBoneflingerOracle = "46027ca0-9d06-42b3-8012-124861c6b558"
	dkaBlackCatOracle          = "10edf7e0-d6be-4510-819f-8834bedfba41"
	dkaElgaudInquisitorOracle  = "e84164cf-c3bf-45ae-9a0a-9dfeacd36768"
	dkaErdwalRipperOracle      = "6a207599-188f-49fc-8932-1974fc716f0b"
	dkaForgeDevilOracle        = "be7a38a3-c27f-4917-a59c-763d0b36e427"
	dkaGravetillerWurmOracle   = "020d0c91-da68-40da-8561-5522377f3d07"
	dkaHeadlessSkaabOracle     = "203365d4-4793-47af-840e-418bd5569d50"
	dkaNiblisOfTheMistOracle   = "e059122d-b8d6-4878-bebb-e20ce6311d68"
	dkaNiblisOfTheUrnOracle    = "3ea7a555-d6fa-4cc9-9348-784e6f9a8cfc"
	dkaScreechingSkaabOracle   = "c339982a-45d6-4c1a-aefc-c72427b40744"
	dkaShriekgeistOracle       = "e8dd4341-e7e5-49f7-b173-e4c3195e3334"
	dkaUlvenwaldBearOracle     = "2a79c29c-e539-49c5-b10c-1d5fd77842c9"
	dkaMidnightGuardOracle     = "f5430b0e-e77f-47b8-ae6d-30bcff4d7e35"
	dkaMoonveilDragonOracle    = "e0cd4619-9a83-41f2-a349-358feaf2ef2f"
	dkaDawntreaderElkOracle    = "de6d0248-79f6-48df-bdca-8ce207b2ec08"
	dkaTorchFiendOracle        = "15f8049c-f21d-4cd0-aa81-1a533936cf77"
	dkaThrabenHereticOracle    = "67247945-17b7-4cb8-ba65-d0dfec367b77"
	dkaSkirsdagFlayerOracle    = "38ea058d-f8cd-4bd9-9142-8f0d3d56d571"
	dkaDerangedOutcastOracle   = "115f1397-fa24-4412-80b7-e2a6ddedcd4c"
	dkaHauntedFengrafOracle    = "7c6143f3-ad2c-4d7f-9041-aa59f01d8fb7"

	dkaBearType  = "Creature — Bear"
	dkaHumanType = "Creature — Human Peasant"
)

// dkaGraveyardCard puts a card of the given type line into p's graveyard.
func dkaGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// dkaSeats is the active seat and the next one.
func dkaSeats(g *game.Game) (me, opp *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}

// dkaFlashback casts a card from the active seat's graveyard for its
// flashback cost and resolves it.
func dkaFlashback(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	me, _ := dkaSeats(g)
	id := effectstest.SeedGraveyardCard(t, g, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: targets}); err != nil {
		t.Fatalf("flashback %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Errorf("%s cast with flashback was not exiled", name)
	}
	return id
}

// --- every card is in the catalog ----------------------------------------

func TestDka5571EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		dkaHollowhengeBeastOracle, dkaRussetWolvesOracle, dkaSanctuaryCatOracle, dkaHighbornGhoulOracle,
		dkaKessigRecluseOracle, dkaNephaliaSeakiteOracle, dkaNearheathStalkerOracle, dkaVengefulVampireOracle,
		dkaStranglerootGeistOracle, dkaVorapedeOracle, dkaSilverclawGriffinOracle, dkaSomberwaldDryadOracle,
		dkaExecutionersHoodOracle, dkaChantOfTheSkifsang, dkaBarTheDoorOracle, dkaBoneToAshOracle,
		dkaBurningOilOracle, dkaChillOfForebodingOracle, dkaCrushingVinesOracle, dkaDeadlyAllureOracle,
		dkaFiresOfUndeathOracle, dkaGrimFloweringOracle, dkaGriptideOracle, dkaHarrowingJourneyOracle,
		dkaMysticRetrievalOracle, dkaReapTheSeagrafOracle, dkaSavingGraspOracle, dkaSkillfulLungeOracle,
		dkaWildHungerOracle, dkaBriarpackAlphaOracle, dkaFarbogBoneflingerOracle, dkaBlackCatOracle,
		dkaElgaudInquisitorOracle, dkaErdwalRipperOracle, dkaForgeDevilOracle, dkaGravetillerWurmOracle,
		dkaHeadlessSkaabOracle, dkaNiblisOfTheMistOracle, dkaNiblisOfTheUrnOracle, dkaScreechingSkaabOracle,
		dkaShriekgeistOracle, dkaUlvenwaldBearOracle, dkaMidnightGuardOracle, dkaMoonveilDragonOracle,
		dkaDawntreaderElkOracle, dkaTorchFiendOracle, dkaThrabenHereticOracle, dkaSkirsdagFlayerOracle,
		dkaDerangedOutcastOracle, dkaHauntedFengrafOracle,
	} {
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

// --- printed keywords -----------------------------------------------------------

func TestDka5571KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Highborn Ghoul", "Creature — Zombie", dkaHighbornGhoulOracle, []string{"intimidate"}},
		{"Kessig Recluse", "Creature — Spider", dkaKessigRecluseOracle, []string{"reach", "deathtouch"}},
		{"Nephalia Seakite", "Creature — Bird", dkaNephaliaSeakiteOracle, []string{"flash", "flying"}},
		{"Nearheath Stalker", "Creature — Vampire Rogue", dkaNearheathStalkerOracle, []string{"undying"}},
		{"Vengeful Vampire", "Creature — Vampire", dkaVengefulVampireOracle, []string{"flying", "undying"}},
		{"Strangleroot Geist", "Creature — Spirit", dkaStranglerootGeistOracle, []string{"haste", "undying"}},
		{"Vorapede", "Creature — Insect", dkaVorapedeOracle, []string{"vigilance", "trample", "undying"}},
		{"Silverclaw Griffin", "Creature — Griffin", dkaSilverclawGriffinOracle, []string{"flying", "first strike"}},
		{"Somberwald Dryad", "Creature — Dryad", dkaSomberwaldDryadOracle, []string{"forestwalk"}},
		{"Briarpack Alpha", "Creature — Wolf", dkaBriarpackAlphaOracle, []string{"flash"}},
		{"Elgaud Inquisitor", "Creature — Human Cleric", dkaElgaudInquisitorOracle, []string{"lifelink"}},
		{"Erdwal Ripper", "Creature — Vampire", dkaErdwalRipperOracle, []string{"haste"}},
		{"Gravetiller Wurm", "Creature — Wurm", dkaGravetillerWurmOracle, []string{"trample"}},
		{"Niblis of the Mist", "Creature — Spirit", dkaNiblisOfTheMistOracle, []string{"flying"}},
		{"Niblis of the Urn", "Creature — Spirit", dkaNiblisOfTheUrnOracle, []string{"flying"}},
		{"Shriekgeist", "Creature — Spirit", dkaShriekgeistOracle, []string{"flying"}},
		{"Moonveil Dragon", "Creature — Dragon", dkaMoonveilDragonOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.AssertKeywords(t, g, id, tc.keywords...)
		})
	}
}

// Undying returns each of them once, with a +1/+1 counter.
func TestDka5571UndyingReturnsOnceWithACounter(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Nearheath Stalker", "Creature — Vampire Rogue", dkaNearheathStalkerOracle},
		{"Vengeful Vampire", "Creature — Vampire", dkaVengefulVampireOracle},
		{"Strangleroot Geist", "Creature — Spirit", dkaStranglerootGeistOracle},
		{"Vorapede", "Creature — Insect", dkaVorapedeOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 1)
			effectstest.KillCreature(t, g, me.ID, id)
			g.RunStateChecksForTest()
			effectstest.PassPriorityAroundTable(t, g)
			back := effectstest.BattlefieldIDNamed(g, me.ID, tc.name)
			if back == uuid.Nil {
				t.Fatalf("%s did not return", tc.name)
			}
			if got := effectstest.CountersOn(g, back, game.CounterPlusOne); got != 1 {
				t.Errorf("returned with %d +1/+1 counters, want 1", got)
			}
			effectstest.KillCreature(t, g, me.ID, back)
			g.RunStateChecksForTest()
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.BattlefieldIDNamed(g, me.ID, tc.name) != uuid.Nil {
				t.Error("it returned a second time with a counter on it")
			}
		})
	}
}

// --- equipment and aura -------------------------------------------------------

func TestDka5571ExecutionersHoodGrantsIntimidate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	hood := effectstest.PushCatalogPermanent(g, me.ID, "Executioner's Hood", "Artifact — Equipment", dkaExecutionersHoodOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", dkaBearType, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "intimidate") {
		t.Fatal("the Bear has intimidate before it is equipped")
	}
	effectstest.EquipTo(t, g, me.ID, hood, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "intimidate") {
		t.Error("the equipped creature has no intimidate")
	}
}

func TestDka5571ChantOfTheSkifsangGivesMinusThirteen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := dkaSeats(g)
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Chant of the Skifsang", "Enchantment — Aura", dkaChantOfTheSkifsang, effectstest.CardTarget(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, giant); p != -9 || tough != 4 {
		t.Errorf("the enchanted creature is %d/%d, want -9/4", p, tough)
	}
}

// --- the spells --------------------------------------------------------------------

func TestDka5571BarTheDoorPumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := dkaSeats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", dkaBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", dkaBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bar the Door", "Instant", dkaBarTheDoorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 2 || tough != 6 {
		t.Errorf("your creature is %d/%d, want 2/6", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestDka5571BoneToAshCountersOnlyACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := dkaSeats(g)
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	ash := effectstest.B13HandCard(me, "Bone to Ash", "Instant", dkaBoneToAshOracle)
	if err := g.CastSpell(me.ID, ash, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}); err == nil {
		t.Fatal("Bone to Ash targeted an instant spell")
	}

	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ = dkaSeats(g)
	bear := effectstest.B13HandCard(me, "Bear", dkaBearType, "")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Bear: %v", err)
	}
	ash = effectstest.B13HandCard(me, "Bone to Ash", "Instant", dkaBoneToAshOracle)
	if err := g.CastSpell(me.ID, ash, game.CastSpellParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("CastSpell Bone to Ash: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestDka5571BurningOilHitsOnlyACombatant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raider := effectstest.B12Creature(g, me.ID, "Raider", dkaHumanType, 4, 4)
	idle := effectstest.B12Creature(g, me.ID, "Idle", dkaHumanType, 4, 4)
	effectstest.BrAttack(t, g, raider)
	oil := effectstest.B13HandCard(me, "Burning Oil", "Instant", dkaBurningOilOracle)
	if err := g.CastSpell(me.ID, oil, game.CastSpellParams{Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("Burning Oil targeted a creature that isn't in combat")
	}
	if err := g.CastSpell(me.ID, oil, game.CastSpellParams{Targets: effectstest.CardTarget(raider)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, raider).DamageMarked; got != 3 {
		t.Errorf("the attacker has %d damage, want 3", got)
	}
}

func TestDka5571ChillOfForebodingMillsEachPlayerFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	libs := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		libs[s.ID] = s.Library.Size()
	}
	effectstest.CastCatalogSpell(t, g, "Chill of Foreboding", "Sorcery", dkaChillOfForebodingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if got := libs[s.ID] - s.Library.Size(); got != 5 {
			t.Errorf("%s milled %d, want 5", s.Name, got)
		}
	}
}

func TestDka5571CrushingVinesEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	_, opp := dkaSeats(g)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", dkaBearType, 2, 2)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 2, Toughness: 2, Keywords: []string{"flying"}})
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if _, err := effectstest.CastModalErr(t, g, "Crushing Vines", "Instant", dkaCrushingVinesOracle, []int{0}, effectstest.CardTarget(walker)); err == nil {
		t.Fatal("the flying bullet targeted a creature without flying")
	}
	effectstest.CastModal(t, g, "Crushing Vines", "Instant", dkaCrushingVinesOracle, []int{0}, effectstest.CardTarget(flyer))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Error("the flyer survived")
	}
	effectstest.CastModal(t, g, "Crushing Vines", "Instant", dkaCrushingVinesOracle, []int{1}, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestDka5571DeadlyAllureGrantsDeathtouchAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", dkaBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Deadly Allure", "Sorcery", dkaDeadlyAllureOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the creature has no deathtouch")
	}
	other := effectstest.B12Creature(g, me.ID, "Other", dkaBearType, 2, 2)
	dkaFlashback(t, g, "Deadly Allure", "Sorcery", dkaDeadlyAllureOracle, effectstest.CardTarget(other))
	if !effectstest.HasEffectiveKeyword(t, g, other, "deathtouch") {
		t.Error("the flashback cast granted no deathtouch")
	}
}

func TestDka5571FiresOfUndeathDealsTwoAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	_, opp := dkaSeats(g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Fires of Undeath", "Instant", dkaFiresOfUndeathOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("the player is at %d, want %d", opp.Life, life-2)
	}
	dkaFlashback(t, g, "Fires of Undeath", "Instant", dkaFiresOfUndeathOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	if opp.Life != life-4 {
		t.Errorf("the player is at %d after the flashback, want %d", opp.Life, life-4)
	}
}

func TestDka5571GrimFloweringDrawsPerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := dkaSeats(g)
	effectstest.PushGraveyardCardForTest(me, "One")
	effectstest.PushGraveyardCardForTest(me, "Two")
	dkaGraveyardCard(me, "Rock", "Artifact")
	effectstest.PushGraveyardCardForTest(opp, "Theirs")
	effectstest.CastCatalogSpell(t, g, "Grim Flowering", "Sorcery", dkaGrimFloweringOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestDka5571GriptideTucksTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	_, opp := dkaSeats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", dkaBearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Griptide", "Instant", dkaGriptideOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature is still on the battlefield")
	}
	if top := effectstest.LibraryTopIDs(opp, 1); top[0] != bear {
		t.Error("the creature is not on top of its owner's library")
	}
}

func TestDka5571HarrowingJourneyTargetDrawsThreeAndLosesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	_, opp := dkaSeats(g)
	hand, life := opp.Hand.Size(), opp.Life
	effectstest.CastCatalogSpell(t, g, "Harrowing Journey", "Sorcery", dkaHarrowingJourneyOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 3 {
		t.Errorf("the target drew %d, want 3", got)
	}
	if opp.Life != life-3 {
		t.Errorf("the target is at %d, want %d", opp.Life, life-3)
	}
}

func TestDka5571MysticRetrievalReturnsAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	bear := dkaGraveyardCard(me, "Bear", dkaBearType)
	spell := dkaGraveyardCard(me, "Old Spell", "Instant")
	if err := effectstest.CastCatalogSpellErr(t, g, "Mystic Retrieval", "Sorcery", dkaMysticRetrievalOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("Mystic Retrieval targeted a creature card")
	}
	effectstest.CastCatalogSpell(t, g, "Mystic Retrieval", "Sorcery", dkaMysticRetrievalOracle, effectstest.CardTarget(spell))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("the instant card did not return to hand")
	}
}

func TestDka5571ReapTheSeagrafMakesAZombieAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	effectstest.CastCatalogSpell(t, g, "Reap the Seagraf", "Sorcery", dkaReapTheSeagrafOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Zombie"); got != 1 {
		t.Fatalf("%d Zombie tokens, want 1", got)
	}
	dkaFlashback(t, g, "Reap the Seagraf", "Sorcery", dkaReapTheSeagrafOracle, nil)
	if got := effectstest.B43TokensNamed(g, me.ID, "Zombie"); got != 2 {
		t.Errorf("%d Zombie tokens after the flashback, want 2", got)
	}
}

func TestDka5571SavingGraspReturnsOnlyACreatureYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := dkaSeats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", dkaBearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", dkaBearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Saving Grasp", "Instant", dkaSavingGraspOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("Saving Grasp targeted a creature you don't own")
	}
	effectstest.CastCatalogSpell(t, g, "Saving Grasp", "Instant", dkaSavingGraspOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your creature did not return to your hand")
	}
}

func TestDka5571PumpsGrantAKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
		p, tough              int
	}{
		{"Skillful Lunge", dkaSkillfulLungeOracle, "first strike", 4, 2},
		{"Wild Hunger", dkaWildHungerOracle, "trample", 5, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, _ := dkaSeats(g)
			bear := effectstest.B12Creature(g, me.ID, "Bear", dkaBearType, 2, 2)
			effectstest.CastCatalogSpell(t, g, tc.name, "Instant", tc.oracle, effectstest.CardTarget(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}

func TestDka5571WildHungerFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", dkaBearType, 2, 2)
	dkaFlashback(t, g, "Wild Hunger", "Instant", dkaWildHungerOracle, effectstest.CardTarget(bear))
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 5/3", p, tough)
	}
}

// --- enters and dies triggers ----------------------------------------------------

func TestDka5571EntersPumpsTargetCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Briarpack Alpha", "Creature — Wolf", dkaBriarpackAlphaOracle, 5, 5},
		{"Farbog Boneflinger", "Creature — Zombie", dkaFarbogBoneflingerOracle, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := dkaSeats(g)
			ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 3, 3)
			effectstest.PickCard(t, g, me.ID, ogre)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, ogre); p != tc.p || tough != tc.tough {
				t.Errorf("the target is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
		})
	}
}

func TestDka5571ForgeDevilPingsACreatureAndYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := dkaSeats(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Forge Devil", "Creature — Devil", dkaForgeDevilOracle, 1, 1)
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 1 {
		t.Errorf("the target has %d damage, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestDka5571BlackCatMakesTheTargetDiscardAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := dkaSeats(g)
	cat := effectstest.B12Push(g, me.ID, "Black Cat", "Creature — Zombie Cat", dkaBlackCatOracle, 1, 1)
	effectstest.D2dSetHand(opp, 3)
	effectstest.B27Kill(g, cat)
	g.RunStateChecksForTest()
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != 2 {
		t.Errorf("the opponent holds %d cards, want 2", got)
	}
}

func TestDka5571ElgaudInquisitorLeavesASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := dkaSeats(g)
	id := effectstest.B12Push(g, me.ID, "Elgaud Inquisitor", "Creature — Human Cleric", dkaElgaudInquisitorOracle, 2, 2)
	effectstest.B27Kill(g, id)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("%d Spirit tokens, want 1", got)
	}
}

func TestDka5571GravetillerWurmEntersWithCountersOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		_, opp := dkaSeats(g)
		if died {
			bear := effectstest.B12Creature(g, opp.ID, "Bear", dkaBearType, 1, 1)
			effectstest.KillCreature(t, g, opp.ID, bear)
			g.RunStateChecksForTest()
		}
		wurm := effectstest.CastHoldCreature(t, g, "Gravetiller Wurm", "Creature — Wurm", dkaGravetillerWurmOracle, 4, 4)
		want := 0
		if died {
			want = 4
		}
		if got := effectstest.CountersOn(g, wurm, game.CounterPlusOne); got != want {
			t.Errorf("died=%v: %d +1/+1 counters, want %d", died, got, want)
		}
	}
}

func TestDka5571HeadlessSkaabExilesACreatureCardAndEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Headless Skaab", TypeLine: "Creature — Zombie Warrior", OracleID: dkaHeadlessSkaabOracle,
		Power: 3, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no creature card in the graveyard")
	}
	fuel := effectstest.PushGraveyardCardForTest(me, "Fuel")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, ExileIDs: []uuid.UUID{fuel}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Exile.Contains(fuel) {
		t.Error("the creature card was not exiled")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Headless Skaab did not resolve")
	}
	if !effectstest.IsTapped(g, id) {
		t.Error("Headless Skaab entered untapped")
	}
}

func TestDka5571NiblisOfTheMistMayTapOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := dkaSeats(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastHoldCreature(t, g, "Niblis of the Mist", "Creature — Spirit", dkaNiblisOfTheMistOracle, 2, 1)
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, ogre) {
		t.Error("the target is untapped")
	}
}

func TestDka5571NiblisOfTheUrnMayTapOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := dkaSeats(g)
	urn := effectstest.B12Push(g, me.ID, "Niblis of the Urn", "Creature — Spirit", dkaNiblisOfTheUrnOracle, 1, 1)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, urn)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, ogre)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, ogre) {
		t.Error("the target is untapped")
	}
}

func TestDka5571ScreechingSkaabMillsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	library := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Screeching Skaab", "Creature — Zombie", dkaScreechingSkaabOracle, 2, 1)
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestDka5571UlvenwaldBearNeedsADeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := dkaSeats(g)
	effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastHoldCreature(t, g, "Ulvenwald Bear", dkaBearType, dkaUlvenwaldBearOracle, 2, 2)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("the morbid trigger fired with no death this turn")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = dkaSeats(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", dkaBearType, 1, 1)
	effectstest.KillCreature(t, g, opp.ID, victim)
	g.RunStateChecksForTest()
	effectstest.CastHoldCreature(t, g, "Ulvenwald Bear", dkaBearType, dkaUlvenwaldBearOracle, 2, 2)
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ogre, game.CounterPlusOne); got != 2 {
		t.Errorf("the target has %d +1/+1 counters, want 2", got)
	}
}

func TestDka5571MidnightGuardUntapsWhenAnotherCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := dkaSeats(g)
	guard := effectstest.B12Push(g, me.ID, "Midnight Guard", "Creature — Human Soldier", dkaMidnightGuardOracle, 2, 3)
	effectstest.TapForTest(t, g, guard)
	effectstest.CastHoldCreature(t, g, "Bear", dkaBearType, "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, guard) {
		t.Error("the Guard is still tapped")
	}
}

// --- combat damage triggers ---------------------------------------------------------

func TestDka5571ErdwalRipperGrowsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := dkaSeats(g)
	ripper := effectstest.B12Push(g, me.ID, "Erdwal Ripper", "Creature — Vampire", dkaErdwalRipperOracle, 2, 1)
	effectstest.DealCombatDamage(g, me.ID, ripper, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ripper, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestDka5571ShriekgeistMillsTheDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := dkaSeats(g)
	geist := effectstest.B12Push(g, me.ID, "Shriekgeist", "Creature — Spirit", dkaShriekgeistOracle, 1, 1)
	mine, theirs := me.Library.Size(), opp.Library.Size()
	effectstest.DealCombatDamage(g, me.ID, geist, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := theirs - opp.Library.Size(); got != 2 {
		t.Errorf("the damaged player milled %d, want 2", got)
	}
	if me.Library.Size() != mine {
		t.Error("the controller milled")
	}
}

// --- activated abilities -------------------------------------------------------------

func TestDka5571MoonveilDragonPumpsYourTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := dkaSeats(g)
	dragon := effectstest.B12Push(g, me.ID, "Moonveil Dragon", "Creature — Dragon", dkaMoonveilDragonOracle, 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", dkaBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", dkaBearType, 2, 2)
	effectstest.ApaMana(me, "R")
	effectstest.ApaActivate(t, g, me, dragon, 0, game.ActivateAbilityParams{})
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("your creature has power %d, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, dragon); p != 6 {
		t.Errorf("the Dragon has power %d, want 6", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestDka5571DawntreaderElkFetchesABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	elk := effectstest.B12Push(g, me.ID, "Dawntreader Elk", "Creature — Elk", dkaDawntreaderElkOracle, 2, 2)
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, elk, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if g.Battlefield.Contains(elk) {
		t.Error("the Elk was not sacrificed")
	}
	got, ok := effectstest.AangCardOnBF(g, forest)
	if !ok {
		t.Fatal("the basic Forest is not on the battlefield")
	}
	if !got.Tapped {
		t.Error("the land entered untapped")
	}
}

func TestDka5571TorchFiendDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := dkaSeats(g)
	fiend := effectstest.B12Push(g, me.ID, "Torch Fiend", "Creature — Devil", dkaTorchFiendOracle, 2, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, fiend, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if g.Battlefield.Contains(fiend) {
		t.Error("the Fiend was not sacrificed")
	}
}

func TestDka5571ThrabenHereticExilesACreatureCardFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := dkaSeats(g)
	heretic := effectstest.B12Push(g, me.ID, "Thraben Heretic", "Creature — Human Wizard", dkaThrabenHereticOracle, 2, 2)
	dead := effectstest.PushGraveyardCardForTest(opp, "Dead")
	if err := g.ActivateCatalogAbility(me.ID, heretic, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(dead)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled")
	}
	if !effectstest.IsTapped(g, heretic) {
		t.Error("the Heretic did not tap")
	}
}

func TestDka5571SkirsdagFlayerSacrificesAHumanToDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := dkaSeats(g)
	flayer := effectstest.B12Push(g, me.ID, "Skirsdag Flayer", "Creature — Human Cleric", dkaSkirsdagFlayerOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", dkaBearType, 2, 2)
	villager := effectstest.B12Creature(g, me.ID, "Villager", dkaHumanType, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	effectstest.ApaMana(me, "B", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, flayer, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardTarget(victim), SacrificeIDs: []uuid.UUID{bear},
	}); err == nil {
		t.Fatal("a non-Human paid the sacrifice")
	}
	effectstest.ApaActivate(t, g, me, flayer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim), SacrificeIDs: []uuid.UUID{villager}})
	if g.Battlefield.Contains(villager) {
		t.Error("the Human was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived")
	}
}

func TestDka5571DerangedOutcastSacrificesAHumanForTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	outcast := effectstest.B12Push(g, me.ID, "Deranged Outcast", "Creature — Human Rogue", dkaDerangedOutcastOracle, 2, 1)
	villager := effectstest.B12Creature(g, me.ID, "Villager", dkaHumanType, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", dkaBearType, 2, 2)
	effectstest.ApaMana(me, "G", "C")
	effectstest.ApaActivate(t, g, me, outcast, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear), SacrificeIDs: []uuid.UUID{villager}})
	if g.Battlefield.Contains(villager) {
		t.Error("the Human was not sacrificed")
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the target has %d +1/+1 counters, want 2", got)
	}
}

func TestDka5571HauntedFengrafReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := dkaSeats(g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Haunted Fengraf", "Land", dkaHauntedFengrafOracle, false)
	rock := dkaGraveyardCard(me, "Rock", "Artifact")
	dead := effectstest.PushGraveyardCardForTest(me, "Dead")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
	if !me.Graveyard.Contains(rock) {
		t.Error("a noncreature card left the graveyard")
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
}
