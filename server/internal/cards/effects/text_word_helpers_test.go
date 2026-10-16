package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// text_word_helpers_test.go — #3872 slice 2: the words a card file hands
// a helper as an ARGUMENT, the words of a cost, and the words of a source
// that has left the battlefield all read as the object's text now says
// them (CR 612.2, ADR 0160's 2026-10-16 follow-up). One rules test per
// door (textWordDoors in text_word_guard_test.go) and per engine seam.

const (
	twhBloodMoonOracle    = "94fac5fe-97d5-4c12-a80c-8efff9d853ae"
	twhYavimayaOracle     = "8dd5f5af-d2d8-4356-8617-8381081b930c"
	twhHeroismOracle      = "74716642-fc8c-4f62-a556-154ed0b4af0e"
	twhKukemssaOracle     = "49e3ac82-7c22-4dec-b61a-b581148c0419"
	twhLilianaOracle      = "92301066-8904-41bc-84bd-4aa286cbab9c"
	twhCabalCoffersOracle = "7358e164-5704-4e78-9b21-6a9bf2a968ce"
	twhLeechriddenOracle  = "d83c86c1-126d-49e9-9b13-9e55784c49c5"
	twhUtopiaSprawlOracle = "00d8efa6-a2d9-4249-8da7-b45173675329"
	twhPollutedDelta      = "ef86989d-ce80-4e55-aece-7d11710eeffa"
)

// Magical Hack's indices: the five basic land types in CR 305.6 order,
// then the four left once the first answer is taken out.
const (
	twhPlains, twhIsland, twhSwamp, twhMountain, twhForest = 0, 1, 2, 3, 4
)

// twhNewType is the second prompt's index for `to` once `from` is gone.
func twhNewType(from, to int) int {
	if to > from {
		return to - 1
	}
	return to
}

// twhHack resolves Magical Hack on `target`, replacing land type `from`
// with `to` (CR 305.6 indices).
func twhHack(t *testing.T, g *game.Game, target uuid.UUID, from, to int) {
	t.Helper()
	twChange(t, g, "Magical Hack", twHackOracle, target, from, twhNewType(from, to))
}

// twhSleight resolves Sleight of Mind on `target`, replacing colour
// `from` with `to` (WUBRG indices).
func twhSleight(t *testing.T, g *game.Game, target uuid.UUID, from, to int) {
	t.Helper()
	twChange(t, g, "Sleight of Mind", twSleightOracle, target, from, twhNewType(from, to))
}

// Blood Moon's "Mountains" is a land-type word of its text: changed to
// "Forests", nonbasic lands are Forests and not Mountains (SetsBasicLandType
// reads its subtypes through the source).
func TestTextWordArgBloodMoonChangedToForestMakesNonbasicLandsForests(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	volcanic := pushLandFor(g, opp.ID, "Volcanic Island", "Land — Island Mountain")
	moon := pushCatalogPermanent(g, me.ID, "Blood Moon", "Enchantment", twhBloodMoonOracle, false)
	if c := gaLayered(t, g, volcanic); !c.HasSubtype("Mountain") || c.HasSubtype("Island") {
		t.Fatal("setup: under the printed Blood Moon a nonbasic land is a Mountain only")
	}
	twhHack(t, g, moon, twhMountain, twhForest)
	c := gaLayered(t, g, volcanic)
	if !c.HasSubtype("Forest") || c.HasSubtype("Mountain") || c.HasSubtype("Island") {
		t.Errorf("Blood Moon now reads \"nonbasic lands are Forests\": subtypes %v", c.Effective().Subtypes)
	}
}

// Yavimaya's "Each land is also a Forest" (EachLandIsAlso) changed to
// "an Island".
func TestTextWordArgYavimayaChangedToIslandMakesEachLandAnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	yavimaya := pushCatalogPermanent(g, me.ID, "Yavimaya, Cradle of Growth", "Legendary Land", twhYavimayaOracle, false)
	twhHack(t, g, yavimaya, twhForest, twhIsland)
	c := gaLayered(t, g, plains)
	if !c.HasSubtype("Island") || c.HasSubtype("Forest") {
		t.Errorf("each land is now also an Island: Plains has %v", c.Effective().Subtypes)
	}
}

// Heroism's "Sacrifice a white creature" under a Sleight of Mind (white →
// green) is "Sacrifice a green creature", and the validator, the move
// enumerator and the view all say so: the cost reads its words through
// the source although the walk is not targeting (costSourceLocked).
func TestTextWordArgHeroismCostChangedToGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	heroism := pushCatalogPermanent(g, me.ID, "Heroism", "Enchantment", twhHeroismOracle, false)
	knownToTable(g, heroism)
	white := seedColoredCreature(g, me.ID, "White Knight", []string{"W"}, 0)
	green := seedColoredCreature(g, me.ID, "Green Bear", []string{"G"}, 0)
	twhSleight(t, g, heroism, 0, 4) // white → green

	// The shared candidate walk (the view's options, the enumerator's pool).
	spec, ok := Lookup(twhHeroismOracle)
	if !ok {
		t.Fatal("Heroism is not in the catalog")
	}
	var cands []uuid.UUID
	g.WithWriteLock(func() {
		cands = g.SacrificeCostCandidatesForEffect(me.ID, heroism, spec.Activated[0].Cost.SacrificeOther)
	})
	if !twhHas(cands, green) || twhHas(cands, white) {
		t.Errorf("sacrifice candidates %v: want the green creature and not the white one", cands)
	}

	// The view's picker.
	view := protocol.ViewOfGameFor(g, me.ID.String())
	var opts []string
	for _, c := range view.Battlefield.Cards {
		if c.InstanceID == heroism.String() && len(c.ActivatedAbilities) > 0 && c.ActivatedAbilities[0].SacrificeOptions != nil {
			opts = c.ActivatedAbilities[0].SacrificeOptions.Cards
		}
	}
	if !twhHasString(opts, green.String()) || twhHasString(opts, white.String()) {
		t.Errorf("view sacrifice options %v: want the green creature and not the white one", opts)
	}

	// The enumerator's moves.
	offered := map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != "activate_ability" || m.Source != heroism {
			continue
		}
		var p struct {
			SacrificeIDs []uuid.UUID `json:"sacrifice_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		for _, id := range p.SacrificeIDs {
			offered[id] = true
		}
	}
	if !offered[green] || offered[white] {
		t.Errorf("enumerated sacrifices %v: want the green creature and not the white one", offered)
	}

	// The validator.
	if err := g.ActivateCatalogAbility(me.ID, heroism, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{white}}); err == nil {
		t.Error("a white creature no longer pays \"Sacrifice a green creature\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, heroism, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{green}}); err != nil {
		t.Errorf("a green creature pays it: %v", err)
	}
}

// Kukemssa Serpent under Magical Hack (Island → Forest): "{U}, Sacrifice
// a Forest: Target land an opponent controls becomes a Forest". Its
// sacrifice of the last Forest sets off "When you control no Forests",
// which sacrifices it before the ability resolves, so the ability reads
// the words its source LAST had (CR 608.2h).
func TestTextWordArgKukemssaSerpentChangedToForest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	island := pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	mountain := pushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	serpent := pushCatalogPermanent(g, me.ID, "Kukemssa Serpent", "Creature — Serpent", twhKukemssaOracle, false)
	twhHack(t, g, serpent, twhIsland, twhForest)

	activate := func(sac uuid.UUID) error {
		g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{U}") })
		return g.ActivateCatalogAbility(me.ID, serpent, 0, game.ActivateAbilityParams{
			SacrificeIDs: []uuid.UUID{sac},
			Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: mountain}},
		})
	}
	if err := activate(island); err == nil {
		t.Fatal("an Island no longer pays \"Sacrifice a Forest\"")
	}
	if err := activate(forest); err != nil {
		t.Fatalf("a Forest pays it: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if twOnBattlefield(g, serpent) {
		t.Fatal("setup: with no Forest left, the Serpent is sacrificed")
	}
	c := gaLayered(t, g, mountain)
	if !c.HasSubtype("Forest") || c.HasSubtype("Island") {
		t.Errorf("the land becomes a Forest, as its departed source's text said: subtypes %v", c.Effective().Subtypes)
	}
}

// Liliana of the Dark Realms changed to count Islands, whose −3 takes her
// last loyalty: she dies to the state-based action before the ability
// resolves, and X is still the number of Islands, read from her last-known
// text (CR 608.2h).
func TestTextWordArgLilianaChangedToIslandCountsIslandsAfterSheDies(t *testing.T) {
	g := newCatalogGame(t)
	me, lili := awbLiliana(t, g, 3)
	awbSwamps(g, me.ID, 1)
	for i := 0; i < 3; i++ {
		pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	}
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 5, 5)
	twhHack(t, g, lili, twhSwamp, twhIsland)

	if err := g.ActivateCatalogAbility(me.ID, lili, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if twOnBattlefield(g, lili) {
		t.Fatal("setup: Liliana with no loyalty is put into the graveyard")
	}
	answerOptionPick(t, g, me.ID, 0) // +X/+X
	if got := effectivePower(t, g, bear); got != 5+3 {
		t.Errorf("power %d, want 8: X counts the three Islands her text last named", got)
	}
}

// A fetchland changed by Magical Hack and sacrificed to its own ability
// searches for the type its text last named (the search lends the
// departed source's words).
func TestTextWordArgPollutedDeltaChangedToForestFetchesAForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	delta := pushCatalogPermanent(g, me.ID, "Polluted Delta", "Land", twhPollutedDelta, false)
	forestCard := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	islandCard := pushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	twhHack(t, g, delta, twhIsland, twhForest)
	if err := g.ActivateCatalogAbility(me.ID, delta, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if !twhHas(c.SearchCards, forestCard) || twhHas(c.SearchCards, islandCard) {
		t.Errorf("the search offers %v: want the Forest card and not the Island card", c.SearchCards)
	}
}

// Cabal Coffers' "for each Swamp you control" (ProducedPerPermanent over
// MatchLandSubtype) changed to Islands.
func TestTextWordArgCabalCoffersChangedToIslandCountsIslands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	coffers := pushCatalogPermanent(g, me.ID, "Cabal Coffers", "Land", twhCabalCoffersOracle, false)
	pushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	twhHack(t, g, coffers, twhSwamp, twhIsland)
	spec, _ := Lookup(twhCabalCoffersOracle)
	var amount game.ManaAmount
	found := false
	for _, ab := range spec.ManaAbilities {
		if ab.ProducedAmount != nil {
			g.WithWriteLock(func() { amount = ab.ProducedAmount(g, me.ID, coffers) })
			found = true
		}
	}
	if !found {
		t.Fatal("setup: Cabal Coffers has no scaled mana ability")
	}
	if amount.N != 2 {
		t.Errorf("Cabal Coffers now adds {B} for each Island: %d, want 2", amount.N)
	}
}

// Leechridden Swamp's "two or more black permanents" (ControlsAtLeast
// over MatchColor, an activation condition) changed to red.
func TestTextWordArgLeechriddenSwampChangedToRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := pushCatalogPermanent(g, me.ID, "Leechridden Swamp", "Land — Swamp", twhLeechriddenOracle, false)
	seedColoredCreature(g, me.ID, "Red One", []string{"R"}, 0)
	seedColoredCreature(g, me.ID, "Red Two", []string{"R"}, 0)
	spec, _ := Lookup(twhLeechriddenOracle)
	cond := func() bool {
		var ok bool
		g.WithWriteLock(func() { ok = spec.Activated[0].Condition(g, me.ID, swamp) })
		return ok
	}
	if cond() {
		t.Fatal("setup: two red permanents are not two black ones")
	}
	twhSleight(t, g, swamp, 2, 3) // black → red
	if !cond() {
		t.Error("\"two or more red permanents\": the condition holds")
	}
}

// Circle of Protection: Red under a Sleight of Mind (red → green) offers
// green sources, the shield's QueryColors read through its text; Story
// Circle's CHOSEN colour is not a word of its text and stays.
func TestTextWordArgCircleOfProtectionChangedToGreenAndStoryCircleUnchanged(t *testing.T) {
	t.Run("printed colour follows", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		cop := pushCatalogPermanent(g, me.ID, "Circle of Protection: Red", "Enchantment", pr7CoPRed, false)
		red := pr7Creature(g, opp.ID, "Red Dragon", 5, "R")
		green := pr7Creature(g, opp.ID, "Green Bear", 2, "G")
		twhSleight(t, g, cop, 3, 4) // red → green
		pr7Activate(t, g, me.ID, cop, 0, game.ActivateAbilityParams{})
		c := pr7SourcePrompt(t, g)
		if !pr7Offers(c, green) || pr7Offers(c, red) {
			t.Errorf("offers %v: want the green creature and not the red one", c.ChooseCards)
		}
	})
	t.Run("chosen colour stays", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		circle := pushCatalogPermanent(g, me.ID, "Story Circle", "Enchantment", pr7StoryCircle, false)
		findBattlefieldCardForTest(g, circle).ChosenColor = "B"
		black := pr7Creature(g, opp.ID, "Black", 3, "B")
		red := pr7Creature(g, opp.ID, "Red", 3, "R")
		twhSleight(t, g, circle, 2, 3) // black → red, which Story Circle never prints
		pr7Activate(t, g, me.ID, circle, 0, game.ActivateAbilityParams{})
		c := pr7SourcePrompt(t, g)
		if !pr7Offers(c, black) || pr7Offers(c, red) {
			t.Errorf("offers %v: the chosen colour is not text, so black only", c.ChooseCards)
		}
	})
}

// Utopia Sprawl's "Enchant Forest" is its own text: changed to "Enchant
// Island", an Aura on a Forest is attached illegally and is put into its
// owner's graveyard (CR 704.5m).
func TestTextWordArgUtopiaSprawlChangedToIslandFallsOffAForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	sprawl := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Utopia Sprawl", TypeLine: "Enchantment — Aura", OracleID: twhUtopiaSprawlOracle,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: forest}, ChosenColor: "G",
	})
	passPriorityAroundTable(t, g)
	if !twOnBattlefield(g, sprawl) {
		t.Fatal("setup: Utopia Sprawl on a Forest stays")
	}
	twhHack(t, g, sprawl, twhForest, twhIsland)
	passPriorityAroundTable(t, g)
	if twOnBattlefield(g, sprawl) {
		t.Error("\"Enchant Island\" on a Forest: the Aura is put into the graveyard")
	}
}

func twhHas(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func twhHasString(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
