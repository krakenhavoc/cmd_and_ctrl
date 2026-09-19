package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exit_payout_cards_test.go — #993 on the five cards the widened lint
// found, resolved through the real priority loop.
//
// exile_payout_cards_test.go is the same file for #911's exile, and the
// boards here are its boards with a different verb: a CR 903.9 prompt
// held open by a commander, and a CR 614 window that cancels the move
// outright. What each card owes on those boards is the whole of the
// triage, and it is not the same answer for all five:
//
//   - Boomerang Basics and Chain of Vapor owe only ORDER. Their second
//     sentence is about a PLAYER and is not introduced by an "if you
//     do", so it happens whichever way the question is answered — the
//     Path to Exile reading (ADR 0013 §5m item 2, §5v). What must not
//     happen is the draw, or the next question, arriving while the
//     first question is still on the table.
//   - Ruthless Technomancer owes the ANSWER. "If you do, create a
//     number of Treasure tokens equal to that creature's power" is
//     gated on the sacrifice, and the card used to read the live board
//     on the next line — which says "still on the battlefield, so not
//     sacrificed" for a leg that is merely paused.
//   - Hermit Druid and Consuming Aberration owe SEQUENCE. Both used to
//     write the next instruction on the line after a mill that can
//     pause; the Aberration's loop then re-read the top of the library
//     and milled the same paused card again.

// --- Boomerang Basics ----------------------------------------------

// TestBoomerangBasicsWaitsForTheCommandZoneAnswer is the ordering half.
// The draw is the bounce's continuation, so it cannot land while the
// owner is still being asked about the command zone.
//
// Both answers draw, and that is the rules call (#993, ADR 0013 §5v):
// "if you controlled that permanent" is a condition about the PLAYER's
// past relationship to the spell's target, true before anything moves
// and beyond any replacement's reach. A commander of yours that took
// CR 903.9's offer was still yours.
func TestBoomerangBasicsWaitsForTheCommandZoneAnswer(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{
		{"to the command zone: the draw still happens", true},
		{"to the hand: the card and the draw", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			mine := pushVanillaCreature(g, me.ID, "My Commander", 1, 1)
			markCommanderCard(t, g, me, mine)

			castCatalogSpell(t, g, "Boomerang Basics", "Sorcery — Lesson",
				b42BoomerangBasicsOracle, b42CardTarget(mine))
			handAfterCast := me.Hand.Size()
			passPriorityAroundTable(t, g)

			if got := me.Hand.Size(); got != handAfterCast {
				t.Fatalf("hand %d → %d while the CR 903.9 prompt is open, want unchanged — "+
					"the draw is the return's continuation", handAfterCast, got)
			}
			if !g.Battlefield.Contains(mine) {
				t.Fatal("nothing moves until the command-zone question is answered")
			}

			if tc.commandZone {
				b36AcceptCommandZone(t, g, me.ID)
				if !me.Command.Contains(mine) {
					t.Fatal("accepting puts the commander in the command zone")
				}
				if got := me.Hand.Size(); got != handAfterCast+1 {
					t.Errorf("hand %d → %d, want +1 (the draw only) — you still controlled that "+
						"permanent, so the clause is true however the return was replaced",
						handAfterCast, got)
				}
				return
			}
			b21DeclineCommandZone(t, g, me.ID)
			if !me.Hand.Contains(mine) {
				t.Fatal("declining returns the permanent to its owner's hand")
			}
			if got := me.Hand.Size(); got != handAfterCast+2 {
				t.Errorf("hand %d → %d, want +2 (the returned card and the draw)", handAfterCast, got)
			}
		})
	}
}

// TestBoomerangBasicsStillDrawsNothingForAnOpponentsPermanent is the
// control: the continuation must not have turned the rider into an
// unconditional cantrip.
func TestBoomerangBasicsStillDrawsNothingForAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)

	castCatalogSpell(t, g, "Boomerang Basics", "Sorcery — Lesson",
		b42BoomerangBasicsOracle, b42CardTarget(theirs))
	handAfterCast := me.Hand.Size()
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(theirs) {
		t.Fatal("an opponent's permanent is returned too")
	}
	if got := me.Hand.Size(); got != handAfterCast {
		t.Errorf("hand %d → %d, want unchanged — you did not control that permanent", handAfterCast, got)
	}
}

// --- Chain of Vapor -------------------------------------------------

// TestChainOfVaporAsksTheChainAfterTheReturnSettles. The chain's first
// question goes to the bounced permanent's controller, and they are
// already answering one — a commander of theirs is asking about the
// command zone. Two prompts at once, in the wrong order, is the bug
// ADR 0013 §5m item 2 fixed for Path to Exile.
func TestChainOfVaporAsksTheChainAfterTheReturnSettles(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Commander", 1, 1)
	markCommanderCard(t, g, opp, theirs)
	seedLandOnBattlefield(g, opp.ID, "Their Swamp", "Basic Land — Swamp")

	castCatalogSpell(t, g, "Chain of Vapor", "Instant", chainOfVaporOracle, b42CardTarget(theirs))
	passPriorityAroundTable(t, g)

	if n := countPendingFor(g, opp.ID, game.PendingChoiceConfirm); n != 0 {
		t.Fatalf("%d confirm prompts for the permanent's controller while the CR 903.9 question "+
			"is still open, want 0 — the chain is the return's continuation", n)
	}
	b21DeclineCommandZone(t, g, opp.ID)
	if !opp.Hand.Contains(theirs) {
		t.Fatal("declining returns the permanent to its owner's hand")
	}
	if n := countPendingFor(g, opp.ID, game.PendingChoiceConfirm); n != 1 {
		t.Errorf("%d confirm prompts once the return has settled, want 1 — \"then that permanent's "+
			"controller may sacrifice a land\"", n)
	}
}

// TestChainOfVaporStillAsksWhenTheReturnWasReplacedAway is the other
// half of the same rules call: the second sentence is about a PLAYER
// and is not an "if you do", so a commander that took the command zone
// still leaves its controller being asked about a land.
func TestChainOfVaporStillAsksWhenTheReturnWasReplacedAway(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Commander", 1, 1)
	markCommanderCard(t, g, opp, theirs)
	seedLandOnBattlefield(g, opp.ID, "Their Swamp", "Basic Land — Swamp")

	castCatalogSpell(t, g, "Chain of Vapor", "Instant", chainOfVaporOracle, b42CardTarget(theirs))
	passPriorityAroundTable(t, g)
	b36AcceptCommandZone(t, g, opp.ID)

	if !opp.Command.Contains(theirs) {
		t.Fatal("accepting puts the commander in the command zone")
	}
	if n := countPendingFor(g, opp.ID, game.PendingChoiceConfirm); n != 1 {
		t.Errorf("%d confirm prompts, want 1 — the chain is not gated on the return", n)
	}
}

// TestChainOfVaporCopyQuestionWaitsForTheSacrificedLand is the
// SacrificeChoice half of the same fix. Its doc says "Then runs once
// the permanent has gone", and the prompt's continuation used to call
// the fire-and-forget sacrifice and then run the clause on the next
// line — so a sacrificed commander LAND had its owner offered the copy
// while they were still being asked about the command zone.
func TestChainOfVaporCopyQuestionWaitsForTheSacrificedLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	land := pushLand(g, opp.ID, "Their Island")
	markCommanderCard(t, g, opp, land)

	castCatalogSpell(t, g, "Chain of Vapor", "Instant", chainOfVaporOracle, b42CardTarget(theirs))
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, opp.ID, true)

	sac := latestChooseCardsFor(g, opp.ID)
	if sac == nil {
		t.Fatalf("they pick which land to sacrifice: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	if n := countPendingFor(g, opp.ID, game.PendingChoiceConfirm); n != 0 {
		t.Fatalf("%d copy questions while the sacrificed land's CR 903.9 prompt is open, want 0 — "+
			"SacrificeChoice.Then runs once the permanent has gone", n)
	}
	b21DeclineCommandZone(t, g, opp.ID)
	if g.Battlefield.Contains(land) {
		t.Fatal("declining sends the sacrificed land to its owner's graveyard")
	}
	if n := countPendingFor(g, opp.ID, game.PendingChoiceConfirm); n != 1 {
		t.Errorf("%d copy questions once the sacrifice has settled, want 1", n)
	}
}

// --- Ruthless Technomancer ------------------------------------------

// TestRuthlessTechnomancerPaysForASacrificedCommander is the clause
// that IS gated, and the outcome the old live-board read got wrong. A
// sacrificed commander sits on the battlefield while its owner answers
// CR 903.9, so "is it still on the battlefield?" said "not sacrificed"
// and no Treasure was ever made.
//
// Both answers pay, and CR 701.17a is why: the keyword action is the
// controller's move OFF the battlefield, and a replacement rewrites
// only where the permanent goes.
func TestRuthlessTechnomancerPaysForASacrificedCommander(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{
		{"to the command zone: still sacrificed", true},
		{"to the graveyard: sacrificed", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			victim := b12Creature(g, me.ID, "Dear Friend", "Legendary Creature — Human", 3, 3)
			markCommanderCard(t, g, me, victim)

			technomancerETB(t, g, me, victim)

			if n, _ := b14Tokens(g, me.ID, "Treasure"); n != 0 {
				t.Fatalf("%d Treasures while the CR 903.9 prompt is open, want 0 — the clause "+
					"waits for the answer", n)
			}
			if tc.commandZone {
				b36AcceptCommandZone(t, g, me.ID)
				if !me.Command.Contains(victim) {
					t.Fatal("accepting puts the commander in the command zone")
				}
			} else {
				b21DeclineCommandZone(t, g, me.ID)
				if !me.Graveyard.Contains(victim) {
					t.Fatal("declining puts it in its owner's graveyard")
				}
			}
			if n, _ := b14Tokens(g, me.ID, "Treasure"); n != 3 {
				t.Errorf("%d Treasures, want 3 — the creature's power, and it WAS sacrificed "+
					"(CR 701.17a: only where it went was replaced)", n)
			}
		})
	}
}

// TestRuthlessTechnomancerPaysNothingForACancelledSacrifice is the
// other side of the gate: the window kept the creature on the
// battlefield, so nothing was sacrificed and "if you do" is false.
func TestRuthlessTechnomancerPaysNothingForACancelledSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := b12Creature(g, me.ID, "Stubborn Friend", "Creature — Human", 3, 3)
	registerExitReplacement(t, g, victim, "it can't leave the battlefield",
		func(ev *game.ReplacementEvent) { ev.Cancel() })

	technomancerETB(t, g, me, victim)

	if !g.Battlefield.Contains(victim) {
		t.Fatal("control: the cancelled move leaves the creature on the battlefield")
	}
	if n, _ := b14Tokens(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("%d Treasures, want 0 — nothing was sacrificed, so \"if you do\" is false", n)
	}
}

// technomancerETB resolves a Ruthless Technomancer under `me` with its
// enter trigger accepted and `victim` chosen, and settles the table up
// to whatever prompt the sacrifice itself opens.
func technomancerETB(t *testing.T, g *game.Game, me *game.Player, victim uuid.UUID) {
	t.Helper()
	castAndResolveCreature(t, g, "Ruthless Technomancer", "Creature — Human Wizard",
		b19RuthlessTechnomancerOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
}

// --- Hermit Druid ---------------------------------------------------

// TestHermitDruidPutsTheLandInHandAfterTheMillSettles. "Put that card
// into your hand and all other cards revealed this way into your
// graveyard" is one instruction about a settled run, and one of the
// cards above the land is a commander whose owner has been asked about
// the command zone. The land used to reach hand while that question was
// open and cards above it were still in the library.
func TestHermitDruidPutsTheLandInHandAfterTheMillSettles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	druid := pushCatalogPermanent(g, me.ID, "Hermit Druid", "Creature — Human Druid",
		b21HermitDruidOracle, false)

	// Top of the library downward: a commander card, a spell, then the
	// basic land that ends the run.
	b10LibraryTop(me, "My Swamp", "Basic Land — Swamp", "", 0, 0)
	b10LibraryTop(me, "Filler", "Instant", "{U}", 0, 0)
	commander := b10LibraryTop(me, "My Commander", "Legendary Creature — Human", "{2}{B}", 2, 2)
	markCommanderCard(t, g, me, commander)
	land := me.Library.Cards[len(me.Library.Cards)-3].InstanceID

	advanceToMain(t, g)
	b06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Hermit Druid: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Hand.Contains(land) {
		t.Fatal("the basic land is in hand while the CR 903.9 prompt is still open — the hand-off " +
			"is the mill's continuation")
	}
	b21DeclineCommandZone(t, g, me.ID)

	if !me.Hand.Contains(land) {
		t.Errorf("the basic land should be in hand once the mill has settled; it is in %s",
			b12ZoneOf(g, land))
	}
	if !me.Graveyard.Contains(commander) {
		t.Error("declining puts the commander card in its owner's graveyard with the rest of the run")
	}
}

// --- Consuming Aberration -------------------------------------------

// TestConsumingAberrationDoesNotReMillAPausedCommander is the loop that
// spun. The old body read the top of the library, milled one card, and
// asked whether the card it had read was a land; a milled commander's
// prompt leaves the card exactly where it was, so the next pass read
// the same card and milled it again — round and round until the fuse
// blew or somebody answered.
//
// The run is chosen up front now, so the commander is one leg of it and
// the rest of the run waits behind its answer.
func TestConsumingAberrationDoesNotReMillAPausedCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b12Push(g, me.ID, "Consuming Aberration", "Creature — Horror", b14ConsumingAberrationOracle, 0, 0)
	// The Aberration is a */* equal to the cards in opponents'
	// graveyards, so it needs one to survive the CR 704.5f check long
	// enough to trigger at all.
	pushGraveyardCardForTest(opp, "Dead A")

	// Top downward: a commander card, a spell, then the land.
	b10LibraryTop(opp, "Their Land", "Basic Land — Swamp", "", 0, 0)
	b10LibraryTop(opp, "Their Spell", "Instant", "{U}", 0, 0)
	commander := b10LibraryTop(opp, "Their Commander", "Legendary Creature — Human", "{2}{B}", 2, 2)
	markCommanderCard(t, g, opp, commander)
	before := opp.Library.Size()

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	if !opp.Library.Contains(commander) {
		t.Fatal("nothing moves for the paused leg until the command-zone question is answered")
	}
	b21DeclineCommandZone(t, g, opp.ID)

	if got := before - opp.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3 (up to and including the land) — a paused leg must not shorten "+
			"the run, and must not be milled twice", got)
	}
	if !opp.Graveyard.Contains(commander) {
		t.Error("declining puts the commander card in its owner's graveyard")
	}
	if n := countInZone(opp.Graveyard, commander); n != 1 {
		t.Errorf("the commander card is in the graveyard %d times, want 1 — the loop used to "+
			"re-read the top of the library and mill the same paused card again", n)
	}
}

// --- local helpers ---------------------------------------------------

// countPendingFor counts the pending choices of one kind addressed to
// one seat.
func countPendingFor(g *game.Game, chooser uuid.UUID, kind game.PendingChoiceKind) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind && c.Chooser == chooser {
			n++
		}
	}
	return n
}

// countInZone counts how many times one instance ID appears in a zone.
// A card can only be in a zone once; the count is what makes "milled
// twice" a failure rather than a silent pass.
func countInZone(z *game.Zone, id uuid.UUID) int {
	if z == nil {
		return 0
	}
	n := 0
	for _, c := range z.Cards {
		if c.InstanceID == id {
			n++
		}
	}
	return n
}
