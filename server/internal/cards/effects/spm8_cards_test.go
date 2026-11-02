package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm8_cards_test.go — Marvel's Spider-Man, batch 8 (#4460). Each test plays
// the real catalog entry and asserts what its text says.

const (
	spm8SpectacularOracle = "eed5e0cc-8f88-4779-bf10-6a823c72df7f"
	spm8BotOracle         = "e4bb33b2-adde-4176-9274-7598b559aa6a"
	spm8ByteOracle        = "845472ca-3fcb-4bcd-953c-96eb8d914af9"
	spm8GirlOracle        = "d459b3f5-a18c-4c17-a855-dc147f7c2161"
	spm8HamOracle         = "00b50215-e832-417f-9c80-2bc3050b6ebb"
	spm8IslandersOracle   = "2d1a15dc-3123-4425-b036-a3b5a83fcdf8"
	spm82099Oracle        = "19845456-3d96-48a4-89fa-9b8e583774c9"
	spm8IndiaOracle       = "8933297c-62f1-4df9-91b8-2d3481db77e5"
	spm8NoMoreOracle      = "163f1fce-70c8-44e3-a679-f97d52cd1e9a"
	spm8NoirOracle        = "2efd6494-c5e1-41de-bc6e-8490ff046cd1"
)

func TestSpm8EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spm8SpectacularOracle: "Spectacular Spider-Man",
		spm8BotOracle:         "Spider-Bot",
		spm8ByteOracle:        "Spider-Byte, Web Warden",
		spm8GirlOracle:        "Spider-Girl, Legacy Hero",
		spm8HamOracle:         "Spider-Ham, Peter Porker",
		spm8IslandersOracle:   "Spider-Islanders",
		spm82099Oracle:        "Spider-Man 2099",
		spm8IndiaOracle:       "Spider-Man India",
		spm8NoMoreOracle:      "Spider-Man No More",
		spm8NoirOracle:        "Spider-Man Noir",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func spm8Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return *c
}

// --- Spectacular Spider-Man ---

func TestSpm8SpectacularHasFlashAndGainsFlyingForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Spectacular Spider-Man", "Legendary Creature — Spider Human Hero", spm8SpectacularOracle, 3, 2)
	if !hasKeywordOnBattlefield(t, g, id, "flash") {
		t.Error("no flash")
	}
	if hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Fatal("flying before the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability was paid for with no mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("no flying after the ability resolved")
	}
}

func TestSpm8SpectacularSacrificeProtectsTheRestOfYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Spectacular Spider-Man", "Legendary Creature — Spider Human Hero", spm8SpectacularOracle, 3, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the sacrifice is a cost: he should already be in the graveyard")
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !hasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("my bear lacks %s", kw)
		}
		if hasKeywordOnBattlefield(t, g, theirs, kw) {
			t.Errorf("the opponent's bear gained %s", kw)
		}
	}
	// The set is fixed as the ability resolves (CR 611.2c).
	late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if hasKeywordOnBattlefield(t, g, late, "hexproof") {
		t.Error("a creature that arrived later gained hexproof")
	}
}

// --- Spider-Bot ---

func spm8BotGame(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Sol Ring", "Artifact"),
		searchTestLand("Dryad Arbor", "Land Creature — Forest Dryad"),
		searchTestLand("Plains", "Basic Land — Plains"),
	)
	castHoldCreature(t, g, "Spider-Bot", "Artifact Creature — Spider Robot Scout", spm8BotOracle, 2, 1)
	return g, me
}

func TestSpm8SpiderBotPutsABasicLandOnTop(t *testing.T) {
	g, me := spm8BotGame(t)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Sol Ring") != uuid.Nil || searchOptionNamed(g, c, "Dryad Arbor") != uuid.Nil {
		t.Error("only basic land cards are legal finds")
	}
	if searchOptionNamed(g, c, "Forest") == uuid.Nil || searchOptionNamed(g, c, "Plains") == uuid.Nil {
		t.Fatal("both basics are offered")
	}
	hand := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Plains")
	passPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.Name != "Plains" {
		t.Errorf("top card %q, want Plains", top.Name)
	}
	if me.Library.Size() != 4 || me.Hand.Size() != hand {
		t.Errorf("library %d hand %d: the card must stay in the library", me.Library.Size(), me.Hand.Size())
	}
}

func TestSpm8SpiderBotMayDeclineTheSearch(t *testing.T) {
	g, me := spm8BotGame(t)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	before := libraryIDs(me)
	if err := answerSearchApply(g, c, me.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sameIDList(before, libraryIDs(me)) {
		t.Error("declining the search shuffled or moved the library")
	}
}

// --- Spider-Byte, Web Warden ---

func TestSpm8SpiderByteBouncesANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := seedPermanentWithOracle(g, opp.ID, "Their Forest", "Basic Land — Forest", "")
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castHoldCreature(t, g, "Spider-Byte, Web Warden", "Legendary Creature — Spider Avatar Hero", spm8ByteOracle, 2, 2)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Fatal("a land was a legal target")
	}
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(rock) || g.Battlefield.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land moved")
	}
	_ = me
}

func TestSpm8SpiderByteChoosingNoTargetBouncesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castHoldCreature(t, g, "Spider-Byte, Web Warden", "Legendary Creature — Spider Avatar Hero", spm8ByteOracle, 2, 2)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
		t.Fatalf("pick no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was bounced with no target chosen")
	}
}

// --- Spider-Girl, Legacy Hero ---

func TestSpm8SpiderGirlFliesOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Spider-Girl, Legacy Hero", "Legendary Creature — Spider Human Hero", spm8GirlOracle, 2, 2)
	if !hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("no flying on my turn")
	}
	advanceToUpkeepOf(t, g, 1)
	if hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("flying on the opponent's turn")
	}
}

func TestSpm8SpiderGirlLeavingMakesAHumanCitizen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Spider-Girl, Legacy Hero", "Legendary Creature — Spider Human Hero", spm8GirlOracle, 2, 2)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}) })
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Human Citizen", me.ID); n != 1 {
		t.Fatalf("Human Citizen tokens = %d, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Human Citizen" {
			if c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Human") || !c.HasSubtype("Citizen") ||
				len(c.Colors) != 2 {
				t.Errorf("token %+v is not a 1/1 green and white Human Citizen", c)
			}
		}
	}
}

// --- Spider-Ham, Peter Porker ---

func TestSpm8SpiderHamFoodAndAnthem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := b12Creature(g, me.ID, "My Spider", "Creature — Spider", 1, 1)
	wolf := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	elf := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 2, 2)
	theirBoar := b12Creature(g, opp.ID, "Their Boar", "Creature — Boar", 2, 2)
	ham := castHoldCreature(t, g, "Spider-Ham, Peter Porker", "Legendary Creature — Spider Boar Hero", spm8HamOracle, 2, 2)

	if n := tokensNamedControlledBy(g, "Food", me.ID); n != 1 {
		t.Errorf("Food tokens = %d, want 1", n)
	}
	for _, c := range []struct {
		name string
		id   uuid.UUID
		p, t int
	}{
		{"my Spider", spider, 2, 2}, {"my Wolf", wolf, 3, 3}, {"my Elf", elf, 2, 2},
		{"their Boar", theirBoar, 2, 2}, {"Spider-Ham himself", ham, 2, 2},
	} {
		p, tough := effectivePTOf(t, g, c.id)
		if p != c.p || tough != c.t {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.p, c.t)
		}
	}
}

// --- Spider-Islanders (mayhem) ---

func TestSpm8SpiderIslandersMayhemCostsOneRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	// Put in the graveyard without being discarded: the zone is closed to it.
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Spider-Islanders", TypeLine: "Creature — Spider Horror Citizen",
		ManaCost: "{3}{R}", OracleID: spm8IslandersOracle, Owner: me.ID, Controller: me.ID})
	floatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err == nil {
		t.Fatal("cast from the graveyard without having discarded it")
	}
}

func TestSpm8SpiderIslandersMayhemCastAfterDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := mayhemHandCard(g, me, "Spider-Islanders", "Creature — Spider Horror Citizen", "{3}{R}", spm8IslandersOracle)
	mayhemDiscard(t, g, me, id)
	floatForTest(g, me, "CCCR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("the printed cost works from the graveyard")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err == nil {
		t.Fatal("one mana paid mayhem {1}{R}")
	}
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Spider-Islanders did not enter")
	}
}

// --- Spider-Man 2099 ---

func spm82099Hand(g *game.Game, me *game.Player) uuid.UUID {
	return handCardFull(me, "Spider-Man 2099", "Legendary Creature — Spider Human Hero", "{U}{R}", spm82099Oracle, []string{"U", "R"})
}

func TestSpm82099CannotBeCastDuringYourFirstThreeTurns(t *testing.T) {
	for turns, wantOK := range map[int]bool{1: false, 2: false, 3: false, 4: true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		me.TurnsBegun = turns
		toMainForCost(t, g)
		id := spm82099Hand(g, me)
		floatForTest(g, me, "UR")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if wantOK && err != nil {
			t.Errorf("turn %d: cast refused: %v", turns, err)
		}
		if !wantOK && err == nil {
			t.Errorf("turn %d: cast allowed", turns)
		}
	}
}

func spm82099InPlay(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	id := b12Push(g, me.ID, "Spider-Man 2099", "Legendary Creature — Spider Human Hero", spm82099Oracle, 2, 3)
	for _, kw := range []string{"double strike", "vigilance"} {
		if !hasKeywordOnBattlefield(t, g, id, kw) {
			t.Fatalf("no %s", kw)
		}
	}
	return g, me, opp, id
}

func TestSpm82099EndStepDamageAfterCastingFromTheGraveyard(t *testing.T) {
	g, me, opp, _ := spm82099InPlay(t)
	isl := mayhemHandCard(g, me, "Spider-Islanders", "Creature — Spider Horror Citizen", "{3}{R}", spm8IslandersOracle)
	mayhemDiscard(t, g, me, isl)
	floatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, isl, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	before := opp.Life
	advanceTo(t, g, game.StepEnd)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent life %d -> %d, want 2 damage (his power)", before, opp.Life)
	}
}

func TestSpm82099EndStepDamageAfterPlayingALandFromTheLibrary(t *testing.T) {
	g, me, opp, _ := spm82099InPlay(t)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Courser of Kruphix",
		TypeLine: "Enchantment Creature — Centaur", OracleID: courserOracle,
		Owner: me.ID, Controller: me.ID,
	})
	top := seedTopOfLibrary(me, game.Card{Name: "Top Island", TypeLine: "Basic Land — Island"})
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play the land off the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	before := opp.Life
	advanceTo(t, g, game.StepEnd)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent life %d -> %d, want 2 damage", before, opp.Life)
	}
}

func TestSpm82099NothingForPlaysFromYourHand(t *testing.T) {
	g, me, opp, _ := spm82099InPlay(t)
	castHoldCreature(t, g, "Hand Bear", "Creature — Bear", "", 2, 2)
	playLandFromHand(t, g, "Hand Island", "")
	before := opp.Life
	advanceTo(t, g, game.StepEnd)
	if ch := latestPickTarget(g, me.ID); ch != nil {
		t.Fatal("the end-step trigger fired with only plays from hand")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("opponent life changed: %d -> %d", before, opp.Life)
	}
}

// --- Spider-Man India ---

func TestSpm8IndiaCreatureSpellCountersAndGivesFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Spider-Man India", "Legendary Creature — Spider Human Hero", spm8IndiaOracle, 4, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hand Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was a legal target")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 {
		t.Errorf("counters on the bear = %d, want 1", plusOnes(g, bear))
	}
	if !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the bear did not gain flying")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "flying") || plusOnes(g, theirs) != 0 {
		t.Error("the opponent's creature was affected")
	}
}

func TestSpm8IndiaIgnoresNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Spider-Man India", "Legendary Creature — Spider Human Hero", spm8IndiaOracle, 4, 4)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hand Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if ch := latestPickTarget(g, me.ID); ch != nil {
		t.Fatal("a noncreature spell triggered Pavitr's Seva")
	}
}

func TestSpm8IndiaWebSlingingNeedsATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	india := handCardFull(me, "Spider-Man India", "Legendary Creature — Spider Human Hero", "{3}{G}{W}", spm8IndiaOracle, []string{"G", "W"})
	untapped := b12Creature(g, me.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	floatForTest(g, me, "CGW")
	if err := g.CastSpell(me.ID, india, game.CastSpellParams{Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{untapped}}); err == nil {
		t.Fatal("web-slinging was paid with an untapped creature")
	}
	if err := g.CastSpell(me.ID, india, game.CastSpellParams{Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{tapped}}); err != nil {
		t.Fatalf("web-slinging: %v", err)
	}
	if !me.Hand.Contains(tapped) {
		t.Error("the tapped creature did not return to hand")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(india) {
		t.Error("Spider-Man India did not enter")
	}
}

// --- Spider-Man No More ---

func TestSpm8NoMoreMakesACitizenWithDefender(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Commander", TypeLine: "Legendary Creature — Elf Druid",
		Power: 5, Toughness: 5, Keywords: []string{"flying", "trample"}, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Spider-Man No More", "Enchantment — Aura", spm8NoMoreOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	c := spm8Live(t, g, victim)
	if !c.IsCreature() || !c.IsLegendary() {
		t.Errorf("type line %q: it keeps its card types and supertypes", c.TypeLine)
	}
	if !c.HasSubtype("Citizen") || c.HasSubtype("Elf") || c.HasSubtype("Druid") {
		t.Errorf("subtypes of %q: want only Citizen", c.TypeLine)
	}
	if p, tough := effectivePTOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("size %d/%d, want 1/1", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, victim, "defender") {
		t.Error("no defender")
	}
	for _, kw := range []string{"flying", "trample"} {
		if hasKeywordOnBattlefield(t, g, victim, kw) {
			t.Errorf("still has %s", kw)
		}
	}
	// Counters still land on top of the set base size (layer 7d).
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, victim, game.CounterPlusOne, 2); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if c := spm8Live(t, g, victim); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("with two counters %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Spider-Man Noir ---

func spm8NoirGame(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	noir := b12Push(g, me.ID, "Spider-Man Noir", "Legendary Creature — Spider Human Hero", spm8NoirOracle, 4, 4)
	seedLibrary(me, "Card A", "Card B", "Card C", "Card D")
	return g, me, opp, noir
}

func TestSpm8NoirLoneAttackerGetsACounterThenSurveilsX(t *testing.T) {
	g, me, opp, _ := spm8NoirGame(t)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, bear, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 2 {
		t.Fatalf("counters = %d, want 2", plusOnes(g, bear))
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil %d cards, want X = 2 (the counters on it)", len(c.ScryCards))
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards[:1], c.ScryCards[1:]); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("graveyard = %d, want 1 surveilled card", me.Graveyard.Size())
	}
}

func TestSpm8NoirCountsEveryKindOfCounter(t *testing.T) {
	g, me, opp, _ := spm8NoirGame(t)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, bear, game.CounterMinusOne, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	// Both counters are on it as the ability resolves (the +1/+1 and -1/-1
	// annihilate only when state-based actions next run, CR 704.5q), so X
	// counts the -1/-1 counter too: "the number of counters on it".
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil %d cards, want X = 2", len(c.ScryCards))
	}
}

func TestSpm8NoirDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g, me, opp, noir := spm8NoirGame(t)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear, noir)
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 0 || plusOnes(g, noir) != 0 {
		t.Error("a counter landed though two creatures attacked")
	}
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("a surveil prompt opened")
	}
}

func TestSpm8NoirTriggersForHisOwnLoneAttack(t *testing.T) {
	g, me, opp, noir := spm8NoirGame(t)
	declareAttack(t, g, opp.ID, noir)
	passPriorityAroundTable(t, g)
	if plusOnes(g, noir) != 1 {
		t.Errorf("counters on Noir = %d, want 1", plusOnes(g, noir))
	}
	if c := surveilChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Errorf("surveil prompt = %+v, want X = 1", c)
	}
}
