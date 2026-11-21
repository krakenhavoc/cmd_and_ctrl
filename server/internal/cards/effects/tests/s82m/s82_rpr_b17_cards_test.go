package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b17_cards_test.go — reprints in DMC/AFC, batch 17 (#5908). Each test
// plays the real catalog entry through the engine.

const (
	rprB17DuneblastOracle  = "14104a52-c4e8-4755-8196-7fa5c5a35f36"
	rprB17EchoingOracle    = "ef100544-3e6d-4661-a1d1-8b02e58c059c"
	rprB17FloodPlainOracle = "9043d8d5-b38a-406f-a44c-49f579c644f0"
	rprB17FusionOracle     = "a152674f-be29-40ab-8dd9-376fd4eb3bb8"
	rprB17GarnaOracle      = "97cb993c-90ae-49ff-8ccb-b4fe317a43ef"
	rprB17GlintEyeOracle   = "d6b6d965-1401-4566-adb3-90b03218af8d"
	rprB17GrasslandsOracle = "e80bd454-8bc5-4921-90cf-6ad28a27a88b"
	rprB17HeroOracle       = "5a66802a-76f9-4a31-a776-24110dcfca64"
	rprB17PodiumOracle     = "85292bfb-32d7-449e-9ebb-5be828e8ecf6"
	rprB17KnightOracle     = "170856f4-7c9d-494b-9bae-20e4d9eba0da"
)

func TestRprB17EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB17DuneblastOracle:  "Duneblast",
		rprB17EchoingOracle:    "Echoing Truth",
		rprB17FloodPlainOracle: "Flood Plain",
		rprB17FusionOracle:     "Fusion Elemental",
		rprB17GarnaOracle:      "Garna, the Bloodflame",
		rprB17GlintEyeOracle:   "Glint-Eye Nephilim",
		rprB17GrasslandsOracle: "Grasslands",
		rprB17HeroOracle:       "Hero of Precinct One",
		rprB17PodiumOracle:     "Heroes' Podium",
		rprB17KnightOracle:     "Knight of New Alara",
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

func rprB17PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.SocB14PT(t, g, id)
}

func TestRprB17KnightOfNewAlaraPumpsOtherMulticoloredByColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Knight of New Alara", TypeLine: "Creature — Human Knight", OracleID: rprB17KnightOracle, Power: 2, Toughness: 2, Colors: []string{"G", "W"}})
	two := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Two", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"U", "R"}})
	three := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Three", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"U", "R", "B"}})
	mono := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mono", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"U"}})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"U", "R"}})
	for _, c := range []struct {
		id   uuid.UUID
		want int
		what string
	}{{knight, 2, "the Knight itself"}, {two, 4, "a two-color creature"}, {three, 5, "a three-color creature"}, {mono, 2, "a monocolored creature"}, {theirs, 2, "an opponent's creature"}} {
		if p, tg := rprB17PT(t, g, c.id); p != c.want || tg != c.want {
			t.Errorf("%s is %d/%d, want %d/%d", c.what, p, tg, c.want, c.want)
		}
	}
}

func TestRprB17HeroesPodiumCountsOtherLegendaryCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Heroes' Podium", "Legendary Artifact", rprB17PodiumOracle, false)
	a := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "A", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	plain := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Plain", TypeLine: "Creature — Human", Power: 2, Toughness: 2})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	if p, _ := rprB17PT(t, g, a); p != 2 {
		t.Errorf("a lone legend gets nothing, got power %d", p)
	}
	b := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "B", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	c := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "C", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	for _, id := range []uuid.UUID{a, b, c} {
		if p, tg := rprB17PT(t, g, id); p != 4 || tg != 4 {
			t.Errorf("with three legends each is %d/%d, want 4/4", p, tg)
		}
	}
	if p, _ := rprB17PT(t, g, plain); p != 2 {
		t.Errorf("a nonlegendary creature got power %d, want 2", p)
	}
	if p, _ := rprB17PT(t, g, theirs); p != 2 {
		t.Errorf("an opponent's legend got power %d, want 2", p)
	}
}

func TestRprB17HeroesPodiumLooksAtXAndTakesALegend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	podium := effectstest.PushCatalogPermanent(g, me.ID, "Heroes' Podium", "Legendary Artifact", rprB17PodiumOracle, false)
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Filler", "Creature — Elf"},
		{"Deep Legend", "Legendary Creature — Elf"},
		{"Too Deep", "Legendary Creature — Elf"},
	})
	if err := g.ActivateCatalogAbility(me.ID, podium, 0, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no Heroes' Podium prompt")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != top[1] || prompt.ChooseMax != 1 || prompt.ChooseMin != 0 {
		t.Fatalf("offered %v (min %d, max %d), want the legend among the top two alone, optional", prompt.ChooseCards, prompt.ChooseMin, prompt.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[1])
	if !me.Hand.Contains(top[1]) {
		t.Error("the legendary creature is not in hand")
	}
	if !effectstest.FindBattlefieldCardForTest(g, podium).Tapped {
		t.Error("the Podium should have tapped for its cost")
	}
	if lib := me.Library.Cards; len(lib) == 0 || lib[0].InstanceID != top[0] || lib[len(lib)-1].InstanceID != top[2] {
		t.Error("the unlooked-at legend should be on top and the filler on the bottom")
	}
}

func TestRprB17HeroMakesAHumanForMulticoloredSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Hero of Precinct One", "Creature — Human Warrior", rprB17HeroOracle, false)
	humans := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Human" && c.IsToken() {
				n++
			}
		}
		return n
	}
	cast := func(colors []string) {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID, Colors: colors})
		for g.Turn.Step != game.StepPrecombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast([]string{"U"})
	if humans() != 0 {
		t.Fatal("a monocolored spell made a Human")
	}
	cast([]string{"U", "R"})
	if humans() != 1 {
		t.Fatalf("a multicolored spell made %d Humans, want 1", humans())
	}
}

func TestRprB17GlintEyeDrawsThatManyAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glint := effectstest.PushCatalogPermanent(g, me.ID, "Glint-Eye Nephilim", "Creature — Nephilim", rprB17GlintEyeOracle, false)
	effectstest.FindBattlefieldCardForTest(g, glint).Power = 2
	effectstest.FindBattlefieldCardForTest(g, glint).Toughness = 2
	handBefore := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, glint)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40-2 {
		t.Fatalf("opponent life %d, want 38", opp.Life)
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d -> %d, want two cards drawn", handBefore, got)
	}
}

func TestRprB17GlintEyePumpDiscardsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glint := effectstest.PushCatalogPermanent(g, me.ID, "Glint-Eye Nephilim", "Creature — Nephilim", rprB17GlintEyeOracle, false)
	effectstest.FindBattlefieldCardForTest(g, glint).Power = 2
	effectstest.FindBattlefieldCardForTest(g, glint).Toughness = 2
	if p, _ := rprB17PT(t, g, glint); p != 2 {
		t.Fatalf("base power %d", p)
	}
	// A discard cost with nothing named to discard is refused.
	if err := g.ActivateCatalogAbility(me.ID, glint, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without discarding a card")
	}
	handBefore := me.Hand.Size() + 1
	if err := g.ActivateCatalogAbility(me.ID, glint, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{effectstest.PushHandCard(g, me)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore-1 {
		t.Errorf("hand %d -> %d, want one card discarded", handBefore, got)
	}
	if p, tg := rprB17PT(t, g, glint); p != 3 || tg != 3 {
		t.Errorf("after the pump it is %d/%d, want 3/3", p, tg)
	}
}

func TestRprB17GarnaReturnsCreaturesPutInGraveyardThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fresh := uuid.New()
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: fresh, Name: "Fresh", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID, PutInGraveyardThisTurn: true})
	spell := me.Graveyard.Size()
	_ = spell
	other := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Other", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	opp := effectstest.ApaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.CastCatalogSpell(t, g, "Garna, the Bloodflame", "Legendary Creature — Human Warrior", rprB17GarnaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(fresh) {
		t.Error("the creature put into the graveyard this turn was not returned")
	}
	if me.Hand.Contains(old) || !me.Graveyard.Contains(old) {
		t.Error("an earlier turn's creature was returned")
	}
	if !rprB17HasHaste(g, other) {
		t.Error("another creature you control lacks haste")
	}
	if rprB17HasHaste(g, opp) {
		t.Error("an opponent's creature gained haste")
	}
}

func TestRprB17EchoingTruthReturnsEveryPermanentWithTheName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1})
	b := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1})
	other := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3})
	effectstest.CastCatalogSpell(t, g, "Echoing Truth", "Instant", rprB17EchoingOracle, []game.TargetRef{{Kind: game.TargetCard, ID: a}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a same-name permanent stayed on the battlefield")
	}
	if !me.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("each Goblin should return to its owner's hand")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a differently named permanent was returned")
	}
}

func TestRprB17EchoingTruthRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.ApaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	err := effectstest.CastCatalogSpellErr(t, g, "Echoing Truth", "Instant", rprB17EchoingOracle, []game.TargetRef{{Kind: game.TargetCard, ID: land}})
	if err == nil {
		t.Fatal("a land is not a legal target")
	}
}

func TestRprB17DuneblastKeepsTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	effectstest.CastCatalogSpell(t, g, "Duneblast", "Sorcery", rprB17DuneblastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if choice == nil {
		t.Fatalf("no creature prompt: %+v", g.PendingChoices)
	}
	if choice.ChooseMin != 0 || choice.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want up to one", choice.ChooseMin, choice.ChooseMax)
	}
	if err := g.ResolveOwnPermanents(choice.ID, me.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("choose an opponent's creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the unchosen creature survived")
	}
	if !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(rock) {
		t.Error("the chosen creature and the artifact should remain")
	}
}

func TestRprB17DuneblastWithNoChoiceDestroysEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.CastCatalogSpell(t, g, "Duneblast", "Sorcery", rprB17DuneblastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if choice == nil {
		t.Fatal("no creature prompt")
	}
	if err := g.ResolveOwnPermanents(choice.ID, me.ID, nil); err != nil {
		t.Fatalf("choose none: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("choosing none should destroy every creature")
	}
}

func TestRprB17GrasslandsAndFloodPlainFetchTheirTypes(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		good, bad    []string
	}{
		{"Grasslands", rprB17GrasslandsOracle, []string{"Basic Land — Plains", "Land — Forest Island"}, []string{"Basic Land — Island"}},
		{"Flood Plain", rprB17FloodPlainOracle, []string{"Basic Land — Island", "Land — Plains Swamp"}, []string{"Basic Land — Forest"}},
	} {
		for _, want := range tc.good {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			fetched := effectstest.StapleLibraryCard(me, "Target", want)
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("%s activate: %v", tc.name, err)
			}
			if _, still := effectstest.AangCardOnBF(g, land); still {
				t.Errorf("%s was not sacrificed as a cost", tc.name)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.TakeSearchPrompts(t, g)
			got, ok := effectstest.AangCardOnBF(g, fetched)
			if !ok {
				t.Errorf("%s did not fetch %q", tc.name, want)
			} else if got.Tapped {
				t.Errorf("%s: the fetched land should arrive untapped", tc.name)
			}
		}
		for _, no := range tc.bad {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			effectstest.StapleLibraryCard(me, "Wrong", no)
			before := effectstest.StapleCountLands(g)
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("%s activate: %v", tc.name, err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.StapleCountLands(g); got != before-1 {
				t.Errorf("%s fetched a wrong-type land %q", tc.name, no)
			}
		}
	}
}

func rprB17HasHaste(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && game.HasKeyword(&c, "haste")
}
