package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// chained_choice_test.go — the enumerator half of the chained-choice
// work, and the reason it exists at all.
//
// `EnumerateFor` returns ONLY a choice's answers while a seat owes one.
// So a prompt kind the enumerator does not know is not a cosmetic gap:
// it is a bot seat handed an empty move list, asleep, holding a live
// table with a human on it. That is #544, and #499 counts four kinds
// still in that state. These tests are the promise that the two kinds
// added here are not a fifth and sixth.

func handOf(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for _, c := range p.Hand.Cards {
		if len(out) == n {
			break
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// TestConfirmEnumeratesBothBranches — and dispatchAll proves the engine
// accepts every one of them, which is the property the wedge in #544
// turned out not to have.
func TestConfirmEnumeratesBothBranches(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueConfirmForEffect(game.ConfirmPrompt{
			Chooser:      me.ID,
			Question:     "Sylvan Library — pay 4 life to keep it?",
			AcceptLabel:  "Pay 4 life",
			DeclineLabel: "Put it on top",
			OnAccept:     func(*game.Game) error { return nil },
			OnDecline:    func(*game.Game) error { return nil },
		})
	})

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 2 {
		t.Fatalf("enumerated %d answers, want 2: %v", len(moves), labels(moves))
	}
	// The card's own words, not "yes" / "no": a bot's narration and a
	// human's move list read the same prompt.
	if !hasLabel(moves, "Sylvan Library — pay 4 life to keep it?: Pay 4 life") {
		t.Errorf("accept branch is unlabelled: %v", labels(moves))
	}
	if !hasLabel(moves, "Sylvan Library — pay 4 life to keep it?: Put it on top") {
		t.Errorf("decline branch is unlabelled: %v", labels(moves))
	}
	dispatchAll(t, g, me.ID, moves)
}

// TestConfirmIsTheOnlyThingOnOffer — the short-circuit that makes an
// unenumerated kind fatal, stated as a test so the next person sees
// why the case above is not optional.
func TestConfirmIsTheOnlyThingOnOffer(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	if len(legal.EnumerateFor(g, me.ID)) == 0 {
		t.Fatal("setup: the active seat should have moves in its main phase")
	}
	g.WithWriteLock(func() {
		g.QueueConfirmForEffect(game.ConfirmPrompt{Chooser: me.ID, Question: "?"})
	})
	moves := legal.EnumerateFor(g, me.ID)
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Fatalf("a seat owing a choice was offered %q as well — if the choice's own "+
				"answers were missing, this list would be empty and the seat would sleep",
				m.Label)
		}
	}
	if len(moves) == 0 {
		t.Fatal("a seat owing a confirm was offered nothing at all")
	}
}

// TestChooseCardsEnumeratesExactlyTheLegalSets. The bounds live on the
// choice, so every set offered is a set the resolver accepts and every
// set the resolver accepts is offerable — the invariant Myriad
// Landscape's Validate hook broke.
func TestChooseCardsEnumeratesExactlyTheLegalSets(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	cards := handOf(me, 3)
	if len(cards) != 3 {
		t.Fatalf("setup: wanted 3 cards in hand, got %d", len(cards))
	}
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  me.ID,
			Question: "Sylvan Library — choose two cards drawn this turn",
			Cards:    cards,
			Min:      2,
			Max:      2,
			Zone:     game.ZoneHand,
			Then:     func(*game.Game, []uuid.UUID) error { return nil },
		})
	})

	moves := legal.EnumerateFor(g, me.ID)
	// C(3,2) = 3, and no "choose nothing": the floor is two.
	if len(moves) != 3 {
		t.Fatalf("enumerated %d answers, want 3: %v", len(moves), labels(moves))
	}
	for _, m := range moves {
		if hasLabel([]legal.Move{m}, "Sylvan Library — choose two cards drawn this turn: choose nothing") {
			t.Error("offered an empty answer to a prompt whose floor is two")
		}
	}
	dispatchAll(t, g, me.ID, moves)
}

// TestChooseCardsOffersTheEmptyAnswerWhenItsFloorIsZero — the
// always-terminating answer, offered first, for the same reason search
// offers "fail to find" first (#544 candidate fix 3).
func TestChooseCardsOffersTheEmptyAnswerWhenItsFloorIsZero(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	cards := handOf(me, 2)
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  me.ID,
			Question: "Put back",
			Cards:    cards,
			Min:      0,
			Zone:     game.ZoneHand,
			Then:     func(*game.Game, []uuid.UUID) error { return nil },
		})
	})
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("a seat owing a choose-cards prompt was offered nothing")
	}
	if moves[0].Label != "Put back: choose nothing" {
		t.Errorf("first answer is %q, want the empty one", moves[0].Label)
	}
	dispatchAll(t, g, me.ID, moves)
}
