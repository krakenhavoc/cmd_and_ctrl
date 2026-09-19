package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch13_test.go — card-level coverage for the card-coverage
// roadmap's batch 13 (#306, `edhrec_rank` 1421–1524): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b13-prefixed.

const (
	b13RuggedHighlandsOracle      = "6c922206-6e68-4dcd-9559-88da1074f2c4"
	b13HedronCrabOracle           = "7216f974-3c84-40ef-b904-82019900c204"
	b13CorpseKnightOracle         = "0c77b805-728d-4ae4-88a9-223f4b7b52e0"
	b13UnderworldDreamsOracle     = "967cf377-ae26-464d-85ac-8448b5a911f7"
	b13ImperiousPerfectOracle     = "3fa71348-fa4d-4f39-a451-cf1570591991"
	b13DrownInTheLochOracle       = "0f264e5b-264e-4e97-9a8d-8ae1d6a286ce"
	b13BoltwaveOracle             = "bb8cccc8-e56f-42fa-a0d7-fc889b5e1c7b"
	b13GoldveinHydraOracle        = "2b62543f-a475-457a-a96b-b5d070383d3c"
	b13InspiringVantageOracle     = "3f17c60e-923a-4392-9da8-87d9ded009b7"
	b13LongRiversPullOracle       = "f1993767-1d07-49c8-b8dc-04ec9840a999"
	b13DarkDealOracle             = "c527eb80-ccac-40b8-8377-c31121613128"
	b13KessigFlamebreatherOracle  = "eadd9559-6dcb-4b96-8c95-58abddd0e120"
	b13BeledrosWitherbloomOracle  = "90194ff1-db61-463f-b5a3-15cd85311d0e"
	b13VibrantCityscapeOracle     = "6a7f3e1f-6798-4644-b64c-7765f81f0938"
	b13WittyRoastmasterOracle     = "34edbf1b-0826-49f1-a04d-15cc8c543b22"
	b13NoMercyOracle              = "b9538d53-480b-481a-abbf-83ab17e1a45b"
	b13FellTheMightyOracle        = "1fbb2a7c-8093-4729-b8fe-cf032d99470f"
	b13ThaliaHereticCatharOracle  = "13be47e4-9da2-4f9e-baf0-db96f7777ccb"
	b13ConcordantCrossroadsOracle = "ff01b408-6d17-40a3-9efd-a1b341ec1307"
	b13WorldShaperOracle          = "3c075bb6-1831-4521-bd8d-4ed2825ae796"
	b13LaezelOracle               = "c066f921-e349-45d1-8ec3-0955d10bbf19"
	b13FarhavenElfOracle          = "4ce2357f-93e6-40ca-beca-8f4e15adc464"
	b13SunbakedCanyonOracle       = "f97fd068-b83a-4621-bf8c-cc96e880ce90"
	b13HeliodsInterventionOracle  = "e7564d66-767c-4cd9-a5f0-0f2488a4a74b"
	b13ProsperousInnkeeperOracle  = "da785227-cf8a-4d44-9e7c-fc909ea868f2"
	b13TwitchingDollOracle        = "fd6e1967-237a-41f6-bbf4-2c869f9447c8"
	b13CopperlineGorgeOracle      = "a05f641c-15c9-43dc-ae0d-1ea372fd33d5"
	b13DourPortMageOracle         = "cf58e309-00e8-438e-813e-2e1c1002db23"
	b13CrystalVeinOracle          = "616d6013-24f4-4999-9bf3-5b0764e52fa6"
	b13GratuitousViolenceOracle   = "7c340a39-4ee0-4ba1-bb66-6674f8020fda"
	b13WaveGoodbyeOracle          = "06ef46c6-00ba-40e6-b866-d0095ab83749"
	b13DranaOracle                = "89b24b5f-d837-4274-8877-8ff7dc2708ba"
	b13RadiantGroveOracle         = "32c91719-f3dd-4cc7-9e32-7d5ccf18f07c"
	b13InGarruksWakeAlreadyOID    = "a6899b94-427d-4851-a474-4087e0a0918a"
)

// b13OpponentCasts puts a card of the given type line in an
// opponent's hand and has them cast it during the active seat's
// precombat main, returning the spell's ID while it is on the stack.
func b13OpponentCasts(t *testing.T, g *game.Game, opp *game.Player, name, typeLine, oracle, manaCost string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("opponent CastSpell %s: %v", name, err)
	}
	return id
}

// b13PlayAs walks the cursor to the given seat's precombat main and
// has that seat play a card from hand — a land play or a cast, the
// engine decides by type.
func b13PlayAs(t *testing.T, g *game.Game, seat int, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	aangAdvanceToMain(t, g, seat)
	p := g.Seats[seat]
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s as seat %d: %v", name, seat, err)
	}
	return id
}

// b13Tapped reads a battlefield card's tapped state.
func b13Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return b12Card(t, g, id).Tapped
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Four are rows
// in cycle tables (Rugged Highlands in gain_lands.go, Inspiring
// Vantage and Copperline Gorge in fastlands.go, Radiant Grove in
// dominaria_united_duals.go), so a transposed row is invisible until
// someone plays that exact card. One of the issue's 43 was already
// on main — In Garruk's Wake (#382) — pinned here so the table
// matches the issue.
func TestBatch13CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b13RuggedHighlandsOracle:      "Rugged Highlands",
		b13HedronCrabOracle:           "Hedron Crab",
		b13CorpseKnightOracle:         "Corpse Knight",
		b13UnderworldDreamsOracle:     "Underworld Dreams",
		b13ImperiousPerfectOracle:     "Imperious Perfect",
		b13DrownInTheLochOracle:       "Drown in the Loch",
		b13BoltwaveOracle:             "Boltwave",
		b13GoldveinHydraOracle:        "Goldvein Hydra",
		b13InspiringVantageOracle:     "Inspiring Vantage",
		b13LongRiversPullOracle:       "Long River's Pull",
		b13DarkDealOracle:             "Dark Deal",
		b13KessigFlamebreatherOracle:  "Kessig Flamebreather",
		b13BeledrosWitherbloomOracle:  "Beledros Witherbloom",
		b13VibrantCityscapeOracle:     "Vibrant Cityscape",
		b13WittyRoastmasterOracle:     "Witty Roastmaster",
		b13NoMercyOracle:              "No Mercy",
		b13FellTheMightyOracle:        "Fell the Mighty",
		b13ThaliaHereticCatharOracle:  "Thalia, Heretic Cathar",
		b13ConcordantCrossroadsOracle: "Concordant Crossroads",
		b13WorldShaperOracle:          "World Shaper",
		b13LaezelOracle:               "Lae'zel, Vlaakith's Champion",
		b13FarhavenElfOracle:          "Farhaven Elf",
		b13SunbakedCanyonOracle:       "Sunbaked Canyon",
		b13HeliodsInterventionOracle:  "Heliod's Intervention",
		b13ProsperousInnkeeperOracle:  "Prosperous Innkeeper",
		b13TwitchingDollOracle:        "Twitching Doll",
		b13CopperlineGorgeOracle:      "Copperline Gorge",
		b13DourPortMageOracle:         "Dour Port-Mage",
		b13CrystalVeinOracle:          "Crystal Vein",
		b13GratuitousViolenceOracle:   "Gratuitous Violence",
		b13WaveGoodbyeOracle:          "Wave Goodbye",
		b13DranaOracle:                "Drana, Liberator of Malakir",
		b13RadiantGroveOracle:         "Radiant Grove",
		b13InGarruksWakeAlreadyOID:    "In Garruk's Wake",
	}
	if len(want) != 34 {
		t.Fatalf("the batch registers 34 cards, the table lists %d", len(want))
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

// --- the four cycle rows -------------------------------------------

func TestB13LandRowsProduceTheirPrintedColours(t *testing.T) {
	for _, tc := range []struct {
		oracle, name, produced string
		triggers               int
	}{
		{b13RuggedHighlandsOracle, "Rugged Highlands", "{R|G}", 1},
		{b13InspiringVantageOracle, "Inspiring Vantage", "{R|W}", 0},
		{b13CopperlineGorgeOracle, "Copperline Gorge", "{R|G}", 0},
		{b13RadiantGroveOracle, "Radiant Grove", "{G|W}", 0},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok {
			t.Fatalf("%s not registered", tc.name)
		}
		if len(spec.Replacements) != 1 {
			t.Errorf("%s: %d replacements, want 1 (enters tapped …)", tc.name, len(spec.Replacements))
		}
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
		if len(spec.Triggered) != tc.triggers {
			t.Errorf("%s: %d triggers, want %d", tc.name, len(spec.Triggered), tc.triggers)
		}
	}
}

func TestB13RuggedHighlandsEntersTappedAndGainsALife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	id := b12PlayFromHand(t, g, "Rugged Highlands", "Land", b13RuggedHighlandsOracle, game.CastSpellParams{})
	if !b13Tapped(t, g, id) {
		t.Fatal("Rugged Highlands enters tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
	if triggerOnStack(g, id) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the life is a trigger with a response window")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
}

func TestB13FastlandRowsEnterUntappedEarly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	vantage := b12PlayFromHand(t, g, "Inspiring Vantage", "Land", b13InspiringVantageOracle, game.CastSpellParams{})
	if b13Tapped(t, g, vantage) {
		t.Error("with two other lands Inspiring Vantage enters untapped")
	}
	gorge := b12PlayFromHand(t, g, "Copperline Gorge", "Land", b13CopperlineGorgeOracle, game.CastSpellParams{})
	if !b13Tapped(t, g, gorge) {
		t.Error("with three other lands Copperline Gorge enters tapped")
	}
}

func TestB13RadiantGroveEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := b12PlayFromHand(t, g, "Radiant Grove", "Land — Forest Plains", b13RadiantGroveOracle, game.CastSpellParams{})
	if !b13Tapped(t, g, id) {
		t.Fatal("Radiant Grove enters tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
}

// --- Hedron Crab ---------------------------------------------------

func TestB13HedronCrabMillsATargetPlayerOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Hedron Crab", "Creature — Crab", b13HedronCrabOracle, 0, 2)
	before, lib := opp.Graveyard.Size(), opp.Library.Size()

	b12PlayFromHand(t, g, "Island", "Basic Land — Island", "", game.CastSpellParams{})
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != before+3 || opp.Library.Size() != lib-3 {
		t.Errorf("opponent graveyard %d → %d, library %d → %d; want three milled", before, opp.Graveyard.Size(), lib, opp.Library.Size())
	}
	// An opponent's land is not yours.
	b13PlayAs(t, g, 1, "Their Island", "Basic Land — Island", "")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's land must not trigger the Crab")
	}
	if opp.Graveyard.Size() != before+3 {
		t.Errorf("opponent graveyard %d, want still %d", opp.Graveyard.Size(), before+3)
	}
}

// --- Corpse Knight / Witty Roastmaster / Prosperous Innkeeper --------

func TestB13CorpseKnightDrainsForEachOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := make([]int, 0, 3)
	for _, p := range g.Seats[1:] {
		before = append(before, p.Life)
	}
	castCatalogSpell(t, g, "Corpse Knight", "Creature — Zombie Knight", b13CorpseKnightOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i] {
			t.Fatalf("the Knight's own entry must not trigger it: opponent %d %d → %d", i+1, before[i], p.Life)
		}
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 2) })
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-3 {
			t.Errorf("a Bear and two Goblins: opponent %d %d → %d, want -3", i+1, before[i], p.Life)
		}
	}
	if me.Life != 40 {
		t.Errorf("the controller loses nothing: %d", me.Life)
	}
}

func TestB13WittyRoastmasterPingsForEachOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	b12Push(g, me.ID, "Witty Roastmaster", "Creature — Devil Citizen", b13WittyRoastmasterOracle, 3, 2)
	before := opp.Life
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent %d → %d, want -1", before, opp.Life)
	}
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("a noncreature permanent must not trigger it: %d → %d", before, opp.Life)
	}
}

func TestB13ProsperousInnkeeperMakesATreasureAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Prosperous Innkeeper", "Creature — Halfling Citizen", b13ProsperousInnkeeperOracle, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("the ETB makes one Treasure, got %d", n)
	}
	before := me.Life
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("another creature entering: life %d → %d, want +1", before, me.Life)
	}
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("only the Innkeeper's own entry makes a Treasure, got %d", n)
	}
}

// --- Underworld Dreams ---------------------------------------------

func TestB13UnderworldDreamsPingsAnOpponentPerCardDrawn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Underworld Dreams", "Enchantment", b13UnderworldDreamsOracle, 0, 0)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("two draws: opponent %d → %d, want -2", before, opp.Life)
	}
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("the controller's own draw must not trigger it: %d → %d", mine, me.Life)
	}
}

// --- Imperious Perfect ---------------------------------------------

func TestB13ImperiousPerfectPumpsOtherElvesAndMakesThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	perfect := b12Push(g, me.ID, "Imperious Perfect", "Creature — Elf Warrior", b13ImperiousPerfectOracle, 2, 2)
	elf := b12Creature(g, me.ID, "Llanowar", "Creature — Elf Druid", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirElf := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	if p := effectivePower(t, g, elf); p != 2 {
		t.Errorf("another Elf you control gets +1/+1: power %d, want 2", p)
	}
	if p := effectivePower(t, g, perfect); p != 2 {
		t.Errorf("the Perfect does not pump itself: power %d, want 2", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Elf is untouched: power %d, want 2", p)
	}
	if p := effectivePower(t, g, theirElf); p != 1 {
		t.Errorf("an opponent's Elf is untouched: power %d, want 1", p)
	}
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateCatalogAbility(me.ID, perfect, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	tok := findBattlefieldByName(g, "Elf Warrior")
	if tok == uuid.Nil {
		t.Fatal("the ability makes an Elf Warrior token")
	}
	if p := effectivePower(t, g, tok); p != 2 {
		t.Errorf("the token is an Elf and gets +1/+1: power %d, want 2", p)
	}
}

// --- Drown in the Loch ---------------------------------------------

func TestB13DrownInTheLochScalesWithTheControllersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	// Empty graveyard: a one-drop is not a legal target.
	if err := g.CastSpell(me.ID, b13HandCard(me, "Drown in the Loch", "Instant", b13DrownInTheLochOracle),
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err == nil {
		t.Fatal("with an empty graveyard a mana value 1 spell is not a legal target")
	}
	pushGraveyardCardForTest(opp, "Dead")
	if err := g.CastSpell(me.ID, b13HandCard(me, "Drown in the Loch", "Instant", b13DrownInTheLochOracle),
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("with one card in their graveyard the Bolt is a legal target: %v", err)
	}
	life := me.Life
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Error("the Bolt should be countered")
	}

	// The creature mode, against a three-drop.
	theirs := b10Creature(g, opp.ID, "Their Three", "Creature — Beast", "{2}{G}", 3, 3)
	if err := g.CastSpell(me.ID, b13HandCard(me, "Drown in the Loch", "Instant", b13DrownInTheLochOracle),
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("two cards in their graveyard cannot reach a mana value 3 creature")
	}
	pushGraveyardCardForTest(opp, "Dead Two")
	if err := g.CastSpell(me.ID, b13HandCard(me, "Drown in the Loch", "Instant", b13DrownInTheLochOracle),
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err != nil {
		t.Fatalf("three cards in their graveyard reach it: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature should be destroyed")
	}
}

// b13HandCard seeds a card into a player's hand and returns its ID.
func b13HandCard(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// --- Boltwave / Kessig Flamebreather -------------------------------

func TestB13BoltwaveHitsEachOpponentForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := make([]int, 0, 3)
	for _, p := range g.Seats[1:] {
		before = append(before, p.Life)
	}
	castCatalogSpell(t, g, "Boltwave", "Sorcery", b13BoltwaveOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i]-3 {
			t.Errorf("opponent %d: %d → %d, want -3", i+1, before[i], p.Life)
		}
	}
	if me.Life != 40 {
		t.Errorf("the caster is untouched: %d", me.Life)
	}
}

func TestB13KessigFlamebreatherPingsOnNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kessig Flamebreather", "Creature — Human Shaman", b13KessigFlamebreatherOracle, 1, 3)
	before := opp.Life
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("an artifact spell: opponent %d → %d, want -1", before, opp.Life)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("a creature spell must not trigger it: %d → %d", before, opp.Life)
	}
}

// --- Goldvein Hydra ------------------------------------------------

func TestB13GoldveinHydraEntersWithXAndPaysOutTappedTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hydra := b12PlayFromHand(t, g, "Goldvein Hydra", "Creature — Hydra", b13GoldveinHydraOracle, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("X=3: the Hydra survives its entry")
	}
	if got := b12Counter(t, g, hydra, "+1/+1"); got != 3 {
		t.Fatalf("X=3: %d counters, want 3", got)
	}
	if p := b12Card(t, g, hydra).CurrentPower(); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	for _, kw := range []string{"vigilance", "trample", "haste"} {
		if !containsString(effectiveAbilities(t, g, hydra), kw) {
			t.Errorf("printed %s did not reach the effective abilities", kw)
		}
	}
	// One more counter from elsewhere, then it dies: four Treasures,
	// all tapped.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, "+1/+1", 1) })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hydra) })
	passPriorityAroundTable(t, g)
	treasures := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			treasures++
			if !c.Tapped {
				t.Error("the Treasures enter tapped")
			}
		}
	}
	if treasures != 4 {
		t.Errorf("a 4-power Hydra dying makes %d Treasures, want 4", treasures)
	}
}

// X=0 is a legal announcement (CR 601.2b) and the printed outcome is
// that nothing survives it: the Hydra enters with no counters as the
// 0/0 it prints, the toughness check puts it into its owner's
// graveyard (CR 704.5f), and a 0-power Hydra dying makes no
// Treasures. Cast through a printing, because that is what tells the
// engine the 0 is printed rather than the importer's stand-in —
// printed_zero_body_test.go has the rest of the family and #691 the
// rule.
func TestB13GoldveinHydraWithXZeroIsAZeroThatPaysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hydra := castPrinted(t, g, "Goldvein Hydra", "Creature — Hydra", b13GoldveinHydraOracle, game.CastSpellParams{XValue: 0})
	if g.Battlefield.Contains(hydra) {
		t.Fatal("a Goldvein Hydra cast for X=0 stayed on the battlefield — CR 704.5f")
	}
	if !me.Graveyard.Contains(hydra) {
		t.Error("it did not reach its owner's graveyard")
	}
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("power 0 makes no Treasures, got %d", n)
	}
}

// --- Long River's Pull ---------------------------------------------

func TestB13LongRiversPullCountersACreatureSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err := b09TryCast(t, g, "Long River's Pull", "Instant", b13LongRiversPullOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err == nil {
		t.Fatal("without the gift a noncreature spell is not a legal target")
	}
	passPriorityAroundTable(t, g)
	// A creature spell is cast on its controller's own turn; the
	// Pull answers it from across the table.
	bear := b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, b13HandCard(me, "Long River's Pull", "Instant", b13LongRiversPullOracle),
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("Long River's Pull at a creature spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell should be countered")
	}
}

// --- Dark Deal -----------------------------------------------------

func TestB13DarkDealWheelsEveryoneForOneFewer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	// Seven in hand each after the opening draw; the caster's Deal
	// leaves the hand with six, so they redraw five.
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	castCatalogSpell(t, g, "Dark Deal", "Sorcery", b13DarkDealOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != mine-1 {
		t.Errorf("caster: %d in hand → %d, want %d", mine, me.Hand.Size(), mine-1)
	}
	if opp.Hand.Size() != theirs-1 {
		t.Errorf("opponent: %d in hand → %d, want %d", theirs, opp.Hand.Size(), theirs-1)
	}
	if me.Graveyard.Size() < mine {
		t.Errorf("the discards reach the graveyard: %d cards", me.Graveyard.Size())
	}
}

// --- Beledros Witherbloom ------------------------------------------

func TestB13BeledrosMakesAPestOnEveryUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beledros := b12Push(g, me.ID, "Beledros Witherbloom", "Legendary Creature — Elder Dragon", b13BeledrosWitherbloomOracle, 4, 4)
	if !containsString(effectiveAbilities(t, g, beledros), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	// The next upkeep is an opponent's — still a Pest for Beledros's
	// controller.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Pest"); n != 1 {
		t.Fatalf("an opponent's upkeep: %d Pests, want 1", n)
	}
	// Round the table: seats 2 and 3 each have an upkeep, then yours
	// — four Pests a round at a four-player table.
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Pest"); n != 4 {
		t.Errorf("after every seat's upkeep: %d Pests, want 4", n)
	}
	pest := b12Card(t, g, findBattlefieldByName(g, "Pest"))
	if !pest.HasColor("B") || !pest.HasColor("G") || !pest.HasSubtype("Pest") {
		t.Errorf("the Pest is a black and green Pest, got %v %q", pest.Colors, pest.TypeLine)
	}
}

// --- Vibrant Cityscape ---------------------------------------------

func TestB13VibrantCityscapeFetchesABasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cityscape := b12Push(g, me.ID, "Vibrant Cityscape", "Land", b13VibrantCityscapeOracle, 0, 0)
	dual := stapleLibraryCard(me, "Watery Grave", "Land — Island Swamp")
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, cityscape, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(cityscape) {
		t.Fatal("the Cityscape is sacrificed at announce")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the basic Forest should be fetched")
	}
	if !b13Tapped(t, g, forest) {
		t.Error("the fetched land enters tapped")
	}
	if g.Battlefield.Contains(dual) {
		t.Error("a nonbasic is not a legal find")
	}
}

// --- No Mercy ------------------------------------------------------

func TestB13NoMercyDestroysTheCreatureThatDamagedYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "No Mercy", "Enchantment", b13NoMercyOracle, 0, 0)
	attacker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)

	aangAdvanceToMain(t, g, 1)
	attackWith(t, g, me.ID, attacker)
	if triggerOnStack(g, findBattlefieldByName(g, "No Mercy")) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("combat damage to the controller should trigger No Mercy")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the creature that dealt the damage should be destroyed")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature that dealt no damage is untouched")
	}
	// Non-combat damage from a creature counts too; damage from a
	// spell does not.
	ping := b12Creature(g, opp.ID, "Pinger", "Creature — Wizard", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(ping, me.ID, 1) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ping) {
		t.Error("a creature pinging you is destroyed too")
	}
	before := len(g.Events)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventTrigger && ev.Label != "" {
			t.Errorf("damage from no creature must not trigger it: %+v", ev)
		}
	}
}

// --- Fell the Mighty -----------------------------------------------

func TestB13FellTheMightyDestroysEverythingBiggerThanTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Creature(g, me.ID, "Small", "Creature — Human", 1, 1)
	same := b12Creature(g, opp.ID, "Same", "Creature — Bear", 2, 2)
	target := b12Creature(g, me.ID, "Target", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Beast", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Beast", 4, 4)

	castCatalogSpell(t, g, "Fell the Mighty", "Sorcery", b13FellTheMightyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{small, same, target} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a creature with power at most the target's survives")
		}
	}
	for _, id := range []uuid.UUID{mine, theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a creature with greater power is destroyed, yours included")
		}
	}
}

// --- Thalia, Heretic Cathar ----------------------------------------

func TestB13ThaliaTapsOpponentsCreaturesAndNonbasicLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thalia := b12Push(g, me.ID, "Thalia, Heretic Cathar", "Legendary Creature — Human Soldier", b13ThaliaHereticCatharOracle, 3, 2)
	if !containsString(effectiveAbilities(t, g, thalia), "first strike") {
		t.Error("printed first strike did not reach the effective abilities")
	}
	// The controller's own permanents are untouched.
	mine := b12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if b13Tapped(t, g, mine) {
		t.Error("Thalia leaves her controller's creatures alone")
	}
	// An opponent's basic enters untapped, a nonbasic tapped.
	basic := b13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")
	if b13Tapped(t, g, basic) {
		t.Error("an opponent's basic land enters untapped")
	}
	nonbasic := b13PlayAs(t, g, 2, "Ancient Tomb", "Land", "")
	if !b13Tapped(t, g, nonbasic) {
		t.Error("an opponent's nonbasic land enters tapped")
	}
	theirs := b13PlayAs(t, g, 3, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, theirs) {
		t.Error("an opponent's creature enters tapped")
	}
	// An opponent's noncreature, nonland permanent is untouched.
	rock := b13PlayAs(t, g, 1, "Rock", "Artifact", "")
	passPriorityAroundTable(t, g)
	if b13Tapped(t, g, rock) {
		t.Error("an opponent's artifact is not a creature or a nonbasic land")
	}
}

// --- Concordant Crossroads -----------------------------------------

func TestB13ConcordantCrossroadsGivesEveryCreatureHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Concordant Crossroads", "World Enchantment", b13ConcordantCrossroadsOracle, 0, 0)
	mine := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !containsString(effectiveAbilities(t, g, id), "haste") {
			t.Errorf("every creature has haste, %s does not", b12Card(t, g, id).Name)
		}
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Errorf("a creature cast this turn attacks under the Crossroads: %v", err)
	}
}

// --- World Shaper --------------------------------------------------

func TestB13WorldShaperMillsOnAttackAndReturnsLandsOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaper := b12Push(g, me.ID, "World Shaper", "Creature — Merfolk Shaman", b13WorldShaperOracle, 3, 3)
	forest := b10LibraryTop(me, "Forest", "Basic Land — Forest", "", 0, 0)
	b10LibraryTop(me, "Spell", "Sorcery", "{G}", 0, 0)
	b10LibraryTop(me, "Island", "Basic Land — Island", "", 0, 0)
	before := me.Graveyard.Size()

	declareAttack(t, g, opp.ID, shaper)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != before+3 {
		t.Fatalf("graveyard %d → %d, want three milled", before, me.Graveyard.Size())
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shaper) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Error("the milled Forest should return to the battlefield")
	}
	if !b13Tapped(t, g, forest) {
		t.Error("the returned land is tapped")
	}
	if findBattlefieldByName(g, "Spell") != uuid.Nil {
		t.Error("only land cards return")
	}
	if n := countBattlefieldNamed(g, me.ID, "Island"); n != 1 {
		t.Errorf("both milled lands return: %d Islands", n)
	}
}

// --- Lae'zel, Vlaakith's Champion ----------------------------------

func TestB13LaezelAddsOneCounterOfEachKindYouPut(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Lae'zel, Vlaakith's Champion", "Legendary Creature — Gith Warrior", b13LaezelOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	walker := b12Permanent(g, me.ID, "Walker", "Legendary Planeswalker — Test")
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(bear, "+1/+1", 2)
		_ = g.AddCounterForEffect(walker, "loyalty", 1)
		_ = g.AddCounterForEffect(rock, "charge", 1)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 1)
	})
	if got := b12Counter(t, g, bear, "+1/+1"); got != 3 {
		t.Errorf("two +1/+1 counters become three: got %d", got)
	}
	if got := b12Counter(t, g, walker, "loyalty"); got != 2 {
		t.Errorf("one loyalty counter on your planeswalker becomes two: got %d", got)
	}
	if got := b12Counter(t, g, rock, "charge"); got != 1 {
		t.Errorf("an artifact is neither a creature nor a planeswalker: got %d", got)
	}
	if got := b12Counter(t, g, theirs, "+1/+1"); got != 1 {
		t.Errorf("an opponent's creature is untouched: got %d", got)
	}
	// A removal is not a placement, and an opponent's resolution is
	// not you putting.
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, "+1/+1", -1)
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: opp.ID})
		_ = g.AddCounterForEffect(bear, "+1/+1", 1)
	})
	if got := b12Counter(t, g, bear, "+1/+1"); got != 3 {
		t.Errorf("3 -1 +1 = 3 with neither the removal nor an opponent's placement increased: got %d", got)
	}
}

// --- Farhaven Elf --------------------------------------------------

func TestB13FarhavenElfMayFetchABasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Farhaven Elf", "Creature — Elf Druid", b13FarhavenElfOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the Forest should be fetched")
	}
	if !b13Tapped(t, g, forest) {
		t.Error("the fetched land enters tapped")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	forest2 := stapleLibraryCard(me2, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g2, "Farhaven Elf", "Creature — Elf Druid", b13FarhavenElfOracle, nil)
	passPriorityAroundTable(t, g2)
	answerLatestTriggerPrompt(t, g2, me2.ID, false)
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(forest2) {
		t.Error("declining the prompt fetches nothing")
	}
}

// --- Sunbaked Canyon / Crystal Vein --------------------------------

func TestB13SunbakedCanyonPaysLifeForManaAndCashesInForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	canyon := b12Push(g, me.ID, "Sunbaked Canyon", "Land", b13SunbakedCanyonOracle, 0, 0)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, canyon, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	if n := b10ResolveAllManaPicks(t, g, me.ID, "W"); n != 1 {
		t.Fatalf("expected one colour pick, answered %d", n)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
	b08Untap(g, canyon)
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, canyon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	if g.Battlefield.Contains(canyon) {
		t.Error("the Canyon was sacrificed")
	}
}

func TestB13CrystalVeinTapsForOneOrCracksForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vein := b12Push(g, me.ID, "Crystal Vein", "Land", b13CrystalVeinOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, vein, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for one: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	if !g.Battlefield.Contains(vein) {
		t.Fatal("the first ability does not sacrifice")
	}
	b08Untap(g, vein)
	if err := g.ActivateManaAbility(me.ID, vein, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack for two: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 3 {
		t.Errorf("pool %v, want three colourless", got)
	}
	if g.Battlefield.Contains(vein) {
		t.Error("the second ability sacrifices the Vein")
	}
}

// --- Heliod's Intervention -----------------------------------------

func TestB13HeliodsInterventionDestroysXTargetsOrGainsTwiceX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	aura := b12Permanent(g, opp.ID, "Aura", "Enchantment")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	two := []game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: aura}}
	if err := g.CastSpell(me.ID, b13HandCard(me, "Heliod's Intervention", "Instant", b13HeliodsInterventionOracle),
		game.CastSpellParams{XValue: 1, Modes: []int{0}, Targets: two}); err == nil {
		t.Fatal("X=1 with two targets is not a legal announcement")
	}
	if err := g.CastSpell(me.ID, b13HandCard(me, "Heliod's Intervention", "Instant", b13HeliodsInterventionOracle),
		game.CastSpellParams{XValue: 1, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	if err := g.CastSpell(me.ID, b13HandCard(me, "Heliod's Intervention", "Instant", b13HeliodsInterventionOracle),
		game.CastSpellParams{XValue: 2, Modes: []int{0}, Targets: two}); err != nil {
		t.Fatalf("X=2 with two targets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(aura) {
		t.Error("both targets should be destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was never a target")
	}

	before := me.Life
	if err := g.CastSpell(me.ID, b13HandCard(me, "Heliod's Intervention", "Instant", b13HeliodsInterventionOracle),
		game.CastSpellParams{XValue: 3, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("the lifegain mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+6 {
		t.Errorf("X=3: life %d → %d, want +6", before, me.Life)
	}
}

// --- Twitching Doll ------------------------------------------------

func TestB13TwitchingDollBanksNestCountersIntoSpiders(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	doll := b12Push(g, me.ID, "Twitching Doll", "Artifact Creature — Spider Toy", b13TwitchingDollOracle, 2, 2)
	for i := 1; i <= 2; i++ {
		b08Untap(g, doll)
		if err := g.ActivateManaAbility(me.ID, doll, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility %d: %v", i, err)
		}
		if n := b10ResolveAllManaPicks(t, g, me.ID, "U"); n != 1 {
			t.Fatalf("expected one colour pick, answered %d", n)
		}
		if got := b12Counter(t, g, doll, "nest"); got != i {
			t.Fatalf("after %d taps: %d nest counters", i, got)
		}
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("pool %v, want [U U]", got)
	}
	// One counter of another kind counts too.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(doll, "+1/+1", 1) })
	b08Untap(g, doll)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, doll, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(doll) {
		t.Fatal("the Doll is sacrificed at announce")
	}
	passPriorityAroundTable(t, g)
	spiders := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spider" && c.Controller == me.ID {
			spiders++
			if p := effectivePower(t, g, c.InstanceID); p != 2 {
				t.Errorf("Spider power %d, want 2", p)
			}
			if !containsString(effectiveAbilities(t, g, c.InstanceID), "reach") {
				t.Error("the Spiders have reach")
			}
		}
	}
	if spiders != 3 {
		t.Errorf("three counters make %d Spiders, want 3", spiders)
	}
}

func TestB13TwitchingDollSacrificeIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	doll := b12Push(g, me.ID, "Twitching Doll", "Artifact Creature — Spider Toy", b13TwitchingDollOracle, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, doll, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the sacrifice ability is sorcery-speed only")
	}
}

// --- Dour Port-Mage ------------------------------------------------

func TestB13DourPortMageDrawsWhenACreatureLeavesWithoutDying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := b12Push(g, me.ID, "Dour Port-Mage", "Creature — Frog Wizard", b13DourPortMageOracle, 1, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	third := b12Creature(g, me.ID, "Third", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	// The Port-Mage cannot return itself.
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mage}},
	}); err == nil {
		t.Fatal("\"another target creature\" excludes the Port-Mage")
	}
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Fatal("an opponent's creature is not \"you control\"")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, bear) != game.ZoneHand {
		t.Fatal("the Bear should be in hand")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("the bounced Bear plus one drawn: hand %d → %d, want +2", hand, me.Hand.Size())
	}
	// Dying is not leaving without dying.
	hand = me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a death must not draw: hand %d → %d", hand, me.Hand.Size())
	}
	// Two leaving at once is one draw.
	fourth := b12Creature(g, me.ID, "Fourth", "Creature — Bear", 2, 2)
	hand = me.Hand.Size()
	g.WithWriteLock(func() { _ = g.BounceCardsToHandForEffect([]uuid.UUID{third, fourth}) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("two bounced plus ONE drawn: hand %d → %d, want +3", hand, me.Hand.Size())
	}
	// An opponent's creature leaving is not yours.
	hand = me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("an opponent's creature leaving must not draw: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Gratuitous Violence -------------------------------------------

func TestB13GratuitousViolenceDoublesYourCreaturesDamageOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Gratuitous Violence", "Enchantment", b13GratuitousViolenceOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("a 2-power attacker under the Violence: %d → %d, want -4", before, opp.Life)
	}
	before = opp.Life
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-3 {
		t.Errorf("a Bolt is not a creature: %d → %d, want -3", before, opp.Life)
	}
	// An opponent's creature is not yours.
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 2) })
	if me.Life != mine-2 {
		t.Errorf("an opponent's creature's damage is not doubled: %d → %d, want -2", mine, me.Life)
	}
}

// --- Wave Goodbye --------------------------------------------------

func TestB13WaveGoodbyeReturnsEveryCreatureWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	countered := b12Creature(g, me.ID, "Countered", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(countered, "+1/+1", 1) })
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	charged := b12Creature(g, opp.ID, "Charged", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(charged, "-1/-1", 1) })
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")

	castCatalogSpell(t, g, "Wave Goodbye", "Sorcery", b13WaveGoodbyeOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(countered) {
		t.Error("a creature with a +1/+1 counter stays")
	}
	for _, id := range []uuid.UUID{mine, theirs, charged} {
		if b12ZoneOf(g, id) != game.ZoneHand {
			t.Errorf("%s should be in its owner's hand", b12Card(t, g, id).Name)
		}
	}
	if findBattlefieldByName(g, "Goblin") != uuid.Nil {
		t.Error("a bounced token ceases to exist")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature is untouched")
	}
}

// --- Drana, Liberator of Malakir -----------------------------------

// Drana has FIRST STRIKE, so this combat has two damage steps
// (CR 510.4) and her trigger belongs to the first one. Since #914 the
// cursor cannot leave a step owing what is on its stack (CR 117.4),
// so the counters land in the first-strike step — which is the whole
// point of the card, and is why the Bear that follows her in hits for
// 3 rather than 2: 2 + 3 = 5. Before #914 the skip-ahead button
// walked past her trigger and the Bear dealt 2.
func TestB13DranaGrowsEveryAttackerWhenSheConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drana := b12Push(g, me.ID, "Drana, Liberator of Malakir", "Legendary Creature — Vampire Ally", b13DranaOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	homebody := b12Creature(g, me.ID, "Homebody", "Creature — Bear", 2, 2)
	for _, kw := range []string{"flying", "first strike"} {
		if !containsString(effectiveAbilities(t, g, drana), kw) {
			t.Errorf("printed %s did not reach the effective abilities", kw)
		}
	}
	before := opp.Life
	attackWith(t, g, opp.ID, drana, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != before-5 {
		t.Errorf("opponent %d → %d, want -5 (her 2 first-strike, then a Bear grown to 3)", before, opp.Life)
	}
	for _, id := range []uuid.UUID{drana, bear} {
		if got := b12Counter(t, g, id, "+1/+1"); got != 1 {
			t.Errorf("%s: %d counters, want 1", b12Card(t, g, id).Name, got)
		}
	}
	if got := b12Counter(t, g, homebody, "+1/+1"); got != 0 {
		t.Errorf("a creature that stayed home got %d counters", got)
	}
}

// --- the log-read helpers ------------------------------------------

func TestB13LastKnownCountersReadTheLastTotalBeforeLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, "+1/+1", 3)
		_ = g.AddCounterForEffect(bear, "+1/+1", -1)
		_ = g.AddCounterForEffect(bear, "nest", 2)
		_ = g.DestroyPermanentForEffect(bear)
	})
	g.WithWriteLock(func() {
		if got := b13LastKnownCounters(g, bear, "+1/+1"); got != 2 {
			t.Errorf("+1/+1: %d, want 2", got)
		}
		if got := b13LastKnownCounters(g, bear, "-1/-1"); got != 0 {
			t.Errorf("-1/-1 was never placed: %d", got)
		}
		if got := b13LastKnownCounterTotal(g, bear); got != 4 {
			t.Errorf("total: %d, want 4", got)
		}
	})
	// A second life of the same instance starts from zero.
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardForEffect(bear, game.ZoneBattlefield)
		_ = g.DestroyPermanentForEffect(bear)
		if got := b13LastKnownCounters(g, bear, "+1/+1"); got != 0 {
			t.Errorf("an earlier life's counters are not this one's: %d", got)
		}
	})
}
