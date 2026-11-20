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
// in cycle tables (Razorverge Thicket in fastlands.go, Sodden Verdure
// in battle_lands.go, Gold Myr in mana_myr.go, Tangled Islet in
// dominaria_united_duals.go), so a transposed row is invisible until
// someone plays that exact card. Two of the issue's 36 were already
// on main — Grave Titan (#73) and Reverberate (#419) — pinned here so
// the table matches the issue.
func TestBatch12CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B12AwakenTheWoodsOracle:       "Awaken the Woods",
		effectstest.B12LaeliaOracle:               "Laelia, the Blade Reforged",
		effectstest.B12ValakutOracle:              "Valakut, the Molten Pinnacle",
		effectstest.B12KrarkClanIronworksOracle:   "Krark-Clan Ironworks",
		effectstest.B12SkirkProspectorOracle:      "Skirk Prospector",
		effectstest.B12ColossalMajestyOracle:      "Colossal Majesty",
		effectstest.B12PersistOracle:              "Persist",
		effectstest.B12PristineTalismanOracle:     "Pristine Talisman",
		effectstest.B12SoulGuideLanternOracle:     "Soul-Guide Lantern",
		effectstest.B12ThirdPathIconoclastOracle:  "Third Path Iconoclast",
		effectstest.B12CrimeNovelistOracle:        "Crime Novelist",
		effectstest.B12SigilOfTheEmptyThroneOrcl:  "Sigil of the Empty Throne",
		effectstest.B12BonehoardDracosaurOracle:   "Bonehoard Dracosaur",
		effectstest.B12ArchangelOfThuneOracle:     "Archangel of Thune",
		effectstest.B12SkullProphetOracle:         "Skull Prophet",
		effectstest.B12MartialCoupOracle:          "Martial Coup",
		effectstest.B12TerastodonOracle:           "Terastodon",
		effectstest.B12LifecraftersBestiaryOracle: "Lifecrafter's Bestiary",
		effectstest.B12AllWillBeOneOracle:         "All Will Be One",
		effectstest.B12ThousandYearStormOracle:    "Thousand-Year Storm",
		effectstest.B12RazorvergeThicketOracle:    "Razorverge Thicket",
		effectstest.B12TimeWipeOracle:             "Time Wipe",
		effectstest.B12CatharCommandoOracle:       "Cathar Commando",
		effectstest.B12AltarOfTheBroodOracle:      "Altar of the Brood",
		effectstest.B12DuskshellCrawlerOracle:     "Duskshell Crawler",
		effectstest.B12SoddenVerdureOracle:        "Sodden Verdure",
		effectstest.B12GoldMyrOracle:              "Gold Myr",
		effectstest.B12TangledIsletOracle:         "Tangled Islet",
		effectstest.B12GraveTitanAlreadyOID:       "Grave Titan",
		effectstest.B12ReverberateAlreadyOID:      "Reverberate",
	}
	if len(want) != 30 {
		t.Fatalf("the batch registers 30 cards, the table lists %d", len(want))
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

// The two cycle rows produce their printed pairs behind their printed
// conditions — the loop-leak canary every land table carries.
func TestB12LandRowsProduceTheirPrintedColours(t *testing.T) {
	for _, tc := range []struct{ oracle, name, produced string }{
		{effectstest.B12RazorvergeThicketOracle, "Razorverge Thicket", "{G|W}"},
		{effectstest.B12SoddenVerdureOracle, "Sodden Verdure", "{G|U}"},
		{effectstest.B12TangledIsletOracle, "Tangled Islet", "{G|U}"},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Fatalf("%s not registered", tc.name)
		}
		if len(spec.Replacements) != 1 {
			t.Errorf("%s: %d replacements, want 1 (enters tapped unless …)", tc.name, len(spec.Replacements))
		}
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

func TestB12GoldMyrTapsForWhiteOnceAwake(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	myr := effectstest.PushCatalogPermanent(g, me.ID, "Gold Myr", "Artifact Creature — Myr", effectstest.B12GoldMyrOracle, true)
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a Myr that entered this turn is summoning sick")
	}
	effectstest.B10Awake(g, myr)
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
}

func TestB12TangledIsletEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12PlayFromHand(t, g, "Tangled Islet", "Land — Forest Island", effectstest.B12TangledIsletOracle, game.CastSpellParams{})
	if !effectstest.B12Card(t, g, id).Tapped {
		t.Fatal("Tangled Islet enters tapped")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
}

func TestB12SoddenVerdureEntersUntappedWithTwoBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	id := effectstest.B12PlayFromHand(t, g, "Sodden Verdure", "Land — Forest Island", effectstest.B12SoddenVerdureOracle, game.CastSpellParams{})
	if effectstest.B12Card(t, g, id).Tapped {
		t.Error("with two basics Sodden Verdure enters untapped")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.SeedLandOnBattlefield(g2, me2.ID, "Forest", "Basic Land — Forest")
	id2 := effectstest.B12PlayFromHand(t, g2, "Sodden Verdure", "Land — Forest Island", effectstest.B12SoddenVerdureOracle, game.CastSpellParams{})
	if !effectstest.B12Card(t, g2, id2).Tapped {
		t.Error("with one basic Sodden Verdure enters tapped")
	}
}

func TestB12RazorvergeThicketEntersUntappedEarly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	id := effectstest.B12PlayFromHand(t, g, "Razorverge Thicket", "Land", effectstest.B12RazorvergeThicketOracle, game.CastSpellParams{})
	if effectstest.B12Card(t, g, id).Tapped {
		t.Error("with two other lands the fastland enters untapped")
	}
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	id2 := effectstest.B12PlayFromHand(t, g, "Razorverge Thicket", "Land", effectstest.B12RazorvergeThicketOracle, game.CastSpellParams{})
	if !effectstest.B12Card(t, g, id2).Tapped {
		t.Error("with four other lands the fastland enters tapped")
	}
}

// --- Awaken the Woods ----------------------------------------------

func TestB12AwakenTheWoodsMakesXForestDryadLandCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12PlayFromHand(t, g, "Awaken the Woods", "Sorcery", effectstest.B12AwakenTheWoodsOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, id) != game.ZoneGraveyard {
		t.Fatal("the sorcery should be in the graveyard")
	}
	var dryads []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest Dryad" && c.Controller == me.ID {
			dryads = append(dryads, c)
		}
	}
	if len(dryads) != 3 {
		t.Fatalf("X=3 made %d Dryads, want 3", len(dryads))
	}
	d := dryads[0]
	if !d.IsLand() || !d.IsCreature() || !d.HasSubtype("Forest") || !d.HasSubtype("Dryad") {
		t.Errorf("the token must be a Forest Dryad land creature, got %q", d.TypeLine)
	}
	if !effects.IsToken(d) {
		t.Error("the Dryad must be a token")
	}
	abilities := game.ManaAbilitiesForCard(d)
	if len(abilities) != 1 || abilities[0].Produced != "{G}" {
		t.Errorf("a Forest token taps for {G} from its type line, got %+v", abilities)
	}
	// It is a creature that just entered, so it is summoning sick.
	if err := g.ActivateManaAbility(me.ID, d.InstanceID, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a Dryad that entered this turn must be summoning sick")
	}
}

// --- Laelia, the Blade Reforged ------------------------------------

func TestB12LaeliaAttacksForACardAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	laelia := effectstest.B12Push(g, me.ID, "Laelia, the Blade Reforged", "Legendary Creature — Spirit Warrior", effectstest.B12LaeliaOracle, 2, 2)
	top := effectstest.B10LibraryTop(me, "Top Card", "Sorcery", "{R}", 0, 0)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(laelia, victim.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Fatal("the top card should be in exile with a play grant for Laelia's controller")
	}
	if got := effectstest.B12Counter(t, g, laelia, "+1/+1"); got != 1 {
		t.Errorf("Laelia's own exile grows her: %d counters, want 1", got)
	}
	if p := effectstest.B12Card(t, g, laelia).CurrentPower(); p != 3 {
		t.Errorf("Laelia power %d, want 3", p)
	}
}

func TestB12LaeliaGrowsOncePerBatchOfExiledCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	laelia := effectstest.B12Push(g, me.ID, "Laelia, the Blade Reforged", "Legendary Creature — Spirit Warrior", effectstest.B12LaeliaOracle, 2, 2)
	a := effectstest.PushGraveyardCardForTest(me, "Dead A")
	b := effectstest.PushGraveyardCardForTest(me, "Dead B")
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Dead")

	// One effect exiling two of your graveyard cards is one batch.
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(a)
		_ = g.ExileCardForEffect(b)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, laelia, "+1/+1"); got != 1 {
		t.Errorf("two cards in one batch: %d counters, want 1", got)
	}

	// An opponent's graveyard is not yours.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, laelia, "+1/+1"); got != 1 {
		t.Errorf("an opponent's card leaving their graveyard grew Laelia: %d counters", got)
	}
}

// --- Valakut, the Molten Pinnacle ----------------------------------

func TestB12ValakutEntersTappedAndTapsForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12PlayFromHand(t, g, "Valakut, the Molten Pinnacle", "Land", effectstest.B12ValakutOracle, game.CastSpellParams{})
	if !effectstest.B12Card(t, g, id).Tapped {
		t.Fatal("Valakut enters tapped")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
	effectstest.B08Untap(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
}

func TestB12ValakutBoltsOnTheSixthMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Valakut, the Molten Pinnacle", "Land", effectstest.B12ValakutOracle, 0, 0)
	for i := 0; i < 4; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	before := opp.Life

	// Four others: no trigger at all.
	effectstest.B12PlayFromHand(t, g, "Mountain", "Basic Land — Mountain", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("four other Mountains must not trigger Valakut")
	}

	// Five others: the prompt, the target, the damage.
	effectstest.B12PlayFromHand(t, g, "Mountain", "Basic Land — Mountain", "", game.CastSpellParams{})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent %d → %d, want -3", before, opp.Life)
	}
	// The subtype is what matters: a Forest is not a Mountain.
	prompts := effectstest.B06TriggerPromptsFor(g, me.ID)
	effectstest.B12PlayFromHand(t, g, "Dryad Arbor", "Land Creature — Forest Dryad", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B06TriggerPromptsFor(g, me.ID) != prompts {
		t.Error("a Forest is not a Mountain")
	}
}

func TestB12ValakutRechecksTheFiveMountainsOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Valakut, the Molten Pinnacle", "Land", effectstest.B12ValakutOracle, 0, 0)
	var mountains []uuid.UUID
	for i := 0; i < 5; i++ {
		mountains = append(mountains, effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain"))
	}
	before := opp.Life
	effectstest.B12PlayFromHand(t, g, "Mountain", "Basic Land — Mountain", "", game.CastSpellParams{})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	// In response, one of the other Mountains leaves. CR 603.4 is
	// checked before the "you may" (#3907): nothing is asked.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mountains[0]) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("a trigger whose condition failed asked its \"you may\"")
	}
	if opp.Life != before {
		t.Errorf("with only four other Mountains at resolution nothing happens; opponent %d → %d", before, opp.Life)
	}
}

// --- Krark-Clan Ironworks / Skirk Prospector -----------------------

func TestB12KrarkClanIronworksEatsArtifactsForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kci := effectstest.B12Push(g, me.ID, "Krark-Clan Ironworks", "Artifact", effectstest.B12KrarkClanIronworksOracle, 0, 0)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, kci, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature that is not an artifact is not a legal sacrifice")
	}
	if err := g.ActivateManaAbility(me.ID, kci, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("pool %v, want [C C]", got)
	}
	if effectstest.B12ZoneOf(g, rock) != game.ZoneGraveyard {
		t.Error("the sacrificed artifact should be in the graveyard")
	}
	// The Ironworks is an artifact, so it can eat itself.
	if err := g.ActivateManaAbility(me.ID, kci, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{kci}}); err != nil {
		t.Fatalf("self-sacrifice: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 4 {
		t.Errorf("pool %v, want four colourless", got)
	}
	if g.Battlefield.Contains(kci) {
		t.Error("the Ironworks should have sacrificed itself")
	}
}

func TestB12SkirkProspectorEatsGoblinsForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prospector := effectstest.B12Push(g, me.ID, "Skirk Prospector", "Creature — Goblin", effectstest.B12SkirkProspectorOracle, 1, 1)
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, prospector, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a Bear is not a Goblin")
	}
	if err := g.ActivateManaAbility(me.ID, prospector, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{goblin}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
	if err := g.ActivateManaAbility(me.ID, prospector, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{prospector}}); err != nil {
		t.Fatalf("self-sacrifice: %v", err)
	}
	if g.Battlefield.Contains(prospector) {
		t.Error("the Prospector is a Goblin and can eat itself")
	}
}

// --- Colossal Majesty ----------------------------------------------

func TestB12ColossalMajestyDrawsWithAFourPowerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Colossal Majesty", "Enchantment", effectstest.B12ColossalMajestyOracle, 0, 0)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// The draw step adds one; the Majesty must add nothing.
	if me.Hand.Size() != before {
		t.Fatalf("no power-4 creature: hand %d → %d during the upkeep", before, me.Hand.Size())
	}

	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Beast", 4, 4)
	effectstest.B12ToMyNextUpkeep(t, g)
	before = me.Hand.Size()
	if len(g.PendingTriggers) == 0 && effectstest.TriggerOnStack(g, effectstest.FindBattlefieldByName(g, "Colossal Majesty")) == nil {
		t.Fatal("the upkeep trigger should be waiting")
	}
	// Removed in response: the intervening-if fails on resolution.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(big) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Errorf("the power-4 creature left in response; hand %d → %d, want unchanged", before, me.Hand.Size())
	}

	effectstest.B12Creature(g, me.ID, "Bigger", "Creature — Beast", 5, 5)
	effectstest.B12ToMyNextUpkeep(t, g)
	before = me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("with a power-5 creature: hand %d → %d, want +1", before, me.Hand.Size())
	}
}

// --- Persist -------------------------------------------------------

func TestB12PersistReturnsANonlegendaryCreatureWithAMinusCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	legend := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: legend, Name: "Dead Legend", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})

	if err := b09TryCast(t, g, "Persist", "Sorcery", effectstest.B12PersistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: legend}}); err == nil {
		t.Fatal("a legendary creature card is not a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Persist", "Sorcery", effectstest.B12PersistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the Bear should be back on the battlefield")
	}
	if got := effectstest.B12Counter(t, g, dead, "-1/-1"); got != 1 {
		t.Errorf("-1/-1 counters %d, want 1", got)
	}
	if p := effectstest.B12Card(t, g, dead).CurrentPower(); p != 1 {
		t.Errorf("a 2/2 with a -1/-1 counter has power %d, want 1", p)
	}
	if c := effectstest.B12Card(t, g, dead); c.Controller != me.ID {
		t.Error("the creature returns under its owner's control")
	}
}

// --- Pristine Talisman ---------------------------------------------

func TestB12PristineTalismanAddsColorlessAndALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Push(g, me.ID, "Pristine Talisman", "Artifact", effectstest.B12PristineTalismanOracle, 0, 0)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
}

// --- Soul-Guide Lantern --------------------------------------------

func TestB12SoulGuideLanternExilesATargetCardOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Dead")
	mine := effectstest.PushGraveyardCardForTest(me, "My Dead")

	effectstest.CastCatalogSpell(t, g, "Soul-Guide Lantern", "Artifact", effectstest.B12SoulGuideLanternOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneExile {
		t.Error("the targeted card should be exiled")
	}
	if effectstest.B12ZoneOf(g, mine) != game.ZoneGraveyard {
		t.Error("the other card stays")
	}
}

func TestB12SoulGuideLanternSweepsOpponentsGraveyardsOrDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	lantern := effectstest.B12Push(g, me.ID, "Soul-Guide Lantern", "Artifact", effectstest.B12SoulGuideLanternOracle, 0, 0)
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Dead")
	others := effectstest.PushGraveyardCardForTest(other, "Other Dead")
	mine := effectstest.PushGraveyardCardForTest(me, "My Dead")

	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, lantern, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneExile || effectstest.B12ZoneOf(g, others) != game.ZoneExile {
		t.Error("every opponent's graveyard should be exiled")
	}
	if effectstest.B12ZoneOf(g, mine) != game.ZoneGraveyard {
		t.Error("your own graveyard is untouched")
	}
	if effectstest.B12ZoneOf(g, lantern) != game.ZoneGraveyard {
		t.Error("the Lantern was sacrificed")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	lantern2 := effectstest.B12Push(g2, me2.ID, "Soul-Guide Lantern", "Artifact", effectstest.B12SoulGuideLanternOracle, 0, 0)
	effectstest.AdvanceToMain(t, g2)
	me2.ManaPool.AddMana(game.ManaToken{Color: "C"})
	before := me2.Hand.Size()
	if err := g2.ActivateCatalogAbility(me2.ID, lantern2, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Hand.Size() != before+1 {
		t.Errorf("hand %d → %d, want +1", before, me2.Hand.Size())
	}
	if g2.Battlefield.Contains(lantern2) {
		t.Error("the Lantern was sacrificed")
	}
}

// --- Third Path Iconoclast / Sigil of the Empty Throne --------------

func TestB12ThirdPathIconoclastMakesArtifactSoldiersForNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Third Path Iconoclast", "Creature — Human Monk", effectstest.B12ThirdPathIconoclastOracle, 2, 1)
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	soldiers := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && c.Controller == me.ID {
			soldiers++
			if !c.IsArtifact() || !c.IsCreature() || !c.IsColorless() {
				t.Errorf("the Soldier is a colorless artifact creature, got %q colours %v", c.TypeLine, c.Colors)
			}
		}
	}
	if soldiers != 1 {
		t.Fatalf("a Bolt should make one Soldier, made %d", soldiers)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("a creature spell must not trigger it: %d Soldiers", n)
	}
}

func TestB12SigilOfTheEmptyThroneMakesAnAngelPerEnchantmentCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sigil of the Empty Throne", "Enchantment", effectstest.B12SigilOfTheEmptyThroneOrcl, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Some Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	angel := effectstest.FindBattlefieldByName(g, "Angel")
	if angel == uuid.Nil {
		t.Fatal("an enchantment cast should make an Angel")
	}
	if p := effectstest.EffectivePower(t, g, angel); p != 4 {
		t.Errorf("Angel power %d, want 4", p)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, angel), "flying") {
		t.Error("the Angel has flying")
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Angel"); n != 1 {
		t.Errorf("an artifact must not trigger it: %d Angels", n)
	}
}

// --- Crime Novelist ------------------------------------------------

func TestB12CrimeNovelistGrowsAndAddsRedWhenAnArtifactIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	novelist := effectstest.B12Push(g, me.ID, "Crime Novelist", "Creature — Goblin Bard", effectstest.B12CrimeNovelistOracle, 1, 3)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, treasure, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack the Treasure: %v", err)
	}
	effectstest.B10ResolveAllManaPicks(t, g, me.ID, "G")
	if effectstest.TriggerOnStack(g, novelist) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("sacrificing the Treasure should trigger the Novelist")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, novelist, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	red := 0
	for _, c := range effectstest.Batch01PoolColors(me) {
		if c == "R" {
			red++
		}
	}
	if red != 1 {
		t.Errorf("pool %v, want one {R} from the Novelist", effectstest.Batch01PoolColors(me))
	}
}

// --- Bonehoard Dracosaur -------------------------------------------

func TestB12BonehoardDracosaurExilesTwoAndPaysForEachKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Bonehoard Dracosaur", "Creature — Dinosaur Dragon", effectstest.B12BonehoardDracosaurOracle, 5, 5)
	spell := effectstest.B10LibraryTop(me, "A Spell", "Sorcery", "{R}", 0, 0)
	land := effectstest.B10LibraryTop(me, "A Mountain", "Basic Land — Mountain", "", 0, 0)

	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{spell, land} {
		if !effectstest.B12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("%s should be exiled with a play grant", id)
		}
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Dinosaur"); n != 1 {
		t.Errorf("a land exiled: %d Dinosaurs, want 1", n)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("a nonland exiled: %d Treasures, want 1", n)
	}

	// Two lands: one Dinosaur, no Treasure. Past this turn's draw step
	// first, so the draw takes a filler rather than a seeded land.
	effectstest.AdvanceTo(t, g, game.StepDraw)
	effectstest.B10LibraryTop(me, "Mountain A", "Basic Land — Mountain", "", 0, 0)
	effectstest.B10LibraryTop(me, "Mountain B", "Basic Land — Mountain", "", 0, 0)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Dinosaur"); n != 2 {
		t.Errorf("two lands make one more Dinosaur: %d, want 2", n)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("two lands make no Treasure: %d, want 1", n)
	}
}

// --- Archangel of Thune --------------------------------------------

func TestB12ArchangelOfThuneGrowsTheTeamOnLifegain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.B12Push(g, me.ID, "Archangel of Thune", "Creature — Angel", effectstest.B12ArchangelOfThuneOracle, 3, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{angel, bear} {
		if got := effectstest.B12Counter(t, g, id, "+1/+1"); got != 1 {
			t.Errorf("one lifegain event: %d counters, want 1", got)
		}
	}
	if got := effectstest.B12Counter(t, g, theirs, "+1/+1"); got != 0 {
		t.Errorf("an opponent's creature grew: %d", got)
	}
	// Losing life is not gaining it.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, bear, "+1/+1"); got != 1 {
		t.Errorf("life loss triggered the Archangel: %d counters", got)
	}
}

// --- Skull Prophet -------------------------------------------------

func TestB12SkullProphetTapsForManaOrMillsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prophet := effectstest.B12Push(g, me.ID, "Skull Prophet", "Creature — Human Druid", effectstest.B12SkullProphetOracle, 3, 1)
	effectstest.AdvanceToMain(t, g)
	before := me.Graveyard.Size()
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, prophet, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("mill: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != before+2 || me.Library.Size() != lib-2 {
		t.Errorf("graveyard %d → %d, library %d → %d; want two milled", before, me.Graveyard.Size(), lib, me.Library.Size())
	}
	if err := g.ActivateManaAbility(me.ID, prophet, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a tapped Prophet cannot tap for mana")
	}
	effectstest.B08Untap(g, prophet)
	if err := g.ActivateManaAbility(me.ID, prophet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if n := effectstest.B10ResolveAllManaPicks(t, g, me.ID, "B"); n != 1 {
		t.Fatalf("expected one colour pick, answered %d", n)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
}

// --- Martial Coup --------------------------------------------------

func TestB12MartialCoupWipesOnlyAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.B12PlayFromHand(t, g, "Martial Coup", "Sorcery", effectstest.B12MartialCoupOracle, game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Soldier"); n != 2 {
		t.Fatalf("X=2: %d Soldiers, want 2", n)
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Fatal("X=2 destroys nothing")
	}

	effectstest.B12PlayFromHand(t, g, "Martial Coup", "Sorcery", effectstest.B12MartialCoupOracle, game.CastSpellParams{XValue: 5})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Soldier"); n != 5 {
		t.Errorf("X=5: %d Soldiers survive, want the five new ones", n)
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("X=5 destroys every other creature, yours included")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && (len(c.Colors) != 1 || c.Colors[0] != "W") {
			t.Errorf("the Soldiers are white, got %v", c.Colors)
		}
	}
}

// --- Terastodon ----------------------------------------------------

func TestB12TerastodonTradesNoncreaturePermanentsForElephants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myLand := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	theirRock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Terastodon", "Creature — Elephant", effectstest.B12TerastodonOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p.PickTargetMin != 0 || p.PickTargetMax != 3 {
		t.Fatalf("up to three: min %d max %d", p.PickTargetMin, p.PickTargetMax)
	}
	if effectstest.HasID(p.PickTargetCards, theirBear) {
		t.Error("a creature is not a legal target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: myLand},
		{Kind: game.TargetCard, ID: theirRock},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myLand) || g.Battlefield.Contains(theirRock) {
		t.Error("both targets should be destroyed")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Elephant"); n != 1 {
		t.Errorf("my land's controller (me) gets one Elephant, got %d", n)
	}
	if n := effectstest.CountBattlefieldNamed(g, opp.ID, "Elephant"); n != 1 {
		t.Errorf("their Rock's controller gets one Elephant, got %d", n)
	}
	if g.Battlefield.Contains(theirBear) == false {
		t.Error("the creature was never a target")
	}
}

func TestB12TerastodonPaysNothingForAnIndestructiblePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forge := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Rock", TypeLine: "Artifact",
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Terastodon", "Creature — Elephant", effectstest.B12TerastodonOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, forge)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forge) {
		t.Fatal("an indestructible artifact survives")
	}
	if n := effectstest.CountBattlefieldNamed(g, opp.ID, "Elephant"); n != 0 {
		t.Errorf("nothing reached a graveyard, so no Elephant; got %d", n)
	}
}

func TestB12TerastodonMayChooseNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	myLand := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Terastodon", "Creature — Elephant", effectstest.B12TerastodonOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining with zero targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(myLand) {
		t.Error("nothing chosen, nothing destroyed")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Elephant"); n != 0 {
		t.Errorf("no Elephants for nothing, got %d", n)
	}
}

// --- Lifecrafter's Bestiary ----------------------------------------

func TestB12LifecraftersBestiaryScriesOnUpkeepAndSellsADrawPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Lifecrafter's Bestiary", "Artifact", effectstest.B12LifecraftersBestiaryOracle, 0, 0)

	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var scry *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceScry && c.Chooser == me.ID {
			scry = c
		}
	}
	if scry == nil || len(scry.ScryCards) != 1 {
		t.Fatalf("the upkeep should queue a scry 1, got %+v", scry)
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}

	effectstest.AdvanceToMain(t, g)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a creature cast should ask the controller to pay {G}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if me.Hand.Size() != before+1 {
		t.Errorf("paying draws one: hand %d → %d", before, me.Hand.Size())
	}

	before = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an artifact is not a creature spell")
	}
	if me.Hand.Size() != before {
		t.Errorf("hand %d → %d, want unchanged", before, me.Hand.Size())
	}
}

// --- All Will Be One -----------------------------------------------

func TestB12AllWillBeOnePingsForCountersYouPlace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "All Will Be One", "Enchantment", effectstest.B12AllWillBeOneOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	before := opp.Life

	// A Duskshell Crawler's entry puts one counter on the Bear: one
	// placement, one trigger, one damage.
	effectstest.CastCatalogSpell(t, g, "Duskshell Crawler", "Creature — Insect", effectstest.B12DuskshellCrawlerOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, theirBear) {
		t.Errorf("the target is an opponent or something they control: %v", p.PickTargetCards)
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("one counter: opponent %d → %d, want -1", before, opp.Life)
	}

	// Three counters at once is one trigger for three.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(bear, "+1/+1", 3)
	})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirBear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirBear) {
		t.Fatal("three damage does not kill a 5/5")
	}
	if got := effectstest.B12Card(t, g, theirBear).DamageMarked; got != 3 {
		t.Errorf("their Bear took %d, want 3", got)
	}
}

func TestB12AllWillBeOneIgnoresOtherPlayersCountersAndRemovals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "All Will Be One", "Enchantment", effectstest.B12AllWillBeOneOracle, 0, 0)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// An opponent's resolution putting counters on their own creature.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: opp.ID})
		_ = g.AddCounterForEffect(theirBear, "+1/+1", 2)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent putting counters must not trigger it")
	}
	// Removing counters is not placing them.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(theirBear, "+1/+1", -1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a removal must not trigger it")
	}
}

// Counters put on a PLAYER fire it too (#3172): printed "on a permanent
// or player". The damage is the number that landed, a removal is not a
// placement, and an opposing placement is not yours.
func TestB12AllWillBeOneFiresForCountersOnAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "All Will Be One", "Enchantment", effectstest.B12AllWillBeOneOracle, 0, 0)
	before := opp.Life

	// You give the opponent two poison counters: one trigger for two.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddPlayerCounterByForEffect(me.ID, opp.ID, "poison", 2)
	})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("two poison counters: opponent %d -> %d, want -2", before, opp.Life)
	}

	// Energy with no named placer falls back to the resolving player.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, 3)
	})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-5 {
		t.Errorf("three energy: opponent life %d, want %d", opp.Life, before-5)
	}

	// A removal is not a placement, and an opponent's placement is not yours.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddPlayerCounterByForEffect(me.ID, opp.ID, "poison", -1)
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: opp.ID})
		_ = g.AddPlayerCounterByForEffect(opp.ID, opp.ID, "poison", 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a removal or an opponent's placement must not trigger it")
	}
}

// --- Thousand-Year Storm -------------------------------------------

func TestB12ThousandYearStormCopiesForEachEarlierSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Thousand-Year Storm", "Enchantment", effectstest.B12ThousandYearStormOracle, 0, 0)
	before := opp.Life

	// First spell of the turn: no copies.
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-3 {
		t.Fatalf("first Bolt: %d → %d, want -3", before, opp.Life)
	}
	// Second: one copy.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	// The trigger resolves first and offers new targets for the copy.
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-9 {
		t.Errorf("second Bolt plus one copy: %d → %d, want -9", before, opp.Life)
	}
	// A creature spell is not counted and does not trigger.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Third instant: two copies.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	for copies := 0; copies < 2; copies++ {
		for i := 0; i < 8 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-18 {
		t.Errorf("third Bolt plus two copies: %d → %d, want -18", before, opp.Life)
	}
}

// --- Time Wipe -----------------------------------------------------

// #1337: the return is a resolution-time choice (ChoosePermanents),
// not a cast-time target, so the prompt only appears once the spell
// resolves, offers only the caster's own creatures, and the caster
// needs no creature of their own to cast the spell at all.
func TestB12TimeWipeSavesOneCreatureAndWipesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saved := effectstest.B12Creature(g, me.ID, "Saved", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Time Wipe", "Sorcery", effectstest.B12TimeWipeOracle, nil)
	if effectstest.PendingOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("the return asked for a target; the printed text names none")
	}
	effectstest.PassPriorityAroundTable(t, g)

	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if choice == nil {
		t.Fatal("Time Wipe queued no return pick")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range choice.ChooseCards {
		offered[id] = true
	}
	if !offered[saved] || !offered[mine] {
		t.Error("both of the caster's own creatures should be offered")
	}
	if offered[theirs] {
		t.Error("only a creature you control can be returned")
	}
	if choice.ChooseMin != 1 {
		t.Error("the return is mandatory when the caster controls a creature")
	}

	if err := g.ResolveOwnPermanents(choice.ID, me.ID, []uuid.UUID{saved}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if effectstest.B12ZoneOf(g, saved) != game.ZoneHand {
		t.Error("the chosen creature returns to hand")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("every other creature is destroyed")
	}
}

// With no creature of their own, the caster is simply skipped
// (CR 608.2c) and the sweep still runs — the printed card is a plain
// five-mana Wrath in that case, not uncastable.
func TestB12TimeWipeCastableWithNoCreatureOfYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Time Wipe", "Sorcery", effectstest.B12TimeWipeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		t.Error("a caster with no creature was still asked to pick one")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the sweep did not run when the caster had no creature to save")
	}
}

// --- Cathar Commando -----------------------------------------------

func TestB12CatharCommandoSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commando := effectstest.B12Push(g, me.ID, "Cathar Commando", "Creature — Human Soldier", effectstest.B12CatharCommandoOracle, 3, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, commando), "flash") {
		t.Error("printed flash did not reach the effective abilities")
	}
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, commando, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, commando, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(commando) {
		t.Fatal("the Commando is sacrificed at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
}

// --- Altar of the Brood --------------------------------------------

func TestB12AltarOfTheBroodMillsEachOpponentPerPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Altar of the Brood", "Artifact", effectstest.B12AltarOfTheBroodOracle, 0, 0)
	before := make([]int, 0, 3)
	for _, p := range g.Seats[1:] {
		before = append(before, p.Graveyard.Size())
	}
	mine := me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Graveyard.Size() != before[i]+1 {
			t.Errorf("opponent %d: graveyard %d → %d, want +1", i+1, before[i], p.Graveyard.Size())
		}
	}
	if me.Graveyard.Size() != mine {
		t.Error("the controller does not mill")
	}
	// Two tokens at once: two triggers, two cards each.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Graveyard.Size() != before[i]+3 {
			t.Errorf("opponent %d after two tokens: graveyard %d, want %d", i+1, p.Graveyard.Size(), before[i]+3)
		}
	}
}

// --- Duskshell Crawler ---------------------------------------------

func TestB12DuskshellCrawlerCountersATargetAndGrantsTrampleToCounteredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	crawler := effectstest.CastCatalogSpell(t, g, "Duskshell Crawler", "Creature — Insect", effectstest.B12DuskshellCrawlerOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, bear, "+1/+1"); got != 1 {
		t.Fatalf("the Bear has %d counters, want 1", got)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "trample") {
		t.Error("a creature you control with a +1/+1 counter has trample")
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, crawler), "trample") {
		t.Error("the Crawler itself has no counter and no trample")
	}
	// An opponent's countered creature gets nothing from your Crawler.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirBear, "+1/+1", 1) })
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, theirBear), "trample") {
		t.Error("the grant is for creatures you control")
	}
	// The counter comes off, the trample goes with it.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", -1) })
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "trample") {
		t.Error("no counter, no trample")
	}
}
