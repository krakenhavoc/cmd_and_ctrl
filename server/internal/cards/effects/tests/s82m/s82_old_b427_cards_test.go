package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b427_cards_test.go — Duel Decks: Izzet vs. Golgari (S82 older sets,
// batch 427, #5567). Each test plays the real catalog entry through the engine.

const (
	oldB427Galvanoth  = "0b4fa1ba-9597-4db1-a77c-975ed839f57f"
	oldB427Wurm       = "6e5f908a-f4f1-4718-ac39-34176e401756"
	oldB427Firemind   = "2037659f-2efe-4321-afaf-961bbec35e9e"
	oldB427Intellect  = "8038337b-aff9-49aa-800c-0e12d2688765"
	oldB427Wand       = "4b9e529c-4fbd-4944-828d-7fbc8f26c991"
	oldB427Mortis     = "6de27cba-5500-4448-abef-68e69fe519dc"
	oldB427Gelectrode = "8039df5c-b26b-4277-8226-4c08a5a69fbb"
	oldB427Thunder    = "9c5bc3a7-42fe-4f8f-b799-ece7dd98ea4d"
	oldB427Svogthos   = "a34a70b8-02e5-4e8c-a9e7-b21c5a11dddf"
	oldB427Twilight   = "ae0a1d9c-19cb-42ee-97c3-464e38e84615"
	oldB427Glean      = "bb7389fa-ab01-4ca2-bd18-03cf5b4cedf6"
	oldB427Weevil     = "dfa8de88-9222-435f-a56e-c32e91fd5289"
	oldB427Vacuumelt  = "eb421a1b-147f-422b-bb8d-483a182a4096"
	oldB427Bear       = "Creature — Bear"
)

func TestOldB427EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB427Galvanoth:  "Galvanoth",
		oldB427Wurm:       "Boneyard Wurm",
		oldB427Firemind:   "Invoke the Firemind",
		oldB427Intellect:  "Overwhelming Intellect",
		oldB427Wand:       "Sphinx-Bone Wand",
		oldB427Mortis:     "Vigor Mortis",
		oldB427Gelectrode: "Gelectrode",
		oldB427Thunder:    "Thunderheads",
		oldB427Svogthos:   "Svogthos, the Restless Tomb",
		oldB427Twilight:   "Twilight's Call",
		oldB427Glean:      "Gleancrawler",
		oldB427Weevil:     "Brain Weevil",
		oldB427Vacuumelt:  "Vacuumelt",
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

func oldB427Hand(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

func oldB427Grave(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Galvanoth --------------------------------------------------------------

func oldB427GalvanothUpkeep(t *testing.T, topName, topType, topCost string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Galvanoth", "Creature — Beast", oldB427Galvanoth, 3, 3)
	top := effectstest.SeedLibraryTop(me, topName, topType, topCost)
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, top
}

func TestOldB427GalvanothCastsATopInstantFree(t *testing.T) {
	g, me, top := oldB427GalvanothUpkeep(t, "Free Shock", "Instant", "{R}")
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, top)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, top, game.CastSpellParams{Targets: effectstest.PlayerRefs(g.Seats[1].ID)}); err != nil {
		t.Fatalf("casting the top card: %v", err)
	}
	if !g.Stack.Contains(top) {
		t.Error("the card is not on the stack")
	}
}

func TestOldB427GalvanothLeavesADeclinedCardOnTop(t *testing.T) {
	g, me, top := oldB427GalvanothUpkeep(t, "Free Shock", "Instant", "{R}")
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if c, err := me.Library.Top(); err != nil || c.InstanceID != top {
		t.Error("the declined card did not stay on top of the library")
	}
}

func TestOldB427GalvanothOffersNothingForACreatureOrALand(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, cost string }{
		{"Top Bear", oldB427Bear, "{1}{G}"},
		{"Top Forest", "Basic Land — Forest", ""},
	} {
		g, me, top := oldB427GalvanothUpkeep(t, tc.name, tc.typeLine, tc.cost)
		if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
			t.Errorf("%s: a cast was offered", tc.name)
		}
		if c, err := me.Library.Top(); err != nil || c.InstanceID != top {
			t.Errorf("%s: it left the top of the library", tc.name)
		}
	}
}

// --- Boneyard Wurm ----------------------------------------------------------

func TestOldB427BoneyardWurmCountsCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := effectstest.B12Push(g, me.ID, "Boneyard Wurm", "Creature — Wurm", oldB427Wurm, 0, 0)
	oldB427Grave(me, "Dead A", oldB427Bear)
	oldB427Grave(me, "Dead B", oldB427Bear)
	oldB427Grave(me, "Spell", "Sorcery")
	oldB427Grave(opp, "Their Dead", oldB427Bear)
	if p, tough := effectstest.PtOf(t, g, wurm); p != 2 || tough != 2 {
		t.Errorf("Wurm is %d/%d, want 2/2 (mine only, creatures only)", p, tough)
	}
	oldB427Grave(me, "Dead C", oldB427Bear)
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.PtOf(t, g, wurm); p != 3 || tough != 3 {
		t.Errorf("Wurm is %d/%d, want 3/3", p, tough)
	}
}

// --- Invoke the Firemind ----------------------------------------------------

func TestOldB427InvokeTheFiremindDrawsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	hand := me.Hand.Size()
	effectstest.B19CastXModal(t, g, "Invoke the Firemind", "Sorcery", oldB427Firemind, "{X}{U}{U}{R}", 3, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

func TestOldB427InvokeTheFiremindDealsXToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	effectstest.B19CastXModal(t, g, "Invoke the Firemind", "Sorcery", oldB427Firemind, "{X}{U}{U}{R}", 4, []int{1}, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 4 {
		t.Errorf("the target took %d, want 4", got)
	}
}

// --- Overwhelming Intellect -------------------------------------------------

func TestOldB427OverwhelmingIntellectCountersACreatureSpellAndDrawsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	bear := oldB427Hand(me, "Three Bear", oldB427Bear, "{2}{G}", "")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Bear: %v", err)
	}
	intellect := oldB427Hand(me, "Overwhelming Intellect", "Instant", "{4}{U}{U}", oldB427Intellect)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, intellect, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast Overwhelming Intellect: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
	if got := me.Hand.Size() - (hand - 1); got != 3 {
		t.Errorf("drew %d, want 3 (the Bear's mana value)", got)
	}
	_ = opp
}

func TestOldB427OverwhelmingIntellectCantTargetANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	intellect := oldB427Hand(me, "Overwhelming Intellect", "Instant", "{4}{U}{U}", oldB427Intellect)
	if err := g.CastSpell(me.ID, intellect, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}); err == nil {
		t.Error("it targeted an instant spell")
	}
}

// --- Sphinx-Bone Wand -------------------------------------------------------

func oldB427WandCast(t *testing.T, g *game.Game, me, target *game.Player, answer bool) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Some Instant", "Instant", "", nil)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PbPlayer(target.ID)); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, answer)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB427SphinxBoneWandGrowsWithEachInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wand := effectstest.PushCatalogPermanent(g, me.ID, "Sphinx-Bone Wand", "Artifact", oldB427Wand, false)
	life := opp.Life
	oldB427WandCast(t, g, me, opp, true)
	if got := life - opp.Life; got != 1 {
		t.Fatalf("first trigger dealt %d, want 1", got)
	}
	life = opp.Life
	oldB427WandCast(t, g, me, opp, true)
	if got := life - opp.Life; got != 2 {
		t.Errorf("second trigger dealt %d, want 2", got)
	}
	if got := effectstest.FindBattlefieldCardForTest(g, wand).Counters["charge"]; got != 2 {
		t.Errorf("%d charge counters, want 2", got)
	}
}

func TestOldB427SphinxBoneWandDecliningDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wand := effectstest.PushCatalogPermanent(g, me.ID, "Sphinx-Bone Wand", "Artifact", oldB427Wand, false)
	life := opp.Life
	oldB427WandCast(t, g, me, opp, false)
	if opp.Life != life || effectstest.FindBattlefieldCardForTest(g, wand).Counters["charge"] != 0 {
		t.Errorf("declining still put a counter or dealt damage (life %d -> %d)", life, opp.Life)
	}
}

// --- Vigor Mortis -----------------------------------------------------------

func oldB427CastMortis(t *testing.T, mana string) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dead := oldB427Grave(me, "Dead Bear", oldB427Bear)
	spell := oldB427Hand(me, "Vigor Mortis", "Sorcery", "{2}{B}{B}", oldB427Mortis)
	effectstest.FloatForTest(g, me, mana)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(dead)}); err != nil {
		t.Fatalf("cast Vigor Mortis: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature did not return to the battlefield")
	}
	return g, dead
}

func TestOldB427VigorMortisAddsACounterWhenGreenWasSpent(t *testing.T) {
	g, dead := oldB427CastMortis(t, "BBGC")
	if got := effectstest.Pr7Counters(g, dead); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1 ({G} was spent)", got)
	}
}

func TestOldB427VigorMortisAddsNoCounterWithoutGreen(t *testing.T) {
	g, dead := oldB427CastMortis(t, "BBCC")
	if got := effectstest.Pr7Counters(g, dead); got != 0 {
		t.Errorf("%d +1/+1 counters, want 0 (no {G} spent)", got)
	}
}

// --- Gelectrode -------------------------------------------------------------

func TestOldB427GelectrodePingsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gel := effectstest.PushCatalogPermanent(g, me.ID, "Gelectrode", "Creature — Weird", oldB427Gelectrode, false)
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, gel, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 1 {
		t.Errorf("dealt %d, want 1", got)
	}
	if !effectstest.PermanentTapped(g, gel) {
		t.Error("the {T} cost did not tap it")
	}
}

func TestOldB427GelectrodeMayUntapWhenYouCastAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gel := effectstest.PushCatalogPermanent(g, me.ID, "Gelectrode", "Creature — Weird", oldB427Gelectrode, false)
	effectstest.FindBattlefieldCardByID(g, gel).Tapped = true
	effectstest.CastCatalogSpell(t, g, "Some Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PermanentTapped(g, gel) {
		t.Error("it did not untap")
	}
}

func TestOldB427GelectrodeIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gel := effectstest.PushCatalogPermanent(g, me.ID, "Gelectrode", "Creature — Weird", oldB427Gelectrode, false)
	effectstest.FindBattlefieldCardByID(g, gel).Tapped = true
	effectstest.CastCatalogSpell(t, g, "Some Bear", oldB427Bear, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PermanentTapped(g, gel) {
		t.Error("a creature spell untapped it")
	}
}

// --- Thunderheads -----------------------------------------------------------

func TestOldB427ThunderheadsMakesAWeirdPerCopyAndExilesThemAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spell := oldB427Hand(me, "Thunderheads", "Instant", "{2}{U}", oldB427Thunder)
	effectstest.FloatForTest(g, me, "UUUCCC")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{OptionalCosts: []int{0}, Strict: true}); err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	weirds := effectstest.MidB03Tokens(g, me.ID, "Weird")
	if len(weirds) != 2 {
		t.Fatalf("%d Weird tokens, want 2 (the spell and its copy)", len(weirds))
	}
	for _, w := range weirds {
		if p, tough := effectstest.PtOf(t, g, w.InstanceID); p != 3 || tough != 3 {
			t.Errorf("a Weird is %d/%d, want 3/3", p, tough)
		}
		for _, kw := range []string{"defender", "flying"} {
			if !effectstest.HasKeywordOnBattlefield(t, g, w.InstanceID, kw) {
				t.Errorf("a Weird has no %s", kw)
			}
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MidB03Tokens(g, me.ID, "Weird"); len(got) != 0 {
		t.Errorf("%d Weird tokens remain after the end step, want 0", len(got))
	}
}

// --- Svogthos, the Restless Tomb --------------------------------------------

func TestOldB427SvogthosTapsForColourless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.S58p6Push(g, me.ID, "Svogthos, the Restless Tomb", "Land", oldB427Svogthos, "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %v, want one mana", me.ManaPool.Units())
	}
}

func TestOldB427SvogthosBecomesAPlantZombieSizedByYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Svogthos, the Restless Tomb", "Land", oldB427Svogthos, false)
	oldB427Grave(me, "Dead A", oldB427Bear)
	oldB427Grave(me, "Dead B", oldB427Bear)
	oldB427Grave(me, "Spell", "Sorcery")
	effectstest.ApaMana(me, "C", "C", "C", "B", "G")
	effectstest.ApaActivate(t, g, me, land, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, land)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
	for _, st := range []string{"Plant", "Zombie"} {
		if !c.HasSubtype(st) {
			t.Errorf("no %s subtype", st)
		}
	}
	if cols := c.EffectiveColors(); len(cols) != 2 || cols[0] != "B" || cols[1] != "G" {
		t.Errorf("colours %v, want black and green", cols)
	}
	if p, tough := effectstest.PtOf(t, g, land); p != 2 || tough != 2 {
		t.Errorf("it is %d/%d, want 2/2", p, tough)
	}
	oldB427Grave(me, "Dead C", oldB427Bear)
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.PtOf(t, g, land); p != 3 || tough != 3 {
		t.Errorf("after another creature card hit the graveyard it is %d/%d, want 3/3", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.S58p6View(t, g, land).IsCreature() {
		t.Error("it is still a creature after the turn")
	}
}

// --- Twilight's Call --------------------------------------------------------

func TestOldB427TwilightsCallReturnsEveryPlayersCreaturesUnderTheirOwnControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := oldB427Grave(me, "My Dead", oldB427Bear)
	theirs := oldB427Grave(opp, "Their Dead", oldB427Bear)
	spellCard := oldB427Grave(me, "A Sorcery", "Sorcery")
	spell := oldB427Hand(me, "Twilight's Call", "Sorcery", "{4}{B}{B}", oldB427Twilight)
	effectstest.FloatForTest(g, me, "BBCCCC")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || c.Controller != me.ID {
		t.Error("my creature did not return under my control")
	}
	if c := effectstest.FindBattlefieldCardByID(g, theirs); c == nil || c.Controller != opp.ID {
		t.Error("their creature did not return under their control")
	}
	if !me.Graveyard.Contains(spellCard) {
		t.Error("a noncreature card left the graveyard")
	}
}

func TestOldB427TwilightsCallHasFlashOnlyForTwoMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := oldB427Hand(me, "Twilight's Call", "Sorcery", "{4}{B}{B}", oldB427Twilight)
	effectstest.FloatForTest(g, me, "BBCCCC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast at instant speed for its printed cost")
	}
	effectstest.FloatForTest(g, me, "BBCCCCCC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast at instant speed paying {2} more: %v", err)
	}
}

// --- Gleancrawler -----------------------------------------------------------

func TestOldB427GleancrawlerReturnsOnlyCreaturesThatDiedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crawler := effectstest.B12Push(g, me.ID, "Gleancrawler", "Creature — Insect Horror", oldB427Glean, 6, 6)
	if !effectstest.HasKeywordOnBattlefield(t, g, crawler, "trample") {
		t.Error("no trample")
	}
	old := oldB427Grave(me, "Old Dead", oldB427Bear) // in the graveyard since before this turn's log
	died := effectstest.B12Creature(g, me.ID, "Fresh Dead", oldB427Bear, 2, 2)
	effectstest.B25Destroy(g, died)
	discarded := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: discarded, Name: "Discarded", TypeLine: oldB427Bear, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, discarded); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(died) {
		t.Error("the creature that died this turn did not return to hand")
	}
	if me.Hand.Contains(old) || !me.Graveyard.Contains(old) {
		t.Error("a creature card that was already in the graveyard returned")
	}
	if me.Hand.Contains(discarded) || !me.Graveyard.Contains(discarded) {
		t.Error("a creature card that was put there from the hand returned")
	}
}

// --- Brain Weevil -----------------------------------------------------------

func TestOldB427BrainWeevilSacrificesToMakeThemDiscardTwoAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	weevil := effectstest.PushCatalogPermanent(g, me.ID, "Brain Weevil", "Creature — Insect", oldB427Weevil, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, weevil, "intimidate") {
		t.Error("no intimidate")
	}
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	hand := opp.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, weevil, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, weevil) {
		t.Error("the Weevil was not sacrificed")
	}
	if c := effectstest.LatestChooseCardsFor(g, opp.ID); c != nil {
		if err := g.ResolveChooseCards(c.ID, opp.ID, c.ChooseCards[:2]); err != nil {
			t.Fatal(err)
		}
	}
	if got := hand - opp.Hand.Size(); got != 2 {
		t.Errorf("the opponent discarded %d, want 2", got)
	}
}

func TestOldB427BrainWeevilCantBeActivatedAtInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	weevil := effectstest.PushCatalogPermanent(g, me.ID, "Brain Weevil", "Creature — Insect", oldB427Weevil, false)
	if err := g.ActivateCatalogAbility(me.ID, weevil, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Vacuumelt --------------------------------------------------------------

func TestOldB427VacuumeltBouncesOnePerReplicate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", oldB427Bear, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", oldB427Bear, 2, 2)
	spell := oldB427Hand(me, "Vacuumelt", "Sorcery", "{2}{U}", oldB427Vacuumelt)
	effectstest.FloatForTest(g, me, "UUUCCC")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardTarget(a), OptionalCosts: []int{0}, Strict: true}); err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardTarget(b)); err != nil {
				t.Fatalf("ResolvePickTargets: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a targeted creature is still on the battlefield")
	}
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("a bounced creature is not in its owner's hand")
	}
}
