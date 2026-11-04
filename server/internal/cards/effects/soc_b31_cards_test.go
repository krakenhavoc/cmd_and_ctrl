package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b31_cards_test.go — Secrets of Strixhaven, batch 31 (#4456). Each test
// plays the real catalog entry through the engine and asserts what the
// printed text says.

const (
	socB31GardensOracle    = "638ff242-63d5-457d-a7a6-40ad51052e2e"
	socB31SurveyorOracle   = "247cc471-fed5-41e3-8b6a-a10bbc129b36"
	socB31ProfessorOracle  = "4b8736e2-fc41-4a25-bb6c-7351afc40981"
	socB31SlothOracle      = "2970d09b-4bef-468e-8de5-f994c820ccc0"
	socB31MascotOracle     = "b8ecddf7-ba89-40dd-9f63-5bad27dd850d"
	socB31HamletOracle     = "78a2972c-14f4-41f3-99f9-167948bdd73a"
	socB31WranglerOracle   = pigmentWranglerOracleID
	socB31ApprenticeOracle = "4b0e5626-0282-42e5-abef-a079b26a8e5f"
	socB31TroveOracle      = "d604c75e-9d81-49de-80bc-08f4ef3ee975"
)

func TestSocB31EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB31GardensOracle:         "Paradox Gardens",
		socB31SurveyorOracle:        "Paradox Surveyor",
		socB31ProfessorOracle:       "Pensive Professor",
		socB31SlothOracle:           "Pestbrood Sloth",
		socB31MascotOracle:          "Pest Mascot",
		socB31HamletOracle:          "Petrified Hamlet",
		socB31WranglerOracle:        "Pigment Wrangler",
		socB31WranglerOracle + "#1": "Striking Palette",
		socB31ApprenticeOracle:      "Poisoner's Apprentice",
		socB31TroveOracle:           "Potioner's Trove",
	}
	for oracle, name := range want {
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

// --- Paradox Gardens ----------------------------------------------------------

func TestSocB31GardensEntersTappedMakesGOrUAndSurveilsForFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")
	id := playLandFromHand(t, g, "Paradox Gardens", socB31GardensOracle)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	untapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the land made {B}")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("{T}: Add {G}: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	untapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{T}: Add {U}: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	untapForTest(g, id)
	floatForTest(g, me, "CCGG")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated without {U} in the pool")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "CCGU")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{G}{U},{T}: surveil 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
	top := c.ScryCards[0]
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not go to the graveyard")
	}
	if card, _ := battlefieldCard(g, id); !card.Tapped {
		t.Error("the {T} cost did not tap the land")
	}
}

// --- Paradox Surveyor ---------------------------------------------------------

func TestSocB31SurveyorOffersOnlyLandsAndXCardsFromTheTopFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Hill", TypeLine: "Land"},
		game.Card{Name: "Fireball", TypeLine: "Sorcery", ManaCost: "{X}{R}"},
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Pump", TypeLine: "Instant", ManaCost: "{1}{G}", OracleID: ""},
		game.Card{Name: "Sixth", TypeLine: "Land"},
	)
	lib := me.Library.Size()
	msc17Play(t, g, me, game.Card{Name: "Paradox Surveyor", TypeLine: "Creature — Elf Druid", OracleID: socB31SurveyorOracle, Power: 3, Toughness: 3})
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choice was offered")
	}
	if len(pick.ChooseCards) != 2 || !hasID(pick.ChooseCards, ids[1]) || !hasID(pick.ChooseCards, ids[2]) {
		t.Fatalf("offered %v: only the land and the {X} card among the top five are legal (not the sixth card's land)", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, ids[2])
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the {X} card did not reach the hand")
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card taken", lib, me.Library.Size())
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[5] {
		t.Error("the sixth card must now be on top; the other four went to the bottom")
	}
	for _, id := range []uuid.UUID{ids[0], ids[1], ids[3], ids[4]} {
		if me.Hand.Contains(id) {
			t.Error("a card that was not picked reached the hand")
		}
	}
}

func TestSocB31SurveyorMayTakeNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Hill", TypeLine: "Land"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Bear 2", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Bear 3", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Bear 4", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Sixth", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	msc17Play(t, g, me, game.Card{Name: "Paradox Surveyor", TypeLine: "Creature — Elf Druid", OracleID: socB31SurveyorOracle, Power: 3, Toughness: 3})
	answerChooseCards(t, g, me.ID)
	if me.Hand.Contains(ids[0]) {
		t.Error("the land reached the hand though the player took nothing")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[5] {
		t.Error("the sixth card must be on top: all five looked-at cards go to the bottom")
	}
}

// --- Pensive Professor --------------------------------------------------------

func TestSocB31ProfessorDrawsWheneverItGetsAPlusOnePlusOneCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D", "E")
	prof := b12Push(g, me.ID, "Pensive Professor", "Creature — Human Wizard", socB31ProfessorOracle, 0, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	cast := func(cost string, n int) {
		t.Helper()
		if err := socB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: cost}, "G", n, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", cost, err)
		}
		passPriorityAroundTable(t, g)
	}
	cast("{G}", 1) // 1 > power 0: increment puts a counter, the Professor draws
	if n := plusOnes(g, prof); n != 1 {
		t.Fatalf("counters %d after a one-mana spell, want 1", n)
	}
	// The helper puts the spell in hand to cast it, so a net +1 is the draw.
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d after incrementing, want one card drawn", hand, got)
	}
	cast("{G}", 1) // 1 is not greater than 1/3
	if n := plusOnes(g, prof); n != 1 {
		t.Fatalf("counters %d after a second one-mana spell, want still 1", n)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after a spell that did not increment, want %d (no second draw)", got, hand+1)
	}
}

func TestSocB31ProfessorDrawsForACounterFromAnotherSourceButNotARemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D", "E")
	prof := b12Push(g, me.ID, "Pensive Professor", "Creature — Human Wizard", socB31ProfessorOracle, 0, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	addCountersForTest(g, prof, game.CounterPlusOne, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d after two counters at once, want exactly one draw", hand, got)
	}
	addCountersForTest(g, prof, game.CounterMinusOne, 1)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after a -1/-1 counter, want no further draw", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(prof, game.CounterPlusOne, -1) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after removing a +1/+1 counter, want no further draw", got)
	}
}

// --- Pestbrood Sloth ----------------------------------------------------------

func TestSocB31SlothMakesTwoPestsThatGainLifeWhenTheyAttackWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sloth := b12Push(g, me.ID, "Pestbrood Sloth", "Creature — Plant Sloth", socB31SlothOracle, 4, 4)
	if countNamed(g, "Pest") != 0 {
		t.Fatal("a Pest exists before the Sloth died")
	}
	b15Destroy(t, g, sloth)
	if countNamed(g, "Pest") != 2 {
		t.Fatalf("%d Pests after the Sloth died, want 2", countNamed(g, "Pest"))
	}
	var pestID uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Pest" {
			continue
		}
		pestID = c.InstanceID
		pest := s58p6View(t, g, c.InstanceID)
		if !pest.IsToken() || pest.CurrentPower() != 1 || pest.CurrentToughness() != 1 || len(pest.Colors) != 2 || c.Controller != me.ID {
			t.Fatalf("a Pest is token %v %d/%d colors %v controller %v, want yours, a 1/1 black and green token", pest.IsToken(), pest.CurrentPower(), pest.CurrentToughness(), pest.Colors, c.Controller)
		}
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	life := me.Life
	declareAttack(t, g, opp.ID, pestID)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d when a Pest attacked, want +1", life, me.Life)
	}
}

func TestSocB31SlothMakesNothingWhenItLeavesWithoutDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sloth := b12Push(g, me.ID, "Pestbrood Sloth", "Creature — Plant Sloth", socB31SlothOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(sloth) })
	passPriorityAroundTable(t, g)
	if countNamed(g, "Pest") != 0 {
		t.Error("exiling the Sloth made Pests: only dying does")
	}
}

// --- Pest Mascot --------------------------------------------------------------

func TestSocB31MascotGrowsOncePerLifeGainEvent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mascot := b12Push(g, me.ID, "Pest Mascot", "Creature — Pest Ape", socB31MascotOracle, 2, 3)
	advanceToMain(t, g)
	sos1GainLife(t, g) // gains 3 life in one event
	if n := plusOnes(g, mascot); n != 1 {
		t.Fatalf("counters %d after gaining 3 life in one event, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 4) })
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, mascot); n != 1 {
		t.Errorf("counters %d after an opponent gained life, want still 1", n)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, mascot); n != 2 {
		t.Errorf("counters %d after a second gain, want 2", n)
	}
}

// --- Petrified Hamlet ---------------------------------------------------------

func TestSocB31HamletNamesALandBansItsActivatedAbilitiesAndGivesItColorlessMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gardens := b21Push(g, me.ID, "Paradox Gardens", "Land", socB31GardensOracle, 0, 0)
	forest := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	hill := gaPut(g, me.ID, "Hill", "Land", "")
	hamlet := playLandFromHand(t, g, "Petrified Hamlet", socB31HamletOracle)
	passPriorityAroundTable(t, g)
	if c := pendingOfKind(g, game.PendingChoiceCardName); c == nil || c.Chooser != me.ID {
		t.Fatal("the Hamlet's enter trigger did not ask its controller for a name")
	}
	// Until the name is chosen it names nothing.
	if n := grantedManaRowsOf(g, gardens) + grantedManaRowsOf(g, forest); n != 0 {
		t.Fatalf("%d granted mana rows before a name was chosen, want 0", n)
	}
	actNameFor(t, g, me.ID, "Paradox Gardens")
	if n := grantedManaRowsOf(g, gardens); n != 1 {
		t.Errorf("the named land has %d granted mana rows, want 1", n)
	}
	if n := grantedManaRowsOf(g, forest) + grantedManaRowsOf(g, hill) + grantedManaRowsOf(g, hamlet); n != 0 {
		t.Errorf("%d granted rows on lands with other names, want 0", n)
	}
	// Its surveil is a non-mana activated ability and can't be activated; its
	// mana abilities (its own and the granted {C}) still can.
	floatForTest(g, me, "CCGU")
	if err := g.ActivateCatalogAbility(me.ID, gardens, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the named land's surveil: err = %v, want ErrCantActivate", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, gardens, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Errorf("the named land's own mana ability: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	untapForTest(g, gardens)
	if err := tapGrantedMana(t, g, me.ID, gardens, "C"); err != nil {
		t.Fatalf("the granted {T}: Add {C}: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
}

func TestSocB31HamletNamingABasicLandGivesItColorlessManaAndAnyPlayersCopyToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	theirs := gaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	other := gaPut(g, me.ID, "Island", "Basic Land — Island", "")
	playLandFromHand(t, g, "Petrified Hamlet", socB31HamletOracle)
	passPriorityAroundTable(t, g)
	actNameFor(t, g, me.ID, "forest")
	if grantedManaRowsOf(g, mine) != 1 || grantedManaRowsOf(g, theirs) != 1 {
		t.Error("every land with the chosen name gets the ability, whoever controls it (a basic land's name is a legal choice)")
	}
	if grantedManaRowsOf(g, other) != 0 {
		t.Error("an Island is not named Forest")
	}
	if err := tapGrantedMana(t, g, me.ID, mine, "C"); err != nil {
		t.Fatalf("tap the Forest for {C}: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
}

func TestSocB31HamletLeavesANonlandOfThatNameAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	krenko := actPushKrenko(g, me.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	playLandFromHand(t, g, "Petrified Hamlet", socB31HamletOracle)
	passPriorityAroundTable(t, g)
	actNameFor(t, g, me.ID, "Krenko, Mob Boss")
	if grantedManaRowsOf(g, krenko) != 0 {
		t.Error("a creature gained a land's mana ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); errors.Is(err, game.ErrCantActivate) {
		t.Errorf("Krenko's ability is refused though a land card name was meant: %v", err)
	}
}

// --- Pigment Wrangler // Striking Palette -------------------------------------

func socB31Wrangler(owner uuid.UUID) game.Card {
	return preparationCard(owner, socB31WranglerOracle, []string{"flying"},
		game.Face{Name: "Pigment Wrangler", TypeLine: "Creature — Orc Sorcerer", ManaCost: "{4}{R}", Power: 4, Toughness: 4},
		game.Face{Name: "Striking Palette", TypeLine: "Sorcery", ManaCost: "{R}"})
}

func TestSocB31WranglerEntersPreparedAndPaletteCopiesTheNextInstant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	c := socB31Wrangler(me.ID)
	castFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Pigment Wrangler did not enter prepared")
	}
	fraBCastCopy(t, g, me, "Striking Palette")
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("casting the copy did not unprepare the Wrangler")
	}
	life := opp.Life
	castBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	if opp.Life != life-6 {
		t.Errorf("life %d -> %d, want 6 damage: the Bolt and its copy", life, opp.Life)
	}
	// It was for the next spell only.
	life = opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 damage: a second Bolt is not copied", life, opp.Life)
	}
}

func TestSocB31PaletteDoesNotCopyANoninstantNonsorcerySpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	c := socB31Wrangler(me.ID)
	castFromHandAndResolve(t, g, me, c)
	fraBCastCopy(t, g, me, "Striking Palette")
	before := countNamed(g, "Bear")
	msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if countNamed(g, "Bear") != before+1 {
		t.Error("a creature spell was copied")
	}
}

// --- Poisoner's Apprentice ----------------------------------------------------

func TestSocB31ApprenticeShrinksAnOpponentsCreatureOnlyIfYouGainedLifeThisTurn(t *testing.T) {
	cast := func(t *testing.T, gain bool) (g *game.Game, victim, bystander, mine uuid.UUID) {
		g = newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim = b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
		bystander = b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		mine = b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		if gain {
			sos1GainLife(t, g)
		}
		msc17Cast(t, g, me, game.Card{Name: "Poisoner's Apprentice", TypeLine: "Creature — Orc Warlock", OracleID: socB31ApprenticeOracle, Power: 2, Toughness: 2}, game.CastSpellParams{})
		passPriorityAroundTable(t, g)
		pickIfAsked(t, g, me.ID, victim)
		passPriorityAroundTable(t, g)
		return g, victim, bystander, mine
	}
	t.Run("with life gained", func(t *testing.T) {
		g, victim, bystander, mine := cast(t, true)
		if got := effectivePower(t, g, victim); got != 1 {
			t.Errorf("the chosen creature's power %d, want 1 (5/5 gets -4/-4)", got)
		}
		if got := effectivePower(t, g, bystander); got != 2 {
			t.Errorf("an untargeted creature's power %d, want 2", got)
		}
		if got := effectivePower(t, g, mine); got != 2 {
			t.Errorf("my creature's power %d, want 2", got)
		}
	})
	t.Run("without life gained", func(t *testing.T) {
		g, victim, _, _ := cast(t, false)
		if got := effectivePower(t, g, victim); got != 5 {
			t.Errorf("the chosen creature's power %d with no life gained, want 5", got)
		}
	})
}

func TestSocB31ApprenticeKillsASmallCreatureAndOnlyTargetsOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	sos1GainLife(t, g)
	msc17Cast(t, g, me, game.Card{Name: "Poisoner's Apprentice", TypeLine: "Creature — Orc Warlock", OracleID: socB31ApprenticeOracle, Power: 2, Toughness: 2}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pickIfAsked(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("-4/-4 did not kill the opponent's 2/2")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's own creature is not a legal target")
	}
}

// --- Potioner's Trove ---------------------------------------------------------

func TestSocB31TroveMakesAnyColorAndGainsLifeOnlyAfterAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trove := b21Push(g, me.ID, "Potioner's Trove", "Artifact", socB31TroveOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, trove, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("{T}: Add one mana of any color: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	untapForTest(g, trove)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, trove, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the life ability activated before any instant or sorcery was cast")
	}
	// A creature spell does not turn it on.
	msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.ActivateCatalogAbility(me.ID, trove, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the life ability activated after only a creature spell")
	}
	castCatalogSpell(t, g, "Filler", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, trove, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate after an instant: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if err := g.ActivateCatalogAbility(me.ID, trove, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Trove activated again while tapped")
	}
}
