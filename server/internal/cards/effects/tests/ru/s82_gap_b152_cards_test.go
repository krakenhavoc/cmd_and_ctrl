package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b152_cards_test.go — GAP batch 152, Jumpstart 2022 creatures
// (#5948). Every card is played through the engine.

const (
	gap152FrontierMastodonOracle  = "ae81040b-572d-4906-bacc-e7bae16e1130"
	gap152GiantCaterpillarOracle  = "43a515f2-b53a-4bb4-bb6e-0d79c9162359"
	gap152GiantLadybugOracle      = "abfe31f1-7c6d-4a7a-9f44-7f3e4166d9f1"
	gap152HavenwoodWurmOracle     = "a4ddff34-d0ae-481f-870b-b911b249250f"
	gap152IridescentHornbeetleOra = "3270265f-fa9f-4ef0-a1a3-a68fc2349bf1"
	gap152KraulForagersOracle     = "f414f997-06f3-4741-a9a5-9753c182e862"
	gap152NagaVitalistOracle      = "9cea8783-c7d7-49f9-b23d-324309efc140"
	gap152OashraCultivatorOracle  = "c07bbe66-96ae-4fab-82b9-fc6dc4f549c3"
	gap152OnduGiantOracle         = "ed8b2887-15fa-4e62-83ac-55973ac6f76c"
	gap152QuarryBeetleOracle      = "fff66029-24bb-4354-a650-99a205fd168f"
)

func gap152Seats(g *game.Game) (me, opp *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}

func TestGap152EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		gap152FrontierMastodonOracle, gap152GiantCaterpillarOracle, gap152GiantLadybugOracle,
		gap152HavenwoodWurmOracle, gap152IridescentHornbeetleOra, gap152KraulForagersOracle,
		gap152NagaVitalistOracle, gap152OashraCultivatorOracle, gap152OnduGiantOracle,
		gap152QuarryBeetleOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestGap152PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Giant Ladybug", "Creature — Insect", gap152GiantLadybugOracle, []string{"reach"}},
		{"Havenwood Wurm", "Creature — Wurm", gap152HavenwoodWurmOracle, []string{"flash", "trample"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me, _ := gap152Seats(g)
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		abilities := effectstest.EffectiveAbilities(t, g, id)
		for _, kw := range tc.keywords {
			if !effectstest.HasAbility(abilities, kw) {
				t.Errorf("%s has no %s", tc.name, kw)
			}
		}
	}
}

func TestGap152FrontierMastodonEntersWithACounterOnlyWithAFourPowerCreature(t *testing.T) {
	for _, big := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := gap152Seats(g)
		if big {
			effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
		} else {
			effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
		}
		id := effectstest.CastHoldCreature(t, g, "Frontier Mastodon", "Creature — Elephant", gap152FrontierMastodonOracle, 3, 2)
		want := 0
		if big {
			want = 1
		}
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != want {
			t.Errorf("big=%v: %d +1/+1 counters, want %d", big, got, want)
		}
	}
}

func TestGap152GiantCaterpillarMakesAButterflyAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := gap152Seats(g)
	cat := effectstest.B12Push(g, me.ID, "Giant Caterpillar", "Creature — Insect", gap152GiantCaterpillarOracle, 3, 3)
	effectstest.ApaMana(me, "G")
	effectstest.ApaActivate(t, g, me, cat, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(cat) {
		t.Fatal("the Caterpillar was not sacrificed")
	}
	butterflies := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Butterfly" && c.IsToken() && c.Controller == me.ID {
				n++
			}
		}
		return n
	}
	if butterflies() != 0 {
		t.Fatal("the Butterfly came before the end step")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if butterflies() != 1 {
		t.Fatalf("%d Butterflies at the end step, want 1", butterflies())
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Butterfly" {
			if c.Power != 1 || c.Toughness != 1 || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") || !c.HasSubtype("Insect") {
				t.Errorf("the Butterfly is %d/%d %q", c.Power, c.Toughness, c.TypeLine)
			}
		}
	}
}

func TestGap152GiantLadybugPutsABasicLandOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap152Seats(g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
	)
	effectstest.CastHoldCreature(t, g, "Giant Ladybug", "Creature — Insect", gap152GiantLadybugOracle, 4, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.Name != "Forest" {
		t.Errorf("the top card is %s, want Forest", top.Name)
	}
	_ = ids
}

func TestGap152OnduGiantFetchesABasicLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap152Seats(g)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
	)
	effectstest.CastHoldCreature(t, g, "Ondu Giant", "Creature — Giant Druid", gap152OnduGiantOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == me.ID {
			found = true
			if !c.Tapped {
				t.Error("the land entered untapped")
			}
		}
	}
	if !found {
		t.Error("no Forest on the battlefield")
	}
}

func TestGap152OashraCultivatorSacrificesForABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := gap152Seats(g)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	cult := effectstest.B12Push(g, me.ID, "Oashra Cultivator", "Creature — Human Druid", gap152OashraCultivatorOracle, 0, 3)
	effectstest.ApaMana(me, "G", "C", "C")
	effectstest.ApaActivate(t, g, me, cult, 0, game.ActivateAbilityParams{})
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cult) {
		t.Error("the Cultivator was not sacrificed")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == me.ID && c.Tapped {
			found = true
		}
	}
	if !found {
		t.Error("no tapped Forest on the battlefield")
	}
}

func TestGap152QuarryBeetleReturnsALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap152Seats(g)
	land := effectstest.PushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.CastHoldCreature(t, g, "Quarry Beetle", "Creature — Insect", gap152QuarryBeetleOracle, 4, 5)
	effectstest.PickCard(t, g, me.ID, land)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("the land card did not return to the battlefield")
	}
}

func TestGap152KraulForagersGainsLifePerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap152Seats(g)
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardTyped(me, "Dead", "Creature — Bear")
	}
	effectstest.PushGraveyardCardTyped(me, "Spell", "Sorcery")
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Kraul Foragers", "Creature — Insect Scout", gap152KraulForagersOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestGap152NagaVitalistProducesWhatYourLandsCould(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := gap152Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	naga := effectstest.B12Push(g, me.ID, "Naga Vitalist", "Creature — Snake Druid", gap152NagaVitalistOracle, 1, 2)
	spec, _ := effects.Lookup(gap152NagaVitalistOracle)
	if len(spec.ManaAbilities) != 1 || !spec.ManaAbilities[0].DerivesFromOtherSources {
		t.Fatalf("mana abilities: %+v", spec.ManaAbilities)
	}
	if err := g.ActivateManaAbility(me.ID, naga, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the Vitalist: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("the pool holds %d mana, want 1", me.ManaPool.Total())
	}
}

func TestGap152NagaVitalistMakesNothingWithoutLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := gap152Seats(g)
	naga := effectstest.B12Push(g, me.ID, "Naga Vitalist", "Creature — Snake Druid", gap152NagaVitalistOracle, 1, 2)
	_ = g.ActivateManaAbility(me.ID, naga, 0, game.ManaAbilityParams{})
	if me.ManaPool.Total() != 0 {
		t.Errorf("the pool holds %d mana with no lands, want 0", me.ManaPool.Total())
	}
}

func TestGap152IridescentHornbeetleMakesInsectsForCountersPutThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := gap152Seats(g)
	effectstest.B12Push(g, me.ID, "Iridescent Hornbeetle", "Creature — Insect", gap152IridescentHornbeetleOra, 3, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterByForEffect(me.ID, mine, game.CounterPlusOne, 2)
		_ = g.AddCounterByForEffect(me.ID, theirs, game.CounterPlusOne, 5)
		_ = g.AddCounterByForEffect(opp.ID, mine, game.CounterPlusOne, 3)
	})
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginEndStep, Actor: me.ID, Step: game.StepEnd}) })
	effectstest.PassPriorityAroundTable(t, g)
	insects := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Insect" && c.IsToken() {
			insects++
		}
	}
	if insects != 2 {
		t.Errorf("%d Insect tokens, want 2 (counters you put on your own creatures only)", insects)
	}
	_ = uuid.Nil
}
