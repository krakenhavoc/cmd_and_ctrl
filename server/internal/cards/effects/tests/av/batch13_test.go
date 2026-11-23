package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
		effectstest.B13RuggedHighlandsOracle:      "Rugged Highlands",
		effectstest.B13HedronCrabOracle:           "Hedron Crab",
		effectstest.B13CorpseKnightOracle:         "Corpse Knight",
		effectstest.B13UnderworldDreamsOracle:     "Underworld Dreams",
		effectstest.B13ImperiousPerfectOracle:     "Imperious Perfect",
		effectstest.B13DrownInTheLochOracle:       "Drown in the Loch",
		effectstest.B13BoltwaveOracle:             "Boltwave",
		effectstest.B13GoldveinHydraOracle:        "Goldvein Hydra",
		effectstest.B13InspiringVantageOracle:     "Inspiring Vantage",
		effectstest.B13LongRiversPullOracle:       "Long River's Pull",
		effectstest.B13DarkDealOracle:             "Dark Deal",
		effectstest.B13KessigFlamebreatherOracle:  "Kessig Flamebreather",
		effectstest.B13BeledrosWitherbloomOracle:  "Beledros Witherbloom",
		effectstest.B13VibrantCityscapeOracle:     "Vibrant Cityscape",
		effectstest.B13WittyRoastmasterOracle:     "Witty Roastmaster",
		effectstest.B13NoMercyOracle:              "No Mercy",
		effectstest.B13FellTheMightyOracle:        "Fell the Mighty",
		effectstest.B13ThaliaHereticCatharOracle:  "Thalia, Heretic Cathar",
		effectstest.B13ConcordantCrossroadsOracle: "Concordant Crossroads",
		effectstest.B13WorldShaperOracle:          "World Shaper",
		effectstest.B13LaezelOracle:               "Lae'zel, Vlaakith's Champion",
		effectstest.B13FarhavenElfOracle:          "Farhaven Elf",
		effectstest.B13SunbakedCanyonOracle:       "Sunbaked Canyon",
		effectstest.B13HeliodsInterventionOracle:  "Heliod's Intervention",
		effectstest.B13ProsperousInnkeeperOracle:  "Prosperous Innkeeper",
		effectstest.B13TwitchingDollOracle:        "Twitching Doll",
		effectstest.B13CopperlineGorgeOracle:      "Copperline Gorge",
		effectstest.B13DourPortMageOracle:         "Dour Port-Mage",
		effectstest.B13CrystalVeinOracle:          "Crystal Vein",
		effectstest.B13GratuitousViolenceOracle:   "Gratuitous Violence",
		effectstest.B13WaveGoodbyeOracle:          "Wave Goodbye",
		effectstest.B13DranaOracle:                "Drana, Liberator of Malakir",
		effectstest.B13RadiantGroveOracle:         "Radiant Grove",
		effectstest.B13InGarruksWakeAlreadyOID:    "In Garruk's Wake",
	}
	if len(want) != 34 {
		t.Fatalf("the batch registers 34 cards, the table lists %d", len(want))
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

// --- the four cycle rows -------------------------------------------

func TestB13LandRowsProduceTheirPrintedColours(t *testing.T) {
	for _, tc := range []struct {
		oracle, name, produced string
		triggers               int
	}{
		{effectstest.B13RuggedHighlandsOracle, "Rugged Highlands", "{R|G}", 1},
		{effectstest.B13InspiringVantageOracle, "Inspiring Vantage", "{R|W}", 0},
		{effectstest.B13CopperlineGorgeOracle, "Copperline Gorge", "{R|G}", 0},
		{effectstest.B13RadiantGroveOracle, "Radiant Grove", "{G|W}", 0},
	} {
		spec, ok := effects.Lookup(tc.oracle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	id := effectstest.B12PlayFromHand(t, g, "Rugged Highlands", "Land", effectstest.B13RuggedHighlandsOracle, game.CastSpellParams{})
	if !effectstest.B13Tapped(t, g, id) {
		t.Fatal("Rugged Highlands enters tapped")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
	if effectstest.TriggerOnStack(g, id) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the life is a trigger with a response window")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
}

func TestB13FastlandRowsEnterUntappedEarly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	vantage := effectstest.B12PlayFromHand(t, g, "Inspiring Vantage", "Land", effectstest.B13InspiringVantageOracle, game.CastSpellParams{})
	if effectstest.B13Tapped(t, g, vantage) {
		t.Error("with two other lands Inspiring Vantage enters untapped")
	}
	gorge := effectstest.B12PlayFromHand(t, g, "Copperline Gorge", "Land", effectstest.B13CopperlineGorgeOracle, game.CastSpellParams{})
	if !effectstest.B13Tapped(t, g, gorge) {
		t.Error("with three other lands Copperline Gorge enters tapped")
	}
}

func TestB13RadiantGroveEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12PlayFromHand(t, g, "Radiant Grove", "Land — Forest Plains", effectstest.B13RadiantGroveOracle, game.CastSpellParams{})
	if !effectstest.B13Tapped(t, g, id) {
		t.Fatal("Radiant Grove enters tapped")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
}

// --- Hedron Crab ---------------------------------------------------

func TestB13HedronCrabMillsATargetPlayerOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Hedron Crab", "Creature — Crab", effectstest.B13HedronCrabOracle, 0, 2)
	before, lib := opp.Graveyard.Size(), opp.Library.Size()

	effectstest.B12PlayFromHand(t, g, "Island", "Basic Land — Island", "", game.CastSpellParams{})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != before+3 || opp.Library.Size() != lib-3 {
		t.Errorf("opponent graveyard %d → %d, library %d → %d; want three milled", before, opp.Graveyard.Size(), lib, opp.Library.Size())
	}
	// An opponent's land is not yours.
	effectstest.B13PlayAs(t, g, 1, "Their Island", "Basic Land — Island", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's land must not trigger the Crab")
	}
	if opp.Graveyard.Size() != before+3 {
		t.Errorf("opponent graveyard %d, want still %d", opp.Graveyard.Size(), before+3)
	}
}

// --- Corpse Knight / Witty Roastmaster / Prosperous Innkeeper --------

func TestB13CorpseKnightDrainsForEachOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := make([]int, 0, 3)
	for _, p := range g.Seats[1:] {
		before = append(before, p.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Corpse Knight", "Creature — Zombie Knight", effectstest.B13CorpseKnightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != before[i] {
			t.Fatalf("the Knight's own entry must not trigger it: opponent %d %d → %d", i+1, before[i], p.Life)
		}
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 2) })
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	effectstest.B12Push(g, me.ID, "Witty Roastmaster", "Creature — Devil Citizen", effectstest.B13WittyRoastmasterOracle, 3, 2)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent %d → %d, want -1", before, opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("a noncreature permanent must not trigger it: %d → %d", before, opp.Life)
	}
}

func TestB13ProsperousInnkeeperMakesATreasureAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Prosperous Innkeeper", "Creature — Halfling Citizen", effectstest.B13ProsperousInnkeeperOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("the ETB makes one Treasure, got %d", n)
	}
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("another creature entering: life %d → %d, want +1", before, me.Life)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("only the Innkeeper's own entry makes a Treasure, got %d", n)
	}
}

// --- Underworld Dreams ---------------------------------------------

func TestB13UnderworldDreamsPingsAnOpponentPerCardDrawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Underworld Dreams", "Enchantment", effectstest.B13UnderworldDreamsOracle, 0, 0)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("two draws: opponent %d → %d, want -2", before, opp.Life)
	}
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("the controller's own draw must not trigger it: %d → %d", mine, me.Life)
	}
}

// --- Imperious Perfect ---------------------------------------------

func TestB13ImperiousPerfectPumpsOtherElvesAndMakesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	perfect := effectstest.B12Push(g, me.ID, "Imperious Perfect", "Creature — Elf Warrior", effectstest.B13ImperiousPerfectOracle, 2, 2)
	elf := effectstest.B12Creature(g, me.ID, "Llanowar", "Creature — Elf Druid", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirElf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	if p := effectstest.EffectivePower(t, g, elf); p != 2 {
		t.Errorf("another Elf you control gets +1/+1: power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, perfect); p != 2 {
		t.Errorf("the Perfect does not pump itself: power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Elf is untouched: power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, theirElf); p != 1 {
		t.Errorf("an opponent's Elf is untouched: power %d, want 1", p)
	}
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateCatalogAbility(me.ID, perfect, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	tok := effectstest.FindBattlefieldByName(g, "Elf Warrior")
	if tok == uuid.Nil {
		t.Fatal("the ability makes an Elf Warrior token")
	}
	if p := effectstest.EffectivePower(t, g, tok); p != 2 {
		t.Errorf("the token is an Elf and gets +1/+1: power %d, want 2", p)
	}
}

// --- Drown in the Loch ---------------------------------------------

func TestB13DrownInTheLochScalesWithTheControllersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	// Empty graveyard: a one-drop is not a legal target.
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Drown in the Loch", "Instant", effectstest.B13DrownInTheLochOracle),
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err == nil {
		t.Fatal("with an empty graveyard a mana value 1 spell is not a legal target")
	}
	effectstest.PushGraveyardCardForTest(opp, "Dead")
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Drown in the Loch", "Instant", effectstest.B13DrownInTheLochOracle),
		game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("with one card in their graveyard the Bolt is a legal target: %v", err)
	}
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Error("the Bolt should be countered")
	}

	// The creature mode, against a three-drop.
	theirs := effectstest.B10Creature(g, opp.ID, "Their Three", "Creature — Beast", "{2}{G}", 3, 3)
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Drown in the Loch", "Instant", effectstest.B13DrownInTheLochOracle),
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("two cards in their graveyard cannot reach a mana value 3 creature")
	}
	effectstest.PushGraveyardCardForTest(opp, "Dead Two")
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Drown in the Loch", "Instant", effectstest.B13DrownInTheLochOracle),
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err != nil {
		t.Fatalf("three cards in their graveyard reach it: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature should be destroyed")
	}
}

// --- Boltwave / Kessig Flamebreather -------------------------------

func TestB13BoltwaveHitsEachOpponentForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := make([]int, 0, 3)
	for _, p := range g.Seats[1:] {
		before = append(before, p.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Boltwave", "Sorcery", effectstest.B13BoltwaveOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Kessig Flamebreather", "Creature — Human Shaman", effectstest.B13KessigFlamebreatherOracle, 1, 3)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("an artifact spell: opponent %d → %d, want -1", before, opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("a creature spell must not trigger it: %d → %d", before, opp.Life)
	}
}

// --- Goldvein Hydra ------------------------------------------------

func TestB13GoldveinHydraEntersWithXAndPaysOutTappedTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hydra := effectstest.B12PlayFromHand(t, g, "Goldvein Hydra", "Creature — Hydra", effectstest.B13GoldveinHydraOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("X=3: the Hydra survives its entry")
	}
	if got := effectstest.B12Counter(t, g, hydra, "+1/+1"); got != 3 {
		t.Fatalf("X=3: %d counters, want 3", got)
	}
	if p := effectstest.B12Card(t, g, hydra).CurrentPower(); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	for _, kw := range []string{"vigilance", "trample", "haste"} {
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, hydra), kw) {
			t.Errorf("printed %s did not reach the effective abilities", kw)
		}
	}
	// One more counter from elsewhere, then it dies: four Treasures,
	// all tapped.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, "+1/+1", 1) })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hydra) })
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hydra := effectstest.CastPrinted(t, g, "Goldvein Hydra", "Creature — Hydra", effectstest.B13GoldveinHydraOracle, game.CastSpellParams{XValue: 0})
	if g.Battlefield.Contains(hydra) {
		t.Fatal("a Goldvein Hydra cast for X=0 stayed on the battlefield — CR 704.5f")
	}
	if !me.Graveyard.Contains(hydra) {
		t.Error("it did not reach its owner's graveyard")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("power 0 makes no Treasures, got %d", n)
	}
}

// --- Long River's Pull ---------------------------------------------

func TestB13LongRiversPullCountersACreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err := b09TryCast(t, g, "Long River's Pull", "Instant", effectstest.B13LongRiversPullOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err == nil {
		t.Fatal("without the gift a noncreature spell is not a legal target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A creature spell is cast on its controller's own turn; the
	// Pull answers it from across the table.
	bear := effectstest.B13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Long River's Pull", "Instant", effectstest.B13LongRiversPullOracle),
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("Long River's Pull at a creature spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell should be countered")
	}
}

// --- Dark Deal -----------------------------------------------------

func TestB13DarkDealWheelsEveryoneForOneFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// Seven in hand each after the opening draw; the caster's Deal
	// leaves the hand with six, so they redraw five.
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Dark Deal", "Sorcery", effectstest.B13DarkDealOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beledros := effectstest.B12Push(g, me.ID, "Beledros Witherbloom", "Legendary Creature — Elder Dragon", effectstest.B13BeledrosWitherbloomOracle, 4, 4)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, beledros), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
	// The next upkeep is an opponent's — still a Pest for Beledros's
	// controller.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Pest"); n != 1 {
		t.Fatalf("an opponent's upkeep: %d Pests, want 1", n)
	}
	// Round the table: seats 2 and 3 each have an upkeep, then yours
	// — four Pests a round at a four-player table.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Pest"); n != 4 {
		t.Errorf("after every seat's upkeep: %d Pests, want 4", n)
	}
	pest := effectstest.B12Card(t, g, effectstest.FindBattlefieldByName(g, "Pest"))
	if !pest.HasColor("B") || !pest.HasColor("G") || !pest.HasSubtype("Pest") {
		t.Errorf("the Pest is a black and green Pest, got %v %q", pest.Colors, pest.TypeLine)
	}
}

// The activated ability has no {T} in its cost, so CR 302.6
// summoning sickness never gates it — SummonedThisTurn is left true
// to prove that. One land is genuinely tapped for mana (through its
// own mana ability, not by hand-setting Tapped) and the other is
// tapped some other way; both must come back up.
func TestB13BeledrosUntapsAllLandsForTenLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beledros := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: beledros, Name: "Beledros Witherbloom", TypeLine: "Legendary Creature — Elder Dragon",
		OracleID: effectstest.B13BeledrosWitherbloomOracle, Power: 4, Toughness: 4,
		Owner: me.ID, Controller: me.ID, SummonedThisTurn: true,
	})
	me.Life = 10

	forest := effectstest.SeedPermanentFor(g, me.ID, "Forest", "Basic Land — Forest")
	swamp := effectstest.SeedPermanentFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Forest for mana: %v", err)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == swamp {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})

	if err := g.ActivateCatalogAbility(me.ID, beledros, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility (a no-tap ability must not be gated by summoning sickness): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if me.Life != 0 {
		t.Errorf("life = %d, want 0 after paying 10", me.Life)
	}
	f, _ := effectstest.BattlefieldCard(g, forest)
	if f.Tapped {
		t.Error("the land tapped for mana this turn should be untapped")
	}
	s, _ := effectstest.BattlefieldCard(g, swamp)
	if s.Tapped {
		t.Error("every land you control should untap")
	}
}

func TestB13BeledrosCannotUntapBelowTenLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beledros := effectstest.B12Push(g, me.ID, "Beledros Witherbloom", "Legendary Creature — Elder Dragon", effectstest.B13BeledrosWitherbloomOracle, 4, 4)
	me.Life = 9

	if err := g.ActivateCatalogAbility(me.ID, beledros, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activating for 10 life while at 9 life should be refused")
	}
}

// "Activate only once each turn" — a second attempt this turn must be
// refused outright, not merely a no-op after paying again.
func TestB13BeledrosOnlyOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beledros := effectstest.B12Push(g, me.ID, "Beledros Witherbloom", "Legendary Creature — Elder Dragon", effectstest.B13BeledrosWitherbloomOracle, 4, 4)
	me.Life = 100

	if err := g.ActivateCatalogAbility(me.ID, beledros, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	lifeAfterFirst := me.Life

	if err := g.ActivateCatalogAbility(me.ID, beledros, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a second activation this turn should be refused (Activate only once each turn)")
	}
	if me.Life != lifeAfterFirst {
		t.Error("a refused activation must not pay any life")
	}
}

// --- Vibrant Cityscape ---------------------------------------------

func TestB13VibrantCityscapeFetchesABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cityscape := effectstest.B12Push(g, me.ID, "Vibrant Cityscape", "Land", effectstest.B13VibrantCityscapeOracle, 0, 0)
	dual := effectstest.StapleLibraryCard(me, "Watery Grave", "Land — Island Swamp")
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, cityscape, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(cityscape) {
		t.Fatal("the Cityscape is sacrificed at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the basic Forest should be fetched")
	}
	if !effectstest.B13Tapped(t, g, forest) {
		t.Error("the fetched land enters tapped")
	}
	if g.Battlefield.Contains(dual) {
		t.Error("a nonbasic is not a legal find")
	}
}

// --- No Mercy ------------------------------------------------------

func TestB13NoMercyDestroysTheCreatureThatDamagedYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "No Mercy", "Enchantment", effectstest.B13NoMercyOracle, 0, 0)
	attacker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)

	effectstest.AangAdvanceToMain(t, g, 1)
	effectstest.AttackWith(t, g, me.ID, attacker)
	if effectstest.TriggerOnStack(g, effectstest.FindBattlefieldByName(g, "No Mercy")) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("combat damage to the controller should trigger No Mercy")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the creature that dealt the damage should be destroyed")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature that dealt no damage is untouched")
	}
	// Non-combat damage from a creature counts too; damage from a
	// spell does not.
	ping := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Wizard", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(ping, me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ping) {
		t.Error("a creature pinging you is destroyed too")
	}
	before := len(g.Events)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventTrigger && ev.Label != "" {
			t.Errorf("damage from no creature must not trigger it: %+v", ev)
		}
	}
}

// --- Fell the Mighty -----------------------------------------------

func TestB13FellTheMightyDestroysEverythingBiggerThanTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Human", 1, 1)
	same := effectstest.B12Creature(g, opp.ID, "Same", "Creature — Bear", 2, 2)
	target := effectstest.B12Creature(g, me.ID, "Target", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Beast", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Beast", 4, 4)

	effectstest.CastCatalogSpell(t, g, "Fell the Mighty", "Sorcery", effectstest.B13FellTheMightyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thalia := effectstest.B12Push(g, me.ID, "Thalia, Heretic Cathar", "Legendary Creature — Human Soldier", effectstest.B13ThaliaHereticCatharOracle, 3, 2)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, thalia), "first strike") {
		t.Error("printed first strike did not reach the effective abilities")
	}
	// The controller's own permanents are untouched.
	mine := effectstest.B12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B13Tapped(t, g, mine) {
		t.Error("Thalia leaves her controller's creatures alone")
	}
	// An opponent's basic enters untapped, a nonbasic tapped.
	basic := effectstest.B13PlayAs(t, g, 1, "Mountain", "Basic Land — Mountain", "")
	if effectstest.B13Tapped(t, g, basic) {
		t.Error("an opponent's basic land enters untapped")
	}
	nonbasic := effectstest.B13PlayAs(t, g, 2, "Ancient Tomb", "Land", "")
	if !effectstest.B13Tapped(t, g, nonbasic) {
		t.Error("an opponent's nonbasic land enters tapped")
	}
	theirs := effectstest.B13PlayAs(t, g, 3, "Their Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B13Tapped(t, g, theirs) {
		t.Error("an opponent's creature enters tapped")
	}
	// An opponent's noncreature, nonland permanent is untouched.
	rock := effectstest.B13PlayAs(t, g, 1, "Rock", "Artifact", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B13Tapped(t, g, rock) {
		t.Error("an opponent's artifact is not a creature or a nonbasic land")
	}
}

// --- Concordant Crossroads -----------------------------------------

func TestB13ConcordantCrossroadsGivesEveryCreatureHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Concordant Crossroads", "World Enchantment", effectstest.B13ConcordantCrossroadsOracle, 0, 0)
	mine := effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, id), "haste") {
			t.Errorf("every creature has haste, %s does not", effectstest.B12Card(t, g, id).Name)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Errorf("a creature cast this turn attacks under the Crossroads: %v", err)
	}
}

// --- World Shaper --------------------------------------------------

func TestB13WorldShaperMillsOnAttackAndReturnsLandsOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaper := effectstest.B12Push(g, me.ID, "World Shaper", "Creature — Merfolk Shaman", effectstest.B13WorldShaperOracle, 3, 3)
	forest := effectstest.B10LibraryTop(me, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.B10LibraryTop(me, "Spell", "Sorcery", "{G}", 0, 0)
	effectstest.B10LibraryTop(me, "Island", "Basic Land — Island", "", 0, 0)
	before := me.Graveyard.Size()

	effectstest.DeclareAttack(t, g, opp.ID, shaper)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != before+3 {
		t.Fatalf("graveyard %d → %d, want three milled", before, me.Graveyard.Size())
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shaper) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Error("the milled Forest should return to the battlefield")
	}
	if !effectstest.B13Tapped(t, g, forest) {
		t.Error("the returned land is tapped")
	}
	if effectstest.FindBattlefieldByName(g, "Spell") != uuid.Nil {
		t.Error("only land cards return")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Island"); n != 1 {
		t.Errorf("both milled lands return: %d Islands", n)
	}
}

// --- Lae'zel, Vlaakith's Champion ----------------------------------

func TestB13LaezelAddsOneCounterOfEachKindYouPut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Lae'zel, Vlaakith's Champion", "Legendary Creature — Gith Warrior", effectstest.B13LaezelOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	walker := effectstest.B12Permanent(g, me.ID, "Walker", "Legendary Planeswalker — Test")
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(me.ID, bear, "+1/+1", 2)
		_ = g.AddCounterByForEffect(me.ID, walker, "loyalty", 1)
		_ = g.AddCounterByForEffect(me.ID, rock, "charge", 1)
		_ = g.AddCounterByForEffect(me.ID, theirs, "+1/+1", 1)
	})
	if got := effectstest.B12Counter(t, g, bear, "+1/+1"); got != 3 {
		t.Errorf("two +1/+1 counters become three: got %d", got)
	}
	if got := effectstest.B12Counter(t, g, walker, "loyalty"); got != 2 {
		t.Errorf("one loyalty counter on your planeswalker becomes two: got %d", got)
	}
	if got := effectstest.B12Counter(t, g, rock, "charge"); got != 1 {
		t.Errorf("an artifact is neither a creature nor a planeswalker: got %d", got)
	}
	if got := effectstest.B12Counter(t, g, theirs, "+1/+1"); got != 1 {
		t.Errorf("an opponent's creature is untouched: got %d", got)
	}
	// A removal is not a placement, and an opponent putting a counter
	// on your creature is not you putting.
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, "+1/+1", -1)
		_ = g.AddCounterByForEffect(opp.ID, bear, "+1/+1", 1)
	})
	if got := effectstest.B12Counter(t, g, bear, "+1/+1"); got != 3 {
		t.Errorf("3 -1 +1 = 3 with neither the removal nor an opponent's placement increased: got %d", got)
	}
}

// --- Farhaven Elf --------------------------------------------------

func TestB13FarhavenElfMayFetchABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Farhaven Elf", "Creature — Elf Druid", effectstest.B13FarhavenElfOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the Forest should be fetched")
	}
	if !effectstest.B13Tapped(t, g, forest) {
		t.Error("the fetched land enters tapped")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	forest2 := effectstest.StapleLibraryCard(me2, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g2, "Farhaven Elf", "Creature — Elf Druid", effectstest.B13FarhavenElfOracle, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerLatestTriggerPrompt(t, g2, me2.ID, false)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(forest2) {
		t.Error("declining the prompt fetches nothing")
	}
}

// --- Sunbaked Canyon / Crystal Vein --------------------------------

func TestB13SunbakedCanyonPaysLifeForManaAndCashesInForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	canyon := effectstest.B12Push(g, me.ID, "Sunbaked Canyon", "Land", effectstest.B13SunbakedCanyonOracle, 0, 0)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, canyon, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	if n := effectstest.B10ResolveAllManaPicks(t, g, me.ID, "W"); n != 1 {
		t.Fatalf("expected one colour pick, answered %d", n)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
	effectstest.B08Untap(g, canyon)
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, canyon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	if g.Battlefield.Contains(canyon) {
		t.Error("the Canyon was sacrificed")
	}
}

func TestB13CrystalVeinTapsForOneOrCracksForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vein := effectstest.B12Push(g, me.ID, "Crystal Vein", "Land", effectstest.B13CrystalVeinOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, vein, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for one: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	if !g.Battlefield.Contains(vein) {
		t.Fatal("the first ability does not sacrifice")
	}
	effectstest.B08Untap(g, vein)
	if err := g.ActivateManaAbility(me.ID, vein, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack for two: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 3 {
		t.Errorf("pool %v, want three colourless", got)
	}
	if g.Battlefield.Contains(vein) {
		t.Error("the second ability sacrifices the Vein")
	}
}

// --- Heliod's Intervention -----------------------------------------

func TestB13HeliodsInterventionDestroysXTargetsOrGainsTwiceX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	aura := effectstest.B12Permanent(g, opp.ID, "Aura", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	two := []game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: aura}}
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Heliod's Intervention", "Instant", effectstest.B13HeliodsInterventionOracle),
		game.CastSpellParams{XValue: 1, Modes: []int{0}, Targets: two}); err == nil {
		t.Fatal("X=1 with two targets is not a legal announcement")
	}
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Heliod's Intervention", "Instant", effectstest.B13HeliodsInterventionOracle),
		game.CastSpellParams{XValue: 1, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Heliod's Intervention", "Instant", effectstest.B13HeliodsInterventionOracle),
		game.CastSpellParams{XValue: 2, Modes: []int{0}, Targets: two}); err != nil {
		t.Fatalf("X=2 with two targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(aura) {
		t.Error("both targets should be destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was never a target")
	}

	before := me.Life
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Heliod's Intervention", "Instant", effectstest.B13HeliodsInterventionOracle),
		game.CastSpellParams{XValue: 3, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("the lifegain mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+6 {
		t.Errorf("X=3: life %d → %d, want +6", before, me.Life)
	}
}

// --- Twitching Doll ------------------------------------------------

func TestB13TwitchingDollBanksNestCountersIntoSpiders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	doll := effectstest.B12Push(g, me.ID, "Twitching Doll", "Artifact Creature — Spider Toy", effectstest.B13TwitchingDollOracle, 2, 2)
	for i := 1; i <= 2; i++ {
		effectstest.B08Untap(g, doll)
		if err := g.ActivateManaAbility(me.ID, doll, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility %d: %v", i, err)
		}
		if n := effectstest.B10ResolveAllManaPicks(t, g, me.ID, "U"); n != 1 {
			t.Fatalf("expected one colour pick, answered %d", n)
		}
		if got := effectstest.B12Counter(t, g, doll, "nest"); got != i {
			t.Fatalf("after %d taps: %d nest counters", i, got)
		}
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("pool %v, want [U U]", got)
	}
	// One counter of another kind counts too.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(doll, "+1/+1", 1) })
	effectstest.B08Untap(g, doll)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, doll, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(doll) {
		t.Fatal("the Doll is sacrificed at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	spiders := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spider" && c.Controller == me.ID {
			spiders++
			if p := effectstest.EffectivePower(t, g, c.InstanceID); p != 2 {
				t.Errorf("Spider power %d, want 2", p)
			}
			if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, c.InstanceID), "reach") {
				t.Error("the Spiders have reach")
			}
		}
	}
	if spiders != 3 {
		t.Errorf("three counters make %d Spiders, want 3", spiders)
	}
}

func TestB13TwitchingDollSacrificeIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	doll := effectstest.B12Push(g, me.ID, "Twitching Doll", "Artifact Creature — Spider Toy", effectstest.B13TwitchingDollOracle, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, doll, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the sacrifice ability is sorcery-speed only")
	}
}

// --- Dour Port-Mage ------------------------------------------------

func TestB13DourPortMageDrawsWhenACreatureLeavesWithoutDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Dour Port-Mage", "Creature — Frog Wizard", effectstest.B13DourPortMageOracle, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	third := effectstest.B12Creature(g, me.ID, "Third", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, bear) != game.ZoneHand {
		t.Fatal("the Bear should be in hand")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("the bounced Bear plus one drawn: hand %d → %d, want +2", hand, me.Hand.Size())
	}
	// Dying is not leaving without dying.
	hand = me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a death must not draw: hand %d → %d", hand, me.Hand.Size())
	}
	// Two leaving at once is one draw.
	fourth := effectstest.B12Creature(g, me.ID, "Fourth", "Creature — Bear", 2, 2)
	hand = me.Hand.Size()
	g.WithWriteLock(func() { _ = g.BounceCardsToHandForEffect([]uuid.UUID{third, fourth}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("two bounced plus ONE drawn: hand %d → %d, want +3", hand, me.Hand.Size())
	}
	// An opponent's creature leaving is not yours.
	hand = me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("an opponent's creature leaving must not draw: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Gratuitous Violence -------------------------------------------

func TestB13GratuitousViolenceDoublesYourCreaturesDamageOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Gratuitous Violence", "Enchantment", effectstest.B13GratuitousViolenceOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("a 2-power attacker under the Violence: %d → %d, want -4", before, opp.Life)
	}
	before = opp.Life
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-3 {
		t.Errorf("a Bolt is not a creature: %d → %d, want -3", before, opp.Life)
	}
	// An opponent's creature is not yours.
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 2) })
	if me.Life != mine-2 {
		t.Errorf("an opponent's creature's damage is not doubled: %d → %d, want -2", mine, me.Life)
	}
}

// --- Wave Goodbye --------------------------------------------------

func TestB13WaveGoodbyeReturnsEveryCreatureWithoutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	countered := effectstest.B12Creature(g, me.ID, "Countered", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(countered, "+1/+1", 1) })
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	charged := effectstest.B12Creature(g, opp.ID, "Charged", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(charged, "-1/-1", 1) })
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Wave Goodbye", "Sorcery", effectstest.B13WaveGoodbyeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(countered) {
		t.Error("a creature with a +1/+1 counter stays")
	}
	for _, id := range []uuid.UUID{mine, theirs, charged} {
		if effectstest.B12ZoneOf(g, id) != game.ZoneHand {
			t.Errorf("%s should be in its owner's hand", effectstest.B12Card(t, g, id).Name)
		}
	}
	if effectstest.FindBattlefieldByName(g, "Goblin") != uuid.Nil {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drana := effectstest.B12Push(g, me.ID, "Drana, Liberator of Malakir", "Legendary Creature — Vampire Ally", effectstest.B13DranaOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	homebody := effectstest.B12Creature(g, me.ID, "Homebody", "Creature — Bear", 2, 2)
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, drana), kw) {
			t.Errorf("printed %s did not reach the effective abilities", kw)
		}
	}
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, drana, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-5 {
		t.Errorf("opponent %d → %d, want -5 (her 2 first-strike, then a Bear grown to 3)", before, opp.Life)
	}
	for _, id := range []uuid.UUID{drana, bear} {
		if got := effectstest.B12Counter(t, g, id, "+1/+1"); got != 1 {
			t.Errorf("%s: %d counters, want 1", effectstest.B12Card(t, g, id).Name, got)
		}
	}
	if got := effectstest.B12Counter(t, g, homebody, "+1/+1"); got != 0 {
		t.Errorf("a creature that stayed home got %d counters", got)
	}
}
