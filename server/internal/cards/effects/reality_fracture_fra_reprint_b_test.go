package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_reprint_b_test.go — the older cards reprinted in
// Reality Fracture products, slice fra-reprint-b.

const (
	rfrbProteusStaff      = "97024e8e-dfde-4769-bc27-c3d6fad19e7c"
	rfrbRestlessAnchorage = "91320daf-f69c-4350-b0fc-4bb37a6904b1"
	rfrbRestlessSpire     = "0ca4e80e-c19c-4b74-b531-c5a4dc5a8ba9"
	rfrbSharkTyphoon      = "8c0520fa-276b-4d21-b4a9-dce1fce59f6b"
	rfrbSkrelvsHive       = "06d219ff-0083-4c2a-b5b3-2b84bb58f57e"
	rfrbStaffStoryteller  = "0c4e2c90-c17b-42cc-b4d7-cf75970fbe90"
	rfrbSyntheticDestiny  = "640a7d2c-42c4-4ea7-bd5b-72f12a65785e"
	rfrbTetsuko           = "ceeeacbc-01b0-4421-aaca-2ce6cdbe45d7"
	rfrbWhiteSunsTwilight = "a38828be-781e-4340-9c6f-f40b1d34773f"
)

func TestFraReprintBCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]struct {
		name         string
		completeness Completeness
	}{
		rfrbProteusStaff:      {"Proteus Staff", CompletenessFull},
		rfrbRestlessAnchorage: {"Restless Anchorage", CompletenessFull},
		rfrbRestlessSpire:     {"Restless Spire", CompletenessFull},
		rfrbSharkTyphoon:      {"Shark Typhoon", CompletenessCaveats},
		rfrbSkrelvsHive:       {"Skrelv's Hive", CompletenessFull},
		rfrbStaffStoryteller:  {"Staff of the Storyteller", CompletenessFull},
		rfrbSyntheticDestiny:  {"Synthetic Destiny", CompletenessFull},
		rfrbTetsuko:           {"Tetsuko Umezawa, Fugitive", CompletenessFull},
		rfrbWhiteSunsTwilight: {"White Sun's Twilight", CompletenessFull},
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want.name)
			continue
		}
		if spec.Name != want.name || spec.Completeness != want.completeness {
			t.Errorf("%s: name %q completeness %v", want.name, spec.Name, spec.Completeness)
		}
	}
}

func rfrbPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

// rfrbOrderTheRestOnTheBottom answers the put-in-library prompt the
// reveal's "rest on the bottom in any order" raises, when there is one
// (a single card is not a choice and asks nothing).
func rfrbOrderTheRestOnTheBottom(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	order := putInLibraryChoiceFor(g, chooser)
	if order == nil {
		return
	}
	if err := g.ResolvePutInLibrary(order.ID, chooser, order.ScryCards, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
}

func rfrbActivate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// --- the Phyrexian Mite token ---------------------------------------

func TestPhyrexianMiteTokenIsToxicAndCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if err := (CreateToken{Controller: opp.ID, Template: PhyrexianMiteToken(), N: 1}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	id := findBattlefieldByName(g, "Phyrexian Mite")
	if id == uuid.Nil {
		t.Fatal("no Mite")
	}
	c := s58p6View(t, g, id)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Phyrexian") {
		t.Errorf("type line %q", c.TypeLine)
	}
	if got := game.ToxicTotal(&c); got != 1 {
		t.Errorf("toxic %d, want 1", got)
	}
	if !game.Restricted(&c, game.CantBlock) {
		t.Error("the Mite can block")
	}
	if len(c.Colors) != 0 {
		t.Errorf("colours %v, want colourless", c.Colors)
	}
}

// --- Skrelv's Hive ------------------------------------------------------

func TestSkrelvsHiveMakesAMiteAndCostsALifeEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfrbPush(g, me.ID, "Skrelv's Hive", "Enchantment", rfrbSkrelvsHive, 0, 0)
	life := me.Life
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, me.Life)
	}
	if n := countBattlefieldByName(g, "Phyrexian Mite"); n != 1 {
		t.Errorf("%d Mites, want 1", n)
	}
}

func TestSkrelvsHiveGivesToxicCreaturesLifelinkOnlyWhenCorrupted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfrbPush(g, me.ID, "Skrelv's Hive", "Enchantment", rfrbSkrelvsHive, 0, 0)
	g.WithWriteLock(func() {
		_ = (CreateToken{Controller: me.ID, Template: PhyrexianMiteToken(), N: 1}).Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	mite := findBattlefieldByName(g, "Phyrexian Mite")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if slices.Contains(s58p6View(t, g, mite).Effective().Abilities, "lifelink") {
		t.Fatal("lifelink before an opponent has three poison counters")
	}
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 3); err != nil {
			t.Fatal(err)
		}
	})
	if !slices.Contains(s58p6View(t, g, mite).Effective().Abilities, "lifelink") {
		t.Error("corrupted: the Mite has no lifelink")
	}
	if slices.Contains(s58p6View(t, g, bear).Effective().Abilities, "lifelink") {
		t.Error("a creature without toxic got lifelink")
	}
}

// --- Tetsuko Umezawa, Fugitive ---------------------------------------

func TestTetsukoMakesSmallCreaturesUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfrbPush(g, me.ID, "Tetsuko Umezawa, Fugitive", "Legendary Creature — Human Rogue", rfrbTetsuko, 1, 3)
	small := b12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 4)
	thin := b12Creature(g, me.ID, "Wisp", "Creature — Wisp", 4, 1)
	big := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	brAttack(t, g, small, thin, big)

	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, blocker, small) || brOffers(offered, blocker, thin) {
		t.Errorf("a power-or-toughness-1 creature is blockable: %v", offered)
	}
	if !brOffers(offered, blocker, big) {
		t.Errorf("a 2/2 should be blockable: %v", offered)
	}
	brRefusal(t, g.DeclareBlocker(blocker, small), game.BlockReasonCantBeBlocked)
	if err := g.DeclareBlocker(blocker, big); err != nil {
		t.Fatalf("block the 2/2: %v", err)
	}
}

func TestTetsukoOnlyHelpsItsControllersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfrbPush(g, opp.ID, "Tetsuko Umezawa, Fugitive", "Legendary Creature — Human Rogue", rfrbTetsuko, 1, 3)
	mine := b12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	brAttack(t, g, mine)
	if err := g.DeclareBlocker(blocker, mine); err != nil {
		t.Fatalf("the opponent's Tetsuko shielded my creature: %v", err)
	}
}

// --- Restless Anchorage ------------------------------------------------

func rfrbAnimate(t *testing.T, g *game.Game, me *game.Player, land uuid.UUID, colors ...string) {
	t.Helper()
	for _, c := range colors {
		fillPoolColored(me, c, 1)
	}
	rfrbActivate(t, g, me, land, game.ActivateAbilityParams{})
}

func TestRestlessAnchorageEntersTappedAndBecomesABird(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if id := b12PlayFromHand(t, g, "Restless Anchorage", "Land", rfrbRestlessAnchorage, game.CastSpellParams{}); !isTapped(g, id) {
		t.Error("entered untapped")
	}
	land := s58p6Push(g, me.ID, "Restless Anchorage", "Land", rfrbRestlessAnchorage, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	if s58p6View(t, g, land).IsCreature() {
		t.Fatal("an idle land is a creature")
	}
	fillPool(me, 1)
	rfrbAnimate(t, g, me, land, "W", "U")
	c := s58p6View(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Bird") {
		t.Fatalf("type line %q", c.Effective().Subtypes)
	}
	if !slices.Equal(c.EffectiveColors(), []string{"W", "U"}) {
		t.Errorf("colours %v", c.EffectiveColors())
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want 2/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Abilities, "flying") {
		t.Errorf("no flying: %v", c.Effective().Abilities)
	}
}

func TestRestlessAnchorageAttackMakesAMap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := s58p6Push(g, me.ID, "Restless Anchorage", "Land", rfrbRestlessAnchorage, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPool(me, 1)
	rfrbAnimate(t, g, me, land, "W", "U")
	declareAttack(t, g, opp.ID, land)
	if triggerOnStack(g, land) == nil {
		t.Fatal("no attack trigger")
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Map"); n != 1 {
		t.Errorf("%d Map tokens, want 1", n)
	}
}

// --- Restless Spire --------------------------------------------------

func TestRestlessSpireIsATwoOneWithFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := s58p6Push(g, me.ID, "Restless Spire", "Land", rfrbRestlessSpire, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	if slices.Contains(s58p6View(t, g, land).Effective().Abilities, "first strike") {
		t.Fatal("an idle land has first strike")
	}
	rfrbAnimate(t, g, me, land, "U", "R")
	c := s58p6View(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Elemental") {
		t.Fatalf("subtypes %v", c.Effective().Subtypes)
	}
	if !slices.Equal(c.EffectiveColors(), []string{"U", "R"}) {
		t.Errorf("colours %v", c.EffectiveColors())
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 1 {
		t.Errorf("%d/%d, want 2/1", c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Abilities, "first strike") {
		t.Errorf("no first strike on my turn: %v", c.Effective().Abilities)
	}
}

func TestRestlessSpireAttackScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := s58p6Push(g, me.ID, "Restless Spire", "Land", rfrbRestlessSpire, "", 0, 0)
	pushLibraryCardForTest(me, game.Card{Name: "Top"})
	advanceTo(t, g, game.StepPrecombatMain)
	rfrbAnimate(t, g, me, land, "U", "R")
	declareAttack(t, g, opp.ID, land)
	if triggerOnStack(g, land) == nil {
		t.Fatal("no attack trigger")
	}
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) == 0 {
		t.Error("no scry prompt")
	}
}

// --- Staff of the Storyteller ---------------------------------------

func TestStaffOfTheStorytellerMakesASpiritCountsItAndDrawsOffTheCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Staff of the Storyteller", "Artifact", rfrbStaffStoryteller, nil)
	passPriorityAroundTable(t, g)
	staff := findBattlefieldByName(g, "Staff of the Storyteller")
	if staff == uuid.Nil {
		t.Fatal("no Staff")
	}
	if countBattlefieldByName(g, "Spirit") != 1 {
		t.Fatalf("%d Spirits, want 1", countBattlefieldByName(g, "Spirit"))
	}
	if n := mustBattlefieldCard(t, g, staff).Counters["story"]; n != 1 {
		t.Fatalf("%d story counters after the Spirit, want 1", n)
	}
	pushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	hand := len(me.Hand.Cards)
	fillPoolColored(me, "W", 1)
	rfrbActivate(t, g, me, staff, game.ActivateAbilityParams{})
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn", hand, len(me.Hand.Cards))
	}
	if n := mustBattlefieldCard(t, g, staff).Counters["story"]; n != 0 {
		t.Errorf("%d story counters left, want 0", n)
	}
}

func TestStaffOfTheStorytellerNeedsACounterToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	staff := rfrbPush(g, me.ID, "Staff of the Storyteller", "Artifact", rfrbStaffStoryteller, 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with no story counter")
	}
}

func TestStaffOfTheStorytellerCountsOneTriggerPerBatchOfTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	staff := rfrbPush(g, me.ID, "Staff of the Storyteller", "Artifact", rfrbStaffStoryteller, 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	castXSpell(t, g, "White Sun's Twilight", "Sorcery", rfrbWhiteSunsTwilight, "{X}{W}{W}", 3, nil)
	passPriorityAroundTable(t, g)
	if n := mustBattlefieldCard(t, g, staff).Counters["story"]; n != 1 {
		t.Errorf("%d story counters for one batch of three Mites, want 1", n)
	}
}

// --- White Sun's Twilight ------------------------------------------

func TestWhiteSunsTwilightGainsLifeAndMakesMites(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	castXSpell(t, g, "White Sun's Twilight", "Sorcery", rfrbWhiteSunsTwilight, "{X}{W}{W}", 4, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, me.Life)
	}
	if n := countBattlefieldByName(g, "Phyrexian Mite"); n != 4 {
		t.Errorf("%d Mites, want 4", n)
	}
	if mustBattlefieldCard(t, g, bear) == nil {
		t.Error("X=4 destroyed a creature")
	}
}

func TestWhiteSunsTwilightAtFiveDestroysOtherCreaturesButNotItsMites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	castXSpell(t, g, "White Sun's Twilight", "Sorcery", rfrbWhiteSunsTwilight, "{X}{W}{W}", 5, nil)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, mine) != nil || findBattlefieldCardByID(g, theirs) != nil {
		t.Error("X=5 left a creature standing")
	}
	if n := countBattlefieldByName(g, "Phyrexian Mite"); n != 5 {
		t.Errorf("%d Mites, want all 5 spared", n)
	}
}

// --- Proteus Staff ------------------------------------------------------

func TestProteusStaffReplacesACreatureWithTheNextCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := rfrbPush(g, me.ID, "Proteus Staff", "Artifact", rfrbProteusStaff, 0, 0)
	ox := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	// The library's top is its last element: a land, a land, then Drake.
	opp.Library.Cards = []game.Card{
		{InstanceID: uuid.New(), Name: "Deep Land", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID},
		{InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID},
		{InstanceID: uuid.New(), Name: "Top Land", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID},
	}
	advanceTo(t, g, game.StepPrecombatMain)
	fillPoolColored(me, "U", 1)
	fillPool(me, 2)
	rfrbActivate(t, g, me, staff, game.ActivateAbilityParams{Targets: cardTarget(ox)})
	rfrbOrderTheRestOnTheBottom(t, g, opp.ID)

	if findBattlefieldCardByID(g, ox) != nil {
		t.Error("the Ox is still on the battlefield")
	}
	drake := findBattlefieldByName(g, "Drake")
	if drake == uuid.Nil {
		t.Fatal("the Drake did not enter")
	}
	if got := mustBattlefieldCard(t, g, drake).Controller; got != opp.ID {
		t.Errorf("the Drake entered under %s, want the Ox's controller", got)
	}
	names := map[string]bool{}
	for _, c := range opp.Library.Cards {
		names[c.Name] = true
	}
	if !names["Ox"] || !names["Top Land"] || !names["Deep Land"] || len(opp.Library.Cards) != 3 {
		t.Errorf("library %v, want Ox and both lands", names)
	}
}

func TestProteusStaffIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := rfrbPush(g, me.ID, "Proteus Staff", "Artifact", rfrbProteusStaff, 0, 0)
	ox := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	advanceTo(t, g, game.StepBeginCombat)
	fillPoolColored(me, "U", 1)
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{Targets: cardTarget(ox)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Synthetic Destiny --------------------------------------------------

func TestSyntheticDestinyExilesYourCreaturesThenReturnsThatManyFromTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Cub", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	// Top first: Land, Imp, Land, Drake, Wurm, ...
	me.Library.Cards = []game.Card{
		{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Creature — Wurm", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID},
		{InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID},
		{InstanceID: uuid.New(), Name: "Land B", TypeLine: "Land", Owner: me.ID, Controller: me.ID},
		{InstanceID: uuid.New(), Name: "Imp", TypeLine: "Creature — Imp", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID},
		{InstanceID: uuid.New(), Name: "Land A", TypeLine: "Land", Owner: me.ID, Controller: me.ID},
	}
	castCatalogSpell(t, g, "Synthetic Destiny", "Instant", rfrbSyntheticDestiny, nil)
	passPriorityAroundTable(t, g)

	if findBattlefieldCardByID(g, a) != nil || findBattlefieldCardByID(g, b) != nil {
		t.Fatal("my creatures were not exiled")
	}
	if findBattlefieldCardByID(g, theirs) == nil {
		t.Fatal("the opponent's creature was exiled")
	}
	if findBattlefieldByName(g, "Imp") != uuid.Nil {
		t.Fatal("the creatures came back before the end step")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Imp") == uuid.Nil || findBattlefieldByName(g, "Drake") == uuid.Nil {
		t.Error("the two revealed creatures did not enter")
	}
	if findBattlefieldByName(g, "Spare") != uuid.Nil {
		t.Error("a third creature entered; only two were exiled")
	}
	if findBattlefieldByName(g, "Land A") != uuid.Nil {
		t.Error("a land entered the battlefield")
	}
	if len(me.Library.Cards) != 3 {
		t.Errorf("library has %d cards, want 3 (two lands and the spare)", len(me.Library.Cards))
	}
}

// --- Shark Typhoon --------------------------------------------------------

func TestSharkTyphoonMakesASharkThatBigAsTheSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfrbPush(g, me.ID, "Shark Typhoon", "Enchantment", rfrbSharkTyphoon, 0, 0)
	castXSpell(t, g, "Divination", "Sorcery", "", "{2}{U}", 0, nil)
	if n := countBattlefieldByName(g, "Shark"); n != 0 {
		t.Fatalf("%d Sharks before the trigger resolved", n)
	}
	passPriorityAroundTable(t, g)
	id := findBattlefieldByName(g, "Shark")
	if id == uuid.Nil {
		t.Fatal("no Shark")
	}
	c := s58p6View(t, g, id)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Abilities, "flying") || !slices.Equal(c.EffectiveColors(), []string{"U"}) {
		t.Errorf("abilities %v colours %v", c.Effective().Abilities, c.EffectiveColors())
	}
}

func TestSharkTyphoonIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfrbPush(g, me.ID, "Shark Typhoon", "Enchantment", rfrbSharkTyphoon, 0, 0)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Shark"); n != 0 {
		t.Errorf("%d Sharks off a creature spell", n)
	}
}

func TestSharkTyphoonCyclingMakesAnXByXShark(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	id := pushCatalogHandCard(me, "Shark Typhoon", "Enchantment", rfrbSharkTyphoon)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{U}"); err != nil {
		t.Fatal(err)
	}
	pushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand { // cycled card out, card drawn in
		t.Errorf("hand %d -> %d, want the same size", hand, len(me.Hand.Cards))
	}
	sh := findBattlefieldByName(g, "Shark")
	if sh == uuid.Nil {
		t.Fatal("no Shark")
	}
	if c := s58p6View(t, g, sh); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}
