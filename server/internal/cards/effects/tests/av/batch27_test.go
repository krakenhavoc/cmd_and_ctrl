package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b27Hands reads every seat's hand size.
func b27Hands(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Hand.Size())
	}
	return out
}

// b27HandKnownBy counts the cards in p's hand that `viewer` has seen.
func b27HandKnownBy(p *game.Player, viewer uuid.UUID) int {
	n := 0
	for _, c := range p.Hand.Cards {
		if c.KnownBy[viewer] {
			n++
		}
	}
	return n
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name.
func TestBatch27CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B27RenataOracle:                "Renata, Called to the Hunt",
		effectstest.B27UrzasBaubleOracle:           "Urza's Bauble",
		effectstest.B27GingerbreadCabinOracle:      "Gingerbread Cabin",
		effectstest.B27EarthquakeOracle:            "Earthquake",
		effectstest.B27RevengeOfRavensOracle:       "Revenge of Ravens",
		effectstest.B27ArchetypeOfAggressionOracle: "Archetype of Aggression",
		effectstest.B27HellToPayOracle:             "Hell to Pay",
		effectstest.B27FinaleOfGloryOracle:         "Finale of Glory",
		effectstest.B27HeadlessRiderOracle:         "Headless Rider",
		effectstest.B27OneWithTheMachineOracle:     "One with the Machine",
		effectstest.B27ChiefOfTheFoundryOracle:     "Chief of the Foundry",
		effectstest.B27OswaldFiddlebenderOracle:    "Oswald Fiddlebender",
		effectstest.B27MagewrightsStoneOracle:      "Magewright's Stone",
		effectstest.B27WolverineRidersOracle:       "Wolverine Riders",
		effectstest.B27LegionLoyaltyOracle:         "Legion Loyalty",
		effectstest.B27DuplicantOracle:             "Duplicant",
		effectstest.B27FarmerCottonOracle:          "Farmer Cotton",
		effectstest.B27MemorialToFollyOracle:       "Memorial to Folly",
		effectstest.B27OrcishLumberjackOracle:      "Orcish Lumberjack",
		effectstest.B27RiteOfTheDragoncallerOracle: "Rite of the Dragoncaller",
		effectstest.B27FlumphOracle:                "Flumph",
		effectstest.B27TashasHideousLaughterOracle: "Tasha's Hideous Laughter",
		effectstest.B27FoundationBreakerOracle:     "Foundation Breaker",
		effectstest.B27TheGooseMotherOracle:        "The Goose Mother",
		effectstest.B27CuriousAltisaurOracle:       "Curious Altisaur",
		effectstest.B27SanctumOfUginOracle:         "Sanctum of Ugin",
		effectstest.B27DragonMageOracle:            "Dragon Mage",
		// #1213 put the #660 discard component on
		// effects.ManaAbilityCost, so the batch's second declared
		// skip is a registered card now.
		"ba95f24d-42da-48ce-bcf1-1b7c4b3c45b5": "Skirge Familiar",
	}
	if len(want) != 28 {
		t.Fatalf("the batch registers 28 cards, the table lists %d", len(want))
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
	// The remaining declared skip stays out until its seam lands: a
	// choose-a-card-from-hand entry choice (Ugin's Labyrinth's
	// imprint).
	//
	// Skirge Familiar LEFT this list in #1213, which put the #660
	// discard component on effects.ManaAbilityCost.
	for oracle, name := range map[string]string{
		"d565cd3d-68d4-4039-9e45-7e69e31d0ffb": "Ugin's Labyrinth",
	} {
		if _, ok := effects.Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #389 but is registered — update the issue", name)
		}
	}
}

// --- the statics ---------------------------------------------------

func TestB27RenataPowerIsDevotionAndOthersEnterWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	renata := effectstest.B27Push(g, me.ID, "Renata, Called to the Hunt", "Legendary Enchantment Creature — Demigod", effectstest.B27RenataOracle, "{2}{G}{G}", 0, 3, "G")
	if got := effectstest.EffectivePower(t, g, renata); got != 2 {
		t.Errorf("Renata alone: power %d, want 2 (her own {G}{G})", got)
	}
	effectstest.B27Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", "{G}", 1, 1, "G")
	if got := effectstest.EffectivePower(t, g, renata); got != 3 {
		t.Errorf("with another {G} permanent: power %d, want 3", got)
	}
	if effectstest.CounterCount(g, renata, "+1/+1") != 0 {
		t.Error("Renata herself never gets the counter")
	}
	bear := effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 1 {
		t.Errorf("a creature entering under Renata's controller: %d +1/+1 counters, want 1", got)
	}
	// #762: a created token takes the same entry pipeline, so it gets
	// the counter exactly as a cast creature does.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.BlackZombieToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, effectstest.FindBattlefieldByName(g, "Zombie"), "+1/+1"); got != 1 {
		t.Errorf("a token: %d counters, want 1", got)
	}
	if spec, _ := effects.Lookup(effectstest.B27RenataOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the token gap is closed — the caveat must be gone")
	}
}

func TestB27ArchetypeOfAggressionGivesAndTakesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Rhino", "Creature — Rhino", 4, 4, "trample")
	if effectstest.HasEffectiveKeyword(t, g, mine, "trample") || !effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Fatal("before the Archetype: the Rhino tramples, the Bear does not")
	}
	archetype := effectstest.B21Push(g, me.ID, "Archetype of Aggression", "Enchantment Creature — Human Warrior", effectstest.B27ArchetypeOfAggressionOracle, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") || !effectstest.HasEffectiveKeyword(t, g, archetype, "trample") {
		t.Error("creatures you control have trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("creatures your opponents control lose trample")
	}
	// A printed trampler entering after the Archetype loses it too...
	late := effectstest.PushTribalCreature(g, opp.ID, "Late Rhino", "Creature — Rhino", 4, 4, "trample")
	if effectstest.HasEffectiveKeyword(t, g, late, "trample") {
		t.Error("a printed keyword is in the baseline, so a later opposing trampler loses it")
	}
	// ...and so does one whose trample is re-granted by a newer static:
	// a catalog trampler's PrintedKeywords static carries its own, later
	// timestamp. This was the declared gap until #1651 made "can't have"
	// a strip after the whole layer-6 bucket (CR 101.2).
	dreadmaw := effectstest.B21Push(g, opp.ID, "Colossal Dreadmaw", "Creature — Dinosaur", "08c7db90-c0cf-4482-b7ee-bb033e5996d2", 6, 6)
	if effectstest.HasEffectiveKeyword(t, g, dreadmaw, "trample") {
		t.Error("can't have or gain trample: a later catalog trampler loses it too")
	}
	effectstest.B27Kill(g, archetype)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, mine, "trample") || !effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("with the Archetype gone everything is as printed again")
	}
}

func TestB27ChiefOfTheFoundryPumpsOtherArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chief := effectstest.B21Push(g, me.ID, "Chief of the Foundry", "Artifact Creature — Construct", effectstest.B27ChiefOfTheFoundryOracle, 2, 3)
	myrArtifact := effectstest.B21Push(g, me.ID, "Myr", "Artifact Creature — Myr", "", 1, 1)
	bear := effectstest.B21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B21Push(g, opp.ID, "Their Myr", "Artifact Creature — Myr", "", 1, 1)
	if effectstest.EffectivePower(t, g, chief) != 2 || effectstest.EffectiveToughness(t, g, chief) != 3 {
		t.Error("the Chief does not pump itself")
	}
	if effectstest.EffectivePower(t, g, myrArtifact) != 2 || effectstest.EffectiveToughness(t, g, myrArtifact) != 2 {
		t.Error("another artifact creature you control gets +1/+1")
	}
	if effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirs) != 1 {
		t.Error("a non-artifact creature and an opponent's artifact creature are untouched")
	}
}

func TestB27DuplicantWearsTheExiledCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wizard := effectstest.B27Push(g, opp.ID, "Archmage", "Legendary Creature — Human Wizard", "", "{3}{U}{U}", 5, 6, "U")
	token := effectstest.PushToken(g, opp.ID, effects.BlackZombieToken())
	dup := effectstest.B20CastCreature(t, g, me, "Duplicant", "Artifact Creature — Shapeshifter", effectstest.B27DuplicantOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetCards, token) {
		t.Error("a token is not a legal imprint target")
	}
	effectstest.PickCard(t, g, me.ID, wizard)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ExileHas(g, wizard) {
		t.Fatal("the chosen creature is exiled")
	}
	if got := effectstest.EffectivePower(t, g, dup); got != 5 {
		t.Errorf("Duplicant power %d, want the Archmage's 5", got)
	}
	if got := effectstest.EffectiveToughness(t, g, dup); got != 6 {
		t.Errorf("Duplicant toughness %d, want the Archmage's 6", got)
	}
	subs := effectstest.EffectiveSubtypes(t, g, dup)
	for _, want := range []string{"Human", "Wizard", "Shapeshifter"} {
		if !effectstest.ContainsString(subs, want) {
			t.Errorf("Duplicant subtypes %v, want %s", subs, want)
		}
	}
	types := effectstest.EffectiveTypes(t, g, dup)
	if !effectstest.ContainsString(types, "Artifact") || !effectstest.ContainsString(types, "Creature") {
		t.Errorf("Duplicant stays an artifact creature: %v", types)
	}
	// The imprinted card leaves exile for its owner's library: the borrowed
	// body is given back at once (#3172), with no battlefield change.
	g.WithWriteLock(func() { _ = g.TuckToLibraryForEffect(wizard, false) })
	if got := effectstest.EffectivePower(t, g, dup); got != 2 {
		t.Errorf("Duplicant power %d after the card left exile, want its own 2", got)
	}
	if got := effectstest.EffectiveToughness(t, g, dup); got != 4 {
		t.Errorf("Duplicant toughness %d after the card left exile, want its own 4", got)
	}
	if effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, dup), "Wizard") {
		t.Error("Duplicant keeps the Wizard type after the card left exile")
	}
}

func TestB27DuplicantDeclinedStaysAShapeshifter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, opp.ID, "Archmage", "Creature — Human Wizard", "", "{3}{U}{U}", 5, 6, "U")
	dup := effectstest.B20CastCreature(t, g, me, "Duplicant", "Artifact Creature — Shapeshifter", effectstest.B27DuplicantOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, dup) != 2 || effectstest.EffectiveToughness(t, g, dup) != 4 {
		t.Error("nothing exiled: Duplicant is its printed 2/4")
	}
	if subs := effectstest.EffectiveSubtypes(t, g, dup); len(subs) != 1 || subs[0] != "Shapeshifter" {
		t.Errorf("subtypes %v, want just Shapeshifter", subs)
	}
}

// #1239: a Duplicant that leaves the battlefield and returns is a new
// CR 400.7 object with no memory of what the earlier incarnation
// exiled, even though its instance ID and the trigger's label are
// unchanged. Without the b27ExiledWith epoch check this test pins,
// the returning Duplicant still reads the Archmage its first
// incarnation exiled and keeps wearing it.
func TestB27DuplicantReturningHasNoMemoryOfItsPreviousExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archmage := effectstest.B27Push(g, opp.ID, "Archmage", "Legendary Creature — Human Wizard", "", "{3}{U}{U}", 5, 6, "U")
	dup := effectstest.B20CastCreature(t, g, me, "Duplicant", "Artifact Creature — Shapeshifter", effectstest.B27DuplicantOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, archmage)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ExileHas(g, archmage) {
		t.Fatal("the first incarnation exiles the Archmage")
	}
	if got := effectstest.EffectivePower(t, g, dup); got != 5 {
		t.Fatalf("first incarnation power %d, want 5 (the Archmage's)", got)
	}

	// Duplicant leaves and returns — same instance ID, a new object
	// per CR 400.7 — and this time its ETB is declined, so nothing new
	// is exiled. The Archmage is still sitting in exile the whole
	// time: it is still "exiled with" the OLD Duplicant, and must not
	// be read by the new one.
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(dup) })
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, dup, game.CastSpellParams{}); err != nil {
		t.Fatalf("recast Duplicant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.ExileHas(g, archmage) {
		t.Fatal("the Archmage is still in exile, untouched by the second incarnation")
	}
	if got := effectstest.EffectivePower(t, g, dup); got != 2 {
		t.Errorf("returning Duplicant power %d, want its printed 2 — the old incarnation's exile must not carry over", got)
	}
	if got := effectstest.EffectiveToughness(t, g, dup); got != 4 {
		t.Errorf("returning Duplicant toughness %d, want its printed 4", got)
	}
	if subs := effectstest.EffectiveSubtypes(t, g, dup); len(subs) != 1 || subs[0] != "Shapeshifter" {
		t.Errorf("returning Duplicant subtypes %v, want just Shapeshifter", subs)
	}
}

// --- the lands -----------------------------------------------------

func TestB27GingerbreadCabinEntersUntappedWithThreeForestsAndMakesFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	cabin := effectstest.B12PlayFromHand(t, g, "Gingerbread Cabin", "Land — Forest", effectstest.B27GingerbreadCabinOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B27Tapped(t, g, cabin) {
		t.Error("two other Forests: the Cabin enters tapped")
	}
	if effectstest.B27CountNamed(g, me.ID, "Food") != 0 {
		t.Error("a tapped entry makes no Food")
	}
	// The Cabin is itself a Forest, so it is the third for the next.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, 0)
	cabin2 := effectstest.B12PlayFromHand(t, g, "Gingerbread Cabin", "Land — Forest", effectstest.B27GingerbreadCabinOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B27Tapped(t, g, cabin2) {
		t.Error("three other Forests: the Cabin enters untapped")
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("an untapped entry makes a Food: %d", got)
	}
	if err := g.ActivateManaAbility(me.ID, cabin2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want G", got)
	}
}

func TestB27MemorialToFollyEntersTappedAndRaisesDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	memorial := effectstest.PlayLandFromHand(t, g, "Memorial to Folly", effectstest.B27MemorialToFollyOracle)
	if !effectstest.B27Tapped(t, g, memorial) {
		t.Error("the Memorial enters tapped")
	}
	effectstest.B22Untap(g, memorial)
	if err := g.ActivateManaAbility(me.ID, memorial, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want B", got)
	}
	me.ManaPool = nil
	effectstest.B22Untap(g, memorial)
	dead := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	effectstest.B06AddMana(me, "C", "C", "B")
	effectstest.B16Activate(t, g, me.ID, memorial, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)})
	if !me.Hand.Contains(dead) || me.Graveyard.Contains(dead) {
		t.Error("the creature card returns to hand")
	}
	if g.Battlefield.Contains(memorial) || !me.Graveyard.Contains(memorial) {
		t.Error("the Memorial is sacrificed by its own ability")
	}
	if spec, _ := effects.Lookup(effectstest.B27MemorialToFollyOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the Memorial is whole")
	}
}

func TestB27SanctumOfUginSacrificesToTutorAColorlessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sanctum := effectstest.B21Push(g, me.ID, "Sanctum of Ugin", "Land", effectstest.B27SanctumOfUginOracle, 0, 0)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Ulamog", TypeLine: "Legendary Creature — Eldrazi", ManaCost: "{10}"},
		game.Card{Name: "Wurmcoil", TypeLine: "Artifact Creature — Wurm", ManaCost: "{6}"},
		game.Card{Name: "Green Wurm", TypeLine: "Creature — Wurm", ManaCost: "{5}{G}", Colors: []string{"G"}},
		game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{7}"},
	)
	// A six-drop artifact creature is colorless but too small.
	effectstest.CastWithCost(t, g, "Steel Hellkite", "Artifact Creature — Dragon", "{6}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("mana value 6 does not trigger the Sanctum")
	}
	effectstest.CastWithCost(t, g, "Kozilek", "Legendary Creature — Eldrazi", "{10}", "")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sanctum) || !me.Graveyard.Contains(sanctum) {
		t.Fatal("the Sanctum is sacrificed")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two colorless creature cards: the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Green Wurm") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("only colorless creature cards are offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Wurmcoil")
	if !b22HandHasNamed(me, "Wurmcoil") {
		t.Error("the pick goes to hand")
	}
}

// --- the mana sources ----------------------------------------------

func TestB27OrcishLumberjackTradesAForestForThreeMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jack := effectstest.PushCatalogPermanent(g, me.ID, "Orcish Lumberjack", "Creature — Orc", effectstest.B27OrcishLumberjackOracle, false)
	if err := g.ActivateManaAbility(me.ID, jack, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("no Forest to sacrifice: the ability cannot be activated")
	}
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	plains := effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	if err := g.ActivateManaAbility(me.ID, jack, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{plains}}); err == nil {
		t.Fatal("a Plains is not a Forest")
	}
	if err := g.ActivateManaAbility(me.ID, jack, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{forest}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if g.Battlefield.Contains(forest) || !effectstest.B27Tapped(t, g, jack) {
		t.Error("the Forest is sacrificed and the Lumberjack taps")
	}
	picks := 0
	for _, color := range []string{"R", "R", "G"} {
		c := effectstest.ManaPickFor(g, me.ID)
		if c == nil {
			break
		}
		if len(c.ColorOptions) != 2 {
			t.Errorf("each slot offers R and G, got %v", c.ColorOptions)
		}
		if err := g.ResolveManaChoice(c.ID, me.ID, color); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
		picks++
	}
	if picks != 3 || effectstest.ManaPickFor(g, me.ID) != nil {
		t.Errorf("three colour picks, one per slot: %d", picks)
	}
	if got := effectstest.PoolColors(me); len(got) != 3 || got[0] != "R" || got[1] != "R" || got[2] != "G" {
		t.Errorf("pool %v, want [R R G]", got)
	}
	if spec, _ := effects.Lookup(effectstest.B27OrcishLumberjackOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the Lumberjack is whole")
	}
}

// --- the spells ----------------------------------------------------

func TestB27EarthquakeHitsGroundCreaturesAndEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bird := effectstest.PushTribalCreature(g, opp.ID, "Bird", "Creature — Bird", 2, 2, "flying")
	mine := effectstest.PushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	before := effectstest.B27Lives(g)
	effectstest.CastXSpell(t, g, "Earthquake", "Sorcery", effectstest.B27EarthquakeOracle, "{X}{R}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 without flying dies to X=2")
	}
	if !g.Battlefield.Contains(bird) || effectstest.DamageMarkedOn(g, bird) != 0 {
		t.Error("a flyer is untouched")
	}
	if !g.Battlefield.Contains(mine) || effectstest.DamageMarkedOn(g, mine) != 2 {
		t.Error("your own ground creature takes it too")
	}
	for i, l := range effectstest.B27Lives(g) {
		if l != before[i]-2 {
			t.Errorf("seat %d life %d, want %d (each player takes X)", i, l, before[i]-2)
		}
	}
}

func TestB27HellToPayPaysOutExcessInTappedTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastXSpell(t, g, "Hell to Pay", "Sorcery", effectstest.B27HellToPayOracle, "{X}{R}", 5, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear dies")
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Treasure"); got != 3 {
		t.Errorf("5 damage to a 2/2: %d Treasures, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && !c.Tapped {
			t.Error("the Treasures enter tapped")
		}
	}
	// Damage already marked lowers what was lethal.
	ox := effectstest.PushTribalCreature(g, opp.ID, "Ox", "Creature — Ox", 4, 4)
	effectstest.B27Damage(g, me.ID, ox, 3)
	effectstest.CastXSpell(t, g, "Hell to Pay", "Sorcery", effectstest.B27HellToPayOracle, "{X}{R}", 2, effectstest.CardRefs(ox))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Treasure"); got != 4 {
		t.Errorf("2 damage to a 4/4 with 3 marked: one more Treasure, got %d total", got)
	}
	// No excess, no Treasure.
	wall := effectstest.PushTribalCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	effectstest.CastXSpell(t, g, "Hell to Pay", "Sorcery", effectstest.B27HellToPayOracle, "{X}{R}", 3, effectstest.CardRefs(wall))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Treasure"); got != 4 {
		t.Errorf("3 damage to a 0/6 is no excess, got %d total", got)
	}
}

func TestB27FinaleOfGloryMakesSoldiersAndAngelsAtTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastXSpell(t, g, "Finale of Glory", "Sorcery", effectstest.B27FinaleOfGloryOracle, "{X}{W}{W}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B27CountNamed(g, me.ID, "Soldier") != 3 || effectstest.B27CountNamed(g, me.ID, "Angel") != 0 {
		t.Errorf("X=3: %d Soldiers and %d Angels, want 3 and 0", effectstest.B27CountNamed(g, me.ID, "Soldier"), effectstest.B27CountNamed(g, me.ID, "Angel"))
	}
	effectstest.CastXSpell(t, g, "Finale of Glory", "Sorcery", effectstest.B27FinaleOfGloryOracle, "{X}{W}{W}", 10, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B27CountNamed(g, me.ID, "Soldier") != 13 || effectstest.B27CountNamed(g, me.ID, "Angel") != 10 {
		t.Errorf("X=10: %d Soldiers and %d Angels, want 13 and 10", effectstest.B27CountNamed(g, me.ID, "Soldier"), effectstest.B27CountNamed(g, me.ID, "Angel"))
	}
	for _, c := range g.Battlefield.Cards {
		switch c.Name {
		case "Soldier":
			if c.Power != 2 || !effectstest.ContainsString(c.Keywords, "vigilance") || !effectstest.ContainsString(c.Colors, "W") {
				t.Errorf("Soldier %+v, want a 2/2 white vigilance token", c)
			}
		case "Angel":
			if c.Power != 4 || !effectstest.ContainsString(c.Keywords, "flying") || !effectstest.ContainsString(c.Keywords, "vigilance") {
				t.Errorf("Angel %+v, want a 4/4 flying vigilance token", c)
			}
		}
	}
}

func TestB27OneWithTheMachineDrawsTheGreatestArtifactManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "One with the Machine", "Sorcery", effectstest.B27OneWithTheMachineOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("no artifacts: nothing drawn")
	}
	effectstest.B27Push(g, me.ID, "Signet", "Artifact", "", "{2}", 0, 0)
	effectstest.B27Push(g, me.ID, "Wurmcoil", "Artifact Creature — Wurm", "", "{6}", 6, 6)
	effectstest.B27Push(g, opp.ID, "Their Colossus", "Artifact Creature — Golem", "", "{11}", 11, 11)
	hand = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "One with the Machine", "Sorcery", effectstest.B27OneWithTheMachineOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("drew %d, want 6 (your greatest artifact, not the table's)", got)
	}
}

func TestB27FarmerCottonMakesXHalflingsAndXFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastXSpell(t, g, "Farmer Cotton", "Legendary Creature — Halfling Peasant", effectstest.B27FarmerCottonOracle, "{X}{G}{W}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B27CountNamed(g, me.ID, "Halfling") != 2 || effectstest.B27CountNamed(g, me.ID, "Food") != 2 {
		t.Errorf("X=2: %d Halflings and %d Food, want 2 and 2", effectstest.B27CountNamed(g, me.ID, "Halfling"), effectstest.B27CountNamed(g, me.ID, "Food"))
	}
	if effectstest.B27CountNamed(g, me.ID, "Farmer Cotton") != 1 {
		t.Error("Cotton himself is on the battlefield")
	}
	// #1357: the ETB tokens now come from a real trigger reading
	// CastX(), closing the last declared gap.
	if spec, _ := effects.Lookup(effectstest.B27FarmerCottonOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("no caveat should remain now that the ETB trigger reads CastX()")
	}
}

func TestB27TashasHideousLaughterExilesToTwentyManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp1, opp2, opp3 := g.Seats[1], g.Seats[2], g.Seats[3]
	// opp1: 6 + 6 + 6 + 6 = 24 on the fourth card; the fifth stays.
	// (seedSearchLibrary pushes to the bottom, so the first card
	// listed is the top.)
	effectstest.SeedSearchLibrary(opp1,
		game.Card{Name: "Top", TypeLine: "Creature", ManaCost: "{6}"},
		game.Card{Name: "Second", TypeLine: "Creature", ManaCost: "{6}"},
		game.Card{Name: "Third", TypeLine: "Creature", ManaCost: "{6}"},
		game.Card{Name: "Fourth", TypeLine: "Creature", ManaCost: "{6}"},
		game.Card{Name: "Deep", TypeLine: "Creature", ManaCost: "{6}"},
	)
	// opp2: lands and cheap spells never reach 20 — the whole
	// library goes, and they do not lose for it.
	effectstest.SeedSearchLibrary(opp2,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"},
	)
	// opp3: one twenty-drop on top.
	effectstest.SeedSearchLibrary(opp3,
		game.Card{Name: "Emrakul", TypeLine: "Creature", ManaCost: "{20}"},
		game.Card{Name: "Under", TypeLine: "Creature", ManaCost: "{1}"},
	)
	effectstest.CastCatalogSpell(t, g, "Tasha's Hideous Laughter", "Sorcery", effectstest.B27TashasHideousLaughterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp1.Library.Size() != 1 || opp1.Library.Cards[0].Name != "Deep" {
		t.Errorf("opp1 exiles until 20: %d cards left, want the one deepest", opp1.Library.Size())
	}
	if opp2.Library.Size() != 0 || opp2.AttemptedEmptyDraw {
		t.Error("opp2 exiles the whole library and does not lose for it")
	}
	if opp3.Library.Size() != 1 {
		t.Error("opp3 exiles the twenty-drop and stops")
	}
	if g.Seats[0].Library.Size() == 0 {
		t.Error("the caster's library is untouched")
	}
	for _, name := range []string{"Top", "Second", "Third", "Fourth", "Forest", "Bolt", "Emrakul"} {
		found := false
		for _, c := range g.Exile.Cards {
			if c.Name == name {
				found = true
			}
		}
		if !found {
			t.Errorf("%s should be in exile", name)
		}
	}
	if len(opp1.Graveyard.Cards)+len(opp2.Graveyard.Cards)+len(opp3.Graveyard.Cards) != 0 {
		t.Error("exile, not mill: nothing goes to a graveyard")
	}
}

// --- the activations -----------------------------------------------

func TestB27UrzasBaubleLooksAtACardAndDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bauble := effectstest.CastCatalogSpell(t, g, "Urza's Bauble", "Artifact", effectstest.B27UrzasBaubleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bauble) {
		t.Fatal("the Bauble resolves to the battlefield")
	}
	if b27HandKnownBy(opp, me.ID) != 0 {
		t.Fatal("before the activation the opponent's hand is hidden")
	}
	effectstest.B16Activate(t, g, me.ID, bauble, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if g.Battlefield.Contains(bauble) || !me.Graveyard.Contains(bauble) {
		t.Error("the Bauble is sacrificed by its own ability")
	}
	if got := b27HandKnownBy(opp, me.ID); got != 1 {
		t.Errorf("the activator sees exactly one card in the hand: %d", got)
	}
	if b27HandKnownBy(opp, g.Seats[2].ID) != 0 {
		t.Error("nobody else sees anything")
	}
	if n := len(g.DelayedTriggers); n != 1 {
		t.Fatalf("delayed triggers = %d, want the upkeep draw", n)
	}
	hand := me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if len(g.PendingTriggers)+len(g.StackMeta) == 0 {
		t.Fatal("the draw is on the stack at the next upkeep, whoever's it is")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, got)
	}
}

func TestB27OswaldFiddlebenderPodsArtifactsUpOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oswald := effectstest.PushCatalogPermanent(g, me.ID, "Oswald Fiddlebender", "Legendary Creature — Gnome Artificer", effectstest.B27OswaldFiddlebenderOracle, false)
	signet := effectstest.B27Push(g, me.ID, "Signet", "Artifact", "", "{2}", 0, 0)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Three Rock", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Other Three", TypeLine: "Artifact Creature — Construct", ManaCost: "{3}"},
		game.Card{Name: "Four Rock", TypeLine: "Artifact", ManaCost: "{4}"},
		game.Card{Name: "Three Elf", TypeLine: "Creature — Elf", ManaCost: "{2}{G}"},
	)
	spec, _ := effects.Lookup(effectstest.B27OswaldFiddlebenderOracle)
	if len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed {
		t.Fatal("one ability, sorcery speed")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W")
	effectstest.B16Activate(t, g, me.ID, oswald, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{signet}})
	if g.Battlefield.Contains(signet) || !effectstest.B27Tapped(t, g, oswald) {
		t.Fatal("the artifact is sacrificed and Oswald taps")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two artifacts at mana value 3: the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Four Rock") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Three Elf") != uuid.Nil {
		t.Error("only artifact cards with mana value exactly 3 are offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Other Three")
	if effectstest.FindBattlefieldByName(g, "Other Three") == uuid.Nil {
		t.Error("the pick enters the battlefield")
	}
}

func TestB27MagewrightsStoneUntapsOnlyTapAbilityCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stone := effectstest.PushCatalogPermanent(g, me.ID, "Magewright's Stone", "Artifact", effectstest.B27MagewrightsStoneOracle, false)
	elves := effectstest.PushCatalogPermanent(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", effectstest.FyndhornElvesOracle, false)
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Tap(g, elves)
	effectstest.B16Tap(g, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a creature with no tap ability is not a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(elves)})
	if effectstest.B27Tapped(t, g, elves) {
		t.Error("the mana Elf is untapped")
	}
	if !effectstest.B27Tapped(t, g, stone) {
		t.Error("the Stone taps to do it")
	}
	if spec, _ := effects.Lookup(effectstest.B27MagewrightsStoneOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the printed target set is reachable now (#3860)")
	}
}

// #3860: the clause is about what the creature HAS, so a creature the
// catalog has no entry for is a legal target when its printed text has
// a "{T}: ..." ability. The importer stamps that from the Scryfall
// text (game.PrintsTapAbilityText); a fixture sets the stamp directly.
func TestMagewrightsStoneUntapsAnUncataloguedCreatureWithAPrintedTapAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stone := effectstest.PushCatalogPermanent(g, me.ID, "Magewright's Stone", "Artifact", effectstest.B27MagewrightsStoneOracle, false)
	printed := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Uncatalogued Mystic", TypeLine: "Creature — Human Wizard",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, PrintsTapAbility: true,
	})
	plain := effectstest.PushTribalCreature(g, me.ID, "Uncatalogued Bear", "Creature — Bear", 2, 2)
	effectstest.B16Tap(g, printed)
	effectstest.B16Tap(g, plain)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(plain)}); err == nil {
		t.Fatal("a creature whose text has no {T} ability is not a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(printed)})
	if effectstest.B27Tapped(t, g, printed) {
		t.Error("the uncatalogued creature with a printed {T} ability is untapped")
	}
}

// A multi-faced card is judged on the face that is up (CR 712.8): the
// front has no tap ability, the back does.
func TestMagewrightsStoneJudgesAMultiFacedCreatureOnTheFaceThatIsUp(t *testing.T) {
	front := game.Face{Name: "Front", TypeLine: "Creature — Human", OracleText: "Vigilance"}
	back := game.Face{Name: "Back", TypeLine: "Creature — Werewolf", OracleText: "{T}: Add {G}."}
	card := game.Card{Name: "Front", TypeLine: "Creature — Human", Layout: "transform", Faces: []game.Face{front, back}}
	card.SetFace(0)
	if card.PrintedTapAbility() {
		t.Error("the front face prints no tap ability")
	}
	card.SetFace(1)
	if !card.PrintedTapAbility() {
		t.Error("the back face prints one")
	}
}

// A copy takes the copied card's text (CR 707.2), so the stamp is a
// copiable value.
func TestPrintsTapAbilityIsACopiableValue(t *testing.T) {
	src := game.Card{Name: "Mystic", TypeLine: "Creature — Wizard", PrintsTapAbility: true}
	if v := game.CopiableValuesOf(src); !v.PrintsTapAbility {
		t.Error("a copy of a creature with a printed tap ability keeps it")
	}
}

// --- the triggers --------------------------------------------------

func TestB27RevengeOfRavensDrainsPerAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Revenge of Ravens", "Enchantment", effectstest.B27RevengeOfRavensOracle, 0, 0)
	a := effectstest.PushVanillaCreature(g, opp.ID, "Raider A", 1, 1)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Raider B", 1, 1)
	c := effectstest.PushVanillaCreature(g, opp.ID, "Raider C", 1, 1)
	other := g.Seats[2]
	effectstest.AdvanceToMainOf(t, g, 1)
	meBefore, oppBefore, otherBefore := me.Life, opp.Life, other.Life
	// One declaration (CR 508.1, #3621): all three attackers are staged
	// before it locks in.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ id, at uuid.UUID }{{a, me.ID}, {b, me.ID}, {c, other.ID}} {
		if err := g.DeclareAttacker(d.id, d.at); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppBefore-2 {
		t.Errorf("attacker's controller lost %d, want 2 (one per creature attacking you)", oppBefore-opp.Life)
	}
	if me.Life != meBefore+2 {
		t.Errorf("you gained %d, want 2", me.Life-meBefore)
	}
	if other.Life != otherBefore {
		t.Error("a creature attacking someone else does not trigger it")
	}
}

// #4864, CR 508.3a: "attacks you or a planeswalker you control". A
// creature attacking your planeswalker drains; one attacking a battle
// you protect does not, though you are its defending player.
func TestB27RevengeOfRavensCountsYourPlaneswalkerButNotYourBattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B21Push(g, me.ID, "Revenge of Ravens", "Enchantment", effectstest.B27RevengeOfRavensOracle, 0, 0)
	walker := effectstest.P1g2Planeswalker(g, me.ID, 5)
	battle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
		Owner: other.ID, Controller: other.ID, ProtectorPlayerID: me.ID,
		Counters: map[string]int{game.CounterDefense: 5},
	})
	a := effectstest.PushVanillaCreature(g, opp.ID, "Raider A", 1, 1)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Raider B", 1, 1)
	effectstest.AdvanceToMainOf(t, g, 1)
	meBefore, oppBefore := me.Life, opp.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ id, at uuid.UUID }{{a, walker}, {b, battle}} {
		if err := g.DeclareAttacker(d.id, d.at); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppBefore-1 {
		t.Errorf("attacker's controller lost %d, want 1 (the planeswalker attack only)", oppBefore-opp.Life)
	}
	if me.Life != meBefore+1 {
		t.Errorf("you gained %d, want 1", me.Life-meBefore)
	}
}

func TestB27HeadlessRiderSpawnsForNontokenZombieDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := effectstest.B21Push(g, me.ID, "Headless Rider", "Creature — Zombie", effectstest.B27HeadlessRiderOracle, 3, 1)
	zombie := effectstest.PushTribalCreature(g, me.ID, "Gravecrawler", "Creature — Zombie", 2, 1)
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B27Kill(g, zombie)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Fatalf("a nontoken Zombie you control died: %d tokens, want 1", got)
	}
	effectstest.B27Kill(g, theirs)
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("an opponent's Zombie and a non-Zombie: still %d tokens, want 1", got)
	}
	token := effectstest.FindBattlefieldByName(g, "Zombie")
	effectstest.B27Kill(g, token)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Zombie"); got != 0 {
		t.Errorf("a token Zombie dying makes nothing: %d", got)
	}
	effectstest.B27Kill(g, rider)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("the Rider's own death makes one: %d", got)
	}
}

func TestB27WolverineRidersSpawnEachUpkeepAndGainOnElves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Wolverine Riders", "Creature — Elf Warrior", effectstest.B27WolverineRidersOracle, 4, 4)
	life := me.Life
	// The opponent's upkeep: a token, and the Elf entering gains 1.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Elf Warrior"); got != 1 {
		t.Fatalf("each upkeep, yours or not: %d tokens, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("the token is an Elf with toughness 1: life %d → %d", life, me.Life)
	}
	// Seats 2 and 3 and then your own: three more.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Elf Warrior"); got != 4 {
		t.Errorf("after a full round of upkeeps: %d tokens, want 4", got)
	}
	life = me.Life
	effectstest.B20CastCreature(t, g, me, "Elvish Archdruid", "Creature — Elf Druid", "", 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("another Elf entering: gain its toughness, life %d → %d", life, me.Life)
	}
	life = me.Life
	effectstest.B20CastCreature(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Error("a non-Elf gains nothing")
	}
}

func TestB27RiteOfTheDragoncallerMakesADragonPerInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Rite of the Dragoncaller", "Enchantment", effectstest.B27RiteOfTheDragoncallerOracle, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Dragon"); got != 1 {
		t.Fatalf("an instant: %d Dragons, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Dragon"); got != 1 {
		t.Errorf("a creature spell: still %d Dragons", got)
	}
	dragon := effectstest.FindBattlefieldByName(g, "Dragon")
	if effectstest.EffectivePower(t, g, dragon) != 5 || !effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("the Dragon is a 5/5 with flying")
	}
}

func TestB27FlumphDrawsForYouAndATargetOpponentWhenDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flumph := effectstest.B21Push(g, me.ID, "Flumph", "Creature — Jellyfish", effectstest.B27FlumphOracle, 0, 4)
	if !effectstest.HasEffectiveKeyword(t, g, flumph, "defender") || !effectstest.HasEffectiveKeyword(t, g, flumph, "flying") {
		t.Error("defender and flying are printed")
	}
	before := b27Hands(g)
	effectstest.B27Damage(g, opp.ID, flumph, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	after := b27Hands(g)
	if after[0] != before[0]+1 || after[1] != before[1]+1 {
		t.Errorf("hands %v → %v, want you and the chosen opponent +1", before, after)
	}
	if after[2] != before[2] || after[3] != before[3] {
		t.Error("the other opponents draw nothing")
	}
}

func TestB27CuriousAltisaurDrawsWhenADinosaurConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	altisaur := effectstest.B21Push(g, me.ID, "Curious Altisaur", "Creature — Dinosaur", effectstest.B27CuriousAltisaurOracle, 2, 5)
	raptor := effectstest.PushTribalCreature(g, me.ID, "Raptor", "Creature — Dinosaur", 3, 3)
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, altisaur, "reach") || !effectstest.HasEffectiveKeyword(t, g, altisaur, "vigilance") {
		t.Error("reach and vigilance are printed")
	}
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, raptor, opp.ID, 3)
	effectstest.DealCombatDamageToPlayer(g, altisaur, opp.ID, 2)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (one per Dinosaur, none for the Bear)", got)
	}
}

func TestB27DragonMageWheelsTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B21Push(g, me.ID, "Dragon Mage", "Creature — Dragon Wizard", effectstest.B27DragonMageOracle, 5, 5)
	for _, p := range g.Seats {
		effectstest.FillHandTo(t, g, p, 3)
	}
	effectstest.DealCombatDamageToPlayer(g, mage, opp.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != 7 {
			t.Errorf("seat %d hand %d, want 7", i, p.Hand.Size())
		}
		if p.Graveyard.Size() < 3 {
			t.Errorf("seat %d discarded %d, want the whole hand of 3", i, p.Graveyard.Size())
		}
	}
	// Noncombat damage is not the trigger.
	hands := b27Hands(g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mage, opp.ID, 5) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, h := range b27Hands(g) {
		if h != hands[i] {
			t.Error("noncombat damage does not wheel")
		}
	}
}

func TestB27FoundationBreakerDestroysOnEntryAndEvokes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B21Push(g, opp.ID, "Mana Rock", "Artifact", "", 0, 0)
	bear := effectstest.PushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	breaker := effectstest.CastAndResolveCreature(t, g, "Foundation Breaker", "Creature — Elemental", effectstest.B27FoundationBreakerOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetCards, bear) {
		t.Error("a creature is not an artifact or enchantment")
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the chosen artifact is destroyed")
	}
	if !g.Battlefield.Contains(breaker) {
		t.Error("hard-cast, the Breaker stays")
	}
	// Evoked: the destroy still happens and the Breaker is sacrificed.
	curse := effectstest.B21Push(g, opp.ID, "Curse", "Enchantment", "", 0, 0)
	evoked := effectstest.CastWithAltCost(t, g, "Foundation Breaker", "Creature — Elemental", effectstest.B27FoundationBreakerOracle, "evoke")
	effectstest.B27Settle(t, g, me.ID, curse)
	if g.Battlefield.Contains(curse) {
		t.Error("evoked, the enchantment is still destroyed")
	}
	if g.Battlefield.Contains(evoked) || !me.Graveyard.Contains(evoked) {
		t.Error("evoked, the Breaker is sacrificed")
	}
}

func TestB27TheGooseMotherEntersWithCountersFoodAndEatsToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goose := effectstest.CastXSpell(t, g, "The Goose Mother", "Legendary Creature — Bird Hydra", effectstest.B27TheGooseMotherOracle, "{X}{G}{U}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, goose, "+1/+1"); got != 3 {
		t.Errorf("X=3: %d +1/+1 counters, want 3", got)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("half of 3 rounded up: %d Food, want 2", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, goose, "flying") {
		t.Error("flying is printed")
	}
	// Attack: the prompt, the Food pick, the sacrifice, the draw.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == goose {
			g.Battlefield.Cards[i].SummonedThisTurn = false
			g.Battlefield.Cards[i].EnteredBattlefieldAt = 0
		}
	}
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, goose)
	effectstest.PassPriorityAroundTable(t, g)
	food := effectstest.FindBattlefieldByName(g, "Food")
	if prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, food) {
		t.Fatal("the Food is chosen as the trigger resolves, with a floor of zero for the \"may\"")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, food)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(food) || effectstest.B27CountNamed(g, me.ID, "Food") != 1 {
		t.Error("the chosen Food is sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, got)
	}
}

func TestB27TheGooseMotherWithNoFoodIsSilent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goose := effectstest.B21Push(g, me.ID, "The Goose Mother", "Legendary Creature — Bird Hydra", effectstest.B27TheGooseMotherOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, goose)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("no Food: no prompt at all (CR 603.3d)")
	}
}

func TestB27LegionLoyaltyMyriadCopiesAttackTheOtherOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.B21Push(g, me.ID, "Legion Loyalty", "Enchantment", effectstest.B27LegionLoyaltyOracle, 0, 0)
	knight := effectstest.PushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3, "vigilance")
	effectstest.DeclareAttack(t, g, opp1.ID, knight)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	copies := 0
	attacking := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Knight" || c.InstanceID == knight {
			continue
		}
		copies++
		if !c.Tapped || !effects.IsToken(c) || c.Controller != me.ID {
			t.Errorf("copy %+v should be a tapped token you control", c.Name)
		}
		attacking[c.AttackingTarget]++
	}
	if copies != 2 {
		t.Fatalf("%d copies, want one per other opponent (2)", copies)
	}
	if attacking[opp2.ID] != 1 || attacking[opp3.ID] != 1 || attacking[opp1.ID] != 0 {
		t.Errorf("copies attack %v, want one each at the two non-defending opponents", attacking)
	}
	if n := len(g.DelayedTriggers); n != 1 {
		t.Fatalf("delayed triggers = %d, want the end-of-combat exile", n)
	}
	l1, l2, l3 := opp1.Life, opp2.Life, opp3.Life
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if opp1.Life != l1-3 || opp2.Life != l2-3 || opp3.Life != l3-3 {
		t.Errorf("every opponent takes 3: %d %d %d → %d %d %d", l1, l2, l3, opp1.Life, opp2.Life, opp3.Life)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Knight"); got != 1 {
		t.Errorf("at end of combat the copies are exiled: %d Knights left, want the original", got)
	}
	if !g.Battlefield.Contains(knight) {
		t.Error("the original stays")
	}
}

func TestB27LegionLoyaltyIsSilentInADuel(t *testing.T) {
	g := effectstest.NewDuelCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Legion Loyalty", "Enchantment", effectstest.B27LegionLoyaltyOracle, 0, 0)
	knight := effectstest.PushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, knight)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Error("no other opponent: myriad has nothing to ask")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B27CountNamed(g, me.ID, "Knight") != 1 {
		t.Error("no copies")
	}
}
