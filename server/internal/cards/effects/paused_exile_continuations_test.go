package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_exile_continuations_test.go — #894. Two single-card exile
// loops that ran on past a leg the CR 903.9 prompt had PAUSED.
//
// Living Death is the one that lost a card: its exile / sacrifice /
// reanimate swap moved on while a commander card's owner was being
// asked about the command zone, so when they declined, that commander
// arrived in exile with the "put all cards they exiled this way onto
// the battlefield" step already finished — stranded, permanently.
// Path to Exile is the one that only looked wrong: it offered the
// basic-land search while the same question was still on the table.
//
// The shared answer is the continuation the exile already has: the
// batch for Living Death (ExileCardsThenForEffect, which reports the
// cards that really reached exile — CR 400.7), ExileTarget.Then for
// the single-card cases.

// livingDeathBoard seats the swap this file tests: `opp` has a
// commander creature card and a plain creature card in their
// graveyard, and `me` has one creature on the battlefield. Returns the
// commander card and the plain card.
func livingDeathBoard(t *testing.T, g *game.Game, me, opp *game.Player) (commander, dead, alive uuid.UUID) {
	t.Helper()
	commander = b17GraveyardCard(opp, "Their Commander", "Legendary Creature — Human", "{2}{B}")
	markCommanderCard(t, g, opp, commander)
	dead = pushGraveyardCardForTest(opp, "Their Dead")
	alive = seedCreature(g, "My Alive", me.ID)
	return commander, dead, alive
}

// TestLivingDeathWaitsForACommanderCardsExileAnswer is the issue. The
// swap holds until the CR 903.9 question is answered, and BOTH answers
// leave a board with the commander somewhere real.
func TestLivingDeathWaitsForACommanderCardsExileAnswer(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{{"to the command zone", true}, {"to exile, and back", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			commander, dead, alive := livingDeathBoard(t, g, me, opp)

			castCatalogSpell(t, g, "Living Death", "Sorcery", b03LivingDeathOracle, nil)
			passPriorityAroundTable(t, g)

			// Nothing past the exile has happened: the sacrifice and
			// the reanimation are both in the continuation.
			if !g.Battlefield.Contains(alive) {
				t.Fatal("the sacrifice waits for the paused exile leg")
			}
			if countBattlefieldNamed(g, opp.ID, "Their Dead") != 0 {
				t.Fatal("nothing is reanimated while the CR 903.9 prompt is open")
			}

			if tc.commandZone {
				b36AcceptCommandZone(t, g, opp.ID)
			} else {
				b21DeclineCommandZone(t, g, opp.ID)
			}

			if g.Battlefield.Contains(alive) {
				t.Error("the living creature is sacrificed once the prompt is answered")
			}
			if !me.Graveyard.Contains(alive) {
				t.Error("the sacrificed creature stays in the graveyard — it was not exiled this way")
			}
			if countBattlefieldNamed(g, opp.ID, "Their Dead") != 1 {
				t.Errorf("the plain graveyard creature comes back; battlefield holds %d",
					countBattlefieldNamed(g, opp.ID, "Their Dead"))
			}
			if g.Exile.Contains(dead) {
				t.Error("nothing that was exiled this way is left in exile")
			}

			if tc.commandZone {
				if !opp.Command.Contains(commander) {
					t.Error("the commander took the offer and is in the command zone")
				}
				if countBattlefieldNamed(g, opp.ID, "Their Commander") != 0 {
					t.Error(`a commander that went to the command zone was not "exiled this way", ` +
						"so it does not come back (CR 400.7)")
				}
				return
			}
			if g.Exile.Contains(commander) {
				t.Error("the declined commander is NOT stranded in exile — this is the bug")
			}
			if countBattlefieldNamed(g, opp.ID, "Their Commander") != 1 {
				t.Error("declining exiles the commander card and the swap puts it onto the battlefield " +
					"with everything else it exiled this way")
			}
		})
	}
}

// TestLivingDeathUndoAcrossTheCommanderPromptReplays — the undo
// contract the batch signs, on the card: rewind into the open prompt,
// answer the OTHER way, and the board follows that answer.
func TestLivingDeathUndoAcrossTheCommanderPromptReplays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commander, _, _ := livingDeathBoard(t, g, me, opp)
	oppID := opp.ID

	castCatalogSpell(t, g, "Living Death", "Sorcery", b03LivingDeathOracle, nil)
	passPriorityAroundTable(t, g)
	promptOpen := g.Clone()

	b36AcceptCommandZone(t, g, oppID)
	if !g.Seats[1].Command.Contains(commander) {
		t.Fatal("accepting puts the commander in the command zone")
	}

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Seats[1].Graveyard.Contains(commander) {
		t.Fatal("the rewind puts the commander card back in its graveyard")
	}

	b21DeclineCommandZone(t, g, oppID)
	if countBattlefieldNamed(g, oppID, "Their Commander") != 1 {
		t.Error("the replayed answer exiles the commander and the swap brings it back")
	}
	if g.Exile.Contains(commander) {
		t.Error("nothing is stranded in exile on the replay either")
	}
}

// TestPathToExileOffersTheSearchOnlyAfterTheExileAnswer is the
// ordering half of the issue. The outcome was always right; what was
// wrong is that the victim was asked to search while their own
// CR 903.9 question was still open.
func TestPathToExileOffersTheSearchOnlyAfterTheExileAnswer(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	commander := b36Commander(g, victim.ID, "Their Commander")
	forest := pushLibraryCardForTest(victim, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	castCatalogSpell(t, g, "Path to Exile", "Instant",
		"d683d985-9888-4d21-8b5f-69e69ce4a03b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: commander}},
	)
	passPriorityAroundTable(t, g)

	if searchChoiceFor(g, victim.ID) != nil {
		t.Fatal("no search is offered while the CR 903.9 prompt is open — the exile is still being answered")
	}
	b21DeclineCommandZone(t, g, victim.ID)

	if !g.Exile.Contains(commander) {
		t.Fatal("declining exiles the commander")
	}
	if searchChoiceFor(g, victim.ID) == nil {
		t.Fatal("the search is offered once the exile has landed")
	}
	answerSearchByID(t, g, victim.ID, forest)
	if !g.Battlefield.Contains(forest) {
		t.Error("and it still fetches the basic land")
	}
}

// TestPathToExileStillSearchesWhenTheCommanderTakesTheCommandZone —
// the search is a separate sentence, not an "if you do", so it is
// offered whichever way the question is answered. Only its order
// changed.
func TestPathToExileStillSearchesWhenTheCommanderTakesTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	commander := b36Commander(g, victim.ID, "Their Commander")
	pushLibraryCardForTest(victim, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	castCatalogSpell(t, g, "Path to Exile", "Instant",
		"d683d985-9888-4d21-8b5f-69e69ce4a03b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: commander}},
	)
	passPriorityAroundTable(t, g)
	b36AcceptCommandZone(t, g, victim.ID)

	if !victim.Command.Contains(commander) {
		t.Fatal("the commander took the offer")
	}
	if searchChoiceFor(g, victim.ID) == nil {
		t.Error("the search is still offered — the exile's outcome does not gate it")
	}
}

// TestFlickerReturnsACommanderOnlyOnceItsExileLands — the Flicker
// primitive (Y'shtola Rhul) is Living Death's shape with one card:
// exile it, then put that same card back. The return used to run with
// the CR 903.9 prompt open, find nothing in exile, and leave the
// commander there for good.
func TestFlickerReturnsACommanderOnlyOnceItsExileLands(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
		wantBack    int
	}{
		{"to the command zone", true, 0},
		{"to exile, and back", false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			commander := b36Commander(g, me.ID, "My Commander")

			g.WithWriteLock(func() {
				item := &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()}
				if err := (Flicker{Target: commander}).Apply(ctxFor(g, item)); err != nil {
					t.Fatalf("Flicker: %v", err)
				}
			})
			if !g.Battlefield.Contains(commander) {
				t.Fatal("nothing has moved yet: the flicker's exile is waiting on the CR 903.9 answer")
			}

			if tc.commandZone {
				b36AcceptCommandZone(t, g, me.ID)
			} else {
				b21DeclineCommandZone(t, g, me.ID)
			}

			if g.Battlefield.Contains(commander) {
				t.Error("the flicker really happened: the pre-exile object is gone (CR 400.7)")
			}
			if got := countBattlefieldNamed(g, me.ID, "My Commander"); got != tc.wantBack {
				t.Errorf("battlefield holds %d copies of the commander, want %d", got, tc.wantBack)
			}
			if g.Exile.Contains(commander) {
				t.Error("the flickered commander is not left in exile")
			}
			if tc.commandZone && !me.Command.Contains(commander) {
				t.Error("a commander that took the offer is in the command zone, not on the battlefield")
			}
		})
	}
}
