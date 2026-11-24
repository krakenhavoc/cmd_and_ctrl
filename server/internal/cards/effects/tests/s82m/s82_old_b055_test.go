package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b055_test.go — Commander 2021, batch 55 (#5504): token makers.
// Each test plays the real catalog entry and asserts what its text says.

const (
	oldB055Zaffai        = "02d529df-6299-4e8f-9794-70c8061efdc4"
	oldB055Troll         = "2a3712f9-6bed-4068-9064-2b5240fc7b28"
	oldB055ParadoxZone   = "2fff8441-1f5a-423b-ab98-93016266f5c5"
	oldB055PiaNalaar     = "715336f5-a708-488b-9f30-5702a9013af0"
	oldB055Satchel       = "75a4ddf5-1b3c-4dde-a6e4-4346e46cebf1"
	oldB055Summonings    = "7d5149c6-0e7e-438c-a554-5235c85e2426"
	oldB055Excavation    = "bcd3e6a1-7b76-408d-9236-3791c3e1064a"
	oldB055Contract      = "cc4f1023-09e4-4f27-8712-be72e9694a80"
	oldB055TasteOfDeath  = "f92673fe-cd73-48ae-bfbd-aa750582b80b"
	oldB055StartingLife  = 40
	oldB055TokenElemName = "Elemental"
)

func oldB055Seats(g *game.Game) (*game.Player, *game.Player) {
	i := g.Turn.ActiveSeat
	return g.Seats[i], g.Seats[(i+1)%len(g.Seats)]
}

// oldB055Tokens are the ids of the tokens named `name` that `controller`
// controls.
func oldB055Tokens(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestOldB055EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB055Zaffai:       "Zaffai, Thunder Conductor",
		oldB055Troll:        "Gluttonous Troll",
		oldB055ParadoxZone:  "Paradox Zone",
		oldB055PiaNalaar:    "Pia Nalaar",
		oldB055Satchel:      "Druidic Satchel",
		oldB055Summonings:   "Metallurgic Summonings",
		oldB055Excavation:   "Excavation Technique",
		oldB055Contract:     "Tempting Contract",
		oldB055TasteOfDeath: "Taste of Death",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is %v, want Full", name, spec.Completeness)
		}
	}
}

// --- Zaffai, Thunder Conductor ---------------------------------------------

func oldB055CastSpellWithZaffai(t *testing.T, cost string) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Zaffai, Thunder Conductor", "Legendary Creature — Human Shaman", oldB055Zaffai, false)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Test Spell", "Instant", cost, "", game.CastSpellParams{})
	effectstest.AnswerScryIfAsked(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me
}

func TestOldB055ZaffaiScriesOnALowManaValueSpell(t *testing.T) {
	g, me := oldB055CastSpellWithZaffai(t, "{2}{R}")
	if n := effectstest.CountOnBattlefield(g, oldB055TokenElemName, me.ID); n != 0 {
		t.Errorf("%d Elementals from a mana value 3 spell, want none", n)
	}
	for _, p := range g.Seats {
		if p.Life != oldB055StartingLife {
			t.Errorf("a life total moved: %d", p.Life)
		}
	}
}

func TestOldB055ZaffaiMakesAnElementalAtFiveButNoDamage(t *testing.T) {
	g, me := oldB055CastSpellWithZaffai(t, "{4}{R}")
	if n := effectstest.CountOnBattlefield(g, oldB055TokenElemName, me.ID); n != 1 {
		t.Fatalf("%d Elementals, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == oldB055TokenElemName && c.IsToken() {
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 4 || tough != 4 {
				t.Errorf("the Elemental is %d/%d, want 4/4", p, tough)
			}
		}
	}
	for _, p := range g.Seats {
		if p.Life != oldB055StartingLife {
			t.Errorf("a life total moved: %d", p.Life)
		}
	}
}

func TestOldB055ZaffaiAtTenDealsTenToARandomOpponentAndMakesAnElemental(t *testing.T) {
	g, me := oldB055CastSpellWithZaffai(t, "{9}{R}")
	if n := effectstest.CountOnBattlefield(g, oldB055TokenElemName, me.ID); n != 1 {
		t.Errorf("%d Elementals, want 1", n)
	}
	hit := 0
	for _, p := range g.Seats {
		switch p.Life {
		case oldB055StartingLife:
		case oldB055StartingLife - 10:
			if p.ID == me.ID {
				t.Error("Zaffai's controller was hit")
			}
			hit++
		default:
			t.Errorf("unexpected life %d", p.Life)
		}
	}
	if hit != 1 {
		t.Errorf("%d players took 10, want exactly one opponent", hit)
	}
}

// --- Gluttonous Troll -------------------------------------------------------

func TestOldB055TrollMakesAFoodForEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Gluttonous Troll", "Creature — Troll", "{2}{B}{G}", oldB055Troll, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Food", me.ID); n != len(g.Seats)-1 {
		t.Errorf("%d Foods, want %d (one per opponent)", n, len(g.Seats)-1)
	}
}

func TestOldB055TrollSacrificesAnotherNonlandPermanentForPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	troll := effectstest.B12Push(g, me.ID, "Gluttonous Troll", "Creature — Troll", oldB055Troll, 3, 3)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Basic Land — Swamp")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{troll}}); err == nil {
		t.Fatal("the Troll paid with itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the permanent was not sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, troll); p != 5 || tough != 5 {
		t.Errorf("the Troll is %d/%d, want 5/5", p, tough)
	}
}

// --- Paradox Zone -----------------------------------------------------------

func TestOldB055ParadoxZoneEntersWithAGrowthCounterThenDoublesAndMakesAFractal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	effectstest.AdvanceToMain(t, g)
	zone := effectstest.CastFromHandForTest(t, g, me, "Paradox Zone", "Enchantment", "{4}{G}", oldB055ParadoxZone, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, zone, "growth"); n != 1 {
		t.Fatalf("%d growth counters on entering, want 1", n)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, zone, "growth"); n != 2 {
		t.Errorf("%d growth counters after the first end step, want 2", n)
	}
	frac := oldB055Tokens(g, me.ID, "Fractal")
	if len(frac) != 1 {
		t.Fatalf("%d Fractals, want 1", len(frac))
	}
	if p, tough := effectstest.PtOf(t, g, frac[0]); p != 2 || tough != 2 {
		t.Errorf("the Fractal is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB055ParadoxZoneDoublesAnExistingPileBeforeCounting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	zone := effectstest.PushCatalogPermanent(g, me.ID, "Paradox Zone", "Enchantment", oldB055ParadoxZone, false)
	effectstest.AddCountersForTest(g, zone, "growth", 3)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, zone, "growth"); n != 6 {
		t.Errorf("%d growth counters, want 6", n)
	}
	frac := oldB055Tokens(g, me.ID, "Fractal")
	if len(frac) != 1 {
		t.Fatalf("%d Fractals, want 1", len(frac))
	}
	if p, tough := effectstest.PtOf(t, g, frac[0]); p != 6 || tough != 6 {
		t.Errorf("the Fractal is %d/%d, want 6/6 (after doubling)", p, tough)
	}
}

// --- Pia Nalaar -------------------------------------------------------------

func TestOldB055PiaNalaarMakesAFlyingThopter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Pia Nalaar", "Legendary Creature — Human Artificer", "{2}{R}", oldB055PiaNalaar, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	thopters := oldB055Tokens(g, me.ID, "Thopter")
	if len(thopters) != 1 {
		t.Fatalf("%d Thopters, want 1", len(thopters))
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, thopters[0], "flying") {
		t.Error("the Thopter does not fly")
	}
	if p, tough := effectstest.PtOf(t, g, thopters[0]); p != 1 || tough != 1 {
		t.Errorf("the Thopter is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB055PiaNalaarPumpsOnlyAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	pia := effectstest.B12Push(g, me.ID, "Pia Nalaar", "Legendary Creature — Human Artificer", oldB055PiaNalaar, 2, 2)
	golem := effectstest.B12Push(g, me.ID, "Golem", "Artifact Creature — Golem", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, pia, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(pia)}); err == nil {
		t.Fatal("pumped a creature that is not an artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, pia, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(golem)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, golem); p != 3 || tough != 2 {
		t.Errorf("the Golem is %d/%d, want 3/2", p, tough)
	}
}

func TestOldB055PiaNalaarSacrificesAnArtifactSoACreatureCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB055Seats(g)
	pia := effectstest.B12Push(g, me.ID, "Pia Nalaar", "Legendary Creature — Human Artificer", oldB055PiaNalaar, 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Basic Land — Mountain")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, pia, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(wall), SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land paid the artifact cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, pia, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(wall), SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if !effectstest.MidB03CantBlock(t, g, wall) {
		t.Error("the Wall can still block")
	}
}

// --- Druidic Satchel --------------------------------------------------------

func oldB055ActivateSatchel(t *testing.T, top game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	satchel := effectstest.PushCatalogPermanent(g, me.ID, "Druidic Satchel", "Artifact", oldB055Satchel, false)
	top.InstanceID = uuid.New()
	top.Owner, top.Controller = me.ID, me.ID
	me.Library.PushTop(top)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, satchel, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, top.InstanceID
}

func TestOldB055SatchelCreatureCardMakesASaproling(t *testing.T) {
	g, me, id := oldB055ActivateSatchel(t, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 1 {
		t.Errorf("%d Saprolings, want 1", n)
	}
	if !me.Library.Contains(id) {
		t.Error("the revealed creature card left the library")
	}
	if me.Life != oldB055StartingLife {
		t.Errorf("life %d, want unchanged", me.Life)
	}
}

func TestOldB055SatchelLandCardEntersTheBattlefield(t *testing.T) {
	g, me, id := oldB055ActivateSatchel(t, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if !g.Battlefield.Contains(id) {
		t.Error("the land card did not enter the battlefield")
	}
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 0 {
		t.Errorf("%d Saprolings from a land, want none", n)
	}
}

func TestOldB055SatchelNoncreatureNonlandCardGainsTwoLife(t *testing.T) {
	g, me, id := oldB055ActivateSatchel(t, game.Card{Name: "Shock", TypeLine: "Instant"})
	if me.Life != oldB055StartingLife+2 {
		t.Errorf("life %d, want %d", me.Life, oldB055StartingLife+2)
	}
	if !me.Library.Contains(id) || g.Battlefield.Contains(id) {
		t.Error("the revealed card moved")
	}
}

func TestOldB055SatchelCreatureLandDoesBoth(t *testing.T) {
	g, me, id := oldB055ActivateSatchel(t, game.Card{Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1})
	if !g.Battlefield.Contains(id) {
		t.Error("the creature land did not enter the battlefield")
	}
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 1 {
		t.Errorf("%d Saprolings, want 1", n)
	}
	if me.Life != oldB055StartingLife {
		t.Errorf("life %d, want unchanged", me.Life)
	}
}

// --- Metallurgic Summonings -------------------------------------------------

func TestOldB055SummoningsMakesAnXByXConstructForAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Metallurgic Summonings", "Enchantment", oldB055Summonings, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Test Spell", "Sorcery", "{2}{R}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	constructs := oldB055Tokens(g, me.ID, "Construct")
	if len(constructs) != 1 {
		t.Fatalf("%d Constructs, want 1", len(constructs))
	}
	if p, tough := effectstest.PtOf(t, g, constructs[0]); p != 3 || tough != 3 {
		t.Errorf("the Construct is %d/%d, want 3/3", p, tough)
	}
	c, _ := effectstest.BattlefieldCard(g, constructs[0])
	if !c.IsArtifact() || !c.IsCreature() || len(c.Colors) != 0 {
		t.Errorf("the Construct is %q with colours %v, want a colourless artifact creature", c.TypeLine, c.Colors)
	}
	// A creature spell does not trigger it.
	effectstest.CastFromHandForTest(t, g, me, "Test Bear", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB055Tokens(g, me.ID, "Construct")); got != 1 {
		t.Errorf("%d Constructs after a creature spell, want still 1", got)
	}
}

func TestOldB055SummoningsActivationNeedsSixArtifactsAndReturnsInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	enchantment := effectstest.PushCatalogPermanent(g, me.ID, "Metallurgic Summonings", "Enchantment", oldB055Summonings, false)
	shock := effectstest.B17GraveyardCard(me, "Shock", "Instant", "{R}")
	fireball := effectstest.B17GraveyardCard(me, "Fireball", "Sorcery", "{2}{R}")
	bear := effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	for i := 0; i < 5; i++ {
		effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, enchantment, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with five artifacts")
	}
	effectstest.B12Permanent(g, me.ID, "Sixth Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, enchantment, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with six artifacts: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(shock) || !me.Hand.Contains(fireball) {
		t.Error("an instant or sorcery card did not return")
	}
	if me.Hand.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("the creature card returned")
	}
	if g.Battlefield.Contains(enchantment) || !g.Exile.Contains(enchantment) {
		t.Error("the enchantment was not exiled as a cost")
	}
}

// --- Excavation Technique ---------------------------------------------------

func TestOldB055ExcavationTechniqueDestroysAndItsControllerGetsTwoTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB055Seats(g)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Excavation Technique", TypeLine: "Sorcery", ManaCost: "{3}{W}", OracleID: oldB055Excavation, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTargets(land), Strict: true}); err == nil {
		t.Fatal("targeted a land")
	}
	effectstest.CastFromHandForTest(t, g, me, "Excavation Technique", "Sorcery", "{3}{W}", oldB055Excavation,
		game.CastSpellParams{Targets: effectstest.CardTargets(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the permanent was not destroyed")
	}
	if n := effectstest.CountOnBattlefield(g, "Treasure", opp.ID); n != 2 {
		t.Errorf("its controller got %d Treasures, want 2", n)
	}
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 0 {
		t.Errorf("the caster got %d Treasures, want none", n)
	}
}

// --- Tempting Contract ------------------------------------------------------

func TestOldB055TemptingContractEachAcceptingOpponentGivesTheControllerATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Tempting Contract", "Artifact", oldB055Contract, false)
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	accepted := 0
	for i := 0; i < len(g.Seats); i++ {
		var asked *game.Player
		for _, p := range g.Seats {
			if p.ID != me.ID && effectstest.LatestConfirmFor(g, p.ID) != nil {
				asked = p
				break
			}
		}
		if asked == nil {
			break
		}
		yes := accepted < 2
		effectstest.AnswerMayChoice(t, g, asked.ID, yes)
		if yes {
			accepted++
			if n := effectstest.CountOnBattlefield(g, "Treasure", asked.ID); n != 1 {
				t.Errorf("an accepting opponent has %d Treasures, want 1", n)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if accepted != 2 {
		t.Fatalf("%d opponents were offered and accepted, want 2", accepted)
	}
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 2 {
		t.Errorf("the controller has %d Treasures, want one per accepting opponent (2)", n)
	}
}

func TestOldB055TemptingContractNothingIsCreatedWhenEveryoneDeclines(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB055Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Tempting Contract", "Artifact", oldB055Contract, false)
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < len(g.Seats); i++ {
		var asked *game.Player
		for _, p := range g.Seats {
			if p.ID != me.ID && effectstest.LatestConfirmFor(g, p.ID) != nil {
				asked = p
				break
			}
		}
		if asked == nil {
			break
		}
		effectstest.AnswerMayChoice(t, g, asked.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, p := range g.Seats {
		if n := effectstest.CountOnBattlefield(g, "Treasure", p.ID); n != 0 {
			t.Errorf("%d Treasures for a player after everyone declined", n)
		}
	}
}

// --- Taste of Death ---------------------------------------------------------

func TestOldB055TasteOfDeathMakesEveryPlayerSacrificeThreeThenYouGetThreeFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB055Seats(g)
	mine := make([]uuid.UUID, 4)
	theirs := make([]uuid.UUID, 2)
	for i := range mine {
		mine[i] = effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	}
	for i := range theirs {
		theirs[i] = effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Wolf", 2, 2)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Taste of Death", "Sorcery", "{4}{B}{B}", oldB055TasteOfDeath, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 30; i++ {
		progressed := false
		for _, p := range g.Seats {
			if effectstest.SacrificeChoiceFor(g, p.ID) == nil {
				continue
			}
			prompt := effectstest.SacrificeChoiceFor(g, p.ID)
			var picks []uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Controller == p.ID && c.IsCreature() && len(picks) < max(prompt.ChooseMin, 1) {
					picks = append(picks, c.InstanceID)
				}
			}
			if len(picks) == 0 {
				break
			}
			effectstest.AnswerSacrifice(t, g, p.ID, picks...)
			progressed = true
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !progressed {
			break
		}
	}
	mineLeft := 0
	for _, id := range mine {
		if g.Battlefield.Contains(id) {
			mineLeft++
		}
	}
	if mineLeft != 1 {
		t.Errorf("%d of my four creatures remain, want 1", mineLeft)
	}
	for _, id := range theirs {
		if g.Battlefield.Contains(id) {
			t.Error("an opponent with two creatures kept one")
		}
	}
	if n := effectstest.CountOnBattlefield(g, "Food", me.ID); n != 3 {
		t.Errorf("%d Foods, want 3", n)
	}
}
