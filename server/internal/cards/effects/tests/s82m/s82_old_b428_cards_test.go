package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b428_cards_test.go — Magic 2013: card draw and library, pump and
// combat tricks, +1/+1 and other counters (S82 older sets, batch 428, #5568).
// Each test plays the real catalog entry through the engine.

const (
	oldB428Shade      = "44f87b6e-0cc8-4088-8a94-8272b6b26869"
	oldB428Gem        = "6b1e4940-11ac-4d85-98f4-3f50d1060b96"
	oldB428Tracker    = "b6265654-84c3-469b-9afa-f752fd64abe0"
	oldB428Revelation = "dcf438a7-adc4-4658-b75a-2dd633f595a7"
	oldB428Realms     = "e6acafb7-856f-4125-a908-a620635b8ba3"
	oldB428Nefarox    = "0632f0f9-8d4f-4e94-8b53-1275b7f21a57"
	oldB428Nocturnus  = "2109c49e-dc81-4a10-8265-3cf8c90db523"
	oldB428Phantasm   = "ab56df8f-0ba3-4d66-bde2-4c9b4acf07e5"
	oldB428Gorgon     = "1e65fac0-23fc-4a32-bd36-4e3f7ebfe33d"
	oldB428Mutiny     = "5599ddb6-1404-4884-b6ca-c1852a521079"
	oldB428Dragon     = "d424cab6-84be-4e55-8f4e-af11b1c50ede"
	oldB428Bear       = "Creature — Bear"
)

func TestOldB428EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB428Shade:      "Liliana's Shade",
		oldB428Gem:        "Gem of Becoming",
		oldB428Tracker:    "Mwonvuli Beast Tracker",
		oldB428Revelation: "Diabolic Revelation",
		oldB428Realms:     "Boundless Realms",
		oldB428Nefarox:    "Nefarox, Overlord of Grixis",
		oldB428Nocturnus:  "Vampire Nocturnus",
		oldB428Phantasm:   "Jace's Phantasm",
		oldB428Gorgon:     "Xathrid Gorgon",
		oldB428Mutiny:     "Mark of Mutiny",
		oldB428Dragon:     "Slumbering Dragon",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB428Lib(p *game.Player, name, typeLine string, keywords ...string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Keywords: keywords, Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID})
	return id
}

func oldB428Hand(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Liliana's Shade --------------------------------------------------------

func TestOldB428LilianasShadeFetchesASwampWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	swamp := oldB428Lib(me, "Swamp", "Basic Land — Swamp")
	forest := oldB428Lib(me, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Liliana's Shade", "Creature — Shade", oldB428Shade, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a Forest was offered for a Swamp search")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, swamp)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(swamp) {
		t.Error("the Swamp did not go to hand")
	}
	if me.Hand.Contains(forest) {
		t.Error("the Forest went to hand")
	}
}

func TestOldB428LilianasShadePumpsForB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shade := effectstest.B12Push(g, me.ID, "Liliana's Shade", "Creature — Shade", oldB428Shade, 1, 1)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		effectstest.ApaMana(me, "B")
		effectstest.ApaActivate(t, g, me, shade, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.PtOf(t, g, shade); p != 3 || tough != 3 {
		t.Errorf("Shade is %d/%d, want 3/3", p, tough)
	}
}

// --- Gem of Becoming --------------------------------------------------------

func oldB428GemGame(t *testing.T) (*game.Game, *game.Player, map[string]uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := map[string]uuid.UUID{
		"island":   oldB428Lib(me, "Island", "Basic Land — Island"),
		"island 2": oldB428Lib(me, "Island Two", "Basic Land — Island"),
		"swamp":    oldB428Lib(me, "Swamp", "Basic Land — Swamp"),
		"mountain": oldB428Lib(me, "Mountain", "Basic Land — Mountain"),
		"volcanic": oldB428Lib(me, "Volcanic Island", "Land — Island Mountain"),
		"forest":   oldB428Lib(me, "Forest", "Basic Land — Forest"),
	}
	gem := effectstest.PushCatalogPermanent(g, me.ID, "Gem of Becoming", "Artifact", oldB428Gem, false)
	effectstest.ApaMana(me, "C", "C", "C")
	effectstest.ApaActivate(t, g, me, gem, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	ids["gem"] = gem
	return g, me, ids
}

func TestOldB428GemOfBecomingFetchesOneOfEachAndIsSacrificed(t *testing.T) {
	g, me, ids := oldB428GemGame(t)
	if effects.OnBattlefield(g, ids["gem"]) {
		t.Error("the Gem was not sacrificed")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids["island"], ids["swamp"], ids["mountain"])
	effectstest.PassPriorityAroundTable(t, g)
	for _, k := range []string{"island", "swamp", "mountain"} {
		if !me.Hand.Contains(ids[k]) {
			t.Errorf("the %s did not go to hand", k)
		}
	}
	if me.Hand.Contains(ids["forest"]) || me.Hand.Contains(ids["island 2"]) {
		t.Error("an extra card went to hand")
	}
}

func TestOldB428GemOfBecomingRefusesTwoOfTheSameLandType(t *testing.T) {
	g, me, ids := oldB428GemGame(t)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids["island"], ids["island 2"]}); err == nil {
		t.Error("two Islands were accepted for the Island slot")
	}
	// A dual land fills the one slot it is needed for.
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids["volcanic"], ids["island"], ids["swamp"]}); err != nil {
		t.Errorf("Volcanic Island as the Mountain, an Island and a Swamp was refused: %v", err)
	}
}

// --- Mwonvuli Beast Tracker -------------------------------------------------

func TestOldB428MwonvuliBeastTrackerPutsTheFoundCreatureOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plain := oldB428Lib(me, "Plain Bear", oldB428Bear)
	trample := oldB428Lib(me, "Trampler", oldB428Bear, "trample")
	oldB428Lib(me, "Hexproof Land", "Basic Land — Forest", "hexproof")
	oldB428Lib(me, "Filler", "Sorcery")
	size := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Mwonvuli Beast Tracker", "Creature — Human Scout", oldB428Tracker, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Plain Bear") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Hexproof Land") != uuid.Nil {
		t.Error("a creature without the keywords, or a noncreature, was offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, trample)
	effectstest.PassPriorityAroundTable(t, g)
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != trample {
		t.Errorf("the top of the library is %q, want the Trampler", top.Name)
	}
	if me.Library.Size() != size {
		t.Errorf("library %d -> %d, want unchanged", size, me.Library.Size())
	}
	_ = plain
}

// --- Diabolic Revelation ----------------------------------------------------

func TestOldB428DiabolicRevelationTakesUpToXCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := oldB428Lib(me, "Card A", "Sorcery")
	b := oldB428Lib(me, "Card B", "Creature — Bear")
	oldB428Lib(me, "Card C", "Sorcery")
	oldB428Lib(me, "Card D", "Sorcery")
	if err := effectstest.CastX(t, g, "Diabolic Revelation", "Sorcery", oldB428Revelation, 2, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{a, b, oldB428Lib(me, "Card E", "Sorcery")}); err == nil {
		t.Error("three cards were accepted for X = 2")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("two cards refused: %v", err)
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the chosen cards did not go to hand")
	}
}

// --- Boundless Realms -------------------------------------------------------

func TestOldB428BoundlessRealmsFetchesBasicsEqualToYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.B12Push(g, me.ID, "Plains", "Basic Land — Plains", "", 0, 0)
	}
	f1 := oldB428Lib(me, "Forest", "Basic Land — Forest")
	f2 := oldB428Lib(me, "Forest", "Basic Land — Forest")
	f3 := oldB428Lib(me, "Forest", "Basic Land — Forest")
	f4 := oldB428Lib(me, "Forest", "Basic Land — Forest")
	nonbasic := oldB428Lib(me, "Fancy Land", "Land")
	if err := effectstest.CastX(t, g, "Boundless Realms", "Sorcery", oldB428Realms, 0, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{f1, f2, f3, f4}); err == nil {
		t.Error("four lands were accepted with three lands controlled")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{nonbasic}); err == nil {
		t.Error("a nonbasic land was accepted")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{f1, f2, f3}); err != nil {
		t.Fatalf("three basics refused: %v", err)
	}
	for _, id := range []uuid.UUID{f1, f2, f3} {
		if !g.Battlefield.Contains(id) || !effectstest.PermanentTapped(g, id) {
			t.Errorf("a fetched Forest is not on the battlefield tapped")
		}
	}
}

// --- Nefarox, Overlord of Grixis --------------------------------------------

func TestOldB428NefaroxMakesTheDefenderSacrificeWhenItAttacksAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nef := effectstest.B12Push(g, me.ID, "Nefarox, Overlord of Grixis", "Legendary Creature — Demon", oldB428Nefarox, 5, 5)
	for _, kw := range []string{"flying", game.KeywordExalted} {
		if !effectstest.HasKeywordOnBattlefield(t, g, nef, kw) {
			t.Errorf("no %s", kw)
		}
	}
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB428Bear, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(nef, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	// Exalted and the edict trigger together: order them.
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, victim) {
		t.Error("the defending player's creature was not sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, nef); p != 6 || tough != 6 {
		t.Errorf("Nefarox is %d/%d after attacking alone, want 6/6 (exalted)", p, tough)
	}
}

func TestOldB428NefaroxMakesNoOneSacrificeWhenItAttacksWithAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nef := effectstest.B12Push(g, me.ID, "Nefarox, Overlord of Grixis", "Legendary Creature — Demon", oldB428Nefarox, 5, 5)
	friend := effectstest.B12Creature(g, me.ID, "Friend", oldB428Bear, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB428Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, nef, friend)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, victim) {
		t.Error("a creature was sacrificed although Nefarox did not attack alone")
	}
}

// --- Vampire Nocturnus ------------------------------------------------------

func TestOldB428VampireNocturnusPumpsVampiresWhileTheTopCardIsBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nocturnus := effectstest.B12Push(g, me.ID, "Vampire Nocturnus", "Creature — Vampire", oldB428Nocturnus, 3, 3)
	bat := effectstest.B12Creature(g, me.ID, "Fanged Vampire", "Creature — Vampire", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB428Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 1, 1)
	effectstest.SeedLibraryTop(me, "Red Card", "Sorcery", "{R}")
	g.BumpLayerVersionForTest()
	if p, _ := effectstest.PtOf(t, g, nocturnus); p != 3 {
		t.Errorf("Nocturnus is %d power with a red top card, want 3", p)
	}
	effectstest.SeedLibraryTop(me, "Black Card", "Sorcery", "{1}{B}")
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.PtOf(t, g, nocturnus); p != 5 || tough != 4 {
		t.Errorf("Nocturnus is %d/%d with a black top card, want 5/4", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, bat); p != 3 || tough != 2 {
		t.Errorf("the other Vampire is %d/%d, want 3/2", p, tough)
	}
	for _, id := range []uuid.UUID{nocturnus, bat} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
			t.Error("a Vampire has no flying")
		}
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 || effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("a non-Vampire was changed")
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 1 {
		t.Error("an opponent's Vampire was pumped")
	}
}

func TestOldB428VampireNocturnusRevealsTheTopCard(t *testing.T) {
	spec, ok := effects.Lookup(oldB428Nocturnus)
	if !ok || spec.LibraryTopVisible != game.LibraryTopRevealed {
		t.Errorf("LibraryTopVisible = %v, want revealed", spec.LibraryTopVisible)
	}
}

// --- Jace's Phantasm --------------------------------------------------------

func TestOldB428JacesPhantasmGrowsWhenAnOpponentHasTenCardsInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ph := effectstest.B12Push(g, me.ID, "Jace's Phantasm", "Creature — Illusion", oldB428Phantasm, 1, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, ph, "flying") {
		t.Error("no flying")
	}
	for i := 0; i < 9; i++ {
		opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	for i := 0; i < 12; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.PtOf(t, g, ph); p != 1 || tough != 1 {
		t.Fatalf("Phantasm is %d/%d with nine cards in their graveyard (twelve in mine), want 1/1", p, tough)
	}
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Tenth", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.PtOf(t, g, ph); p != 5 || tough != 5 {
		t.Errorf("Phantasm is %d/%d with ten, want 5/5", p, tough)
	}
}

// --- Xathrid Gorgon ---------------------------------------------------------

func TestOldB428XathridGorgonPetrifiesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gorgon := effectstest.PushCatalogPermanent(g, me.ID, "Xathrid Gorgon", "Creature — Gorgon", oldB428Gorgon, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, gorgon, "deathtouch") {
		t.Error("no deathtouch")
	}
	// A creature of theirs with an activated ability of its own.
	shade := effectstest.PushCatalogPermanent(g, opp.ID, "Liliana's Shade", "Creature — Shade", oldB428Shade, false)
	effectstest.ApaMana(me, "C", "C", "B")
	effectstest.B16Activate(t, g, me.ID, gorgon, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(shade)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, shade, "petrification"); got != 1 {
		t.Errorf("%d petrification counters, want 1", got)
	}
	c := effectstest.S58p6View(t, g, shade)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("creature=%v artifact=%v, want both (in addition to its other types)", c.IsCreature(), c.IsArtifact())
	}
	if cols := c.EffectiveColors(); len(cols) != 0 {
		t.Errorf("colours %v, want colourless", cols)
	}
	if !effectstest.HasEffectiveKeyword(t, g, shade, "defender") {
		t.Error("no defender")
	}
	effectstest.ApaMana(opp, "B")
	if err := g.ActivateCatalogAbility(opp.ID, shade, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the petrified creature's ability was activated")
	}
	// It lasts: it is still petrified on a later turn.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, shade, "defender") {
		t.Error("the effect ended with the turn")
	}
}

// --- Mark of Mutiny ---------------------------------------------------------

func TestOldB428MarkOfMutinyStealsGrowsUntapsAndGivesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB428Bear, 2, 2)
	effectstest.FindBattlefieldCardByID(g, theirs).Tapped = true
	spell := oldB428Hand(me, "Mark of Mutiny", "Sorcery", "{2}{R}", oldB428Mutiny)
	effectstest.FloatForTest(g, me, "RCC")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, theirs); c.Controller != me.ID {
		t.Fatalf("controller is %v, want me", c.Controller)
	}
	if effectstest.PermanentTapped(g, theirs) {
		t.Error("it is still tapped")
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("no haste")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, theirs); c.Controller != opp.ID {
		t.Error("control did not return at end of turn")
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter is gone after control returned (%d)", got)
	}
}

// --- Slumbering Dragon ------------------------------------------------------

func TestOldB428SlumberingDragonGrowsWhenACreatureAttacksItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	dragon := effectstest.B12Push(g, opp.ID, "Slumbering Dragon", "Creature — Dragon", oldB428Dragon, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB428Bear, 2, 2)
	effectstest.B12Creature(g, other.ID, "Bystander", oldB428Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, dragon, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after an attack on its controller, want 1", got)
	}
}

func TestOldB428SlumberingDragonIgnoresAnAttackOnAnotherPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	dragon := effectstest.B12Push(g, opp.ID, "Slumbering Dragon", "Creature — Dragon", oldB428Dragon, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB428Bear, 2, 2)
	effectstest.DeclareAttack(t, g, other.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, dragon, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters after an attack on someone else, want 0", got)
	}
}

func TestOldB428SlumberingDragonAttacksOnlyWithFiveCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Slumbering Dragon", "Creature — Dragon", oldB428Dragon, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, dragon, "flying") {
		t.Error("no flying")
	}
	effectstest.SncCounter(g, dragon, game.CounterPlusOne, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dragon, opp.ID); err == nil {
		t.Fatal("it attacked with four +1/+1 counters")
	}
	effectstest.SncCounter(g, dragon, game.CounterPlusOne, 1)
	if err := g.DeclareAttacker(dragon, opp.ID); err != nil {
		t.Errorf("it can't attack with five +1/+1 counters: %v", err)
	}
}
