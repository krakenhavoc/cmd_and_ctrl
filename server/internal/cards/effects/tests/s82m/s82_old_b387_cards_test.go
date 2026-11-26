package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 older sets, Commander 2013, batch 387 (#5560).

const (
	oldB387RestoreOracle     = "cf3165bf-1d59-4e63-8650-bbda119e6335"
	oldB387HippogriffOracle  = "d121108e-f0bc-469b-bf94-e5e5308014a2"
	oldB387SuccessionOracle  = "1b95970c-e7eb-41c4-a8d2-9889b64b3c63"
	oldB387KongmingOracle    = "21e9e1a9-5d6d-473e-adab-6a1e8e2b0ebd"
	oldB387KeyruneOracle     = "7266b491-54e6-4393-a448-d5ae99d965c6"
	oldB387ReflectionsOracle = "76c142ef-0f07-4215-8d11-d25f7114c70d"
	oldB387PastureOracle     = "9eaadbbc-818b-4c21-9d4b-1bba48504d38"
)

func TestOldB387EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB387RestoreOracle:     "Restore",
		oldB387HippogriffOracle:  "Razor Hippogriff",
		oldB387SuccessionOracle:  "Order of Succession",
		oldB387KongmingOracle:    "Kongming, \"Sleeping Dragon\"",
		oldB387KeyruneOracle:     "Azorius Keyrune",
		oldB387ReflectionsOracle: "Tempt with Reflections",
		oldB387PastureOracle:     "Springjack Pasture",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB387RestorePutsAnyGraveyardsLandOntoTheBattlefieldUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	nonland := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: nonland, Name: "Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Restore", "Sorcery", oldB387RestoreOracle, effectstest.CardTarget(nonland)); err == nil {
		t.Fatal("a creature card was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Restore", "Sorcery", oldB387RestoreOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || effectstest.ControllerOf(t, g, land) != me.ID {
		t.Error("the land did not enter under my control")
	}
}

func TestOldB387RazorHippogriffReturnsAnArtifactAndGainsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	art := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: art, Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	life := me.Life
	id := effectstest.CastCatalogSpell(t, g, "Razor Hippogriff", "Creature — Hippogriff", oldB387HippogriffOracle, nil)
	_ = id
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.AnswerFirstPickTarget(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !me.Hand.Contains(art) {
		t.Error("the artifact card did not return to hand")
	}
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d", me.Life, life+4)
	}
}

func TestOldB387KongmingPumpsOnlyTheOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	king := effectstest.B12Push(g, me.ID, "Kongming, \"Sleeping Dragon\"", "Legendary Creature — Human Advisor", oldB387KongmingOracle, 2, 2)
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, g.Seats[1].ID)
	if p, tough := effectstest.PtOf(t, g, king); p != 2 || tough != 2 {
		t.Errorf("Kongming is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("my bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their bear is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB387AzoriusKeyruneAnimatesIntoAFlyingBird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rune := effectstest.PushCatalogPermanent(g, me.ID, "Azorius Keyrune", "Artifact", oldB387KeyruneOracle, false)
	if err := effectstest.ConB1Activate(t, g, me, rune, 0, "W", game.ActivateAbilityParams{}); err == nil {
		// Row 0 of the activated list is the animation, which needs {W}{U}.
		t.Fatal("animated with {W} alone")
	}
	if err := effectstest.ConB1Activate(t, g, me, rune, 0, "WU", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("animate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, rune); p != 2 || tough != 2 {
		t.Errorf("the Keyrune is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, rune, "flying") {
		t.Error("the Bird lacks flying")
	}
}

func TestOldB387TemptWithReflectionsCopiesForTheCasterAndEachAcceptingOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	elk := effectstest.B12Push(g, me.ID, "Test Elk", "Creature — Elk", "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Tempt with Reflections", "Sorcery", oldB387ReflectionsOracle, effectstest.CardTarget(elk))
	effectstest.PassPriorityAroundTable(t, g)
	// Every opponent is asked in turn order; the first accepts, the rest decline.
	for i := 1; i < len(g.Seats); i++ {
		opp := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		effectstest.AnswerMayChoice(t, g, opp.ID, i == 1)
	}
	effectstest.PassPriorityAroundTable(t, g)
	copies := func(p *game.Player) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p.ID && c.Name == "Test Elk" && c.InstanceID != elk {
				n++
			}
		}
		return n
	}
	if got := copies(me); got != 2 {
		t.Errorf("I control %d token copies, want 2 (one, plus one for the accepting opponent)", got)
	}
	if got := copies(g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]); got != 1 {
		t.Errorf("the accepting opponent controls %d copies, want 1", got)
	}
}

func TestOldB387SpringjackPastureMakesGoatsAndSacrificesThemForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pasture := effectstest.PushCatalogPermanent(g, me.ID, "Springjack Pasture", "Land", oldB387PastureOracle, false)
	if err := effectstest.ConB1Activate(t, g, me, pasture, 0, "CCCC", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("make a Goat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	goats := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Goat" {
			goats++
		}
	}
	if goats != 1 {
		t.Fatalf("%d Goats, want 1", goats)
	}
	// Untap and sacrifice two Goats for two mana of one colour and 2 life.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pasture {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	a := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Goat", TypeLine: "Token Creature — Goat", Power: 0, Toughness: 1})
	b := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Goat", TypeLine: "Token Creature — Goat", Power: 0, Toughness: 1})
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, pasture, 1, game.ManaAbilityParams{XValue: 2, SacrificeIDs: []uuid.UUID{a, b}, Colors: []string{"G"}}); err != nil {
		t.Fatalf("sacrifice two Goats: %v", err)
	}
	green := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "G" {
			green += tok.N()
		}
	}
	if green != 2 {
		t.Errorf("%d green mana, want 2", green)
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the sacrificed Goats are still on the battlefield")
	}
}

func TestOldB387OrderOfSuccessionPassesCreaturesAroundTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	n := len(g.Seats)
	start := g.Turn.ActiveSeat
	bears := make([]uuid.UUID, n)
	for i, p := range g.Seats {
		bears[i] = effectstest.SeedBear(g, p.ID)
	}
	effectstest.CastCatalogSpell(t, g, "Order of Succession", "Sorcery", oldB387SuccessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, g.Seats[start].ID, 0)
	for i := 0; i < n; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.PendingOfKind(g, game.PendingChoiceChooseCards)
		if c == nil {
			t.Fatalf("pick %d: no choose_cards prompt is open", i)
		}
		wantChooser := g.Seats[(start+i)%n].ID
		if c.Chooser != wantChooser {
			t.Fatalf("pick %d is asked of %s, want %s", i, c.Chooser, wantChooser)
		}
		if len(c.ChooseCards) != 1 {
			t.Fatalf("pick %d offers %v, want the next player's one creature", i, c.ChooseCards)
		}
		if err := g.ResolveChooseCards(c.ID, c.Chooser, c.ChooseCards[:1]); err != nil {
			t.Fatalf("pick %d: %v", i, err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := range g.Seats {
		next := (i + 1) % n
		// Left: seat i chose seat i+1's bear and now controls it.
		if got := effectstest.ControllerOf(t, g, bears[next]); got != g.Seats[i].ID {
			t.Errorf("seat %d's bear is controlled by %s, want seat %d's player", next, got, i)
		}
	}
}
