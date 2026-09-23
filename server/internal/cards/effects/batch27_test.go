package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch27_test.go — card-level coverage for the card-coverage
// roadmap's batch 27 (#389, `edhrec_rank` 2839–2939): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, damage, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b27-prefixed.

const (
	b27RenataOracle                = "6761b077-7a89-42c9-93ab-675fe4231564"
	b27UrzasBaubleOracle           = "17dbbca3-ac1c-4d4e-9618-3e66ac3ccd24"
	b27GingerbreadCabinOracle      = "fa98c367-0312-49c6-abef-72e5ead4cc7d"
	b27EarthquakeOracle            = "9a40614b-50a3-422c-849e-53c8b7d3d204"
	b27RevengeOfRavensOracle       = "dd96f145-73eb-4a6d-bcfd-5d6313fac1f9"
	b27ArchetypeOfAggressionOracle = "263408e6-b315-4af5-8cb8-3fd1aa88e48c"
	b27HellToPayOracle             = "7dc7d90d-5979-4c4a-9183-392e0e0b878a"
	b27FinaleOfGloryOracle         = "68273453-1059-4eab-8c27-5d96f998f3b1"
	b27HeadlessRiderOracle         = "d4fdacd7-3101-44e2-a880-dde7326137a4"
	b27OneWithTheMachineOracle     = "757eadac-dd29-4a7f-8683-5bc823168326"
	b27ChiefOfTheFoundryOracle     = "4fdfa41a-75a5-4b36-8c9e-083e26d137e2"
	b27OswaldFiddlebenderOracle    = "dbc9ea19-cf68-41d3-88a8-b5ab8df75c5a"
	b27MagewrightsStoneOracle      = "9e500094-c430-4c10-89cc-97f4f18b08bc"
	b27WolverineRidersOracle       = "09070db6-01f6-4a8a-b167-a7825e6959f4"
	b27LegionLoyaltyOracle         = "24e39d26-6a2b-461e-b194-bda575bcf898"
	b27DuplicantOracle             = "ea86abfa-6cab-4ef0-8463-34136fc25b59"
	b27FarmerCottonOracle          = "11f1d2b7-0c2b-40df-bf90-3c55b23449af"
	b27MemorialToFollyOracle       = "2bc38f14-0314-4351-8138-e2b8bf041404"
	b27OrcishLumberjackOracle      = "383e3fe4-8558-4561-8632-6eadb5d5963c"
	b27RiteOfTheDragoncallerOracle = "5097f4e6-50af-4641-909f-db44abf0ce32"
	b27FlumphOracle                = "bc328acc-8521-4581-88c3-99a7cb2c1cdb"
	b27TashasHideousLaughterOracle = "e352f5b9-6406-4914-bc79-f24608be6bc9"
	b27FoundationBreakerOracle     = "8b7a3613-f1bd-4262-9b11-b631167c2c2d"
	b27TheGooseMotherOracle        = "de595f1b-3f7d-45e0-a31b-ed23e5d1ee48"
	b27CuriousAltisaurOracle       = "80686908-abb5-4728-a5f6-71baca27f467"
	b27SanctumOfUginOracle         = "72cb5dcd-9b24-435c-921a-3766108374c4"
	b27DragonMageOracle            = "eab71e4e-27c3-4c41-b95f-259c1d14b97a"
)

// b27Lives reads every seat's life total.
func b27Lives(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Life)
	}
	return out
}

// b27Hands reads every seat's hand size.
func b27Hands(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Hand.Size())
	}
	return out
}

// b27Push seeds a permanent with a mana cost, colours and P/T on the
// battlefield with a layer timestamp.
func b27Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Colors: colors, Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b27Kill destroys a permanent through the effect API and leaves
// whatever it triggered pending.
func b27Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
}

// b27Damage deals noncombat damage to a creature through the effect
// API and leaves whatever it triggered pending.
func b27Damage(g *game.Game, source, target uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(source, target, n) })
}

// b27CountNamed counts battlefield permanents `controller` controls
// with the given name.
func b27CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// b27Tapped reads a battlefield card's tapped state.
func b27Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
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

// b27Settle answers every prompt addressed to `chooser` until the
// stack is empty: trigger-order prompts in the offered order,
// yes/no trigger prompts with yes, and pick_target prompts with
// `pick` (a card) — then passes priority. For the cards whose entry
// queues an evoke sacrifice beside an optional targeted trigger.
func b27Settle(t *testing.T, g *game.Game, chooser, pick uuid.UUID) {
	t.Helper()
	for i := 0; i < 40; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c == nil || c.Chooser != chooser {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerOrder:
				answerTriggerOrderInOfferedOrder(t, g)
			case game.PendingChoiceTriggerPrompt:
				if err := g.ResolveTriggerPrompt(c.ID, chooser, true); err != nil {
					t.Fatalf("ResolveTriggerPrompt: %v", err)
				}
			case game.PendingChoicePickTarget:
				if err := g.ResolvePickTarget(c.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: pick}); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
			default:
				continue
			}
			answered = true
			break
		}
		if answered {
			continue
		}
		if stackFullyEmpty(g) && len(g.PendingChoices) == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name.
func TestBatch27CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b27RenataOracle:                "Renata, Called to the Hunt",
		b27UrzasBaubleOracle:           "Urza's Bauble",
		b27GingerbreadCabinOracle:      "Gingerbread Cabin",
		b27EarthquakeOracle:            "Earthquake",
		b27RevengeOfRavensOracle:       "Revenge of Ravens",
		b27ArchetypeOfAggressionOracle: "Archetype of Aggression",
		b27HellToPayOracle:             "Hell to Pay",
		b27FinaleOfGloryOracle:         "Finale of Glory",
		b27HeadlessRiderOracle:         "Headless Rider",
		b27OneWithTheMachineOracle:     "One with the Machine",
		b27ChiefOfTheFoundryOracle:     "Chief of the Foundry",
		b27OswaldFiddlebenderOracle:    "Oswald Fiddlebender",
		b27MagewrightsStoneOracle:      "Magewright's Stone",
		b27WolverineRidersOracle:       "Wolverine Riders",
		b27LegionLoyaltyOracle:         "Legion Loyalty",
		b27DuplicantOracle:             "Duplicant",
		b27FarmerCottonOracle:          "Farmer Cotton",
		b27MemorialToFollyOracle:       "Memorial to Folly",
		b27OrcishLumberjackOracle:      "Orcish Lumberjack",
		b27RiteOfTheDragoncallerOracle: "Rite of the Dragoncaller",
		b27FlumphOracle:                "Flumph",
		b27TashasHideousLaughterOracle: "Tasha's Hideous Laughter",
		b27FoundationBreakerOracle:     "Foundation Breaker",
		b27TheGooseMotherOracle:        "The Goose Mother",
		b27CuriousAltisaurOracle:       "Curious Altisaur",
		b27SanctumOfUginOracle:         "Sanctum of Ugin",
		b27DragonMageOracle:            "Dragon Mage",
		// #1213 put the #660 discard component on
		// effects.ManaAbilityCost, so the batch's second declared
		// skip is a registered card now.
		"ba95f24d-42da-48ce-bcf1-1b7c4b3c45b5": "Skirge Familiar",
	}
	if len(want) != 28 {
		t.Fatalf("the batch registers 28 cards, the table lists %d", len(want))
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
	// The remaining declared skip stays out until its seam lands: a
	// choose-a-card-from-hand entry choice (Ugin's Labyrinth's
	// imprint).
	//
	// Skirge Familiar LEFT this list in #1213, which put the #660
	// discard component on effects.ManaAbilityCost.
	for oracle, name := range map[string]string{
		"d565cd3d-68d4-4039-9e45-7e69e31d0ffb": "Ugin's Labyrinth",
	} {
		if _, ok := Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #389 but is registered — update the issue", name)
		}
	}
}

// --- the statics ---------------------------------------------------

func TestB27RenataPowerIsDevotionAndOthersEnterWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	renata := b27Push(g, me.ID, "Renata, Called to the Hunt", "Legendary Enchantment Creature — Demigod", b27RenataOracle, "{2}{G}{G}", 0, 3, "G")
	if got := effectivePower(t, g, renata); got != 2 {
		t.Errorf("Renata alone: power %d, want 2 (her own {G}{G})", got)
	}
	b27Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", "{G}", 1, 1, "G")
	if got := effectivePower(t, g, renata); got != 3 {
		t.Errorf("with another {G} permanent: power %d, want 3", got)
	}
	if counterCount(g, renata, "+1/+1") != 0 {
		t.Error("Renata herself never gets the counter")
	}
	bear := castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 1 {
		t.Errorf("a creature entering under Renata's controller: %d +1/+1 counters, want 1", got)
	}
	// #762: a created token takes the same entry pipeline, so it gets
	// the counter exactly as a cast creature does.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, BlackZombieToken(), 1) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, findBattlefieldByName(g, "Zombie"), "+1/+1"); got != 1 {
		t.Errorf("a token: %d counters, want 1", got)
	}
	if spec, _ := Lookup(b27RenataOracle); spec.Completeness != CompletenessFull {
		t.Error("the token gap is closed — the caveat must be gone")
	}
}

func TestB27ArchetypeOfAggressionGivesAndTakesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := pushTribalCreature(g, opp.ID, "Rhino", "Creature — Rhino", 4, 4, "trample")
	if hasEffectiveKeyword(t, g, mine, "trample") || !hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Fatal("before the Archetype: the Rhino tramples, the Bear does not")
	}
	archetype := b21Push(g, me.ID, "Archetype of Aggression", "Enchantment Creature — Human Warrior", b27ArchetypeOfAggressionOracle, 3, 2)
	if !hasEffectiveKeyword(t, g, mine, "trample") || !hasEffectiveKeyword(t, g, archetype, "trample") {
		t.Error("creatures you control have trample")
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("creatures your opponents control lose trample")
	}
	// A printed trampler entering after the Archetype loses it too...
	late := pushTribalCreature(g, opp.ID, "Late Rhino", "Creature — Rhino", 4, 4, "trample")
	if hasEffectiveKeyword(t, g, late, "trample") {
		t.Error("a printed keyword is in the baseline, so a later opposing trampler loses it")
	}
	// ...unless the trample is re-granted by a newer static — the
	// declared gap, pinned here: a catalog trampler's PrintedKeywords
	// static carries its own, later timestamp.
	dreadmaw := b21Push(g, opp.ID, "Colossal Dreadmaw", "Creature — Dinosaur", "08c7db90-c0cf-4482-b7ee-bb033e5996d2", 6, 6)
	if !hasEffectiveKeyword(t, g, dreadmaw, "trample") {
		t.Error("the declared gap has closed — a later catalog trampler now loses trample; update the card's caveat")
	}
	b27Kill(g, archetype)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, mine, "trample") || !hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("with the Archetype gone everything is as printed again")
	}
}

func TestB27ChiefOfTheFoundryPumpsOtherArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chief := b21Push(g, me.ID, "Chief of the Foundry", "Artifact Creature — Construct", b27ChiefOfTheFoundryOracle, 2, 3)
	myrArtifact := b21Push(g, me.ID, "Myr", "Artifact Creature — Myr", "", 1, 1)
	bear := b21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := b21Push(g, opp.ID, "Their Myr", "Artifact Creature — Myr", "", 1, 1)
	if effectivePower(t, g, chief) != 2 || effectiveToughness(t, g, chief) != 3 {
		t.Error("the Chief does not pump itself")
	}
	if effectivePower(t, g, myrArtifact) != 2 || effectiveToughness(t, g, myrArtifact) != 2 {
		t.Error("another artifact creature you control gets +1/+1")
	}
	if effectivePower(t, g, bear) != 2 || effectivePower(t, g, theirs) != 1 {
		t.Error("a non-artifact creature and an opponent's artifact creature are untouched")
	}
}

func TestB27DuplicantWearsTheExiledCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wizard := b27Push(g, opp.ID, "Archmage", "Legendary Creature — Human Wizard", "", "{3}{U}{U}", 5, 6, "U")
	token := pushToken(g, opp.ID, BlackZombieToken())
	dup := b20CastCreature(t, g, me, "Duplicant", "Artifact Creature — Shapeshifter", b27DuplicantOracle, 2, 4)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, token) {
		t.Error("a token is not a legal imprint target")
	}
	pickCard(t, g, me.ID, wizard)
	passPriorityAroundTable(t, g)
	if !exileHas(g, wizard) {
		t.Fatal("the chosen creature is exiled")
	}
	if got := effectivePower(t, g, dup); got != 5 {
		t.Errorf("Duplicant power %d, want the Archmage's 5", got)
	}
	if got := effectiveToughness(t, g, dup); got != 6 {
		t.Errorf("Duplicant toughness %d, want the Archmage's 6", got)
	}
	subs := effectiveSubtypes(t, g, dup)
	for _, want := range []string{"Human", "Wizard", "Shapeshifter"} {
		if !containsString(subs, want) {
			t.Errorf("Duplicant subtypes %v, want %s", subs, want)
		}
	}
	types := effectiveTypes(t, g, dup)
	if !containsString(types, "Artifact") || !containsString(types, "Creature") {
		t.Errorf("Duplicant stays an artifact creature: %v", types)
	}
	if spec, _ := Lookup(b27DuplicantOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the stale-record corner must be declared")
	}
}

func TestB27DuplicantDeclinedStaysAShapeshifter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, opp.ID, "Archmage", "Creature — Human Wizard", "", "{3}{U}{U}", 5, 6, "U")
	dup := b20CastCreature(t, g, me, "Duplicant", "Artifact Creature — Shapeshifter", b27DuplicantOracle, 2, 4)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, dup) != 2 || effectiveToughness(t, g, dup) != 4 {
		t.Error("nothing exiled: Duplicant is its printed 2/4")
	}
	if subs := effectiveSubtypes(t, g, dup); len(subs) != 1 || subs[0] != "Shapeshifter" {
		t.Errorf("subtypes %v, want just Shapeshifter", subs)
	}
}

// --- the lands -----------------------------------------------------

func TestB27GingerbreadCabinEntersUntappedWithThreeForestsAndMakesFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	cabin := b12PlayFromHand(t, g, "Gingerbread Cabin", "Land — Forest", b27GingerbreadCabinOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if !b27Tapped(t, g, cabin) {
		t.Error("two other Forests: the Cabin enters tapped")
	}
	if b27CountNamed(g, me.ID, "Food") != 0 {
		t.Error("a tapped entry makes no Food")
	}
	// The Cabin is itself a Forest, so it is the third for the next.
	advanceToNextSeatsTurn(t, g)
	advanceToMainOf(t, g, 0)
	cabin2 := b12PlayFromHand(t, g, "Gingerbread Cabin", "Land — Forest", b27GingerbreadCabinOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if b27Tapped(t, g, cabin2) {
		t.Error("three other Forests: the Cabin enters untapped")
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("an untapped entry makes a Food: %d", got)
	}
	if err := g.ActivateManaAbility(me.ID, cabin2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want G", got)
	}
}

func TestB27MemorialToFollyEntersTappedAndRaisesDead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	memorial := playLandFromHand(t, g, "Memorial to Folly", b27MemorialToFollyOracle)
	if !b27Tapped(t, g, memorial) {
		t.Error("the Memorial enters tapped")
	}
	b22Untap(g, memorial)
	if err := g.ActivateManaAbility(me.ID, memorial, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want B", got)
	}
	me.ManaPool = nil
	b22Untap(g, memorial)
	dead := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	b06AddMana(me, "C", "C", "B")
	b16Activate(t, g, me.ID, memorial, 0, game.ActivateAbilityParams{Targets: cardRefs(dead)})
	if !me.Hand.Contains(dead) || me.Graveyard.Contains(dead) {
		t.Error("the creature card returns to hand")
	}
	if g.Battlefield.Contains(memorial) || !me.Graveyard.Contains(memorial) {
		t.Error("the Memorial is sacrificed by its own ability")
	}
	if spec, _ := Lookup(b27MemorialToFollyOracle); spec.Completeness != CompletenessFull {
		t.Error("the Memorial is whole")
	}
}

func TestB27SanctumOfUginSacrificesToTutorAColorlessCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sanctum := b21Push(g, me.ID, "Sanctum of Ugin", "Land", b27SanctumOfUginOracle, 0, 0)
	seedSearchLibrary(me,
		game.Card{Name: "Ulamog", TypeLine: "Legendary Creature — Eldrazi", ManaCost: "{10}"},
		game.Card{Name: "Wurmcoil", TypeLine: "Artifact Creature — Wurm", ManaCost: "{6}"},
		game.Card{Name: "Green Wurm", TypeLine: "Creature — Wurm", ManaCost: "{5}{G}", Colors: []string{"G"}},
		game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{7}"},
	)
	// A six-drop artifact creature is colorless but too small.
	castWithCost(t, g, "Steel Hellkite", "Artifact Creature — Dragon", "{6}", "")
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("mana value 6 does not trigger the Sanctum")
	}
	castWithCost(t, g, "Kozilek", "Legendary Creature — Eldrazi", "{10}", "")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sanctum) || !me.Graveyard.Contains(sanctum) {
		t.Fatal("the Sanctum is sacrificed")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two colorless creature cards: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Green Wurm") != uuid.Nil || searchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("only colorless creature cards are offered")
	}
	answerSearchNamed(t, g, me.ID, "Wurmcoil")
	if !b22HandHasNamed(me, "Wurmcoil") {
		t.Error("the pick goes to hand")
	}
}

// --- the mana sources ----------------------------------------------

func TestB27OrcishLumberjackTradesAForestForThreeMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jack := pushCatalogPermanent(g, me.ID, "Orcish Lumberjack", "Creature — Orc", b27OrcishLumberjackOracle, false)
	if err := g.ActivateManaAbility(me.ID, jack, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("no Forest to sacrifice: the ability cannot be activated")
	}
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	plains := seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	if err := g.ActivateManaAbility(me.ID, jack, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{plains}}); err == nil {
		t.Fatal("a Plains is not a Forest")
	}
	if err := g.ActivateManaAbility(me.ID, jack, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{forest}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if g.Battlefield.Contains(forest) || !b27Tapped(t, g, jack) {
		t.Error("the Forest is sacrificed and the Lumberjack taps")
	}
	picks := 0
	for _, color := range []string{"R", "R", "G"} {
		c := manaPickFor(g, me.ID)
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
	if picks != 3 || manaPickFor(g, me.ID) != nil {
		t.Errorf("three colour picks, one per slot: %d", picks)
	}
	if got := poolColors(me); len(got) != 3 || got[0] != "R" || got[1] != "R" || got[2] != "G" {
		t.Errorf("pool %v, want [R R G]", got)
	}
	if spec, _ := Lookup(b27OrcishLumberjackOracle); spec.Completeness != CompletenessFull {
		t.Error("the Lumberjack is whole")
	}
}

// --- the spells ----------------------------------------------------

func TestB27EarthquakeHitsGroundCreaturesAndEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bird := pushTribalCreature(g, opp.ID, "Bird", "Creature — Bird", 2, 2, "flying")
	mine := pushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	before := b27Lives(g)
	castXSpell(t, g, "Earthquake", "Sorcery", b27EarthquakeOracle, "{X}{R}", 2, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 without flying dies to X=2")
	}
	if !g.Battlefield.Contains(bird) || damageMarkedOn(g, bird) != 0 {
		t.Error("a flyer is untouched")
	}
	if !g.Battlefield.Contains(mine) || damageMarkedOn(g, mine) != 2 {
		t.Error("your own ground creature takes it too")
	}
	for i, l := range b27Lives(g) {
		if l != before[i]-2 {
			t.Errorf("seat %d life %d, want %d (each player takes X)", i, l, before[i]-2)
		}
	}
}

func TestB27HellToPayPaysOutExcessInTappedTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castXSpell(t, g, "Hell to Pay", "Sorcery", b27HellToPayOracle, "{X}{R}", 5, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear dies")
	}
	if got := b27CountNamed(g, me.ID, "Treasure"); got != 3 {
		t.Errorf("5 damage to a 2/2: %d Treasures, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && !c.Tapped {
			t.Error("the Treasures enter tapped")
		}
	}
	// Damage already marked lowers what was lethal.
	ox := pushTribalCreature(g, opp.ID, "Ox", "Creature — Ox", 4, 4)
	b27Damage(g, me.ID, ox, 3)
	castXSpell(t, g, "Hell to Pay", "Sorcery", b27HellToPayOracle, "{X}{R}", 2, cardRefs(ox))
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Treasure"); got != 4 {
		t.Errorf("2 damage to a 4/4 with 3 marked: one more Treasure, got %d total", got)
	}
	// No excess, no Treasure.
	wall := pushTribalCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	castXSpell(t, g, "Hell to Pay", "Sorcery", b27HellToPayOracle, "{X}{R}", 3, cardRefs(wall))
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Treasure"); got != 4 {
		t.Errorf("3 damage to a 0/6 is no excess, got %d total", got)
	}
}

func TestB27FinaleOfGloryMakesSoldiersAndAngelsAtTen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "Finale of Glory", "Sorcery", b27FinaleOfGloryOracle, "{X}{W}{W}", 3, nil)
	passPriorityAroundTable(t, g)
	if b27CountNamed(g, me.ID, "Soldier") != 3 || b27CountNamed(g, me.ID, "Angel") != 0 {
		t.Errorf("X=3: %d Soldiers and %d Angels, want 3 and 0", b27CountNamed(g, me.ID, "Soldier"), b27CountNamed(g, me.ID, "Angel"))
	}
	castXSpell(t, g, "Finale of Glory", "Sorcery", b27FinaleOfGloryOracle, "{X}{W}{W}", 10, nil)
	passPriorityAroundTable(t, g)
	if b27CountNamed(g, me.ID, "Soldier") != 13 || b27CountNamed(g, me.ID, "Angel") != 10 {
		t.Errorf("X=10: %d Soldiers and %d Angels, want 13 and 10", b27CountNamed(g, me.ID, "Soldier"), b27CountNamed(g, me.ID, "Angel"))
	}
	for _, c := range g.Battlefield.Cards {
		switch c.Name {
		case "Soldier":
			if c.Power != 2 || !containsString(c.Keywords, "vigilance") || !containsString(c.Colors, "W") {
				t.Errorf("Soldier %+v, want a 2/2 white vigilance token", c)
			}
		case "Angel":
			if c.Power != 4 || !containsString(c.Keywords, "flying") || !containsString(c.Keywords, "vigilance") {
				t.Errorf("Angel %+v, want a 4/4 flying vigilance token", c)
			}
		}
	}
}

func TestB27OneWithTheMachineDrawsTheGreatestArtifactManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "One with the Machine", "Sorcery", b27OneWithTheMachineOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("no artifacts: nothing drawn")
	}
	b27Push(g, me.ID, "Signet", "Artifact", "", "{2}", 0, 0)
	b27Push(g, me.ID, "Wurmcoil", "Artifact Creature — Wurm", "", "{6}", 6, 6)
	b27Push(g, opp.ID, "Their Colossus", "Artifact Creature — Golem", "", "{11}", 11, 11)
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "One with the Machine", "Sorcery", b27OneWithTheMachineOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("drew %d, want 6 (your greatest artifact, not the table's)", got)
	}
}

func TestB27FarmerCottonMakesXHalflingsAndXFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "Farmer Cotton", "Legendary Creature — Halfling Peasant", b27FarmerCottonOracle, "{X}{G}{W}", 2, nil)
	passPriorityAroundTable(t, g)
	if b27CountNamed(g, me.ID, "Halfling") != 2 || b27CountNamed(g, me.ID, "Food") != 2 {
		t.Errorf("X=2: %d Halflings and %d Food, want 2 and 2", b27CountNamed(g, me.ID, "Halfling"), b27CountNamed(g, me.ID, "Food"))
	}
	if b27CountNamed(g, me.ID, "Farmer Cotton") != 1 {
		t.Error("Cotton himself is on the battlefield")
	}
	if spec, _ := Lookup(b27FarmerCottonOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the resolve-time posture must be declared")
	}
}

func TestB27TashasHideousLaughterExilesToTwentyManaValue(t *testing.T) {
	g := newCatalogGame(t)
	opp1, opp2, opp3 := g.Seats[1], g.Seats[2], g.Seats[3]
	// opp1: 6 + 6 + 6 + 6 = 24 on the fourth card; the fifth stays.
	// (seedSearchLibrary pushes to the bottom, so the first card
	// listed is the top.)
	seedSearchLibrary(opp1,
		game.Card{Name: "Top", TypeLine: "Creature", ManaCost: "{6}"},
		game.Card{Name: "Second", TypeLine: "Creature", ManaCost: "{6}"},
		game.Card{Name: "Third", TypeLine: "Creature", ManaCost: "{6}"},
		game.Card{Name: "Fourth", TypeLine: "Creature", ManaCost: "{6}"},
		game.Card{Name: "Deep", TypeLine: "Creature", ManaCost: "{6}"},
	)
	// opp2: lands and cheap spells never reach 20 — the whole
	// library goes, and they do not lose for it.
	seedSearchLibrary(opp2,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"},
	)
	// opp3: one twenty-drop on top.
	seedSearchLibrary(opp3,
		game.Card{Name: "Emrakul", TypeLine: "Creature", ManaCost: "{20}"},
		game.Card{Name: "Under", TypeLine: "Creature", ManaCost: "{1}"},
	)
	castCatalogSpell(t, g, "Tasha's Hideous Laughter", "Sorcery", b27TashasHideousLaughterOracle, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bauble := castCatalogSpell(t, g, "Urza's Bauble", "Artifact", b27UrzasBaubleOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bauble) {
		t.Fatal("the Bauble resolves to the battlefield")
	}
	if b27HandKnownBy(opp, me.ID) != 0 {
		t.Fatal("before the activation the opponent's hand is hidden")
	}
	b16Activate(t, g, me.ID, bauble, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
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
	advanceToUpkeepOf(t, g, 1)
	if len(g.PendingTriggers)+len(g.StackMeta) == 0 {
		t.Fatal("the draw is on the stack at the next upkeep, whoever's it is")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, got)
	}
}

func TestB27OswaldFiddlebenderPodsArtifactsUpOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oswald := pushCatalogPermanent(g, me.ID, "Oswald Fiddlebender", "Legendary Creature — Gnome Artificer", b27OswaldFiddlebenderOracle, false)
	signet := b27Push(g, me.ID, "Signet", "Artifact", "", "{2}", 0, 0)
	seedSearchLibrary(me,
		game.Card{Name: "Three Rock", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Other Three", TypeLine: "Artifact Creature — Construct", ManaCost: "{3}"},
		game.Card{Name: "Four Rock", TypeLine: "Artifact", ManaCost: "{4}"},
		game.Card{Name: "Three Elf", TypeLine: "Creature — Elf", ManaCost: "{2}{G}"},
	)
	spec, _ := Lookup(b27OswaldFiddlebenderOracle)
	if len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed {
		t.Fatal("one ability, sorcery speed")
	}
	advanceToMain(t, g)
	b06AddMana(me, "W")
	b16Activate(t, g, me.ID, oswald, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{signet}})
	if g.Battlefield.Contains(signet) || !b27Tapped(t, g, oswald) {
		t.Fatal("the artifact is sacrificed and Oswald taps")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two artifacts at mana value 3: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Four Rock") != uuid.Nil || searchOptionNamed(g, c, "Three Elf") != uuid.Nil {
		t.Error("only artifact cards with mana value exactly 3 are offered")
	}
	answerSearchNamed(t, g, me.ID, "Other Three")
	if findBattlefieldByName(g, "Other Three") == uuid.Nil {
		t.Error("the pick enters the battlefield")
	}
}

func TestB27MagewrightsStoneUntapsOnlyTapAbilityCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stone := pushCatalogPermanent(g, me.ID, "Magewright's Stone", "Artifact", b27MagewrightsStoneOracle, false)
	elves := pushCatalogPermanent(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", fyndhornElvesOracle, false)
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b16Tap(g, elves)
	b16Tap(g, bear)
	advanceToMain(t, g)
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a creature with no tap ability is not a legal target")
	}
	b16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{Targets: cardRefs(elves)})
	if b27Tapped(t, g, elves) {
		t.Error("the mana Elf is untapped")
	}
	if !b27Tapped(t, g, stone) {
		t.Error("the Stone taps to do it")
	}
	if spec, _ := Lookup(b27MagewrightsStoneOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the catalog-only target set must be declared")
	}
}

// --- the triggers --------------------------------------------------

func TestB27RevengeOfRavensDrainsPerAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Revenge of Ravens", "Enchantment", b27RevengeOfRavensOracle, 0, 0)
	a := pushVanillaCreature(g, opp.ID, "Raider A", 1, 1)
	b := pushVanillaCreature(g, opp.ID, "Raider B", 1, 1)
	c := pushVanillaCreature(g, opp.ID, "Raider C", 1, 1)
	other := g.Seats[2]
	advanceToMainOf(t, g, 1)
	meBefore, oppBefore, otherBefore := me.Life, opp.Life, other.Life
	declareAttack(t, g, me.ID, a, b)
	declareAttack(t, g, other.ID, c)
	passPriorityAroundTable(t, g)
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

func TestB27HeadlessRiderSpawnsForNontokenZombieDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := b21Push(g, me.ID, "Headless Rider", "Creature — Zombie", b27HeadlessRiderOracle, 3, 1)
	zombie := pushTribalCreature(g, me.ID, "Gravecrawler", "Creature — Zombie", 2, 1)
	theirs := pushTribalCreature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b27Kill(g, zombie)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Fatalf("a nontoken Zombie you control died: %d tokens, want 1", got)
	}
	b27Kill(g, theirs)
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("an opponent's Zombie and a non-Zombie: still %d tokens, want 1", got)
	}
	token := findBattlefieldByName(g, "Zombie")
	b27Kill(g, token)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Zombie"); got != 0 {
		t.Errorf("a token Zombie dying makes nothing: %d", got)
	}
	b27Kill(g, rider)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Zombie"); got != 1 {
		t.Errorf("the Rider's own death makes one: %d", got)
	}
}

func TestB27WolverineRidersSpawnEachUpkeepAndGainOnElves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Wolverine Riders", "Creature — Elf Warrior", b27WolverineRidersOracle, 4, 4)
	life := me.Life
	// The opponent's upkeep: a token, and the Elf entering gains 1.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Elf Warrior"); got != 1 {
		t.Fatalf("each upkeep, yours or not: %d tokens, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("the token is an Elf with toughness 1: life %d → %d", life, me.Life)
	}
	// Seats 2 and 3 and then your own: three more.
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Elf Warrior"); got != 4 {
		t.Errorf("after a full round of upkeeps: %d tokens, want 4", got)
	}
	life = me.Life
	b20CastCreature(t, g, me, "Elvish Archdruid", "Creature — Elf Druid", "", 2, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("another Elf entering: gain its toughness, life %d → %d", life, me.Life)
	}
	life = me.Life
	b20CastCreature(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Error("a non-Elf gains nothing")
	}
}

func TestB27RiteOfTheDragoncallerMakesADragonPerInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Rite of the Dragoncaller", "Enchantment", b27RiteOfTheDragoncallerOracle, 0, 0)
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Dragon"); got != 1 {
		t.Fatalf("an instant: %d Dragons, want 1", got)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Dragon"); got != 1 {
		t.Errorf("a creature spell: still %d Dragons", got)
	}
	dragon := findBattlefieldByName(g, "Dragon")
	if effectivePower(t, g, dragon) != 5 || !hasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("the Dragon is a 5/5 with flying")
	}
}

func TestB27FlumphDrawsForYouAndATargetOpponentWhenDamaged(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flumph := b21Push(g, me.ID, "Flumph", "Creature — Jellyfish", b27FlumphOracle, 0, 4)
	if !hasEffectiveKeyword(t, g, flumph, "defender") || !hasEffectiveKeyword(t, g, flumph, "flying") {
		t.Error("defender and flying are printed")
	}
	before := b27Hands(g)
	b27Damage(g, opp.ID, flumph, 1)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	after := b27Hands(g)
	if after[0] != before[0]+1 || after[1] != before[1]+1 {
		t.Errorf("hands %v → %v, want you and the chosen opponent +1", before, after)
	}
	if after[2] != before[2] || after[3] != before[3] {
		t.Error("the other opponents draw nothing")
	}
}

func TestB27CuriousAltisaurDrawsWhenADinosaurConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	altisaur := b21Push(g, me.ID, "Curious Altisaur", "Creature — Dinosaur", b27CuriousAltisaurOracle, 2, 5)
	raptor := pushTribalCreature(g, me.ID, "Raptor", "Creature — Dinosaur", 3, 3)
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, altisaur, "reach") || !hasEffectiveKeyword(t, g, altisaur, "vigilance") {
		t.Error("reach and vigilance are printed")
	}
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, raptor, opp.ID, 3)
	dealCombatDamageToPlayer(g, altisaur, opp.ID, 2)
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (one per Dinosaur, none for the Bear)", got)
	}
}

func TestB27DragonMageWheelsTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := b21Push(g, me.ID, "Dragon Mage", "Creature — Dragon Wizard", b27DragonMageOracle, 5, 5)
	for _, p := range g.Seats {
		fillHandTo(t, g, p, 3)
	}
	dealCombatDamageToPlayer(g, mage, opp.ID, 5)
	passPriorityAroundTable(t, g)
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
	passPriorityAroundTable(t, g)
	for i, h := range b27Hands(g) {
		if h != hands[i] {
			t.Error("noncombat damage does not wheel")
		}
	}
}

func TestB27FoundationBreakerDestroysOnEntryAndEvokes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b21Push(g, opp.ID, "Mana Rock", "Artifact", "", 0, 0)
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	breaker := castAndResolveCreature(t, g, "Foundation Breaker", "Creature — Elemental", b27FoundationBreakerOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, bear) {
		t.Error("a creature is not an artifact or enchantment")
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the chosen artifact is destroyed")
	}
	if !g.Battlefield.Contains(breaker) {
		t.Error("hard-cast, the Breaker stays")
	}
	// Evoked: the destroy still happens and the Breaker is sacrificed.
	curse := b21Push(g, opp.ID, "Curse", "Enchantment", "", 0, 0)
	evoked := castWithAltCost(t, g, "Foundation Breaker", "Creature — Elemental", b27FoundationBreakerOracle, "evoke")
	b27Settle(t, g, me.ID, curse)
	if g.Battlefield.Contains(curse) {
		t.Error("evoked, the enchantment is still destroyed")
	}
	if g.Battlefield.Contains(evoked) || !me.Graveyard.Contains(evoked) {
		t.Error("evoked, the Breaker is sacrificed")
	}
}

func TestB27TheGooseMotherEntersWithCountersFoodAndEatsToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goose := castXSpell(t, g, "The Goose Mother", "Legendary Creature — Bird Hydra", b27TheGooseMotherOracle, "{X}{G}{U}", 3, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, goose, "+1/+1"); got != 3 {
		t.Errorf("X=3: %d +1/+1 counters, want 3", got)
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("half of 3 rounded up: %d Food, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, goose, "flying") {
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
	declareAttack(t, g, opp.ID, goose)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	food := findBattlefieldByName(g, "Food")
	pickCard(t, g, me.ID, food)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(food) || b27CountNamed(g, me.ID, "Food") != 1 {
		t.Error("the chosen Food is sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, got)
	}
}

func TestB27TheGooseMotherWithNoFoodIsSilent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goose := b21Push(g, me.ID, "The Goose Mother", "Legendary Creature — Bird Hydra", b27TheGooseMotherOracle, 2, 2)
	declareAttack(t, g, opp.ID, goose)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil || latestPickTarget(g, me.ID) != nil {
		t.Error("no Food: no prompt at all (CR 603.3d)")
	}
}

func TestB27LegionLoyaltyMyriadCopiesAttackTheOtherOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	b21Push(g, me.ID, "Legion Loyalty", "Enchantment", b27LegionLoyaltyOracle, 0, 0)
	knight := pushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3, "vigilance")
	declareAttack(t, g, opp1.ID, knight)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	copies := 0
	attacking := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Knight" || c.InstanceID == knight {
			continue
		}
		copies++
		if !c.Tapped || !IsToken(c) || c.Controller != me.ID {
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
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if opp1.Life != l1-3 || opp2.Life != l2-3 || opp3.Life != l3-3 {
		t.Errorf("every opponent takes 3: %d %d %d → %d %d %d", l1, l2, l3, opp1.Life, opp2.Life, opp3.Life)
	}
	if got := b27CountNamed(g, me.ID, "Knight"); got != 1 {
		t.Errorf("at end of combat the copies are exiled: %d Knights left, want the original", got)
	}
	if !g.Battlefield.Contains(knight) {
		t.Error("the original stays")
	}
}

func TestB27LegionLoyaltyIsSilentInADuel(t *testing.T) {
	g := newDuelCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Legion Loyalty", "Enchantment", b27LegionLoyaltyOracle, 0, 0)
	knight := pushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3)
	declareAttack(t, g, opp.ID, knight)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("no other opponent: myriad has nothing to ask")
	}
	passPriorityAroundTable(t, g)
	if b27CountNamed(g, me.ID, "Knight") != 1 {
		t.Error("no copies")
	}
}
