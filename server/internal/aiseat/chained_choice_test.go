package aiseat_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// chained_choice_test.go — can a bot finish a chain?
//
// #544 is the reason this file exists. A seat that owes a choice is
// enumerated that choice's answers and NOTHING else, so a prompt the
// enumerator cannot answer is a sleeping bot holding a live table —
// which happened to a human opponent on turn 2. A chain multiplies the
// exposure: every link is another prompt that has to be answerable,
// and the last link is the one that has to actually END.
//
// So this walks the real policy through the real engine, one enumerated
// move at a time, and fails on the two shapes that wedge: an empty move
// list, and an answer the engine refuses.
//
// Deliberately NOT behind requireGameTests. It is not a whole game — it
// is a bounded loop over one card's prompts, in milliseconds — and the
// gate on the whole-game tests is precisely why #544 reached
// production.

// newSettledTable is newRoom past the mulligan window: every seat has
// kept, so the enumerator is out of its "keep or mulligan" world and
// back to offering real moves.
func newSettledTable(t *testing.T, seed uint64) *game.Game {
	t.Helper()
	g := newRoom(t, 2, seed).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// driveChoices runs the policy against the seat for as long as it owes
// a pending choice, dispatching each decision. Returns the labels it
// took. Fails on an empty move list (the #544 wedge) or a rejected
// dispatch (the #543 one).
func driveChoices(t *testing.T, g *game.Game, pol aiseat.Policy, seat uuid.UUID, maxSteps int) []string {
	t.Helper()
	var taken []string
	for step := 0; step < maxSteps; step++ {
		if !owesChoice(g, seat) {
			return taken
		}
		moves := legal.EnumerateFor(g, seat)
		if len(moves) == 0 {
			t.Fatalf("step %d: the seat owes a choice and was offered NOTHING — "+
				"this is the #544 wedge: the runner sleeps and the table stops. "+
				"Open choices: %s", step, openChoices(g, seat))
		}
		in := aiseat.Input{
			View:  protocol.ViewOfGameFor(g, seat.String()),
			Seat:  seat,
			Moves: moves,
		}
		d, err := pol.Decide(context.Background(), in)
		if err != nil {
			t.Fatalf("step %d: policy failed: %v", step, err)
		}
		if d.Index == aiseat.Decline {
			t.Fatalf("step %d: the policy declined a window it owes an answer in. "+
				"A decline with no pass on offer returns the runner to sleep "+
				"(runner.go) — a chain link must be answerable, not skippable.", step)
		}
		if d.Index < 0 || d.Index >= len(moves) {
			t.Fatalf("step %d: policy returned out-of-range index %d of %d", step, d.Index, len(moves))
		}
		mv := moves[d.Index]
		if err := actions.Dispatch(g, actions.Action{
			Type:   actions.Type(mv.Type),
			Player: mv.Player,
			Caller: seat,
			Params: mv.Params,
		}); err != nil {
			t.Fatalf("step %d: the engine refused an ENUMERATED answer %q: %v — "+
				"an enumerator that offers what the resolver rejects is #544",
				step, mv.Label, err)
		}
		taken = append(taken, mv.Label)
	}
	t.Fatalf("the chain did not terminate in %d steps; took %v", maxSteps, taken)
	return nil
}

func owesChoice(g *game.Game, seat uuid.UUID) bool {
	owed := false
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == seat {
				owed = true
				return
			}
		}
	})
	return owed
}

func openChoices(g *game.Game, seat uuid.UUID) string {
	var b strings.Builder
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == seat {
				b.WriteString(string(c.Kind))
				b.WriteString("(" + c.Reason + ") ")
			}
		}
	})
	return b.String()
}

// TestHeuristicFinishesPondersShuffleChain — reorder, then the shuffle
// question, then done, with the draw landing either way.
func TestHeuristicFinishesPondersShuffleChain(t *testing.T) {
	g := newSettledTable(t, 4)
	seat := g.Seats[0]
	handBefore := seat.Hand.Size()

	g.WithWriteLock(func() {
		g.LookAtTopThenForEffect(seat.ID, uuid.Nil, 3, func(g *game.Game) error {
			g.QueueConfirmForEffect(game.ConfirmPrompt{
				Chooser:      seat.ID,
				Question:     "Ponder — shuffle your library?",
				AcceptLabel:  "Shuffle",
				DeclineLabel: "Keep that order",
				OnAccept: func(g *game.Game) error {
					if err := g.ShuffleLibraryForEffect(seat.ID); err != nil {
						return err
					}
					return g.DrawNForEffect(seat.ID, 1)
				},
				OnDecline: func(g *game.Game) error {
					return g.DrawNForEffect(seat.ID, 1)
				},
			})
			return nil
		})
	})

	took := driveChoices(t, g, heuristic.New(), seat.ID, 8)
	if len(took) != 2 {
		t.Errorf("took %d answers, want 2 (the reorder and the shuffle): %v", len(took), took)
	}
	if seat.Hand.Size() != handBefore+1 {
		t.Errorf("hand is %d, want %d — the draw at the end of the chain never ran",
			seat.Hand.Size(), handBefore+1)
	}
}

// TestHeuristicFinishesSylvanLibrarysThreeLinkChain is the whole card,
// bot-side: pick two cards, then answer a pay-or-put-back question per
// card, each queued by the answer before it.
func TestHeuristicFinishesSylvanLibrarysThreeLinkChain(t *testing.T) {
	g := newSettledTable(t, 9)
	seat := g.Seats[0]

	// Stand in for the trigger's resolution: draw two, then ask.
	var drawn []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(seat.ID, 2); err != nil {
			t.Fatalf("draw: %v", err)
		}
		drawn = g.CardsDrawnThisTurnFor(seat.ID)
	})
	if len(drawn) < 2 {
		t.Fatalf("seeded %d drawn cards, want at least 2", len(drawn))
	}
	handBefore := seat.Hand.Size()
	lifeBefore := seat.Life
	libBefore := seat.Library.Size()

	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  seat.ID,
			Question: "Sylvan Library — choose two cards drawn this turn",
			Cards:    drawn,
			Min:      2,
			Max:      2,
			Zone:     game.ZoneHand,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				return sylvanChainForTest(g, seat.ID, picked)
			},
		})
	})

	took := driveChoices(t, g, heuristic.New(), seat.ID, 12)
	// One pick plus one question per chosen card.
	if len(took) != 3 {
		t.Errorf("took %d answers, want 3: %v", len(took), took)
	}
	// Whatever the policy chose, the books balance: every card either
	// stayed in hand and cost 4 life, or went back to the library.
	paid := (lifeBefore - seat.Life) / 4
	putBack := seat.Library.Size() - libBefore
	if paid+putBack != 2 {
		t.Errorf("paid for %d and put back %d, want 2 cards accounted for", paid, putBack)
	}
	if seat.Hand.Size() != handBefore-putBack {
		t.Errorf("hand is %d, want %d", seat.Hand.Size(), handBefore-putBack)
	}
}

// TestHeuristicWillNotPayItsLastLifeToAChainedPrompt — the hole #547
// closed for activated abilities, reached through a prompt instead. A
// confirm declares what its accept branch charges, the enumerator
// stamps it on the move, and the policy prices it: a bot on 4 life
// answering "pay 4 life to keep this card" would be answering "lose the
// game", and it must take the other branch instead.
func TestHeuristicWillNotPayItsLastLifeToAChainedPrompt(t *testing.T) {
	g := newSettledTable(t, 11)
	seat := g.Seats[0]
	paid := false
	g.WithWriteLock(func() {
		seat.Life = 4
		g.QueueConfirmForEffect(game.ConfirmPrompt{
			Chooser:      seat.ID,
			Question:     "Sylvan Library — pay 4 life to keep it?",
			AcceptLabel:  "Pay 4 life",
			DeclineLabel: "Put it on top",
			LifeCost:     4,
			OnAccept: func(*game.Game) error {
				paid = true
				return nil
			},
		})
	})

	took := driveChoices(t, g, heuristic.New(), seat.ID, 4)
	if paid {
		t.Errorf("the bot paid its last 4 life; it took %v", took)
	}
	if len(took) != 1 || !strings.Contains(took[0], "Put it on top") {
		t.Errorf("took %v, want the decline branch", took)
	}
}

// sylvanChainForTest mirrors the card's per-card chain without
// importing the effects package (which would import legal's catalog
// hooks into this test's dependency graph for no benefit).
func sylvanChainForTest(g *game.Game, controller uuid.UUID, remaining []uuid.UUID) error {
	if len(remaining) == 0 {
		return nil
	}
	card, rest := remaining[0], remaining[1:]
	putBack := func(g *game.Game) error {
		if err := g.TuckToLibraryForEffect(card, false); err != nil {
			return err
		}
		return sylvanChainForTest(g, controller, rest)
	}
	g.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      controller,
		Question:     "Sylvan Library — pay 4 life to keep it?",
		AcceptLabel:  "Pay 4 life",
		DeclineLabel: "Put it on top",
		OnAccept: func(g *game.Game) error {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, controller, -4); err != nil {
				return err
			}
			return sylvanChainForTest(g, controller, rest)
		},
		OnDecline: putBack,
	})
	return nil
}

// --- #798: which card a bot gives up --------------------------------
//
// The three shapes an effect discard reaches a bot in, end to end:
// Mind Rot's forced count, a loot (draw, then discard), and a rummage
// (discard, then draw). All three are one PendingChoiceChooseCards
// over the discarder's own hand since #797, so what is being tested is
// the same policy rule three times against the real engine — the bot
// names the card it wants least, not the first one in hand order.

// botCard mints a card owned and controlled by one seat.
func botCard(owner uuid.UUID, name, typeLine, cost string, power, tough int) game.Card {
	c := game.NewCard(name, owner)
	c.TypeLine = typeLine
	c.ManaCost = cost
	c.Power, c.Toughness = power, tough
	// A card dealt straight into a zone has no knowers, and a seat
	// that is not a knower of its own hand card reads it as a blank
	// (S13.5). The draw path does this for every card a player draws;
	// a test that skips it hands the policy three anonymous cards and
	// proves nothing about which one it would pitch.
	c.AddKnower(owner)
	return c
}

// seedDiscardBoard gives the seat a known hand, worst card LAST so
// that hand order and card value disagree, and puts six untapped
// Mountains on the battlefield: past Config.LandsWanted a seventh land
// is the cheapest card the bot holds, which is what makes the answer
// predictable without pinning the tuning numbers themselves.
//
// Caller must hold the write lock.
func seedDiscardBoard(g *game.Game, seat *game.Player) {
	seat.Hand.Cards = nil
	seat.Hand.PushTop(botCard(seat.ID, "Dragon", "Creature — Dragon", "{4}{R}{R}", 6, 6))
	seat.Hand.PushTop(botCard(seat.ID, "Bear", "Creature — Bear", "{1}{R}", 2, 2))
	seat.Hand.PushTop(botCard(seat.ID, "Mountain", "Basic Land — Mountain", "", 0, 0))
	for i := 0; i < 6; i++ {
		g.Battlefield.PushTop(botCard(seat.ID, "Mountain", "Basic Land — Mountain", "", 0, 0))
	}
}

// zoneNames reads the names in one of a seat's zones under the read
// lock — the policy and the engine are both done by the time it runs,
// but the lock is the contract.
func zoneNames(g *game.Game, z *game.Zone) []string {
	var out []string
	g.ReadSnapshot(func() {
		for _, c := range z.Cards {
			out = append(out, c.Name)
		}
	})
	return out
}

func TestHeuristicDiscardsItsWorstCardNotItsFirst(t *testing.T) {
	g := newSettledTable(t, 17)
	seat := g.Seats[0]
	g.WithWriteLock(func() {
		seedDiscardBoard(g, seat)
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   seat.ID,
			N:        1,
			Question: "Mind Rot — discard a card",
		})
	})

	took := driveChoices(t, g, heuristic.New(), seat.ID, 4)
	if len(took) != 1 {
		t.Fatalf("took %d answers, want 1: %v", len(took), took)
	}
	if got := zoneNames(g, seat.Graveyard); len(got) != 1 || got[0] != "Mountain" {
		t.Errorf("graveyard %v, want the Mountain — the bot pitched in hand order", got)
	}
	if got := zoneNames(g, seat.Hand); len(got) != 2 || got[0] != "Dragon" || got[1] != "Bear" {
		t.Errorf("hand %v, want the Dragon and the Bear kept", got)
	}
}

// "Discard up to two cards": the engine's floor is zero, so the bot
// owes nothing and keeps everything. The prompt still has to be
// answered — a seat owing a choice is offered nothing else (#544).
func TestHeuristicDiscardsNoMoreThanItMustWhenThePromptSaysUpTo(t *testing.T) {
	g := newSettledTable(t, 19)
	seat := g.Seats[0]
	g.WithWriteLock(func() {
		seedDiscardBoard(g, seat)
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   seat.ID,
			N:        2,
			UpTo:     true,
			Question: "Rummage — discard up to 2 cards",
		})
	})

	driveChoices(t, g, heuristic.New(), seat.ID, 4)
	if got := zoneNames(g, seat.Graveyard); len(got) != 0 {
		t.Errorf("graveyard %v, want nothing discarded for an 'up to' prompt", got)
	}
	if got := zoneNames(g, seat.Hand); len(got) != 3 {
		t.Errorf("hand %v, want all three cards kept", got)
	}
}

// A loot draws first, so the card it just drew is in the hand the
// prompt is built from: the bot keeps the better of the two and pitches
// the other. Here the drawn card is the worst thing it holds, which is
// the case a "discard what you just drew" bot gets wrong.
func TestHeuristicLootKeepsTheBetterCard(t *testing.T) {
	g := newSettledTable(t, 21)
	seat := g.Seats[0]
	g.WithWriteLock(func() {
		seedDiscardBoard(g, seat)
		// Two cards in hand, and the third drawn off the top by the
		// loot itself.
		seat.Hand.Cards = seat.Hand.Cards[:2]
		seat.Library.PushTop(botCard(seat.ID, "Mountain", "Basic Land — Mountain", "", 0, 0))
		if err := g.DrawNForEffect(seat.ID, 1); err != nil {
			t.Fatalf("draw: %v", err)
		}
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   seat.ID,
			N:        1,
			Question: "Faithless Looting — discard a card",
		})
	})

	driveChoices(t, g, heuristic.New(), seat.ID, 4)
	if got := zoneNames(g, seat.Graveyard); len(got) != 1 || got[0] != "Mountain" {
		t.Errorf("graveyard %v, want the land it drew", got)
	}
	if got := zoneNames(g, seat.Hand); len(got) != 2 || got[0] != "Dragon" || got[1] != "Bear" {
		t.Errorf("hand %v, want the two spells kept", got)
	}
}

// A rummage discards and THEN draws: the prompt's continuation is the
// draw, so answering it has to leave the bot one card better off with
// its worst card in the graveyard.
func TestHeuristicRummageDiscardsTheWorstCardThenDraws(t *testing.T) {
	g := newSettledTable(t, 23)
	seat := g.Seats[0]
	var libBefore int
	g.WithWriteLock(func() {
		seedDiscardBoard(g, seat)
		libBefore = seat.Library.Size()
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   seat.ID,
			N:        1,
			Question: "Syphon Mind — discard a card",
			Then:     func(g *game.Game) error { return g.DrawNForEffect(seat.ID, 1) },
		})
	})

	driveChoices(t, g, heuristic.New(), seat.ID, 4)
	if got := zoneNames(g, seat.Graveyard); len(got) != 1 || got[0] != "Mountain" {
		t.Errorf("graveyard %v, want the Mountain", got)
	}
	if got := zoneNames(g, seat.Hand); len(got) != 3 {
		t.Errorf("hand %v, want three cards — two kept and one drawn", got)
	}
	if got := seat.Library.Size(); got != libBefore-1 {
		t.Errorf("library %d, want %d — the draw after the discard never ran", got, libBefore-1)
	}
}
