package testst

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// twhSleight resolves Sleight of Mind on `target`, replacing colour
// `from` with `to` (WUBRG indices).
func twhSleight(t *testing.T, g *game.Game, target uuid.UUID, from, to int) {
	t.Helper()
	effectstest.TwChange(t, g, "Sleight of Mind", effectstest.TwSleightOracle, target, from, effectstest.TwhNewType(from, to))
}

// Blood Moon's "Mountains" is a land-type word of its text: changed to
// "Forests", nonbasic lands are Forests and not Mountains (SetsBasicLandType
// reads its subtypes through the source).
func TestTextWordArgBloodMoonChangedToForestMakesNonbasicLandsForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	volcanic := effectstest.PushLandFor(g, opp.ID, "Volcanic Island", "Land — Island Mountain")
	moon := effectstest.PushCatalogPermanent(g, me.ID, "Blood Moon", "Enchantment", effectstest.TwhBloodMoonOracle, false)
	if c := effectstest.GaLayered(t, g, volcanic); !c.HasSubtype("Mountain") || c.HasSubtype("Island") {
		t.Fatal("setup: under the printed Blood Moon a nonbasic land is a Mountain only")
	}
	effectstest.TwhHack(t, g, moon, effectstest.TwhMountain, effectstest.TwhForest)
	c := effectstest.GaLayered(t, g, volcanic)
	if !c.HasSubtype("Forest") || c.HasSubtype("Mountain") || c.HasSubtype("Island") {
		t.Errorf("Blood Moon now reads \"nonbasic lands are Forests\": subtypes %v", c.Effective().Subtypes)
	}
}

// Yavimaya's "Each land is also a Forest" (EachLandIsAlso) changed to
// "an Island".
func TestTextWordArgYavimayaChangedToIslandMakesEachLandAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	yavimaya := effectstest.PushCatalogPermanent(g, me.ID, "Yavimaya, Cradle of Growth", "Legendary Land", effectstest.TwhYavimayaOracle, false)
	effectstest.TwhHack(t, g, yavimaya, effectstest.TwhForest, effectstest.TwhIsland)
	c := effectstest.GaLayered(t, g, plains)
	if !c.HasSubtype("Island") || c.HasSubtype("Forest") {
		t.Errorf("each land is now also an Island: Plains has %v", c.Effective().Subtypes)
	}
}

// Heroism's "Sacrifice a white creature" under a Sleight of Mind (white →
// green) is "Sacrifice a green creature", and the validator, the move
// enumerator and the view all say so: the cost reads its words through
// the source although the walk is not targeting (costSourceLocked).
func TestTextWordArgHeroismCostChangedToGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	heroism := effectstest.PushCatalogPermanent(g, me.ID, "Heroism", "Enchantment", effectstest.TwhHeroismOracle, false)
	effectstest.KnownToTable(g, heroism)
	white := effectstest.SeedColoredCreature(g, me.ID, "White Knight", []string{"W"}, 0)
	green := effectstest.SeedColoredCreature(g, me.ID, "Green Bear", []string{"G"}, 0)
	twhSleight(t, g, heroism, 0, 4) // white → green

	// The shared candidate walk (the view's options, the enumerator's pool).
	spec, ok := effects.Lookup(effectstest.TwhHeroismOracle)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	mountain := effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	serpent := effectstest.PushCatalogPermanent(g, me.ID, "Kukemssa Serpent", "Creature — Serpent", effectstest.TwhKukemssaOracle, false)
	effectstest.TwhHack(t, g, serpent, effectstest.TwhIsland, effectstest.TwhForest)

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
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if twOnBattlefield(g, serpent) {
		t.Fatal("setup: with no Forest left, the Serpent is sacrificed")
	}
	c := effectstest.GaLayered(t, g, mountain)
	if !c.HasSubtype("Forest") || c.HasSubtype("Island") {
		t.Errorf("the land becomes a Forest, as its departed source's text said: subtypes %v", c.Effective().Subtypes)
	}
}

// Liliana of the Dark Realms changed to count Islands, whose −3 takes her
// last loyalty: she dies to the state-based action before the ability
// resolves, and X is still the number of Islands, read from her last-known
// text (CR 608.2h).
func TestTextWordArgLilianaChangedToIslandCountsIslandsAfterSheDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, lili := effectstest.AwbLiliana(t, g, 3)
	effectstest.AwbSwamps(g, me.ID, 1)
	for i := 0; i < 3; i++ {
		effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	}
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 5, 5)
	effectstest.TwhHack(t, g, lili, effectstest.TwhSwamp, effectstest.TwhIsland)

	if err := g.ActivateCatalogAbility(me.ID, lili, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if twOnBattlefield(g, lili) {
		t.Fatal("setup: Liliana with no loyalty is put into the graveyard")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // +X/+X
	if got := effectstest.EffectivePower(t, g, bear); got != 5+3 {
		t.Errorf("power %d, want 8: X counts the three Islands her text last named", got)
	}
}

// A fetchland changed by Magical Hack and sacrificed to its own ability
// searches for the type its text last named (the search lends the
// departed source's words).
func TestTextWordArgPollutedDeltaChangedToForestFetchesAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	delta := effectstest.PushCatalogPermanent(g, me.ID, "Polluted Delta", "Land", effectstest.TwhPollutedDelta, false)
	forestCard := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	islandCard := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.TwhHack(t, g, delta, effectstest.TwhIsland, effectstest.TwhForest)
	if err := g.ActivateCatalogAbility(me.ID, delta, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	coffers := effectstest.PushCatalogPermanent(g, me.ID, "Cabal Coffers", "Land", effectstest.TwhCabalCoffersOracle, false)
	effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.TwhHack(t, g, coffers, effectstest.TwhSwamp, effectstest.TwhIsland)
	spec, _ := effects.Lookup(effectstest.TwhCabalCoffersOracle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := effectstest.PushCatalogPermanent(g, me.ID, "Leechridden Swamp", "Land — Swamp", effectstest.TwhLeechriddenOracle, false)
	effectstest.SeedColoredCreature(g, me.ID, "Red One", []string{"R"}, 0)
	effectstest.SeedColoredCreature(g, me.ID, "Red Two", []string{"R"}, 0)
	spec, _ := effects.Lookup(effectstest.TwhLeechriddenOracle)
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
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		cop := effectstest.PushCatalogPermanent(g, me.ID, "Circle of Protection: Red", "Enchantment", effectstest.Pr7CoPRed, false)
		red := effectstest.Pr7Creature(g, opp.ID, "Red Dragon", 5, "R")
		green := effectstest.Pr7Creature(g, opp.ID, "Green Bear", 2, "G")
		twhSleight(t, g, cop, 3, 4) // red → green
		effectstest.Pr7Activate(t, g, me.ID, cop, 0, game.ActivateAbilityParams{})
		c := effectstest.Pr7SourcePrompt(t, g)
		if !effectstest.Pr7Offers(c, green) || effectstest.Pr7Offers(c, red) {
			t.Errorf("offers %v: want the green creature and not the red one", c.ChooseCards)
		}
	})
	t.Run("chosen colour stays", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		circle := effectstest.PushCatalogPermanent(g, me.ID, "Story Circle", "Enchantment", effectstest.Pr7StoryCircle, false)
		effectstest.FindBattlefieldCardForTest(g, circle).ChosenColor = "B"
		black := effectstest.Pr7Creature(g, opp.ID, "Black", 3, "B")
		red := effectstest.Pr7Creature(g, opp.ID, "Red", 3, "R")
		twhSleight(t, g, circle, 2, 3) // black → red, which Story Circle never prints
		effectstest.Pr7Activate(t, g, me.ID, circle, 0, game.ActivateAbilityParams{})
		c := effectstest.Pr7SourcePrompt(t, g)
		if !effectstest.Pr7Offers(c, black) || effectstest.Pr7Offers(c, red) {
			t.Errorf("offers %v: the chosen colour is not text, so black only", c.ChooseCards)
		}
	})
}

// Utopia Sprawl's "Enchant Forest" is its own text: changed to "Enchant
// Island", an Aura on a Forest is attached illegally and is put into its
// owner's graveyard (CR 704.5m).
func TestTextWordArgUtopiaSprawlChangedToIslandFallsOffAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	sprawl := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Utopia Sprawl", TypeLine: "Enchantment — Aura", OracleID: effectstest.TwhUtopiaSprawlOracle,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: forest}, ChosenColor: "G",
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !twOnBattlefield(g, sprawl) {
		t.Fatal("setup: Utopia Sprawl on a Forest stays")
	}
	effectstest.TwhHack(t, g, sprawl, effectstest.TwhForest, effectstest.TwhIsland)
	effectstest.PassPriorityAroundTable(t, g)
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
