package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chosen_player_cards_test.go — #980, the catalog half of the CR
// 614.12 "as this enters, choose a player" choice.
//
// Two cards, on purpose. True-Name Nemesis is the one the seam was
// named after and proves CR 702.16k's player quality end to end;
// Sawhorn Nemesis reads the SAME stored field from a replacement
// effect and never touches protection.go, which is what says the field
// is general machinery rather than a protection back door.
//
// Each test asserts the refusal AND the permission. A player quality
// that collapsed into "everybody" would pass half of either.

const (
	trueNameNemesisOracle = "112322ad-8f66-4cd4-98a1-f425d61a69ce"
	sawhornNemesisOracle  = "2f9107c5-991a-4c20-9b77-2e2fb4b9dc53"
)

// answerChosenPlayer answers the open as-enters prompt with `seat`,
// through the wire resolver rather than by stamping the field — the
// prompt is the half of the card this file is testing.
func answerChosenPlayer(t *testing.T, g *game.Game, chooser, seat uuid.UUID) {
	t.Helper()
	c := pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatalf("no choose-a-player prompt is open: %+v", g.PendingChoices)
	}
	idx := -1
	for i, opt := range c.PickOptions {
		if opt.Player == seat {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("seat %v is not offered by %+v", seat, c.PickOptions)
	}
	if err := g.ResolveOptionPick(c.ID, chooser, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
}

// castNemesis casts the named creature from the active seat's hand,
// resolves it, and answers its as-enters prompt with `seat`.
func castNemesis(t *testing.T, g *game.Game, name, oracle string, seat uuid.UUID) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := castCatalogSpell(t, g, name, "Creature — Test", oracle, nil)
	passPriorityAroundTable(t, g)
	answerChosenPlayer(t, g, active.ID, seat)
	return id
}

// TestTrueNameNemesisChoosesAPlayerAsItEnters — the prompt is asked as
// the creature enters, every seat is offered, and the answer lands on
// the permanent rather than on a stack item.
func TestTrueNameNemesisChoosesAPlayerAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	id := castCatalogSpell(t, g, "True-Name Nemesis", "Creature — Merfolk Rogue", trueNameNemesisOracle, nil)
	passPriorityAroundTable(t, g)

	c := pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatalf("no as-enters prompt: %+v", g.PendingChoices)
	}
	if c.Chooser != active.ID {
		t.Errorf("the controller chooses: %v, want %v", c.Chooser, active.ID)
	}
	if len(c.PickOptions) != len(g.Seats) {
		t.Errorf("every seat is offered (CR 614.12 names no restriction): %d of %d", len(c.PickOptions), len(g.Seats))
	}
	if got := g.ChosenPlayerOf(id); got != uuid.Nil {
		t.Errorf("nobody is chosen until the prompt is answered: %v", got)
	}

	answerChosenPlayer(t, g, active.ID, victim.ID)
	if got := g.ChosenPlayerOf(id); got != victim.ID {
		t.Errorf("the answer lands on the permanent: %v, want %v", got, victim.ID)
	}
}

// TestTrueNameNemesisIsProtectedFromTheChosenPlayerOnly walks the
// printed ability through the engine's one reader: the quality is the
// PLAYER, resolved off this permanent, and it names exactly one seat.
func TestTrueNameNemesisIsProtectedFromTheChosenPlayerOnly(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]

	id := castNemesis(t, g, "True-Name Nemesis", trueNameNemesisOracle, victim.ID)

	printed := protectionsOn(t, g, id)
	if !hasString(printed, "the chosen player") {
		t.Fatalf("protections = %v, want the printed quality", printed)
	}
	// The display string never holds a UUID — that is the whole reason
	// the seat lives on the quality and not in the token.
	for _, p := range printed {
		if _, err := uuid.Parse(p); err == nil {
			t.Errorf("a raw UUID reached a display string: %q", p)
		}
	}

	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != id {
				continue
			}
			if !game.ProtectedFrom(c, &game.Characteristic{Controller: victim.ID}) {
				t.Error("a source the chosen player controls has the quality")
			}
			if game.ProtectedFrom(c, &game.Characteristic{Controller: other.ID}) {
				t.Error("a source anybody else controls does not — the Nemesis is a blank against the rest of the table")
			}
		}
	})
}

// TestTrueNameNemesisCannotBeBlockedByTheChosenPlayer is the same
// quality through a second of the four DEBT checks, from the catalog
// side: the one that makes the card what it is in play.
func TestTrueNameNemesisCannotBeBlockedByTheChosenPlayer(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]

	id := castNemesis(t, g, "True-Name Nemesis", trueNameNemesisOracle, victim.ID)
	theirs := b31Push(g, victim.ID, "Chosen Player's Blocker", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	others := b31Push(g, other.ID, "Another Seat's Blocker", "Creature — Bear", "", "{1}{G}", 2, 2, "G")

	var refused, allowed game.BlockRefusal
	g.ReadSnapshot(func() {
		card := func(want uuid.UUID) *game.Card {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == want {
					return &g.Battlefield.Cards[i]
				}
			}
			t.Fatalf("card %v is not on the battlefield", want)
			return nil
		}
		refused = g.BlockPairRefusalLocked(card(id), card(theirs))
		allowed = g.BlockPairRefusalLocked(card(id), card(others))
	})
	if refused.Legal() {
		t.Error("the chosen player's creature cannot block it (CR 702.16f)")
	}
	if !allowed.Legal() {
		t.Errorf("anybody else's creature can: %q", allowed.Reason)
	}
}

// TestSawhornNemesisDoublesDamageToTheChosenPlayer — the same stored
// field read by something that is not protection at all.
func TestSawhornNemesisDoublesDamageToTheChosenPlayer(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]

	castNemesis(t, g, "Sawhorn Nemesis", sawhornNemesisOracle, victim.ID)

	victimLife, otherLife := victim.Life, other.Life
	source := b31Push(g, other.ID, "Some Pinger", "Artifact", "", "{2}", 0, 0)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, victim.ID, 3); err != nil {
			t.Fatalf("damage to the chosen player: %v", err)
		}
		if err := g.DealDamageToPlayerForEffect(source, other.ID, 3); err != nil {
			t.Fatalf("damage to another seat: %v", err)
		}
	})
	if got := victimLife - victim.Life; got != 6 {
		t.Errorf("the chosen player lost %d, want 6 (doubled)", got)
	}
	if got := otherLife - other.Life; got != 3 {
		t.Errorf("another seat lost %d, want 3 (untouched)", got)
	}
}

// TestSawhornNemesisDoublesDamageToTheChosenPlayersPermanents is the
// printed clause's second half — "or a permanent they control" — and
// the arm Fiendish Duo's player-only version does not have.
func TestSawhornNemesisDoublesDamageToTheChosenPlayersPermanents(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]

	castNemesis(t, g, "Sawhorn Nemesis", sawhornNemesisOracle, victim.ID)

	theirs := b31Push(g, victim.ID, "Chosen Player's Bear", "Creature — Bear", "", "{1}{G}", 2, 9, "G")
	others := b31Push(g, other.ID, "Another Seat's Bear", "Creature — Bear", "", "{1}{G}", 2, 9, "G")
	source := b31Push(g, other.ID, "Some Pinger", "Artifact", "", "{2}", 0, 0)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(source, theirs, 2); err != nil {
			t.Fatalf("damage to the chosen player's creature: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(source, others, 2); err != nil {
			t.Fatalf("damage to another seat's creature: %v", err)
		}
	})
	marked := func(id uuid.UUID) int {
		out := -1
		g.ReadSnapshot(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					out = g.Battlefield.Cards[i].DamageMarked
				}
			}
		})
		return out
	}
	if got := marked(theirs); got != 4 {
		t.Errorf("the chosen player's creature has %d damage, want 4 (doubled)", got)
	}
	if got := marked(others); got != 2 {
		t.Errorf("another seat's creature has %d damage, want 2 (untouched)", got)
	}
}

// TestSawhornNemesisDoublesNothingBeforeItIsAnswered — the window
// between entering and the answer. The weaker direction: an unchosen
// player is nobody, so a replacement that read it as "everybody" would
// double every damage event at the table.
func TestSawhornNemesisDoublesNothingBeforeItIsAnswered(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	castCatalogSpell(t, g, "Sawhorn Nemesis", "Creature — Dinosaur", sawhornNemesisOracle, nil)
	passPriorityAroundTable(t, g)
	if pendingOfKind(g, game.PendingChoiceOptionPick) == nil {
		t.Fatalf("the prompt is open and unanswered: %+v", g.PendingChoices)
	}

	before := victim.Life
	source := b31Push(g, victim.ID, "Some Pinger", "Artifact", "", "{2}", 0, 0)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, victim.ID, 3); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := before - victim.Life; got != 3 {
		t.Errorf("lost %d, want 3 — nobody has been chosen yet", got)
	}
}
