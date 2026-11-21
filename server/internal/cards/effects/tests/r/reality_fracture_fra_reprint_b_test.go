package testsr

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFraReprintBCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]struct {
		name         string
		completeness effects.Completeness
	}{
		effectstest.RfrbProteusStaff:      {"Proteus Staff", effects.CompletenessFull},
		effectstest.RfrbRestlessAnchorage: {"Restless Anchorage", effects.CompletenessFull},
		effectstest.RfrbRestlessSpire:     {"Restless Spire", effects.CompletenessFull},
		effectstest.RfrbSharkTyphoon:      {"Shark Typhoon", effects.CompletenessFull},
		effectstest.RfrbSkrelvsHive:       {"Skrelv's Hive", effects.CompletenessFull},
		effectstest.RfrbStaffStoryteller:  {"Staff of the Storyteller", effects.CompletenessFull},
		effectstest.RfrbSyntheticDestiny:  {"Synthetic Destiny", effects.CompletenessFull},
		effectstest.RfrbTetsuko:           {"Tetsuko Umezawa, Fugitive", effects.CompletenessFull},
		effectstest.RfrbWhiteSunsTwilight: {"White Sun's Twilight", effects.CompletenessFull},
	} {
		spec, ok := effects.Lookup(oracle)
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
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

// rfrbOrderTheRestOnTheBottom answers the put-in-library prompt the
// reveal's "rest on the bottom in any order" raises, when there is one
// (a single card is not a choice and asks nothing).
func rfrbOrderTheRestOnTheBottom(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	order := effectstest.PutInLibraryChoiceFor(g, chooser)
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
	effectstest.PassPriorityAroundTable(t, g)
}

// --- the Phyrexian Mite token ---------------------------------------

func TestPhyrexianMiteTokenIsToxicAndCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if err := (effects.CreateToken{Controller: opp.ID, Template: effects.PhyrexianMiteToken(), N: 1}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	id := effectstest.FindBattlefieldByName(g, "Phyrexian Mite")
	if id == uuid.Nil {
		t.Fatal("no Mite")
	}
	c := effectstest.S58p6View(t, g, id)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rfrbPush(g, me.ID, "Skrelv's Hive", "Enchantment", effectstest.RfrbSkrelvsHive, 0, 0)
	life := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, me.Life)
	}
	if n := effectstest.CountBattlefieldByName(g, "Phyrexian Mite"); n != 1 {
		t.Errorf("%d Mites, want 1", n)
	}
}

func TestSkrelvsHiveGivesToxicCreaturesLifelinkOnlyWhenCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfrbPush(g, me.ID, "Skrelv's Hive", "Enchantment", effectstest.RfrbSkrelvsHive, 0, 0)
	g.WithWriteLock(func() {
		_ = (effects.CreateToken{Controller: me.ID, Template: effects.PhyrexianMiteToken(), N: 1}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	mite := effectstest.FindBattlefieldByName(g, "Phyrexian Mite")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if slices.Contains(effectstest.S58p6View(t, g, mite).Effective().Abilities, "lifelink") {
		t.Fatal("lifelink before an opponent has three poison counters")
	}
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 3); err != nil {
			t.Fatal(err)
		}
	})
	if !slices.Contains(effectstest.S58p6View(t, g, mite).Effective().Abilities, "lifelink") {
		t.Error("corrupted: the Mite has no lifelink")
	}
	if slices.Contains(effectstest.S58p6View(t, g, bear).Effective().Abilities, "lifelink") {
		t.Error("a creature without toxic got lifelink")
	}
}

// --- Tetsuko Umezawa, Fugitive ---------------------------------------

func TestTetsukoMakesSmallCreaturesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfrbPush(g, me.ID, "Tetsuko Umezawa, Fugitive", "Legendary Creature — Human Rogue", effectstest.RfrbTetsuko, 1, 3)
	small := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 4)
	thin := effectstest.B12Creature(g, me.ID, "Wisp", "Creature — Wisp", 4, 1)
	big := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.BrAttack(t, g, small, thin, big)

	offered := effectstest.BrOffered(t, g, opp.ID)
	if effectstest.BrOffers(offered, blocker, small) || effectstest.BrOffers(offered, blocker, thin) {
		t.Errorf("a power-or-toughness-1 creature is blockable: %v", offered)
	}
	if !effectstest.BrOffers(offered, blocker, big) {
		t.Errorf("a 2/2 should be blockable: %v", offered)
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(blocker, small), game.BlockReasonCantBeBlocked)
	if err := g.DeclareBlocker(blocker, big); err != nil {
		t.Fatalf("block the 2/2: %v", err)
	}
}

func TestTetsukoOnlyHelpsItsControllersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfrbPush(g, opp.ID, "Tetsuko Umezawa, Fugitive", "Legendary Creature — Human Rogue", effectstest.RfrbTetsuko, 1, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.BrAttack(t, g, mine)
	if err := g.DeclareBlocker(blocker, mine); err != nil {
		t.Fatalf("the opponent's Tetsuko shielded my creature: %v", err)
	}
}

// --- Restless Anchorage ------------------------------------------------

func rfrbAnimate(t *testing.T, g *game.Game, me *game.Player, land uuid.UUID, colors ...string) {
	t.Helper()
	for _, c := range colors {
		effectstest.FillPoolColored(me, c, 1)
	}
	rfrbActivate(t, g, me, land, game.ActivateAbilityParams{})
}

func TestRestlessAnchorageEntersTappedAndBecomesABird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if id := effectstest.B12PlayFromHand(t, g, "Restless Anchorage", "Land", effectstest.RfrbRestlessAnchorage, game.CastSpellParams{}); !effectstest.IsTapped(g, id) {
		t.Error("entered untapped")
	}
	land := effectstest.S58p6Push(g, me.ID, "Restless Anchorage", "Land", effectstest.RfrbRestlessAnchorage, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if effectstest.S58p6View(t, g, land).IsCreature() {
		t.Fatal("an idle land is a creature")
	}
	effectstest.FillPool(me, 1)
	rfrbAnimate(t, g, me, land, "W", "U")
	c := effectstest.S58p6View(t, g, land)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.S58p6Push(g, me.ID, "Restless Anchorage", "Land", effectstest.RfrbRestlessAnchorage, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPool(me, 1)
	rfrbAnimate(t, g, me, land, "W", "U")
	effectstest.DeclareAttack(t, g, opp.ID, land)
	if effectstest.TriggerOnStack(g, land) == nil {
		t.Fatal("no attack trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Map"); n != 1 {
		t.Errorf("%d Map tokens, want 1", n)
	}
}

// --- Restless Spire --------------------------------------------------

func TestRestlessSpireIsATwoOneWithFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.S58p6Push(g, me.ID, "Restless Spire", "Land", effectstest.RfrbRestlessSpire, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if slices.Contains(effectstest.S58p6View(t, g, land).Effective().Abilities, "first strike") {
		t.Fatal("an idle land has first strike")
	}
	rfrbAnimate(t, g, me, land, "U", "R")
	c := effectstest.S58p6View(t, g, land)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.S58p6Push(g, me.ID, "Restless Spire", "Land", effectstest.RfrbRestlessSpire, "", 0, 0)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Top"})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	rfrbAnimate(t, g, me, land, "U", "R")
	effectstest.DeclareAttack(t, g, opp.ID, land)
	if effectstest.TriggerOnStack(g, land) == nil {
		t.Fatal("no attack trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) == 0 {
		t.Error("no scry prompt")
	}
}

// --- Staff of the Storyteller ---------------------------------------

func TestStaffOfTheStorytellerMakesASpiritCountsItAndDrawsOffTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Staff of the Storyteller", "Artifact", effectstest.RfrbStaffStoryteller, nil)
	effectstest.PassPriorityAroundTable(t, g)
	staff := effectstest.FindBattlefieldByName(g, "Staff of the Storyteller")
	if staff == uuid.Nil {
		t.Fatal("no Staff")
	}
	if effectstest.CountBattlefieldByName(g, "Spirit") != 1 {
		t.Fatalf("%d Spirits, want 1", effectstest.CountBattlefieldByName(g, "Spirit"))
	}
	if n := effectstest.MustBattlefieldCard(t, g, staff).Counters["story"]; n != 1 {
		t.Fatalf("%d story counters after the Spirit, want 1", n)
	}
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	hand := len(me.Hand.Cards)
	effectstest.FillPoolColored(me, "W", 1)
	rfrbActivate(t, g, me, staff, game.ActivateAbilityParams{})
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn", hand, len(me.Hand.Cards))
	}
	if n := effectstest.MustBattlefieldCard(t, g, staff).Counters["story"]; n != 0 {
		t.Errorf("%d story counters left, want 0", n)
	}
}

func TestStaffOfTheStorytellerNeedsACounterToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := rfrbPush(g, me.ID, "Staff of the Storyteller", "Artifact", effectstest.RfrbStaffStoryteller, 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with no story counter")
	}
}

func TestStaffOfTheStorytellerCountsOneTriggerPerBatchOfTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := rfrbPush(g, me.ID, "Staff of the Storyteller", "Artifact", effectstest.RfrbStaffStoryteller, 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.CastXSpell(t, g, "White Sun's Twilight", "Sorcery", effectstest.RfrbWhiteSunsTwilight, "{X}{W}{W}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.MustBattlefieldCard(t, g, staff).Counters["story"]; n != 1 {
		t.Errorf("%d story counters for one batch of three Mites, want 1", n)
	}
}

// --- White Sun's Twilight ------------------------------------------

func TestWhiteSunsTwilightGainsLifeAndMakesMites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.CastXSpell(t, g, "White Sun's Twilight", "Sorcery", effectstest.RfrbWhiteSunsTwilight, "{X}{W}{W}", 4, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, me.Life)
	}
	if n := effectstest.CountBattlefieldByName(g, "Phyrexian Mite"); n != 4 {
		t.Errorf("%d Mites, want 4", n)
	}
	if effectstest.MustBattlefieldCard(t, g, bear) == nil {
		t.Error("X=4 destroyed a creature")
	}
}

func TestWhiteSunsTwilightAtFiveDestroysOtherCreaturesButNotItsMites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	effectstest.CastXSpell(t, g, "White Sun's Twilight", "Sorcery", effectstest.RfrbWhiteSunsTwilight, "{X}{W}{W}", 5, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardByID(g, mine) != nil || effectstest.FindBattlefieldCardByID(g, theirs) != nil {
		t.Error("X=5 left a creature standing")
	}
	if n := effectstest.CountBattlefieldByName(g, "Phyrexian Mite"); n != 5 {
		t.Errorf("%d Mites, want all 5 spared", n)
	}
}

// --- Proteus Staff ------------------------------------------------------

func TestProteusStaffReplacesACreatureWithTheNextCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := rfrbPush(g, me.ID, "Proteus Staff", "Artifact", effectstest.RfrbProteusStaff, 0, 0)
	ox := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	// The library's top is its last element: a land, a land, then Drake.
	opp.Library.Cards = []game.Card{
		{InstanceID: uuid.New(), Name: "Deep Land", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID},
		{InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID},
		{InstanceID: uuid.New(), Name: "Top Land", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID},
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPool(me, 2)
	rfrbActivate(t, g, me, staff, game.ActivateAbilityParams{Targets: effectstest.CardTarget(ox)})
	rfrbOrderTheRestOnTheBottom(t, g, opp.ID)

	if effectstest.FindBattlefieldCardByID(g, ox) != nil {
		t.Error("the Ox is still on the battlefield")
	}
	drake := effectstest.FindBattlefieldByName(g, "Drake")
	if drake == uuid.Nil {
		t.Fatal("the Drake did not enter")
	}
	if got := effectstest.MustBattlefieldCard(t, g, drake).Controller; got != opp.ID {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := rfrbPush(g, me.ID, "Proteus Staff", "Artifact", effectstest.RfrbProteusStaff, 0, 0)
	ox := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(ox)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Synthetic Destiny --------------------------------------------------

func TestSyntheticDestinyExilesYourCreaturesThenReturnsThatManyFromTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Cub", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	// Top first: Land, Imp, Land, Drake, Wurm, ...
	me.Library.Cards = []game.Card{
		{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Creature — Wurm", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID},
		{InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID},
		{InstanceID: uuid.New(), Name: "Land B", TypeLine: "Land", Owner: me.ID, Controller: me.ID},
		{InstanceID: uuid.New(), Name: "Imp", TypeLine: "Creature — Imp", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID},
		{InstanceID: uuid.New(), Name: "Land A", TypeLine: "Land", Owner: me.ID, Controller: me.ID},
	}
	effectstest.CastCatalogSpell(t, g, "Synthetic Destiny", "Instant", effectstest.RfrbSyntheticDestiny, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.FindBattlefieldCardByID(g, a) != nil || effectstest.FindBattlefieldCardByID(g, b) != nil {
		t.Fatal("my creatures were not exiled")
	}
	if effectstest.FindBattlefieldCardByID(g, theirs) == nil {
		t.Fatal("the opponent's creature was exiled")
	}
	if effectstest.FindBattlefieldByName(g, "Imp") != uuid.Nil {
		t.Fatal("the creatures came back before the end step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Imp") == uuid.Nil || effectstest.FindBattlefieldByName(g, "Drake") == uuid.Nil {
		t.Error("the two revealed creatures did not enter")
	}
	if effectstest.FindBattlefieldByName(g, "Spare") != uuid.Nil {
		t.Error("a third creature entered; only two were exiled")
	}
	if effectstest.FindBattlefieldByName(g, "Land A") != uuid.Nil {
		t.Error("a land entered the battlefield")
	}
	if len(me.Library.Cards) != 3 {
		t.Errorf("library has %d cards, want 3 (two lands and the spare)", len(me.Library.Cards))
	}
}

// --- Shark Typhoon --------------------------------------------------------

func TestSharkTyphoonMakesASharkThatBigAsTheSpellsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rfrbPush(g, me.ID, "Shark Typhoon", "Enchantment", effectstest.RfrbSharkTyphoon, 0, 0)
	effectstest.CastXSpell(t, g, "Divination", "Sorcery", "", "{2}{U}", 0, nil)
	if n := effectstest.CountBattlefieldByName(g, "Shark"); n != 0 {
		t.Fatalf("%d Sharks before the trigger resolved", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	id := effectstest.FindBattlefieldByName(g, "Shark")
	if id == uuid.Nil {
		t.Fatal("no Shark")
	}
	c := effectstest.S58p6View(t, g, id)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Abilities, "flying") || !slices.Equal(c.EffectiveColors(), []string{"U"}) {
		t.Errorf("abilities %v colours %v", c.Effective().Abilities, c.EffectiveColors())
	}
}

func TestSharkTyphoonIgnoresCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rfrbPush(g, me.ID, "Shark Typhoon", "Enchantment", effectstest.RfrbSharkTyphoon, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Shark"); n != 0 {
		t.Errorf("%d Sharks off a creature spell", n)
	}
}

func TestSharkTyphoonCyclingMakesAnXByXShark(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogHandCard(me, "Shark Typhoon", "Enchantment", effectstest.RfrbSharkTyphoon)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{U}"); err != nil {
		t.Fatal(err)
	}
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand { // cycled card out, card drawn in
		t.Errorf("hand %d -> %d, want the same size", hand, len(me.Hand.Cards))
	}
	sh := effectstest.FindBattlefieldByName(g, "Shark")
	if sh == uuid.Nil {
		t.Fatal("no Shark")
	}
	if c := effectstest.S58p6View(t, g, sh); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}
