package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_cg2_cards_test.go — Final Fantasy (FIN), the C–G cards #5225 left
// (#4462), and the Chocobo Bird token four of them make. Each test plays
// the real catalog entry and asserts what its text says.

const (
	finCG2EvilReawakenedOracle = "264f944b-386b-4378-9764-4a8bde246809"
	finCG2CombatTutorialOracle = "2a483d95-ff99-40b4-9071-6e78585090d5"
	finCG2CapitalCityOracle    = "2c96ab90-155b-4bf4-acc9-65a2f0cd3189"
	finCG2MoogleOracle         = "f718b520-d79d-4b91-9d85-de1f1c90c176"
	finCG2DemonWallOracle      = "077b5ed2-05c2-4a16-b5eb-794884ee1c21"
	finCG2BeaversOracle        = "9108fffb-90f4-444e-8d3e-46a3d4ff3681"
	finCG2CactuarOracle        = "3e61d7c4-263f-4077-bdbc-d79020f2a479"
	finCG2DeadlyEmbraceOracle  = "32ad99a9-b5ed-41a1-8060-22265f7c5da1"
	finCG2SunCrystOracle       = "0ffeec93-89d7-42c2-8abf-802e9f6dc992"
	finCG2ElixirOracle         = "3e978db3-77ef-4b93-8f8f-e414e27254bd"
	finCG2OchuOracle           = "09b71c95-7b5d-4cee-a505-e0dda72715b7"
	finCG2GladiolusOracle      = "fb5a05d1-baa5-459d-a2e5-438f4a60b82f"
	finCG2LanceOracle          = "a66c35fc-d5db-4cfd-9b85-d70b001d83d3"
	finCG2CargoShipOracle      = "93e336e3-2be7-4581-8ab3-9bea422d81c3"
	finCG2GysahlOracle         = "eddfe7d7-6868-4b81-b46c-d0765e5210e7"
	finCG2ChocoCometOracle     = "07b28ed7-eab8-4560-9994-3befcba876d3"
	finCG2RacetrackOracle      = "7522f7fd-6fda-4ed3-9131-21bf91616a73"
	finCG2MountainCallOracle   = "a0dd3eb4-f6f1-4bfb-935c-032dd5af3811"
)

func TestFinCG2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finCG2EvilReawakenedOracle: "Evil Reawakened",
		finCG2CombatTutorialOracle: "Combat Tutorial",
		finCG2CapitalCityOracle:    "Capital City",
		finCG2MoogleOracle:         "Cloudbound Moogle",
		finCG2DemonWallOracle:      "Demon Wall",
		finCG2BeaversOracle:        "Commune with Beavers",
		finCG2CactuarOracle:        "Cactuar",
		finCG2DeadlyEmbraceOracle:  "Deadly Embrace",
		finCG2SunCrystOracle:       "Fate of the Sun-Cryst",
		finCG2ElixirOracle:         "Elixir",
		finCG2OchuOracle:           "Gran Pulse Ochu",
		finCG2GladiolusOracle:      "Gladiolus Amicitia",
		finCG2LanceOracle:          "Dragoon's Lance",
		finCG2CargoShipOracle:      "Cargo Ship",
		finCG2GysahlOracle:         "Gysahl Greens",
		finCG2ChocoCometOracle:     "Choco-Comet",
		finCG2RacetrackOracle:      "Chocobo Racetrack",
		finCG2MountainCallOracle:   "Call the Mountain Chocobo",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// finCG2Birds is the Chocobo Birds `controller` has, as the layers see
// them.
func finCG2Birds(t *testing.T, g *game.Game, controller uuid.UUID) []game.Card {
	t.Helper()
	var out []game.Card
	for _, c := range effectstest.SocB8Named(g, controller, "Bird") {
		out = append(out, effectstest.S58p6View(t, g, c.InstanceID))
	}
	return out
}

// finCG2OneBird asserts `controller` has exactly n Birds, each a green
// Bird creature token, and returns them.
func finCG2OneBird(t *testing.T, g *game.Game, controller uuid.UUID, n int) []game.Card {
	t.Helper()
	birds := finCG2Birds(t, g, controller)
	if len(birds) != n {
		t.Fatalf("%d Birds, want %d", len(birds), n)
	}
	for _, b := range birds {
		if !b.IsToken() || !b.IsCreature() || !b.HasSubtype("Bird") || len(b.Effective().Colors) != 1 || b.Effective().Colors[0] != "G" {
			t.Errorf("the Bird is %+v, want a green Bird creature token", b)
		}
	}
	return birds
}

// --- the Chocobo Bird token ---------------------------------------------------------

func TestFinCG2ChocoboBirdGetsPlusOneZeroOnEachLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Gysahl Greens", "Sorcery", finCG2GysahlOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	bird := finCG2OneBird(t, g, me.ID, 1)[0]
	if bird.CurrentPower() != 2 || bird.CurrentToughness() != 2 {
		t.Fatalf("the Bird is %d/%d, want 2/2", bird.CurrentPower(), bird.CurrentToughness())
	}
	if text := game.TokenTextForCard(bird); text != "Whenever a land you control enters, this token gets +1/+0 until end of turn." {
		t.Errorf("token text %q", text)
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.S58p6View(t, g, bird.InstanceID); v.CurrentPower() != 4 || v.CurrentToughness() != 2 {
		t.Errorf("after two lands the Bird is %d/%d, want 4/2", v.CurrentPower(), v.CurrentToughness())
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if v := effectstest.S58p6View(t, g, bird.InstanceID); v.CurrentPower() != 2 {
		t.Errorf("next turn the Bird has power %d, want 2", v.CurrentPower())
	}
}

// --- Gysahl Greens ------------------------------------------------------------------

func TestFinCG2GysahlGreensFlashesBackForASecondBird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	offers := game.AlternativeCostsOfferedFromZone(finCG2GysahlOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "flashback" || offers[0].ManaCost != "{6}{G}" || !offers[0].ExileOnLeavingStack {
		t.Fatalf("graveyard offers %+v, want flashback {6}{G} that exiles", offers)
	}
	id := effectstest.SeedGraveyardCard(t, g, "Gysahl Greens", "Sorcery", finCG2GysahlOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	finCG2OneBird(t, g, me.ID, 1)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back Greens are not in exile")
	}
}

// --- Choco-Comet --------------------------------------------------------------------

func TestFinCG2ChocoCometDealsXAndMakesABird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.CastXSpell(t, g, "Choco-Comet", "Sorcery", finCG2ChocoCometOracle, "{X}{R}{R}", 3,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent lost %d life, want 3", got)
	}
	finCG2OneBird(t, g, me.ID, 1)

	// X = 0 at a creature: no damage, still a Bird.
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastXSpell(t, g, "Choco-Comet", "Sorcery", finCG2ChocoCometOracle, "{X}{R}{R}", 0, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, bear); c == nil || c.DamageMarked != 0 {
		t.Error("X = 0 dealt damage")
	}
	finCG2OneBird(t, g, me.ID, 2)
}

// --- Chocobo Racetrack --------------------------------------------------------------

func TestFinCG2ChocoboRacetrackMakesABirdPerLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Chocobo Racetrack", "Artifact", finCG2RacetrackOracle, 0, 0)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	first := finCG2OneBird(t, g, me.ID, 1)[0]
	if first.CurrentPower() != 2 {
		t.Errorf("the land that made the Bird pumped it: power %d, want 2", first.CurrentPower())
	}
	// Two landfalls now, the Racetrack's and the Bird's: either order.
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	if effectstest.TriggerOrderPrompt(g) != nil {
		effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
	finCG2OneBird(t, g, me.ID, 2)
	if v := effectstest.S58p6View(t, g, first.InstanceID); v.CurrentPower() != 3 {
		t.Errorf("after the second land the first Bird has power %d, want 3", v.CurrentPower())
	}
}

// --- Call the Mountain Chocobo ------------------------------------------------------

func TestFinCG2CallTheMountainChocoboFindsAMountainThenMakesABird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for _, c := range []game.Card{
		{Name: "Forest", TypeLine: "Basic Land — Forest"},
		{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		{Name: "Stomping Ground", TypeLine: "Land — Mountain Forest"},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Library.PushTop(c)
	}
	effectstest.CastCatalogSpell(t, g, "Call the Mountain Chocobo", "Sorcery", finCG2MountainCallOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Stomping Ground") == uuid.Nil {
		t.Error("the search should offer any Mountain card and no other land")
	}
	if n := len(finCG2Birds(t, g, me.ID)); n != 0 {
		t.Fatalf("%d Birds before the search was answered", n)
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Mountain")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasNamed(me, "Mountain") {
		t.Error("the Mountain is not in hand")
	}
	finCG2OneBird(t, g, me.ID, 1)
	offers := game.AlternativeCostsOfferedFromZone(finCG2MountainCallOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "flashback" || offers[0].ManaCost != "{5}{R}" {
		t.Errorf("graveyard offers %+v, want flashback {5}{R}", offers)
	}
}

// --- Evil Reawakened ----------------------------------------------------------------

func TestFinCG2EvilReawakenedReturnsWithTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	theirs := effectstest.PushGraveyardCreature(g, opp.ID, "Their Bear", "{1}{G}")
	if err := effectstest.CastCatalogSpellErr(t, g, "Evil Reawakened", "Sorcery", finCG2EvilReawakenedOracle, effectstest.CardRefs(theirs)); err == nil {
		t.Error("a creature card in an opponent's graveyard was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Evil Reawakened", "Sorcery", finCG2EvilReawakenedOracle, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, mine)
	if c == nil {
		t.Fatal("the creature did not return")
	}
	if got := c.Counters[game.CounterPlusOne]; got != 2 || c.Controller != me.ID {
		t.Errorf("returned with %d +1/+1 counters under %s, want 2 under its owner", got, c.Controller)
	}
}

// --- Combat Tutorial ----------------------------------------------------------------

func TestFinCG2CombatTutorialDrawsTwoAndCountersYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "One", "Two", "Three")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	player := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	if err := effectstest.CastCatalogSpellErr(t, g, "Combat Tutorial", "Sorcery", finCG2CombatTutorialOracle,
		[]game.TargetRef{player, {Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("an opponent's creature was a legal target for the counter")
	}
	hand := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Combat Tutorial", "Sorcery", finCG2CombatTutorialOracle,
		[]game.TargetRef{player, {Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("the target player drew %d, want 2", got)
	}
	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Error("your creature did not get a +1/+1 counter")
	}

	// Up to one: the player alone is a legal cast.
	effectstest.CastCatalogSpell(t, g, "Combat Tutorial", "Sorcery", finCG2CombatTutorialOracle, []game.TargetRef{player})
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Capital City -------------------------------------------------------------------

func TestFinCG2CapitalCityTapsForColorlessFiltersAndCycles(t *testing.T) {
	spec, _ := effects.Lookup(finCG2CapitalCityOracle)
	if len(spec.ManaAbilities) != 2 ||
		spec.ManaAbilities[0].Produced != "{C}" || !spec.ManaAbilities[0].Cost.Tap ||
		spec.ManaAbilities[1].Produced != "{W|U|B|R|G}" || spec.ManaAbilities[1].Cost.Mana != "{1}" || !spec.ManaAbilities[1].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: {C} and {1}, {T}: any color", spec.ManaAbilities)
	}
	if spec.Replacements != nil {
		t.Error("Capital City should enter untapped")
	}

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	city := effectstest.PlayLandFromHand(t, g, "Capital City", finCG2CapitalCityOracle)
	before := me.ManaPool.CountColor("C")
	effectstest.ActivateManaFor(t, g, me.ID, city, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("C"); got != before+1 {
		t.Errorf("colorless in the pool %d → %d, want one more", before, got)
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.SeedLibrary(g2.Seats[g2.Turn.ActiveSeat], "Drawn")
	id, cycler := effectstest.CycleFromHand(t, g2, "Capital City", "Land — Town", finCG2CapitalCityOracle, "{C}{C}")
	hand := cycler.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g2)
	if !cycler.Graveyard.Contains(id) {
		t.Error("the cycled City is not in the graveyard")
	}
	if got := cycler.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d after cycling, want one card drawn", hand, got)
	}
}

// --- Cloudbound Moogle --------------------------------------------------------------

func TestFinCG2CloudboundMoogleFliesAndCountersATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	moogle := effectstest.CastRingCreature(t, g, "Cloudbound Moogle", "Creature — Moogle", finCG2MoogleOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, moogle, "flying") {
		t.Error("the Moogle has no flying")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, bear); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Error("the targeted creature (any creature) did not get a +1/+1 counter")
	}
	spec, _ := effects.Lookup(finCG2MoogleOracle)
	want := effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")).Label
	if len(spec.Activated) != 1 || spec.Activated[0].Label != want {
		t.Errorf("activated rows %+v, want plainscycling {2}", spec.Activated)
	}
}

// --- Demon Wall ---------------------------------------------------------------------

func TestFinCG2DemonWallAttacksOnceItHasACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B12Push(g, me.ID, "Demon Wall", "Artifact Creature — Demon Wall", finCG2DemonWallOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, wall, "menace") || !effectstest.HasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("Demon Wall lacks defender or menace")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Fatal("a Demon Wall with no counter attacked")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	wall2 := effectstest.B12Push(g2, me2.ID, "Demon Wall", "Artifact Creature — Demon Wall", finCG2DemonWallOracle, 3, 3)
	effectstest.AdvanceToMain(t, g2)
	effectstest.FloatMana(t, g2, me2, "{C}{C}{C}{C}{C}{B}")
	effectstest.B16Activate(t, g2, me2.ID, wall2, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	if v := effectstest.S58p6View(t, g2, wall2); v.CurrentPower() != 5 || v.CurrentToughness() != 5 {
		t.Fatalf("after the activation it is %d/%d, want 5/5", v.CurrentPower(), v.CurrentToughness())
	}
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(wall2, opp2.ID); err != nil {
		t.Fatalf("a Demon Wall with counters could not attack: %v", err)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g2, wall2, "defender") {
		t.Error("the waiver removed defender")
	}
}

// --- Commune with Beavers -----------------------------------------------------------

func TestFinCG2CommuneWithBeaversTakesAnArtifactCreatureOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cards := []struct{ Name, TypeLine string }{
		{"Rock", "Artifact"},
		{"Shock", "Instant"},
		{"Forest", "Basic Land — Forest"},
		{"Bear", "Creature — Bear"},
	}
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: ids[i], Name: cards[i].Name, TypeLine: cards[i].TypeLine, Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Commune with Beavers", "Sorcery", finCG2BeaversOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	// The top three only: the Bear is fourth.
	if len(offered) != 2 || !offered[ids[0]] || !offered[ids[2]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the artifact and the land, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the land is not in hand")
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
}

// --- Cactuar ------------------------------------------------------------------------

func TestFinCG2CactuarReturnsAtYourEndStepUnlessItEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	old := effectstest.B12Push(g, me.ID, "Cactuar", "Creature — Plant", finCG2CactuarOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, old, "trample") {
		t.Error("Cactuar has no trample")
	}
	fresh := effectstest.CastRingCreature(t, g, "Cactuar", "Creature — Plant", finCG2CactuarOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(old) || !me.Hand.Contains(old) {
		t.Error("the Cactuar that was already there did not return to hand")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("the Cactuar that entered this turn returned to hand")
	}
}

func TestFinCG2CactuarStaysOnAnotherPlayersEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Push(g, opp.ID, "Cactuar", "Creature — Plant", finCG2CactuarOracle, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("Cactuar returned at another player's end step")
	}
}

// --- Deadly Embrace -----------------------------------------------------------------

func TestFinCG2DeadlyEmbraceDestroysThenDrawsPerDeathThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	early := effectstest.B12Creature(g, opp.ID, "Early Bear", "Creature — Bear", 2, 2)
	target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Deadly Embrace", "Sorcery", finCG2DeadlyEmbraceOracle, effectstest.CardRefs(mine)); err == nil {
		t.Error("your own creature was a legal target")
	}
	effectstest.B18Kill(t, g, early)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Deadly Embrace", "Sorcery", finCG2DeadlyEmbraceOracle, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Error("the target survived")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (the earlier death and the target)", got)
	}
}

// --- Fate of the Sun-Cryst ----------------------------------------------------------

func TestFinCG2FateOfTheSunCrystCostsLessAtATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tired := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Tired", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	fresh := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Fresh", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	rock := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Tapped Rock", TypeLine: "Artifact", Tapped: true})
	price := func(target uuid.UUID) int {
		return effectstest.SelfPriced(t, g, me, finCG2SunCrystOracle, "Instant", "{4}{W}", game.ZoneHand, effectstest.CardRefs(target)).ManaValue()
	}
	if got := price(tired); got != 3 {
		t.Errorf("at a tapped creature: %d, want 3", got)
	}
	if got := price(fresh); got != 5 {
		t.Errorf("at an untapped creature: %d, want 5", got)
	}
	if got := price(rock); got != 5 {
		t.Errorf("at a tapped artifact: %d, want 5", got)
	}
	effectstest.CastCatalogSpell(t, g, "Fate of the Sun-Cryst", "Instant", finCG2SunCrystOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the nonland permanent survived")
	}
	land := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if err := effectstest.CastCatalogSpellErr(t, g, "Fate of the Sun-Cryst", "Instant", finCG2SunCrystOracle, effectstest.CardRefs(land)); err == nil {
		t.Error("a land was a legal target")
	}
}

// --- Elixir -------------------------------------------------------------------------

func TestFinCG2ElixirEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	elixir := effectstest.CastCatalogSpell(t, g, "Elixir", "Artifact", finCG2ElixirOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, elixir); c == nil || !c.Tapped {
		t.Error("Elixir did not enter tapped")
	}
}

func TestFinCG2ElixirShufflesNonlandCardsBackAndGainsThatMuchLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elixir := effectstest.B12Push(g, me.ID, "Elixir", "Artifact", finCG2ElixirOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.PushGraveyardCard(g, me, "Spent Spell")
	bear := effectstest.PushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	land := effectstest.GraveyardCardOf(g, me, "Old Forest", "Basic Land — Forest", "", "")
	lib, life := me.Library.Size(), me.Life
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, elixir, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Library.Contains(spell) || !me.Library.Contains(bear) {
		t.Error("a nonland card was not shuffled into the library")
	}
	if !me.Graveyard.Contains(land) {
		t.Error("the land left the graveyard")
	}
	if got := me.Library.Size() - lib; got != 2 {
		t.Errorf("library grew by %d, want 2", got)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	if !g.Exile.Contains(elixir) {
		t.Error("Elixir was not exiled as the cost")
	}
}

// --- Gran Pulse Ochu ----------------------------------------------------------------

func TestFinCG2GranPulseOchuGrowsByPermanentCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ochu := effectstest.B12Push(g, me.ID, "Gran Pulse Ochu", "Creature — Plant Beast", finCG2OchuOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, ochu, "deathtouch") {
		t.Error("the Ochu has no deathtouch")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.PushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	effectstest.GraveyardCardOf(g, me, "Old Forest", "Basic Land — Forest", "", "")
	effectstest.PushGraveyardCard(g, me, "Spent Spell")
	effectstest.PushGraveyardCreature(g, opp.ID, "Their Bear", "{1}{G}")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, ochu, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.S58p6View(t, g, ochu); v.CurrentPower() != 3 || v.CurrentToughness() != 3 {
		t.Fatalf("it is %d/%d, want 3/3 (two permanent cards in your graveyard)", v.CurrentPower(), v.CurrentToughness())
	}
	// The count was taken as the ability resolved.
	effectstest.PushGraveyardCreature(g, me.ID, "Later Bear", "{1}{G}")
	if v := effectstest.S58p6View(t, g, ochu); v.CurrentPower() != 3 {
		t.Errorf("a later graveyard card changed the bonus: power %d", v.CurrentPower())
	}
}

// --- Gladiolus Amicitia -------------------------------------------------------------

func TestFinCG2GladiolusFetchesALandTappedAndPumpsAnotherOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	glad := effectstest.CastRingCreature(t, g, "Gladiolus Amicitia", "Legendary Creature — Human Warrior", finCG2GladiolusOracle, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, forest); c == nil || !c.Tapped {
		t.Fatal("the land is not on the battlefield tapped")
	}
	// The fetched land is a landfall: Gladiolus itself is not a legal target.
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the landfall")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: glad}); err == nil {
		t.Fatal("Gladiolus was a legal target for its own landfall")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.S58p6View(t, g, bear); v.CurrentPower() != 4 || v.CurrentToughness() != 4 {
		t.Errorf("the other creature is %d/%d, want 4/4", v.CurrentPower(), v.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the other creature has no trample")
	}
}

// --- Dragoon's Lance ----------------------------------------------------------------

func TestFinCG2DragoonsLanceMakesAKnightHeroThatFliesOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	lance := effectstest.CastCatalogSpell(t, g, "Dragoon's Lance", "Artifact — Equipment", finCG2LanceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.YsBHero(t, g, lance)
	finMPHeroPT(t, g, hero, 2, 1, "Knight")
	finMPEquipLabel(t, finCG2LanceOracle, "{4}")
	if !effectstest.HasEffectiveKeyword(t, g, hero, "flying") {
		t.Error("on your turn the Hero has no flying")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, hero, "flying") {
		t.Error("on another player's turn the Hero has flying")
	}
	finMPHeroPT(t, g, hero, 2, 1, "Knight")
}

// --- Cargo Ship ---------------------------------------------------------------------

func TestFinCG2CargoShipTapsForArtifactManaAndCrewsWithOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ship := effectstest.B12Push(g, me.ID, "Cargo Ship", "Artifact — Vehicle", finCG2CargoShipOracle, 2, 3)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, ship, kw) {
			t.Errorf("Cargo Ship has no %s", kw)
		}
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, ship, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Ship: %v", err)
	}
	cost, _ := game.ParseCost("{1}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) || !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("the mana refused an artifact spell or an artifact's ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) || me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana paid for a nonartifact spell or ability")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	ship2 := effectstest.B12Push(g2, me2.ID, "Cargo Ship", "Artifact — Vehicle", finCG2CargoShipOracle, 2, 3)
	crewer := effectstest.B12Creature(g2, me2.ID, "Squire", "Creature — Human", 1, 1)
	effectstest.AdvanceToMain(t, g2)
	if c, _ := effectstest.BattlefieldCard(g2, ship2); c.IsCreature() {
		t.Fatal("the Ship is a creature before it is crewed")
	}
	effectstest.CrewBy(t, g2, me2.ID, ship2, crewer)
	effectstest.PassPriorityAroundTable(t, g2)
	if v := effectstest.S58p6View(t, g2, ship2); !v.IsCreature() {
		t.Error("power 1 did not crew the Ship")
	}
}
