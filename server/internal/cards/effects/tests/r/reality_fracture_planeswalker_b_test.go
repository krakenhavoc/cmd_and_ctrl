package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const rfpbWalkerLine = "Legendary Planeswalker — Test"

func rfpbTargets(ids ...uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Targets: effectstest.CardRefs(ids...)}
}

// rfpbSettle passes priority around the table, answering every
// pick_target prompt that comes up with refs, until nothing is left.
func rfpbSettle(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	for i := 0; i < 6; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestPickTarget(g, chooser)
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

func rfpbPlayerRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

// --- Loyal Tutor ---------------------------------------------------

func TestRFPBLoyalTutorPutsAPlaneswalkerOnTopRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Needle Walker", TypeLine: "Legendary Planeswalker — Needle"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Decoy", TypeLine: "Creature — Bear"})
	libBefore, handBefore := me.Library.Size(), me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Loyal Tutor", "Instant", effectstest.RfpbLoyalTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, needle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	first := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Walker One", TypeLine: "Legendary Planeswalker — One"})
	second := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Walker Two", TypeLine: "Legendary Planeswalker — Two"})
	decoy := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Decoy", TypeLine: "Creature — Bear"})

	effectstest.CastCatalogSpell(t, g, "Loyal Tutor", "Instant", effectstest.RfpbLoyalTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt with two planeswalkers in the library")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{decoy}); err == nil {
		t.Error("a creature was accepted as the tutor's pick")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, second)
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != second {
		t.Fatalf("the picked planeswalker is not on top (err %v)", err)
	}
	if !me.Library.Contains(first) {
		t.Error("the other planeswalker must stay in the library")
	}
}

func TestRFPBLoyalTutorFindsNothingWithoutAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Decoy", TypeLine: "Creature — Bear"})
	libBefore := me.Library.Size()

	effectstest.CastCatalogSpell(t, g, "Loyal Tutor", "Instant", effectstest.RfpbLoyalTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Library.Size() != libBefore {
		t.Errorf("library %d, want %d — nothing may move", me.Library.Size(), libBefore)
	}
}

// --- the three commons ---------------------------------------------

func TestRFPBCommonsEnterTappedUnlessYouControlAPlaneswalker(t *testing.T) {
	for _, land := range []struct{ name, oracle string }{
		{"Innovative Commons", effectstest.RfpbInnovativeCommonsOracle},
		{"Konstrari Annex", effectstest.RfpbKonstrariAnnexOracle},
		{"Meticulous Commons", effectstest.RfpbMeticulousCommonsOracle},
	} {
		t.Run(land.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			bare := effectstest.PlayLandFromHand(t, g, land.name, land.oracle)
			if c, ok := effectstest.BattlefieldCard(g, bare); !ok || !c.Tapped {
				t.Error("without a planeswalker the land must enter tapped")
			}
			effectstest.RfpbWalker(g, me.ID, "Some Walker", rfpbWalkerLine, "", 3)
			withWalker := effectstest.PlayLandFromHand(t, g, land.name, land.oracle)
			if c, ok := effectstest.BattlefieldCard(g, withWalker); !ok || c.Tapped {
				t.Error("with a planeswalker the land must enter untapped")
			}
			// An opponent's planeswalker does not count.
			g2 := effectstest.NewCatalogGame(t)
			effectstest.RfpbWalker(g2, g2.Seats[1].ID, "Their Walker", rfpbWalkerLine, "", 3)
			theirs := effectstest.PlayLandFromHand(t, g2, land.name, land.oracle)
			if c, ok := effectstest.BattlefieldCard(g2, theirs); !ok || !c.Tapped {
				t.Error("an opponent's planeswalker must not untap the land")
			}
		})
	}
}

// --- Garruk, Curse Breaker -----------------------------------------

func TestRFPBGarrukCurseBreakerDrawsOnlyForPowerFourOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukCurseBreakerOracle, 5)

	before := me.Hand.Size()
	effectstest.RfpbCastCreature(t, g, "Big", 4) // pushed to hand (+1), cast (-1), drew (+1)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d: a power-4 creature draws a card", got, before+1)
	}
	before = me.Hand.Size()
	effectstest.RfpbCastCreature(t, g, "Small", 3)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d: a power-3 creature draws nothing", got, before)
	}
}

func TestRFPBGarrukCurseBreakerPlusTwoUntapsUpToTwoLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.RfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukCurseBreakerOracle, 5)
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		l := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(l) })
		lands = append(lands, l)
	}
	effectstest.B16Activate(t, g, me.ID, walker, 0, rfpbTargets(lands[0], lands[1]))
	for i, l := range lands {
		c, _ := effectstest.BattlefieldCard(g, l)
		if want := i == 2; c.Tapped != want {
			t.Errorf("land %d tapped=%v, want %v", i, c.Tapped, want)
		}
	}
	if got := effectstest.LoyaltyCount(g, walker); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestRFPBGarrukCurseBreakerMinusThreeMakesATrampleBeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.RfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukCurseBreakerOracle, 5)
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	beast := effectstest.FindBattlefieldByName(g, "Beast")
	if beast == uuid.Nil {
		t.Fatal("no Beast token")
	}
	if p := effectstest.EffectivePower(t, g, beast); p != 4 {
		t.Errorf("Beast power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, beast, "trample") {
		t.Error("the Beast must have trample")
	}
	if got := effectstest.LoyaltyCount(g, walker); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

// −4: until your next turn, creatures attacking one of your opponents get
// +2/+2 and trample — including an opponent's own creatures attacking
// another opponent — and nothing attacking YOU does.
func TestRFPBGarrukCurseBreakerMinusFourPumpsCreaturesAttackingYourOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.RfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukCurseBreakerOracle, 5)
	effectstest.B16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})

	effectstest.AdvanceToNextSeatsTurn(t, g)
	attackerSeat := g.Seats[g.Turn.ActiveSeat]
	var other *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID && s.ID != attackerSeat.ID {
			other = s
			break
		}
	}
	atOpp := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: attackerSeat.ID, Controller: attackerSeat.ID,
	})
	atMe := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raider Two", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: attackerSeat.ID, Controller: attackerSeat.ID,
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atOpp, other.ID); err != nil {
		t.Fatalf("attack the other player: %v", err)
	}
	if err := g.DeclareAttacker(atMe, me.ID); err != nil {
		t.Fatalf("attack Garruk's controller: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if p := effectstest.EffectivePower(t, g, atOpp); p != 5 {
		t.Errorf("a creature attacking one of Garruk's opponents has power %d, want 3+2", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, atOpp, "trample") {
		t.Error("that creature must gain trample")
	}
	if p := effectstest.EffectivePower(t, g, atMe); p != 3 {
		t.Errorf("a creature attacking Garruk's controller got pumped (power %d)", p)
	}
}

// --- Garruk, Veiled Butcher ----------------------------------------

func TestRFPBGarrukVeiledButcherExilesOpponentCreaturesThatWouldDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.RfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukVeiledButcherOracle, 5)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

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
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	walker := effectstest.RfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukVeiledButcherOracle, 5)
	target := effectstest.B16Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	effectstest.B16Activate(t, g, me.ID, walker, 0, rfpbTargets(target))
	if p, tough := effectstest.RfPT(t, g, target); p != 1 || tough != 4 {
		t.Errorf("-4/-1 gave %d/%d, want 1/4", p, tough)
	}
	if got := effectstest.LoyaltyCount(g, walker); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestRFPBGarrukVeiledButcherMinusTwoEdictMakesABeastOnlyIfYouSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.RfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukVeiledButcherOracle, 5)
	fodder := effectstest.B16Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	var theirs []uuid.UUID
	for _, s := range g.Seats {
		if s.ID != me.ID {
			theirs = append(theirs, effectstest.B16Creature(g, s.ID, "Their Bear", "Creature — Bear", 2, 2))
		}
	}
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	for i, s := range g.Seats[1:] {
		effectstest.AnswerSacrifice(t, g, s.ID, theirs[i])
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Beast") == uuid.Nil {
		t.Error("I sacrificed a creature: the Beast token must be created")
	}
	for _, id := range theirs {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Error("every player sacrifices a creature of their choice")
		}
	}
}

func TestRFPBGarrukVeiledButcherMinusTwoWithNoCreatureOfYoursMakesNoBeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.RfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukVeiledButcherOracle, 5)
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Beast") != uuid.Nil {
		t.Error("no Beast without a sacrificed creature")
	}
}

func TestRFPBGarrukVeiledButcherMinusThreeDrawsPerOpponentWhoDidNotDiscardTwoNonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.RfpbWalker(g, me.ID, "Garruk, Veiled Butcher", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukVeiledButcherOracle, 5)
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
		spells[i] = effectstest.HandCardFull(opps[0], "Spell", "Sorcery", "", "", nil)
	}
	for _, o := range opps[1:] {
		effectstest.HandCardFull(o, "Plains", "Basic Land — Plains", "", "", nil)
		effectstest.HandCardFull(o, "Plains", "Basic Land — Plains", "", "", nil)
	}
	handBefore := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	for _, o := range opps {
		if effectstest.DiscardOwed(g, o.ID) > 0 {
			if o == opps[0] {
				effectstest.AnswerDiscard(t, g, o.ID, spells[0], spells[1])
			} else {
				effectstest.DiscardFromHand(t, g, o.ID)
			}
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	want := len(opps) - 1 // everyone but opponent 0 discarded lands
	if got := me.Hand.Size() - handBefore; got != want {
		t.Errorf("drew %d cards, want %d", got, want)
	}
}

// --- Jace, Multiverse Architect ------------------------------------

func TestRFPBJaceMultiversePlusOneDrawsTwoThenBottomsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := effectstest.RfpbWalker(g, me.ID, "Jace, Multiverse Architect", "Legendary Planeswalker — Jace", effectstest.RfpbJaceMultiverseOracle, 4)
	handBefore := me.Hand.Size()
	libBefore := me.Library.Size()
	pick := me.Hand.Cards[0].InstanceID

	effectstest.B16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	effectstest.AnswerChooseCards(t, g, me.ID, pick)

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
	if got := effectstest.LoyaltyCount(g, jace); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestRFPBJaceMultiverseMinusThreeExilesAnotherAndRevealsUntilACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := effectstest.RfpbWalker(g, me.ID, "Jace, Multiverse Architect", "Legendary Planeswalker — Jace", effectstest.RfpbJaceMultiverseOracle, 4)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
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
	effectstest.B16Activate(t, g, me.ID, jace, 1, rfpbTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(bear) {
		t.Error("the targeted creature must be exiled")
	}
	if _, ok := effectstest.BattlefieldCard(g, prize); !ok {
		t.Error("the revealed creature must enter the battlefield")
	}
	if me.Library.Cards[0].InstanceID != miss1 && me.Library.Cards[0].InstanceID != miss2 {
		t.Error("the revealed misses must go on the bottom of the library")
	}
	if got := effectstest.LoyaltyCount(g, jace); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

// --- Identity Echo -------------------------------------------------

func TestRFPBIdentityEchoExilesYourPermanentAndPutsTheRevealedOneOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	echo := effectstest.PushCatalogPermanent(g, me.ID, "Identity Echo", "Enchantment", effectstest.RfpbIdentityEchoOracle, false)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	prize := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: prize, Name: "Prize", TypeLine: "Creature — Giant", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	miss := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: miss, Name: "Miss", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.RfpbAddMana(me, "R", "R", "R", "R")

	effectstest.B16Activate(t, g, me.ID, echo, 0, rfpbTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(bear) {
		t.Error("the target must be exiled")
	}
	if _, ok := effectstest.BattlefieldCard(g, prize); !ok {
		t.Error("the revealed creature must enter the battlefield")
	}
	if me.Library.Cards[0].InstanceID != miss {
		t.Error("the revealed miss goes to the bottom")
	}
}

func TestRFPBIdentityEchoIsSorceryOnlyAndNeedsYourOwnTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	echo := effectstest.PushCatalogPermanent(g, me.ID, "Identity Echo", "Enchantment", effectstest.RfpbIdentityEchoOracle, false)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.RfpbAddMana(me, "R", "R", "R", "R", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, echo, 0, rfpbTargets(theirs)); err == nil {
		t.Error("Identity Echo exiled an opponent's creature")
	}
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.RfpbAddMana(me, "R", "R", "R", "R", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, echo, 0, rfpbTargets(mine)); err == nil {
		t.Error("Identity Echo activated outside a main phase")
	}
}

// --- Living Library ------------------------------------------------

func TestRFPBLivingLibraryShufflesAnOpponentsCreatureIntoTheirLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lib := effectstest.B12Push(g, me.ID, "Living Library", "Artifact Creature — Book Illusion", effectstest.RfpbLivingLibraryOracle, 0, 4)
	victim := effectstest.B16Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	effectstest.RfpbAddMana(me, "G", "G", "G", "G", "G", "G")
	libBefore := opp.Library.Size()

	effectstest.B16Activate(t, g, me.ID, lib, 0, rfpbTargets(victim))
	effectstest.PassPriorityAroundTable(t, g)

	if _, ok := effectstest.BattlefieldCard(g, lib); ok {
		t.Error("Living Library must be sacrificed")
	}
	if _, ok := effectstest.BattlefieldCard(g, victim); ok || !opp.Library.Contains(victim) {
		t.Error("the creature must be shuffled into its owner's library")
	}
	if got := opp.Library.Size(); got != libBefore+1 {
		t.Errorf("their library %d, want %d", got, libBefore+1)
	}
}

func TestRFPBLivingLibraryRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := effectstest.B12Push(g, me.ID, "Living Library", "Artifact Creature — Book Illusion", effectstest.RfpbLivingLibraryOracle, 0, 4)
	mine := effectstest.B16Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	effectstest.RfpbAddMana(me, "G", "G", "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, lib, 0, rfpbTargets(mine)); err == nil {
		t.Error("Living Library targeted its controller's own creature")
	}
}

// --- Gideon's Memorial ---------------------------------------------

func TestRFPBGideonsMemorialAnthemTouchesOnlyYourCreatureTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gideon's Memorial", "Legendary Artifact", effectstest.RfpbGideonsMemorialOracle, false)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	card := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	var theirs uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.Controller == opp.ID {
			theirs = c.InstanceID
		}
	}
	if p := effectstest.EffectivePower(t, g, goblin); p != 2 {
		t.Errorf("my token power %d, want 1+1", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, goblin, "vigilance") {
		t.Error("my token must have vigilance")
	}
	if p := effectstest.EffectivePower(t, g, card); p != 2 || effectstest.HasEffectiveKeyword(t, g, card, "vigilance") {
		t.Error("a nontoken creature must be untouched")
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 || effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's token must be untouched")
	}
}

func TestRFPBGideonsMemorialManaPaysOnlyForPlaneswalkerSpells(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	mem := effectstest.PushCatalogPermanent(g, me.ID, "Gideon's Memorial", "Legendary Artifact", effectstest.RfpbGideonsMemorialOracle, false)
	if err := g.ActivateManaAbility(me.ID, mem, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap the Memorial: %v", err)
	}
	bear := effectstest.HandSpell(me, "Green Creature", "Creature — Bear", "{G}")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}), "the Memorial's mana for a creature spell")
	walker := effectstest.HandSpell(me, "Green Walker", "Legendary Planeswalker — Test", "{G}")
	if err := g.CastSpell(me.ID, walker, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Memorial's {G} for a planeswalker spell: %v", err)
	}
}

func TestRFPBGideonsMemorialChannelDamagesAnAttackerFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mem := effectstest.HandCardFull(me, "Gideon's Memorial", "Legendary Artifact", "{1}{W}", effectstest.RfpbGideonsMemorialOracle, []string{"W"})
	attacker := effectstest.B16Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 4)
	idle := effectstest.B16Creature(g, me.ID, "Idle", "Creature — Ogre", 3, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("declare attacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.RfpbAddMana(me, "W", "W")

	if err := g.ActivateCatalogAbility(me.ID, mem, 0, rfpbTargets(idle)); err == nil {
		t.Fatal("the channel hit a creature that is neither attacking nor blocking")
	}
	if err := g.ActivateCatalogAbility(me.ID, mem, 0, rfpbTargets(attacker)); err != nil {
		t.Fatalf("channel at the attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, attacker); ok {
		t.Error("4 damage must kill a 3/4 attacker")
	}
	if !me.Graveyard.Contains(mem) {
		t.Error("the Memorial is discarded as a cost")
	}
}

// --- Gideon the Oathless -------------------------------------------

func TestRFPBGideonOathlessPingsThePlayerWhoseCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, opp.ID, "Gideon the Oathless", "Legendary Creature — Human Mercenary", effectstest.RfpbGideonOathlessOracle, 3, 3)
	myLife, oppLife := me.Life, opp.Life

	effectstest.RfpbCastCreature(t, g, "Mine", 2)
	if me.Life != myLife-1 {
		t.Errorf("my life %d, want %d: a creature an opponent of Gideon's controller controls entered", me.Life, myLife-1)
	}
	if opp.Life != oppLife {
		t.Errorf("Gideon's controller lost life: %d", opp.Life)
	}
}

func TestRFPBGideonOathlessIgnoresItsControllersOwnCreaturesAndAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Gideon the Oathless", "Legendary Creature — Human Mercenary", effectstest.RfpbGideonOathlessOracle, 3, 3)
	walker := effectstest.RfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukCurseBreakerOracle, 5)
	life := me.Life
	effectstest.RfpbCastCreature(t, g, "Mine", 2)
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, opp.ID, "Gideon the Oathless", "Legendary Creature — Human Mercenary", effectstest.RfpbGideonOathlessOracle, 3, 3)
	walker := effectstest.RfpbWalker(g, me.ID, "Garruk, Curse Breaker", "Legendary Planeswalker — Garruk", effectstest.RfpbGarrukCurseBreakerOracle, 5)
	before := me.Life
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{}) // +2, makes no creature
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("life %d, want %d: an opponent activated a loyalty ability", me.Life, before-1)
	}
}

func TestRFPBGideonOathlessCarriesItsWard(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.RfpbGideonOathlessOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Gideon the Oathless is complete since ward—discard landed (#3172): %+v", spec.Caveats)
	}
}

// --- Lich's Relic --------------------------------------------------

func TestRFPBLichsRelicGivesPlusTwoPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Lich's Relic", "Artifact — Equipment", effectstest.RfpbLichsRelicOracle, false)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.RfpbAddMana(me, "B", "B")
	effectstest.B16Activate(t, g, me.ID, relic, 0, rfpbTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.RfPT(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("equipped bear is %d/%d, want 4/3", p, tough)
	}
}

func TestRFPBLichsRelicPaidDestroysOnePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var victims []uuid.UUID
	var refs []game.TargetRef
	for _, s := range g.Seats {
		if s.ID != me.ID {
			v := effectstest.B16Creature(g, s.ID, "Bear", "Creature — Bear", 2, 2)
			victims = append(victims, v)
			refs = append(refs, effectstest.RfpbCardRef(v))
		}
	}
	effectstest.RfpbCastRelic(t, g, me, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	// One clause per opponent, asked one prompt at a time.
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestPickTarget(g, me.ID)
		if pick == nil {
			break
		}
		var ref []game.TargetRef
		for _, r := range refs {
			if effectstest.HasID(pick.PickTargetCards, r.ID) {
				ref = []game.TargetRef{r}
			}
		}
		if err := g.ResolvePickTargets(pick.ID, me.ID, ref); err != nil {
			t.Fatalf("resolve the target prompt: %v", err)
		}
	}
	for _, v := range victims {
		if _, ok := effectstest.BattlefieldCard(g, v); ok {
			t.Error("a victim survived the Relic")
		}
	}
}

func TestRFPBLichsRelicTakesOnlyOnePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B16Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B16Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.RfpbCastRelic(t, g, me, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the destroy trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{effectstest.RfpbCardRef(a), effectstest.RfpbCardRef(b)}); err == nil {
		t.Error("two creatures of the same opponent were accepted: one per opponent")
	}
}

func TestRFPBLichsRelicDeclinedPaymentDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.RfpbCastRelic(t, g, me, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); !ok {
		t.Error("the creature died although {2} was not paid")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt appeared although {2} was not paid")
	}
}

// --- Liliana the Faultless / the Repentant -------------------------

func TestRFPBLilianaFaultlessGainsLifeAndGrantsHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lili := effectstest.B12Push(g, me.ID, "Liliana the Faultless", "Legendary Creature — Human Cleric", effectstest.RfpbLilianaFaultlessOracle, 1, 1)
	life := me.Life
	bear := effectstest.RfpbCastCreature(t, g, "Bear", 2)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d: another creature entered", me.Life, life+1)
	}
	effectstest.RfpbAddMana(me, "W")
	pitch := effectstest.HandCardFull(me, "Pitch", "Sorcery", "", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(lili), DiscardIDs: []uuid.UUID{pitch}}); err == nil {
		t.Error("Liliana targeted herself")
	}
	effectstest.B16Activate(t, g, me.ID, lili, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), DiscardIDs: []uuid.UUID{pitch}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the bear must gain hexproof")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard is a cost")
	}
}

func TestRFPBLilianaRepentantMillsAndReanimatesOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lili := effectstest.B12Push(g, me.ID, "Liliana the Repentant", "Legendary Creature — Human Warlock", effectstest.RfpbLilianaRepentantOracle, 2, 2)
	libBefore := me.Library.Size()
	effectstest.RfpbCastCreature(t, g, "Bear", 2)
	if got := me.Library.Size(); got != libBefore-2 {
		t.Errorf("library %d, want %d: mill two", got, libBefore-2)
	}
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen", TypeLine: "Creature — Giant", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.RfpbAddMana(me, "B", "B", "B", "B", "B", "B")
	effectstest.B16Activate(t, g, me.ID, lili, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, dead); !ok {
		t.Error("the creature must return to the battlefield")
	}
	if p, _ := effectstest.RfPT(t, g, lili); p != 3 {
		t.Errorf("Liliana has power %d, want 2+1", p)
	}
	effectstest.RfpbAddMana(me, "B", "B", "B", "B", "B", "B")
	dead2 := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead2, Name: "Fallen Two", TypeLine: "Creature — Giant", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead2)}); err == nil {
		t.Error("an exhaust ability activated twice")
	}
}

func TestRFPBLilianaRepentantRefusesAnOpponentsGraveyardAndALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lili := effectstest.B12Push(g, me.ID, "Liliana the Repentant", "Legendary Creature — Human Warlock", effectstest.RfpbLilianaRepentantOracle, 2, 2)
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Creature — Giant", Owner: opp.ID, Controller: opp.ID})
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
	effectstest.RfpbAddMana(me, "B", "B", "B", "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("reanimated a card from an opponent's graveyard")
	}
	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)}); err == nil {
		t.Error("reanimated a land card")
	}
}

// --- Mabel ---------------------------------------------------------

func rfpbCastMabel(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	mabel := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mabel, Name: "Mabel, Bitter Recluse", TypeLine: "Legendary Creature — Mouse Warlock",
		OracleID: effectstest.RfpbMabelOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, mabel, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Mabel: %v", err)
	}
	rfpbSettle(t, g, me.ID, effectstest.RfpbCardRef(target))
}

func TestRFPBMabelRemovesUpToThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B16Creature(g, opp.ID, "Hydra", "Creature — Hydra", 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(target, game.CounterPlusOne, 5) })
	rfpbCastMabel(t, g, me, target)
	for i := 0; i < 3; i++ {
		effectstest.AnswerOptionPick(t, g, me.ID, 1)
	}
	c, _ := effectstest.BattlefieldCard(g, target)
	if got := c.Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("counters left %d, want 5-3", got)
	}
}

func TestRFPBMabelMayStopEarlyAndHasDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B16Creature(g, opp.ID, "Hydra", "Creature — Hydra", 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(target, game.CounterPlusOne, 5) })
	rfpbCastMabel(t, g, me, target)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	c, _ := effectstest.BattlefieldCard(g, target)
	if got := c.Counters[game.CounterPlusOne]; got != 4 {
		t.Errorf("counters left %d, want 4 after stopping", got)
	}
	mabel := effectstest.FindBattlefieldByName(g, "Mabel, Bitter Recluse")
	if mabel == uuid.Nil || !effectstest.HasEffectiveKeyword(t, g, mabel, "deathtouch") {
		t.Error("Mabel must have deathtouch")
	}
}

// --- Massacre Girl, Most Wanted ------------------------------------

func TestRFPBMassacreGirlDrainsWhenYourCreatureOrPlaneswalkerDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Massacre Girl, Most Wanted", "Legendary Creature — Human Assassin", effectstest.RfpbMassacreGirlOracle, 4, 4)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	walker := effectstest.RfpbWalker(g, me.ID, "Doomed Walker", rfpbWalkerLine, "", 3)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	girl := effectstest.B12Push(g, me.ID, "Massacre Girl, Most Wanted", "Legendary Creature — Human Assassin", effectstest.RfpbMassacreGirlOracle, 4, 4)
	src := effectstest.B16Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.RfPT(t, g, girl); p != 5 {
		t.Errorf("power %d, want 5 after noncombat damage to an opponent", p)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.RfPT(t, g, girl); p != 5 {
		t.Errorf("power %d: damage to her own controller must not grow her", p)
	}
}
