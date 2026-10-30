package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m17_cards_test.go — Marvel Super Heroes Commander, M–Z batch 17
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	mscM17HeraldOracle       = "be02005c-0e37-4dba-ac8e-3c69d06bbec2"
	mscM17VoyagerOracle      = "3ce1f550-d044-4cb1-991b-e98449f74507"
	mscM17StiltManOracle     = "8ae1bda5-8f70-48ec-9df0-89e58a4ab21c"
	mscM17CloneSagaOracle    = "0f0282c6-aceb-4879-8ec8-482d0501204a"
	mscM17NullificationOrcl  = "2fe0ebf5-52ed-4e92-9b93-81e9ae564439"
	mscM17SyndicationOracle  = "4777fa96-12a5-42d5-840f-ddd85f4a4b68"
	mscM17AssistantOracle    = "69db9cb8-9152-4d1b-9ac9-16aac42c0908"
	mscM17GalactusOracle     = "4232995b-68c0-4514-8cde-bc62b9d1cbaa"
	mscM17VillainType        = "Creature — Human Villain"
	mscM17LegendaryVillain   = "Legendary Creature — Human Villain"
	mscM17GalactusName       = "Galactus, Devourer of Worlds"
	mscM17CloneSagaTypeLine  = "Enchantment — Saga"
	mscM17SyndicationTypeLne = "Enchantment — Plan"
)

func TestMscM17EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM17HeraldOracle:      "Silver Surfer, Galactus's Herald",
		mscM17VoyagerOracle:     "Silver Surfer, Cosmic Voyager",
		mscM17StiltManOracle:    "Stilt-Man, Towering Terror",
		mscM17CloneSagaOracle:   "The Clone Saga",
		mscM17NullificationOrcl: "Ultimate Nullification",
		mscM17SyndicationOracle: "Villainous Syndication",
		mscM17AssistantOracle:   "Virtual Assistant",
	} {
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

func mscM17Controller(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Controller
}

// --- Virtual Assistant -----------------------------------------------------------

func mscM17AssistantTable(t *testing.T) (g *game.Game, me *game.Player, assistant uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	assistant = pushCatalogPermanent(g, me.ID, "Virtual Assistant", "Artifact Creature — Illusion Advisor", mscM17AssistantOracle, false)
	seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	return g, me, assistant
}

func TestMscM17VirtualAssistantHasDefender(t *testing.T) {
	g, _, assistant := mscM17AssistantTable(t)
	if !hasEffectiveKeyword(t, g, assistant, "defender") {
		t.Error("Virtual Assistant lacks defender")
	}
}

func TestMscM17VirtualAssistantMakesARobotWhenYouCastUsingTeamwork(t *testing.T) {
	g, me, _ := mscM17AssistantTable(t)
	a := pushVanillaCreature(g, me.ID, "Elf A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Elf B", 1, 1)
	if _, err := castPaying(t, g, "Crossover Collaboration", "Instant", crossoverCollaborationOracle, nil,
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := mshB1Tokens(g, me.ID, "Robot"); got != 1 {
		t.Fatalf("Robot tokens = %d, want 1 for a spell cast using teamwork", got)
	}
	tok := mshB1Token(t, g, me.ID, "Robot")
	if !tok.HasSubtype("Hero") || !tok.IsArtifact() || !tok.IsCreature() || len(tok.Colors) != 0 || tok.Power != 1 || tok.Toughness != 1 {
		t.Errorf("token = %s %v %d/%d, want a 1/1 colorless Robot Hero artifact creature", tok.TypeLine, tok.Colors, tok.Power, tok.Toughness)
	}
	if !hasEffectiveKeyword(t, g, tok.InstanceID, "flying") {
		t.Error("the token lacks flying")
	}
}

func TestMscM17VirtualAssistantIgnoresASpellCastWithoutTeamwork(t *testing.T) {
	g, me, _ := mscM17AssistantTable(t)
	// The same spell, teamwork left unpaid, and a spell with no teamwork.
	if _, err := castPaying(t, g, "Crossover Collaboration", "Instant", crossoverCollaborationOracle, nil,
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)

	if got := mshB1Tokens(g, me.ID, "Robot"); got != 0 {
		t.Errorf("Robot tokens = %d, want 0 without teamwork", got)
	}
}

// --- Ultimate Nullification ----------------------------------------------------

func TestMscM17UltimateNullificationNeedsALegendaryCreatureToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	floatForTest(g, me, "CCCCW")
	for name, params := range map[string]game.CastSpellParams{
		"no sacrifice":            {Strict: true},
		"a nonlegendary creature": {Strict: true, SacrificeIDs: []uuid.UUID{plain}},
		"a legendary noncreature": {Strict: true, SacrificeIDs: []uuid.UUID{b12Permanent(g, me.ID, "Legend Relic", "Legendary Artifact")}},
		"an opponent's legend":    {Strict: true, SacrificeIDs: []uuid.UUID{b12Push(g, g.Seats[1].ID, "Theirs", mscM17LegendaryVillain, "", 2, 2)}},
	} {
		if _, err := castWithTapParams(t, g, "Ultimate Nullification", "Sorcery", "{4}{W}", mscM17NullificationOrcl, params); err == nil {
			t.Errorf("Ultimate Nullification was cast with %s", name)
		}
	}
	if !g.Battlefield.Contains(plain) {
		t.Error("a refused cast sacrificed a creature")
	}
}

func TestMscM17UltimateNullificationExilesCreaturesAndGraveyardsThenGoesToTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	legend := b12Push(g, me.ID, "Mighty Legend", mscM17LegendaryVillain, "", 3, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	myGrave := gyCard(me, "My Old Spell", "Instant", "", "")
	theirGrave := gyCard(opp, "Their Old Spell", "Sorcery", "", "")
	library := me.Library.Size()
	floatForTest(g, me, "CCCCW")

	spell, err := castWithTapParams(t, g, "Ultimate Nullification", "Sorcery", "{4}{W}", mscM17NullificationOrcl,
		game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{legend}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"the sacrificed legend": legend, "my Bear": mine, "their Bear": theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s is still on the battlefield", name)
		}
	}
	if !g.Battlefield.Contains(relic) {
		t.Error("a noncreature permanent was exiled")
	}
	for name, id := range map[string]uuid.UUID{"my Bear": mine, "their Bear": theirs, "my graveyard card": myGrave, "their graveyard card": theirGrave} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", name)
		}
	}
	if me.Graveyard.Size() != 0 || opp.Graveyard.Size() != 0 {
		t.Errorf("graveyards hold %d and %d cards, want none", me.Graveyard.Size(), opp.Graveyard.Size())
	}
	if me.Library.Size() != library+1 || me.Library.Cards[0].InstanceID != spell {
		t.Errorf("Ultimate Nullification is not on the bottom of its owner's library (size %d, want %d)", me.Library.Size(), library+1)
	}
}

// --- Villainous Syndication --------------------------------------------------------

func mscM17SyndicationTable(t *testing.T) (g *game.Game, me *game.Player, plan uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	powerUpMain(t, g)
	plan = pushCatalogPermanent(g, me.ID, "Villainous Syndication", mscM17SyndicationTypeLne, mscM17SyndicationOracle, false)
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	return g, me, plan
}

func mscM17Activate(t *testing.T, g *game.Game, me *game.Player, plan uuid.UUID, tap ...uuid.UUID) error {
	t.Helper()
	return g.ActivateCatalogAbility(me.ID, plan, 0, game.ActivateAbilityParams{TapIDs: tap})
}

func TestMscM17SyndicationTapsAVillainToMillAndCountAPlan(t *testing.T) {
	g, me, plan := mscM17SyndicationTable(t)
	villain := b12Creature(g, me.ID, "Henchman", mscM17VillainType, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	library, graveyard := me.Library.Size(), me.Graveyard.Size()

	if err := mscM17Activate(t, g, me, plan, bear); err == nil {
		t.Fatal("a non-Villain paid the cost")
	}
	if err := mscM17Activate(t, g, me, plan); err == nil {
		t.Fatal("the ability was activated without tapping a Villain")
	}
	if err := mscM17Activate(t, g, me, plan, villain); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, villain) {
		t.Error("the Villain was not tapped")
	}
	if me.Library.Size() != library-1 || me.Graveyard.Size() != graveyard+1 {
		t.Errorf("library %d graveyard %d, want one card milled", me.Library.Size(), me.Graveyard.Size())
	}
	if got := countersOn(g, plan, "plan"); got != 1 {
		t.Errorf("plan counters = %d, want 1", got)
	}
	// An already-tapped Villain can't pay it again.
	if err := mscM17Activate(t, g, me, plan, villain); err == nil {
		t.Error("a tapped Villain paid the cost")
	}
}

func TestMscM17SyndicationIsSorcerySpeed(t *testing.T) {
	g, me, plan := mscM17SyndicationTable(t)
	villain := b12Creature(g, me.ID, "Henchman", mscM17VillainType, 1, 1)
	advanceTo(t, g, game.StepBeginCombat)
	if err := mscM17Activate(t, g, me, plan, villain); err == nil {
		t.Error("the ability was activated outside a main phase")
	}
}

func TestMscM17SyndicationFourthCounterSacrificesAndReturnsACreature(t *testing.T) {
	g, me, plan := mscM17SyndicationTable(t)
	var villains []uuid.UUID
	for i := 0; i < 4; i++ {
		villains = append(villains, b12Creature(g, me.ID, "Henchman", mscM17VillainType, 1, 1))
	}
	fallen := gyCard(me, "Fallen Champion", "Creature — Human Hero", "", "")
	spell := gyCard(me, "Old Spell", "Instant", "", "")

	for i := 0; i < 3; i++ {
		if err := mscM17Activate(t, g, me, plan, villains[i]); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(plan) {
			t.Fatalf("the plan was sacrificed after %d counters", i+1)
		}
	}
	if err := mscM17Activate(t, g, me, plan, villains[3]); err != nil {
		t.Fatalf("activate 4: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(plan) {
		t.Fatal("the fourth plan counter did not sacrifice the enchantment")
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, spell) {
			t.Errorf("a noncreature card was offered: %v", p.PickTargetCards)
		}
		pickCard(t, g, me.ID, fallen)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fallen) {
		t.Error("the creature card did not return to the battlefield")
	}
	if c, ok := battlefieldCard(g, fallen); ok && c.Controller != me.ID {
		t.Error("the creature returned under another player's control")
	}
}

// --- The Clone Saga ------------------------------------------------------------------

func mscM17CastCloneSaga(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	seedLibrary(me, "Bin Me", "Keep Me", "Third", "Fourth", "Fifth", "Sixth")
	saga := castCatalogSpell(t, g, "The Clone Saga", mscM17CloneSagaTypeLine, mscM17CloneSagaOracle, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("chapter I asked no surveil: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 3 {
		t.Errorf("surveil looks at %d cards, want 3", len(c.ScryCards))
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards[:1], c.ScryCards[1:]); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	return saga
}

func TestMscM17CloneSagaChapterOneSurveilsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	graveyard := me.Graveyard.Size()
	mscM17CastCloneSaga(t, g)
	if me.Graveyard.Size() != graveyard+1 {
		t.Errorf("graveyard = %d, want %d: one of the three surveilled cards was binned", me.Graveyard.Size(), graveyard+1)
	}
}

func mscM17CastLegend(t *testing.T, g *game.Game, me *game.Player, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Legendary Creature — Human Hero",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestMscM17CloneSagaChapterTwoCopiesTheNextCreatureSpellWithoutLegendary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mscM17CastCloneSaga(t, g)
	powerUpPassTo(t, g, 0)
	passPriorityAroundTable(t, g) // chapter II

	// A noncreature spell is not the one: the trigger keeps waiting.
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if got := len(battlefieldIDsNamed(g, "Stan Lee")); got != 0 {
		t.Fatalf("%d Stan Lees before any was cast", got)
	}

	mscM17CastLegend(t, g, me, "Stan Lee")
	ids := battlefieldIDsNamed(g, "Stan Lee")
	if len(ids) != 2 {
		t.Fatalf("%d Stan Lees on the battlefield, want the card and its copy (a legendary copy would have lost the legend rule)", len(ids))
	}
	legendary, tokens := 0, 0
	for _, id := range ids {
		c, _ := battlefieldCard(g, id)
		if c.IsToken() {
			tokens++
		}
		if strings.Contains(c.TypeLine, "Legendary") {
			legendary++
		}
	}
	if tokens != 1 || legendary != 1 {
		t.Errorf("tokens %d legendary %d, want one token copy and one legendary original", tokens, legendary)
	}

	// Used up: the next creature spell is not copied.
	mscM17CastLegend(t, g, me, "Jack Kirby")
	if got := len(battlefieldIDsNamed(g, "Jack Kirby")); got != 1 {
		t.Errorf("%d Jack Kirbys, want 1: the trigger copies only the next creature spell", got)
	}
}

func TestMscM17CloneSagaChapterThreeDrawsForTheNamedCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saga := mscM17CastCloneSaga(t, g)
	named := b12Creature(g, me.ID, "Clone A", "Creature — Clone", 1, 1)
	twin := b12Creature(g, me.ID, "Clone A", "Creature — Clone", 1, 1)
	other := b12Creature(g, me.ID, "Clone B", "Creature — Clone", 1, 1)
	powerUpPassTo(t, g, 0)
	passPriorityAroundTable(t, g) // chapter II
	powerUpPassTo(t, g, 0)
	passPriorityAroundTable(t, g) // chapter III asks for the name
	actNameFor(t, g, me.ID, "Clone A")
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}

	seedLibrary(me, "D1", "D2", "D3", "D4", "D5")
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, named, twin, other)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2: one per Clone A that connected, none for Clone B", got)
	}
}

func TestMscM17CloneSagaChapterThreeEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mscM17CastCloneSaga(t, g)
	named := b12Creature(g, me.ID, "Clone A", "Creature — Clone", 1, 1)
	powerUpPassTo(t, g, 0)
	passPriorityAroundTable(t, g)
	powerUpPassTo(t, g, 0)
	passPriorityAroundTable(t, g)
	actNameFor(t, g, me.ID, "Clone A")
	passPriorityAroundTable(t, g)

	powerUpPassTo(t, g, 0) // next of my turns: the trigger was for the turn chapter III resolved
	seedLibrary(me, "D1", "D2", "D3", "D4", "D5")
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, named)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards on a later turn, want 0", got)
	}
}

// --- Silver Surfer, Galactus's Herald -----------------------------------------------

func TestMscM17HeraldHasFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	herald := pushCatalogPermanent(g, me.ID, "Silver Surfer, Galactus's Herald", "Legendary Creature — Alien Hero", mscM17HeraldOracle, false)
	if !hasEffectiveKeyword(t, g, herald, "flying") {
		t.Error("Silver Surfer lacks flying")
	}
}

func mscM17CastHerald(t *testing.T, g *game.Game, me *game.Player) (galactus uuid.UUID, other uuid.UUID) {
	t.Helper()
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bystander", TypeLine: "Creature — Bear"},
		game.Card{Name: mscM17GalactusName, TypeLine: "Legendary Creature — Elder Alien", OracleID: mscM17GalactusOracle},
	)
	castCatalogSpell(t, g, "Silver Surfer, Galactus's Herald", "Legendary Creature — Alien Hero", mscM17HeraldOracle, nil)
	passPriorityAroundTable(t, g)
	return ids[1], ids[0]
}

func TestMscM17HeraldFindsGalactusOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	galactus, other := mscM17CastHerald(t, g, me)
	answerSearchByID(t, g, me.ID, galactus)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(galactus) {
		t.Error("Galactus is not in hand")
	}
	if me.Hand.Contains(other) {
		t.Error("another card was found")
	}
}

func TestMscM17HeraldMayFindNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	galactus, _ := mscM17CastHerald(t, g, me)
	answerSearchByID(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(galactus) {
		t.Error("Galactus was found after declining")
	}
}

func TestMscM17HeraldForcesATargetCreatureToAttackThePlayerItHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	herald := pushCatalogPermanent(g, me.ID, "Silver Surfer, Galactus's Herald", "Legendary Creature — Alien Hero", mscM17HeraldOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, herald)
	pick := game.TargetRef{Kind: game.TargetCard, ID: bear}
	assistCardSettle(t, g, me.ID, &pick, true)

	// Their turns in between owe nothing of my Bear; on my next turn it
	// must attack the player the Surfer hit.
	powerUpPassTo(t, g, 0)
	advanceTo(t, g, game.StepDeclareAttackers)
	var must map[uuid.UUID][]uuid.UUID
	g.WithWriteLock(func() { must = g.MustAttackForEffect() })
	if got := must[bear]; len(got) != 1 || got[0] != opp.ID {
		t.Fatalf("Bear must attack %v, want only the player Silver Surfer hit (%v)", got, opp.ID)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("pass with the Bear home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(bear, third.ID); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("attacking another player = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Errorf("attacking the player Silver Surfer hit: %v", err)
	}
}

func TestMscM17HeraldsRequirementEndsAfterYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	herald := pushCatalogPermanent(g, me.ID, "Silver Surfer, Galactus's Herald", "Legendary Creature — Alien Hero", mscM17HeraldOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, herald)
	pick := game.TargetRef{Kind: game.TargetCard, ID: bear}
	assistCardSettle(t, g, me.ID, &pick, true)

	powerUpPassTo(t, g, 0) // my next turn: still owed
	powerUpPassTo(t, g, 0) // the turn after: the effect ran out
	advanceTo(t, g, game.StepDeclareAttackers)
	var must map[uuid.UUID][]uuid.UUID
	g.WithWriteLock(func() { must = g.MustAttackForEffect() })
	if got := must[bear]; len(got) != 0 {
		t.Errorf("Bear still owes an attack on %v two turns later", got)
	}
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Errorf("attacking another player after the effect ended: %v", err)
	}
}

// --- Stilt-Man, Towering Terror -------------------------------------------------------

func mscM17StiltManTable(t *testing.T) (g *game.Game, me, opp *game.Player, stilt uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	stilt = pushCatalogPermanent(g, me.ID, "Stilt-Man, Towering Terror", "Legendary Creature — Human Rogue Villain", mscM17StiltManOracle, false)
	return g, me, opp, stilt
}

func TestMscM17StiltManHasReach(t *testing.T) {
	g, _, _, stilt := mscM17StiltManTable(t)
	if !hasEffectiveKeyword(t, g, stilt, "reach") {
		t.Error("Stilt-Man lacks reach")
	}
}

func TestMscM17StiltManStealsANoncreatureNonlandPermanentThatPlayerControls(t *testing.T) {
	g, me, opp, _ := mscM17StiltManTable(t)
	villain := b12Creature(g, me.ID, "Henchman", mscM17VillainType, 2, 2)
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	third := b12Permanent(g, g.Seats[2].ID, "Third's Relic", "Artifact")

	attackWith(t, g, opp.ID, villain)
	passPriorityAroundTable(t, g)
	offered := mscM13OpenPickTargetCards(g, me.ID)
	if !hasID(offered, relic) {
		t.Fatalf("the Relic was not offered: %v", offered)
	}
	for name, id := range map[string]uuid.UUID{"a land": land, "a creature": bear, "another player's permanent": third} {
		if hasID(offered, id) {
			t.Errorf("%s was offered", name)
		}
	}
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)

	if got := mscM17Controller(t, g, relic); got != me.ID {
		t.Fatalf("the Relic is controlled by %v, want me", got)
	}
	var cant bool
	g.WithWriteLock(func() { cant = g.CantBeSacrificedForEffect(relic) })
	if !cant {
		t.Error("the stolen permanent can be sacrificed")
	}
}

func TestMscM17StiltManKeepsTheControlThroughYourNextTurnOnly(t *testing.T) {
	g, me, opp, _ := mscM17StiltManTable(t)
	villain := b12Creature(g, me.ID, "Henchman", mscM17VillainType, 2, 2)
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	attackWith(t, g, opp.ID, villain)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)

	powerUpPassTo(t, g, 1)
	if got := mscM17Controller(t, g, relic); got != me.ID {
		t.Errorf("on their turn the Relic is controlled by %v, want me", got)
	}
	powerUpPassTo(t, g, 0)
	if got := mscM17Controller(t, g, relic); got != me.ID {
		t.Errorf("on my next turn the Relic is controlled by %v, want me", got)
	}
	powerUpPassTo(t, g, 1)
	if got := mscM17Controller(t, g, relic); got != opp.ID {
		t.Errorf("after my next turn the Relic is controlled by %v, want its owner back", got)
	}
	var cant bool
	g.WithWriteLock(func() { cant = g.CantBeSacrificedForEffect(relic) })
	if cant {
		t.Error("the Relic still can't be sacrificed after the turn ended")
	}
}

func TestMscM17StiltManTriggersOncePerPlayerForManyVillainsAndNotForOthers(t *testing.T) {
	g, me, opp, stilt := mscM17StiltManTable(t)
	v1 := b12Creature(g, me.ID, "Henchman A", mscM17VillainType, 2, 2)
	v2 := b12Creature(g, me.ID, "Henchman B", mscM17VillainType, 2, 2)
	b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	attackWith(t, g, opp.ID, v1, v2)
	passPriorityAroundTable(t, g)
	if got := mscM13TriggersFrom(g, stilt); got != 1 {
		t.Errorf("%d Stilt-Man triggers for two Villains hitting one player, want 1", got)
	}
}

func TestMscM17StiltManIgnoresNonVillains(t *testing.T) {
	g, me, opp, stilt := mscM17StiltManTable(t)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Henchman", mscM17VillainType, 2, 2)
	b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := mscM13TriggersFrom(g, stilt); got != 0 {
		t.Errorf("%d triggers for a Bear, want 0", got)
	}
	_ = theirs
}

// --- Silver Surfer, Cosmic Voyager -----------------------------------------------------

func TestMscM17VoyagerHasFlashAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	voyager := pushCatalogPermanent(g, me.ID, "Silver Surfer, Cosmic Voyager", "Legendary Creature — Alien Hero", mscM17VoyagerOracle, false)
	for _, kw := range []string{"flash", "flying"} {
		if !hasEffectiveKeyword(t, g, voyager, kw) {
			t.Errorf("Silver Surfer lacks %s", kw)
		}
	}
}

func TestMscM17VoyagerBlinksAtTheNextEndStepWithLandsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushCatalogPermanent(g, me.ID, "Plain Bear", "Creature — Bear", "", false)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	untouched := b12Creature(g, me.ID, "Untouched Bear", "Creature — Bear", 2, 2)

	castAndResolveCreature(t, g, "Silver Surfer, Cosmic Voyager", "Legendary Creature — Alien Hero", mscM17VoyagerOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt for the enters trigger")
	}
	if hasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's permanent was offered")
	}
	for _, id := range p.PickTargetCards {
		if c, ok := battlefieldCard(g, id); ok && c.Name == "Silver Surfer, Cosmic Voyager" {
			t.Error("Silver Surfer offered himself")
		}
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(bear, relic, forest)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the Bear": bear, "the Relic": relic, "the Forest": forest} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not exiled", name)
		}
		if !g.Exile.Contains(id) {
			t.Errorf("%s is not in exile", name)
		}
	}
	if !g.Battlefield.Contains(untouched) {
		t.Error("an untargeted permanent was exiled")
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	var back []uuid.UUID
	for _, name := range []string{"Plain Bear", "Relic", "Forest"} {
		ids := battlefieldIDsNamed(g, name)
		if len(ids) != 1 {
			t.Fatalf("%d %s on the battlefield after the end step began, want 1", len(ids), name)
		}
		back = append(back, ids[0])
		if got := mscM17Controller(t, g, ids[0]); got != me.ID {
			t.Errorf("%s returned under %v, want its owner", name, got)
		}
	}
	if tappedForTest(t, g, back[0]) || tappedForTest(t, g, back[1]) {
		t.Error("a nonland permanent returned tapped")
	}
	if !tappedForTest(t, g, back[2]) {
		t.Error("the land returned untapped: if a land enters this way, it enters tapped")
	}
}
