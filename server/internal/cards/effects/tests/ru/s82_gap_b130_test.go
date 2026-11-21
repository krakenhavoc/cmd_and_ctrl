package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b130_test.go — GAP batch 130 (S82 #5948), The Lord of the Rings:
// Tales of Middle-earth: Forge Anew; Tale of Tinúviel; Isildur's Fateful
// Strike; Shelob's Ambush; Pippin's Bravery; Shower of Arrows; Stew the
// Coneys; Friendly Rivalry; Cast into the Fire; Display of Power; Gimli's Fury;
// Rush the Room.

const (
	gapForgeAnewOracle   = "4df579df-fdbb-418a-af2d-709e140ee569"
	gapTinuvielOracle    = "bce30bc5-b059-42bb-a9a4-5886f5607160"
	gapIsildurOracle     = "a43e98d1-684a-4198-bc7e-877b9f6d5357"
	gapShelobOracle      = "c2816923-9daf-4277-adf7-20427d14ba4a"
	gapPippinOracle      = "c7ec52b1-295a-4daa-86fd-dcbe6fb412d7"
	gapShowerOracle      = "efcbca5a-743c-4b16-a7cf-b07335b31ced"
	gapStewOracle        = "f6f1b8da-5e7c-4121-9b0f-fd6678b04991"
	gapRivalryOracle     = "e6850561-3c57-4354-8850-c467a6449563"
	gapCastFireOracle    = "b24ed296-e17d-4e36-8a86-a370868b0136"
	gapDisplayOracle     = "99140891-face-4015-aacd-1309e87d8f9f"
	gapGimlisFuryOracle  = "1b550c27-14b0-45fb-b17e-fcb061dc6ad0"
	gapRushTheRoomOracle = "223fefa3-64a7-457f-90a2-89973640f554"
)

func TestGapB130CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapForgeAnewOracle:   "Forge Anew",
		gapTinuvielOracle:    "Tale of Tinúviel",
		gapIsildurOracle:     "Isildur's Fateful Strike",
		gapShelobOracle:      "Shelob's Ambush",
		gapPippinOracle:      "Pippin's Bravery",
		gapShowerOracle:      "Shower of Arrows",
		gapStewOracle:        "Stew the Coneys",
		gapRivalryOracle:     "Friendly Rivalry",
		gapCastFireOracle:    "Cast into the Fire",
		gapDisplayOracle:     "Display of Power",
		gapGimlisFuryOracle:  "Gimli's Fury",
		gapRushTheRoomOracle: "Rush the Room",
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

// gapInstant casts an instant from the active seat's hand in the current
// step, strictly, and returns the cast's error.
func gapInstant(t *testing.T, g *game.Game, name, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: true})
}

func gapPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id)
}

// --- Forge Anew ---------------------------------------------------------------

func TestGapForgeAnewReturnsAnEquipmentCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sword, Name: "Bonesplitter", TypeLine: "Artifact — Equipment", OracleID: effectstest.BonesplitterOracle, Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Forge Anew", "Enchantment", gapForgeAnewOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a creature card was a legal target")
	}
	effectstest.PickCard(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, sword) != game.ZoneBattlefield {
		t.Errorf("the Equipment is in %s, want the battlefield", effectstest.B12ZoneOf(g, sword))
	}
	if effectstest.B12ZoneOf(g, bear) != game.ZoneGraveyard {
		t.Errorf("the creature card moved to %s", effectstest.B12ZoneOf(g, bear))
	}
}

func gapEquipAttempt(g *game.Game, who, equipment, creature uuid.UUID) error {
	return g.ActivateCatalogAbility(who, equipment, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}}, Strict: true,
	})
}

func TestGapForgeAnewEquipsAtInstantSpeedDuringYourTurnForFreeOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)
	sword := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", effectstest.BonesplitterOracle)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPool(me, 1)
	if err := gapEquipAttempt(g, me.ID, sword, bear); err == nil {
		t.Fatal("equip was activated in combat without Forge Anew")
	}
	effectstest.B12Push(g, me.ID, "Forge Anew", "Enchantment", gapForgeAnewOracle, 0, 0)
	row := effectstest.EquipRow(t, g, sword)
	tw, ok := effectstest.TwinOf(g, me.ID, sword, row)
	if !ok {
		t.Fatal("no {0} alternative is offered for the first equip of the turn")
	}
	me.ManaPool = game.ManaPool{}
	if err := g.ActivateCatalogAbility(me.ID, sword, tw.Index, game.ActivateAbilityParams{
		Ref: tw.Ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("equip at instant speed for {0}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, sword, bear) {
		t.Fatal("the Equipment did not attach")
	}
	if _, ok := effectstest.TwinOf(g, me.ID, sword, row); ok {
		t.Error("a second {0} equip is offered in the same turn")
	}
	if err := gapEquipAttempt(g, me.ID, sword, other); err == nil {
		t.Error("the second equip was free")
	}
	effectstest.FillPool(me, 1)
	if err := gapEquipAttempt(g, me.ID, sword, other); err != nil {
		t.Errorf("the second equip for its printed {1} at instant speed: %v", err)
	}
}

func TestGapForgeAnewGivesNothingDuringAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)
	sword := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", effectstest.BonesplitterOracle)
	effectstest.B12Push(g, me.ID, "Forge Anew", "Enchantment", gapForgeAnewOracle, 0, 0)
	g.Turn.ActiveSeat = 1
	effectstest.AdvanceToMainOf(t, g, 1)
	if err := gapEquipAttempt(g, me.ID, sword, bear); err == nil {
		t.Error("equip was activated on an opponent's turn")
	}
}

// --- Tale of Tinúviel ---------------------------------------------------------

func TestGapTinuvielPlaysItsChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	saga := effectstest.CastCatalogSpell(t, g, "Tale of Tinúviel", "Enchantment — Saga", gapTinuvielOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "indestructible") {
		t.Fatal("chapter I: the creature is not indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	if effectstest.B12ZoneOf(g, mine) != game.ZoneBattlefield {
		t.Error("the indestructible creature was destroyed")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, dead) != game.ZoneBattlefield {
		t.Errorf("chapter II: the creature card is in %s", effectstest.B12ZoneOf(g, dead))
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter III asked for no targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(mine, friend)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "lifelink") || !effectstest.HasKeywordOnBattlefield(t, g, friend, "lifelink") {
		t.Error("chapter III: the two creatures lack lifelink")
	}
	if g.Battlefield.Contains(saga) {
		t.Fatal("the Saga was not sacrificed after III")
	}
	// The Saga is gone: the indestructible grant ended with it.
	if effectstest.HasKeywordOnBattlefield(t, g, mine, "indestructible") {
		t.Error("the creature is still indestructible without the Saga")
	}
}

// --- Isildur's Fateful Strike -------------------------------------------------

func TestGapIsildurNeedsALegendaryCreatureOrPlaneswalkerToCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Isildur's Fateful Strike", gapIsildurOracle, nil, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("cast without a legendary creature")
	}
	effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	if err := gapInstant(t, g, "Isildur's Fateful Strike", gapIsildurOracle, nil, effectstest.CardRefs(theirs)); err != nil {
		t.Fatalf("cast with a legendary creature: %v", err)
	}
}

func gapFillHand(p *game.Player, n int) []uuid.UUID {
	p.Hand.Cards = nil
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Card", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func TestGapIsildurDestroysAndMakesItsControllerExileDownToFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	hand := gapFillHand(opp, 7)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Isildur's Fateful Strike", gapIsildurOracle, nil, effectstest.CardRefs(theirs)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneGraveyard {
		t.Fatalf("the creature is in %s, want the graveyard", effectstest.B12ZoneOf(g, theirs))
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to choose the exiled cards")
	}
	if err := g.ResolveChooseCards(effectstest.LatestChooseCardsFor(g, opp.ID).ID, opp.ID, hand[:2]); err == nil {
		t.Fatal("exiling two cards was accepted, three are over four")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, hand[:3]...)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 4 {
		t.Errorf("the controller has %d cards in hand, want 4", opp.Hand.Size())
	}
	for _, id := range hand[:3] {
		if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("a chosen card is in %s, want exile", effectstest.B12ZoneOf(g, id))
		}
	}
}

func TestGapIsildurAsksNothingAtFourCardsOrFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	gapFillHand(opp, 4)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Isildur's Fateful Strike", gapIsildurOracle, nil, effectstest.CardRefs(theirs)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, opp.ID) != nil || opp.Hand.Size() != 4 {
		t.Errorf("hand %d, prompt %v; want 4 cards and no prompt", opp.Hand.Size(), effectstest.LatestChooseCardsFor(g, opp.ID))
	}
}

// --- Shelob's Ambush ----------------------------------------------------------

func TestGapShelobsAmbushPumpsGivesDeathtouchAndMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Shelob's Ambush", gapShelobOracle, nil, effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := gapPT(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the creature is %d/%d, want 3/4", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("no deathtouch")
	}
	if effectstest.CountBattlefieldByName(g, "Food") != 1 {
		t.Errorf("Foods = %d, want 1", effectstest.CountBattlefieldByName(g, "Food"))
	}
}

// --- Pippin's Bravery ---------------------------------------------------------

func gapPippin(t *testing.T, food, sacrifice bool) (int, int, bool) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	var f uuid.UUID
	if food {
		f = effectstest.PushToken(g, me.ID, effects.FoodToken())
	}
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Pippin's Bravery", gapPippinOracle, nil, effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if food {
		if sacrifice {
			effectstest.AnswerOwnPermanents(t, g, me.ID, f)
		} else {
			effectstest.AnswerOwnPermanents(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
	} else if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("asked to sacrifice a Food with none")
	}
	p, tough := gapPT(t, g, bear)
	return p - 2, tough - 2, food && g.Battlefield.Contains(f)
}

func TestGapPippinsBraveryGivesFourFourForAFoodAndTwoTwoOtherwise(t *testing.T) {
	if p, tough, kept := gapPippin(t, true, true); p != 4 || tough != 4 || kept {
		t.Errorf("sacrificed: +%d/+%d, Food kept %v, want +4/+4 and the Food gone", p, tough, kept)
	}
	if p, tough, kept := gapPippin(t, true, false); p != 2 || tough != 2 || !kept {
		t.Errorf("declined: +%d/+%d, Food kept %v, want +2/+2 and the Food kept", p, tough, kept)
	}
	if p, tough, _ := gapPippin(t, false, false); p != 2 || tough != 2 {
		t.Errorf("no Food: +%d/+%d, want +2/+2", p, tough)
	}
}

// --- Shower of Arrows ---------------------------------------------------------

func TestGapShowerOfArrowsDestroysItsThreeKindsAndScries(t *testing.T) {
	for _, c := range []struct{ name, typeLine string }{
		{"Rock", "Artifact"}, {"Aura Thing", "Enchantment"}, {"Hawk", "Creature — Bird"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		var target uuid.UUID
		if c.name == "Hawk" {
			target = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hawk", TypeLine: c.typeLine,
				Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
		} else {
			target = effectstest.PushPermanentForTest(g, opp.ID, c.name, "", c.typeLine)
		}
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID})
		effectstest.AdvanceToMain(t, g)
		if err := gapInstant(t, g, "Shower of Arrows", gapShowerOracle, nil, effectstest.CardRefs(target)); err != nil {
			t.Fatalf("%s: cast: %v", c.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.B12ZoneOf(g, target) != game.ZoneGraveyard {
			t.Errorf("%s is in %s, want the graveyard", c.name, effectstest.B12ZoneOf(g, target))
		}
		effectstest.AnswerScryKeepAll(t, g, me.ID)
	}
}

func TestGapShowerOfArrowsRefusesAGroundCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Shower of Arrows", gapShowerOracle, nil, effectstest.CardRefs(bear)); err == nil {
		t.Error("a creature without flying was a legal target")
	}
}

// --- Stew the Coneys ----------------------------------------------------------

func TestGapStewTheConeysBitesAndMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 3)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Stew the Coneys", gapStewOracle, nil, effectstest.CardRefs(theirs, mine)); err == nil {
		t.Fatal("the targets were accepted in the wrong order")
	}
	if err := gapInstant(t, g, "Stew the Coneys", gapStewOracle, nil, effectstest.CardRefs(mine, theirs)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneGraveyard {
		t.Errorf("the 2/3 survived 3 damage (%s)", effectstest.B12ZoneOf(g, theirs))
	}
	if c := effectstest.LayeredCard(t, g, mine); c.DamageMarked != 0 {
		t.Errorf("the biter took %d damage, it is a one-sided bite", c.DamageMarked)
	}
	if effectstest.CountBattlefieldByName(g, "Food") != 1 {
		t.Errorf("Foods = %d, want 1", effectstest.CountBattlefieldByName(g, "Food"))
	}
}

// --- Friendly Rivalry ---------------------------------------------------------

func TestGapFriendlyRivalryAddsUpTheTwoPowers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Giant", 5, 5)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Friendly Rivalry", gapRivalryOracle, nil, effectstest.CardRefs(mine, mine, theirs)); err == nil {
		t.Fatal("the same creature was chosen for both of mine")
	}
	if err := gapInstant(t, g, "Friendly Rivalry", gapRivalryOracle, nil, effectstest.CardRefs(mine, legend, theirs)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneGraveyard {
		t.Errorf("the 5/5 took 2+3 damage and is in %s", effectstest.B12ZoneOf(g, theirs))
	}
}

func TestGapFriendlyRivalryRefusesANonLegendarySecondCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Giant", 5, 5)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Friendly Rivalry", gapRivalryOracle, nil, effectstest.CardRefs(mine, plain, theirs)); err == nil {
		t.Error("a non-legendary creature was the second target")
	}
	// With the optional clause left empty the single source deals its power.
	if err := gapInstant(t, g, "Friendly Rivalry", gapRivalryOracle, nil, effectstest.CardRefs(mine, theirs)); err != nil {
		t.Fatalf("cast with one source: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, theirs); c.DamageMarked != 2 {
		t.Errorf("the victim took %d, want 2", c.DamageMarked)
	}
}

// --- Cast into the Fire -------------------------------------------------------

func TestGapCastIntoTheFireDealsOneToEachOfUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 1)
	c3 := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Cast into the Fire", gapCastFireOracle, []int{0}, effectstest.CardRefs(a, b, c3)); err == nil {
		t.Fatal("three targets were accepted")
	}
	if err := gapInstant(t, g, "Cast into the Fire", gapCastFireOracle, []int{0}, effectstest.CardRefs(a, b)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, a).DamageMarked; got != 1 {
		t.Errorf("A took %d, want 1", got)
	}
	if effectstest.B12ZoneOf(g, b) != game.ZoneGraveyard {
		t.Errorf("the 2/1 survived (%s)", effectstest.B12ZoneOf(g, b))
	}
	if got := effectstest.LayeredCard(t, g, c3).DamageMarked; got != 0 {
		t.Errorf("an untargeted creature took %d", got)
	}
}

func TestGapCastIntoTheFireExilesAnArtifactAndOnlyAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.PushPermanentForTest(g, opp.ID, "Rock", "", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Cast into the Fire", gapCastFireOracle, []int{1}, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("a non-artifact creature was a legal target of the exile mode")
	}
	if err := gapInstant(t, g, "Cast into the Fire", gapCastFireOracle, []int{1}, effectstest.CardRefs(rock)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, rock) != game.ZoneExile {
		t.Errorf("the artifact is in %s, want exile", effectstest.B12ZoneOf(g, rock))
	}
}

// --- Display of Power ---------------------------------------------------------

func TestGapDisplayOfPowerCopiesEachTargetedSpellAndCannotBeCopied(t *testing.T) {
	spec, _ := effects.Lookup(gapDisplayOracle)
	if !spec.CantBeCopied {
		t.Error("Display of Power does not say it can't be copied")
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast the bolt: %v", err)
	}
	if err := gapInstant(t, g, "Display of Power", gapDisplayOracle, nil, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err != nil {
		t.Fatalf("cast Display of Power: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Errorf("the opponent is at %d, want %d (the bolt and its copy)", opp.Life, life-6)
	}
}

func TestGapDisplayOfPowerWithNoTargetsStillResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	if err := gapInstant(t, g, "Display of Power", gapDisplayOracle, nil, nil); err != nil {
		t.Fatalf("a zero-target cast was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Gimli's Fury -------------------------------------------------------------

func TestGapGimlisFuryGivesTrampleOnlyToALegendaryCreature(t *testing.T) {
	for _, legendary := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		typeLine := "Creature — Dwarf"
		if legendary {
			typeLine = "Legendary Creature — Dwarf"
		}
		dwarf := effectstest.B12Creature(g, me.ID, "Dwarf", typeLine, 2, 2)
		effectstest.AdvanceToMain(t, g)
		if err := gapInstant(t, g, "Gimli's Fury", gapGimlisFuryOracle, nil, effectstest.CardRefs(dwarf)); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := gapPT(t, g, dwarf); p != 5 || tough != 4 {
			t.Errorf("legendary %v: %d/%d, want 5/4", legendary, p, tough)
		}
		if got := effectstest.HasKeywordOnBattlefield(t, g, dwarf, "trample"); got != legendary {
			t.Errorf("legendary %v: trample %v", legendary, got)
		}
	}
}

// --- Rush the Room ------------------------------------------------------------

func TestGapRushTheRoomGivesHasteOnlyToGoblinsAndOrcs(t *testing.T) {
	for typeLine, haste := range map[string]bool{
		"Creature — Bear": false, "Creature — Goblin": true, "Creature — Orc Army": true,
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		c := effectstest.B12Creature(g, me.ID, "Target", typeLine, 2, 2)
		effectstest.AdvanceToMain(t, g)
		if err := gapInstant(t, g, "Rush the Room", gapRushTheRoomOracle, nil, effectstest.CardRefs(c)); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.EffectivePower(t, g, c); p != 3 {
			t.Errorf("%s: power %d, want 3", typeLine, p)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, c, "first strike") {
			t.Errorf("%s: no first strike", typeLine)
		}
		if got := effectstest.HasKeywordOnBattlefield(t, g, c, "haste"); got != haste {
			t.Errorf("%s: haste %v, want %v", typeLine, got, haste)
		}
	}
}
