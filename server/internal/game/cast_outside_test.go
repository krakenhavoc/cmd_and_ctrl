package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cast_outside_test.go — #3955: a card outside the game is cast or played
// through a per-instance permission, from the caster's own container, and
// from nowhere else.

// outsideCard is a creature with a real cost sitting in seat 0's
// container, so a free cast is distinguishable from a paid one.
func outsideCreature(name, cost string) Card {
	c := sbCard(name, "Creature — Eldrazi")
	c.ManaCost = cost
	return c
}

func outsideLand(name string) Card {
	return sbCard(name, "Land")
}

// outsideGameInMain is a started game with seat 0 in its precombat main
// phase and `side` outside the game.
func outsideGameInMain(t *testing.T, side ...Card) *Game {
	t.Helper()
	g := newSideboardGame(t, side...)
	advanceTo(t, g, StepPrecombatMain)
	return g
}

func grantFree(g *Game, p *Player, id uuid.UUID) bool {
	var ok bool
	g.WithWriteLock(func() { ok = g.FreeCastFromOutsideForEffect(p.ID, uuid.New(), id, "Test Source") })
	return ok
}

func grantPlay(g *Game, p *Player, id uuid.UUID) bool {
	var ok bool
	g.WithWriteLock(func() { ok = g.PlayFromOutsideForEffect(p.ID, uuid.New(), id, "Test Wish") })
	return ok
}

func outsideOnStack(g *Game, id uuid.UUID) bool {
	return g.Stack.Contains(id)
}

// The headline: a free cast from outside the game puts the card on the
// stack, pays nothing for a {10} spell, and the card leaves the
// container (CR 601.2a).
func TestFreeCastFromOutsideCastsWithoutPaying(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("Big Eldrazi", "{10}"))
	me := g.Seats[0]
	id := me.SideboardCards()[0].InstanceID

	if !grantFree(g, me, id) {
		t.Fatal("the grant was refused for a card in the container")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "outside", Strict: true}); err != nil {
		t.Fatalf("free cast from outside: %v", err)
	}
	if !outsideOnStack(g, id) {
		t.Error("the spell is not on the stack")
	}
	if me.Sideboard.Contains(id) {
		t.Error("the card is still outside the game after it was cast")
	}
	if c, ok := g.LookupCardForEffect(id); !ok || c.Controller != me.ID {
		t.Errorf("the spell on the stack is %+v (found %v)", c, ok)
	}
}

// Nothing a card prints opens the container: without a permission the
// cast is refused, and the card stays where it was.
func TestCastFromOutsideNeedsAPermission(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("Big Eldrazi", "{1}"))
	me := g.Seats[0]
	id := me.SideboardCards()[0].InstanceID
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "outside"})
	if !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Errorf("cast with no permission: %v, want ErrCastZoneNotAllowed", err)
	}
	if !me.Sideboard.Contains(id) || outsideOnStack(g, id) {
		t.Error("a refused cast moved the card")
	}
}

// A permission is the caster's own: another seat's card, or a card in
// another seat's container, is not reachable by naming the zone.
func TestCastFromOutsideIsTheCastersOwnContainer(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("Big Eldrazi", "{1}"))
	me, opp := g.Seats[0], g.Seats[1]
	id := me.SideboardCards()[0].InstanceID
	if !grantFree(g, me, id) {
		t.Fatal("grant refused")
	}
	// The opponent names the zone and the card: their container does not
	// hold it.
	if err := g.CastSpell(opp.ID, id, CastSpellParams{FromZone: "outside"}); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("opponent cast a card from another seat's container: %v", err)
	}
	// And a grant over a card the holder does not own is refused outright.
	if grantFree(g, opp, id) {
		t.Error("a permission was granted over a card in somebody else's container")
	}
}

// "Cast" is a cast: a land is not a spell, so the free cast refuses it
// (CR 305.1) while the play window plays it, as a land drop.
func TestFreeCastRefusesALandAndPlayPlaysIt(t *testing.T) {
	g := outsideGameInMain(t, outsideLand("Forest"), outsideLand("Island"))
	me := g.Seats[0]
	forest, island := me.SideboardCards()[0].InstanceID, me.SideboardCards()[1].InstanceID

	if !grantFree(g, me, forest) {
		t.Fatal("grant refused")
	}
	if err := g.CastSpell(me.ID, forest, CastSpellParams{FromZone: "outside"}); !errors.Is(err, ErrNoPlayPermission) {
		t.Errorf("a free cast played a land: %v", err)
	}

	if !grantPlay(g, me, island) {
		t.Fatal("play grant refused")
	}
	before := me.Hand.Size()
	drops := g.LandDropsRemainingLocked(me.ID)
	if err := g.CastSpell(me.ID, island, CastSpellParams{FromZone: "outside"}); err != nil {
		t.Fatalf("play a land from outside: %v", err)
	}
	if !g.Battlefield.Contains(island) {
		t.Error("the land is not on the battlefield")
	}
	if me.Sideboard.Contains(island) {
		t.Error("the land is still outside the game")
	}
	if me.Hand.Size() != before {
		t.Error("the hand changed")
	}
	if got := g.LandDropsRemainingLocked(me.ID); got != drops-1 {
		t.Errorf("land drops left %d, want %d: playing a land from outside counts as the land drop", got, drops-1)
	}
}

// "Play ... this turn" pays the card's own cost and keeps its own timing:
// an unpaid {1} is refused under strict mana, and a sorcery-speed card is
// refused in the opponent's turn.
func TestPlayFromOutsidePaysAndKeepsItsTiming(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("Costly", "{3}"))
	me := g.Seats[0]
	id := me.SideboardCards()[0].InstanceID
	if !grantPlay(g, me, id) {
		t.Fatal("grant refused")
	}
	var short *InsufficientManaError
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "outside", Strict: true}); !errors.As(err, &short) {
		t.Fatalf("a Wish play with no mana: %v, want InsufficientManaError", err)
	}
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "outside", Strict: true}); err != nil {
		t.Fatalf("paid play: %v", err)
	}
	if !outsideOnStack(g, id) {
		t.Error("not on the stack")
	}
}

// The free cast is pass-closed: the holder's next priority pass ends the
// window and the card STAYS outside the game, uncast.
func TestFreeCastFromOutsideLapsesOnPassAndStaysOutside(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("Big Eldrazi", "{10}"))
	me := g.Seats[0]
	id := me.SideboardCards()[0].InstanceID
	if !grantFree(g, me, id) {
		t.Fatal("grant refused")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !me.Sideboard.Contains(id) {
		t.Fatal("the card left the container when its window lapsed")
	}
	if len(me.CastPermissions) != 0 {
		t.Errorf("the lapsed grant is still stored: %+v", me.CastPermissions)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "outside"}); err == nil {
		t.Error("a lapsed free cast still cast")
	}
}

// "This turn" ends with the turn: the play window is swept at cleanup and
// the card stays outside.
func TestPlayFromOutsideEndsWithTheTurn(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("Costly", "{1}"))
	me := g.Seats[0]
	id := me.SideboardCards()[0].InstanceID
	if !grantPlay(g, me, id) {
		t.Fatal("grant refused")
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
	if len(me.CastPermissions) != 0 {
		t.Errorf("the play window outlived the turn: %+v", me.CastPermissions)
	}
	if !me.Sideboard.Contains(id) {
		t.Error("the card left the container")
	}
}

// A live play window is not swept as "the card is gone" at the next step:
// the sweep finds the named card in the container (it is in no zone the
// card index reaches), so Wish's play is still good in the second main
// phase.
func TestPlayFromOutsideSurvivesStepChanges(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("Costly", "{1}"))
	me := g.Seats[0]
	id := me.SideboardCards()[0].InstanceID
	if !grantPlay(g, me, id) {
		t.Fatal("grant refused")
	}
	advanceTo(t, g, StepPostcombatMain)
	if len(me.CastPermissions) != 1 {
		t.Fatalf("the play window was swept mid-turn: %+v", me.CastPermissions)
	}
	fundOutside(me)
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "outside", Strict: true}); err != nil {
		t.Fatalf("play in the second main phase: %v", err)
	}
}

func fundOutside(p *Player) {
	p.ManaPool.AddMana(ManaToken{Color: "C"})
}

// The window is cast per card: casting one granted card leaves the others
// open ("any number of Eldrazi cards", Spawnsire of Ulamog).
func TestFreeCastsFromOutsideAreIndependent(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("One", "{5}"), outsideCreature("Two", "{6}"), outsideCreature("Three", "{7}"))
	me := g.Seats[0]
	ids := []uuid.UUID{me.SideboardCards()[0].InstanceID, me.SideboardCards()[1].InstanceID, me.SideboardCards()[2].InstanceID}
	for _, id := range ids {
		if !grantFree(g, me, id) {
			t.Fatal("grant refused")
		}
	}
	for _, id := range ids {
		if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "outside"}); err != nil {
			t.Fatalf("cast %s: %v", id, err)
		}
	}
	for _, id := range ids {
		if !outsideOnStack(g, id) {
			t.Errorf("%s is not on the stack", id)
		}
	}
	if me.Sideboard.Size() != 0 {
		t.Errorf("%d cards are still outside", me.Sideboard.Size())
	}
}

// ADR 0167 section 5: whoever makes a player's decisions is offered
// nothing from outside the game, and cannot cast from it either.
func TestCastFromOutsideIsClosedToAControlledPlayer(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("Big Eldrazi", "{10}"))
	me, opp := g.Seats[0], g.Seats[1]
	id := me.SideboardCards()[0].InstanceID
	if !grantFree(g, me, id) {
		t.Fatal("grant refused")
	}
	g.WithWriteLock(func() {
		g.PlayerControls = append(g.PlayerControls, PlayerControl{
			Controller: opp.ID, Controlled: me.ID, Until: ControlUntilTurn, Stage: ControlTurn, TurnSeq: g.Turn.Seq,
		})
	})
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "outside"}); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("a controlled player cast from outside the game: %v", err)
	}
	var ok bool
	g.WithWriteLock(func() { ok = g.FreeCastFromOutsideForEffect(me.ID, uuid.New(), id, "x") })
	if ok {
		t.Error("a controlled player was granted a cast from outside the game")
	}
}

// The permission survives a deploy: a grant over a card in the container
// round-trips through a snapshot, and the restored table still casts it.
func TestFreeCastFromOutsideSurvivesASnapshot(t *testing.T) {
	g := outsideGameInMain(t, outsideCreature("Big Eldrazi", "{10}"))
	me := g.Seats[0]
	id := me.SideboardCards()[0].InstanceID
	if !grantFree(g, me, id) {
		t.Fatal("grant refused")
	}
	if len(g.Clone().Seats[0].CastPermissions) != 1 {
		t.Fatal("Clone lost the permission")
	}
	_, restored := roundTrip(t, g)
	rme := restored.Seats[0]
	if len(rme.CastPermissions) != 1 || rme.CastPermissions[0].Zone != ZoneOutside ||
		rme.CastPermissions[0].LapseOnPass != LapseStaysOutside {
		t.Fatalf("restored permissions: %+v", rme.CastPermissions)
	}
	if err := castFrom(restored, rme, id); err != nil {
		t.Fatalf("free cast after a restore: %v", err)
	}
	if !restored.Stack.Contains(id) {
		t.Error("the restored cast did not reach the stack")
	}
}

func castFrom(g *Game, p *Player, id uuid.UUID) error {
	return g.CastSpell(p.ID, id, CastSpellParams{FromZone: "outside", Strict: true})
}

// A card file cannot open the container for every copy of itself.
func TestNoCardOpensOutsideByItsOwnText(t *testing.T) {
	g := newActiveGame(t)
	c := NewCard("Anything", g.Seats[0].ID)
	c.TypeLine = "Creature"
	var opens bool
	g.WithWriteLock(func() { opens = g.cardOpensZoneLocked(c, "any-key", ZoneOutside) })
	if opens {
		t.Error("a card opened ZoneOutside by itself")
	}
}
