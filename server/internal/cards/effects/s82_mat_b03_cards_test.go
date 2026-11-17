package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mat_b03_cards_test.go — S82 March of the Machine: The Aftermath,
// batch 3 (#5476): Sigarda, Font of Blessings; Danitha, New Benalia's
// Light; Nahiri, Forged in Fury; Kiora, Sovereign of the Deep; Kolaghan
// Warmonger; Nissa, Resurgent Animist; Pia Nalaar, Consul of Revival;
// The Kenriths' Royal Funeral; Arni Metalbrow.

const (
	s82mSigardaOracle   = "5fe07a76-f261-4f16-8a15-571bed707261"
	s82mDanithaOracle   = "971b30a8-4ab4-4e96-9d50-597da4d52994"
	s82mNahiriOracle    = "87966759-206a-4a2d-bf78-86f727c29972"
	s82mKioraOracle     = "95df54e2-0196-4e54-8144-de5982ce07f5"
	s82mWarmongerOracle = "4c7e6520-d179-4590-afdc-92eb2c0740a6"
	s82mNissaOracle     = "c1fc5923-c3cd-448a-98d1-c154661c2812"
	s82mPiaOracle       = "65d4ebdb-8757-4e5e-9411-d78597e20d57"
	s82mFuneralOracle   = "b4b0a4b7-b36f-4c96-b07b-a86c1efd333e"
	s82mArniOracle      = "d799e58a-0032-4df8-a77f-0d824ec15214"
)

func TestS82MatB03CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82mSigardaOracle:   "Sigarda, Font of Blessings",
		s82mDanithaOracle:   "Danitha, New Benalia's Light",
		s82mNahiriOracle:    "Nahiri, Forged in Fury",
		s82mKioraOracle:     "Kiora, Sovereign of the Deep",
		s82mWarmongerOracle: "Kolaghan Warmonger",
		s82mNissaOracle:     "Nissa, Resurgent Animist",
		s82mPiaOracle:       "Pia Nalaar, Consul of Revival",
		s82mFuneralOracle:   "The Kenriths' Royal Funeral",
		s82mArniOracle:      "Arni Metalbrow",
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

// --- Sigarda, Font of Blessings ---------------------------------------------

func TestS82MatSigardaGrantsHexproofToOtherPermanentsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sig := b12Push(g, me.ID, "Sigarda, Font of Blessings", "Legendary Creature — Angel", s82mSigardaOracle, 4, 4)
	land := b12Permanent(g, me.ID, "My Land", "Land")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for id, want := range map[uuid.UUID]bool{land: true, bear: true, sig: false, theirs: false} {
		if got := hasEffectiveKeyword(t, g, id, "hexproof"); got != want {
			t.Errorf("hexproof on %s = %v, want %v", b12Card(t, g, id).Name, got, want)
		}
	}
	if !hasEffectiveKeyword(t, g, sig, "flying") {
		t.Error("Sigarda lacks flying")
	}
}

func TestS82MatSigardaCastsAngelsAndHumansFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushTopSource(t, g, me.ID, "Sigarda, Font of Blessings", "Legendary Creature — Angel", s82mSigardaOracle)
	for _, tl := range []string{"Creature — Angel", "Creature — Human Wizard", "Instant — Human"} {
		id := s82mTop(me, game.Card{Name: "Top " + tl, TypeLine: tl, ManaCost: "{0}"})
		if !grantedPermissionOn(g, me.ID, id, game.ZoneLibrary).Granted() {
			t.Errorf("%q on top was not opened", tl)
		}
	}
	bear := s82mTop(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if grantedPermissionOn(g, me.ID, bear, game.ZoneLibrary).Granted() {
		t.Error("a Bear on top was opened")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("a Bear was cast off the top")
	}
	land := s82mTop(me, game.Card{Name: "Top Human Land", TypeLine: "Land — Human"})
	if grantedPermissionOn(g, me.ID, land, game.ZoneLibrary).Granted() {
		t.Error("a land on top was opened: the permission casts spells")
	}
	angel := s82mTop(me, game.Card{Name: "Cast Angel", TypeLine: "Creature — Angel", ManaCost: "{0}", Power: 1, Toughness: 1})
	if err := g.CastSpell(me.ID, angel, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("cast an Angel off the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(angel) {
		t.Error("the Angel did not resolve")
	}
}

// --- Danitha, New Benalia's Light -------------------------------------------

func s82mDanithaTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushTopSource(t, g, me.ID, "Danitha, New Benalia's Light", "Legendary Creature — Human Knight", s82mDanithaOracle)
	return g, me
}

func s82mGraveCard(me *game.Player, name, typeLine string) uuid.UUID {
	return graveyardCardOf(nil, me, name, typeLine, "{0}", "")
}

func TestS82MatDanithaCastsOneAuraOrEquipmentPerTurn(t *testing.T) {
	g, me := s82mDanithaTable(t)
	equip := s82mGraveCard(me, "Grave Blade", "Artifact — Equipment")
	aura := s82mGraveCard(me, "Grave Aura", "Enchantment — Aura")
	bear := s82mGraveCard(me, "Grave Bear", "Creature — Bear")
	if !enumeratedFor(g, me.ID, equip) || !enumeratedFor(g, me.ID, aura) {
		t.Fatal("the Aura and the Equipment are not both offered")
	}
	if enumeratedFor(g, me.ID, bear) {
		t.Error("a creature card is offered")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("a creature was cast from the graveyard")
	}
	if err := g.CastSpell(me.ID, equip, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast the Equipment: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(equip) {
		t.Error("the Equipment did not resolve")
	}
	if enumeratedFor(g, me.ID, aura) {
		t.Error("a second Aura-or-Equipment spell is offered the same turn")
	}
}

func TestS82MatDanithaOnlyOnYourOwnTurn(t *testing.T) {
	g, me := s82mDanithaTable(t)
	equip := s82mGraveCard(me, "Grave Blade", "Artifact — Equipment")
	advanceToUpkeepOf(t, g, 1)
	if enumeratedFor(g, me.ID, equip) {
		t.Error("offered on an opponent's turn")
	}
}

// --- Kolaghan Warmonger -----------------------------------------------------

func TestS82MatWarmongerOffersOnlyDragonsFromTheTopSixAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	deeper := s82mTop(me, game.Card{Name: "Seventh", TypeLine: "Creature — Dragon"})
	top := hornTopFive(me, []struct{ name, typeLine string }{
		{"Mountain", "Basic Land — Mountain"},
		{"Whelp", "Creature — Dragon"},
		{"Shock", "Instant"},
		{"Drake", "Creature — Drake"},
		{"Wyrm", "Creature — Wyrm"},
		{"Dragon Egg", "Creature — Egg"},
	})
	warmonger := pushCatalogPermanent(g, me.ID, "Kolaghan Warmonger", "Creature — Ogre Warrior", s82mWarmongerOracle, false)
	finABAttack(t, g, opp.ID, warmonger)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("the attack raised no prompt")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != top[1] || prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Fatalf("offered %v [%d,%d], want only the Dragon, optional", prompt.ChooseCards, prompt.ChooseMin, prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[1]) {
		t.Error("the Dragon is not in hand")
	}
	if c, err := me.Library.Top(); err != nil || c.InstanceID != deeper {
		t.Errorf("the seventh card is not on top after the six went to the bottom: %+v", c)
	}
	for _, i := range []int{0, 2, 3, 4, 5} {
		if !me.Library.Contains(top[i]) {
			t.Errorf("card %d is not in the library", i)
		}
	}
}

// --- Nissa, Resurgent Animist -----------------------------------------------

// s82mLand puts a Forest onto the battlefield from the hand as an effect
// would, so it enters (it is not played) and landfall sees it.
func s82mLand(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("land enters: %v", err)
	}
	return id
}

func TestS82MatNissaAddsManaEachLandfallAndRevealsOnTheSecond(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Nissa, Resurgent Animist", "Legendary Creature — Elf Scout", s82mNissaOracle, false)
	elf := s82mTop(me, game.Card{Name: "Elf Hit", TypeLine: "Creature — Elf Warrior"})
	bear := s82mTop(me, game.Card{Name: "Bear Miss", TypeLine: "Creature — Bear"})
	elemental := s82mTop(me, game.Card{Name: "Elemental Later", TypeLine: "Creature — Elemental"})
	_ = elemental
	// Library top to bottom: Elemental Later, Bear Miss, Elf Hit, filler.
	hand := me.Hand.Size()

	s82mLand(t, g, me)
	passPriorityAroundTable(t, g)
	if n := riderAnswerManaPicks(t, g, me.ID, "G"); n != 1 {
		t.Fatalf("first landfall queued %d colour picks, want 1", n)
	}
	if me.Hand.Size() != hand {
		t.Fatal("the first resolution revealed cards")
	}
	s82mLand(t, g, me)
	passPriorityAroundTable(t, g)
	riderAnswerManaPicks(t, g, me.ID, "G")
	// The second resolution reveals until an Elf or Elemental: the top card.
	if !me.Hand.Contains(elemental) || me.Hand.Size() != hand+1 {
		t.Fatalf("the second resolution did not take the Elemental on top (hand %d→%d)", hand, me.Hand.Size())
	}
	s82mLand(t, g, me)
	passPriorityAroundTable(t, g)
	riderAnswerManaPicks(t, g, me.ID, "G")
	if me.Hand.Size() != hand+1 {
		t.Error("the third resolution revealed cards: only the second does")
	}
	_, _ = elf, bear
}

func TestS82MatNissaRevealRunPutsTheRestOnTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Nissa, Resurgent Animist", "Legendary Creature — Elf Scout", s82mNissaOracle, false)
	elf := s82mTop(me, game.Card{Name: "Elf Hit", TypeLine: "Creature — Elf Warrior"})
	bear := s82mTop(me, game.Card{Name: "Bear Miss", TypeLine: "Creature — Bear"})
	land := s82mTop(me, game.Card{Name: "Land Miss", TypeLine: "Basic Land — Forest"})
	size := me.Library.Size()
	for i := 0; i < 2; i++ {
		s82mLand(t, g, me)
		passPriorityAroundTable(t, g)
		riderAnswerManaPicks(t, g, me.ID, "G")
	}
	if !me.Hand.Contains(elf) {
		var names []string
		for _, c := range me.Hand.Cards {
			names = append(names, c.Name)
		}
		var lib []string
		for i := len(me.Library.Cards) - 1; i >= 0 && len(lib) < 5; i-- {
			lib = append(lib, me.Library.Cards[i].Name)
		}
		t.Fatalf("the Elf was not put into hand: hand %v library top %v", names, lib)
	}
	if me.Library.Size() != size-1 {
		t.Errorf("library %d → %d, want one card fewer", size, me.Library.Size())
	}
	for _, id := range []uuid.UUID{bear, land} {
		if !me.Library.Contains(id) {
			t.Error("a revealed card that missed left the library")
		}
	}
	if c, err := me.Library.Top(); err == nil && (c.InstanceID == bear || c.InstanceID == land) {
		t.Error("a missed card is still on top: the rest go to the bottom")
	}
}

// --- Pia Nalaar, Consul of Revival ------------------------------------------

func TestS82MatPiaMakesAThopterPerCastOrLandPlayFromExileOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Pia Nalaar, Consul of Revival", "Legendary Creature — Human Artificer", s82mPiaOracle, false)
	// A spell cast from hand makes nothing.
	castFromHandForTest(t, g, me, "Hand Bear", "Creature — Bear", "{0}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Thopter"); n != 0 {
		t.Fatalf("a spell cast from hand made %d Thopters", n)
	}
	// A spell cast from exile makes one.
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: spell, Name: "Exiled Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Owner: me.ID, Controller: me.ID})
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Duration: g.UntilEndOfTurnDuration()},
			[]game.Card{{InstanceID: spell, Name: "Exiled Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Owner: me.ID, Controller: me.ID}})
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Thopter"); n != 1 {
		t.Fatalf("a spell cast from exile made %d Thopters, want 1", n)
	}
	// A land played from exile makes another.
	land := uuid.New()
	g.WithWriteLock(func() {
		c := game.Card{InstanceID: land, Name: "Exiled Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
		g.Exile.PushTop(c)
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Duration: g.UntilEndOfTurnDuration()}, []game.Card{c})
	})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play a land from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Thopter"); n != 2 {
		t.Fatalf("a land played from exile left %d Thopters, want 2", n)
	}
}

func TestS82MatPiaGivesThoptersHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pia := pushCatalogPermanent(g, me.ID, "Pia Nalaar, Consul of Revival", "Legendary Creature — Human Artificer", s82mPiaOracle, false)
	thopter := b12Creature(g, me.ID, "Thopter", "Token Artifact Creature — Thopter", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, thopter, "haste") {
		t.Error("a Thopter lacks haste")
	}
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a Bear has haste")
	}
	_ = pia
}

// --- The Kenriths' Royal Funeral ----------------------------------------------

func TestS82MatFuneralExilesLegendsDrawsAndDiscountsLegendarySpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	big := graveyardCardOf(g, me, "Big Legend", "Legendary Creature — Dragon", "{3}{R}{R}", "")
	small := graveyardCardOf(g, me, "Small Legend", "Legendary Creature — Elf", "{1}{G}", "")
	plain := graveyardCardOf(g, me, "Plain Bear", "Creature — Bear", "{2}", "")
	hand, life := me.Hand.Size(), me.Life
	floatForTest(g, me, "WWBB")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "The Kenriths' Royal Funeral", TypeLine: "Legendary Enchantment",
			ManaCost: "{2}{W}{B}", OracleID: s82mFuneralOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// A non-legendary card is not a legal target.
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: plain}}); err == nil {
		t.Fatal("a non-legendary creature card was accepted as a target")
	}
	pickTargetRefs(t, g, me.ID, nil,
		game.TargetRef{Kind: game.TargetCard, ID: big}, game.TargetRef{Kind: game.TargetCard, ID: small})
	passPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{big, small} {
		if !inExile(g, c) {
			t.Fatalf("%s is not in exile", c)
		}
	}
	if !me.Graveyard.Contains(plain) {
		t.Error("the plain Bear left the graveyard")
	}
	if me.Hand.Size() != hand+5 || me.Life != life-5 {
		t.Errorf("hand %d→%d, life %d→%d: want +5 cards and -5 life (greatest mana value 5)", hand, me.Hand.Size(), life, me.Life)
	}
	// Two cards exiled with it: legendary spells cost {2} less.
	floatForTest(g, me, "RR")
	if err := s82mCastErr(g, me, "Legend Spell", "Legendary Creature — Dragon", "{3}{R}"); err != nil {
		t.Errorf("a {3}{R} legendary spell with {R}{R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	floatForTest(g, me, "RR")
	if err := s82mCastErr(g, me, "Plain Spell", "Creature — Bear", "{3}{R}"); err == nil {
		t.Error("a non-legendary spell was discounted")
	}
}

func s82mCastErr(g *game.Game, p *game.Player, name, typeLine, cost string) error {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
}

func TestS82MatFuneralWithNoTargetsDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	floatForTest(g, me, "WWBB")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "The Kenriths' Royal Funeral", TypeLine: "Legendary Enchantment",
			ManaCost: "{2}{W}{B}", OracleID: s82mFuneralOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("with nothing exiled hand %d→%d, life %d→%d", hand, me.Hand.Size(), life, me.Life)
	}
}

// --- Kiora, Sovereign of the Deep ---------------------------------------------

func TestS82MatKioraLooksAtXAndCastsAStrictlyCheaperSpellFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Kiora, Sovereign of the Deep", "Legendary Creature — Merfolk Noble", s82mKioraOracle, false)
	deeper := s82mTop(me, game.Card{Name: "Fourth", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	same := s82mTop(me, game.Card{Name: "Same Value", TypeLine: "Creature — Bear", ManaCost: "{3}", Power: 3, Toughness: 3})
	cheap := s82mTop(me, game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})
	land := s82mTop(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	floatForTest(g, me, "UUU")
	castFromHandForTest(t, g, me, "Sea Kraken", "Creature — Kraken", "{2}{U}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g) // the trigger resolves first
	offer := resolutionCastOfferFor(g, me.ID)
	if offer == nil {
		t.Fatalf("no free-cast offer after casting a Kraken from hand: pending %d, stack %d, kraken on bf %v", len(g.PendingChoices), g.Stack.Size(), b16CountNamed(g, "Sea Kraken"))
	}
	if slices.Contains(offer.DiscardOptions, same) || slices.Contains(offer.DiscardOptions, land) || slices.Contains(offer.DiscardOptions, deeper) {
		t.Errorf("offered %v: a spell of mana value 3 (not less than X = 3), a land and the fourth card are not castable", offer.DiscardOptions)
	}
	if !slices.Contains(offer.DiscardOptions, cheap) {
		t.Fatalf("the mana value 2 spell is not offered: %v", offer.DiscardOptions)
	}
	if err := castThroughResolutionCast(t, g, me.ID, cheap, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the free spell did not resolve")
	}
	if c, err := me.Library.Top(); err != nil || c.InstanceID != deeper {
		t.Errorf("the card below the looked-at three is not on top: only the rest of the three go to the bottom (%+v)", c)
	}
}

func TestS82MatKioraIgnoresOtherTypesAndCastsNotFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Kiora, Sovereign of the Deep", "Legendary Creature — Merfolk Noble", s82mKioraOracle, false)
	s82mTop(me, game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	floatForTest(g, me, "UUU")
	castFromHandForTest(t, g, me, "Big Bear", "Creature — Bear", "{2}{U}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a Bear spell triggered Kiora")
	}
	// A Serpent cast from the top of the library is not cast from hand.
	serpent := s82mTop(me, game.Card{Name: "Top Serpent", TypeLine: "Creature — Serpent", ManaCost: "{0}"})
	g.WithWriteLock(func() {
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneLibrary, Duration: g.UntilEndOfTurnDuration()},
			[]game.Card{{InstanceID: serpent, Name: "Top Serpent", TypeLine: "Creature — Serpent", ManaCost: "{0}", Owner: me.ID, Controller: me.ID}})
	})
	_ = g.CastSpell(me.ID, serpent, game.CastSpellParams{FromZone: "library"})
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a Serpent cast from the library triggered Kiora")
	}
}

// --- Nahiri, Forged in Fury ---------------------------------------------------

func TestS82MatNahiriExilesOnEquippedAttackAndCastsEquipmentFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Nahiri, Forged in Fury", "Legendary Creature — Kor Artificer", s82mNahiriOracle, false)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Spare Blade", "")
	s82mAttach(t, g, blade, bear)
	plain := seedBear(g, me.ID)
	top := s82mTop(me, game.Card{Name: "Free Blade", TypeLine: "Artifact — Equipment", ManaCost: "{5}"})
	before := me.Library.Size()
	finABAttack(t, g, opp.ID, bear, plain)
	passPriorityAroundTable(t, g)
	if !inExile(g, top) {
		t.Fatal("the top card was not exiled by the equipped attacker's trigger")
	}
	// Only the equipped attacker triggered: one card exiled, not two.
	if c, err := me.Library.Top(); err == nil && c.Name == "Free Blade" {
		t.Fatal("the library top is still the exiled card")
	}
	if me.Library.Size() != before-1 {
		t.Fatalf("library size %d → %d: exactly one card should be exiled", before, me.Library.Size())
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast the Equipment without paying its mana cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top) {
		t.Error("the free Equipment did not resolve")
	}
}

func TestS82MatNahiriOtherCardsAreCastAtFullCost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Nahiri, Forged in Fury", "Legendary Creature — Kor Artificer", s82mNahiriOracle, false)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Spare Blade", "")
	s82mAttach(t, g, blade, bear)
	top := s82mTop(me, game.Card{Name: "Pricey Bear", TypeLine: "Creature — Bear", ManaCost: "{4}", Power: 4, Toughness: 4})
	finABAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Fatal("a creature spell was cast without paying its mana cost")
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast with no mana to pay {4}")
	}
	floatForTest(g, me, "GGGG")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast at its printed cost: %v", err)
	}
}

func TestS82MatNahiriAffinityForEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedEquipment(g, me.ID, "Blade A", "")
	seedEquipment(g, me.ID, "Blade B", "")
	floatForTest(g, me, "RRWW")
	// {4}{R}{W} less two generic is {2}{R}{W}.
	if err := s82mCastErrOracle(g, me, "Nahiri, Forged in Fury", "Legendary Creature — Kor Artificer", "{4}{R}{W}", s82mNahiriOracle); err != nil {
		t.Errorf("Nahiri with two Equipment and four mana: %v", err)
	}
}

func s82mCastErrOracle(g *game.Game, p *game.Player, name, typeLine, cost, oracle string) error {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
}

// --- Arni Metalbrow -----------------------------------------------------------

func s82mArniTable(t *testing.T) (g *game.Game, me, opp *game.Player, arni uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	advanceToMain(t, g)
	arni = uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: arni, Name: "Arni Metalbrow", TypeLine: "Legendary Creature — Human Berserker",
		ManaCost: "{2}{R}", OracleID: s82mArniOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	return
}

func TestS82MatArniPaysToPutALesserCreatureAttacking(t *testing.T) {
	g, me, opp, arni := s82mArniTable(t)
	cheap := s82mHandCard(me, "Cheap Bear", "Creature — Bear", "{2}")
	equal := s82mHandCard(me, "Equal Bear", "Creature — Bear", "{3}")
	big := s82mHandCard(me, "Big Bear", "Creature — Bear", "{5}")
	finABAttack(t, g, opp.ID, arni)
	floatForTest(g, me, "RR")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no creature offer after paying")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != cheap {
		t.Fatalf("offered %v, want only the mana value 2 creature (Arni is mana value 3; %v and %v are not lower)", prompt.ChooseCards, equal, big)
	}
	answerChooseCards(t, g, me.ID, cheap)
	answerOptionPick(t, g, me.ID, 0) // the creature Arni's trigger came from is asked about first
	passPriorityAroundTable(t, g)
	c := flipState(t, g, cheap)
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("tapped=%v attacking=%v, want tapped and attacking", c.Tapped, c.AttackingTarget)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left: {1}{R} was not paid", me.ManaPool.Total())
	}
}

func TestS82MatArniDecliningPaysNothingAndPutsNothing(t *testing.T) {
	g, me, opp, arni := s82mArniTable(t)
	cheap := s82mHandCard(me, "Cheap Bear", "Creature — Bear", "{2}")
	finABAttack(t, g, opp.ID, arni)
	floatForTest(g, me, "RR")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	if latestChooseCards(g, me.ID) != nil || !me.Hand.Contains(cheap) || me.ManaPool.Total() != 2 {
		t.Error("declining the payment still did something")
	}
}

func TestS82MatArniTriggersWhenACreatureEntersAttackingButNotWhenItJustEnters(t *testing.T) {
	g, me, opp, arni := s82mArniTable(t)
	s82mHandCard(me, "Cheap Bear", "Creature — Bear", "{0}")
	finABAttack(t, g, opp.ID, arni)
	passPriorityAroundTable(t, g)
	// Decline the attack trigger's payment.
	if pu := s82mDeclineIfAny(t, g, me.ID); !pu {
		t.Fatal("the attack raised no payment offer")
	}
	// A creature that enters not attacking does not trigger Arni.
	b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 1, 1)
	passPriorityAroundTable(t, g)
	if s82mDeclineIfAny(t, g, me.ID) {
		t.Error("a creature entering without attacking triggered Arni")
	}
	// A creature that enters attacking does.
	g.WithWriteLock(func() {
		c := game.Card{InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Goblin", ManaCost: "{2}", Power: 1, Toughness: 1,
			Owner: me.ID, Controller: me.ID, Tapped: true, AttackingTarget: opp.ID}
		g.Battlefield.PushTop(c)
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: c.InstanceID, Actor: me.ID, NewZone: game.ZoneBattlefield})
	})
	passPriorityAroundTable(t, g)
	if !s82mDeclineIfAny(t, g, me.ID) {
		t.Error("a creature entering attacking did not trigger Arni")
	}
}

func s82mHandCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// s82mDeclineIfAny declines an open pay prompt and reports whether there was one.
func s82mDeclineIfAny(t *testing.T, g *game.Game, chooser uuid.UUID) bool {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			answerPayUnless(t, g, chooser, false)
			return true
		}
	}
	return false
}

func s82mAttach(t *testing.T, g *game.Game, gear, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(gear, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
}

// s82mTop pushes a card onto the top of the seat's library, keeping the
// rest (seedTopOfLibrary replaces the whole library).
func s82mTop(p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}
