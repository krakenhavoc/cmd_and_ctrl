package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// blocking_pay_unless_test.go — #567, rebuilt on #997's door. A
// pay_unless prompt can stop the table even though its KIND does not:
// the upkeep pay-or-else (cumulative upkeep's "sacrifice this unless
// you pay", CR 702.24; Stasis; Pact of Negation) is owed before the
// step it was asked in can end.
//
// #794's lesson is that the engine and the enumerator must answer
// "does this stop the table" from ONE function. They now both read
// game.ChoicePromptBlocksTable, and this pins that they agree about a
// per-prompt narrowing as well as about a kind. The prompt is queued to
// a NON-active seat both times, so what is being measured is whether
// the seat holding priority can still act.
func TestPayUnlessBlocksTheTableOnlyWhenTheContractSaysSo(t *testing.T) {
	for _, tc := range []struct {
		name     string
		blocking bool
	}{
		{"an upkeep pay-or-else blocks", true},
		{"a Rhystic tax does not", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			active, payer := g.Seats[0], g.Seats[1]
			g.WithWriteLock(func() {
				var err error
				if tc.blocking {
					err = g.QueueUpkeepPayUnlessForEffect(game.UpkeepPayUnlessPrompt{
						Chooser:   payer.ID,
						Source:    payer.ID,
						Cost:      "{1}",
						Question:  "pay {1}?",
						OnDecline: func(*game.Game) error { return nil },
					})
				} else {
					err = g.QueuePayUnlessForEffect(payer.ID, payer.ID, "{1}", "pay {1}?",
						func(*game.Game) error { return nil })
				}
				if err != nil {
					t.Fatalf("queue: %v", err)
				}
			})

			// The payer is asked their own question either way, and
			// only that — a seat owing a prompt is offered nothing
			// else, blocking or not (#794).
			mine := legal.EnumerateFor(g, payer.ID)
			if len(mine) == 0 {
				t.Fatal("the seat owing the prompt was offered nothing")
			}
			for _, m := range mine {
				if m.Type != legal.TypeResolveChoice {
					t.Errorf("offered %q (%s) while owing a pay_unless", m.Label, m.Type)
				}
			}
			dispatchAll(t, g, payer.ID, mine)

			// The seat holding priority is the one the block is about.
			moves := legal.EnumerateFor(g, active.ID)
			if tc.blocking && len(moves) != 0 {
				t.Errorf("the active seat was offered %v while a blocking pay_unless is open", labels(moves))
			}
			if !tc.blocking && len(moves) == 0 {
				t.Error("an ordinary pay_unless stopped the active seat — ADR 0018 §6 says it must not")
			}
		})
	}
}

// #1014 landed the five-zone cast walk — hand, command, graveyard,
// exile and the top of the library — and this is the one thing that
// walk must not do: run while a prompt is holding the table.
//
// It cannot, structurally: castMoves is reached only after choiceMoves
// and anyBlockingChoiceOpen have had their say (legal.go). The test is
// here rather than trusted because the two landed a day apart — the
// halt is #997's newest reason to stop the table, the walk is #1014's
// newest source of moves — and the failure it catches is a bot casting
// a flashback spell out of its graveyard while the question its own
// upkeep is waiting on goes unanswered.
//
// Each half runs a CONTROL with no prompt open first, because the
// offer has to be there to be suppressed; without that the assertions
// are satisfied by a seat that was never offered anything.

// upkeepTaxTable seeds a main phase (the graveyard offer under test is
// a sorcery, and #1014's walk would not reach it in an upkeep at all)
// with Faithless Looting in `caster`'s graveyard and the mana to flash
// it back. The halt does not care which step raised it — it holds the
// step that asked it, #997 — and is read through the same predicate
// either way, which is the point being made here.
func upkeepTaxTable(t *testing.T, casterSeat int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	for _, p := range g.Seats {
		clearHand(p)
	}
	advanceTo(t, g, game.StepPrecombatMain)
	caster := g.Seats[casterSeat]
	basicLands(g, caster, 3, "Mountain")
	looting := graveyardCard(caster, game.Card{
		Name: "Faithless Looting", TypeLine: "Sorcery", ManaCost: "{1}{R}",
		OracleID: oracleFaithlessLooting,
	})
	return g, caster, looting
}

func queueUpkeepTaxFor(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueueUpkeepPayUnlessForEffect(game.UpkeepPayUnlessPrompt{
			Chooser:   chooser,
			Source:    chooser,
			Cost:      "{1}",
			Question:  "pay {1} or sacrifice it",
			OnDecline: func(*game.Game) error { return nil },
		}); err != nil {
			t.Fatalf("QueueUpkeepPayUnlessForEffect: %v", err)
		}
	})
}

// The seat that OWES the prompt is offered the answer and nothing else
// — no cast, from any zone. #794's rule, now with five zones behind it.
func TestTheSeatOwingAnUpkeepTaxIsOfferedNoCastFromAnyZone(t *testing.T) {
	active := 0
	control, payer, looting := upkeepTaxTable(t, active)
	if payer != control.Seats[control.Turn.ActiveSeat] {
		t.Fatalf("setup: seat %d does not hold priority", active)
	}
	if n := len(castMovesFor(legal.EnumerateFor(control, payer.ID), looting)); n != 1 {
		t.Fatalf("control: the graveyard flashback was offered %d times, want 1", n)
	}

	g, payer, looting := upkeepTaxTable(t, active)
	queueUpkeepTaxFor(t, g, payer.ID)

	mine := legal.EnumerateFor(g, payer.ID)
	if len(mine) == 0 {
		t.Fatal("the seat owing the prompt was offered nothing")
	}
	for _, m := range mine {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("offered %q (%s) while owing an upkeep pay-or-else", m.Label, m.Type)
		}
	}
	if n := len(castMovesFor(mine, looting)); n != 0 {
		t.Errorf("the payer was offered %d graveyard casts while owing the prompt: %v",
			n, labels(mine))
	}
	dispatchAll(t, g, payer.ID, mine)
}

// And the seat that does NOT owe it is offered nothing at all — which
// is what makes this a halt rather than a question one player answers
// in the background. The prompt is queued to a seat that does not hold
// priority so that what is measured is the seat that does: the whole
// five-zone walk, suppressed.
func TestAnUpkeepTaxStopsEveryOtherSeatsCastWalk(t *testing.T) {
	active := 0
	control, caster, looting := upkeepTaxTable(t, active)
	if n := len(castMovesFor(legal.EnumerateFor(control, caster.ID), looting)); n != 1 {
		t.Fatalf("control: the graveyard flashback was offered %d times, want 1", n)
	}

	g, caster, _ := upkeepTaxTable(t, active)
	payer := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	queueUpkeepTaxFor(t, g, payer.ID)

	if moves := legal.EnumerateFor(g, caster.ID); len(moves) != 0 {
		t.Errorf("the seat holding priority was offered %v while an upkeep pay-or-else "+
			"it does not owe holds the table", labels(moves))
	}
	// The payer can still answer through the halt, which is what keeps
	// it from being a wedge.
	mine := legal.EnumerateFor(g, payer.ID)
	if len(mine) == 0 {
		t.Fatal("the seat owing the prompt was offered nothing to answer with")
	}
	dispatchAll(t, g, payer.ID, mine)
}
