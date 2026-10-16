package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// outside_play_right_test.go — #3955, Wish: "you may play a card you own
// from outside the game this turn" is a standing, one-use,
// until-end-of-turn permission over the container. The card is chosen when
// it is played, at the timing it allows; a database pool names it through
// the `outside_play` action, and a cancelled pick spends nothing.

func costedCollection(t *testing.T) {
	t.Helper()
	bear := collectionCard("Costly Bear", "Creature — Bear")
	bear.ManaCost = "{2}"
	forest := collectionCard("Forest", "Basic Land — Forest")
	bolt := collectionCard("Quick Bolt", "Instant")
	bolt.ManaCost = "{1}"
	withCollection(t, bear, forest, bolt)
}

func grantRight(g *Game, p *Player) bool {
	var ok bool
	g.WithWriteLock(func() { ok = g.GrantOutsidePlayRightForEffect(p.ID, uuid.New(), "Wish") })
	return ok
}

func rightLive(g *Game, p *Player) bool {
	var live bool
	g.WithWriteLock(func() { live = g.OutsidePlayRightLocked(p.ID) != nil })
	return live
}

// Not a card chosen when Wish resolves: the permission covers every card
// in the container, and the FIRST one played spends it.
func TestTheWishRightIsSpentByTheFirstCardPlayed(t *testing.T) {
	g := outsideGameInMain(t, withCost(sbCard("First", "Instant"), "{1}"), withCost(sbCard("Second", "Instant"), "{1}"))
	me := g.Seats[0]
	first, second := me.SideboardCards()[0].InstanceID, me.SideboardCards()[1].InstanceID
	if !grantRight(g, me) {
		t.Fatal("the right was not stored")
	}
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	// Either card may be the one played: the choice is made now.
	if err := castFromOutsideStrict(g, me, second); err != nil {
		t.Fatalf("play the second card: %v", err)
	}
	if rightLive(g, me) {
		t.Error("the right survived the card it was spent on")
	}
	if err := castFromOutsideStrict(g, me, first); err == nil {
		t.Error("a second card was played under a one-use right")
	}
	if !me.Sideboard.Contains(first) {
		t.Error("the unplayed card left the container")
	}
}

func castFromOutsideStrict(g *Game, p *Player, id uuid.UUID) error {
	return g.CastSpell(p.ID, id, CastSpellParams{FromZone: "outside", Strict: true})
}

// A land play spends the right too, and is the turn's land drop.
func TestAWishedLandSpendsTheRight(t *testing.T) {
	g := outsideGameInMain(t, outsideLand("Forest"), withCost(sbCard("Spare", "Instant"), "{1}"))
	me := g.Seats[0]
	forest, spare := me.SideboardCards()[0].InstanceID, me.SideboardCards()[1].InstanceID
	if !grantRight(g, me) {
		t.Fatal("the right was not stored")
	}
	if err := castFromOutsideStrict(g, me, forest); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	if rightLive(g, me) {
		t.Error("playing a land did not spend the right")
	}
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	if err := castFromOutsideStrict(g, me, spare); err == nil {
		t.Error("a spell was cast under a right a land had spent")
	}
}

func TestTheWishRightIsHeldByTheCasterOfTheTurnOnly(t *testing.T) {
	g := outsideGameInMain(t, outsideLand("Forest"))
	me, opp := g.Seats[0], g.Seats[1]
	var ok bool
	g.WithWriteLock(func() {
		g.PlayerControls = append(g.PlayerControls, PlayerControl{
			Controller: opp.ID, Controlled: me.ID, Until: ControlUntilTurn, Stage: ControlTurn, TurnSeq: g.Turn.Seq,
		})
		ok = g.GrantOutsidePlayRightForEffect(me.ID, uuid.New(), "Wish")
	})
	if ok || rightLive(g, me) {
		t.Error("a controlled player was granted the right")
	}
}

// The database pool names its card NOW: the action opens the pick, a
// declined pick spends nothing, and the named card is minted where the
// right already covers it and is played at once.
func TestNamingAWishedCardInTheDatabasePool(t *testing.T) {
	costedCollection(t)
	g := newSideboardGame(t) // no sideboard: the database is the pool
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	if err := g.NameOutsideCardToPlay(me.ID); !errors.Is(err, ErrNoPlayPermission) {
		t.Fatalf("naming a card with no right: %v, want ErrNoPlayPermission", err)
	}
	if !grantRight(g, me) {
		t.Fatal("the right was not stored")
	}
	var can bool
	g.WithWriteLock(func() { can = g.CanNameOutsideCardLocked(me.ID) })
	if !can {
		t.Fatal("a database pool holding the right cannot name a card")
	}

	if err := g.NameOutsideCardToPlay(me.ID); err != nil {
		t.Fatalf("NameOutsideCardToPlay: %v", err)
	}
	pick := g.PendingChoices[len(g.PendingChoices)-1]
	if pick.Kind != PendingChoiceOutsideCollection || pick.ChooseDestination != ChoosePlayFromOutside ||
		pick.ChooseMin != 0 || pick.ChooseMax != 1 || pick.Chooser != me.ID {
		t.Fatalf("pick %+v, want a may-name-one over the database, destination play_outside", pick)
	}
	// Cancelled: nothing minted, nothing spent.
	if err := g.ResolveOutsideCollection(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !rightLive(g, me) || len(me.SideboardCards()) != 0 {
		t.Fatalf("a cancelled pick spent the right (live %v) or minted a card (%d)", rightLive(g, me), len(me.SideboardCards()))
	}

	// Named: minted, castable under the right, and the cast spends it.
	if err := g.NameOutsideCardToPlay(me.ID); err != nil {
		t.Fatal(err)
	}
	pick = g.PendingChoices[len(g.PendingChoices)-1]
	if err := g.ResolveOutsideCollection(pick.ID, me.ID, []string{"Costly Bear"}); err != nil {
		t.Fatalf("name the card: %v", err)
	}
	bear := me.SideboardCards()[0]
	if bear.Name != "Costly Bear" || bear.Owner != me.ID {
		t.Fatalf("minted %+v", bear)
	}
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := castFromOutsideStrict(g, me, bear.InstanceID); err != nil {
		t.Fatalf("play the named card: %v", err)
	}
	if rightLive(g, me) || !g.Stack.Contains(bear.InstanceID) {
		t.Error("the named card was not cast, or the right survived it")
	}
}

// A sideboard pool needs no naming: its cards carry the right directly, so
// the action is refused (and never offered).
func TestASideboardPoolCastsUnderTheRightWithoutNaming(t *testing.T) {
	costedCollection(t)
	g := outsideGameInMain(t, outsideLand("Forest"))
	me := g.Seats[0]
	if !grantRight(g, me) {
		t.Fatal("the right was not stored")
	}
	if err := g.NameOutsideCardToPlay(me.ID); !errors.Is(err, ErrNoPlayPermission) {
		t.Errorf("a sideboard pool was asked to name a card: %v", err)
	}
	if err := castFromOutsideStrict(g, me, me.SideboardCards()[0].InstanceID); err != nil {
		t.Errorf("play from the sideboard under the right: %v", err)
	}
}

// The right ends with the turn, uncast.
func TestTheWishRightEndsWithTheTurnAndSpendsNothingUnused(t *testing.T) {
	costedCollection(t)
	g := newSideboardGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	if !grantRight(g, me) {
		t.Fatal("the right was not stored")
	}
	seq := g.Turn.Seq
	for i := 0; g.Turn.Seq == seq; i++ {
		if i > 100 {
			t.Fatal("the turn never ended")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if rightLive(g, me) {
		t.Error("the right outlived the turn")
	}
	if err := g.NameOutsideCardToPlay(me.ID); !errors.Is(err, ErrNoPlayPermission) {
		t.Errorf("naming a card after the turn ended: %v", err)
	}
}
