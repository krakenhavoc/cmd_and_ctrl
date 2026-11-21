package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m21_5511_b1_cards_test.go — Core Set 2021 batch 1 (#5511). Every test
// plays the real catalog entry through the engine.

const (
	m21B1AlpineWatchdogOracle      = "6a0419c3-4bfa-46c1-932b-fe10f2fa2ad7"
	m21B1BloodGluttonOracle        = "065b8e5e-874e-4de3-9f12-41e4b1604561"
	m21B1GarruksGorehornOracle     = "c8384392-1914-4423-9f49-92f5643d15dc"
	m21B1StaunchShieldmateOracle   = "0a1a0911-dcd5-4a2f-8deb-abb5b273b0fb"
	m21B1MistralSingerOracle       = "863a5ddb-2d66-4a07-bafc-38290776d5c0"
	m21B1ForgottenSentinelOracle   = "9bd4a0bd-d10c-434a-8254-89f38740aafd"
	m21B1WardedBattlementsOracle   = "7616dacf-44e2-4b03-b120-875cc2e50ea5"
	m21B1SigiledContenderOracle    = "84ec7820-6608-459c-b172-80e8e3750cf4"
	m21B1KineticAugurOracle        = "98e0b1a8-902d-4285-b13c-e63f2e1692f1"
	m21B1ChandrasMagmuttOracle     = "a2c4537e-70a6-416d-9e8b-f549ad535147"
	m21B1IgneousCurOracle          = "0d899095-cef7-43d2-aaf7-eb73b62f5ec8"
	m21B1TeferisProtegeOracle      = "e84f1727-ce01-44be-aebf-8a0c4e032c9a"
	m21B1AnointedChoristerOracle   = "61859443-e298-41c7-a179-6e6554de91c9"
	m21B1SkywaySniperOracle        = "8d907825-25f9-4b25-9fac-e0a207ed1797"
	m21B1HobblefiendOracle         = "f2b284cd-fa12-4ba1-b297-80105bb48fde"
	m21B1CagedZombieOracle         = "430b3e61-09a5-4390-b377-89830cacf66e"
	m21B1TemperedVeteranOracle     = "c19f3f00-2c7b-4e85-a034-e65760d9cc13"
	m21B1SeasonedHallowbladeOracle = "d03dbb60-e32b-4b74-ae8a-73d428d344cd"
	m21B1BoltHoundOracle           = "60555f06-dc85-4a34-ae3c-fc1cdb923be7"
	m21B1BonePitBruteOracle        = "6de38910-8d0d-4fac-b9db-c80ffd1850ea"
	m21B1GaleSwooperOracle         = "1758e102-22da-4cb3-98d5-76f7501d026d"
	m21B1LibraryLarcenistOracle    = "eb0e6f63-2861-4392-ac0e-64380fbbd8ae"
	m21B1SpinedMegalodonOracle     = "3a124fe5-f525-43e5-9746-501c68a98330"
	m21B1SpiritOfMalevolenceOracle = "15323e45-8260-43dc-b72b-efd661a58d75"
	m21B1StormCallerOracle         = "5c946cc0-5d90-4970-acf8-b14eff177cbc"
	m21B1FalconerAdeptOracle       = "e2106eba-15f5-42b5-b797-6582ae94f3de"
	m21B1TwinbladeAssassinsOracle  = "66093486-c9c3-42c5-97d1-ec0fe54f65b0"
	m21B1SabertoothMaulerOracle    = "da4f4094-b0ce-4b5b-8ac3-8c0a5159220f"
	m21B1GloomSowerOracle          = "e5cc1492-c9cc-48f0-8c11-02ead883675c"
	m21B1WardenOfTheWoodsOracle    = "c8c6924e-9af8-4212-be28-b0b4eaa4c7d8"
	m21B1BurlfistOakOracle         = "95a08f25-8d44-4704-a4d1-a098bf010932"
	m21B1RiseAgainOracle           = "4600e968-fe3a-446a-a306-bd10e6763e0d"
	m21B1FinishingBlowOracle       = "3c8a6c58-f4af-4809-af86-a931313fa71e"
	m21B1EliminateOracle           = "cf307f81-488c-4df5-8b08-ba46e3e7f3d9"
	m21B1AngelicAscensionOracle    = "5a898412-5b1e-4217-ac78-5b4099ed8e16"
	m21B1SecureTheSceneOracle      = "d6a9746c-eba5-4735-aec1-e70edba2c686"
	m21B1MiscastOracle             = "e08c71aa-d35c-4c20-b4db-ae9d36ee3169"
	m21B1ReadTheTidesOracle        = "e5c4ed40-5069-4606-81f2-59dcfa1345bd"
	m21B1RunAfoulOracle            = "5462cb38-7cd7-46e8-a273-35d0d1af45b4"
	m21B1BadDealOracle             = "47258c3e-b4a7-466d-9464-3924f75d9293"
	m21B1RookieMistakeOracle       = "e51680d1-3e6c-4d1b-9705-36f73d7328cf"

	m21B1Bear = "Creature — Bear"
)

// m21B1ControlledNamed counts the battlefield permanents named name that
// controller controls.
func m21B1ControlledNamed(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller {
				n++
			}
		}
	})
	return n
}

// m21B1Flyer puts a 1/1 creature with flying onto owner's battlefield.
func m21B1Flyer(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
}

// --- every card is in the catalog ----------------------------------------

func TestM21B1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m21B1AlpineWatchdogOracle:      "Alpine Watchdog",
		m21B1BloodGluttonOracle:        "Blood Glutton",
		m21B1GarruksGorehornOracle:     "Garruk's Gorehorn",
		m21B1StaunchShieldmateOracle:   "Staunch Shieldmate",
		m21B1MistralSingerOracle:       "Mistral Singer",
		m21B1ForgottenSentinelOracle:   "Forgotten Sentinel",
		m21B1WardedBattlementsOracle:   "Warded Battlements",
		m21B1SigiledContenderOracle:    "Sigiled Contender",
		m21B1KineticAugurOracle:        "Kinetic Augur",
		m21B1ChandrasMagmuttOracle:     "Chandra's Magmutt",
		m21B1IgneousCurOracle:          "Igneous Cur",
		m21B1TeferisProtegeOracle:      "Teferi's Protege",
		m21B1AnointedChoristerOracle:   "Anointed Chorister",
		m21B1SkywaySniperOracle:        "Skyway Sniper",
		m21B1HobblefiendOracle:         "Hobblefiend",
		m21B1CagedZombieOracle:         "Caged Zombie",
		m21B1TemperedVeteranOracle:     "Tempered Veteran",
		m21B1SeasonedHallowbladeOracle: "Seasoned Hallowblade",
		m21B1BoltHoundOracle:           "Bolt Hound",
		m21B1BonePitBruteOracle:        "Bone Pit Brute",
		m21B1GaleSwooperOracle:         "Gale Swooper",
		m21B1LibraryLarcenistOracle:    "Library Larcenist",
		m21B1SpinedMegalodonOracle:     "Spined Megalodon",
		m21B1SpiritOfMalevolenceOracle: "Spirit of Malevolence",
		m21B1StormCallerOracle:         "Storm Caller",
		m21B1FalconerAdeptOracle:       "Falconer Adept",
		m21B1TwinbladeAssassinsOracle:  "Twinblade Assassins",
		m21B1SabertoothMaulerOracle:    "Sabertooth Mauler",
		m21B1GloomSowerOracle:          "Gloom Sower",
		m21B1WardenOfTheWoodsOracle:    "Warden of the Woods",
		m21B1BurlfistOakOracle:         "Burlfist Oak",
		m21B1RiseAgainOracle:           "Rise Again",
		m21B1FinishingBlowOracle:       "Finishing Blow",
		m21B1EliminateOracle:           "Eliminate",
		m21B1AngelicAscensionOracle:    "Angelic Ascension",
		m21B1SecureTheSceneOracle:      "Secure the Scene",
		m21B1MiscastOracle:             "Miscast",
		m21B1ReadTheTidesOracle:        "Read the Tides",
		m21B1RunAfoulOracle:            "Run Afoul",
		m21B1BadDealOracle:             "Bad Deal",
		m21B1RookieMistakeOracle:       "Rookie Mistake",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- printed keywords -----------------------------------------------------------

func TestM21B1PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Alpine Watchdog", "Creature — Dog", m21B1AlpineWatchdogOracle, 2, 2, []string{"vigilance"}},
		{"Blood Glutton", "Creature — Vampire", m21B1BloodGluttonOracle, 4, 3, []string{"lifelink"}},
		{"Mistral Singer", "Creature — Siren", m21B1MistralSingerOracle, 2, 2, []string{"flying", "prowess"}},
		{"Warded Battlements", "Creature — Wall", m21B1WardedBattlementsOracle, 0, 3, []string{"defender"}},
		{"Kinetic Augur", "Creature — Human Shaman", m21B1KineticAugurOracle, 0, 4, []string{"trample"}},
		{"Anointed Chorister", "Creature — Human Cleric", m21B1AnointedChoristerOracle, 1, 1, []string{"lifelink"}},
		{"Skyway Sniper", "Creature — Elf Archer", m21B1SkywaySniperOracle, 1, 2, []string{"reach"}},
		{"Hobblefiend", "Creature — Devil", m21B1HobblefiendOracle, 2, 1, []string{"trample"}},
		{"Bolt Hound", "Creature — Elemental Dog", m21B1BoltHoundOracle, 2, 2, []string{"haste"}},
		{"Bone Pit Brute", "Creature — Cyclops", m21B1BonePitBruteOracle, 4, 5, []string{"menace"}},
		{"Gale Swooper", "Creature — Griffin", m21B1GaleSwooperOracle, 3, 2, []string{"flying"}},
		{"Spined Megalodon", "Creature — Shark", m21B1SpinedMegalodonOracle, 5, 7, []string{"hexproof"}},
		{"Warden of the Woods", "Creature — Treefolk", m21B1WardenOfTheWoodsOracle, 5, 7, []string{"vigilance"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.AssertKeywords(t, g, id, c.keywords...)
		})
	}
}

// The two vanilla bodies have no abilities to exercise: they resolve as
// plain creatures.
func TestM21B1VanillaCreaturesResolve(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Garruk's Gorehorn", "Creature — Beast", m21B1GarruksGorehornOracle, 7, 3},
		{"Staunch Shieldmate", "Creature — Dwarf Soldier", m21B1StaunchShieldmateOracle, 1, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastHoldCreature(t, g, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			if p, tough := effectstest.PtOf(t, g, id); p != c.power || tough != c.toughness {
				t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.power, c.toughness)
			}
		})
	}
}

// --- statics --------------------------------------------------------------------

func TestM21B1ForgottenSentinelEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastHoldCreature(t, g, "Forgotten Sentinel", "Artifact Creature — Golem", m21B1ForgottenSentinelOracle, 4, 3)
	if !effectstest.HmlTapped(t, g, id) {
		t.Error("Forgotten Sentinel entered untapped")
	}
}

func TestM21B1WardedBattlementsPumpsOnlyAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Warded Battlements", "Creature — Wall", m21B1WardedBattlementsOracle, 0, 3)
	raider := effectstest.B12Creature(g, me.ID, "Raider", m21B1Bear, 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", m21B1Bear, 2, 2)
	if p, _ := effectstest.PtOf(t, g, raider); p != 2 {
		t.Fatalf("a creature that isn't attacking has power %d, want 2", p)
	}
	effectstest.BrAttack(t, g, raider)
	if p, tough := effectstest.PtOf(t, g, raider); p != 3 || tough != 2 {
		t.Errorf("the attacker is %d/%d, want 3/2", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, idle); p != 2 {
		t.Errorf("a creature that isn't attacking has power %d, want 2", p)
	}
}

func TestM21B1SigiledContenderHasLifelinkOnlyWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Sigiled Contender", "Creature — Human Warrior", m21B1SigiledContenderOracle, 3, 3)
	if effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
		t.Fatal("lifelink without a +1/+1 counter")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, 1) })
	if !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
		t.Error("no lifelink with a +1/+1 counter")
	}
}

// Kinetic Augur discards two sorceries as it enters, draws two, and its
// power counts the two sorceries now in the graveyard.
func TestM21B1KineticAugurLootsAndCountsInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Draw A", "Draw B", "Draw C")
	a, b := uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: a, Name: "Junk A", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: b, Name: "Junk B", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	augur := effectstest.CastHoldCreature(t, g, "Kinetic Augur", "Creature — Human Shaman", m21B1KineticAugurOracle, 0, 4)
	if p, _ := effectstest.PtOf(t, g, augur); p != 0 {
		t.Fatalf("power %d with no instant or sorcery in the graveyard, want 0", p)
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no discard prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after discarding two and drawing two, want %d", got, hand)
	}
	if p, _ := effectstest.PtOf(t, g, augur); p != 2 {
		t.Errorf("power %d with two sorceries in the graveyard, want 2", p)
	}
}

// --- activated abilities --------------------------------------------------------

func TestM21B1ChandrasMagmuttPingsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mutt := effectstest.B12Push(g, me.ID, "Chandra's Magmutt", "Creature — Elemental Dog", m21B1ChandrasMagmuttOracle, 2, 2)
	life := opp.Life
	effectstest.ApaActivate(t, g, me, mutt, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-1)
	}
	if !effectstest.HmlTapped(t, g, mutt) {
		t.Error("the Magmutt did not tap")
	}
}

func TestM21B1PumpRows(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		mana                   []string
		wantP, wantT           int
	}{
		{"Igneous Cur", "Creature — Elemental Dog", m21B1IgneousCurOracle, 1, 2, []string{"R", "R"}, 3, 2},
		{"Anointed Chorister", "Creature — Human Cleric", m21B1AnointedChoristerOracle, 1, 1, []string{"W", "W", "W", "W", "W"}, 4, 4},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.B12Push(g, me.ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.ApaMana(me, c.mana...)
			effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if p, tough := effectstest.PtOf(t, g, id); p != c.wantP || tough != c.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.wantP, c.wantT)
			}
		})
	}
}

func TestM21B1TeferisProtegeLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	protege := effectstest.B12Push(g, me.ID, "Teferi's Protege", "Creature — Human Wizard", m21B1TeferisProtegeOracle, 2, 3)
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "U", "U")
	effectstest.ApaActivate(t, g, me, protege, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestM21B1SkywaySniperShootsOnlyAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	sniper := effectstest.B12Push(g, me.ID, "Skyway Sniper", "Creature — Elf Archer", m21B1SkywaySniperOracle, 1, 2)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", m21B1Bear, 2, 2)
	flyer := m21B1Flyer(g, opp.ID, "Flyer")
	effectstest.ApaMana(me, "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, sniper, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(walker)}); err == nil {
		t.Fatal("the Sniper targeted a creature without flying")
	}
	effectstest.ApaActivate(t, g, me, sniper, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(flyer)})
	if g.Battlefield.Contains(flyer) {
		t.Error("the 1/1 flyer survived 1 damage")
	}
}

func TestM21B1HobblefiendSacrificesAnotherForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	fiend := effectstest.B12Push(g, me.ID, "Hobblefiend", "Creature — Devil", m21B1HobblefiendOracle, 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m21B1Bear, 2, 2)
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, fiend, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fiend}}); err == nil {
		t.Fatal("Hobblefiend sacrificed itself")
	}
	effectstest.ApaActivate(t, g, me, fiend, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if n := effectstest.PlusOnes(g, fiend); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
}

func TestM21B1CagedZombieNeedsADeathThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	zombie := effectstest.B12Push(g, me.ID, "Caged Zombie", "Creature — Zombie", m21B1CagedZombieOracle, 2, 3)
	victim := effectstest.B12Creature(g, me.ID, "Victim", m21B1Bear, 1, 1)
	effectstest.ApaMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, zombie, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated before any creature died")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(victim); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.ApaActivate(t, g, me, zombie, 0, game.ActivateAbilityParams{})
	for _, s := range g.Seats {
		want := lives[s.ID] - 2
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestM21B1TemperedVeteranCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	vet := effectstest.B12Push(g, me.ID, "Tempered Veteran", "Creature — Human Knight", m21B1TemperedVeteranOracle, 1, 2)
	vet2 := effectstest.B12Push(g, me.ID, "Tempered Veteran", "Creature — Human Knight", m21B1TemperedVeteranOracle, 1, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare", m21B1Bear, 2, 2)
	countered := effectstest.B12Creature(g, me.ID, "Countered", m21B1Bear, 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(countered, game.CounterPlusOne, 1) })
	effectstest.ApaMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, vet, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bare)}); err == nil {
		t.Fatal("the {W} row targeted a creature with no +1/+1 counter")
	}
	effectstest.ApaActivate(t, g, me, vet, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(countered)})
	if n := effectstest.PlusOnes(g, countered); n != 2 {
		t.Errorf("the countered creature has %d counters, want 2", n)
	}
	effectstest.ApaMana(me, "W", "W", "W", "W", "W", "W")
	effectstest.ApaActivate(t, g, me, vet2, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bare)})
	if n := effectstest.PlusOnes(g, bare); n != 1 {
		t.Errorf("the bare creature has %d counters, want 1", n)
	}
}

func TestM21B1SeasonedHallowbladeTapsAndGainsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	blade := effectstest.B12Push(g, me.ID, "Seasoned Hallowblade", "Creature — Human Warrior", m21B1SeasonedHallowbladeOracle, 3, 1)
	card := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: card, Name: "Filler", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.ApaActivate(t, g, me, blade, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}})
	if me.Hand.Contains(card) {
		t.Fatal("the discard was not paid")
	}
	if !effectstest.HmlTapped(t, g, blade) {
		t.Error("not tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, blade, "indestructible") {
		t.Error("no indestructible")
	}
	effectstest.B27Kill(g, blade)
	if !g.Battlefield.Contains(blade) {
		t.Error("the indestructible Hallowblade was destroyed")
	}
}

// --- attack triggers ------------------------------------------------------------

func TestM21B1BoltHoundPumpsTheOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hound := effectstest.B12Push(g, me.ID, "Bolt Hound", "Creature — Elemental Dog", m21B1BoltHoundOracle, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", m21B1Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m21B1Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, hound)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, other); p != 3 || tough != 2 {
		t.Errorf("your other creature is %d/%d, want 3/2", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, hound); p != 2 {
		t.Errorf("the Hound has power %d, want 2", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestM21B1LibraryLarcenistDrawsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	l := effectstest.B12Push(g, me.ID, "Library Larcenist", "Creature — Merfolk Rogue", m21B1LibraryLarcenistOracle, 1, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, l)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestM21B1SpinedMegalodonScriesOneWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two")
	shark := effectstest.B12Push(g, me.ID, "Spined Megalodon", "Creature — Shark", m21B1SpinedMegalodonOracle, 5, 7)
	effectstest.DeclareAttack(t, g, opp.ID, shark)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scry 1 offered %d cards", len(c.ScryCards))
	}
}

func TestM21B1FalconerAdeptMakesATappedAttackingBird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adept := effectstest.B12Push(g, me.ID, "Falconer Adept", "Creature — Human Soldier", m21B1FalconerAdeptOracle, 2, 3)
	effectstest.DeclareAttack(t, g, opp.ID, adept)
	effectstest.PassPriorityAroundTable(t, g)
	birds := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name != "Bird" || c.Controller != me.ID {
				continue
			}
			birds++
			if !c.Tapped {
				t.Error("the Bird enters tapped")
			}
			if c.AttackingTarget != opp.ID {
				t.Error("and attacking the player the Adept is attacking")
			}
		}
	})
	if birds != 1 {
		t.Errorf("%d Birds, want 1", birds)
	}
}

// --- enters and dies triggers ---------------------------------------------------

func TestM21B1EntersTargetedTriggers(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		wantP, wantT           int
		flying                 bool
	}{
		{"Bone Pit Brute", "Creature — Cyclops", m21B1BonePitBruteOracle, 4, 5, 6, 2, false},
		{"Gale Swooper", "Creature — Griffin", m21B1GaleSwooperOracle, 3, 2, 2, 2, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.B12Creature(g, me.ID, "Bear", m21B1Bear, 2, 2)
			effectstest.CastHoldCreature(t, g, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.PickCard(t, g, me.ID, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != c.wantP || tough != c.wantT {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, c.wantP, c.wantT)
			}
			if got := effectstest.HasEffectiveKeyword(t, g, bear, "flying"); got != c.flying {
				t.Errorf("the Bear has flying %v, want %v", got, c.flying)
			}
		})
	}
}

func TestM21B1StormCallerDealsTwoToEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.CastHoldCreature(t, g, "Storm Caller", "Creature — Ogre Shaman", m21B1StormCallerOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 2
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestM21B1SpiritOfMalevolenceDrainsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := effectstest.B12Push(g, me.ID, "Spirit of Malevolence", "Creature — Spirit", m21B1SpiritOfMalevolenceOracle, 2, 1)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.B27Kill(g, spirit)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 1
		if s.ID == me.ID {
			want = lives[s.ID] + 1
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

// --- end-step triggers ----------------------------------------------------------

func TestM21B1TwinbladeAssassinsDrawOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Twinblade Assassins", "Creature — Elf Assassin", m21B1TwinbladeAssassinsOracle, 5, 4)
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", m21B1Bear, 2, 2)
		effectstest.AdvanceToMain(t, g)
		hand := me.Hand.Size()
		if died {
			effectstest.KillCreature(t, g, opp.ID, bear)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if died {
			want = hand + 1
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("a creature died %v: hand %d, want %d", died, got, want)
		}
	}
}

func TestM21B1SabertoothMaulerGrowsAndUntapsAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		cat := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Sabertooth Mauler", OracleID: m21B1SabertoothMaulerOracle,
			TypeLine: "Creature — Cat", Power: 3, Toughness: 3, Tapped: true})
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", m21B1Bear, 2, 2)
		effectstest.AdvanceToMain(t, g)
		if died {
			effectstest.KillCreature(t, g, opp.ID, bear)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if died {
			want = 1
		}
		if got := effectstest.PlusOnes(g, cat); got != want {
			t.Errorf("a creature died %v: %d counters, want %d", died, got, want)
		}
		if got := effectstest.HmlTapped(t, g, cat); got != !died {
			t.Errorf("a creature died %v: tapped %v, want %v", died, got, !died)
		}
	}
}

// --- other triggers -------------------------------------------------------------

func TestM21B1GloomSowerDrainsEachBlockersController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	sower := effectstest.B12Push(g, me.ID, "Gloom Sower", "Creature — Horror", m21B1GloomSowerOracle, 8, 6)
	wall := effectstest.B12Creature(g, them.ID, "Wall", "Creature — Wall", 0, 9)
	mine, theirs := me.Life, them.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sower, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, sower); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.InitSettle(t, g)
	if them.Life != theirs-2 {
		t.Errorf("the blocker's controller is at %d, want %d", them.Life, theirs-2)
	}
	if me.Life != mine+2 {
		t.Errorf("you are at %d, want %d", me.Life, mine+2)
	}
}

func TestM21B1WardenOfTheWoodsMayDrawTwoWhenAnOpponentTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	warden := effectstest.B12Push(g, me.ID, "Warden of the Woods", "Creature — Treefolk", m21B1WardenOfTheWoodsOracle, 5, 7)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(warden))
	prompt := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if prompt.Source != warden {
		t.Fatalf("the trigger prompt is from %v, want the Warden", prompt.Source)
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("answer the Warden: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	mine := g2.Seats[0]
	w2 := effectstest.B12Push(g2, mine.ID, "Warden of the Woods", "Creature — Treefolk", m21B1WardenOfTheWoodsOracle, 5, 7)
	effectstest.CastCatalogSpell(t, g2, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(w2))
	if effectstest.TriggerOnStack(g2, w2) != nil || effectstest.LatestTriggerPrompt(g2, mine.ID) != nil {
		t.Error("your own spell triggered it")
	}
}

func TestM21B1BurlfistOakGrowsOnEachOfYourDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	oak := effectstest.B12Push(g, me.ID, "Burlfist Oak", "Creature — Treefolk", m21B1BurlfistOakOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, oak); p != 4 || tough != 5 {
		t.Fatalf("the Oak is %d/%d after one draw, want 4/5", p, tough)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("opponent draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, oak); p != 4 {
		t.Errorf("an opponent's draw grew it to power %d", p)
	}
}

// --- the spells -----------------------------------------------------------------

func TestM21B1RiseAgainReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	rock := effectstest.WthGraveyardCard(me, "Old Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Rise Again", "Sorcery", m21B1RiseAgainOracle, effectstest.CardTarget(rock)); err == nil {
		t.Fatal("Rise Again targeted a noncreature card")
	}
	bear := effectstest.WthGraveyardCard(me, "Dead Bear", m21B1Bear)
	effectstest.CastCatalogSpell(t, g, "Rise Again", "Sorcery", m21B1RiseAgainOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.VisBattlefieldNamed(g, me.ID, "Dead Bear"); !ok {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestM21B1FinishingBlowDestroysACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	if err := effectstest.CastCatalogSpellErr(t, g, "Finishing Blow", "Instant", m21B1FinishingBlowOracle, effectstest.CardTarget(land)); err == nil {
		t.Fatal("Finishing Blow targeted a land")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", m21B1Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Finishing Blow", "Instant", m21B1FinishingBlowOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
}

func TestM21B1EliminateTakesOnlyManaValueThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	big := effectstest.B10Creature(g, opp.ID, "Big", "Creature — Giant", "{3}{G}{G}", 5, 5)
	small := effectstest.B10Creature(g, opp.ID, "Small", m21B1Bear, "{1}{G}{G}", 3, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Eliminate", "Instant", m21B1EliminateOracle, effectstest.CardTarget(big)); err == nil {
		t.Fatal("Eliminate targeted a creature with mana value 5")
	}
	effectstest.CastCatalogSpell(t, g, "Eliminate", "Instant", m21B1EliminateOracle, effectstest.CardTarget(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the mana value 3 creature survived")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the mana value 5 creature was destroyed")
	}
}

func TestM21B1ExileAndCompensate(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		victimType, token      string
	}{
		{"Angelic Ascension", "Instant", m21B1AngelicAscensionOracle, m21B1Bear, "Angel"},
		{"Secure the Scene", "Sorcery", m21B1SecureTheSceneOracle, "Artifact", "Soldier"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
			land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
			if err := effectstest.CastCatalogSpellErr(t, g, c.name, c.typeLine, c.oracle, effectstest.CardTarget(land)); err == nil {
				t.Fatalf("%s targeted a land", c.name)
			}
			victim := effectstest.B12Creature(g, opp.ID, "Victim", c.victimType, 2, 2)
			effectstest.CastCatalogSpell(t, g, c.name, c.typeLine, c.oracle, effectstest.CardTarget(victim))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
				t.Error("the permanent was not exiled")
			}
			if n := m21B1ControlledNamed(g, c.token, opp.ID); n != 1 {
				t.Errorf("its controller has %d %s tokens, want 1", n, c.token)
			}
			if n := m21B1ControlledNamed(g, c.token, me.ID); n != 0 {
				t.Errorf("the caster has %d %s tokens, want 0", n, c.token)
			}
		})
	}
}

func TestM21B1AngelicAscensionTokenIsAFourFourFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", m21B1Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Angelic Ascension", "Instant", m21B1AngelicAscensionOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	ids := effectstest.BattlefieldIDsNamed(g, "Angel")
	if len(ids) != 1 {
		t.Fatalf("%d Angels, want 1", len(ids))
	}
	if p, tough := effectstest.PtOf(t, g, ids[0]); p != 4 || tough != 4 {
		t.Errorf("the Angel is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, ids[0], "flying") {
		t.Error("the Angel has no flying")
	}
}

func TestM21B1MiscastCountersAnInstantOrSorceryUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := effectstest.CastCatalogSpell(t, g, "A Bear", m21B1Bear, "", nil)
	if err := effectstest.CastCatalogSpellErr(t, g, "Miscast", "Instant", m21B1MiscastOracle, effectstest.CardTarget(creature)); err == nil {
		t.Fatal("Miscast targeted a creature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	spell := effectstest.CastCatalogSpell(t, g, "A Ponder", "Sorcery", "", nil)
	effectstest.CastCatalogSpell(t, g, "Miscast", "Instant", m21B1MiscastOracle, effectstest.CardTarget(spell))
	effectstest.PassUntilTaxed(t, g, me.ID)
	if c := effectstest.AnswerPayUnless(t, g, me.ID, false); c == nil {
		t.Fatal("no pay-unless prompt")
	}
	if g.Stack.Contains(spell) {
		t.Error("the spell was not countered")
	}
}

func TestM21B1ReadTheTidesEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Read the Tides", "Sorcery", m21B1ReadTheTidesOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	opp := g2.Seats[1]
	a := effectstest.B12Creature(g2, opp.ID, "Bear A", m21B1Bear, 2, 2)
	b := effectstest.B12Creature(g2, opp.ID, "Bear B", m21B1Bear, 2, 2)
	keep := effectstest.B12Creature(g2, opp.ID, "Bear C", m21B1Bear, 2, 2)
	effectstest.CastModal(t, g2, "Read the Tides", "Sorcery", m21B1ReadTheTidesOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g2)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("the two creatures did not return to their owner's hand")
	}
	if !g2.Battlefield.Contains(keep) {
		t.Error("an untargeted creature was returned")
	}
}

func TestM21B1RunAfoulTakesOnlyAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	walker := effectstest.B12Creature(g, opp.ID, "Walker", m21B1Bear, 2, 2)
	flyer := m21B1Flyer(g, opp.ID, "Flyer")
	effectstest.CastCatalogSpell(t, g, "Run Afoul", "Instant", m21B1RunAfoulOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
		for _, id := range c.SacrificeOptions {
			if id == walker {
				t.Fatal("a creature without flying was offered")
			}
		}
		effectstest.AnswerSacrifice(t, g, opp.ID, flyer)
	}
	if g.Battlefield.Contains(flyer) {
		t.Error("the flyer was not sacrificed")
	}
	if !g.Battlefield.Contains(walker) {
		t.Error("the creature without flying was sacrificed")
	}
}

func TestM21B1BadDeal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bad Deal", "Sorcery", m21B1BadDealOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("the caster was asked to discard")
	}
	for round := 0; round < 4; round++ {
		for _, s := range g.Seats {
			if s.ID == me.ID {
				continue
			}
			if c := effectstest.DiscardChoiceFor(g, s.ID); c != nil {
				if c.ChooseMax != 2 {
					t.Errorf("%s owes %d discards, want 2", s.Name, c.ChooseMax)
				}
				effectstest.DiscardFromHand(t, g, s.ID)
			}
		}
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.Life != lives[s.ID]-2 {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, lives[s.ID]-2)
		}
	}
}

func TestM21B1RookieMistakeNeedsTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m21B1Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m21B1Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Rookie Mistake", "Instant", m21B1RookieMistakeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Fatal("one creature was both targets")
	}
	effectstest.CastCatalogSpell(t, g, "Rookie Mistake", "Instant", m21B1RookieMistakeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 2 || tough != 4 {
		t.Errorf("the first target is %d/%d, want 2/4", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 0 || tough != 2 {
		t.Errorf("the second target is %d/%d, want 0/2", p, tough)
	}
}
