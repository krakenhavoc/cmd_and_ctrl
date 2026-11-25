package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b210_cards_test.go — Dominaria cards, S82 OLD batch 210 (#5530),
// played through the real catalog entries.

const (
	oldB210NiambiOracle      = "0c107664-a9fc-4684-8a1b-fa750401a6dd"
	oldB210UrgorosOracle     = "43e7b834-d3a9-4297-bf79-2dab5507bebe"
	oldB210WeightOracle      = "826c5feb-9e82-49e0-bc6f-6c29004f8151"
	oldB210OutburstOracle    = "894a344e-a8a7-4ddc-833d-eea41f1918d7"
	oldB210PrecognitionOracl = "9c3e4bd4-2e1e-4103-adc3-f9ccc259dcdf"
	oldB210MachinationsOracl = "f0adab66-c3fe-4989-b7c8-ab7bdc7f0fd1"
)

func TestOldB210CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB210NiambiOracle:      "Niambi, Faithful Healer",
		oldB210UrgorosOracle:     "Urgoros, the Empty One",
		oldB210WeightOracle:      "Weight of Memory",
		oldB210OutburstOracle:    "Chandra's Outburst",
		oldB210PrecognitionOracl: "Precognition Field",
		oldB210MachinationsOracl: "Temporal Machinations",
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

func oldB210Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB210NiambiFetchesTeferiTimebenderFromLibraryOrGraveyard(t *testing.T) {
	t.Run("library", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB210Seats(g)
		ids := effectstest.SeedSearchLibrary(me,
			game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
			game.Card{Name: "Teferi, Timebender", TypeLine: "Legendary Planeswalker — Teferi"})
		effectstest.LgKnowLibrary(me)
		effectstest.CastCatalogSpell(t, g, "Niambi, Faithful Healer", "Legendary Creature — Human Cleric", oldB210NiambiOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, true)
		effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
		effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(ids[1]) {
			t.Fatal("Teferi, Timebender was not fetched from the library")
		}
	})
	t.Run("graveyard", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB210Seats(g)
		effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		effectstest.LgKnowLibrary(me)
		pw := effectstest.PushGraveyardCardWithTypeLineID(me, "Teferi, Timebender", "Legendary Planeswalker — Teferi")
		effectstest.CastCatalogSpell(t, g, "Niambi, Faithful Healer", "Legendary Creature — Human Cleric", oldB210NiambiOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, true)
		effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
		effectstest.AnswerSearchByID(t, g, me.ID, pw)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(pw) {
			t.Fatal("Teferi, Timebender was not fetched from the graveyard")
		}
	})
}

func oldB210Urgoros(t *testing.T) (g *game.Game, me, opp *game.Player, urgoros uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = oldB210Seats(g)
	urgoros = effectstest.B12Push(g, me.ID, "Urgoros, the Empty One", "Legendary Creature — Specter", oldB210UrgorosOracle, 4, 3)
	effectstest.AssertKeywords(t, g, urgoros, "flying")
	return g, me, opp, urgoros
}

func TestOldB210UrgorosMakesTheDamagedPlayerDiscardAtRandom(t *testing.T) {
	g, me, opp, urgoros := oldB210Urgoros(t)
	effectstest.PushHandCard(g, opp)
	effectstest.PushHandCard(g, opp)
	myHand, theirHand, theirGrave := me.Hand.Size(), opp.Hand.Size(), opp.Graveyard.Size()
	effectstest.DealCombatDamageToPlayer(g, urgoros, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != theirHand-1 {
		t.Fatalf("opponent hand = %d, want %d after the random discard", got, theirHand-1)
	}
	if opp.Graveyard.Size() != theirGrave+1 {
		t.Fatalf("opponent graveyard = %d, want %d", opp.Graveyard.Size(), theirGrave+1)
	}
	if me.Hand.Size() != myHand {
		t.Fatalf("Urgoros's controller drew although the discard happened (hand %d -> %d)", myHand, me.Hand.Size())
	}
}

func TestOldB210UrgorosDrawsWhenTheDamagedPlayerCannotDiscard(t *testing.T) {
	g, me, opp, urgoros := oldB210Urgoros(t)
	for opp.Hand.Size() > 0 {
		opp.Hand.Cards = opp.Hand.Cards[1:]
	}
	myHand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, urgoros, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != myHand+1 {
		t.Fatalf("Urgoros's controller hand = %d, want %d (a draw for the empty hand)", got, myHand+1)
	}
}

func TestOldB210UrgorosIgnoresNoncombatDamage(t *testing.T) {
	g, me, opp, urgoros := oldB210Urgoros(t)
	effectstest.PushHandCard(g, opp)
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	effectstest.EmitNoncombatDamage(g, urgoros, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != theirHand || me.Hand.Size() != myHand {
		t.Fatalf("noncombat damage triggered Urgoros (opp hand %d -> %d, my hand %d -> %d)", theirHand, opp.Hand.Size(), myHand, me.Hand.Size())
	}
}

func TestOldB210WeightOfMemoryDrawsThenMillsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB210Seats(g)
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryTopForTest(me, "Mine")
		effectstest.PushLibraryTopForTest(opp, "Theirs")
	}
	myHand, myLib, theirLib, theirGrave := me.Hand.Size(), me.Library.Size(), opp.Library.Size(), opp.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Weight of Memory", "Sorcery", oldB210WeightOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	// The spell card left the hand when it was cast, so the net change is +3 draws.
	if got := me.Hand.Size() - myHand; got != 3 {
		t.Errorf("hand grew by %d, want 3 (the spell itself is not counted)", got)
	}
	if got := myLib - me.Library.Size(); got != 3 {
		t.Errorf("caster library fell by %d, want 3", got)
	}
	if got := theirLib - opp.Library.Size(); got != 3 {
		t.Errorf("target library fell by %d, want 3", got)
	}
	if got := opp.Graveyard.Size() - theirGrave; got != 3 {
		t.Errorf("target graveyard grew by %d, want 3", got)
	}
}

func TestOldB210WeightOfMemoryMayTargetItsCaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB210Seats(g)
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryTopForTest(me, "Mine")
	}
	lib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Weight of Memory", "Sorcery", oldB210WeightOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - me.Library.Size(); got != 6 {
		t.Errorf("library fell by %d, want 6 (three drawn, three milled)", got)
	}
}

func TestOldB210ChandrasOutburstDamagesAPlayerAndFetchesChandra(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB210Seats(g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Chandra, Bold Pyromancer", TypeLine: "Legendary Planeswalker — Chandra"})
	effectstest.LgKnowLibrary(me)
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Chandra's Outburst", "Sorcery", oldB210OutburstOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-4 {
		t.Errorf("opponent life = %d, want %d", got, before-4)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("Chandra, Bold Pyromancer was not fetched from the library")
	}
}

func TestOldB210ChandrasOutburstFetchesChandraFromTheGraveyardAndHitsAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB210Seats(g)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	chandra := effectstest.PushGraveyardCardWithTypeLineID(me, "Chandra, Bold Pyromancer", "Legendary Planeswalker — Chandra")
	pw := effectstest.PushCatalogPermanent(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "", false)
	effectstest.SetCounters(g, pw, map[string]int{game.CounterLoyalty: 6})
	effectstest.CastCatalogSpell(t, g, "Chandra's Outburst", "Sorcery", oldB210OutburstOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: pw}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchByID(t, g, me.ID, chandra)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, pw, game.CounterLoyalty); got != 2 {
		t.Errorf("planeswalker loyalty = %d, want 2", got)
	}
	if !me.Hand.Contains(chandra) {
		t.Fatal("Chandra, Bold Pyromancer was not fetched from the graveyard")
	}
}

func TestOldB210ChandrasOutburstRefusesACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB210Seats(g)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Chandra's Outburst", "Sorcery", oldB210OutburstOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Fatal("Chandra's Outburst targeted a creature")
	}
}

func TestOldB210PrecognitionFieldCastsOnlyInstantsAndSorceriesFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB210Seats(g)
	effectstest.ToMain(t, g)
	effectstest.PushTopSource(t, g, me.ID, "Precognition Field", "Enchantment", oldB210PrecognitionOracl)
	if !effectstest.TopVisibleTo(g, me.ID, me.ID) || effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Error("the top card must be visible to its owner only")
	}
	for _, tc := range []struct {
		name string
		card game.Card
		want bool
	}{
		{"sorcery", game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"}, true},
		{"instant", game.Card{Name: "Top Instant", TypeLine: "Instant", ManaCost: "{0}"}, true},
		{"creature", game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"}, false},
		{"artifact", game.Card{Name: "Top Rock", TypeLine: "Artifact", ManaCost: "{0}"}, false},
		{"land", game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			top := effectstest.SeedTopOfLibrary(me, tc.card)
			if got := effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted(); got != tc.want {
				t.Fatalf("permission granted = %v, want %v", got, tc.want)
			}
			if !tc.want {
				if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
					t.Errorf("a card the Field does not open was cast off the top")
				}
			}
		})
	}
}

func TestOldB210PrecognitionFieldCastsASorceryFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB210Seats(g)
	effectstest.ToMain(t, g)
	effectstest.PushTopSource(t, g, me.ID, "Precognition Field", "Enchantment", oldB210PrecognitionOracl)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("cast the sorcery off the top: %v", err)
	}
	if me.Library.Contains(top) {
		t.Error("the cast card is still in the library")
	}
}

func TestOldB210PrecognitionFieldExilesTheTopCardForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB210Seats(g)
	effectstest.ToMain(t, g)
	field := effectstest.PushCatalogPermanent(g, me.ID, "Precognition Field", "Enchantment", oldB210PrecognitionOracl, false)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"})

	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Library.Contains(top) {
		t.Error("the card left the library before the ability resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Contains(top) || !effectstest.ExileHas(g, top) {
		t.Error("the top card was not exiled")
	}
}

func TestOldB210TemporalMachinationsBouncesAndDrawsOnlyWithAnArtifact(t *testing.T) {
	for _, withArtifact := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB210Seats(g)
		for i := 0; i < 3; i++ {
			effectstest.PushLibraryTopForTest(me, "Mine")
		}
		bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
		if withArtifact {
			effectstest.PushArtifactForTest(g, me.ID, "Rock")
		}
		hand, theirHand := me.Hand.Size(), opp.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Temporal Machinations", "Sorcery", oldB210MachinationsOracl,
			[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.BattlefieldHas(g, bear) || opp.Hand.Size() != theirHand+1 {
			t.Errorf("artifact=%v: the creature was not returned to its owner's hand", withArtifact)
		}
		want := hand
		if withArtifact {
			want = hand + 1
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("artifact=%v: hand = %d, want %d", withArtifact, got, want)
		}
	}
}
