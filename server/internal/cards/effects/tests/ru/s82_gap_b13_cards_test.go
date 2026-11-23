package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b13_cards_test.go — Innistrad Remastered creatures, GAP batch 13
// (#5948). Each test plays the real catalog entry through the engine.

const (
	gapTowerGeistOracle         = "10e19453-84fb-4845-8db0-c9da52847e09"
	gapSpellQuellerOracle       = "135135f2-1643-40db-91f2-bb7b956d212c"
	gapApothecaryGeistOracle    = "3420da71-0af4-46b4-a994-91a82332d0fb"
	gapAvacynianPriestOracle    = "df7e296d-96de-4280-a4d8-49189544b9e3"
	gapBriselaOracle            = "b23587e0-9fb7-492e-8bb4-f218e8e4ce6c"
	gapBrunaOracle              = "bce7e45d-b2c5-43aa-91a6-3a3c24f30850"
	gapDauntlessCatharOracle    = "c14bd133-ff17-48b3-a1c8-56c93d8c6e43"
	gapDrogskolShieldmateOracle = "470028cf-0b99-4dd1-bb9d-aba91c52ee28"
	gapGiselaOracle             = "f3e23d5e-bd88-4e7c-a3fb-db2a8cb05b22"
	gapGuardianPilgrimsOracle   = "a090bc9c-d253-4a29-87d6-aa849f856d36"
	gapInspiringCaptainOracle   = "4a294001-645b-4691-bb32-c240ba5b5320"
)

func TestGapB13EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapTowerGeistOracle:         "Tower Geist",
		gapSpellQuellerOracle:       "Spell Queller",
		gapApothecaryGeistOracle:    "Apothecary Geist",
		gapAvacynianPriestOracle:    "Avacynian Priest",
		gapBriselaOracle:            "Brisela, Voice of Nightmares",
		gapBrunaOracle:              "Bruna, the Fading Light",
		gapDauntlessCatharOracle:    "Dauntless Cathar",
		gapDrogskolShieldmateOracle: "Drogskol Shieldmate",
		gapGiselaOracle:             "Gisela, the Broken Blade",
		gapGuardianPilgrimsOracle:   "Guardian of Pilgrims",
		gapInspiringCaptainOracle:   "Inspiring Captain",
	}
	for oracle, name := range want {
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

func TestGapB13PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, c := range []struct {
		name, tl, oracle string
		kw               []string
	}{
		{"Tower Geist", "Creature — Spirit", gapTowerGeistOracle, []string{"flying"}},
		{"Spell Queller", "Creature — Spirit", gapSpellQuellerOracle, []string{"flash", "flying"}},
		{"Brisela, Voice of Nightmares", "Legendary Creature — Eldrazi Angel", gapBriselaOracle, []string{"flying", "first strike", "vigilance", "lifelink"}},
		{"Bruna, the Fading Light", "Legendary Creature — Angel Horror", gapBrunaOracle, []string{"flying", "vigilance"}},
		{"Gisela, the Broken Blade", "Legendary Creature — Angel Horror", gapGiselaOracle, []string{"flying", "first strike", "lifelink"}},
		{"Drogskol Shieldmate", "Creature — Spirit Soldier", gapDrogskolShieldmateOracle, []string{"flash"}},
	} {
		id := effectstest.B12Push(g, me.ID, c.name, c.tl, c.oracle, 2, 2)
		effectstest.AssertKeywords(t, g, id, c.kw...)
	}
}

// --- Tower Geist ---

func TestGapB13TowerGeistTakesOneOfTheTopTwoAndMillsTheOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	second := effectstest.PutOnTopForTest(me, "Second Card", "Land")
	top := effectstest.PutOnTopForTest(me, "Top Card", "Artifact")
	yard := me.Graveyard.Size()
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Tower Geist", TypeLine: "Creature — Spirit",
		OracleID: gapTowerGeistOracle, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(second) {
		t.Error("the chosen card is not in hand")
	}
	if me.Hand.Contains(top) || !me.Graveyard.Contains(top) {
		t.Error("the other card is not in the graveyard")
	}
	if me.Graveyard.Size() != yard+1 {
		t.Errorf("graveyard %d, want %d", me.Graveyard.Size(), yard+1)
	}
}

// --- Apothecary Geist ---

func TestGapB13ApothecaryGeistGainsLifeOnlyWithAnotherSpirit(t *testing.T) {
	for _, withSpirit := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if withSpirit {
			effectstest.B12Creature(g, me.ID, "Other Spirit", "Creature — Spirit", 1, 1)
		}
		life := me.Life
		effectstest.CastHoldCreature(t, g, "Apothecary Geist", "Creature — Spirit", gapApothecaryGeistOracle, 2, 3)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if withSpirit {
			want = life + 3
		}
		if me.Life != want {
			t.Errorf("another Spirit %v: life %d, want %d", withSpirit, me.Life, want)
		}
	}
}

// --- Avacynian Priest ---

func TestGapB13AvacynianPriestTapsOnlyANonHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	priest := effectstest.PushCatalogPermanent(g, me.ID, "Avacynian Priest", "Creature — Human Cleric", gapAvacynianPriestOracle, false)
	human := effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human Soldier", 2, 2)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(human), Strict: true}); err == nil {
		t.Fatal("a Human was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, bear) {
		t.Error("the non-Human creature was not tapped")
	}
	if effectstest.TappedForTest(t, g, human) {
		t.Error("the Human was tapped")
	}
	if !effectstest.TappedForTest(t, g, priest) {
		t.Error("the Priest did not tap for its cost")
	}
}

// --- Brisela ---

func TestGapB13BriselaStopsOpponentsCastingSmallSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Brisela, Voice of Nightmares", "Legendary Creature — Eldrazi Angel", gapBriselaOracle, 9, 10)
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	cast := func(cost string) error {
		effectstest.B06AddMana(opp, "R", "R", "R", "R", "R")
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Spell", TypeLine: "Sorcery", ManaCost: cost,
			Owner: opp.ID, Controller: opp.ID})
		effectstest.AdvanceToMain(t, g)
		return g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := cast("{2}{R}"); err == nil {
		t.Error("an opponent cast a mana value 3 spell")
	}
	if err := cast("{3}{R}"); err != nil {
		t.Errorf("an opponent could not cast a mana value 4 spell: %v", err)
	}
}

func TestGapB13BriselaDoesNotStopItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Brisela, Voice of Nightmares", "Legendary Creature — Eldrazi Angel", gapBriselaOracle, 9, 10)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Spell", TypeLine: "Sorcery", ManaCost: "{1}",
		Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("its controller could not cast a cheap spell: %v", err)
	}
}

// --- Bruna, the Fading Light ---

func TestGapB13BrunaReturnsAnAngelOrHumanWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	angel := effectstest.PushCatalogGraveyardCard(me, "Dead Angel", "Creature — Angel", "", 3, 3)
	human := effectstest.PushCatalogGraveyardCard(me, "Dead Human", "Creature — Human Soldier", "", 2, 2)
	bear := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bruna, the Fading Light", TypeLine: "Legendary Creature — Angel Horror",
		OracleID: gapBrunaOracle, Power: 5, Toughness: 7, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, angel)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(human) {
		t.Error("the Human returned too: only one card may return")
	}
	if !g.Battlefield.Contains(angel) || g.Battlefield.Contains(bear) {
		t.Errorf("want the Angel back and no Bear (Angel %v, Bear %v)", g.Battlefield.Contains(angel), g.Battlefield.Contains(bear))
	}
	if g.Battlefield.Contains(bear) {
		t.Error("a Bear returned")
	}
}

// --- Dauntless Cathar ---

func TestGapB13DauntlessCatharMakesASpiritFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, "Dauntless Cathar", "Creature — Human Soldier", gapDauntlessCatharOracle, 3, 2)
	idx := -1
	for i, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: gapDauntlessCatharOracle}) {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no graveyard ability")
	}
	effectstest.B06AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the Cathar was not exiled as a cost")
	}
	spirits := effectstest.BattlefieldIDsNamed(g, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirit tokens, want 1", len(spirits))
	}
	effectstest.AssertKeywords(t, g, spirits[0], "flying")
	if effectstest.EffectivePower(t, g, spirits[0]) != 1 {
		t.Error("the Spirit is not a 1/1")
	}
}

func TestGapB13DauntlessCatharRefusesOutsideTheMainPhaseWithAnEmptyStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, "Dauntless Cathar", "Creature — Human Soldier", gapDauntlessCatharOracle, 3, 2)
	idx := -1
	for i, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: gapDauntlessCatharOracle}) {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			idx = i
		}
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B06AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was activated at beginning of combat")
	}
	// A pool that cannot pay is refused as well.
	effectstest.ToMainPhase(t, g)
}

// --- Drogskol Shieldmate ---

func TestGapB13DrogskolShieldmateGivesTheOthersPlusZeroPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.CastHoldCreature(t, g, "Drogskol Shieldmate", "Creature — Spirit Soldier", gapDrogskolShieldmateOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveToughness(t, g, mine); got != 3 {
		t.Errorf("my other creature has toughness %d, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, theirs); got != 2 {
		t.Errorf("their creature has toughness %d, want 2", got)
	}
	if got := effectstest.EffectiveToughness(t, g, id); got != 3 {
		t.Errorf("the Shieldmate has toughness %d, want 3 (it says other)", got)
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("power %d, want 2", got)
	}
}

// --- Gisela, the Broken Blade ---

func gapMeldPrint() *game.MeldPrint {
	return &game.MeldPrint{
		ResultOracleID:   gapBriselaOracle,
		ResultScryfallID: uuid.NewString(),
		Result: game.Face{
			Name: "Brisela, Voice of Nightmares", TypeLine: "Legendary Creature — Eldrazi Angel",
			Colors: []string{"W"}, Power: 9, Toughness: 10, Keywords: []string{"flying", "first strike", "vigilance", "lifelink"},
		},
	}
}

func gapPushGiselaAndBruna(g *game.Game, giselaOwner, brunaOwner uuid.UUID, brunaName string) (gisela, bruna uuid.UUID) {
	gisela = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gisela, the Broken Blade", OracleID: gapGiselaOracle,
		TypeLine: "Legendary Creature — Angel Horror", Power: 4, Toughness: 3, Colors: []string{"W"},
		Owner: giselaOwner, Controller: giselaOwner,
		Layout: game.LayoutMeld, Meld: gapMeldPrint(), ScryfallID: uuid.NewString(),
	})
	bruna = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: brunaName, OracleID: gapBrunaOracle,
		TypeLine: "Legendary Creature — Angel Horror", Power: 5, Toughness: 7, Colors: []string{"W"},
		Owner: brunaOwner, Controller: brunaOwner,
		Layout: game.LayoutMeld, Meld: gapMeldPrint(), ScryfallID: uuid.NewString(),
	})
	return gisela, bruna
}

func gapBrisela(g *game.Game) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == gapBriselaOracle {
			return c, true
		}
	}
	return game.Card{}, false
}

func TestGapB13GiselaMeldsWithBrunaAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gisela, bruna := gapPushGiselaAndBruna(g, me.ID, me.ID, "Bruna, the Fading Light")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	m, ok := gapBrisela(g)
	if !ok {
		t.Fatalf("no Brisela on the battlefield: %+v", g.Battlefield.Cards)
	}
	if !m.IsMelded() || len(m.MeldedFrom) != 2 {
		t.Errorf("not melded from two cards: %+v", m.MeldedFrom)
	}
	for _, id := range []uuid.UUID{gisela, bruna} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%v is still on the battlefield beside the melded permanent", id)
		}
	}
}

func TestGapB13GiselaDoesNotMeldWithoutBruna(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gisela, _ := gapPushGiselaAndBruna(g, me.ID, me.ID, "Not Bruna")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := gapBrisela(g); ok {
		t.Error("melded without a creature named Bruna, the Fading Light")
	}
	if !g.Battlefield.Contains(gisela) {
		t.Error("Gisela left the battlefield")
	}
}

func TestGapB13GiselaDoesNotMeldWithABrunaYouDoNotOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapPushGiselaAndBruna(g, me.ID, opp.ID, "Bruna, the Fading Light")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := gapBrisela(g); ok {
		t.Error("melded with a Bruna owned by another player")
	}
}

// --- Guardian of Pilgrims ---

func TestGapB13GuardianOfPilgrimsPumpsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Guardian of Pilgrims", "Creature — Spirit Cleric", gapGuardianPilgrimsOracle, 2, 2)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, bear) != 3 || effectstest.EffectiveToughness(t, g, bear) != 3 {
		t.Errorf("bear is %d/%d, want 3/3", effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear))
	}
}

// --- Inspiring Captain ---

func TestGapB13InspiringCaptainPumpsYourCreaturesIncludingItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.CastHoldCreature(t, g, "Inspiring Captain", "Creature — Human Knight", gapInspiringCaptainOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, mine) != 3 || effectstest.EffectiveToughness(t, g, mine) != 3 {
		t.Errorf("my bear is %d/%d, want 3/3", effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine))
	}
	if effectstest.EffectivePower(t, g, id) != 4 {
		t.Errorf("the Captain's power is %d, want 4", effectstest.EffectivePower(t, g, id))
	}
	if effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Errorf("their bear's power is %d, want 2", effectstest.EffectivePower(t, g, theirs))
	}
}

// --- Spell Queller ---

func gapQuellerOverTheirBolt(t *testing.T, g *game.Game, boltCost string) (queller, bolt uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	bolt = uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
		ManaCost: boltCost, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 0 })
	queller = effectstest.CastFromSeat(t, g, me, "Spell Queller", "Creature — Spirit", gapSpellQuellerOracle, nil)
	effectstest.PassUntilOnBattlefield(t, g, queller)
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	return queller, bolt
}

func TestGapB13SpellQuellerExilesASmallSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	_, bolt := gapQuellerOverTheirBolt(t, g, "{R}")
	if !g.Exile.Contains(bolt) {
		t.Fatal("the Bolt was not exiled")
	}
	if me.Life != life {
		t.Errorf("the exiled Bolt still resolved: %d → %d", life, me.Life)
	}
}

func TestGapB13SpellQuellerOffersTheOwnerAFreeCastWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	queller, bolt := gapQuellerOverTheirBolt(t, g, "{R}")
	life := me.Life
	effectstest.B27Kill(g, queller)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, opp.ID, game.MayCastKeywordExileUntil, bolt)
	if err := effectstest.CastThroughResolutionCast(t, g, opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}, Strict: true,
	}); err != nil {
		t.Fatalf("free cast of the exiled Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life > life-3 {
		t.Errorf("life %d, want at most %d after the Bolt", me.Life, life-3)
	}
}

func TestGapB13SpellQuellerCannotTargetASpellWithManaValueFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	big := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: big, Name: "Big Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
		ManaCost: "{4}{R}", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(opp.ID, big, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 0 })
	queller := effectstest.CastFromSeat(t, g, me, "Spell Queller", "Creature — Spirit", gapSpellQuellerOracle, nil)
	effectstest.PassUntilOnBattlefield(t, g, queller)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(big) {
		t.Error("a spell with mana value 5 was exiled")
	}
}
