package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_revealed_view_test.go — ADR 0199, #3953: a hand its owner plays
// with revealed reaches every seat's frame, cards drawn later included,
// for exactly as long as the effect lasts.

const viewHandRevealer = "view-fixture-hand-revealer"

func stubViewHandRevealer(t *testing.T) {
	t.Helper()
	prevLookup := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef {
		if key == viewHandRevealer {
			return &game.CardDef{}
		}
		if prevLookup == nil {
			return nil
		}
		return prevLookup(key)
	}
	prev := game.CatalogHandsRevealed
	game.CatalogHandsRevealed = func(key string) game.HandRevealScope {
		if key == viewHandRevealer {
			return game.HandRevealOpponents
		}
		if prev == nil {
			return game.HandRevealNone
		}
		return prev(key)
	}
	t.Cleanup(func() {
		game.CatalogLookup = prevLookup
		game.CatalogHandsRevealed = prev
	})
}

// handNamesSeenBy is the names of `owner`'s hand cards `viewer`'s frame
// carries face up.
func handNamesSeenBy(t *testing.T, g *game.Game, owner uuid.UUID, viewer string) []string {
	t.Helper()
	fv := FilterViewFor(ViewOfGame(g), viewer)
	for _, s := range fv.Seats {
		if s.ID != owner.String() {
			continue
		}
		var out []string
		for _, c := range s.Hand.Cards {
			if c.KnownByYou && c.Name != "" {
				out = append(out, c.Name)
			}
		}
		return out
	}
	t.Fatalf("no seat %s in the frame", owner)
	return nil
}

func seatHandRevealed(g *game.Game, owner uuid.UUID) []string {
	for _, s := range ViewOfGame(g).Seats {
		if s.ID == owner.String() {
			return s.HandRevealed
		}
	}
	return nil
}

// A granted reveal with Stromgald Spy's duration: every other seat and a
// spectator see the hand, a card drawn while it lasts too, and nobody but
// the owner once the source leaves the battlefield.
func TestRevealedHandReachesEverySeatWhileTheDurationLasts(t *testing.T) {
	g := buildThreeSeatGame(t)
	spyOwner, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	// A card put straight into a hand here carries its owner as a
	// knower, the way a draw makes it.
	inHand := func(name string) game.Card {
		c := game.NewCard(name, victim.ID)
		c.AddKnower(victim.ID)
		return c
	}
	secret := inHand("Secret Plan")
	spy := game.NewCard("Stromgald Spy", spyOwner.ID)
	spy.TypeLine = "Creature — Human Rogue"
	g.WithWriteLock(func() {
		victim.Hand.PushTop(secret)
		g.Battlefield.PushTop(spy)
	})
	for _, viewer := range []string{spyOwner.ID.String(), bystander.ID.String(), SpectatorViewerID} {
		if got := handNamesSeenBy(t, g, victim.ID, viewer); slices.Contains(got, "Secret Plan") {
			t.Fatalf("before the grant, %s already sees the hand: %v", viewer, got)
		}
	}

	g.WithWriteLock(func() {
		d, ok := g.ForAsLongAsOnBattlefieldDuration(spy.InstanceID)
		if !ok {
			t.Fatal("the Spy is not on the battlefield")
		}
		g.GrantHandRevealedForEffect(victim.ID, "Stromgald Spy", spy.InstanceID, d)
	})
	if got := seatHandRevealed(g, victim.ID); !slices.Equal(got, []string{"Stromgald Spy"}) {
		t.Errorf("hand_revealed = %v, want [Stromgald Spy]", got)
	}
	want := victim.Hand.Size()
	for _, viewer := range []string{spyOwner.ID.String(), bystander.ID.String(), SpectatorViewerID} {
		got := handNamesSeenBy(t, g, victim.ID, viewer)
		if !slices.Contains(got, "Secret Plan") || len(got) != want {
			t.Errorf("%s sees %v, want all %d cards of the hand", viewer, got, want)
		}
	}
	// The owner's own cards say every other seat knows them.
	for _, s := range FilterViewFor(ViewOfGame(g), victim.ID.String()).Seats {
		if s.ID != victim.ID.String() {
			continue
		}
		for _, c := range s.Hand.Cards {
			if len(c.KnownBySeats) != 2 {
				t.Errorf("owner's %s: known_by_seats = %v, want the other two seats", c.Name, c.KnownBySeats)
			}
		}
	}

	// A card that arrives later is revealed as well.
	g.WithWriteLock(func() { victim.Hand.PushTop(inHand("Late Draw")) })
	if got := handNamesSeenBy(t, g, victim.ID, bystander.ID.String()); !slices.Contains(got, "Late Draw") {
		t.Errorf("a card drawn while revealed is hidden: %v", got)
	}

	// The Spy leaves: the hand is hidden again, and nothing is remembered.
	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(spy.InstanceID); err != nil {
			t.Fatalf("remove: %v", err)
		}
	})
	if got := seatHandRevealed(g, victim.ID); len(got) != 0 {
		t.Errorf("hand_revealed = %v after the Spy left, want none", got)
	}
	for _, viewer := range []string{spyOwner.ID.String(), bystander.ID.String(), SpectatorViewerID} {
		if got := handNamesSeenBy(t, g, victim.ID, viewer); len(got) != 0 {
			t.Errorf("after the Spy left, %s still sees %v", viewer, got)
		}
	}
	if got := handNamesSeenBy(t, g, victim.ID, victim.ID.String()); len(got) != victim.Hand.Size() {
		t.Errorf("the owner sees %d of their %d cards", len(got), victim.Hand.Size())
	}
}

// Telepathy's shape: every opponent's hand is revealed to every seat,
// and its controller's own hand stays hidden.
func TestOpponentsPlayWithTheirHandsRevealed(t *testing.T) {
	stubViewHandRevealer(t)
	g := buildThreeSeatGame(t)
	me, left, right := g.Seats[0], g.Seats[1], g.Seats[2]
	tp := game.NewCard("Telepathy", me.ID)
	tp.TypeLine = "Enchantment"
	tp.OracleID = viewHandRevealer
	g.WithWriteLock(func() { g.Battlefield.PushTop(tp) })

	if got := handNamesSeenBy(t, g, left.ID, me.ID.String()); len(got) != left.Hand.Size() || len(got) == 0 {
		t.Errorf("Telepathy's controller sees %v of the opponent's %d cards", got, left.Hand.Size())
	}
	if got := handNamesSeenBy(t, g, left.ID, right.ID.String()); len(got) != left.Hand.Size() {
		t.Errorf("another opponent sees %v of the revealed hand of %d", got, left.Hand.Size())
	}
	if got := handNamesSeenBy(t, g, me.ID, left.ID.String()); len(got) != 0 {
		t.Errorf("Telepathy revealed its controller's own hand: %v", got)
	}
	if got := seatHandRevealed(g, me.ID); len(got) != 0 {
		t.Errorf("controller's hand_revealed = %v, want none", got)
	}
	if got := seatHandRevealed(g, right.ID); !slices.Equal(got, []string{"Telepathy"}) {
		t.Errorf("opponent's hand_revealed = %v, want [Telepathy]", got)
	}
}
