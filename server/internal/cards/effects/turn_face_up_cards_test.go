package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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

// manifestTop puts `c` on top of p's library and manifests it, returning
// the face-down permanent (controlled by p).
func manifestTop(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		if id, err = g.ManifestForEffect(p.ID); err != nil {
			t.Fatalf("ManifestForEffect: %v", err)
		}
	})
	if findBattlefieldCardForTest(g, id) == nil {
		t.Fatal("nothing was manifested")
	}
	return id
}

func bearCard() game.Card {
	return game.Card{Name: "Sleeping Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
}

func islandCard() game.Card {
	return game.Card{Name: "Sleeping Island", TypeLine: "Basic Land — Island"}
}

func activateShrieker(t *testing.T, g *game.Game, me *game.Player, shrieker, target uuid.UUID) error {
	t.Helper()
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("G", "C")...) })
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

// answerOwnPermanents answers the open "choose from your permanents"
// prompt (ChoosePermanents on the chooser's own board).
func answerOwnPermanents(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, chooser)
	if c == nil {
		t.Fatal("no own-permanents prompt is open")
	}
	if err := g.ResolveOwnPermanents(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}

// --- Hauntwoods Shrieker ------------------------------------------------

func TestHauntwoodsShriekerAttackManifestsDread(t *testing.T) {
	g, me, _, second := dreadTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	shrieker := pushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	declareAttack(t, g, opp.ID, shrieker)
	answerDread(t, g, me, second)
	requireManifested(t, g, me, second)
}

func TestHauntwoodsShriekerRevealsAndTurnsAnOpponentsCreatureCardFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	shrieker := pushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	target := manifestTop(t, g, opp, bearCard())
	mana := len(me.ManaPool)

	if err := activateShrieker(t, g, me, shrieker, target); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardForTest(g, target)
	if !c.FaceDown {
		t.Fatal("the permanent turned up before the controller was asked")
	}
	for _, p := range g.Seats {
		if !c.IsKnownTo(p.ID) {
			t.Errorf("seat %s was not shown the revealed card", p.Name)
		}
	}
	answerMayChoice(t, g, me.ID, true)

	c = findBattlefieldCardForTest(g, target)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	shrieker := pushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	target := manifestTop(t, g, me, bearCard())
	if err := activateShrieker(t, g, me, shrieker, target); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if !findBattlefieldCardForTest(g, target).FaceDown {
		t.Error("declining still turned it face up")
	}
	if turnedFaceUpEvent(g, target) != nil {
		t.Error("declining emitted EventTurnedFaceUp")
	}
}

// "If it's a creature card" asks the CARD, not the 2/2 body: a manifested
// Island is revealed and stays face down, and nobody is asked anything.
func TestHauntwoodsShriekerRevealsANoncreatureCardAndLeavesItFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	shrieker := pushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	target := manifestTop(t, g, opp, islandCard())
	if findBattlefieldCardForTest(g, target).IsKnownTo(me.ID) {
		t.Fatal("fixture: the Shrieker's controller already knew the opponent's manifest")
	}
	if err := activateShrieker(t, g, me, shrieker, target); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)

	if latestConfirmFor(g, me.ID) != nil {
		t.Error("asked to turn a noncreature card face up")
	}
	c := findBattlefieldCardForTest(g, target)
	if !c.FaceDown {
		t.Error("a manifested Island was turned face up (CR 701.40b)")
	}
	if !c.IsKnownTo(me.ID) {
		t.Error("the reveal did not show the Shrieker's controller the card")
	}
}

func TestHauntwoodsShriekerCannotTargetAFaceUpPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	shrieker := pushDiesCreatureForTest(g, me.ID, "Hauntwoods Shrieker", tfuShriekerOracle, "Creature — Beast Mutant", 3, 3)
	faceUp := pushVanillaCreature(g, me.ID, "Face Up Bear", 2, 2)
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
	g, me, top, second := dreadTable(t)
	pushDiesCreatureForTest(g, me.ID, "Zimone, Mystery Unraveler", tfuZimoneOracle, "Legendary Creature — Human Wizard", 3, 3)

	playLandFromHand(t, g, "Forest", "")
	answerDread(t, g, me, top)
	requireManifested(t, g, me, top)
	if !mdInGraveyard(me, second) {
		t.Fatal("the other card is not in the graveyard")
	}

	// The second landfall this turn: no manifest, a "you may turn one face up".
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil {
		t.Fatal("the second landfall did not offer to turn a permanent face up")
	}
	answerOwnPermanents(t, g, me.ID, top)
	c := findBattlefieldCardForTest(g, top)
	if c.FaceDown || c.Name != "Top Bear" {
		t.Errorf("the chosen permanent is not face up: face_down=%v name=%q", c.FaceDown, c.Name)
	}
	if ev := turnedFaceUpEvent(g, top); ev == nil || ev.Actor != me.ID {
		t.Errorf("event = %+v, want it actored by %v", ev, me.ID)
	}
}

func TestZimoneThirdLandfallStillDoesNotManifestAgain(t *testing.T) {
	g, me, top, _ := dreadTable(t)
	pushDiesCreatureForTest(g, me.ID, "Zimone, Mystery Unraveler", tfuZimoneOracle, "Legendary Creature — Human Wizard", 3, 3)
	playLandFromHand(t, g, "Forest", "")
	answerDread(t, g, me, top)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID) // "you may": take none
	library := me.Library.Size()
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if me.Library.Size() != library {
		t.Error("the third landfall looked at the library again")
	}
}

// With nothing it could turn over, the "otherwise" branch asks nothing:
// a manifested Island stays face down (CR 701.40b) and is not offered.
func TestZimoneDoesNotOfferAManifestedNoncreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushDiesCreatureForTest(g, me.ID, "Zimone, Mystery Unraveler", tfuZimoneOracle, "Legendary Creature — Human Wizard", 3, 3)
	island := manifestTop(t, g, me, islandCard())
	playLandFromHand(t, g, "Forest", "") // first landfall: manifest dread
	passPriorityAroundTable(t, g)
	if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
		// Manifest dread's prompt: manifest the first card offered.
		answerChooseCards(t, g, me.ID, pick.ChooseCards[0])
		passPriorityAroundTable(t, g)
	}
	playLandFromHand(t, g, "Forest", "") // second landfall: turn one face up
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceOwnPermanents {
			for _, id := range c.ChooseCards {
				if id == island {
					t.Error("a manifested Island was offered to be turned face up")
				}
			}
		}
	}
	if !findBattlefieldCardForTest(g, island).FaceDown {
		t.Error("the Island was turned face up")
	}
}

// --- Staff Room ---------------------------------------------------------

func staffRoomRoom(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	roomsBCast(t, g, me, roomsBCard(me.ID, expLabOracle, "Experimental Lab", "{3}{G}", "Staff Room", "{2}{G}"), 1)
	roomsBSettle(t, g, me)
}

func TestStaffRoomOffersToTurnAFaceDownCreatureFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := pushDiesCreatureForTest(g, me.ID, "Hidden Bear", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, id) })
	staffRoomRoom(t, g, me)
	attackWith(t, g, opp.ID, id)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	c := findBattlefieldCardForTest(g, id)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := pushDiesCreatureForTest(g, me.ID, "Hidden Bear", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, id) })
	staffRoomRoom(t, g, me)
	attackWith(t, g, opp.ID, id)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	c := findBattlefieldCardForTest(g, id)
	if !c.FaceDown {
		t.Error("choosing the counter turned it face up")
	}
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("counters = %d, want 1", c.Counters[game.CounterPlusOne])
	}
}

func TestStaffRoomAsksNothingOfAFaceUpCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := pushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	staffRoomRoom(t, g, me)
	attackWith(t, g, opp.ID, id)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("asked a choice with only one option")
	}
	if findBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne] != 1 {
		t.Error("the face-up creature got no counter")
	}
}

func TestTheDeclaredCardsAreNoLongerCaveated(t *testing.T) {
	for _, id := range []string{tfuShriekerOracle, tfuZimoneOracle, expLabOracle} {
		spec, ok := Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want full", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}
