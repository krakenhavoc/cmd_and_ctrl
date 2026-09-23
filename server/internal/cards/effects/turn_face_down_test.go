package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_face_down_test.go — #1209's four proof cards. One test per
// card, each asserting the half of the seam only that card reaches:
//
//	Backslide           the targeted turn, and the way back UP again
//	Cyber Conversion    the turn with no way back up (CR 708.7)
//	Ixidron             the batch, and a size that counts what it hid
//	Master of the Veil  the loop: morph → turned face up → turn one down

// morphCardOnBattlefield seats a creature whose CARD prints morph, so
// the CR 702.37e half of the seam has something to read. The morph
// cast itself is #1194's and is tested there; what matters here is
// that the DECLARATION is on the card while the permanent is face up.
func morphCardOnBattlefield(g *game.Game, owner uuid.UUID, name, oracleID string) uuid.UUID {
	id := uuid.New()
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Human Wizard",
		OracleID:   oracleID,
		ManaCost:   "{1}{U}",
		Power:      1,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

func faceDownCardForTest(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("permanent %s is not on the battlefield", id)
	return game.Card{}
}

// TestBackslideTurnsAMorphBackDownAndItCanComeBackUp is the whole
// point of the card. Willbender is registered with Morph {1}{U}
// (#1194), so it is a legal target for "target creature with a morph
// ability" and, once face down, CR 702.37e offers it its own morph
// cost again — which is what makes Backslide a combat trick rather
// than a Pacifism.
func TestBackslideTurnsAMorphBackDownAndItCanComeBackUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// Willbender's registered oracle ID — it prints Morph {1}{U}.
	victim := morphCardOnBattlefield(g, me.ID, "Willbender", "0aae277e-e58e-4115-b5fd-0459451e17ec")

	castCatalogSpell(t, g, "Backslide", "Instant",
		"6d9746bc-0b72-4ac4-b48a-742b63b1c41b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	c := faceDownCardForTest(t, g, victim)
	if !c.FaceDown || c.FaceDownKind != game.FaceDownTurned {
		t.Fatalf("face-down state = (%v, %q), want (true, %q)", c.FaceDown, c.FaceDownKind, game.FaceDownTurned)
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("the permanent is %d/%d, want the CR 708.2a 2/2",
			c.CurrentPower(), c.CurrentToughness())
	}
	offer := game.TurnFaceUpOffer(c)
	if offer == nil {
		t.Fatal("no way back up — CR 702.37e offers a morph card its own morph cost, " +
			"whatever put the permanent face down")
	}
	if offer.Cost != "{1}{U}" {
		t.Errorf("turn-face-up cost = %q, want Willbender's morph cost {1}{U}", offer.Cost)
	}
}

// TestBackslideRefusesACreatureWithoutAMorphAbility is the clause,
// enforced at announce (CR 601.2c). A creature with no morph is not a
// legal target, so the cast is refused rather than resolving into
// nothing.
func TestBackslideRefusesACreatureWithoutAMorphAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vanilla := pushCreatureToBattlefieldForTest(g, me.ID, "Grizzly Bears")

	err := castCatalogSpellErr(t, g, "Backslide", "Instant",
		"6d9746bc-0b72-4ac4-b48a-742b63b1c41b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: vanilla}})
	if err == nil {
		t.Fatal("a creature with no morph ability was a legal target for Backslide")
	}
	if faceDownCardForTest(t, g, vanilla).FaceDown {
		t.Error("the refused cast turned the creature face down anyway")
	}
}

// TestCyberConversionHidesACreatureForGood is CR 708.7 with nothing
// on the other side of it: no keyword put this permanent face down
// and no keyword takes it back up.
func TestCyberConversionHidesACreatureForGood(t *testing.T) {
	g := newCatalogGame(t)
	victim := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Sheoldred")

	castCatalogSpell(t, g, "Cyber Conversion", "Instant",
		"33761c1a-9848-45bf-b934-123cebab566b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	c := faceDownCardForTest(t, g, victim)
	if !c.FaceDown || c.FaceDownKind != game.FaceDownTurned {
		t.Fatalf("face-down state = (%v, %q), want (true, %q)", c.FaceDown, c.FaceDownKind, game.FaceDownTurned)
	}
	if c.Controller != g.Seats[1].ID {
		t.Error("turning a creature face down changed who controls it")
	}
	if !c.IsKnownTo(g.Seats[1].ID) {
		t.Error("its controller may not look at it (CR 708.5)")
	}
	if c.IsKnownTo(g.Seats[0].ID) {
		t.Error("the caster may look at a face-down permanent they do not control (CR 708.5)")
	}
	if game.TurnFaceUpOffer(c) != nil {
		t.Error("Cyber Conversion's victim can be turned face up — CR 708.7 gives it no way back")
	}
}

// TestIxidronHidesEveryOtherNontokenCreatureAndSizesItself is the
// batch form and the count, which are the same sentence read twice:
// Ixidron is as big as the board it just hid.
func TestIxidronHidesEveryOtherNontokenCreatureAndSizesItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	theirs := pushCreatureToBattlefieldForTest(g, opp.ID, "Theirs")
	tokenID := uuid.New()
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: tokenID,
		Name:       "Zombie",
		TypeLine:   "Token Creature — Zombie",
		Power:      2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	self := castCatalogSpell(t, g, "Ixidron", "Creature — Illusion",
		"a86587ea-14ec-45db-8bac-10b6b9eaeb25", nil)
	passPriorityAroundTable(t, g)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"a creature you control", mine, true},
		{"a creature an opponent controls — \"all OTHER\" is not \"all yours\"", theirs, true},
		{"a token — Ixidron says nontoken", tokenID, false},
		{"Ixidron itself — \"all OTHER\"", self, false},
	} {
		if got := faceDownCardForTest(t, g, tc.id).FaceDown; got != tc.want {
			t.Errorf("%s: face down = %v, want %v", tc.name, got, tc.want)
		}
	}

	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	ix := faceDownCardForTest(t, g, self)
	if ix.CurrentPower() != 2 || ix.CurrentToughness() != 2 {
		t.Errorf("Ixidron is %d/%d, want 2/2 — the two creatures it just hid",
			ix.CurrentPower(), ix.CurrentToughness())
	}
}

// TestIxidronAloneIsAZeroZero is the same characteristic-defining
// ability read the other way, and it is the printed behaviour a
// hard-coded P/T would have got wrong: with nothing else on the
// board there is nothing face down to count.
func TestIxidronAloneIsAZeroZero(t *testing.T) {
	g := newCatalogGame(t)
	self := castCatalogSpell(t, g, "Ixidron", "Creature — Illusion",
		"a86587ea-14ec-45db-8bac-10b6b9eaeb25", nil)
	passPriorityAroundTable(t, g)

	// CR 704.5f will have swept it into the graveyard by now, which
	// is itself the assertion: a 0/0 with no counters dies.
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == self {
			t.Fatalf("Ixidron survived on an empty board as a %d/%d, want a 0/0 that dies",
				c.CurrentPower(), c.CurrentToughness())
		}
	}
}

// TestMasterOfTheVeilTurnsAMorphFaceDownWhenItIsTurnedFaceUp joins
// #1194's direction to #1209's. The trigger is ordinary — it watches
// EventTurnedFaceUp, it is optional and it targets — and the body is
// the new primitive; nothing in the card file knows that the two
// halves were built a seam apart.
func TestMasterOfTheVeilTurnsAMorphFaceDownWhenItIsTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	master := uuid.New()
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: master,
		Name:       "Master of the Veil",
		TypeLine:   "Creature — Human Wizard",
		OracleID:   "ee9104d5-3583-4549-8dd0-5b69274f4b4f",
		ManaCost:   "{2}{U}{U}",
		Power:      2, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	victim := morphCardOnBattlefield(g, me.ID, "Willbender", "0aae277e-e58e-4115-b5fd-0459451e17ec")

	// Put the Master face down the way an effect would, then turn it
	// face up for its morph cost — which is the event its own trigger
	// watches. CR 702.37e is what makes that legal from the `turned`
	// state at all.
	g.WithWriteLock(func() {
		g.TurnFaceDownForEffect(uuid.Nil, master)
		for i := 0; i < 8; i++ {
			g.Seats[g.Turn.ActiveSeat].ManaPool.AddMana(game.ManaToken{Color: "U"})
		}
	})
	if err := g.PerformSpecialAction(me.ID, master, game.SpecialActionTurnFaceUp,
		game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn the Master face up: %v", err)
	}

	// "You MAY turn" is the CR 603.3c optional-trigger prompt; the
	// target is picked only after it is accepted (CR 603.3d).
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if pick == nil || !hasID(pick.PickTargetCards, victim) {
		t.Fatalf("the morph creature is offered as a target: %+v", pick)
	}
	// The Master itself is the other legal target — its own card
	// prints morph — which is the loop the card is built around.
	if !hasID(pick.PickTargetCards, master) {
		t.Error("the Master is not offered itself; CR 702.37e keys on the card, not on the state")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if faceDownCardForTest(t, g, master).FaceDown {
		t.Fatal("the Master is still face down")
	}
	if !faceDownCardForTest(t, g, victim).FaceDown {
		t.Error("the trigger did not turn the targeted morph face down")
	}
	if k := faceDownCardForTest(t, g, victim).FaceDownKind; k != game.FaceDownTurned {
		t.Errorf("kind = %q, want %q — an effect turned it over, not a cast", k, game.FaceDownTurned)
	}
}
