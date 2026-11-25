package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b154_test.go — Core Set 2020, S82 set OLD batch 154 (#5518). Every
// test plays the real catalog entry through the engine.

const (
	oldB154Cavalier   = "98b19afc-7696-46b9-96f0-7f7c60003e1c"
	oldB154Miscreant  = "dd2510cd-356c-4d5a-8e5d-e44872eba60b"
	oldB154Golos      = "efe9c3e0-7315-408e-9903-9d25a681039b"
	oldB154FlameWave  = "9f84fc91-8647-4cc8-9b57-d23a83f85ab1"
	oldB154Regulator  = "d8c95a86-2f05-44f9-8ed7-b1798eaed42f"
	oldB154Matcher    = "d5b33c1a-6e15-4a08-b834-38b0465b0cfb"
	oldB154Atemsis    = "f557aa26-8b79-4116-9249-a067b21b88ef"
	oldB154Lockbox    = "e00810cd-8863-433b-aa7a-683b8f308a9b"
	oldB154ChandraFF  = "408966d2-08c0-4531-a3b4-e666840ee95c"
	oldB154ChandraFFn = "Chandra, Flame's Fury"
)

func TestOldB154CardsAreRegisteredWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB154Cavalier: "Cavalier of Flame", oldB154Miscreant: "Faerie Miscreant",
		oldB154Golos: "Golos, Tireless Pilgrim", oldB154FlameWave: "Chandra's Flame Wave",
		oldB154Regulator: "Chandra's Regulator", oldB154Matcher: "Pattern Matcher",
		oldB154Atemsis: "Atemsis, All-Seeing", oldB154Lockbox: "Diviner's Lockbox",
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

// oldB154OnTop puts a card on top of the player's library and returns its ID.
func oldB154OnTop(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
	g.WithWriteLock(func() { p.Library.PushTop(c) })
	return c.InstanceID
}

// oldB154Active is the active seat and the next opponent.
func oldB154Active(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// --- Cavalier of Flame --------------------------------------------------------

func TestOldB154CavalierPumpsAndHastesYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB154Active(g)
	cav := effectstest.PushCatalogPermanent(g, me.ID, "Cavalier of Flame", "Creature — Elemental Knight", oldB154Cavalier, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, cav, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {1}{R} for one mana")
	}
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, cav, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.CurrentPower(t, g, bear); p != 3 {
		t.Errorf("my Bear has power %d, want 3", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("my Bear has no haste")
	}
	if p := effectstest.CurrentPower(t, g, theirs); p != 2 {
		t.Errorf("their creature has power %d, want 2", p)
	}
}

func TestOldB154CavalierDiscardsAnyNumberThenDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB154Active(g)
	effectstest.Msc18Library(me, 6)
	hand := effectstest.S58HandCards(me, 3)
	effectstest.CastFromHandForTest(t, g, me, "Cavalier of Flame", "Creature — Elemental Knight", "{2}{R}{R}{R}", oldB154Cavalier,
		game.CastSpellParams{Strict: false})
	effectstest.PassPriorityAroundTable(t, g)
	before := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, hand[0], hand[1])
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d → %d, want the same size (discarded two, drew two)", before, got)
	}
	if got := me.Graveyard.Size(); got != 2 {
		t.Errorf("%d cards in the graveyard, want 2", got)
	}
}

func TestOldB154CavalierDiesForDamagePerLandInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB154Active(g)
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	cav := effectstest.PushCatalogPermanent(g, me.ID, "Cavalier of Flame", "Creature — Elemental Knight", oldB154Cavalier, false)
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		})
	}
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	walker := effectstest.S82Walker(g, opp.ID, "Test Walker", "", 5)
	myWalker := effectstest.S82Walker(g, me.ID, "My Walker", "", 5)
	oppLife, otherLife, myLife := opp.Life, other.Life, me.Life
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, cav).DamageMarked = 10 })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || other.Life != otherLife-2 {
		t.Errorf("opponents' life %d and %d, want %d and %d", opp.Life, other.Life, oppLife-2, otherLife-2)
	}
	if me.Life != myLife {
		t.Errorf("my life changed to %d", me.Life)
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 3 {
		t.Errorf("their planeswalker has loyalty %d, want 3", got)
	}
	if got := effectstest.LoyaltyOf(g, myWalker); got != 5 {
		t.Errorf("my planeswalker has loyalty %d, want 5", got)
	}
}

// --- Faerie Miscreant ---------------------------------------------------------

func TestOldB154MiscreantDrawsOnlyWithAnotherMiscreant(t *testing.T) {
	for _, tc := range []struct {
		name    string
		another string
		want    int
	}{
		{"alone", "", 0},
		{"another with the same name", "Faerie Miscreant", 1},
		{"another with a different name", "Faerie Rogue", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, _ := oldB154Active(g)
			effectstest.Msc18Library(me, 4)
			if tc.another != "" {
				effectstest.B12Creature(g, me.ID, tc.another, "Creature — Faerie Rogue", 1, 1)
			}
			hand := me.Hand.Size()
			effectstest.CastFromHandForTest(t, g, me, "Faerie Miscreant", "Creature — Faerie Rogue", "{U}", oldB154Miscreant, game.CastSpellParams{})
			effectstest.PassPriorityAroundTable(t, g)
			// The cast took the Miscreant out of the hand it was put in.
			if got := me.Hand.Size() - hand; got != tc.want {
				t.Errorf("hand grew by %d, want %d", got, tc.want)
			}
		})
	}
}

// --- Golos, Tireless Pilgrim --------------------------------------------------

func TestOldB154GolosFetchesATappedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB154Active(g)
	effectstest.Msc18Library(me, 4)
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Wooded Test Land", TypeLine: "Land"})
	effectstest.CastFromHandForTest(t, g, me, "Golos, Tireless Pilgrim", "Legendary Artifact Creature — Scout", "{5}", oldB154Golos, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Wooded Test Land")
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, forest)
	if !ok || !c.Tapped {
		t.Errorf("land on the battlefield %v, tapped %v, want both", ok, c.Tapped)
	}
}

func TestOldB154GolosExilesThreeAndLetsYouCastThemFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB154Active(g)
	golos := effectstest.PushCatalogPermanent(g, me.ID, "Golos, Tireless Pilgrim", "Legendary Artifact Creature — Scout", oldB154Golos, false)
	effectstest.Msc18Library(me, 6)
	big := oldB154OnTop(g, me, game.Card{Name: "Big Creature", TypeLine: "Creature — Giant", ManaCost: "{6}{G}{G}", Power: 8, Toughness: 8})
	one := oldB154OnTop(g, me, game.Card{Name: "Second", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})
	two := oldB154OnTop(g, me, game.Card{Name: "Third", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})
	libBefore := me.Library.Size()
	effectstest.FloatForTest(g, me, "CCWUBRG")
	if err := g.ActivateCatalogAbility(me.ID, golos, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != libBefore-3 {
		t.Fatalf("library %d → %d, want three cards exiled", libBefore, got)
	}
	exiled := 0
	for _, id := range []uuid.UUID{big, one, two} {
		if effectstest.InExile(g, id) {
			exiled++
		}
	}
	if exiled != 3 {
		t.Fatalf("%d of the three top cards are in exile, want 3", exiled)
	}
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("casting the exiled {6}{G}{G} creature for free: %v", err)
	}
}

// --- Chandra's Flame Wave -----------------------------------------------------

func TestOldB154FlameWaveHitsThePlayerAndTheirCreaturesThenFindsChandra(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB154Active(g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	chandra := effectstest.PushLibraryCardForTest(me, game.Card{Name: oldB154ChandraFFn, TypeLine: "Legendary Planeswalker — Chandra", OracleID: oldB154ChandraFF})
	oppLife := opp.Life
	effectstest.FloatForTest(g, me, "CCCRR")
	effectstest.CastFromHandForTest(t, g, me, "Chandra's Flame Wave", "Sorcery", "{3}{R}{R}", oldB154FlameWave,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary, game.SearchZoneGraveyard)
	effectstest.AnswerSearchNamed(t, g, me.ID, oldB154ChandraFFn)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("their life %d, want %d", opp.Life, oppLife-2)
	}
	if _, ok := effectstest.BattlefieldCard(g, small); ok {
		t.Error("their 2/2 survived")
	}
	if c, ok := effectstest.BattlefieldCard(g, big); !ok || c.DamageMarked != 2 {
		t.Errorf("their 4/4: present %v, damage %d, want 2", ok, c.DamageMarked)
	}
	if c, ok := effectstest.BattlefieldCard(g, mine); !ok || c.DamageMarked != 0 {
		t.Errorf("my creature: present %v, damage %d, want untouched", ok, c.DamageMarked)
	}
	if !effectstest.InHand(me, chandra) {
		t.Error("Chandra, Flame's Fury is not in my hand")
	}
}

// --- Chandra's Regulator ------------------------------------------------------

// oldB154ChandraWalker puts a Chandra planeswalker with its catalog abilities
// on the battlefield.
func oldB154ChandraWalker(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: oldB154ChandraFFn, TypeLine: "Legendary Planeswalker — Chandra", OracleID: oldB154ChandraFF,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	g.RunStateChecksForTest()
	return id
}

func TestOldB154RegulatorCopiesAChandraLoyaltyAbilityForOne(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := oldB154Active(g)
		effectstest.PushCatalogPermanent(g, me.ID, "Chandra's Regulator", "Legendary Artifact", oldB154Regulator, false)
		chandra := oldB154ChandraWalker(g, me.ID)
		oppLife := opp.Life
		effectstest.S82Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.FloatForTest(g, me, "C")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := oppLife - 2
		if pay {
			want = oppLife - 4
		}
		if opp.Life != want {
			t.Errorf("pay=%v: their life %d, want %d", pay, opp.Life, want)
		}
		if got := effectstest.LoyaltyOf(g, chandra); got != 5 {
			t.Errorf("pay=%v: loyalty %d, want 5 (the copy costs no loyalty)", pay, got)
		}
	}
}

func TestOldB154RegulatorIgnoresANonChandraLoyaltyAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB154Active(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Chandra's Regulator", "Legendary Artifact", oldB154Regulator, false)
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Golgari Queen", "131096a3-2d53-4ec2-96d3-fcb58e698a13", 4)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a non-Chandra planeswalker offered the copy")
	}
}

func TestOldB154RegulatorDiscardsAMountainOrARedCardToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB154Active(g)
	effectstest.Msc18Library(me, 4)
	reg := effectstest.PushCatalogPermanent(g, me.ID, "Chandra's Regulator", "Legendary Artifact", oldB154Regulator, false)
	blue := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: blue, Name: "Blue Card", TypeLine: "Sorcery", ManaCost: "{U}", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, reg, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{blue}}); err == nil {
		t.Fatal("discarded a blue card")
	}
	mountain := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: mountain, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, reg, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{mountain}}); err != nil {
		t.Fatalf("activate with a Mountain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want the same (one discarded, one drawn)", hand, got)
	}
}

// --- Pattern Matcher ----------------------------------------------------------

func TestOldB154PatternMatcherFindsACardNamedLikeAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB154Active(g)
	effectstest.Msc18Library(me, 4)
	effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	copyID := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Pattern Matcher", TypeLine: "Artifact Creature — Golem"})
	effectstest.CastFromHandForTest(t, g, me, "Pattern Matcher", "Artifact Creature — Golem", "{4}", oldB154Matcher, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Pattern Matcher") != uuid.Nil {
		t.Error("the Matcher's own name was offered, but it is not ANOTHER creature")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Grizzly Bears")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHand(me, copyID) {
		t.Error("the Grizzly Bears card is not in my hand")
	}
}

// --- Atemsis, All-Seeing ------------------------------------------------------

func TestOldB154AtemsisLootsForTwoOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB154Active(g)
	effectstest.Msc18Library(me, 6)
	atemsis := effectstest.PushCatalogPermanent(g, me.ID, "Atemsis, All-Seeing", "Legendary Creature — Sphinx", oldB154Atemsis, false)
	effectstest.S58HandCards(me, 2)
	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, atemsis, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1 (drew two, discarded one)", hand, got)
	}
}

func TestOldB154AtemsisEndsTheGameOnlyWithSixDifferentManaValues(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values []string
		loses  bool
	}{
		{"six different", []string{"", "{1}", "{2}", "{3}", "{4}", "{5}"}, true},
		{"five different", []string{"", "{1}", "{2}", "{3}", "{4}", "{4}"}, false},
		{"seven cards, five different", []string{"", "", "{1}", "{2}", "{3}", "{4}", "{4}"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := oldB154Active(g)
			atemsis := effectstest.PushCatalogPermanent(g, me.ID, "Atemsis, All-Seeing", "Legendary Creature — Sphinx", oldB154Atemsis, false)
			g.WithWriteLock(func() {
				for _, cost := range tc.values {
					me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card " + cost, TypeLine: "Artifact", ManaCost: cost, Owner: me.ID, Controller: me.ID})
				}
			})
			effectstest.DealCombatDamageToPlayer(g, atemsis, opp.ID, 4)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Eliminated != tc.loses {
				t.Errorf("opponent eliminated = %v, want %v", opp.Eliminated, tc.loses)
			}
		})
	}
}

// --- Diviner's Lockbox --------------------------------------------------------

func TestOldB154LockboxSacrificesAndDrawsThreeOnAHit(t *testing.T) {
	for _, hit := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, _ := oldB154Active(g)
		effectstest.Msc18Library(me, 8)
		box := effectstest.PushCatalogPermanent(g, me.ID, "Diviner's Lockbox", "Artifact", oldB154Lockbox, false)
		oldB154OnTop(g, me, game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"})
		effectstest.FloatForTest(g, me, "C")
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, box, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		name := "Mox Pearl"
		if hit {
			name = "sol ring"
		}
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
		if c == nil {
			t.Fatal("no card-name prompt")
		}
		if err := g.ResolveCardNameChoice(c.ID, me.ID, name); err != nil {
			t.Fatalf("ResolveCardNameChoice: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		_, onBattlefield := effectstest.BattlefieldCard(g, box)
		drew := me.Hand.Size() - hand
		if hit && (onBattlefield || drew != 3) {
			t.Errorf("hit: lockbox still there %v, drew %d, want sacrificed and 3", onBattlefield, drew)
		}
		if !hit && (!onBattlefield || drew != 0) {
			t.Errorf("miss: lockbox there %v, drew %d, want kept and 0", onBattlefield, drew)
		}
	}
}

func TestOldB154LockboxIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB154Active(g)
	box := effectstest.PushCatalogPermanent(g, me.ID, "Diviner's Lockbox", "Artifact", oldB154Lockbox, false)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, box, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated outside a main phase")
	}
}
