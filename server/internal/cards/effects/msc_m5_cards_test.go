package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m5_cards_test.go — Marvel Super Heroes Commander, M–Z half, batch 5
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	mscM5SpaceshiftOracle   = "5c63a244-75d7-4372-baf4-7a3526f240c8"
	mscM5MoleManOracle      = "a38f7ce1-6ccc-4fd3-9e02-631e8933c1f8"
	mscM5MollyOracle        = "e3f0a2a6-7a42-46bc-9736-d0f6a0ce0b37"
	mscM5NamoraOracle       = "d0cd3edd-d908-47c5-bff1-2089c22ab41c"
	mscM5SheHulkOracle      = "db99d52b-bbe6-4a48-8f2c-bf77912dc0a9"
	mscM5ArmorOracle        = "cdcfc66d-3d3a-4311-b864-937e7b624d61"
	mscM5BlackWidowOracle   = "3bc0479f-0f45-4b36-9963-f8eb5973cf47"
	mscM5IronManOracle      = "6abb489e-cc62-42d7-b898-3e5095037e75"
	mscM5ThorOracle         = "d4e4614d-8247-4fc5-8883-e37204939aa4"
	mscM5HulkOracle         = "7398c8b2-373a-41c7-ac6c-7329f5944462"
	mscM5MetalworkOracle    = "64ffe529-ba53-48ec-8af1-b046fbd1f0cc"
	mscM5MetalworkTypeLine  = "Artifact Creature — Construct"
	mscM5SagaTypeLine       = "Enchantment — Saga"
	mscM5PowerUpTypeLine    = "Legendary Creature — Mutant Hero"
	mscM5SheHulkTypeLine    = "Legendary Creature — Gamma Advisor Hero"
	mscM5NamoraTypeLine     = "Legendary Creature — Mutant Merfolk Hero"
	mscM5DefaultLibrarySize = 12
)

func TestMscM5EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM5SpaceshiftOracle: "Spaceshift",
		mscM5MoleManOracle:    "Return of the Mole Man",
		mscM5MollyOracle:      "Molly Hayes, Runaway",
		mscM5NamoraOracle:     "Namora, the Sea Queen",
		mscM5SheHulkOracle:    "She-Hulk, Attorney-at-Law",
		mscM5ArmorOracle:      "Midnight Angel Armor",
		mscM5BlackWidowOracle: "Origin of Black Widow",
		mscM5IronManOracle:    "Origin of Iron Man",
		mscM5ThorOracle:       "Origin of Thor",
		mscM5HulkOracle:       "Origin of the Hulk",
		mscM5MetalworkOracle:  "Metalwork Colossus",
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

func mscM5Card(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// mscM5Library gives a seat a library deep enough for the draws a test makes.
func mscM5Library(p *game.Player) {
	names := make([]string, mscM5DefaultLibrarySize)
	for i := range names {
		names[i] = "Library Card"
	}
	seedLibrary(p, names...)
}

// --- Spaceshift ----------------------------------------------------------------

// The target returns as a new object under its OWNER's control (not the
// controller who held it) with a +1/+1 counter on it.
func TestMscM5SpaceshiftReturnsItUnderItsOwnerWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Stolen Bear", "Creature — Bear", 2, 2)
	battlefieldCardPtr(t, g, bear).Controller = me.ID

	castCatalogSpell(t, g, "Spaceshift", "Instant", mscM5SpaceshiftOracle, []game.TargetRef{mscM5Card(bear)})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Fatal("the old object is still on the battlefield")
	}
	back := findBattlefieldByName(g, "Stolen Bear")
	if back == uuid.Nil {
		t.Fatal("the Bear did not return")
	}
	if c := battlefieldCardPtr(t, g, back); c.Controller != opp.ID {
		t.Errorf("the Bear returned under %s, want its owner %s", c.Controller, opp.ID)
	}
	if got := plusOnes(g, back); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if c := s58p6View(t, g, back); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// An artifact is a legal target and comes back with the counter too.
func TestMscM5SpaceshiftTargetsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := pushPermanent(g, me.ID, game.Card{Name: "Spare Relic", TypeLine: "Artifact"})
	castCatalogSpell(t, g, "Spaceshift", "Instant", mscM5SpaceshiftOracle, []game.TargetRef{mscM5Card(relic)})
	passPriorityAroundTable(t, g)
	back := findBattlefieldByName(g, "Spare Relic")
	if back == uuid.Nil || back == relic {
		t.Fatalf("the artifact did not return as a new object (id %s)", back)
	}
	if got := plusOnes(g, back); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// A land is neither an artifact nor a creature.
func TestMscM5SpaceshiftRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushPermanent(g, me.ID, game.Card{Name: "Spare Forest", TypeLine: "Basic Land — Forest"})
	err := castCatalogSpellErr(t, g, "Spaceshift", "Instant", mscM5SpaceshiftOracle, []game.TargetRef{mscM5Card(land)})
	if err == nil {
		t.Fatal("Spaceshift accepted a land as its target")
	}
}

// --- Return of the Mole Man ----------------------------------------------------

func mscM5MoleMan(g *game.Game, me *game.Player) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, "Return of the Mole Man", "Enchantment", mscM5MoleManOracle, false)
}

func TestMscM5MoleManLandfallMillsTwoWhenYouSayYes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mscM5Library(me)
	mscM5MoleMan(g, me)
	lib := len(me.Library.Cards)

	playLandFromHand(t, g, "Forest", "")
	assistCardSettle(t, g, me.ID, nil, true)

	if got := len(me.Library.Cards); got != lib-2 {
		t.Errorf("library %d -> %d, want two milled", lib, got)
	}
	if got := len(me.Graveyard.Cards); got != 2 {
		t.Errorf("graveyard has %d cards, want 2", got)
	}
}

func TestMscM5MoleManLandfallCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mscM5Library(me)
	mscM5MoleMan(g, me)
	lib := len(me.Library.Cards)

	playLandFromHand(t, g, "Forest", "")
	assistCardSettle(t, g, me.ID, nil, false)

	if got := len(me.Library.Cards); got != lib {
		t.Errorf("library %d -> %d after declining, want unchanged", lib, got)
	}
}

// X counts the permanent cards in the graveyard as the ability resolves, and
// that includes the sacrificed enchantment itself; an instant does not count.
func TestMscM5MoleManMakesAMoloidPerPermanentCardInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mole := mscM5MoleMan(g, me)
	for _, c := range []game.Card{
		{Name: "Dead Bear", TypeLine: "Creature — Bear"},
		{Name: "Dead Relic", TypeLine: "Artifact"},
		{Name: "Dead Bolt", TypeLine: "Instant"},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Graveyard.PushTop(c)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{G}")
	b16Activate(t, g, me.ID, mole, 0, game.ActivateAbilityParams{})

	if g.Battlefield.Contains(mole) {
		t.Error("the enchantment was not sacrificed")
	}
	if got := mshB1Tokens(g, me.ID, "Minion"); got != 3 {
		t.Fatalf("%d Moloid tokens, want 3 (two permanent cards and the enchantment)", got)
	}
	if tok := mshB1Token(t, g, me.ID, "Minion"); tok.Name != "Moloid" || tok.Power != 1 || !tok.HasColor("G") {
		t.Errorf("token is %+v, want a 1/1 green Moloid", tok)
	}
}

func TestMscM5MoleManSacrificeIsSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mole := mscM5MoleMan(g, me)
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{G}")
	err := g.ActivateCatalogAbility(me.ID, mole, 0, game.ActivateAbilityParams{Strict: true})
	if err == nil {
		t.Fatal("the sacrifice ability was activated in the beginning of combat step")
	}
	if !g.Battlefield.Contains(mole) {
		t.Error("the enchantment was sacrificed by a refused activation")
	}
}

func TestMscM5MoleManSacrificeNeedsItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mole := mscM5MoleMan(g, me)
	floatMana(t, g, me, "{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, mole, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability was activated with three mana")
	}
}

// --- Molly Hayes, Runaway ------------------------------------------------------

func TestMscM5MollyPowersUpOnceForTwoCountersAndAnImpulseExile(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	mscM5Library(me)
	top := seedLibraryTop(me, "Top Card", "Instant", "")
	molly := castPowerUpCreature(t, g, me, "Molly Hayes, Runaway", mscM5PowerUpTypeLine, "{2}{R}", "{C}{C}{R}", mscM5MollyOracle, 3)

	if got := chargedPowerUp(t, g, me.ID, molly); got != "{3}" {
		t.Fatalf("the turn she entered: %q, want {3}", got)
	}
	floatMana(t, g, me, "{C}{C}{C}")
	b16Activate(t, g, me.ID, molly, 0, game.ActivateAbilityParams{})
	if got := plusOnes(g, molly); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Fatal("the exiled top card is not playable by me")
	}
	if b12ExiledPlayableBy(t, g, top, opp.ID) {
		t.Error("the exiled top card is playable by an opponent")
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, molly, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}

	powerUpPassTo(t, g, 0)
	if !b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the card is not playable on my next turn")
	}
	powerUpPassTo(t, g, 0)
	if b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the card is still playable after my next turn ended")
	}
}

// --- Namora, the Sea Queen -----------------------------------------------------

func TestMscM5NamoraPowersUpForACounterAndTwoMerfolk(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	namora := castPowerUpCreature(t, g, me, "Namora, the Sea Queen", mscM5NamoraTypeLine, "{2}{U}", "{C}{C}{U}", mscM5NamoraOracle, 2)
	assertKeywords(t, g, namora, "flying")

	if got := chargedPowerUp(t, g, me.ID, namora); got != "{3}" {
		t.Fatalf("the turn she entered: %q, want {3}", got)
	}
	floatMana(t, g, me, "{C}{C}{C}")
	b16Activate(t, g, me.ID, namora, 0, game.ActivateAbilityParams{})
	if got := plusOnes(g, namora); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if got := mshB1Tokens(g, me.ID, "Merfolk"); got != 2 {
		t.Fatalf("%d Merfolk tokens, want 2", got)
	}
	if tok := mshB1Token(t, g, me.ID, "Merfolk"); tok.Power != 1 || tok.Toughness != 1 || !tok.HasColor("U") {
		t.Errorf("token is %+v, want a 1/1 blue Merfolk", tok)
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, namora, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

// --- She-Hulk, Attorney-at-Law -------------------------------------------------

func TestMscM5SheHulkDoublesEveryCounterYouControlAfterHerOwn(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	she := pushPowerUpPermanent(g, me.ID, "She-Hulk, Attorney-at-Law", mscM5SheHulkTypeLine, "{2}{G/W}", mscM5SheHulkOracle)
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	bare := b12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.AddCounterBy(me.ID, ally, game.CounterPlusOne, 2); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounterBy(opp.ID, theirs, game.CounterPlusOne, 3); err != nil {
		t.Fatal(err)
	}

	if got := chargedPowerUp(t, g, me.ID, she); got != "{6}{G/W}" {
		t.Fatalf("a turn after she entered: %q, want {6}{G/W}", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{G}")
	b16Activate(t, g, me.ID, she, 0, game.ActivateAbilityParams{})

	for name, want := range map[string]struct {
		id uuid.UUID
		n  int
	}{
		"She-Hulk (1 then doubled)": {she, 2},
		"my creature with two":      {ally, 4},
		"my creature with none":     {bare, 0},
		"an opponent's creature":    {theirs, 3},
	} {
		if got := plusOnes(g, want.id); got != want.n {
			t.Errorf("%s: %d +1/+1 counters, want %d", name, got, want.n)
		}
	}
	assertKeywords(t, g, she, "vigilance")
}

// --- Midnight Angel Armor ------------------------------------------------------

func TestMscM5ArmorMakesASoldierAndAttachesToIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	armor := castCatalogSpell(t, g, "Midnight Angel Armor", "Artifact — Equipment", mscM5ArmorOracle, nil)
	passPriorityAroundTable(t, g)

	soldier := findBattlefieldByName(g, "Soldier")
	if soldier == uuid.Nil {
		t.Fatal("no Soldier token")
	}
	c := s58p6View(t, g, soldier)
	if c.Controller != me.ID || !c.IsToken() || !c.HasColorWord("W") {
		t.Errorf("the token is %+v, want a white Soldier token under my control", c)
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the Soldier is %d/%d, want 4/4 (1/1 plus +3/+3)", c.CurrentPower(), c.CurrentToughness())
	}
	assertKeywords(t, g, soldier, "flying", "vigilance")
	if a := findBattlefieldCardByID(g, armor); a == nil || !a.IsAttachedTo(soldier) {
		t.Error("the Armor should be attached to the Soldier")
	}
}

func TestMscM5ArmorEquipMovesTheBonusAndTheKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	armor := castCatalogSpell(t, g, "Midnight Angel Armor", "Artifact — Equipment", mscM5ArmorOracle, nil)
	passPriorityAroundTable(t, g)
	soldier := findBattlefieldByName(g, "Soldier")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	equipTo(t, g, me.ID, armor, bear)

	if c := s58p6View(t, g, bear); c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("the equipped Bear is %d/%d, want 5/5", c.CurrentPower(), c.CurrentToughness())
	}
	assertKeywords(t, g, bear, "flying", "vigilance")
	if c := s58p6View(t, g, soldier); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the Soldier is %d/%d after the Armor moved, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	if containsString(effectiveAbilities(t, g, soldier), "flying") {
		t.Error("the Soldier kept flying after the Armor moved")
	}
}

// --- Metalwork Colossus --------------------------------------------------------

func TestMscM5MetalworkCostsLessByTheNoncreatureArtifactsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	if got := selfPricedMV(t, g, me, mscM5MetalworkOracle, mscM5MetalworkTypeLine, "{11}"); got != 11 {
		t.Fatalf("with no artifacts: %d, want 11", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}"})
	pushPermanent(g, me.ID, game.Card{Name: "Gem", TypeLine: "Artifact", ManaCost: "{2}"})
	pushPermanent(g, me.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{4}", Power: 3, Toughness: 3})
	pushPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{5}"})
	if got := selfPricedMV(t, g, me, mscM5MetalworkOracle, mscM5MetalworkTypeLine, "{11}"); got != 6 {
		t.Errorf("with {3} and {2} of artifacts (a creature and an opponent's do not count): %d, want 6", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Huge Relic", TypeLine: "Artifact", ManaCost: "{20}"})
	if got := selfPricedMV(t, g, me, mscM5MetalworkOracle, mscM5MetalworkTypeLine, "{11}"); got != 0 {
		t.Errorf("with more reduction than the cost: %d, want 0", got)
	}
}

func mscM5MetalworkInGraveyard(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Metalwork Colossus", TypeLine: mscM5MetalworkTypeLine,
		OracleID: mscM5MetalworkOracle, Owner: me.ID, Controller: me.ID})
	return id
}

func TestMscM5MetalworkReturnsFromTheGraveyardForTwoArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	colossus := mscM5MetalworkInGraveyard(me)
	a := pushPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	b := pushPermanent(g, me.ID, game.Card{Name: "Gem", TypeLine: "Artifact"})
	keep := pushPermanent(g, me.ID, game.Card{Name: "Bystander", TypeLine: "Artifact"})

	b16Activate(t, g, me.ID, colossus, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}})

	if !me.Hand.Contains(colossus) {
		t.Fatal("the Colossus did not return to hand")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the two artifacts were not sacrificed")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("a third artifact was sacrificed")
	}
}

func TestMscM5MetalworkNeedsTwoArtifactsToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	colossus := mscM5MetalworkInGraveyard(me)
	one := pushPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	err := g.ActivateCatalogAbility(me.ID, colossus, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{one}})
	if err == nil {
		t.Fatal("the ability was activated by sacrificing one artifact")
	}
	if me.Hand.Contains(colossus) || !g.Battlefield.Contains(one) {
		t.Error("a refused activation still moved cards")
	}
}

// --- the Origin sagas ----------------------------------------------------------

// mscM5CastSaga casts a Saga for the active seat and settles chapter I.
func mscM5CastSaga(t *testing.T, g *game.Game, name, oracle string, pick *game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	saga := castCatalogSpell(t, g, name, mscM5SagaTypeLine, oracle, nil)
	assistCardSettle(t, g, me.ID, pick, true)
	return saga
}

// mscM5NextChapter walks to seat 0's next precombat main and settles the
// chapter the lore counter fires.
func mscM5NextChapter(t *testing.T, g *game.Game, pick *game.TargetRef) {
	t.Helper()
	powerUpPassTo(t, g, 0)
	assistCardSettle(t, g, g.Seats[0].ID, pick, true)
}

func TestMscM5OriginOfTheHulkRunsItsThreeChapters(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pick := mscM5Card(bear)

	saga := mscM5CastSaga(t, g, "Origin of the Hulk", mscM5HulkOracle, nil)
	citizen := mshB1Token(t, g, me.ID, "Citizen")
	if citizen.Power != 1 || citizen.Toughness != 1 || !citizen.HasColor("G") || !citizen.HasColor("W") {
		t.Errorf("chapter I made %+v, want a 1/1 green and white Citizen", citizen)
	}

	mscM5NextChapter(t, g, &pick)
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("chapter II: +1/+1 counters = %d, want 2", got)
	}

	mscM5NextChapter(t, g, &pick)
	if c := s58p6View(t, g, bear); c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
		t.Errorf("chapter III: the Bear is %d/%d, want 7/7 (2/2, two counters, +3/+3)", c.CurrentPower(), c.CurrentToughness())
	}
	assertKeywords(t, g, bear, "trample")
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestMscM5OriginOfBlackWidowEdictDeathtouchAndDrain(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	var theirs []uuid.UUID
	for _, opp := range g.Seats[1:] {
		theirs = append(theirs, b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2))
	}

	saga := mscM5CastSaga(t, g, "Origin of Black Widow", mscM5BlackWidowOracle, nil)
	for i, id := range theirs {
		if g.Battlefield.Contains(id) {
			t.Errorf("opponent %d did not sacrifice their only creature", i+1)
		}
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was sacrificed")
	}

	theirsLater := b12Creature(g, g.Seats[1].ID, "Theirs Later", "Creature — Bear", 2, 2)
	mscM5NextChapter(t, g, nil)
	if !containsString(effectiveAbilities(t, g, mine), "deathtouch") {
		t.Error("chapter II: my creature lacks deathtouch")
	}
	if containsString(effectiveAbilities(t, g, theirsLater), "deathtouch") {
		t.Error("chapter II: an opponent's creature gained deathtouch")
	}

	for i, n := range []int{2, 1, 0} {
		for k := 0; k < n; k++ {
			g.Seats[i+1].Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear",
				Owner: g.Seats[i+1].ID, Controller: g.Seats[i+1].ID})
		}
	}
	g.Seats[1].Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Forest", TypeLine: "Basic Land — Forest",
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	life := []int{g.Seats[0].Life, g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}
	mscM5NextChapter(t, g, nil)
	for i, loss := range []int{0, 3, 2, 1} {
		if got := g.Seats[i].Life; got != life[i]-loss {
			t.Errorf("seat %d life %d -> %d, want %d lost", i, life[i], got, loss)
		}
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestMscM5OriginOfIronManHoldsTheCreatureWhileTheSagaRemains(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	mscM5Library(me)
	target := b12Creature(g, opp.ID, "Held Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Free Bear", "Creature — Bear", 2, 2)
	pick := mscM5Card(target)

	saga := mscM5CastSaga(t, g, "Origin of Iron Man", mscM5IronManOracle, &pick)
	if !b16Tapped(t, g, target) {
		t.Fatal("chapter I did not tap the target")
	}
	if b16Tapped(t, g, other) {
		t.Fatal("chapter I tapped a creature it did not target")
	}

	hand := len(me.Hand.Cards)
	// Seat 1's untap step passes on the way to my next turn.
	mscM5NextChapter(t, g, nil)
	if !b16Tapped(t, g, target) {
		t.Error("the held creature untapped while the Saga remained")
	}
	if got := len(me.Hand.Cards); got != hand+3 {
		t.Errorf("hand %d -> %d, want the turn's draw plus chapter II's two", hand, got)
	}

	// Chapter III: an artifact with mana value 5 goes in; 6 and a creature stay.
	cheap := uuid.New()
	pricey := uuid.New()
	body := uuid.New()
	for id, c := range map[uuid.UUID]game.Card{
		cheap:  {Name: "Cheap Relic", TypeLine: "Artifact", ManaCost: "{5}"},
		pricey: {Name: "Pricey Relic", TypeLine: "Artifact", ManaCost: "{6}"},
		body:   {Name: "Hand Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"},
	} {
		c.InstanceID, c.Owner, c.Controller = id, me.ID, me.ID
		me.Hand.PushTop(c)
	}
	powerUpPassTo(t, g, 0)
	pickPrompt := chooseCardsChoiceFor(g, me.ID)
	if pickPrompt == nil {
		passPriorityAroundTable(t, g)
		pickPrompt = chooseCardsChoiceFor(g, me.ID)
	}
	if pickPrompt == nil {
		t.Fatal("chapter III did not offer a put")
	}
	if !hasID(pickPrompt.ChooseCards, cheap) || hasID(pickPrompt.ChooseCards, pricey) || hasID(pickPrompt.ChooseCards, body) {
		t.Errorf("candidates %v: want only the mana value 5 artifact", pickPrompt.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the artifact did not enter")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}

	// The Saga is gone: the creature untaps at its controller's next untap.
	powerUpPassTo(t, g, 1)
	if b16Tapped(t, g, target) {
		t.Error("the creature stayed tapped after the Saga left")
	}
}

func TestMscM5OriginOfIronManChapterOneAllowsNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Origin of Iron Man", mscM5SagaTypeLine, mscM5IronManOracle, nil)
	passPriorityAroundTable(t, g)
	p := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(p.ID, g.Seats[0].ID, nil); err != nil {
		t.Fatalf("choose no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, bear) {
		t.Error("a creature was tapped though no target was chosen")
	}
}

func TestMscM5OriginOfThorRummagesThenCountersEachSpellThenDamages(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mscM5Library(me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pick := mscM5Card(bear)
	junk := handCardForTest(me, "Junk", "Sorcery", "")

	castCatalogSpell(t, g, "Origin of Thor", mscM5SagaTypeLine, mscM5ThorOracle, nil)
	passPriorityAroundTable(t, g)
	hand := len(me.Hand.Cards)
	answerDiscard(t, g, me.ID, junk)
	assistCardSettle(t, g, me.ID, nil, true)
	if got := len(me.Hand.Cards); got != hand-1+2 {
		t.Errorf("chapter I: hand %d -> %d, want a discard and two draws", hand, got)
	}

	powerUpPassTo(t, g, 0)
	assistCardSettle(t, g, me.ID, &pick, true)
	for i := 0; i < 2; i++ {
		castFree(t, g, me, "Free Spell", "Sorcery")
		assistCardSettle(t, g, me.ID, &pick, true)
	}
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("chapter II: %d counters after two spells, want 2", got)
	}

	// On the next turn the "this turn" window of chapter II is over.
	powerUpPassTo(t, g, 0)
	assistCardSettle(t, g, me.ID, &pick, true)
	castFree(t, g, me, "Late Spell", "Sorcery")
	assistCardSettle(t, g, me.ID, &pick, true)
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("a spell cast after chapter II's turn added a counter (%d, want 2)", got)
	}
}

func TestMscM5OriginOfThorChapterThreeDamagesEachOpponentByPower(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mscM5Library(me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pick := mscM5Card(bear)

	castCatalogSpell(t, g, "Origin of Thor", mscM5SagaTypeLine, mscM5ThorOracle, nil)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	assistCardSettle(t, g, me.ID, nil, true)
	powerUpPassTo(t, g, 0)
	assistCardSettle(t, g, me.ID, &pick, true)
	if err := g.AddCounterBy(me.ID, bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}

	life := []int{g.Seats[0].Life, g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}
	mscM5NextChapter(t, g, &pick)
	for i, loss := range []int{0, 3, 3, 3} {
		if got := g.Seats[i].Life; got != life[i]-loss {
			t.Errorf("seat %d life %d -> %d, want %d lost", i, life[i], got, loss)
		}
	}
}
