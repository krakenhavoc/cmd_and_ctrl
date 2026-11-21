package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_face_up_cards_test.go — #2590 / ADR 0082's second 2026-10-07
// amendment through the catalog: Hauntwoods Shrieker, Zimone, Mystery
// Unraveler and Staff Room, the three cards that turn a face-down
// permanent face up as an effect. The engine door
// (game.TurnFaceUpForEffect) is tested in game/face_down_turn_up_effect_test.go.

const (
	tfuShriekerOracle = "09c275bd-8323-485e-81f6-aff1f99d0d03"
	tfuZimoneOracle   = "75bf0f27-9df1-4ab4-97a8-94bb3223bdf9"
)

func islandCard() game.Card {
	return game.Card{Name: "Sleeping Island", TypeLine: "Basic Land — Island"}
}

func activateShrieker(t *testing.T, g *game.Game, me *game.Player, shrieker, target uuid.UUID) error {
	t.Helper()
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("G", "C")...) })
	return g.ActivateCatalogAbility(me.ID, shrieker, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
}

func turnedFaceUpEvent(g *game.Game, id uuid.UUID) *game.Event {
	for i := range g.Events {
		if g.Events[i].Kind == game.EventTurnedFaceUp && g.Events[i].CardID == id {
			return &g.Events[i]
		}
	}
	return nil
}

// --- Hauntwoods Shrieker ------------------------------------------------

func TestHauntwoodsShriekerAttackManifestsDread(t *testing.T) {
	g, me, _, second := effectstest.DreadTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	shrieker := effectstest.PushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, shrieker)
	effectstest.AnswerDread(t, g, me, second)
	effectstest.RequireManifested(t, g, me, second)
}

func TestHauntwoodsShriekerRevealsAndTurnsAnOpponentsCreatureCardFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	shrieker := effectstest.PushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	target := effectstest.ManifestTop(t, g, opp, effectstest.BearCard())
	mana := len(me.ManaPool)

	if err := activateShrieker(t, g, me, shrieker, target); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, target)
	if !c.FaceDown {
		t.Fatal("the permanent turned up before the controller was asked")
	}
	for _, p := range g.Seats {
		if !c.IsKnownTo(p.ID) {
			t.Errorf("seat %s was not shown the revealed card", p.Name)
		}
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)

	c = effectstest.FindBattlefieldCardForTest(g, target)
	if c.FaceDown || c.Name != "Sleeping Bear" {
		t.Fatalf("the creature card is not face up: face_down=%v name=%q", c.FaceDown, c.Name)
	}
	if c.Controller != opp.ID {
		t.Error("turning an opponent's permanent face up changed its controller")
	}
	if ev := turnedFaceUpEvent(g, target); ev == nil || ev.Actor != me.ID || ev.Source != shrieker {
		t.Errorf("event = %+v, want actor %v and source %v", ev, me.ID, shrieker)
	}
	// The activation cost {1}{G} was the whole price: the turn itself is free.
	if len(me.ManaPool)-mana != 0 {
		t.Errorf("pool changed by %d beyond the activation's own cost", len(me.ManaPool)-mana)
	}
}

func TestHauntwoodsShriekerMayDeclineToTurnItUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	shrieker := effectstest.PushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	target := effectstest.ManifestTop(t, g, me, effectstest.BearCard())
	if err := activateShrieker(t, g, me, shrieker, target); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if !effectstest.FindBattlefieldCardForTest(g, target).FaceDown {
		t.Error("declining still turned it face up")
	}
	if turnedFaceUpEvent(g, target) != nil {
		t.Error("declining emitted EventTurnedFaceUp")
	}
}

// "If it's a creature card" asks the CARD, not the 2/2 body: a manifested
// Island is revealed and stays face down, and nobody is asked anything.
func TestHauntwoodsShriekerRevealsANoncreatureCardAndLeavesItFaceDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	shrieker := effectstest.PushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	target := effectstest.ManifestTop(t, g, opp, islandCard())
	if effectstest.FindBattlefieldCardForTest(g, target).IsKnownTo(me.ID) {
		t.Fatal("fixture: the Shrieker's controller already knew the opponent's manifest")
	}
	if err := activateShrieker(t, g, me, shrieker, target); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked to turn a noncreature card face up")
	}
	c := effectstest.FindBattlefieldCardForTest(g, target)
	if !c.FaceDown {
		t.Error("a manifested Island was turned face up (CR 701.40b)")
	}
	if !c.IsKnownTo(me.ID) {
		t.Error("the reveal did not show the Shrieker's controller the card")
	}
}

func TestHauntwoodsShriekerCannotTargetAFaceUpPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	shrieker := effectstest.PushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	faceUp := effectstest.PushVanillaCreature(g, me.ID, "Face Up Bear", 2, 2)
	err := activateShrieker(t, g, me, shrieker, faceUp)
	if err == nil {
		t.Fatal("a face-up permanent was accepted as a face-down target")
	}
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Logf("refused with %v", err)
	}
}

// --- Zimone, Mystery Unraveler ------------------------------------------

func TestZimoneManifestsDreadOnTheFirstLandfallAndTurnsAPermanentUpOnTheNext(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	effectstest.PushDiesCreatureForTest(g, me.ID, "Zimone, Mystery Unraveler", tfuZimoneOracle, "Legendary Creature — Human Wizard", 3, 3)

	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AnswerDread(t, g, me, top)
	effectstest.RequireManifested(t, g, me, top)
	if !effectstest.MdInGraveyard(me, second) {
		t.Fatal("the other card is not in the graveyard")
	}

	// The second landfall this turn: no manifest, a "you may turn one face up".
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil {
		t.Fatal("the second landfall did not offer to turn a permanent face up")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, top)
	c := effectstest.FindBattlefieldCardForTest(g, top)
	if c.FaceDown || c.Name != "Top Bear" {
		t.Errorf("the chosen permanent is not face up: face_down=%v name=%q", c.FaceDown, c.Name)
	}
	if ev := turnedFaceUpEvent(g, top); ev == nil || ev.Actor != me.ID {
		t.Errorf("event = %+v, want it actored by %v", ev, me.ID)
	}
}

func TestZimoneThirdLandfallStillDoesNotManifestAgain(t *testing.T) {
	g, me, top, _ := effectstest.DreadTable(t)
	effectstest.PushDiesCreatureForTest(g, me.ID, "Zimone, Mystery Unraveler", tfuZimoneOracle, "Legendary Creature — Human Wizard", 3, 3)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AnswerDread(t, g, me, top)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID) // "you may": take none
	library := me.Library.Size()
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != library {
		t.Error("the third landfall looked at the library again")
	}
}

// With nothing it could turn over, the "otherwise" branch asks nothing:
// a manifested Island stays face down (CR 701.40b) and is not offered.
func TestZimoneDoesNotOfferAManifestedNoncreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushDiesCreatureForTest(g, me.ID, "Zimone, Mystery Unraveler", tfuZimoneOracle, "Legendary Creature — Human Wizard", 3, 3)
	island := effectstest.ManifestTop(t, g, me, islandCard())
	effectstest.PlayLandFromHand(t, g, "Forest", "") // first landfall: manifest dread
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		// Manifest dread's prompt: manifest the first card offered.
		effectstest.AnswerChooseCards(t, g, me.ID, pick.ChooseCards[0])
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "") // second landfall: turn one face up
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceOwnPermanents {
			for _, id := range c.ChooseCards {
				if id == island {
					t.Error("a manifested Island was offered to be turned face up")
				}
			}
		}
	}
	if !effectstest.FindBattlefieldCardForTest(g, island).FaceDown {
		t.Error("the Island was turned face up")
	}
}

// --- Staff Room ---------------------------------------------------------

func staffRoomRoom(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, effectstest.ExpLabOracle, "Experimental Lab", "{3}{G}", "Staff Room", "{2}{G}"), 1)
	effectstest.RoomsBSettle(t, g, me)
}

func TestStaffRoomOffersToTurnAFaceDownCreatureFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.PushDiesCreatureForTest(g, me.ID, "Hidden Bear", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, id) })
	staffRoomRoom(t, g, me)
	effectstest.AttackWith(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c.FaceDown {
		t.Error("choosing to turn it face up left it face down")
	}
	if c.Counters[game.CounterPlusOne] != 0 {
		t.Error("it was turned face up AND got a counter")
	}
	if ev := turnedFaceUpEvent(g, id); ev == nil || ev.Actor != me.ID {
		t.Errorf("event = %+v", ev)
	}
}

func TestStaffRoomCanPutTheCounterOnAFaceDownCreatureInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.PushDiesCreatureForTest(g, me.ID, "Hidden Bear", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, id) })
	staffRoomRoom(t, g, me)
	effectstest.AttackWith(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.FindBattlefieldCardForTest(g, id)
	if !c.FaceDown {
		t.Error("choosing the counter turned it face up")
	}
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("counters = %d, want 1", c.Counters[game.CounterPlusOne])
	}
}

func TestStaffRoomAsksNothingOfAFaceUpCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.PushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	staffRoomRoom(t, g, me)
	effectstest.AttackWith(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked a choice with only one option")
	}
	if effectstest.FindBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne] != 1 {
		t.Error("the face-up creature got no counter")
	}
}

func TestTheDeclaredCardsAreNoLongerCaveated(t *testing.T) {
	for _, id := range []string{tfuShriekerOracle, tfuZimoneOracle, effectstest.ExpLabOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want full", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}
