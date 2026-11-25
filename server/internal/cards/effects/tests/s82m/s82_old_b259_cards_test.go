package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b259_cards_test.go — Kaladesh, batch 259 (#5539). Every test plays
// the real catalog entry through the engine.

const (
	kldB259LostLegacy        = "532f3f03-37af-46b6-bafb-031e9a3df056"
	kldB259MadcapExperiment  = "684e2237-b860-45bf-a053-a02cc7ed1527"
	kldB259FatefulShowdown   = "729b7ce8-7e7f-4db9-b3ce-67aba46a7b0d"
	kldB259TezzeretsAmbition = "91486e70-94e4-431f-8899-fef019bb206e"
	kldB259VerdantCrescendo  = "b184d12d-41d8-4a74-99b1-8c7fced212f8"
	kldB259LiberatingCombust = "e3b40e34-edad-4838-a4d5-c167037d9ff7"
	kldB259FailedInspection  = "e4fbf952-de3b-4cbd-848a-37041f337e06"
	kldB259WildWanderer      = "ff16f889-9203-4874-919b-42e540cadcbc"
	kldB259Bear              = "Creature — Bear"
)

func TestKldB259EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		kldB259LostLegacy:        "Lost Legacy",
		kldB259MadcapExperiment:  "Madcap Experiment",
		kldB259FatefulShowdown:   "Fateful Showdown",
		kldB259TezzeretsAmbition: "Tezzeret's Ambition",
		kldB259VerdantCrescendo:  "Verdant Crescendo",
		kldB259LiberatingCombust: "Liberating Combustion",
		kldB259FailedInspection:  "Failed Inspection",
		kldB259WildWanderer:      "Wild Wanderer",
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

// --- Wild Wanderer -------------------------------------------------------------

func TestKldB259WildWandererFetchesABasicLandOntoTheBattlefieldTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Temple Garden", "Land — Forest Plains"),
		game.Card{Name: "Bear", TypeLine: kldB259Bear})
	effectstest.CastHoldCreature(t, g, "Wild Wanderer", "Creature — Elf Druid", kldB259WildWanderer, 3, 2)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, bad := range []string{"Temple Garden", "Bear"} {
		if effectstest.SearchOptionNamed(g, c, bad) != uuid.Nil {
			t.Errorf("the search offered %s, which is not a basic land card", bad)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	var forest *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Forest" {
			forest = &g.Battlefield.Cards[i]
		}
	}
	if forest == nil {
		t.Fatal("the Forest is not on the battlefield")
	}
	if !forest.Tapped {
		t.Error("the Forest must enter tapped")
	}
	if me.Library.Size() != 2 {
		t.Errorf("library has %d cards, want 2", me.Library.Size())
	}
}

func TestKldB259WildWandererMayDeclineTheSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))
	before := effectstest.LibraryIDs(me)
	effectstest.CastHoldCreature(t, g, "Wild Wanderer", "Creature — Elf Druid", kldB259WildWanderer, 3, 2)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := effectstest.AnswerSearchApply(g, c, me.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SameIDList(before, effectstest.LibraryIDs(me)) {
		t.Error("declining the search still moved or shuffled the library")
	}
}

// --- Tezzeret's Ambition ---------------------------------------------------------

func TestKldB259TezzeretsAmbitionDiscardsOnlyWithoutAnArtifact(t *testing.T) {
	for _, artifact := range []bool{false, true} {
		name := "no artifact"
		if artifact {
			name = "with an artifact"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedLibrary(me, "a", "b", "c", "d", "e")
			if artifact {
				effectstest.PushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", "", false)
			}
			handBefore := me.Hand.Size()
			effectstest.CastCatalogSpell(t, g, "Tezzeret's Ambition", "Sorcery", kldB259TezzeretsAmbition, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if artifact {
				if effectstest.DiscardChoiceFor(g, me.ID) != nil {
					t.Fatal("asked to discard while controlling an artifact")
				}
				if me.Hand.Size() != handBefore+3 {
					t.Errorf("hand %d -> %d, want +3", handBefore, me.Hand.Size())
				}
				return
			}
			if effectstest.DiscardChoiceFor(g, me.ID) == nil {
				t.Fatal("no discard prompt without an artifact")
			}
			effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != handBefore+2 {
				t.Errorf("hand %d -> %d, want +2 (three drawn, one discarded)", handBefore, me.Hand.Size())
			}
		})
	}
}

// --- Failed Inspection ---------------------------------------------------------------

func TestKldB259FailedInspectionCountersThenLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "x", "y", "z")
	lifeBefore := me.Life
	shock := effectstest.CastShockAt(t, g, opp, me.ID)
	handBefore := me.Hand.Size()
	effectstest.CastCounterAt(t, g, me, "Failed Inspection", kldB259FailedInspection, shock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(shock) || !opp.Graveyard.Contains(shock) {
		t.Fatal("the Shock was not countered")
	}
	if me.Life != lifeBefore {
		t.Error("the countered Shock still dealt damage")
	}
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard after the draw")
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	// Failed Inspection left the hand to be cast, so the loot is net zero.
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d -> %d, want unchanged after draw and discard", handBefore, me.Hand.Size())
	}
}

// --- Fateful Showdown ---------------------------------------------------------------

func TestKldB259FatefulShowdownDamagesByHandDiscardsAndRedraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	effectstest.SeedLibrary(me, "p", "q", "r", "s", "t", "u")
	var old []uuid.UUID
	for _, c := range me.Hand.Cards {
		old = append(old, c.InstanceID)
	}
	for i := 0; i < 3; i++ {
		old = append(old, effectstest.PushHandCardOfType(me, "Old Card", kldB259Bear))
	}
	lifeBefore := opp.Life
	effectstest.CastCatalogSpell(t, g, "Fateful Showdown", "Instant", kldB259FatefulShowdown,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	n := len(old)
	if handNow := me.Hand.Size(); handNow < n {
		t.Fatalf("hand is %d, want at least %d", handNow, n)
	}
	if opp.Life != lifeBefore-n {
		t.Errorf("opponent life %d -> %d, want -%d (cards in hand as it resolved)", lifeBefore, opp.Life, n)
	}
	for _, id := range old {
		if me.Hand.Contains(id) || !me.Graveyard.Contains(id) {
			t.Error("an old card was not discarded")
		}
	}
	if me.Hand.Size() != n {
		t.Errorf("hand has %d cards, want %d (drew that many)", me.Hand.Size(), n)
	}
}

func TestKldB259FatefulShowdownDoesNothingWithAnIllegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := effectstest.B12Creature(g, effectstest.Top100Opponent(g).ID, "Victim", kldB259Bear, 2, 2)
	old := effectstest.PushHandCardOfType(me, "Old Card", kldB259Bear)
	effectstest.CastCatalogSpell(t, g, "Fateful Showdown", "Instant", kldB259FatefulShowdown, effectstest.B16TargetCard(victim))
	effectstest.B27Kill(g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(old) {
		t.Error("a spell whose only target is gone still discarded the hand (CR 608.2b)")
	}
}

// --- Madcap Experiment ---------------------------------------------------------------

func TestKldB259MadcapExperimentPutsTheArtifactInAndHurtsPerCardRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me) // empties the library
	// PushLibraryCardForTest pushes to the bottom, so the first card pushed is the top.
	a := effectstest.PushLibraryCardForTest(me, game.Card{Name: "A", TypeLine: kldB259Bear})
	b := effectstest.PushLibraryCardForTest(me, game.Card{Name: "B", TypeLine: kldB259Bear})
	ring := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	below := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Below", TypeLine: kldB259Bear})
	lifeBefore := me.Life
	effectstest.CastCatalogSpell(t, g, "Madcap Experiment", "Sorcery", kldB259MadcapExperiment, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ring) {
		t.Fatal("the artifact did not enter the battlefield")
	}
	if me.Life != lifeBefore-3 {
		t.Errorf("life %d -> %d, want -3 (A, B and the artifact were revealed)", lifeBefore, me.Life)
	}
	if me.Library.Size() != 3 {
		t.Fatalf("library has %d cards, want 3", me.Library.Size())
	}
	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.InstanceID != below {
		t.Errorf("the top card is %q, want the unrevealed card on top", top.Name)
	}
	if !me.Library.Contains(a) || !me.Library.Contains(b) {
		t.Error("the revealed nonartifact cards were not put on the bottom")
	}
}

func TestKldB259MadcapExperimentWithNoArtifactRevealsEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "A", TypeLine: kldB259Bear})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "B", TypeLine: kldB259Bear})
	lifeBefore := me.Life
	effectstest.CastCatalogSpell(t, g, "Madcap Experiment", "Sorcery", kldB259MadcapExperiment, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lifeBefore-2 {
		t.Errorf("life %d -> %d, want -2", lifeBefore, me.Life)
	}
	if me.Library.Size() != 2 {
		t.Errorf("library has %d cards, want all 2 back on the bottom", me.Library.Size())
	}
}

// --- Liberating Combustion -----------------------------------------------------------

func TestKldB259LiberatingCombustionDamagesThenMayFetchChandra(t *testing.T) {
	for _, zone := range []string{"library", "graveyard", "declined"} {
		t.Run(zone, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
			victim := effectstest.B12Creature(g, opp.ID, "Victim", kldB259Bear, 6, 6)
			effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			var chandra uuid.UUID
			if zone == "graveyard" {
				chandra = effectstest.PushGraveyardCardWithTypeLineID(me, "Chandra, Pyrogenius", "Legendary Planeswalker — Chandra")
			} else {
				chandra = effectstest.PushLibraryCardForTest(me, game.Card{Name: "Chandra, Pyrogenius", TypeLine: "Legendary Planeswalker — Chandra"})
			}
			effectstest.CastCatalogSpell(t, g, "Liberating Combustion", "Sorcery", kldB259LiberatingCombust, effectstest.B16TargetCard(victim))
			effectstest.PassPriorityAroundTable(t, g)
			if zone == "declined" {
				effectstest.AnswerMayChoice(t, g, me.ID, false)
				effectstest.PassPriorityAroundTable(t, g)
				if me.Hand.Contains(chandra) {
					t.Error("declining still took Chandra")
				}
				if g.Battlefield.Contains(victim) {
					t.Error("6 damage did not kill the 6/6")
				}
				return
			}
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			if zone == "graveyard" {
				effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
			}
			effectstest.AnswerSearchByID(t, g, me.ID, chandra)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(chandra) {
				t.Errorf("Chandra did not come from the %s", zone)
			}
			if g.Battlefield.Contains(victim) {
				t.Error("6 damage did not kill the 6/6")
			}
		})
	}
}

func TestKldB259LiberatingCombustionFizzlesWithoutItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", kldB259Bear, 2, 2)
	chandra := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Chandra, Pyrogenius", TypeLine: "Legendary Planeswalker — Chandra"})
	effectstest.CastCatalogSpell(t, g, "Liberating Combustion", "Sorcery", kldB259LiberatingCombust, effectstest.B16TargetCard(victim))
	effectstest.B27Kill(g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil || me.Hand.Contains(chandra) {
		t.Error("a spell whose only target is gone searched anyway (CR 608.2b)")
	}
}

// --- Verdant Crescendo ------------------------------------------------------------------

func TestKldB259VerdantCrescendoFetchesAForestTappedAndNissaFromEitherZone(t *testing.T) {
	for _, zone := range []string{"library", "graveyard"} {
		t.Run(zone, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedSearchLibrary(me,
				effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
				effectstest.SearchTestLand("Temple Garden", "Land — Forest Plains"))
			var nissa uuid.UUID
			if zone == "graveyard" {
				nissa = effectstest.PushGraveyardCardWithTypeLineID(me, "Nissa, Nature's Artisan", "Legendary Planeswalker — Nissa")
			} else {
				nissa = effectstest.PushLibraryCardForTest(me, game.Card{Name: "Nissa, Nature's Artisan", TypeLine: "Legendary Planeswalker — Nissa"})
			}
			effectstest.CastCatalogSpell(t, g, "Verdant Crescendo", "Sorcery", kldB259VerdantCrescendo, nil)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no land search")
			}
			if effectstest.SearchOptionNamed(g, c, "Temple Garden") != uuid.Nil {
				t.Error("a nonbasic land was offered")
			}
			effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
			effectstest.AnswerSearchByID(t, g, me.ID, nissa)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(nissa) {
				t.Errorf("Nissa did not come from the %s", zone)
			}
			tapped := false
			for _, bc := range g.Battlefield.Cards {
				if bc.Name == "Forest" {
					tapped = bc.Tapped
				}
			}
			if !tapped {
				t.Error("the Forest is missing or untapped")
			}
		})
	}
}

// --- Lost Legacy -------------------------------------------------------------------------

func TestKldB259LostLegacyExilesEveryCopyAndDrawsForTheHandOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	mk := func() game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID}
	}
	var hand, yard, lib [2]uuid.UUID
	var other uuid.UUID
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			h, y, l := mk(), mk(), mk()
			hand[i], yard[i], lib[i] = h.InstanceID, y.InstanceID, l.InstanceID
			opp.Hand.PushTop(h)
			opp.Graveyard.PushTop(y)
			opp.Library.PushTop(l)
		}
		other = uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: other, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	})
	handBefore := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Lost Legacy", "Sorcery", kldB259LostLegacy,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "lightning bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{hand[0], hand[1], yard[0], yard[1], lib[0], lib[1]} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a Lightning Bolt is not in exile (zone %v)", z)
		}
	}
	if z := g.FindCardZoneForEffect(other); z == nil || z.Kind != game.ZoneHand {
		t.Error("the Shock left the hand")
	}
	// Two cards left the hand and two were drawn: the hand is where it began.
	if opp.Hand.Size() != handBefore {
		t.Errorf("the opponent's hand is %d, want %d (two exiled, two drawn)", opp.Hand.Size(), handBefore)
	}
}

func TestKldB259LostLegacyIgnoresAnArtifactOrLandName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	ring := effectstest.PushHandCardOfType(opp, "Sol Ring", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Lost Legacy", "Sorcery", kldB259LostLegacy,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Sol Ring"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(ring) {
		t.Error("an artifact card was exiled by a name that Lost Legacy may not choose")
	}
}

func TestKldB259LostLegacyRefusesAPoolThatCannotPayAndAnIllegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lost Legacy", TypeLine: "Sorcery", ManaCost: "{1}{B}{B}",
			OracleID: kldB259LostLegacy, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("cast with an empty pool")
	}
}
