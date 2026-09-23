package game

import (
	"testing"

	"github.com/google/uuid"
)

// sandbox_move_resume_test.go pins #707: the sandbox move_card verb
// (MoveCardByIDAsCommander) finishes its move after the CR 903.9
// prompt, from whatever zone the card was in when the window opened.
//
// CR 903.9 is "from ANYWHERE", and #539 made that true of the prompt:
// a commander moved out of a graveyard, a hand, a library or the stack
// by hand is asked whether it should go to the command zone instead.
// What the admin move did not have was a RESUME. Its own pipeline call
// had no continuation behind it, and the generic resume knew only how
// to finish a BATTLEFIELD exit, so both answers to the prompt did the
// same thing: nothing. The prompt closed, the card stayed where it was,
// and the move was lost.
//
// The fix is not a second resume — it is the shared exit primitive.
// Every destination that is not the battlefield or the stack now goes
// through routeCardToZoneLocked, whose zoneRoute frame carries the
// origin zone and everything the move asked for (the bottom of the
// library, the stack item to retire) across the pause, and whose
// resume lands it.

// sandboxOrigin describes one "commander leaves zone X by hand" case.
type sandboxOrigin struct {
	name string
	// seat puts the commander in its origin zone and returns its ID
	// and the ZoneRef naming that zone.
	seat func(t *testing.T, g *Game, owner *Player) (uuid.UUID, ZoneRef)
	// dst is the move the admin asks for.
	dst func(owner *Player) ZoneRef
	// declined is where the card must land when the owner says "no".
	declined func(g *Game, owner *Player) *Zone
}

func sandboxOrigins() []sandboxOrigin {
	return []sandboxOrigin{
		{
			name: "library to graveyard (a mill by hand)",
			seat: func(t *testing.T, g *Game, owner *Player) (uuid.UUID, ZoneRef) {
				return seatCommander(t, owner.Library, owner),
					ZoneRef{Kind: ZoneLibrary, Owner: owner.ID}
			},
			dst:      func(owner *Player) ZoneRef { return ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID} },
			declined: func(_ *Game, owner *Player) *Zone { return owner.Graveyard },
		},
		{
			name: "stack to graveyard (a counter by hand)",
			seat: func(t *testing.T, g *Game, owner *Player) (uuid.UUID, ZoneRef) {
				id := seatCommander(t, g.Stack, owner)
				spell, _ := g.Stack.Top()
				if _, err := g.Stack.Remove(id); err != nil {
					t.Fatalf("Stack.Remove: %v", err)
				}
				pushStackSpell(t, g, spell)
				return id, ZoneRef{Kind: ZoneStack}
			},
			dst:      func(owner *Player) ZoneRef { return ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID} },
			declined: func(_ *Game, owner *Player) *Zone { return owner.Graveyard },
		},
		{
			name: "hand to graveyard (a discard by hand)",
			seat: func(t *testing.T, g *Game, owner *Player) (uuid.UUID, ZoneRef) {
				return seatCommander(t, owner.Hand, owner),
					ZoneRef{Kind: ZoneHand, Owner: owner.ID}
			},
			dst:      func(owner *Player) ZoneRef { return ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID} },
			declined: func(_ *Game, owner *Player) *Zone { return owner.Graveyard },
		},
		{
			name: "graveyard to exile",
			seat: func(t *testing.T, g *Game, owner *Player) (uuid.UUID, ZoneRef) {
				return seatCommander(t, owner.Graveyard, owner),
					ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID}
			},
			dst:      func(_ *Player) ZoneRef { return ZoneRef{Kind: ZoneExile} },
			declined: func(g *Game, _ *Player) *Zone { return g.Exile },
		},
	}
}

// TestSandboxCommanderMoveFinishesFromEveryOriginZone is the issue.
// Both answers, from each of the four non-battlefield zones a
// commander can be moved out of by hand.
func TestSandboxCommanderMoveFinishesFromEveryOriginZone(t *testing.T) {
	for _, tc := range sandboxOrigins() {
		for _, toCommandZone := range []bool{true, false} {
			answer := "declined"
			if toCommandZone {
				answer = "command zone"
			}
			t.Run(tc.name+" / "+answer, func(t *testing.T) {
				g := newActiveGame(t)
				owner := g.Seats[0]
				cmdID, src := tc.seat(t, g, owner)
				dst := tc.dst(owner)

				if err := g.MoveCardByIDAsCommander(src, dst, cmdID, false); err != nil {
					t.Fatalf("MoveCardByIDAsCommander: %v", err)
				}
				// The prompt gates the move: nothing has left yet.
				srcZone := g.zoneFromRefLocked(src)
				if !srcZone.Contains(cmdID) {
					t.Fatalf("the commander left %s before the prompt was answered", srcZone.Kind)
				}
				prompt := expectCommanderPrompt(t, g, owner)

				if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, toCommandZone); err != nil {
					t.Fatalf("ResolveOptionalReplacement: %v", err)
				}
				if len(g.PendingChoices) != 0 {
					t.Fatalf("%d prompts survived the answer", len(g.PendingChoices))
				}
				want := owner.Command
				if !toCommandZone {
					want = tc.declined(g, owner)
				}
				others := []*Zone{srcZone}
				if want != owner.Command {
					others = append(others, owner.Command)
				}
				assertOnlyIn(t, cmdID, want, others...)
			})
		}
	}
}

// TestSandboxMoveOffTheStackRetiresTheStackItem — the per-zone detail
// the stack case owes. A card moved off the stack by hand is not a
// spell on the stack any more, whichever answer the prompt got, and
// the retirement has to survive the pause with the rest of the route.
func TestSandboxMoveOffTheStackRetiresTheStackItem(t *testing.T) {
	for _, toCommandZone := range []bool{true, false} {
		g := newActiveGame(t)
		owner := g.Seats[0]
		cmdID := seatCommander(t, g.Stack, owner)
		spell, _ := g.Stack.Top()
		if _, err := g.Stack.Remove(cmdID); err != nil {
			t.Fatalf("Stack.Remove: %v", err)
		}
		pushStackSpell(t, g, spell)

		if err := g.MoveCardByIDAsCommander(
			ZoneRef{Kind: ZoneStack},
			ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID},
			cmdID, false,
		); err != nil {
			t.Fatalf("MoveCardByIDAsCommander: %v", err)
		}
		if _, ok := g.StackMeta[cmdID]; !ok {
			t.Fatal("the stack item went away while the prompt was still open")
		}
		prompt := expectCommanderPrompt(t, g, owner)
		if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, toCommandZone); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
		if _, ok := g.StackMeta[cmdID]; ok {
			t.Errorf("StackMeta entry survived the move (answer to-command-zone=%v)", toCommandZone)
		}
	}
}

// TestSandboxTuckToBottomSurvivesTheCommanderPrompt — the other
// per-zone detail. "Library (bottom)" used to be a reorder that ran
// while the card was still in its old zone, so a paused commander
// landed on TOP of the library when its owner declined. The bottom now
// rides the route.
func TestSandboxTuckToBottomSurvivesTheCommanderPrompt(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, owner.Hand, owner)

	if err := g.MoveCardByIDToBottom(
		ZoneRef{Kind: ZoneHand, Owner: owner.ID},
		ZoneRef{Kind: ZoneLibrary, Owner: owner.ID},
		cmdID, false,
	); err != nil {
		t.Fatalf("MoveCardByIDToBottom: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if bottom, err := owner.Library.Bottom(); err != nil || bottom.InstanceID != cmdID {
		t.Errorf("the declined commander is not on the bottom of the library (err=%v)", err)
	}
}

// TestSandboxMoveOfANonCommanderStillMovesImmediately — the common
// case must not have grown a prompt or lost its state check.
func TestSandboxMoveOfANonCommanderStillMovesImmediately(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	owner.Hand.PushTop(Card{
		InstanceID: id,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneHand, Owner: owner.ID},
		ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID},
		id,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("an ordinary move queued %d prompts", len(g.PendingChoices))
	}
	assertOnlyIn(t, id, owner.Graveyard, owner.Hand)
}

// TestSandboxMoveOfAMissingCardIsStillAnError — the exit primitive
// finds the card by scan, so the src ref would be advisory without the
// guard. A stale request naming the zone the card has already left
// must fail rather than move it out of wherever it is now.
func TestSandboxMoveOfAMissingCardIsStillAnError(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := seatCommander(t, owner.Graveyard, owner)

	err := g.MoveCardByIDAsCommander(
		ZoneRef{Kind: ZoneHand, Owner: owner.ID},
		ZoneRef{Kind: ZoneExile},
		id, false,
	)
	if err != ErrCardNotFound {
		t.Fatalf("err = %v, want ErrCardNotFound", err)
	}
	assertOnlyIn(t, id, owner.Graveyard, g.Exile, owner.Command)
}

// TestUndoAcrossASandboxMovePauseReplaysTheSameWay — the resume frame
// is unserialisable continuation state, so the undo stack has to be
// able to rewind a game that is sitting on one. Rewinding to before
// the move drops the prompt with it; rewinding INTO the open prompt
// keeps it, and the second answer must land the card exactly where the
// first one did.
func TestUndoAcrossASandboxMovePauseReplaysTheSameWay(t *testing.T) {
	g := newActiveGame(t)
	// RestoreFrom swaps g.Seats wholesale, so the seat (and its
	// zones) is re-read after every rewind rather than captured once.
	owner := func() *Player { return g.Seats[0] }
	cmdID := seatCommander(t, owner().Graveyard, owner())
	move := func() {
		if err := g.MoveCardByIDAsCommander(
			ZoneRef{Kind: ZoneGraveyard, Owner: owner().ID},
			ZoneRef{Kind: ZoneExile},
			cmdID, true,
		); err != nil {
			t.Fatalf("MoveCardByIDAsCommander: %v", err)
		}
	}
	answer := func() {
		if len(g.PendingChoices) != 1 {
			t.Fatalf("pending choices = %d, want the CR 903.9 prompt", len(g.PendingChoices))
		}
		if err := g.ResolveOptionalReplacement(g.PendingChoices[0].ID, owner().ID, true); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
	}

	// --- rewind to before the move ---
	beforeMove := g.Clone()
	move()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the CR 903.9 prompt", len(g.PendingChoices))
	}
	g.WithWriteLock(func() { g.RestoreFrom(beforeMove) })
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the rewind to before the move", len(g.PendingChoices))
	}
	assertOnlyIn(t, cmdID, owner().Graveyard, g.Exile, owner().Command)

	// --- rewind into the open prompt, then answer again ---
	move()
	promptOpen := g.Clone()
	answer()
	assertOnlyIn(t, cmdID, owner().Command, owner().Graveyard, g.Exile)

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	assertOnlyIn(t, cmdID, owner().Graveyard, g.Exile, owner().Command)
	answer()
	assertOnlyIn(t, cmdID, owner().Command, owner().Graveyard, g.Exile)
}

// TestStaleSandboxMovePromptIsPruned — #701's prune, on the new path.
// Two prompts for one card, both paused, both about a move out of the
// graveyard; answering the first moves the card and the second becomes
// unanswerable at that instant, so it goes with it.
func TestStaleSandboxMovePromptIsPruned(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, owner.Graveyard, owner)

	for i := 0; i < 2; i++ {
		if err := g.MoveCardByIDAsCommander(
			ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID},
			ZoneRef{Kind: ZoneExile},
			cmdID, false,
		); err != nil {
			t.Fatalf("MoveCardByIDAsCommander #%d: %v", i, err)
		}
	}
	if len(g.PendingChoices) != 2 {
		t.Fatalf("expected the two duplicate prompts, got %d", len(g.PendingChoices))
	}

	first := g.PendingChoices[0]
	if err := g.ResolveOptionalReplacement(first.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the stale sibling survived its card leaving: %d prompts still queued",
			len(g.PendingChoices))
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Graveyard, g.Exile)
}

// TestStaleSandboxMoveAnswerIsDroppedNotRefused — the answer path's own
// tolerance. A card that left by a route NO prune sees leaves the
// prompt in flight; answering it must drop the move rather than rip the
// card back out of the battlefield.
//
// #1175 changed the vehicle, not the claim. The reanimation this test
// used to move the card with was the entry door's own gap: an ENTRY
// lands without going through the exit primitive, so nothing pruned the
// sibling prompt — and `pruneChoicesAfterArrivalLocked` now runs
// `pruneStaleZoneChangeChoicesLocked` there, which withdraws the prompt
// before anyone can answer it (that is
// TestAnArrivalPrunesAStaleMoveOutOfTheGraveyard's shape). The raw zone
// mutation below stands in for any door the two funnels do not cover:
// `dropStaleReplacementResumeLocked` is belt-and-braces and has to keep
// working when the belt is not there.
func TestStaleSandboxMoveAnswerIsDroppedNotRefused(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, owner.Graveyard, owner)

	if err := g.MoveCardByIDAsCommander(
		ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID},
		ZoneRef{Kind: ZoneExile},
		cmdID, false,
	); err != nil {
		t.Fatalf("MoveCardByIDAsCommander: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	g.WithWriteLock(func() {
		if _, err := MoveCard(owner.Graveyard, g.Battlefield, cmdID); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})

	base := len(g.Events)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("answering a stale prompt was refused: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts still queued after the answer", len(g.PendingChoices))
	}
	assertOnlyIn(t, cmdID, g.Battlefield, g.Exile, owner.Graveyard, owner.Command)
	var sawBreadcrumb bool
	for _, ev := range g.Events[base:] {
		if ev.Kind == EventEffectError {
			sawBreadcrumb = true
			break
		}
	}
	if !sawBreadcrumb {
		t.Error("no EventEffectError breadcrumb for the dropped prompt")
	}
}
