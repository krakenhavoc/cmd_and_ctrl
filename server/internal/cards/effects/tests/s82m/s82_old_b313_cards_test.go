package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b313_cards_test.go — Battle for Zendikar, batch 313 (S82, #5547).
// Each test plays the real catalog entry through the engine.

const (
	oldB313GrovetenderOracle = "1fc8b186-bb74-4fe1-a2da-e8ce3074f988"
	oldB313MundaOracle       = "4d5da73a-f22a-4446-b5f9-504dfabf1d52"
	oldB313EntanglersOracle  = "6fedd98c-9c58-4ca5-b77b-dcb4cd88d238"
	oldB313HealerOracle      = "b70718c3-a959-402a-a715-7a3c9724a70a"
	oldB313ProcessorOracle   = "1cf82e44-99dd-48d3-a272-8df9835ebc0d"
	oldB313MedicOracle       = "e4d6ed72-f40f-4e4a-bec0-19b1d4f99ebe"
	oldB313EmissaryOracle    = "ef49b53d-e5b6-423a-88f0-51a7b2021c46"
	oldB313TidecallerOracle  = "3aa709b5-0465-4103-aea1-8cfb15086f76"
	oldB313BehemothOracle    = "78ee8cd8-de12-440f-8890-a78ab40b6be4"
	oldB313MarchOracle       = "9d2b9f3b-9ba1-4f28-8cd9-490803190e24"
	oldB313EarthenArmsOracle = "2146fa14-dc6a-4e13-bed0-dda235784ee6"
)

func TestOldB313EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB313GrovetenderOracle: "Grovetender Druids",
		oldB313MundaOracle:       "Munda, Ambush Leader",
		oldB313EntanglersOracle:  "Kor Entanglers",
		oldB313HealerOracle:      "Kalastria Healer",
		oldB313ProcessorOracle:   "Ruin Processor",
		oldB313MedicOracle:       "Stone Haven Medic",
		oldB313EmissaryOracle:    "Drana's Emissary",
		oldB313TidecallerOracle:  "Halimar Tidecaller",
		oldB313BehemothOracle:    "Deathless Behemoth",
		oldB313MarchOracle:       "March from the Tomb",
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

// --- Grovetender Druids ----------------------------------------------------

func TestOldB313GrovetenderDruidsPaysOneForAPlant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	effectstest.CastHoldCreature(t, g, "Grovetender Druids", "Creature — Elf Druid Ally", oldB313GrovetenderOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Plant", me.ID); n != 1 {
		t.Fatalf("%d Plants, want 1", n)
	}
	plant := effectstest.FindBattlefieldByName(g, "Plant")
	if p, tough := effectstest.PtOf(t, g, plant); p != 1 || tough != 1 {
		t.Errorf("the Plant is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB313GrovetenderDruidsDecliningMakesNoPlant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	effectstest.CastHoldCreature(t, g, "Grovetender Druids", "Creature — Elf Druid Ally", oldB313GrovetenderOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Plant", me.ID); n != 0 {
		t.Errorf("%d Plants after declining, want 0", n)
	}
}

// --- Kor Entanglers --------------------------------------------------------

func TestOldB313KorEntanglersTapsAnOpponentsCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Kor Entanglers", "Creature — Kor Soldier Ally", oldB313EntanglersOracle, 3, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("my own creature is offered as a target")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedOnBattlefield(t, g, theirs) {
		t.Error("their creature was not tapped")
	}
	if effectstest.TappedOnBattlefield(t, g, mine) {
		t.Error("my creature was tapped")
	}
}

// --- Kalastria Healer, Drana's Emissary ------------------------------------

func TestOldB313KalastriaHealerDrainsForEachAllyEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastHoldCreature(t, g, "Kalastria Healer", "Creature — Vampire Cleric Ally", oldB313HealerOracle, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 || me.Life != 41 {
		t.Fatalf("after the Healer itself: opponent %d, me %d, want 39 and 41", opp.Life, me.Life)
	}
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("a non-Ally entering drained: opponent %d, want 39", opp.Life)
	}
	effectstest.CastHoldCreature(t, g, "Ally Bear", "Creature — Bear Ally", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 38 || me.Life != 42 {
		t.Errorf("after another Ally: opponent %d, me %d, want 38 and 42", opp.Life, me.Life)
	}
}

func TestOldB313DranasEmissaryDrainsEachUpkeepAndFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spec, _ := effects.Lookup(oldB313EmissaryOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flying" {
		t.Errorf("printed keywords %v, want flying", spec.PrintedKeywords)
	}
	id := effectstest.B12Push(g, me.ID, "Drana's Emissary", "Creature — Vampire Cleric Ally", oldB313EmissaryOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("no flying")
	}
	// Entering is not a trigger: only the upkeep drains.
	effectstest.CastHoldCreature(t, g, "Ally Bear", "Creature — Bear Ally", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40 {
		t.Fatalf("an Ally entering drained: opponent %d, want 40", opp.Life)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 || me.Life != 41 {
		t.Errorf("after the upkeep: opponent %d, me %d, want 39 and 41", opp.Life, me.Life)
	}
}

// --- Stone Haven Medic -----------------------------------------------------

func TestOldB313StoneHavenMedicGainsOneForWAndTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	medic := effectstest.PushCatalogPermanent(g, me.ID, "Stone Haven Medic", "Creature — Kor Cleric", oldB313MedicOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, medic, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without paying {W}")
	}
	effectstest.FloatForTest(g, me, "W")
	effectstest.ApaActivate(t, g, me, medic, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 41 {
		t.Errorf("life %d, want 41", me.Life)
	}
	if !effectstest.TappedOnBattlefield(t, g, medic) {
		t.Error("the Medic is untapped")
	}
}

// --- Ruin Processor --------------------------------------------------------

func TestOldB313RuinProcessorPutsAnOpponentsExiledCardInTheirGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.PushExiledCard(g, opp.ID, "Their Card", "Creature — Bear", false)
	mine := effectstest.PushExiledCard(g, me.ID, "My Card", "Creature — Bear", false)
	effectstest.FloatForTest(g, me, "CCCCCCC")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ruin Processor", TypeLine: "Creature — Eldrazi Processor", ManaCost: "{7}", OracleID: oldB313ProcessorOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// The cast trigger goes on the stack above the spell and resolves first.
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.ChooseCardsPrompt(t, g)
	if effectstest.HasID(p.ChooseCards, mine) {
		t.Error("my own exiled card is offered")
	}
	if !effectstest.HasID(p.ChooseCards, theirs) {
		t.Fatal("the opponent's exiled card is not offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, theirs)
	if !opp.Graveyard.Contains(theirs) {
		t.Error("their card is not in their graveyard")
	}
	if me.Life != 45 {
		t.Errorf("life %d, want 45", me.Life)
	}
}

func TestOldB313RuinProcessorDecliningGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.PushExiledCard(g, opp.ID, "Their Card", "Creature — Bear", false)
	effectstest.FloatForTest(g, me, "CCCCCCC")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ruin Processor", TypeLine: "Creature — Eldrazi Processor", ManaCost: "{7}", OracleID: oldB313ProcessorOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if opp.Graveyard.Contains(theirs) {
		t.Error("the card moved though the player declined")
	}
	if me.Life != 40 {
		t.Errorf("life %d, want 40", me.Life)
	}
}

// --- Halimar Tidecaller ----------------------------------------------------

func TestOldB313HalimarTidecallerReturnsOnlyACardWithAwaken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := effectstest.GraveyardCardOf(g, me, "Plain Spell", "Sorcery", "{1}", "")
	awaken := effectstest.GraveyardCardOf(g, me, "Earthen Arms", "Sorcery", "{2}{G}", oldB313EarthenArmsOracle)
	effectstest.CastHoldCreature(t, g, "Halimar Tidecaller", "Creature — Human Wizard Ally", oldB313TidecallerOracle, 2, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, plain) {
		t.Error("a card without awaken is offered")
	}
	effectstest.PickCard(t, g, me.ID, awaken)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(awaken) {
		t.Error("the card with awaken was not returned")
	}
	if me.Hand.Contains(plain) {
		t.Error("the card without awaken was returned")
	}
}

func TestOldB313HalimarTidecallerGivesLandCreaturesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Halimar Tidecaller", "Creature — Human Wizard Ally", oldB313TidecallerOracle, 2, 3)
	land := effectstest.B12Push(g, me.ID, "Forest", "Land Creature — Forest", "", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirLand := effectstest.B12Push(g, opp.ID, "Island", "Land Creature — Island", "", 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, land, "flying") {
		t.Error("my land creature has no flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("a non-land creature gained flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirLand, "flying") {
		t.Error("an opponent's land creature gained flying")
	}
}

// --- Deathless Behemoth ----------------------------------------------------

func TestOldB313DeathlessBehemothReturnsForTwoScionsAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(oldB313BehemothOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "vigilance" {
		t.Errorf("printed keywords %v, want vigilance", spec.PrintedKeywords)
	}
	card := effectstest.GraveyardCardOf(g, me, "Deathless Behemoth", "Creature — Eldrazi", "{6}", oldB313BehemothOracle)
	scionA := effectstest.B12Push(g, me.ID, "Eldrazi Scion", "Token Creature — Eldrazi Scion", "", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with a single Scion")
	}
	scionB := effectstest.B12Push(g, me.ID, "Eldrazi Scion", "Token Creature — Eldrazi Scion", "", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	_ = bear
	effectstest.ApaActivate(t, g, me, card, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{scionA, scionB}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, card) {
		t.Error("the Behemoth did not return to the hand")
	}
	if _, ok := effectstest.BattlefieldCard(g, scionA); ok {
		t.Error("a Scion was not sacrificed")
	}
}

// --- March from the Tomb ---------------------------------------------------

func TestOldB313MarchFromTheTombReturnsAlliesUpToTotalEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.GraveyardCardOf(g, me, "Ally Five", "Creature — Human Ally", "{4}{W}", "")
	b := effectstest.GraveyardCardOf(g, me, "Ally Three", "Creature — Human Ally", "{2}{W}", "")
	c := effectstest.GraveyardCardOf(g, me, "Ally Two", "Creature — Human Ally", "{1}{W}", "")
	bear := effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}", "")
	effectstest.FloatForTest(g, me, "CCCWB")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "March from the Tomb", TypeLine: "Sorcery", ManaCost: "{3}{W}{B}", OracleID: oldB313MarchOracle, Owner: me.ID, Controller: me.ID})
	tgt := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, i := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: i})
		}
		return out
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: tgt(a, b, c)}); err == nil {
		t.Fatal("total mana value 10 was a legal set")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: tgt(bear)}); err == nil {
		t.Fatal("a non-Ally creature card was a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: tgt(a, b)}); err != nil {
		t.Fatalf("cast with total 8: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, want := range []uuid.UUID{a, b} {
		if _, ok := effectstest.BattlefieldCard(g, want); !ok {
			t.Errorf("a chosen Ally did not return to the battlefield")
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, c); ok {
		t.Error("an unchosen Ally returned")
	}
}

// --- Munda, Ambush Leader --------------------------------------------------

func TestOldB313MundaPutsRevealedAlliesOnTopAndTheRestOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spec, _ := effects.Lookup(oldB313MundaOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "haste" {
		t.Errorf("printed keywords %v, want haste", spec.PrintedKeywords)
	}
	// Top first: Ally 1, Bear 1, Ally 2, Bear 2, then filler.
	filler := effectstest.CardNamed("Filler", "Creature — Bear")
	filler.InstanceID = uuid.New()
	effectstest.LibraryTop(me, filler)
	var ids [4]uuid.UUID
	for i, c := range []game.Card{
		effectstest.CardNamed("Bear 2", "Creature — Bear"),
		effectstest.CardNamed("Ally 2", "Creature — Human Ally"),
		effectstest.CardNamed("Bear 1", "Creature — Bear"),
		effectstest.CardNamed("Ally 1", "Creature — Human Ally"),
	} {
		c.InstanceID = uuid.New()
		ids[3-i] = c.InstanceID
		effectstest.LibraryTop(me, c)
	}
	ally1, bear1, ally2, bear2 := ids[0], ids[1], ids[2], ids[3]
	effectstest.CastHoldCreature(t, g, "Munda, Ambush Leader", "Legendary Creature — Kor Ally", oldB313MundaOracle, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	p := effectstest.ChooseCardsPrompt(t, g)
	if effectstest.HasID(p.ChooseCards, bear1) || effectstest.HasID(p.ChooseCards, bear2) {
		t.Error("a non-Ally card can be revealed")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ally1, ally2)
	// Bottom order for the two Bears, then top order for the two Allies.
	order := effectstest.PutInLibraryPrompt(g)
	if order == nil {
		t.Fatal("no ordering prompt for the rest")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{bear2, bear1}, nil); err != nil {
		t.Fatalf("bottom order: %v", err)
	}
	order = effectstest.PutInLibraryPrompt(g)
	if order == nil {
		t.Fatal("no ordering prompt for the revealed Allies")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{ally2, ally1}); err != nil {
		t.Fatalf("top order: %v", err)
	}
	if got := effectstest.LibraryTopIDs(me, 2); len(got) != 2 || got[0] != ally2 || got[1] != ally1 {
		t.Errorf("library top %v, want [Ally 2, Ally 1]", got)
	}
	n := len(me.Library.Cards)
	if me.Library.Cards[0].InstanceID != bear1 || me.Library.Cards[1].InstanceID != bear2 {
		t.Errorf("the bottom of the library is not Bear 2 then Bear 1 (from the bottom) at %d cards", n)
	}
}
