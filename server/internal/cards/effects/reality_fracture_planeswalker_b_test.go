package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_planeswalker_b_test.go — slice fra-planeswalker-b of
// the Reality Fracture set (#2795).

const (
	rfpbGarrukCurseBreakerOracle  = "184d4672-05a9-4182-8b3f-3561fb33ed62"
	rfpbGarrukVeiledButcherOracle = "4a6972fe-348a-4a19-a1d0-bdf8f1f9c792"
	rfpbGideonOathlessOracle      = "7afa530f-ea51-4bee-af3b-59b83ea5cb25"
	rfpbGideonsMemorialOracle     = "c8f08704-7330-4fa9-b456-2e96aa252a2c"
	rfpbIdentityEchoOracle        = "4dbf8c43-d7c5-4037-875e-6accd87aaf6f"
	rfpbInnovativeCommonsOracle   = "6812e4d3-0034-4fc8-9ee9-520f76a0ea94"
	rfpbJaceMultiverseOracle      = "3321c134-bf5b-4f63-8035-fca0bbdfa86b"
	rfpbKonstrariAnnexOracle      = "31ea4f3f-392a-4319-a95b-eb28790a80e6"
	rfpbLichsRelicOracle          = "0540eaca-0e03-4831-955a-192ae2b87d35"
	rfpbLilianaFaultlessOracle    = "22c2e66d-a2a1-46db-a70a-dff5ad5d6c14"
	rfpbLilianaRepentantOracle    = "5eb4403f-f199-4f75-a7c6-e76783f9b07d"
	rfpbLivingLibraryOracle       = "81f022a7-fd19-4340-b8a0-6b11ab3e9d91"
	rfpbLoyalTutorOracle          = "7e52f151-d1f6-4fcb-9b21-1baecae27da6"
	rfpbMabelOracle               = "36924990-8e3a-434e-bcd1-f03603cb350d"
	rfpbMassacreGirlOracle        = "81f35914-73bf-429f-ba20-32e23ac30159"
	rfpbMeticulousCommonsOracle   = "7fdd471f-ec95-4aa7-8f8a-ad8946de5ab4"
)

const rfpbWalkerLine = "Legendary Planeswalker — Test"

// rfpbWalker seats a planeswalker with the given type line and loyalty.
func rfpbWalker(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, loyalty int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

// rfpbCastCreature casts a creature with the given power from the active
// seat's hand (free) and settles it.
func rfpbCastCreature(t *testing.T, g *game.Game, name string, power int) uuid.UUID {
	t.Helper()
	toMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Test", Power: power, Toughness: power,
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func rfpbTargets(ids ...uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Targets: cardRefs(ids...)}
}

// rfpbSettle passes priority around the table, answering every
// pick_target prompt that comes up with refs, until nothing is left.
func rfpbSettle(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	for i := 0; i < 6; i++ {
		passPriorityAroundTable(t, g)
		pick := latestPickTarget(g, chooser)
		if pick == nil {
			return
		}
		var err error
		if len(refs) == 1 {
			err = g.ResolvePickTarget(pick.ID, chooser, refs[0])
		} else {
			err = g.ResolvePickTargets(pick.ID, chooser, refs)
		}
		if err != nil {
			t.Fatalf("resolve the target prompt: %v", err)
		}
	}
}

func rfpbCardRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

func rfpbPlayerRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

func rfpbAddMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// --- Loyal Tutor ---------------------------------------------------

func TestRFPBLoyalTutorPutsAPlaneswalkerOnTopRevealed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	needle := pushLibraryCardForTest(me, game.Card{Name: "Needle Walker", TypeLine: "Legendary Planeswalker — Needle"})
	pushLibraryCardForTest(me, game.Card{Name: "Decoy", TypeLine: "Creature — Bear"})
	libBefore, handBefore := me.Library.Size(), me.Hand.Size()

	castCatalogSpell(t, g, "Loyal Tutor", "Instant", rfpbLoyalTutorOracle, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchByID(t, g, me.ID, needle)
	}

	top, err := me.Library.Top()
	if err != nil || top.InstanceID != needle {
		t.Fatalf("top of library is not the planeswalker (err %v)", err)
	}
	if me.Hand.Contains(needle) || me.Hand.Size() != handBefore {
		t.Error("the planeswalker went to hand; Loyal Tutor puts it on top")
	}
	if me.Library.Size() != libBefore {
		t.Errorf("library %d, want %d", me.Library.Size(), libBefore)
	}
	if !top.IsKnownTo(opp.ID) {
		t.Error("an opponent does not know the revealed card")
	}
}

// With two planeswalkers to choose between, the controller picks; the
// creature is never an option.
func TestRFPBLoyalTutorOffersOnlyPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	first := pushLibraryCardForTest(me, game.Card{Name: "Walker One", TypeLine: "Legendary Planeswalker — One"})
	second := pushLibraryCardForTest(me, game.Card{Name: "Walker Two", TypeLine: "Legendary Planeswalker — Two"})
	decoy := pushLibraryCardForTest(me, game.Card{Name: "Decoy", TypeLine: "Creature — Bear"})

	castCatalogSpell(t, g, "Loyal Tutor", "Instant", rfpbLoyalTutorOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt with two planeswalkers in the library")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{decoy}); err == nil {
		t.Error("a creature was accepted as the tutor's pick")
	}
	answerSearchByID(t, g, me.ID, second)
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != second {
		t.Fatalf("the picked planeswalker is not on top (err %v)", err)
	}
	if !me.Library.Contains(first) {
		t.Error("the other planeswalker must stay in the library")
	}
}

func TestRFPBLoyalTutorFindsNothingWithoutAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLibraryCardForTest(me, game.Card{Name: "Decoy", TypeLine: "Creature — Bear"})
	libBefore := me.Library.Size()

	castCatalogSpell(t, g, "Loyal Tutor", "Instant", rfpbLoyalTutorOracle, nil)
	passPriorityAroundTable(t, g)

	if me.Library.Size() != libBefore {
		t.Errorf("library %d, want %d — nothing may move", me.Library.Size(), libBefore)
	}
}

// --- the three commons ---------------------------------------------

func TestRFPBCommonsEnterTappedUnlessYouControlAPlaneswalker(t *testing.T) {
	for _, land := range []struct{ name, oracle string }{
		{"Innovative Commons", rfpbInnovativeCommonsOracle},
		{"Konstrari Annex", rfpbKonstrariAnnexOracle},
		{"Meticulous Commons", rfpbMeticulousCommonsOracle},
	} {
		t.Run(land.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			bare := playLandFromHand(t, g, land.name, land.oracle)
			if c, ok := battlefieldCard(g, bare); !ok || !c.Tapped {
				t.Error("without a planeswalker the land must enter tapped")
			}
			rfpbWalker(g, me.ID, "Some Walker", rfpbWalkerLine, "", 3)
			withWalker := playLandFromHand(t, g, land.name, land.oracle)
			if c, ok := battlefieldCard(g, withWalker); !ok || c.Tapped {
				t.Error("with a planeswalker the land must enter untapped")
			}
			// An opponent's planeswalker does not count.
			g2 := newCatalogGame(t)
			rfpbWalker(g2, g2.Seats[1].ID, "Their Walker", rfpbWalkerLine, "", 3)
			theirs := playLandFromHand(t, g2, land.name, land.oracle)
			if c, ok := battlefieldCard(g2, theirs); !ok || !c.Tapped {
				t.Error("an opponent's planeswalker must not untap the land")
			}
		})
	}
}

// --- Garruk, Curse Breaker -----------------------------------------

func TestRFPBGarrukCurseBreakerDrawsOnlyForPowerFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", rfpbGarrukCurseBreakerOracle, 5)

	before := me.Hand.Size()
	rfpbCastCreature(t, g, "Big", 4) // pushed to hand (+1), cast (-1), drew (+1)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d: a power-4 creature draws a card", got, before+1)
	}
	before = me.Hand.Size()
	rfpbCastCreature(t, g, "Small", 3)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d: a power-3 creature draws nothing", got, before)
	}
}

func TestRFPBGarrukCurseBreakerPlusTwoUntapsUpToTwoLands(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := rfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", rfpbGarrukCurseBreakerOracle, 5)
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		l := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(l) })
		lands = append(lands, l)
	}
	b16Activate(t, g, me.ID, walker, 0, rfpbTargets(lands[0], lands[1]))
	for i, l := range lands {
		c, _ := battlefieldCard(g, l)
		if want := i == 2; c.Tapped != want {
			t.Errorf("land %d tapped=%v, want %v", i, c.Tapped, want)
		}
	}
	if got := loyaltyCount(g, walker); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestRFPBGarrukCurseBreakerMinusThreeMakesATrampleBeast(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := rfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", rfpbGarrukCurseBreakerOracle, 5)
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	beast := findBattlefieldByName(g, "Beast")
	if beast == uuid.Nil {
		t.Fatal("no Beast token")
	}
	if p := effectivePower(t, g, beast); p != 4 {
		t.Errorf("Beast power %d, want 4", p)
	}
	if !hasEffectiveKeyword(t, g, beast, "trample") {
		t.Error("the Beast must have trample")
	}
	if got := loyaltyCount(g, walker); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

// −4: until your next turn, creatures attacking one of your opponents get
// +2/+2 and trample — including an opponent's own creatures attacking
// another opponent — and nothing attacking YOU does.
func TestRFPBGarrukCurseBreakerMinusFourPumpsCreaturesAttackingYourOpponents(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := rfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", rfpbGarrukCurseBreakerOracle, 5)
	b16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})

	advanceToNextSeatsTurn(t, g)
	attackerSeat := g.Seats[g.Turn.ActiveSeat]
	var other *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID && s.ID != attackerSeat.ID {
			other = s
			break
		}
	}
	atOpp := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: attackerSeat.ID, Controller: attackerSeat.ID,
	})
	atMe := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raider Two", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: attackerSeat.ID, Controller: attackerSeat.ID,
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atOpp, other.ID); err != nil {
		t.Fatalf("attack the other player: %v", err)
	}
	if err := g.DeclareAttacker(atMe, me.ID); err != nil {
		t.Fatalf("attack Garruk's controller: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if p := effectivePower(t, g, atOpp); p != 5 {
		t.Errorf("a creature attacking one of Garruk's opponents has power %d, want 3+2", p)
	}
	if !hasEffectiveKeyword(t, g, atOpp, "trample") {
		t.Error("that creature must gain trample")
	}
	if p := effectivePower(t, g, atMe); p != 3 {
		t.Errorf("a creature attacking Garruk's controller got pumped (power %d)", p)
	}
}

// --- Garruk, Veiled Butcher ----------------------------------------

func TestRFPBGarrukVeiledButcherExilesOpponentCreaturesThatWouldDie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", rfpbGarrukVeiledButcherOracle, 5)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	if !g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's dying creature must be exiled instead")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	if !me.Graveyard.Contains(mine) {
		t.Error("my own creature must still die")
	}
}

func TestRFPBGarrukVeiledButcherPlusTwoShrinksUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	walker := rfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", rfpbGarrukVeiledButcherOracle, 5)
	target := b16Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	b16Activate(t, g, me.ID, walker, 0, rfpbTargets(target))
	if p, tough := rfPT(t, g, target); p != 1 || tough != 4 {
		t.Errorf("-4/-1 gave %d/%d, want 1/4", p, tough)
	}
	if got := loyaltyCount(g, walker); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestRFPBGarrukVeiledButcherMinusTwoEdictMakesABeastOnlyIfYouSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := rfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", rfpbGarrukVeiledButcherOracle, 5)
	fodder := b16Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	var theirs []uuid.UUID
	for _, s := range g.Seats {
		if s.ID != me.ID {
			theirs = append(theirs, b16Creature(g, s.ID, "Their Bear", "Creature — Bear", 2, 2))
		}
	}
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	answerSacrifice(t, g, me.ID, fodder)
	for i, s := range g.Seats[1:] {
		answerSacrifice(t, g, s.ID, theirs[i])
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Beast") == uuid.Nil {
		t.Error("I sacrificed a creature: the Beast token must be created")
	}
	for _, id := range theirs {
		if _, ok := battlefieldCard(g, id); ok {
			t.Error("every player sacrifices a creature of their choice")
		}
	}
}

func TestRFPBGarrukVeiledButcherMinusTwoWithNoCreatureOfYoursMakesNoBeast(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := rfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", rfpbGarrukVeiledButcherOracle, 5)
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Beast") != uuid.Nil {
		t.Error("no Beast without a sacrificed creature")
	}
}

func TestRFPBGarrukVeiledButcherMinusThreeDrawsPerOpponentWhoDidNotDiscardTwoNonland(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := rfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", rfpbGarrukVeiledButcherOracle, 5)
	var opps []*game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			opps = append(opps, s)
		}
	}
	// Opponent 0 holds two nonland cards; the others hold only lands.
	for _, o := range opps {
		for o.Hand.Size() > 0 {
			o.Hand.Cards = o.Hand.Cards[:len(o.Hand.Cards)-1]
		}
	}
	var spells [2]uuid.UUID
	for i := range spells {
		spells[i] = handCardFull(opps[0], "Spell", "Sorcery", "", "", nil)
	}
	for _, o := range opps[1:] {
		handCardFull(o, "Plains", "Basic Land — Plains", "", "", nil)
		handCardFull(o, "Plains", "Basic Land — Plains", "", "", nil)
	}
	handBefore := me.Hand.Size()
	b16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	for _, o := range opps {
		if discardOwed(g, o.ID) > 0 {
			if o == opps[0] {
				answerDiscard(t, g, o.ID, spells[0], spells[1])
			} else {
				discardFromHand(t, g, o.ID)
			}
		}
	}
	passPriorityAroundTable(t, g)
	want := len(opps) - 1 // everyone but opponent 0 discarded lands
	if got := me.Hand.Size() - handBefore; got != want {
		t.Errorf("drew %d cards, want %d", got, want)
	}
}

// --- Jace, Multiverse Architect ------------------------------------

func TestRFPBJaceMultiversePlusOneDrawsTwoThenBottomsOne(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := rfpbWalker(g, me.ID, "Jace, Multiverse Architect", "Legendary Planeswalker — Jace", rfpbJaceMultiverseOracle, 4)
	handBefore := me.Hand.Size()
	libBefore := me.Library.Size()
	pick := me.Hand.Cards[0].InstanceID

	b16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	answerChooseCards(t, g, me.ID, pick)

	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d, want %d (draw two, bottom one)", got, handBefore+1)
	}
	if me.Hand.Contains(pick) {
		t.Error("the chosen card must leave the hand")
	}
	if me.Library.Cards[0].InstanceID != pick {
		t.Error("the chosen card must be on the BOTTOM of the library")
	}
	if got := me.Library.Size(); got != libBefore-1 {
		t.Errorf("library %d, want %d", got, libBefore-1)
	}
	if got := loyaltyCount(g, jace); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestRFPBJaceMultiverseMinusThreeExilesAnotherAndRevealsUntilACreature(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := rfpbWalker(g, me.ID, "Jace, Multiverse Architect", "Legendary Planeswalker — Jace", rfpbJaceMultiverseOracle, 4)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	prize := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: prize, Name: "Prize", TypeLine: "Creature — Giant", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	miss1 := uuid.New()
	miss2 := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: miss1, Name: "Miss One", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: miss2, Name: "Miss Two", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})

	// Jace may not target himself.
	if err := g.ActivateCatalogAbility(me.ID, jace, 1, rfpbTargets(jace)); err == nil {
		t.Fatal("−3 targeted Jace himself: it says another")
	}
	b16Activate(t, g, me.ID, jace, 1, rfpbTargets(bear))
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(bear) {
		t.Error("the targeted creature must be exiled")
	}
	if _, ok := battlefieldCard(g, prize); !ok {
		t.Error("the revealed creature must enter the battlefield")
	}
	if me.Library.Cards[0].InstanceID != miss1 && me.Library.Cards[0].InstanceID != miss2 {
		t.Error("the revealed misses must go on the bottom of the library")
	}
	if got := loyaltyCount(g, jace); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

// --- Identity Echo -------------------------------------------------

func TestRFPBIdentityEchoExilesYourPermanentAndPutsTheRevealedOneOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	echo := pushCatalogPermanent(g, me.ID, "Identity Echo", "Enchantment", rfpbIdentityEchoOracle, false)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	prize := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: prize, Name: "Prize", TypeLine: "Creature — Giant", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	miss := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: miss, Name: "Miss", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	rfpbAddMana(me, "R", "R", "R", "R")

	b16Activate(t, g, me.ID, echo, 0, rfpbTargets(bear))
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(bear) {
		t.Error("the target must be exiled")
	}
	if _, ok := battlefieldCard(g, prize); !ok {
		t.Error("the revealed creature must enter the battlefield")
	}
	if me.Library.Cards[0].InstanceID != miss {
		t.Error("the revealed miss goes to the bottom")
	}
}

func TestRFPBIdentityEchoIsSorceryOnlyAndNeedsYourOwnTarget(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	echo := pushCatalogPermanent(g, me.ID, "Identity Echo", "Enchantment", rfpbIdentityEchoOracle, false)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rfpbAddMana(me, "R", "R", "R", "R", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, echo, 0, rfpbTargets(theirs)); err == nil {
		t.Error("Identity Echo exiled an opponent's creature")
	}
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	rfpbAddMana(me, "R", "R", "R", "R", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, echo, 0, rfpbTargets(mine)); err == nil {
		t.Error("Identity Echo activated outside a main phase")
	}
}

// --- Living Library ------------------------------------------------

func TestRFPBLivingLibraryShufflesAnOpponentsCreatureIntoTheirLibrary(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lib := b12Push(g, me.ID, "Living Library", "Artifact Creature — Book Illusion", rfpbLivingLibraryOracle, 0, 4)
	victim := b16Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	rfpbAddMana(me, "G", "G", "G", "G", "G", "G")
	libBefore := opp.Library.Size()

	b16Activate(t, g, me.ID, lib, 0, rfpbTargets(victim))
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, lib); ok {
		t.Error("Living Library must be sacrificed")
	}
	if _, ok := battlefieldCard(g, victim); ok || !opp.Library.Contains(victim) {
		t.Error("the creature must be shuffled into its owner's library")
	}
	if got := opp.Library.Size(); got != libBefore+1 {
		t.Errorf("their library %d, want %d", got, libBefore+1)
	}
}

func TestRFPBLivingLibraryRefusesYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := b12Push(g, me.ID, "Living Library", "Artifact Creature — Book Illusion", rfpbLivingLibraryOracle, 0, 4)
	mine := b16Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	rfpbAddMana(me, "G", "G", "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, lib, 0, rfpbTargets(mine)); err == nil {
		t.Error("Living Library targeted its controller's own creature")
	}
}

// --- Gideon's Memorial ---------------------------------------------

func TestRFPBGideonsMemorialAnthemTouchesOnlyYourCreatureTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Gideon's Memorial", "Legendary Artifact", rfpbGideonsMemorialOracle, false)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")
	card := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	var theirs uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.Controller == opp.ID {
			theirs = c.InstanceID
		}
	}
	if p := effectivePower(t, g, goblin); p != 2 {
		t.Errorf("my token power %d, want 1+1", p)
	}
	if !hasEffectiveKeyword(t, g, goblin, "vigilance") {
		t.Error("my token must have vigilance")
	}
	if p := effectivePower(t, g, card); p != 2 || hasEffectiveKeyword(t, g, card, "vigilance") {
		t.Error("a nontoken creature must be untouched")
	}
	if p := effectivePower(t, g, theirs); p != 1 || hasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's token must be untouched")
	}
}

func TestRFPBGideonsMemorialManaPaysOnlyForPlaneswalkerSpells(t *testing.T) {
	g, me, _ := spendTable(t)
	mem := pushCatalogPermanent(g, me.ID, "Gideon's Memorial", "Legendary Artifact", rfpbGideonsMemorialOracle, false)
	if err := g.ActivateManaAbility(me.ID, mem, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap the Memorial: %v", err)
	}
	bear := handSpell(me, "Green Creature", "Creature — Bear", "{G}")
	refusedForMana(t, g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}), "the Memorial's mana for a creature spell")
	walker := handSpell(me, "Green Walker", "Legendary Planeswalker — Test", "{G}")
	if err := g.CastSpell(me.ID, walker, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Memorial's {G} for a planeswalker spell: %v", err)
	}
}

func TestRFPBGideonsMemorialChannelDamagesAnAttackerFromHand(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mem := handCardFull(me, "Gideon's Memorial", "Legendary Artifact", "{1}{W}", rfpbGideonsMemorialOracle, []string{"W"})
	attacker := b16Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 4)
	idle := b16Creature(g, me.ID, "Idle", "Creature — Ogre", 3, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("declare attacker: %v", err)
	}
	lockInAttacks(t, g)
	rfpbAddMana(me, "W", "W")

	if err := g.ActivateCatalogAbility(me.ID, mem, 0, rfpbTargets(idle)); err == nil {
		t.Fatal("the channel hit a creature that is neither attacking nor blocking")
	}
	if err := g.ActivateCatalogAbility(me.ID, mem, 0, rfpbTargets(attacker)); err != nil {
		t.Fatalf("channel at the attacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, attacker); ok {
		t.Error("4 damage must kill a 3/4 attacker")
	}
	if !me.Graveyard.Contains(mem) {
		t.Error("the Memorial is discarded as a cost")
	}
}

// --- Gideon the Oathless -------------------------------------------

func TestRFPBGideonOathlessPingsThePlayerWhoseCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, opp.ID, "Gideon the Oathless", "Legendary Creature — Human Mercenary", rfpbGideonOathlessOracle, 3, 3)
	myLife, oppLife := me.Life, opp.Life

	rfpbCastCreature(t, g, "Mine", 2)
	if me.Life != myLife-1 {
		t.Errorf("my life %d, want %d: a creature an opponent of Gideon's controller controls entered", me.Life, myLife-1)
	}
	if opp.Life != oppLife {
		t.Errorf("Gideon's controller lost life: %d", opp.Life)
	}
}

func TestRFPBGideonOathlessIgnoresItsControllersOwnCreaturesAndAbilities(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Gideon the Oathless", "Legendary Creature — Human Mercenary", rfpbGideonOathlessOracle, 3, 3)
	walker := rfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", rfpbGarrukCurseBreakerOracle, 5)
	life := me.Life
	rfpbCastCreature(t, g, "Mine", 2)
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d: Gideon's own controller is never pinged", me.Life, life)
	}
	for _, s := range g.Seats[1:] {
		if s.Life != me.Life {
			// Every seat starts at the same total; nobody else may have lost any.
			t.Errorf("seat %s has life %d", s.Name, s.Life)
		}
	}
}

func TestRFPBGideonOathlessPingsAnOpponentWhoActivatesALoyaltyAbility(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, opp.ID, "Gideon the Oathless", "Legendary Creature — Human Mercenary", rfpbGideonOathlessOracle, 3, 3)
	walker := rfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", rfpbGarrukCurseBreakerOracle, 5)
	before := me.Life
	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{}) // +2, makes no creature
	passPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("life %d, want %d: an opponent activated a loyalty ability", me.Life, before-1)
	}
}

func TestRFPBGideonOathlessIsDeclaredWithItsWardCaveat(t *testing.T) {
	spec, ok := Lookup(rfpbGideonOathlessOracle)
	if !ok || spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Fatalf("Gideon the Oathless must ship with the ward caveat: %+v", spec.Caveats)
	}
}

// --- Lich's Relic --------------------------------------------------

func TestRFPBLichsRelicGivesPlusTwoPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := pushCatalogPermanent(g, me.ID, "Lich's Relic", "Artifact — Equipment", rfpbLichsRelicOracle, false)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rfpbAddMana(me, "B", "B")
	b16Activate(t, g, me.ID, relic, 0, rfpbTargets(bear))
	passPriorityAroundTable(t, g)
	if p, tough := rfPT(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("equipped bear is %d/%d, want 4/3", p, tough)
	}
}

// rfpbCastRelic casts Lich's Relic for {B} with extra mana in the pool and
// settles the spell, leaving the "you may pay {2}" prompt open.
func rfpbCastRelic(t *testing.T, g *game.Game, me *game.Player, extra int) {
	t.Helper()
	toMainForCost(t, g)
	relic := handCardFull(me, "Lich's Relic", "Artifact — Equipment", "{B}", rfpbLichsRelicOracle, []string{"B"})
	rfpbAddMana(me, "B")
	for i := 0; i < extra; i++ {
		rfpbAddMana(me, "B")
	}
	if err := g.CastSpell(me.ID, relic, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Lich's Relic: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestRFPBLichsRelicPaidDestroysOnePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var victims []uuid.UUID
	var refs []game.TargetRef
	for _, s := range g.Seats {
		if s.ID != me.ID {
			v := b16Creature(g, s.ID, "Bear", "Creature — Bear", 2, 2)
			victims = append(victims, v)
			refs = append(refs, rfpbCardRef(v))
		}
	}
	rfpbCastRelic(t, g, me, 2)
	answerPayUnless(t, g, me.ID, true)
	rfpbSettle(t, g, me.ID, refs...)
	for _, v := range victims {
		if _, ok := battlefieldCard(g, v); ok {
			t.Error("a victim survived the Relic")
		}
	}
}

func TestRFPBLichsRelicTakesOnlyOnePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b16Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b16Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	rfpbCastRelic(t, g, me, 2)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the destroy trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{rfpbCardRef(a), rfpbCardRef(b)}); err == nil {
		t.Error("two creatures of the same opponent were accepted: one per opponent")
	}
}

func TestRFPBLichsRelicDeclinedPaymentDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rfpbCastRelic(t, g, me, 2)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, victim); !ok {
		t.Error("the creature died although {2} was not paid")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt appeared although {2} was not paid")
	}
}

// --- Liliana the Faultless / the Repentant -------------------------

func TestRFPBLilianaFaultlessGainsLifeAndGrantsHexproof(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lili := b12Push(g, me.ID, "Liliana the Faultless", "Legendary Creature — Human Cleric", rfpbLilianaFaultlessOracle, 1, 1)
	life := me.Life
	bear := rfpbCastCreature(t, g, "Bear", 2)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d: another creature entered", me.Life, life+1)
	}
	rfpbAddMana(me, "W")
	pitch := handCardFull(me, "Pitch", "Sorcery", "", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{Targets: cardRefs(lili), DiscardIDs: []uuid.UUID{pitch}}); err == nil {
		t.Error("Liliana targeted herself")
	}
	b16Activate(t, g, me.ID, lili, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), DiscardIDs: []uuid.UUID{pitch}})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the bear must gain hexproof")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard is a cost")
	}
}

func TestRFPBLilianaRepentantMillsAndReanimatesOnce(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lili := b12Push(g, me.ID, "Liliana the Repentant", "Legendary Creature — Human Warlock", rfpbLilianaRepentantOracle, 2, 2)
	libBefore := me.Library.Size()
	rfpbCastCreature(t, g, "Bear", 2)
	if got := me.Library.Size(); got != libBefore-2 {
		t.Errorf("library %d, want %d: mill two", got, libBefore-2)
	}
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen", TypeLine: "Creature — Giant", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	rfpbAddMana(me, "B", "B", "B", "B", "B", "B")
	b16Activate(t, g, me.ID, lili, 0, game.ActivateAbilityParams{Targets: cardRefs(dead)})
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, dead); !ok {
		t.Error("the creature must return to the battlefield")
	}
	if p, _ := rfPT(t, g, lili); p != 3 {
		t.Errorf("Liliana has power %d, want 2+1", p)
	}
	rfpbAddMana(me, "B", "B", "B", "B", "B", "B")
	dead2 := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead2, Name: "Fallen Two", TypeLine: "Creature — Giant", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{Targets: cardRefs(dead2)}); err == nil {
		t.Error("an exhaust ability activated twice")
	}
}

func TestRFPBLilianaRepentantRefusesAnOpponentsGraveyardAndALand(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lili := b12Push(g, me.ID, "Liliana the Repentant", "Legendary Creature — Human Warlock", rfpbLilianaRepentantOracle, 2, 2)
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Creature — Giant", Owner: opp.ID, Controller: opp.ID})
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
	rfpbAddMana(me, "B", "B", "B", "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Error("reanimated a card from an opponent's graveyard")
	}
	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{Targets: cardRefs(land)}); err == nil {
		t.Error("reanimated a land card")
	}
}

// --- Mabel ---------------------------------------------------------

func rfpbCastMabel(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	toMainForCost(t, g)
	mabel := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mabel, Name: "Mabel, Bitter Recluse", TypeLine: "Legendary Creature — Mouse Warlock",
		OracleID: rfpbMabelOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, mabel, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Mabel: %v", err)
	}
	rfpbSettle(t, g, me.ID, rfpbCardRef(target))
}

func TestRFPBMabelRemovesUpToThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := b16Creature(g, opp.ID, "Hydra", "Creature — Hydra", 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(target, game.CounterPlusOne, 5) })
	rfpbCastMabel(t, g, me, target)
	for i := 0; i < 3; i++ {
		answerOptionPick(t, g, me.ID, 1)
	}
	c, _ := battlefieldCard(g, target)
	if got := c.Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("counters left %d, want 5-3", got)
	}
}

func TestRFPBMabelMayStopEarlyAndHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := b16Creature(g, opp.ID, "Hydra", "Creature — Hydra", 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(target, game.CounterPlusOne, 5) })
	rfpbCastMabel(t, g, me, target)
	answerOptionPick(t, g, me.ID, 1)
	answerOptionPick(t, g, me.ID, 0)
	c, _ := battlefieldCard(g, target)
	if got := c.Counters[game.CounterPlusOne]; got != 4 {
		t.Errorf("counters left %d, want 4 after stopping", got)
	}
	mabel := findBattlefieldByName(g, "Mabel, Bitter Recluse")
	if mabel == uuid.Nil || !hasEffectiveKeyword(t, g, mabel, "deathtouch") {
		t.Error("Mabel must have deathtouch")
	}
}

// --- Massacre Girl, Most Wanted ------------------------------------

func TestRFPBMassacreGirlDrainsWhenYourCreatureOrPlaneswalkerDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Massacre Girl, Most Wanted", "Legendary Creature — Human Assassin", rfpbMassacreGirlOracle, 4, 4)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	walker := rfpbWalker(g, me.ID, "Doomed Walker", rfpbWalkerLine, "", 3)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	rfpbSettle(t, g, me.ID, rfpbPlayerRef(opp.ID))
	if me.Life != myLife+1 || opp.Life != oppLife-1 {
		t.Errorf("after a creature died: life %d/%d, want %d/%d", me.Life, opp.Life, myLife+1, oppLife-1)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(walker) })
	rfpbSettle(t, g, me.ID, rfpbPlayerRef(opp.ID))
	if me.Life != myLife+2 || opp.Life != oppLife-2 {
		t.Errorf("after a planeswalker died: life %d/%d", me.Life, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	rfpbSettle(t, g, me.ID, rfpbPlayerRef(opp.ID))
	if me.Life != myLife+2 {
		t.Error("an opponent's creature dying must not trigger her")
	}
}

func TestRFPBMassacreGirlGrowsOnNoncombatDamageToAnOpponentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	girl := b12Push(g, me.ID, "Massacre Girl, Most Wanted", "Legendary Creature — Human Assassin", rfpbMassacreGirlOracle, 4, 4)
	src := b16Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if p, _ := rfPT(t, g, girl); p != 5 {
		t.Errorf("power %d, want 5 after noncombat damage to an opponent", p)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 2) })
	passPriorityAroundTable(t, g)
	if p, _ := rfPT(t, g, girl); p != 5 {
		t.Errorf("power %d: damage to her own controller must not grow her", p)
	}
}
