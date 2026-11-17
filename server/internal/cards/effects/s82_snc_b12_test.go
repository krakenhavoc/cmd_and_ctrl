package effects

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b12_test.go — Streets of New Capenna, S82 batch 12: the cards that
// count mana values among cards in your graveyard.

const (
	sncSanguineSpyOracle   = "03af4b10-900e-4013-bd5d-dd8d132d50fb"
	sncTaintedOracle       = "6944092f-22c6-4b65-840d-8c1b70883127"
	sncGraveyardShiftOracl = "da8a93c6-66e2-492d-8894-397ba9ec5a87"
	sncSewerCrocOracle     = "b01e9ff2-865a-401e-a839-568cd4a0451d"
	sncHeartstabberOracle  = "fee73267-b0ef-4f43-a0bd-7728691d6f17"
	sncNewsieOracle        = "347a5faa-5656-4a36-af44-ab5a3ed267f4"
	sncInfiltratorOracle   = "0d6be8a4-b05b-40fb-a653-5aabf71809e0"
	sncArbiterOracle       = "f15460d9-0133-436c-b0fe-1661989124f0"
	sncOskarOracle         = "14c2c818-0ce2-4809-a5ff-ee4a6253defa"
	sncMaestrosOracle      = "85905607-dceb-4c37-81a3-ab441ce9490c"
)

// sncGrave seeds one nonland card per mana value into p's graveyard (a mana
// value of 0 is a land card), and returns their ids.
func sncGrave(p *game.Player, mvs ...int) []uuid.UUID {
	var ids []uuid.UUID
	for i, mv := range mvs {
		c := game.Card{InstanceID: uuid.New(), Name: fmt.Sprintf("Grave %d-%d", mv, i), Owner: p.ID, Controller: p.ID}
		if mv == 0 {
			c.TypeLine = "Land"
		} else {
			c.TypeLine, c.ManaCost = "Sorcery", fmt.Sprintf("{%d}", mv)
		}
		p.Graveyard.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func sncTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	return g, g.Seats[0], g.Seats[1]
}

func TestSncB12EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncSanguineSpyOracle: "Sanguine Spy", sncTaintedOracle: "Tainted Indulgence",
		sncGraveyardShiftOracl: "Graveyard Shift", sncSewerCrocOracle: "Sewer Crocodile",
		sncHeartstabberOracle: "Aven Heartstabber", sncNewsieOracle: "Snooping Newsie",
		sncInfiltratorOracle: "Syndicate Infiltrator", sncArbiterOracle: "All-Seeing Arbiter",
		sncOskarOracle: "Oskar, Rubbish Reclaimer", sncMaestrosOracle: "Maestros Initiate",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// The threshold counts DIFFERENT mana values: duplicates add nothing, a land is 0.
func TestSncB12SyndicateInfiltratorNeedsFiveDifferentManaValues(t *testing.T) {
	g, me, _ := sncTable(t)
	spy := blbB1Push(g, me.ID, "Syndicate Infiltrator", "Creature — Vampire Wizard", sncInfiltratorOracle, 3, 3)
	sncGrave(me, 1, 1, 2, 2, 3, 3, 4, 4) // four different values, eight cards
	if p, tt := ptOf(t, g, spy); p != 3 || tt != 3 {
		t.Fatalf("four mana values: %d/%d, want 3/3", p, tt)
	}
	extra := sncGrave(me, 0) // a land is mana value 0: the fifth value
	if p, tt := ptOf(t, g, spy); p != 5 || tt != 5 {
		t.Fatalf("five mana values: %d/%d, want 5/5", p, tt)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile}, extra[0]); err != nil {
		t.Fatal(err)
	}
	if p, _ := ptOf(t, g, spy); p != 3 {
		t.Fatalf("the fifth value left the graveyard: power %d, want 3", p)
	}
}

// Only the controller's own graveyard counts.
func TestSncB12InfiltratorIgnoresAnOpponentsGraveyard(t *testing.T) {
	g, me, opp := sncTable(t)
	spy := blbB1Push(g, me.ID, "Syndicate Infiltrator", "Creature — Vampire Wizard", sncInfiltratorOracle, 3, 3)
	sncGrave(opp, 0, 1, 2, 3, 4, 5)
	if p, _ := ptOf(t, g, spy); p != 3 {
		t.Fatalf("an opponent's graveyard counted: power %d", p)
	}
}

func TestSncB12SnoopingNewsieMillsTwoThenTurnsOnLifelinkAndPump(t *testing.T) {
	g, me, _ := sncTable(t)
	sncGrave(me, 1, 2, 3)
	me.Library.Cards = nil
	pushLibraryCardForTest(me, game.Card{Name: "Deep", TypeLine: "Sorcery", ManaCost: "{6}"})
	pushLibraryCardForTest(me, game.Card{Name: "Mill B", TypeLine: "Sorcery", ManaCost: "{5}"})
	pushLibraryCardForTest(me, game.Card{Name: "Mill A", TypeLine: "Sorcery", ManaCost: "{4}"})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Snooping Newsie", TypeLine: "Creature — Human Rogue",
		OracleID: sncNewsieOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g) // the spell
	if p, _ := ptOf(t, g, id); p != 2 || hasKeywordOnBattlefield(t, g, id, "lifelink") {
		t.Fatal("pumped before the entry trigger resolved")
	}
	passPriorityAroundTable(t, g) // the entry trigger
	if got := me.Graveyard.Size(); got != 5 {
		t.Fatalf("graveyard = %d cards, want 5 (two milled)", got)
	}
	if p, tt := ptOf(t, g, id); p != 3 || tt != 3 {
		t.Errorf("five mana values: %d/%d, want 3/3", p, tt)
	}
	if !hasKeywordOnBattlefield(t, g, id, "lifelink") {
		t.Error("five mana values: no lifelink")
	}
}

func TestSncB12AvenHeartstabberPumpsDeathtouchAndMillsThenDrawsWhenItDies(t *testing.T) {
	g, me, _ := sncTable(t)
	me.Library.Cards = nil
	pushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Sorcery", ManaCost: "{1}"})
	pushLibraryCardForTest(me, game.Card{Name: "Mill B", TypeLine: "Sorcery", ManaCost: "{2}"})
	pushLibraryCardForTest(me, game.Card{Name: "Mill A", TypeLine: "Sorcery", ManaCost: "{3}"})
	bird := blbB1Push(g, me.ID, "Aven Heartstabber", "Creature — Bird Assassin", sncHeartstabberOracle, 1, 1)
	sncGrave(me, 1, 2, 3, 4)
	if p, tt := ptOf(t, g, bird); p != 1 || tt != 1 || hasKeywordOnBattlefield(t, g, bird, "deathtouch") {
		t.Fatalf("four values: %d/%d", p, tt)
	}
	sncGrave(me, 5)
	if p, tt := ptOf(t, g, bird); p != 3 || tt != 3 || !hasKeywordOnBattlefield(t, g, bird, "deathtouch") {
		t.Fatalf("five values: %d/%d, want 3/3 with deathtouch", p, tt)
	}
	hand := me.Hand.Size()
	killCreature(t, g, me.ID, bird)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand = %d, want %d", got, hand+1)
	}
	for _, c := range me.Hand.Cards {
		if c.Name == "Drawn" {
			return
		}
	}
	t.Error("the drawn card was not the one below the two milled cards")
}

func TestSncB12TaintedIndulgenceDiscardsOnlyBelowFiveManaValues(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mvs     []int
		discard bool
	}{{"four values", []int{1, 2, 3, 4}, true}, {"five values", []int{1, 2, 3, 4, 5}, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := sncTable(t)
			sncGrave(me, tc.mvs...)
			me.Hand.Cards = nil
			for i := 0; i < 3; i++ {
				pushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Sorcery", ManaCost: "{9}"})
			}
			castCatalogSpell(t, g, "Tainted Indulgence", "Instant", sncTaintedOracle, nil)
			passPriorityAroundTable(t, g)
			prompt := discardChoiceFor(g, me.ID)
			if tc.discard {
				if prompt == nil {
					t.Fatal("no discard prompt below five mana values")
				}
				discardFromHand(t, g, me.ID)
				if got := me.Hand.Size(); got != 1 {
					t.Errorf("hand = %d, want 1", got)
				}
				return
			}
			if prompt != nil {
				t.Fatal("asked to discard with five mana values")
			}
			if got := me.Hand.Size(); got != 2 {
				t.Errorf("hand = %d, want 2", got)
			}
		})
	}
}

func TestSncB12GraveyardShiftHasFlashOnlyWithFiveManaValues(t *testing.T) {
	g, me, _ := sncTable(t)
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	sncGrave(me, 3, 4, 5)
	shift := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: shift, Name: "Graveyard Shift", TypeLine: "Sorcery",
		OracleID: sncGraveyardShiftOracl, Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, shift, game.CastSpellParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Fatal("cast at instant speed with four mana values (bear 2, 3, 4, 5)")
	}
	sncGrave(me, 6) // five: 2, 3, 4, 5, 6
	if err := g.CastSpell(me.ID, shift, game.CastSpellParams{Strict: true, Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("flash with five mana values: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature card did not return")
	}
}

func TestSncB12GraveyardShiftIsASorceryWithoutFiveAndReturnsAtSorcerySpeed(t *testing.T) {
	g, me, _ := sncTable(t)
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	shift := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: shift, Name: "Graveyard Shift", TypeLine: "Sorcery",
		OracleID: sncGraveyardShiftOracl, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, shift, game.CastSpellParams{Strict: true, Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("main phase cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature card did not return")
	}
}

func TestSncB12SewerCrocodileCostsThreeLessWithFiveManaValues(t *testing.T) {
	for _, tc := range []struct {
		name string
		mvs  []int
		pool string
		ok   bool
	}{
		{"four values pays all four", []int{1, 2, 3, 4}, "{C}{C}{C}{U}", true},
		{"four values cannot pay {U} alone", []int{1, 2, 3, 4}, "{U}", false},
		{"five values pays {U}", []int{1, 2, 3, 4, 5}, "{U}", true},
		{"five values still needs the blue", []int{1, 2, 3, 4, 5}, "{C}{C}{C}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := sncTable(t)
			croc := blbB1Push(g, me.ID, "Sewer Crocodile", "Creature — Crocodile", sncSewerCrocOracle, 4, 6)
			sncGrave(me, tc.mvs...)
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.pool); err != nil {
				t.Fatal(err)
			}
			err := g.ActivateCatalogAbility(me.ID, croc, 0, game.ActivateAbilityParams{})
			if (err == nil) != tc.ok {
				t.Fatalf("activation err = %v, want ok=%v", err, tc.ok)
			}
			if !tc.ok {
				return
			}
			passPriorityAroundTable(t, g)
			if restrictionsOf(t, g, croc)&game.CantBeBlocked == 0 {
				t.Error("not unblockable after the ability resolved")
			}
		})
	}
}

func TestSncB12SanguineSpySacrificesAnotherCreatureToSurveil(t *testing.T) {
	g, me, _ := sncTable(t)
	spy := blbB1Push(g, me.ID, "Sanguine Spy", "Creature — Vampire Rogue", sncSanguineSpyOracle, 2, 3)
	fodder := blbB1Push(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	pushLibraryCardForTest(me, game.Card{Name: "Top", TypeLine: "Sorcery", ManaCost: "{1}"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, spy, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{spy}}); err == nil {
		t.Fatal("the Spy may not sacrifice itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, spy, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) == nil && len(g.PendingChoices) == 0 {
		t.Error("surveil 1 asked nothing")
	}
}

func TestSncB12SanguineSpyEndStepPaysLifeForACardOnlyWithFive(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mvs          []int
		answerYes    bool
		wantLifeLoss int
		wantDraw     int
		asked        bool
	}{
		{"four values: no trigger", []int{1, 2, 3, 4}, true, 0, 0, false},
		{"five values, pay", []int{1, 2, 3, 4, 5}, true, 2, 1, true},
		{"five values, decline", []int{1, 2, 3, 4, 5}, false, 0, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := sncTable(t)
			blbB1Push(g, me.ID, "Sanguine Spy", "Creature — Vampire Rogue", sncSanguineSpyOracle, 2, 3)
			sncGrave(me, tc.mvs...)
			pushLibraryCardForTest(me, game.Card{Name: "Top", TypeLine: "Sorcery", ManaCost: "{1}"})
			life, hand := me.Life, me.Hand.Size()
			advanceToEndStep(t, g)
			passPriorityAroundTable(t, g)
			asked := latestConfirmFor(g, me.ID) != nil
			if asked != tc.asked {
				t.Fatalf("asked = %v, want %v", asked, tc.asked)
			}
			if asked {
				answerMayChoice(t, g, me.ID, tc.answerYes)
				passPriorityAroundTable(t, g)
			}
			if got := life - me.Life; got != tc.wantLifeLoss {
				t.Errorf("life lost = %d, want %d", got, tc.wantLifeLoss)
			}
			if got := me.Hand.Size() - hand; got != tc.wantDraw {
				t.Errorf("cards drawn = %d, want %d", got, tc.wantDraw)
			}
		})
	}
}

func TestSncB12MaestrosInitiateLootsFromTheGraveyardForEitherColour(t *testing.T) {
	for _, pool := range []string{"{C}{C}{C}{C}{U}", "{C}{C}{C}{C}{R}"} {
		t.Run(pool, func(t *testing.T) {
			g, me, _ := sncTable(t)
			id := uuid.New()
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Maestros Initiate", TypeLine: "Creature — Human Citizen",
				ManaCost: "{2}{B}", OracleID: sncMaestrosOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
			me.Hand.Cards = nil
			for i := 0; i < 3; i++ {
				pushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Sorcery", ManaCost: "{9}"})
			}
			if err := g.AddManaForEffect(me.ID, uuid.Nil, pool); err != nil {
				t.Fatal(err)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate from the graveyard: %v", err)
			}
			if me.Graveyard.Contains(id) || !g.Exile.Contains(id) {
				t.Error("the card was not exiled as the cost")
			}
			passPriorityAroundTable(t, g)
			discardFromHand(t, g, me.ID)
			if got := me.Hand.Size(); got != 1 {
				t.Errorf("hand = %d, want 1 (drew two, discarded one)", got)
			}
		})
	}
}

func TestSncB12MaestrosInitiateCannotBeActivatedFromTheHand(t *testing.T) {
	g, me, _ := sncTable(t)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Maestros Initiate", TypeLine: "Creature — Human Citizen",
		OracleID: sncMaestrosOracle, Owner: me.ID, Controller: me.ID})
	_ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the ability works only from the graveyard")
	}
}

func TestSncB12AllSeeingArbiterShrinksAnOpponentsCreatureOnEachDiscard(t *testing.T) {
	g, me, opp := sncTable(t)
	big := blbB1Push(g, opp.ID, "Big", "Creature — Ogre", "", 6, 6)
	sncGrave(me, 1, 2, 3)
	me.Hand.Cards = nil
	me.Hand.PushTop(game.Card{Name: "Pitch", TypeLine: "Sorcery", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Sorcery", ManaCost: "{9}"})
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "All-Seeing Arbiter", TypeLine: "Creature — Avatar",
		OracleID: sncArbiterOracle, Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g) // the spell
	passPriorityAroundTable(t, g) // enters: draw two, then discard
	var pitched uuid.UUID
	for _, c := range me.Hand.Cards {
		if c.Name == "Pitch" {
			pitched = c.InstanceID
		}
	}
	answerDiscard(t, g, me.ID, pitched)
	passPriorityAroundTable(t, g) // the discard trigger
	if p, tt := ptOf(t, g, big); p != 2 || tt != 6 {
		t.Fatalf("after discarding a mana value 4 card: %d/%d, want 2/6 (X = 4)", p, tt)
	}
	// Until YOUR next turn: still shrunk through the other seats' turns.
	advanceToUpkeepOf(t, g, 1)
	if p, _ := ptOf(t, g, big); p != 2 {
		t.Errorf("on the opponent's turn power = %d, want 2", p)
	}
	advanceToUpkeepOf(t, g, 0)
	if p, _ := ptOf(t, g, big); p != 6 {
		t.Errorf("on my next turn power = %d, want 6", p)
	}
}

func TestSncB12OskarCostsOneLessPerDifferentManaValue(t *testing.T) {
	for _, tc := range []struct {
		name string
		mvs  []int
		pool string
		ok   bool
	}{
		{"no cards", nil, "{C}{C}{C}{U}{B}", true},
		{"three different values", []int{1, 2, 3, 3, 3}, "{U}{B}", true},
		{"duplicates count once", []int{2, 2, 2}, "{C}{C}{U}{B}", true},
		{"duplicates are not enough", []int{2, 2, 2}, "{C}{U}{B}", false},
		{"never below the coloured part", []int{1, 2, 3, 4, 5, 6}, "{U}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := sncTable(t)
			sncGrave(me, tc.mvs...)
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Oskar, Rubbish Reclaimer", TypeLine: "Legendary Creature — Human Wizard",
				ManaCost: "{3}{U}{B}", OracleID: sncOskarOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.pool); err != nil {
				t.Fatal(err)
			}
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
			if (err == nil) != tc.ok {
				t.Fatalf("cast err = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}

func TestSncB12OskarOffersToCastADiscardedNonlandCard(t *testing.T) {
	g, me, _ := sncTable(t)
	blbB1Push(g, me.ID, "Oskar, Rubbish Reclaimer", "Legendary Creature — Human Wizard", sncOskarOracle, 3, 3)
	me.Hand.Cards = nil
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, bolt)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast the discarded card from the graveyard: %v", err)
	}
	castInsideTheResolution(t, g, bolt)
}

func TestSncB12OskarDeclinedOfferLeavesTheCardAndLandsAreNotOffered(t *testing.T) {
	g, me, _ := sncTable(t)
	blbB1Push(g, me.ID, "Oskar, Rubbish Reclaimer", "Legendary Creature — Human Wizard", sncOskarOracle, 3, 3)
	me.Hand.Cards = nil
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a discarded land was offered for casting")
	}
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	_ = g.AddManaForEffect(me.ID, uuid.Nil, "{R}")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	declineResolutionCast(t, g, me.ID)
	if !me.Graveyard.Contains(bolt) {
		t.Fatal("the declined card left the graveyard")
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("the offer closed on the decline")
	}
}
