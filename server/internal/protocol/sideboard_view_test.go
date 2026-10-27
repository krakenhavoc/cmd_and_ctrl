package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sideboard_view_test.go — ADR 0191 §7: a seat's sideboard is its
// owner's to read; everybody else gets a count at most, and a card an
// effect reveals is public from then on.

func sideboardViewGame(t *testing.T) (*game.Game, []game.Card) {
	t.Helper()
	g := game.NewGame()
	var side []game.Card
	for i := 0; i < 3; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := 0; j < 20; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if i == 0 {
			for _, n := range []string{"Secret Sorcery", "Secret Instant"} {
				c := game.NewCard(n, uuid.Nil)
				c.TypeLine = "Sorcery"
				c.IsSideboard = true
				side = append(side, c)
				deck = append(deck, c)
			}
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.StartWithSource(nil); err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	return g, side
}

func TestSideboardIsTheOwnersAndNobodyElses(t *testing.T) {
	g, side := sideboardViewGame(t)
	owner, opp := g.Seats[0].ID.String(), g.Seats[1].ID.String()
	for _, tc := range []struct {
		name   string
		viewer string
		cards  bool
	}{
		{"owner", owner, true},
		{"opponent", opp, false},
		{"spectator", SpectatorViewerID, false},
		{"admin", "", false},
	} {
		v := ViewOfGameFor(g, tc.viewer)
		seat := v.Seats[0]
		if tc.cards {
			if len(seat.Sideboard) != 2 || seat.Sideboard[0].Name != "Secret Sorcery" || seat.Sideboard[0].InstanceID != side[0].InstanceID.String() {
				t.Errorf("%s: sideboard %+v", tc.name, seat.Sideboard)
			}
			if seat.Sideboard[0].TypeLine != "Sorcery" || !seat.Sideboard[0].KnownByYou {
				t.Errorf("%s: the owner's card is %+v, want it face up", tc.name, seat.Sideboard[0])
			}
		} else {
			if len(seat.Sideboard) != 0 {
				t.Errorf("%s: sees %d sideboard cards", tc.name, len(seat.Sideboard))
			}
			raw, err := json.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			for _, leak := range []string{"Secret Sorcery", "Secret Instant", side[0].InstanceID.String(), side[1].InstanceID.String()} {
				if strings.Contains(string(raw), leak) {
					t.Errorf("%s: the wire carries %q", tc.name, leak)
				}
			}
		}
		// The count is public, like a hand's size.
		if seat.SideboardCount != 2 {
			t.Errorf("%s: sideboard_count %d, want 2", tc.name, seat.SideboardCount)
		}
		if v.Seats[1].SideboardCount != 0 || len(v.Seats[1].Sideboard) != 0 {
			t.Errorf("%s: seat 1 shows a sideboard", tc.name)
		}
	}
}

// ADR 0167 §5 (CR 723.4): the controller of a player's turn sees what
// the controlled player can see, "except information about cards
// outside the game". Exercised by what the controller's view of the
// controlled player's seat carries.
func TestAControllerDoesNotSeeTheControlledPlayersSideboard(t *testing.T) {
	g, _ := sideboardViewGame(t)
	owner, ctrl := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.PlayerControls = append(g.PlayerControls, game.PlayerControl{
			Controller: ctrl.ID, Controlled: owner.ID, Until: game.ControlUntilTurn,
			Stage: game.ControlTurn, TurnSeq: g.Turn.Seq,
		})
	})
	v := ViewOfGameFor(g, ctrl.ID.String())
	if len(v.Seats[0].Sideboard) != 0 {
		t.Errorf("the controller sees %d of the controlled player's sideboard cards", len(v.Seats[0].Sideboard))
	}
	raw, _ := json.Marshal(v)
	if strings.Contains(string(raw), "Secret Sorcery") {
		t.Error("the controller's wire carries a sideboard card name")
	}
}

// The owner's pick prompt shows them the candidates, and the table sees
// only that a choice is open.
func TestAnOutsidePickShowsItsCandidatesToTheOwnerAlone(t *testing.T) {
	g, side := sideboardViewGame(t)
	owner := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(game.OutsidePickPrompt{Player: owner.ID, Question: "Burning Wish — reveal a sorcery card", Max: 1})
	})
	own := ViewOfGameFor(g, owner.ID.String())
	if len(own.PendingChoices) != 1 {
		t.Fatalf("%d pending choices", len(own.PendingChoices))
	}
	pc := own.PendingChoices[0]
	if pc.Kind != string(game.PendingChoiceChooseCards) || len(pc.Options) != 2 || pc.ChooseDestination != string(game.ChooseFromOutside) {
		t.Fatalf("owner's prompt kind %q, %d options, destination %q", pc.Kind, len(pc.Options), pc.ChooseDestination)
	}
	if pc.Options[0].Name != "Secret Sorcery" || pc.Options[0].InstanceID != side[0].InstanceID.String() {
		t.Errorf("owner's first option %+v", pc.Options[0])
	}
	for _, viewer := range []string{g.Seats[1].ID.String(), SpectatorViewerID, ""} {
		v := ViewOfGameFor(g, viewer)
		if len(v.PendingChoices) != 1 {
			t.Fatalf("viewer %q: %d pending choices", viewer, len(v.PendingChoices))
		}
		other := v.PendingChoices[0]
		if len(other.Options) != 0 || other.ChooseMax != 0 {
			t.Errorf("viewer %q sees %d options, max %d", viewer, len(other.Options), other.ChooseMax)
		}
		raw, _ := json.Marshal(v)
		if strings.Contains(string(raw), "Secret Sorcery") || strings.Contains(string(raw), side[0].InstanceID.String()) {
			t.Errorf("viewer %q: the wire carries the sideboard card", viewer)
		}
	}
}

// A revealed card is public: the reveal frame names it, an opponent
// asked to choose among the revealed cards sees both, and the card that
// stays outside is still hidden in the owner's sideboard view of everyone
// else.
func TestARevealedSideboardCardIsPublic(t *testing.T) {
	g, side := sideboardViewGame(t)
	owner, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.RevealOutsideForEffect(game.RevealSpec{
			Player: owner.ID, Reason: "Extrapolate the Impossible — reveal two cards from outside the game",
			Cards: []uuid.UUID{side[0].InstanceID, side[1].InstanceID},
		})
		if _, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
			Chooser: opp.ID, Owner: owner.ID, Question: "choose one", Min: 1, Max: 1,
			Cards: []uuid.UUID{side[0].InstanceID, side[1].InstanceID},
		}); err != nil {
			t.Fatal(err)
		}
	})
	v := ViewOfGameFor(g, opp.ID.String())
	if len(v.Reveals) != 1 || len(v.Reveals[0].Cards) != 2 || v.Reveals[0].Cards[0].Name != "Secret Sorcery" {
		t.Fatalf("reveal frame %+v", v.Reveals)
	}
	if len(v.PendingChoices) != 1 || v.PendingChoices[0].Kind != string(game.PendingChoiceRevealPick) {
		t.Fatalf("pending %+v", v.PendingChoices)
	}
	if got := len(v.PendingChoices[0].Options); got != 2 {
		t.Errorf("the opponent sees %d of the 2 revealed cards", got)
	}
	// Still not in the opponent's view of the sideboard itself.
	if len(v.Seats[0].Sideboard) != 0 {
		t.Error("revealing two cards opened the whole sideboard")
	}
	// And a third seat that was told the same: public.
	third := ViewOfGameFor(g, g.Seats[2].ID.String())
	if len(third.Reveals) != 1 {
		t.Errorf("a third seat saw %d reveal frames", len(third.Reveals))
	}
}
