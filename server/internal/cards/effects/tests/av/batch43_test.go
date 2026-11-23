package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b43Commander seeds a creature flagged as its owner's commander.
func b43Commander(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
	})
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Naya Panorama
// is a row in the Panorama cycle table, so a transposed row would be
// invisible until someone played that exact land.
func TestBatch43CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B43GigantosaurusOracle:      "Gigantosaurus",
		effectstest.B43IndomitableAncientsOracl: "Indomitable Ancients",
		effectstest.B43RaiseTheAlarmOracle:      "Raise the Alarm",
		effectstest.B43DreadboreOracle:          "Dreadbore",
		effectstest.B43NyxFleeceRamOracle:       "Nyx-Fleece Ram",
		effectstest.B43BaronOracle:              "Baron, Airship Kingdom",
		effectstest.B43ThornglintBridgeOracle:   "Thornglint Bridge",
		effectstest.B43UrzasFactoryOracle:       "Urza's Factory",
		effectstest.B43NayaPanoramaOracle:       "Naya Panorama",
		effectstest.B43ShamanOfThePackOracle:    "Shaman of the Pack",
		effectstest.B43SeafloorOracleOracle:     "Seafloor Oracle",
		effectstest.B43NemesisOfReasonOracle:    "Nemesis of Reason",
		effectstest.B43PeekOracle:               "Peek",
		effectstest.B43CaptainSisayOracle:       "Captain Sisay",
		effectstest.B43PlanarBridgeOracle:       "Planar Bridge",
		effectstest.B43NightshadeHarvesterOracl: "Nightshade Harvester",
		effectstest.B43DarienOracle:             "Darien, King of Kjeldor",
		effectstest.B43ElectrickeryOracle:       "Electrickery",
		effectstest.B43AngelicFieldMarshalOracl: "Angelic Field Marshal",
		effectstest.B43BoonOfTheSpiritRealmOrac: "Boon of the Spirit Realm",
		effectstest.B43NoxiousGhoulOracle:       "Noxious Ghoul",
		effectstest.B43DistinguishedConjurerOra: "Distinguished Conjurer",
		effectstest.B43CeruleanWispsOracle:      "Cerulean Wisps",
		effectstest.B43TowerWinderOracle:        "Tower Winder",
		effectstest.B43MooglesValorOracle:       "Moogles' Valor",
		effectstest.B43EowynOracle:              "Éowyn, Shieldmaiden",
		effectstest.B43DeeprootPilgrimageOracle: "Deeproot Pilgrimage",
		effectstest.B43FinneasOracle:            "Finneas, Ace Archer",
		effectstest.B43CalderaPyremawOracle:     "Caldera Pyremaw",
		effectstest.B43ZukosExileOracle:         "Zuko's Exile",
		effectstest.B43DragonlordAtarkaOracle:   "Dragonlord Atarka",
		effectstest.B43MaelstromColossusOracle:  "Maelstrom Colossus",
		effectstest.B43ZephyrBootsOracle:        "Zephyr Boots",
		effectstest.B43ShieldOfTheOversoulOracl: "Shield of the Oversoul",
		effectstest.B43DesertOracle:             "Desert",
		effectstest.B43IfnirDeadlandsOracle:     "Ifnir Deadlands",
		effectstest.B43CryptcallerChariotOracle: "Cryptcaller Chariot",
		effectstest.B43ArchetypeOfCourageOracle: "Archetype of Courage",
		effectstest.B43VizkopaGuildmageOracle:   "Vizkopa Guildmage",
		effectstest.B43RestartSequenceOracle:    "Restart Sequence",
		effectstest.B43FracturedIdentityOracle:  "Fractured Identity",
		effectstest.B43FloodOfTearsOracle:       "Flood of Tears",
		effectstest.B43AgonasaurRexOracle:       "Agonasaur Rex",
		effectstest.B43InterplanarBeaconOracle:  "Interplanar Beacon",
		effectstest.B43SkemfarShadowsageOracle:  "Skemfar Shadowsage",
		effectstest.B43GimlisRecklessMightOracl: "Gimli's Reckless Might",
		effectstest.B43LostJitteOracle:          "Lost Jitte",
	}
	if len(want) != 47 {
		t.Fatalf("the batch registers 47 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// The two vanilla bodies are registered so they resolve automatically
// and declare no abilities. Agonasaur Rex used to be a third row here
// — #1412 gave it cycling and its cycle trigger, so it is no longer
// vanilla and is covered by its own tests below instead.
func TestB43VanillaBodiesDeclareOnlyWhatTheyPrint(t *testing.T) {
	for _, row := range []struct {
		name, oracle string
		keywords     []string
	}{
		{"Gigantosaurus", effectstest.B43GigantosaurusOracle, nil},
		{"Indomitable Ancients", effectstest.B43IndomitableAncientsOracl, nil},
	} {
		spec, ok := effects.Lookup(row.oracle)
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
		{"Baron, Airship Kingdom", "Land — Town", effectstest.B43BaronOracle, "R"},
		{"Thornglint Bridge", "Artifact Land", effectstest.B43ThornglintBridgeOracle, "W"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.B12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if !effectstest.B16Tapped(t, g, land) {
			t.Fatalf("%s enters tapped", row.name)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		effectstest.B28TapForMana(t, g, me.ID, land, row.colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != row.colour {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.colour, got)
		}
	}
}

// Thornglint Bridge is the only indestructible land in the batch, and
// the keyword is what makes it worth playing over a plain tapland.
func TestB43ThornglintBridgeIsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bridge := effectstest.B43Catalog(g, me.ID, "Thornglint Bridge", "Artifact Land", effectstest.B43ThornglintBridgeOracle, 0, 0)
	effectstest.AssertKeywords(t, g, bridge, "indestructible")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bridge) })
	if !g.Battlefield.Contains(bridge) {
		t.Error("an indestructible land survives a destroy")
	}
}

func TestB43UrzasFactoryTapsForColourlessAndBuildsAssemblyWorkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	factory := effectstest.B43Catalog(g, me.ID, "Urza's Factory", "Land — Urza's", effectstest.B43UrzasFactoryOracle, 0, 0)
	effectstest.B28TapForMana(t, g, me.ID, factory, "C")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(factory) })
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, factory, 0, game.ActivateAbilityParams{})
	if n := effectstest.B43TokensNamed(g, me.ID, "Assembly-Worker"); n != 1 {
		t.Errorf("one 2/2 Assembly-Worker: %d", n)
	}
	if !effectstest.B16Tapped(t, g, factory) {
		t.Error("the Factory taps to build")
	}
}

func TestB43NayaPanoramaSacrificesItselfToFetchOneOfThreeBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
	)
	panorama := effectstest.B43Catalog(g, me.ID, "Naya Panorama", "Land", effectstest.B43NayaPanoramaOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, panorama, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the fetch opens a search")
	}
	if effectstest.SearchOptionNamed(g, c, "Island") != uuid.Nil {
		t.Error("an Island is none of Mountain, Forest or Plains")
	}
	mountain := effectstest.SearchOptionNamed(g, c, "Mountain")
	if mountain == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Plains") == uuid.Nil {
		t.Fatal("both the Mountain and the Plains are offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, mountain)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(panorama) {
		t.Error("the Panorama sacrificed itself to pay")
	}
	if !g.Battlefield.Contains(mountain) || !effectstest.B16Tapped(t, g, mountain) {
		t.Error("the basic arrives on the battlefield tapped")
	}
}

func TestB43DesertPingsAnAttackerOnlyInTheEndOfCombatStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	desert := effectstest.B43Catalog(g, me.ID, "Desert", "Land — Desert", effectstest.B43DesertOracle, 0, 0)
	bear := effectstest.B43Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B28TapForMana(t, g, me.ID, desert, "C")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(desert) })
	effectstest.AdvanceToMain(t, g)
	// Not the end of combat step: the condition refuses it.
	if err := g.ActivateCatalogAbility(me.ID, desert, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("the ping is not available in a main phase")
	}
	spec, _ := effects.Lookup(effectstest.B43DesertOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].SorcerySpeed || spec.Activated[0].Condition == nil {
		t.Fatal("one ability, gated by a step CONDITION rather than sorcery speed")
	}
	// The condition really is the end of combat step and nothing else.
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	var open bool
	g.WithWriteLock(func() { open = spec.Activated[0].Condition(g, me.ID, desert) })
	if !open {
		t.Error("the window opens in the end of combat step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	g.WithWriteLock(func() { open = spec.Activated[0].Condition(g, me.ID, desert) })
	if open {
		t.Error("and closes again after it")
	}
	// The damage half is the test below.
}

// The damage half, and the caveat this card shipped with: creatures
// are in combat for the whole end of combat step (CR 511.3, #785), so
// the ping has the legal target the printed card promises in the one
// window it is allowed in.
func TestB43DesertPingsAnAttackerInTheEndOfCombatStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	desert := effectstest.B43Catalog(g, me.ID, "Desert", "Land — Desert", effectstest.B43DesertOracle, 0, 0)
	raider := effectstest.B43Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, raider)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.B16Activate(t, g, me.ID, desert, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(raider)})

	c, ok := effectstest.BattlefieldCard(g, raider)
	if !ok {
		t.Fatalf("the attacker left the battlefield")
	}
	if c.DamageMarked != 1 {
		t.Errorf("the attacker has %d damage marked, want 1 from the Desert", c.DamageMarked)
	}
}

func TestB43IfnirDeadlandsPaysLifeForBlackAndEatsADesertForTwoMinusCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deadlands := effectstest.B43Catalog(g, me.ID, "Ifnir Deadlands", "Land — Desert", effectstest.B43IfnirDeadlandsOracle, 0, 0)
	other := effectstest.B43Catalog(g, me.ID, "Desert", "Land — Desert", effectstest.B43DesertOracle, 0, 0)
	bear := effectstest.B43Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	mine := effectstest.B43Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, deadlands, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility (pay 1 life: add {B}): %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Fatalf("tapped for {B}: pool %v", got)
	}
	if me.Life != life-1 {
		t.Errorf("the black mana costs 1 life: %d → %d", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(deadlands) })
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "C", "C")
	// Your own creature is not "a creature an opponent controls".
	if err := g.ActivateCatalogAbility(me.ID, deadlands, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(mine), SacrificeIDs: []uuid.UUID{other},
	}); err == nil {
		t.Fatal("the counters only go on a creature an opponent controls")
	}
	effectstest.B16Activate(t, g, me.ID, deadlands, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{other},
	})
	if effectstest.CounterOn(g, bear, game.CounterMinusOne) != 2 {
		t.Errorf("two -1/-1 counters: %d", effectstest.CounterOn(g, bear, game.CounterMinusOne))
	}
	if g.Battlefield.Contains(other) {
		t.Error("the other Desert paid the cost")
	}
	if !g.Battlefield.Contains(deadlands) {
		t.Error("sacrificing a different Desert leaves the Deadlands in play")
	}
}

func TestB43InterplanarBeaconGainsLifeOnAPlaneswalkerCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beacon := effectstest.B43Catalog(g, me.ID, "Interplanar Beacon", "Land", effectstest.B43InterplanarBeaconOracle, 0, 0)
	effectstest.B28TapForMana(t, g, me.ID, beacon, "C")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for {C}: pool %v", got)
	}
	_ = beacon
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Some Planeswalker", "Legendary Planeswalker — Test", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("a planeswalker cast gains 1 life: %d → %d", before, me.Life)
	}
	// A creature spell is not a planeswalker spell.
	again := me.Life
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != again {
		t.Errorf("a creature cast gains nothing: %d → %d", again, me.Life)
	}
}

// --- spells --------------------------------------------------------

func TestB43RaiseTheAlarmMakesTwoSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Raise the Alarm", "Instant", effectstest.B43RaiseTheAlarmOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("two 1/1 white Soldiers: %d", n)
	}
}

func TestB43DreadboreDestroysACreatureOrAPlaneswalkerAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B43Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Signet", "Artifact")
	effectstest.AdvanceToMain(t, g)
	if !effectstest.B03CastRefused(t, g, "Dreadbore", "Sorcery", effectstest.B43DreadboreOracle, game.CastSpellParams{Targets: effectstest.CardRefs(rock)}) {
		t.Error("an artifact is not a creature or a planeswalker")
	}
	effectstest.CastCatalogSpell(t, g, "Dreadbore", "Sorcery", effectstest.B43DreadboreOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature is destroyed")
	}

	// The planeswalker half, in its own game so the loyalty counters are
	// the only thing keeping it alive.
	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	walker := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: opp2.ID, Controller: opp2.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	effectstest.CastCatalogSpell(t, g2, "Dreadbore", "Sorcery", effectstest.B43DreadboreOracle, effectstest.CardRefs(walker))
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(walker) {
		t.Error("the planeswalker is destroyed too")
	}
}

func TestB43PeekShowsOnlyYouTheHandAndDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Peek", "Instant", effectstest.B43PeekOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B43Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	theirs := effectstest.B43Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	other := effectstest.B43Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if effectstest.B03CastRefused(t, g, "Electrickery", "Instant", effectstest.B43ElectrickeryOracle, game.CastSpellParams{Targets: effectstest.CardRefs(mine)}) != true {
		t.Error("\"you don't control\" keeps your own creature off the target list")
	}
	effectstest.CastCatalogSpell(t, g, "Electrickery", "Instant", effectstest.B43ElectrickeryOracle, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("1 damage kills an x/1")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("the untargeted x/1 is untouched")
	}
	// Overloaded, it sweeps every creature you don't control and leaves
	// yours alone.
	survivor := effectstest.B43Creature(g, opp.ID, "Their Other Bird", "Creature — Bird", 1, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Electrickery", TypeLine: "Instant", OracleID: effectstest.B43ElectrickeryOracle,
		Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overload cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(survivor) || g.Battlefield.Contains(other) {
		t.Error("the overloaded sweep hits every creature you don't control")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the overloaded sweep is one-sided — your own board survives")
	}
}

func TestB43CeruleanWispsRecoloursUntapsAndCantrips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		Power: 1, Toughness: 1, Colors: []string{"G"}, Tapped: true,
		Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Cerulean Wisps", "Instant", effectstest.B43CeruleanWispsOracle, effectstest.CardRefs(elf))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, elf) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B43Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Moogles' Valor", "Instant", effectstest.B43MooglesValorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Moogle"); n != 2 {
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
		effectstest.AssertKeywords(t, g, id, "indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(moogle) })
	if !g.Battlefield.Contains(moogle) {
		t.Error("the Moogles it just made are indestructible as well")
	}
}

func TestB43ZukosExileExilesAndGivesTheControllerAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B43Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Zuko's Exile", "Instant — Lesson", effectstest.B43ZukosExileOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the permanent is exiled")
	}
	if n := effectstest.B43TokensNamed(g, opp.ID, "Clue"); n != 1 {
		t.Errorf("ITS CONTROLLER gets the Clue, not you: %d under the victim", n)
	}
	if n := effectstest.B43TokensNamed(g, g.Seats[0].ID, "Clue"); n != 0 {
		t.Errorf("the caster gets no Clue: %d", n)
	}
}

func TestB43FracturedIdentityHandsACopyToEveryOtherPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B43Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Fractured Identity", "Sorcery", effectstest.B43FracturedIdentityOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the permanent is exiled")
	}
	for _, p := range g.Seats {
		want := 1
		if p.ID == opp.ID {
			want = 0
		}
		if n := effectstest.B43TokensNamed(g, p.ID, "Grizzly Bears"); n != want {
			t.Errorf("seat %s has %d copies, want %d", p.Name, n, want)
		}
	}
	_ = me
}

func TestB43FloodOfTearsBouncesEverythingAndPaysOffOnFourOfYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 4; i++ {
		effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}
	theirs := effectstest.B43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	me.Hand.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Big Thing", TypeLine: "Creature — Avatar",
		Power: 9, Toughness: 9, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Flood of Tears", "Sorcery", effectstest.B43FloodOfTearsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("all nonland permanents go back, including opponents'")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("lands stay")
	}
	p := effectstest.LatestPutFromHandChoice(g, me.ID)
	if p == nil {
		t.Fatal("four of your own nontoken permanents returned: the payoff prompt opens")
	}
}

func TestB43RestartSequenceReanimatesFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Restart Sequence", "Sorcery", effectstest.B43RestartSequenceOracle, effectstest.CardRefs(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("the creature card returns to the battlefield")
	}
}

// --- creatures with an enters trigger ------------------------------

func TestB43ShamanOfThePackDrainsForEveryElfIncludingItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	effectstest.B43Creature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	effectstest.B43Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 1, 1)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Shaman of the Pack", "Creature — Elf Shaman", effectstest.B43ShamanOfThePackOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// Two Elves plus the Shaman itself; the opponent's Elf does not count.
	if opp.Life != before-3 {
		t.Errorf("target opponent loses 3: %d → %d", before, opp.Life)
	}
}

func TestB43TowerWinderFindsCommandTowerAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Command Tower", TypeLine: "Land"},
		game.Card{Name: "Command Tower", TypeLine: "Land"},
		game.Card{Name: "Command Beacon", TypeLine: "Land"},
	)
	effectstest.CastCatalogSpell(t, g, "Tower Winder", "Creature — Snake", effectstest.B43TowerWinderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger opens a search")
	}
	if effectstest.SearchOptionNamed(g, c, "Command Beacon") != uuid.Nil {
		t.Error("only a card NAMED Command Tower is offered")
	}
	tower := effectstest.SearchOptionNamed(g, c, "Command Tower")
	if tower == uuid.Nil {
		t.Fatal("Command Tower is offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, tower)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(tower) {
		t.Error("it goes to your hand")
	}
}

func TestB43DragonlordAtarkaSplitsFiveAmongOpposingCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B43Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := effectstest.B43Creature(g, opp.ID, "Ox", "Creature — Ox", 4, 4)
	b := effectstest.B43Creature(g, opp.ID, "Elk", "Creature — Elk", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Dragonlord Atarka", "Legendary Creature — Elder Dragon", effectstest.B43DragonlordAtarkaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for targets")
	}
	if effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("\"your opponents control\" keeps your own creatures off the list")
	}
	// #1563: the caster divides — 1 to the first pick, 4 to the second,
	// which the old even division (3 / 2) never offered.
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}, map[uuid.UUID]int{a: 1, b: 4}); err != nil {
		t.Fatalf("ResolvePickTargetsDivided: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if dmg := effectstest.B12Card(t, g, a).DamageMarked; dmg != 1 {
		t.Errorf("first pick was assigned 1: %d", dmg)
	}
	if g.Battlefield.Contains(b) {
		t.Errorf("second pick was assigned 4 — lethal on a 4/4")
	}
}

// --- statics -------------------------------------------------------

func TestB43AngelicFieldMarshalTurnsOnOnlyWithYourOwnCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marshal := effectstest.B43Catalog(g, me.ID, "Angelic Field Marshal", "Creature — Angel", effectstest.B43AngelicFieldMarshalOracl, 3, 3)
	bear := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, marshal); p != 3 {
		t.Errorf("with no commander out the Marshal is 3/3: power %d", p)
	}
	// An OPPONENT'S commander is not "your commander".
	b43Commander(g, opp.ID, "Their Commander")
	if p := effectstest.EffectivePower(t, g, marshal); p != 3 {
		t.Errorf("an opponent's commander does not switch it on: power %d", p)
	}
	b43Commander(g, me.ID, "My Commander")
	if p := effectstest.EffectivePower(t, g, marshal); p != 5 {
		t.Errorf("lieutenant: the Marshal is 5/5: power %d", p)
	}
	effectstest.AssertKeywords(t, g, marshal, "flying", "vigilance")
	effectstest.AssertKeywords(t, g, bear, "vigilance")
	if got := effectstest.EffectiveAbilities(t, g, theirs); effectstest.ContainsString(got, "vigilance") {
		t.Errorf("only creatures YOU control get vigilance: %v", got)
	}
}

func TestB43BoonOfTheSpiritRealmCountsItselfThenGrowsWithEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	boon := effectstest.CastCatalogSpell(t, g, "Boon of the Spirit Realm", "Enchantment", effectstest.B43BoonOfTheSpiritRealmOrac, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, boon, "blessing") != 1 {
		t.Fatalf("\"this enchantment OR another\": its own entry is a counter: %d", effectstest.CounterOn(g, boon, "blessing"))
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("one blessing counter is +1/+1: power %d", p)
	}
	effectstest.CastCatalogSpell(t, g, "Another Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, boon, "blessing") != 2 {
		t.Fatalf("another enchantment is another counter: %d", effectstest.CounterOn(g, boon, "blessing"))
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("two blessing counters is +2/+2: power %d", p)
	}
}

func TestB43ArchetypeOfCourageGivesYourTeamFirstStrikeAndLeavesOpponentsAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archetype := effectstest.B43Catalog(g, me.ID, "Archetype of Courage", "Enchantment Creature — Human Soldier", effectstest.B43ArchetypeOfCourageOracle, 2, 2)
	bear := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AssertKeywords(t, g, archetype, "first strike")
	effectstest.AssertKeywords(t, g, bear, "first strike")
	if got := effectstest.EffectiveAbilities(t, g, theirs); effectstest.ContainsString(got, "first strike") {
		t.Errorf("opponents' creatures get nothing from it: %v", got)
	}
}

func TestB43ShieldOfTheOversoulReadsTheHostsColoursLive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B43Creature(g, me.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	shield := effectstest.CastCatalogSpell(t, g, "Shield of the Oversoul", "Enchantment — Aura", effectstest.B43ShieldOfTheOversoulOracl, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	_ = shield
	if p := effectstest.EffectivePower(t, g, host); p != 3 {
		t.Errorf("green only: +1/+1: power %d", p)
	}
	effectstest.AssertKeywords(t, g, host, "indestructible")
	if got := effectstest.EffectiveAbilities(t, g, host); effectstest.ContainsString(got, "flying") {
		t.Errorf("a green creature that is not white gets no flying: %v", got)
	}
	// A green-white host gets both halves at once.
	gw := effectstest.B43Creature(g, me.ID, "Selesnya Bear", "Creature — Bear", 2, 2, "G", "W")
	shield2 := effectstest.CastCatalogSpell(t, g, "Shield of the Oversoul", "Enchantment — Aura", effectstest.B43ShieldOfTheOversoulOracl, effectstest.CardRefs(gw))
	effectstest.PassPriorityAroundTable(t, g)
	_ = shield2
	if p := effectstest.EffectivePower(t, g, gw); p != 4 {
		t.Errorf("green AND white: +2/+2: power %d", p)
	}
	effectstest.AssertKeywords(t, g, gw, "indestructible", "flying")
}

// --- triggers ------------------------------------------------------

func TestB43NyxFleeceRamGainsALifeEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B43Catalog(g, me.ID, "Nyx-Fleece Ram", "Enchantment Creature — Sheep", effectstest.B43NyxFleeceRamOracle, 0, 5)
	before := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life <= before {
		t.Errorf("your upkeep gains 1 life: %d → %d", before, me.Life)
	}
}

func TestB43NightshadeHarvesterDrainsTheLandsControllerAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harvester := effectstest.B43Catalog(g, me.ID, "Nightshade Harvester", "Creature — Elf Shaman", effectstest.B43NightshadeHarvesterOracl, 2, 2)
	// Your own land drop does nothing.
	before := me.Life
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before || effectstest.CounterOn(g, harvester, game.CounterPlusOne) != 0 {
		t.Fatal("\"a land an OPPONENT controls\" does not see your own")
	}
	oppLife := opp.Life
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.B12PlayFromHand(t, g, "Swamp", "Basic Land — Swamp", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("the LAND'S controller loses 1: %d → %d", oppLife, opp.Life)
	}
	if effectstest.CounterOn(g, harvester, game.CounterPlusOne) != 1 {
		t.Errorf("and the Harvester grows: %d", effectstest.CounterOn(g, harvester, game.CounterPlusOne))
	}
}

func TestB43DarienMakesOneSoldierPerPointOfDamageAndOnlyIfYouSayYes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B43Catalog(g, me.ID, "Darien, King of Kjeldor", "Legendary Creature — Human Soldier", effectstest.B43DarienOracle, 3, 3)
	bolt := effectstest.B43Creature(g, g.Seats[1].ID, "Some Source", "Creature — Elemental", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bolt, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Soldier"); n != 3 {
		t.Errorf("three damage is three Soldiers: %d", n)
	}
	// Life LOSS is not damage, and Darien does not see it.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(bolt, me.ID, -4) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Soldier"); n != 3 {
		t.Errorf("life loss makes no Soldiers: %d", n)
	}
}

func TestB43NoxiousGhoulShrinksEveryNonZombieWhenAnyZombieArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.B43Catalog(g, me.ID, "Noxious Ghoul", "Creature — Zombie", effectstest.B43NoxiousGhoulOracle, 3, 3)
	mine := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirZombie := effectstest.B43Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Gravecrawler", "Creature — Zombie", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p := effectstest.EffectivePower(t, g, id); p != 1 {
			t.Errorf("every non-Zombie creature gets -1/-1: power %d", p)
		}
	}
	for _, id := range []uuid.UUID{ghoul, theirZombie} {
		if p := effectstest.EffectivePower(t, g, id); p != 3 && p != 2 {
			t.Errorf("Zombies are spared: power %d", p)
		}
	}
	if p := effectstest.EffectivePower(t, g, ghoul); p != 3 {
		t.Errorf("the Ghoul is a Zombie and is spared: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirZombie); p != 2 {
		t.Errorf("an opponent's Zombie is spared too: power %d", p)
	}
}

func TestB43SeafloorOracleDrawsOnlyForMerfolkCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Catalog(g, me.ID, "Seafloor Oracle", "Creature — Merfolk Wizard", effectstest.B43SeafloorOracleOracle, 2, 3)
	merfolk := effectstest.B43Creature(g, me.ID, "Merfolk Scout", "Creature — Merfolk Scout", 2, 2)
	bear := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, merfolk, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("one Merfolk connected, so one card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB43NemesisOfReasonMillsTheDefenderTenOnDeclaration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nemesis := effectstest.B43Catalog(g, me.ID, "Nemesis of Reason", "Creature — Leviathan Horror", effectstest.B43NemesisOfReasonOracle, 3, 7)
	before := opp.Library.Size()
	effectstest.AttackWith(t, g, opp.ID, nemesis)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Catalog(g, me.ID, "Deeproot Pilgrimage", "Enchantment", effectstest.B43DeeprootPilgrimageOracle, 0, 0)
	a := effectstest.B43Creature(g, me.ID, "Merfolk One", "Creature — Merfolk", 1, 1)
	b := effectstest.B43Creature(g, me.ID, "Merfolk Two", "Creature — Merfolk", 1, 1)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Merfolk", TypeLine: "Token Creature — Merfolk",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AttackWith(t, g, opp.ID, a, b, token)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Merfolk") - 1; n != 1 {
		t.Errorf("\"one or more\" is ONE trigger per batch: %d new tokens", n)
	}
}

func TestB43FinneasGrowsTokensAndRabbitsThenDrawsOnTenPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	finneas := effectstest.B43Catalog(g, me.ID, "Finneas, Ace Archer", "Legendary Creature — Rabbit Archer", effectstest.B43FinneasOracle, 2, 2)
	rabbit := effectstest.B43Creature(g, me.ID, "Regal Bunnicorn", "Creature — Rabbit", 4, 4)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, finneas)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, rabbit, game.CounterPlusOne) != 1 {
		t.Errorf("a Rabbit gets a counter: %d", effectstest.CounterOn(g, rabbit, game.CounterPlusOne))
	}
	if effectstest.CounterOn(g, token, game.CounterPlusOne) != 1 {
		t.Errorf("a token gets a counter: %d", effectstest.CounterOn(g, token, game.CounterPlusOne))
	}
	if effectstest.CounterOn(g, bear, game.CounterPlusOne) != 0 {
		t.Errorf("a nontoken non-Rabbit gets nothing: %d", effectstest.CounterOn(g, bear, game.CounterPlusOne))
	}
	if effectstest.CounterOn(g, finneas, game.CounterPlusOne) != 0 {
		t.Errorf("\"each OTHER creature\" excludes Finneas: %d", effectstest.CounterOn(g, finneas, game.CounterPlusOne))
	}
	// 2 + 5 + 4 + 2 = 13 after the counters, so the draw happens.
	if me.Hand.Size() != hand+1 {
		t.Errorf("total power 10 or more draws a card: hand %d → %d", hand, me.Hand.Size())
	}
}

// TestB43FinneasDrawsAfterAPausedPlacement is #1290's real-card exit
// criterion: with Doubling Season AND Hardened Scales both on the
// board, the Rabbit's +1/+1 counter placement pauses on a CR 616
// ordering prompt, and "total power 10 or greater" has to be read
// AFTER the placement lands — not while it is still owed to the
// prompt, which is where b13PutCounterOnEach used to leave it before
// the fix chained the batch through AddCounterByKeyedForEffect.
func TestB43FinneasDrawsAfterAPausedPlacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	finneas := effectstest.B43Catalog(g, me.ID, "Finneas, Ace Archer", "Legendary Creature — Rabbit Archer", effectstest.B43FinneasOracle, 2, 2)
	rabbit := effectstest.B43Creature(g, me.ID, "Regal Bunnicorn", "Creature — Rabbit", 5, 5)
	hand := me.Hand.Size()

	// declareAttack (not attackWith) stops at the declare-attackers
	// step rather than advancing on through it, which is what lets us
	// answer the CR 616 prompt before anything tries to walk past it.
	effectstest.DeclareAttack(t, g, opp.ID, finneas)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != hand {
		t.Fatalf("the draw happened with the Rabbit's +1/+1 counter still owed to the CR 616 prompt")
	}

	// Doubling Season first: 1 → 2 → 3.
	effectstest.AnswerOrderBySource(t, g, ds, hs)

	if effectstest.CounterOn(g, rabbit, game.CounterPlusOne) != 3 {
		t.Errorf("Rabbit +1/+1 counters = %d, want 3 ((1*2)+1 with Doubling Season first)", effectstest.CounterOn(g, rabbit, game.CounterPlusOne))
	}
	// finneas 2 (excluded from its own ability) + rabbit (5+3=8) = 10.
	if me.Hand.Size() != hand+1 {
		t.Errorf("total power 10 or more draws a card once the paused placement lands: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestB43CalderaPyremawGrowsThenBurnsForItsNewPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pyremaw := effectstest.B43Catalog(g, me.ID, "Caldera Pyremaw", "Creature — Dragon", effectstest.B43CalderaPyremawOracle, 3, 3)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, pyremaw, game.CounterPlusOne) != 1 {
		t.Fatalf("the counter goes on first: %d", effectstest.CounterOn(g, pyremaw, game.CounterPlusOne))
	}
	if opp.Life != before-4 {
		t.Errorf("a 3/3 with a fresh counter deals 4: %d → %d", before, opp.Life)
	}
}

// TestB43CalderaPyremawBurnsAfterAPausedPlacement is #1290's real-card
// exit criterion: with Doubling Season AND Hardened Scales both on the
// board, the +1/+1 counter placement pauses on a CR 616 ordering
// prompt, and the damage has to be measured off the power the Dragon
// has AFTER the counter lands.
func TestB43CalderaPyremawBurnsAfterAPausedPlacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	pyremaw := effectstest.B43Catalog(g, me.ID, "Caldera Pyremaw", "Creature — Dragon", effectstest.B43CalderaPyremawOracle, 3, 3)
	before := opp.Life

	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != before {
		t.Fatalf("Caldera Pyremaw dealt damage with its +1/+1 counter still owed to the CR 616 prompt")
	}

	// Doubling Season first: 1 → 2 → 3.
	effectstest.AnswerOrderBySource(t, g, ds, hs)

	if effectstest.CounterOn(g, pyremaw, game.CounterPlusOne) != 3 {
		t.Errorf("Caldera Pyremaw +1/+1 counters = %d, want 3 ((1*2)+1 with Doubling Season first)", effectstest.CounterOn(g, pyremaw, game.CounterPlusOne))
	}
	if opp.Life != before-6 {
		t.Errorf("opponent life %d → %d, want -6 (3 base + 3 counters) — damage was measured before the paused placement landed", before, opp.Life)
	}
}

func TestB43EowynNeedsAnotherHumanThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Catalog(g, me.ID, "Éowyn, Shieldmaiden", "Legendary Creature — Human Knight", effectstest.B43EowynOracle, 5, 4)
	// Nothing entered this turn: no trigger.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Human Knight"); n != 0 {
		t.Fatalf("no Human entered this turn, so nothing happens: %d tokens", n)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.CastCatalogSpell(t, g, "Human Soldier", "Creature — Human Soldier", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Human Knight"); n != 2 {
		t.Errorf("a Human entered this turn: two 2/2 Knights: %d", n)
	}
	_ = opp
}

func TestB43DistinguishedConjurerGainsLifeAndBlinksAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	conjurer := effectstest.B43Catalog(g, me.ID, "Distinguished Conjurer", "Creature — Human Wizard", effectstest.B43DistinguishedConjurerOra, 1, 2)
	bear := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("another creature you control entering gains 1 life: %d → %d", before, me.Life)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "C", "C", "C", "C")
	// The Conjurer cannot blink itself — "another".
	if err := g.ActivateCatalogAbility(me.ID, conjurer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(conjurer)}); err == nil {
		t.Fatal("\"another target creature\" excludes the Conjurer")
	}
	effectstest.B16Activate(t, g, me.ID, conjurer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if g.Battlefield.Contains(bear) {
		t.Error("a blinked permanent returns as a NEW object, with a new instance")
	}
	returned, ok := effectstest.B43BattlefieldByName(g, "Bear")
	if !ok {
		t.Fatal("and it does come back")
	}
	if effectstest.CounterOn(g, returned, game.CounterPlusOne) != 0 {
		t.Error("counters fall off the new object")
	}
}

func TestB43ZephyrBootsFlyAndLootOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	boots := effectstest.B43Catalog(g, me.ID, "Zephyr Boots", "Artifact — Equipment", effectstest.B43ZephyrBootsOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, boots, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.AssertKeywords(t, g, bear, "flying")
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("the draw happens before the discard prompt: hand %d → %d", hand, me.Hand.Size())
	}
	if !anyPendingChoiceFor(g, me.ID) {
		t.Error("and then it asks you to discard")
	}
}

func TestB43CryptcallerChariotMakesATappedZombiePerDiscardAndCrews(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chariot := effectstest.B43Catalog(g, me.ID, "Cryptcaller Chariot", "Artifact — Vehicle", effectstest.B43CryptcallerChariotOracle, 5, 5)
	if effectstest.IsCreatureNow(g, chariot) {
		t.Fatal("a Vehicle is not a creature until it is crewed")
	}
	card := me.Hand.Cards[0].InstanceID
	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	effectstest.AnswerDiscard(t, g, me.ID, card)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B43TokensNamed(g, me.ID, "Zombie"); n != 1 {
		t.Fatalf("one discard is one Zombie: %d", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && !c.Tapped {
			t.Error("the Zombies arrive tapped")
		}
	}
	crew := effectstest.B43Creature(g, me.ID, "Big Bear", "Creature — Bear", 3, 3)
	_ = crew
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, chariot, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crew}})
	if !effectstest.IsCreatureNow(g, chariot) {
		t.Error("crewed, it is an artifact creature")
	}
	effectstest.AssertKeywords(t, g, chariot, "menace")
}

// --- activated abilities -------------------------------------------

func TestB43CaptainSisayTutorsOnlyLegendaryCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Sol Ring", TypeLine: "Legendary Artifact"},
		game.Card{Name: "Kozilek", TypeLine: "Legendary Creature — Eldrazi"},
		game.Card{Name: "Signet", TypeLine: "Artifact"},
	)
	sisay := effectstest.B43Catalog(g, me.ID, "Captain Sisay", "Legendary Creature — Human Soldier", effectstest.B43CaptainSisayOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sisay, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the tap opens a search")
	}
	if effectstest.SearchOptionNamed(g, c, "Signet") != uuid.Nil {
		t.Error("a nonlegendary card is not offered")
	}
	ring := effectstest.SearchOptionNamed(g, c, "Sol Ring")
	if ring == uuid.Nil {
		t.Fatal("a legendary ARTIFACT is a legal find — the clause says \"card\"")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ring)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ring) {
		t.Error("it goes to your hand")
	}
	if !effectstest.B16Tapped(t, g, sisay) {
		t.Error("Sisay taps to activate")
	}
}

func TestB43PlanarBridgePutsAnyPermanentCardOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Big Eldrazi", TypeLine: "Creature — Eldrazi"},
		game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Some Instant", TypeLine: "Instant"},
	)
	bridge := effectstest.B43Catalog(g, me.ID, "Planar Bridge", "Legendary Artifact", effectstest.B43PlanarBridgeOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bridge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the activation opens a search")
	}
	if effectstest.SearchOptionNamed(g, c, "Some Instant") != uuid.Nil {
		t.Error("an instant is not a permanent card")
	}
	eldrazi := effectstest.SearchOptionNamed(g, c, "Big Eldrazi")
	if eldrazi == uuid.Nil {
		t.Fatal("the creature card is offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, eldrazi)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(eldrazi) {
		t.Error("it arrives on the battlefield")
	}
	if effectstest.B16Tapped(t, g, eldrazi) {
		t.Error("untapped — the card says nothing about tapped")
	}
}

func TestB43VizkopaGuildmageGrantsLifelinkUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B43Catalog(g, me.ID, "Vizkopa Guildmage", "Creature — Human Wizard", effectstest.B43VizkopaGuildmageOracle, 2, 2)
	bear := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spec, _ := effects.Lookup(effectstest.B43VizkopaGuildmageOracle)
	if len(spec.Activated) != 2 {
		t.Fatalf("both abilities ship: %d", len(spec.Activated))
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "B", "C")
	effectstest.B16Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.AssertKeywords(t, g, bear, "lifelink")
}

// --- cascade -------------------------------------------------------

func TestB43MaelstromColossusCascadesOnCast(t *testing.T) {
	if got := len(game.CatalogTriggers(effectstest.B43MaelstromColossusOracle)); got != 1 {
		t.Fatalf("one cascade trigger, got %d", got)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 3)
	effectstest.CastWithCost(t, g, "Maelstrom Colossus", "Artifact Creature — Golem", "{8}", effectstest.B43MaelstromColossusOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers the free cast: %d, want 1", got)
	}
	if n := effectstest.AnswerAllMayCast(t, g, me.ID, false); n != 1 {
		t.Errorf("answered %d offers, want 1", n)
	}
}

// --- modal triggers and multi-clause targets (#937) -----------------

// Skemfar Shadowsage's mode is chosen as the trigger goes on the
// stack (CR 603.3c), and X is the biggest single tribe on your board
// — not the sum of your creatures.
func TestB43SkemfarShadowsageChoosesAModeThenCountsTheBiggestTribe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Three Elves (two of them also Clerics) and a Bear: the biggest
	// single type is Elf at 3, and the Shadowsage itself makes 4.
	effectstest.B43Creature(g, me.ID, "Elf One", "Creature — Elf Druid", 1, 1)
	effectstest.B43Creature(g, me.ID, "Elf Two", "Creature — Elf Cleric", 1, 1)
	effectstest.B43Creature(g, me.ID, "Elf Three", "Creature — Elf Cleric", 1, 1)
	effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B43Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 1, 1)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Skemfar Shadowsage", "Creature — Elf Cleric", effectstest.B43SkemfarShadowsageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger asks which mode, as it goes on the stack")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Four Elves of yours; the opponent's Elf does not count, and the
	// Bear is not in any tribe with anything.
	if opp.Life != before-4 {
		t.Errorf("each opponent loses 4: %d → %d", before, opp.Life)
	}
}

func TestB43SkemfarShadowsageSecondModeGainsTheSameAmount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B43Creature(g, me.ID, "Zombie One", "Creature — Zombie", 2, 2)
	effectstest.B43Creature(g, me.ID, "Zombie Two", "Creature — Zombie", 2, 2)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Skemfar Shadowsage", "Creature — Elf Cleric", effectstest.B43SkemfarShadowsageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the mode prompt is there")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Two Zombies is the biggest tribe; the Shadowsage is an Elf Cleric
	// and is a tribe of one on its own.
	if me.Life != before+2 {
		t.Errorf("you gain 2: %d → %d", before, me.Life)
	}
}

// Gimli's haste is unconditional; the fight waits for total power 8.
func TestB43GimlisRecklessMightGivesHasteAndFightsOnlyAtEightPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Catalog(g, me.ID, "Gimli's Reckless Might", "Enchantment", effectstest.B43GimlisRecklessMightOracl, 0, 0)
	small := effectstest.B43Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AssertKeywords(t, g, small, "haste")
	// Total power 2: formidable is not met, so no trigger at all.
	effectstest.AttackWith(t, g, opp.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("total power 2 is not 8: the trigger does not fire")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("and nothing fought")
	}
}

func TestB43GimlisRecklessMightFightsWithTwoDifferentTargetClauses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Catalog(g, me.ID, "Gimli's Reckless Might", "Enchantment", effectstest.B43GimlisRecklessMightOracl, 0, 0)
	big := effectstest.B43Creature(g, me.ID, "Big Bear", "Creature — Bear", 5, 5)
	other := effectstest.B43Creature(g, me.ID, "Other Bear", "Creature — Bear", 4, 4)
	theirs := effectstest.B43Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	// The trigger's target prompt opens as the attack declaration is
	// locked in, and it asks CLAUSE BY CLAUSE: the attacker you control
	// first, then the creature it fights.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(big, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	_, _ = g.AdvanceStep()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("total power 9: formidable is met and the trigger asks for targets")
	}
	if effectstest.HasID(p.PickTargetCards, other) {
		t.Error("the first clause is an ATTACKING creature you control — the one at home is not legal")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err != nil {
		t.Fatalf("ResolvePickTarget (the attacker): %v", err)
	}
	p = effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("then the second clause asks for the creature it fights")
	}
	if effectstest.HasID(p.PickTargetCards, big) {
		t.Error("the second clause is a creature you DON'T control")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatalf("ResolvePickTarget (the victim): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B43Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	jitte := effectstest.B43Catalog(g, me.ID, "Lost Jitte", "Legendary Artifact — Equipment", effectstest.B43LostJitteOracle, 0, 0)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, jitte, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})

	// Blocked, so the damage never reaches a player — the Jitte still
	// banks a counter, which is what "deals combat damage" means.
	effectstest.AttackIntoBlocks(t, g, opp.ID, bear)
	if err := g.DeclareBlocker(blocker, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, jitte, "charge") < 1 {
		t.Fatalf("equipped creature dealt combat damage: a charge counter: %d", effectstest.CounterOn(g, jitte, "charge"))
	}

	// Spend it on the untap-a-land bullet.
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(land) })
	effectstest.AdvanceToMainOf(t, g, 0)
	if err := g.ActivateCatalogAbility(me.ID, jitte, 0, game.ActivateAbilityParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility (untap a land): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, land) {
		t.Error("the untap bullet untapped the land")
	}
	if effectstest.CounterOn(g, jitte, "charge") != 0 {
		t.Errorf("the counter was the cost and is gone: %d", effectstest.CounterOn(g, jitte, "charge"))
	}
}

// --- #1412: cycling ---------------------------------------------------

// Agonasaur Rex: cycling {2}{G} discards the Rex and draws a card,
// and the "when you cycle this card" trigger fires from the
// graveyard (CR 702.29c) to put two +1/+1 counters on the chosen
// creature or Vehicle and grant it trample and indestructible until
// end of turn.
func TestB43AgonasaurRexCyclesForACardAndTriggersItsBonus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.B43Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	id, _ := effectstest.CycleFromHand(t, g, "Agonasaur Rex", "Creature — Dinosaur", effectstest.B43AgonasaurRexOracle, "{C}{C}{G}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Rex is not in the graveyard")
	}
	handBefore := len(me.Hand.Cards)

	// The cycle trigger's target prompt is already open by the time
	// activation returns — building a targeted trigger runs
	// synchronously with the event that fires it, the same way the
	// trigger_order prompt appears synchronously for Magmakin
	// Artillerist's two triggers.
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", handBefore, got)
	}
	if got := effectstest.CounterOn(g, target, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters on the target: got %d, want 2", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, target, "trample") {
		t.Error("the cycle trigger did not grant trample")
	}
	if !effectstest.HasEffectiveKeyword(t, g, target, "indestructible") {
		t.Error("the cycle trigger did not grant indestructible")
	}
}

// Cycling is refused without the mana to pay for it — the discard and
// the draw never happen.
func TestB43AgonasaurRexCyclingRefusedWithoutMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := effectstest.PushCatalogHandCard(me, "Agonasaur Rex", "Creature — Dinosaur", effectstest.B43AgonasaurRexOracle)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{2}{G} must be paid to cycle")
	}
	if !me.Hand.Contains(id) {
		t.Error("the Rex was discarded despite the unpaid cost")
	}
}
