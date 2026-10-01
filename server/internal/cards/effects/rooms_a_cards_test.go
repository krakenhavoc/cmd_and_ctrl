package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_a_cards_test.go — ADR 0103 PR 3, first batch: Bottomless Pool,
// Glassworks, Roaring Furnace, Meat Locker and Restricted Office. Each
// is cast by a real half, unlocked through the real special action, and
// every door's effect is read off the game.

// roomsACosts are the printed mana costs, which a cast needs to pay.
var roomsACosts = map[string]string{
	"Bottomless Pool": "{U}", "Locker Room": "{4}{U}",
	"Glassworks": "{2}{R}", "Shattered Yard": "{4}{R}",
	"Roaring Furnace": "{1}{R}", "Steaming Sauna": "{3}{U}{U}",
	"Meat Locker": "{2}{U}", "Drowned Diner": "{3}{U}{U}",
	"Restricted Office": "{2}{W}{W}", "Lecture Hall": "{5}{U}{U}",
}

func roomsAHandCard(owner uuid.UUID, oracle string, left, right string) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
		Layout:     game.LayoutSplit,
		Faces: []game.Face{
			{Name: left, TypeLine: "Enchantment — Room", ManaCost: roomsACosts[left]},
			{Name: right, TypeLine: "Enchantment — Room", ManaCost: roomsACosts[right]},
		},
	}
	c.SettleImported()
	return c
}

// roomsACast puts the Room in `me`'s hand, casts the given face and
// settles the stack, answering target prompts.
func roomsACast(t *testing.T, g *game.Game, me *game.Player, oracle, left, right string, face int) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	c := roomsAHandCard(me.ID, oracle, left, right)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: face}); err != nil {
		t.Fatalf("cast %s: %v", c.Faces[face].Name, err)
	}
	roomsASettle(t, g, me.ID)
	return c.InstanceID
}

func roomsAUnlock(t *testing.T, g *game.Game, me *game.Player, room uuid.UUID, door game.DoorSide) {
	t.Helper()
	if err := g.PerformSpecialAction(me.ID, room, game.SpecialActionUnlock, game.SpecialActionParams{Door: door}); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	roomsASettle(t, g, me.ID)
}

func roomsASettle(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 4; i++ {
		answerAnyPendingTargetPrompts(t, g)
		settleOrdering(t, g, chooser)
	}
}

func roomsAHasDoors(t *testing.T, g *game.Game, room uuid.UUID, left, right bool) {
	t.Helper()
	c, ok := battlefieldCard(g, room)
	if !ok {
		t.Fatal("the Room is not on the battlefield")
	}
	if c.Unlocked.Has(game.DoorLeft) != left || c.Unlocked.Has(game.DoorRight) != right {
		t.Errorf("doors unlocked left=%v right=%v, want %v %v", c.Unlocked.Has(game.DoorLeft), c.Unlocked.Has(game.DoorRight), left, right)
	}
}

func TestBottomlessPoolBouncesOnUnlockAndLockerRoomDraws(t *testing.T) {
	const oracle = "b9ac4856-5cdb-479f-b0f5-7fa0b232a5ff"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	oppHand := opp.Hand.Size()

	// Cast the right half first: Locker Room is unlocked, Bottomless Pool is not.
	room := roomsACast(t, g, me, oracle, "Bottomless Pool", "Locker Room", 1)
	roomsAHasDoors(t, g, room, false, true)
	if _, ok := battlefieldCard(g, victim); !ok || opp.Hand.Size() != oppHand {
		t.Fatal("a locked Bottomless Pool bounced something")
	}
	roomsAUnlock(t, g, me, room, game.DoorLeft)
	roomsAHasDoors(t, g, room, true, true)
	if _, ok := battlefieldCard(g, victim); ok || opp.Hand.Size() != oppHand+1 {
		t.Errorf("unlocking Bottomless Pool: victim on battlefield=%v, opponent hand %d -> %d", ok, oppHand, opp.Hand.Size())
	}

	// Locker Room: one draw when a creature connects.
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Test", 2, 2)
	hand := me.Hand.Size()
	b784AttackEach(t, g, [2]uuid.UUID{attacker, opp.ID})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("Locker Room drew %d cards, want 1", got)
	}
}

func TestBottomlessPoolLockedLockerRoomDoesNotDraw(t *testing.T) {
	const oracle = "b9ac4856-5cdb-479f-b0f5-7fa0b232a5ff"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	room := roomsACast(t, g, me, oracle, "Bottomless Pool", "Locker Room", 0)
	roomsAHasDoors(t, g, room, true, false)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Test", 2, 2)
	hand := me.Hand.Size()
	b784AttackEach(t, g, [2]uuid.UUID{attacker, opp.ID})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("a locked Locker Room drew %d cards", got)
	}
}

func TestGlassworksDamagesAndShatteredYardPingsAtEndStep(t *testing.T) {
	const oracle = "4122720b-cad9-4ebb-b458-be71411c21e3"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	big := pushCreatureToBattlefieldForTest(g, opp.ID, "Big")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == big {
				g.Battlefield.Cards[i].Toughness = 4
			}
		}
	})
	room := roomsACast(t, g, me, oracle, "Glassworks", "Shattered Yard", 0)
	if _, ok := battlefieldCard(g, big); ok {
		t.Error("Glassworks left a 4-toughness creature alive")
	}
	life := opp.Life
	roomsAUnlock(t, g, me, room, game.DoorRight)
	roomsAHasDoors(t, g, room, true, true)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("Shattered Yard: opponent life %d -> %d, want -1", life, opp.Life)
	}
}

func TestRoaringFurnaceDamageEqualsHandAndSteamingSaunaDrawsWithNoHandLimit(t *testing.T) {
	const oracle = "d5f31713-d380-42ba-8052-4b8d9beb3958"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	small := pushCreatureToBattlefieldForTest(g, opp.ID, "Small")
	advanceToMain(t, g)
	for me.Hand.Size() < 4 {
		me.Hand.PushTop(game.NewCard("basic-filler", uuid.Nil))
	}
	// The cast Room leaves the hand first; the trigger then counts what is left.
	want := me.Hand.Size() - 1
	maxHand := func() int {
		var n int
		g.ReadSnapshot(func() { n = g.EffectiveMaxHandSizeLocked(me) })
		return n
	}
	room := roomsACast(t, g, me, oracle, "Roaring Furnace", "Steaming Sauna", 0)
	if _, ok := battlefieldCard(g, small); want >= 2 && ok {
		t.Errorf("Roaring Furnace with %d cards in hand left a 2/2 alive", want)
	}
	if maxHand() == game.NoMaxHandSize {
		t.Fatal("a locked Steaming Sauna already removed the hand limit")
	}
	roomsAUnlock(t, g, me, room, game.DoorRight)
	if maxHand() != game.NoMaxHandSize {
		t.Error("an unlocked Steaming Sauna does not remove the maximum hand size")
	}
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("Steaming Sauna drew %d cards at the end step, want 1", got)
	}
}

func TestMeatLockerStunsAndDrownedDinerLoots(t *testing.T) {
	const oracle = "897eef47-3e99-4a5f-8a3e-ddb06bc95e9a"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	target := pushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	room := roomsACast(t, g, me, oracle, "Meat Locker", "Drowned Diner", 0)
	c, _ := battlefieldCard(g, target)
	if !c.Tapped || c.Counters[game.CounterStun] != 2 {
		t.Errorf("Meat Locker: tapped=%v stun=%d, want tapped and 2", c.Tapped, c.Counters[game.CounterStun])
	}
	hand := me.Hand.Size()
	roomsAUnlock(t, g, me, room, game.DoorRight)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("Drowned Diner drew %d, want 3", got)
	}
	if got := discardOwed(g, me.ID); got != 1 {
		t.Errorf("Drowned Diner owes %d discards, want 1", got)
	}
}

func TestRestrictedOfficeDestroysBigCreaturesAndLectureHallShieldsOthers(t *testing.T) {
	const oracle = "b36286fa-4007-4276-af9c-d8f0da9e94d1"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mk := func(owner uuid.UUID, name string, power int) uuid.UUID {
		return b12Creature(g, owner, name, "Creature — Test", power, 5)
	}
	big, small, theirBig := mk(me.ID, "Big", 3), mk(me.ID, "Small", 2), mk(opp.ID, "Their Big", 4)
	room := roomsACast(t, g, me, oracle, "Restricted Office", "Lecture Hall", 0)
	for id, alive := range map[uuid.UUID]bool{big: false, small: true, theirBig: false} {
		if _, ok := battlefieldCard(g, id); ok != alive {
			t.Errorf("after Restricted Office, creature alive=%v want %v", ok, alive)
		}
	}
	hasHexproof := func(id uuid.UUID) bool { return effectiveAbilitiesContain(t, g, id, "hexproof") }
	if hasHexproof(small) {
		t.Error("a locked Lecture Hall granted hexproof")
	}
	roomsAUnlock(t, g, me, room, game.DoorRight)
	if !hasHexproof(small) {
		t.Error("an unlocked Lecture Hall did not grant hexproof to another permanent")
	}
	if hasHexproof(room) {
		t.Error("Lecture Hall granted hexproof to itself (it says other permanents)")
	}
}
