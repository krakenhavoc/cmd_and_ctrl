package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// resolution_pause_test.go — #1289: state-based actions wait for a
// resolution that is paused on one of its own prompts (CR 704.3).
//
// Every test resolves a real stack item through PassPriority, because
// the bug lives in the boundary passPriorityLocked runs after a
// resolution. Calling the verb directly under the lock would skip that
// boundary and test nothing.

// resolveUntilPrompt passes priority until the top item has resolved
// into a prompt, and fails if it never does.
func resolveUntilPrompt(t *testing.T, g *Game) {
	t.Helper()
	passUntilPrompt(t, g)
	if len(g.PendingChoices) == 0 {
		t.Fatal("the resolution finished without asking anything")
	}
}

// bearForPause parks a 2/2 whose toughness the engine knows, so CR
// 704.5f and CR 704.5g apply to it.
func bearForPause(g *Game, owner uuid.UUID) uuid.UUID {
	id := pushBattlefieldForTest(g, owner, "Bear", "Creature — Bear", "")
	c := findCard(g, id)
	c.Power, c.Toughness = 2, 2
	c.PrintedPTKnown = true
	c.EnteredBattlefieldAt = timeNowUnixNano()
	return id
}

// markLethal puts two damage on a 2/2, inside a resolution.
func markLethal(g *Game, id uuid.UUID) {
	if c, ok := g.battlefieldCardLocked(id); ok {
		c.DamageMarked = 2
	}
}

// tokenDoublerForPause is a sourceless Parallel Lives for one player.
func tokenDoublerForPause(controller uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventTokenCreated},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCreateTokens && ev.TokenController == controller
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.MultiplyTokens(2)
			return nil
		},
		Label: "twice as many tokens",
	}
}

// armiesOf lists the Army creatures a player controls.
func armiesOf(g *Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtype(ArmySubtype) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// TestAnEarthbentBareLandSurvivesAPausedCounterPlacement is #1289's
// first example. Earthbend 3 onto a land with no counters, on a board
// where two counter replacements apply: the land becomes a 0/0 and its
// counters wait on the CR 616 prompt. The land must still be the same
// object when the counters land.
func TestAnEarthbentBareLandSurvivesAPausedCounterPlacement(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")
	registerPausingCounterBoard(g, CounterPlusOne)

	pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "earthbend 3", func(g *Game, it *StackItem) error {
		return g.EarthbendForEffect(it.Controller, it.SourceCardID, land, 3)
	})
	resolveUntilPrompt(t, g)

	prompt := g.PendingChoices[0]
	if prompt.Kind != PendingChoiceReplacementOrder || !prompt.midResolution {
		t.Fatalf("prompt kind=%q midResolution=%v, want the CR 616 prompt, stamped", prompt.Kind, prompt.midResolution)
	}
	if !g.ResolutionPaused() {
		t.Error("ResolutionPaused() = false with the resolution's own prompt open")
	}
	if findCard(g, land) == nil {
		t.Fatal("the 0/0 land died to CR 704.5f while its counters were still owed")
	}
	if n := countEventsOfKind(g, EventLTB); n != 0 {
		t.Fatalf("%d LTB event(s) during the pause, want none", n)
	}

	answerCounterOrder(t, g)

	// ×2 then +1: 3 → 6 → 7 counters, onto the same object.
	if got := counterOn(g, land, CounterPlusOne); got != 7 {
		t.Fatalf("counters on the land = %d, want 7", got)
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still paused after its last prompt was answered")
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("the table cannot move after the answer: %v", err)
	}
}

// TestAFreshArmySurvivesAPausedCounterPlacement is #1289's second
// example. Amass with no Army: the 0/0 token is created, then its
// counter placement pauses.
func TestAFreshArmySurvivesAPausedCounterPlacement(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	registerPausingCounterBoard(g, CounterPlusOne)

	var amassed uuid.UUID
	pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "amass Orcs 2", func(g *Game, it *StackItem) error {
		return g.AmassForEffect(it.Controller, it.SourceCardID, amassTestToken("Orc"), "Orc", 2,
			func(_ *Game, army uuid.UUID) error {
				amassed = army
				return nil
			})
	})
	resolveUntilPrompt(t, g)

	armies := armiesOf(g, me.ID)
	if len(armies) != 1 {
		t.Fatalf("%d Armies on the battlefield during the pause, want the fresh 0/0", len(armies))
	}
	if !g.PendingChoices[0].midResolution {
		t.Fatal("the counter prompt was not stamped as part of the resolution")
	}

	answerCounterOrder(t, g)

	if amassed != armies[0] {
		t.Fatalf("\"the Army you amassed\" = %v, want the token (%v)", amassed, armies[0])
	}
	// ×2 then +1: 2 → 4 → 5.
	if got := counterOn(g, armies[0], CounterPlusOne); got != 5 {
		t.Errorf("Army counters = %d, want 5", got)
	}
}

// TestADoubledArmyCreationHoldsBothArmiesUntilTheCountersLand is the
// issue's worst case: a token doubler makes two 0/0 Armies, the
// choose-an-Army prompt pauses first, then the counter placement
// pauses again. Both Armies live until the resolution is over, and
// then CR 704.5f takes the one that got no counters.
func TestADoubledArmyCreationHoldsBothArmiesUntilTheCountersLand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	registerPausingCounterBoard(g, CounterPlusOne)
	g.WithWriteLock(func() { g.RegisterReplacementForTest(tokenDoublerForPause(me.ID)) })

	pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "amass Orcs 2", func(g *Game, it *StackItem) error {
		return g.AmassForEffect(it.Controller, it.SourceCardID, amassTestToken("Orc"), "Orc", 2, nil)
	})
	resolveUntilPrompt(t, g)

	armies := armiesOf(g, me.ID)
	if len(armies) != 2 {
		t.Fatalf("%d Armies during the choose prompt, want both 0/0s", len(armies))
	}
	pick := g.PendingChoices[0]
	if pick.Kind != PendingChoiceChooseCards || !pick.midResolution {
		t.Fatalf("prompt kind=%q midResolution=%v, want the stamped choose-an-Army prompt", pick.Kind, pick.midResolution)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{armies[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if n := len(armiesOf(g, me.ID)); n != 2 {
		t.Fatalf("%d Armies during the counter prompt, want both still there", n)
	}
	if !g.PendingChoices[0].midResolution {
		t.Fatal("the counter prompt queued by the answer was not stamped")
	}

	answerCounterOrder(t, g)

	left := armiesOf(g, me.ID)
	if len(left) != 1 || left[0] != armies[1] {
		t.Fatalf("Armies after the resolution = %v, want only the chosen one (%v)", left, armies[1])
	}
	if got := counterOn(g, armies[1], CounterPlusOne); got != 5 {
		t.Errorf("chosen Army counters = %d, want 5", got)
	}
}

// TestAResolutionWaitingOnAYesNoHoldsTheSweep is the pause that has
// nothing to do with counters. The ability marks lethal damage and
// then asks "you may remove it". The answer is part of the same
// resolution, so the creature is still there to be saved.
func TestAResolutionWaitingOnAYesNoHoldsTheSweep(t *testing.T) {
	for _, accept := range []bool{true, false} {
		name := "decline"
		if accept {
			name = "accept"
		}
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			bear := bearForPause(g, me.ID)
			pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "hurt, then maybe heal", func(g *Game, it *StackItem) error {
				markLethal(g, bear)
				g.QueueConfirmForEffect(ConfirmPrompt{
					Chooser:  it.Controller,
					Question: "Remove the damage?",
					OnAccept: func(g *Game) error {
						if c, ok := g.battlefieldCardLocked(bear); ok {
							c.DamageMarked = 0
						}
						return nil
					},
				})
				return nil
			})
			resolveUntilPrompt(t, g)

			if findCard(g, bear) == nil {
				t.Fatal("CR 704.5g ran while the resolution was waiting on its yes/no")
			}
			var perr *ChoicePendingError
			if err := g.PassPriority(); !errors.As(err, &perr) {
				t.Fatalf("PassPriority during the pause = %v, want the choice gate", err)
			}

			c := g.PendingChoices[0]
			if err := g.ResolveConfirm(c.ID, me.ID, accept); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			alive := findCard(g, bear) != nil
			if alive != accept {
				t.Errorf("bear alive = %v after answering %v; the sweep must run after the answer, not before", alive, accept)
			}
		})
	}
}

// TestStateChecksInsideAResolutionWait: a state check asked for while
// the resolution function is still running does nothing (CR 704.3).
// The sweep runs once, after the item has finished.
func TestStateChecksInsideAResolutionWait(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := bearForPause(g, me.ID)
	aliveInside := false
	pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "hurt, then check", func(g *Game, it *StackItem) error {
		markLethal(g, bear)
		g.runStateChecksLocked()
		_, aliveInside = g.battlefieldCardLocked(bear)
		return nil
	})
	passUntilStackEmpty(t, g)

	if !aliveInside {
		t.Error("the sweep ran in the middle of the resolution")
	}
	if findCard(g, bear) != nil {
		t.Error("the sweep never ran after the resolution finished")
	}
}

// TestANonBlockingPromptDoesNotHoldTheSweep:a pay_unless asked of
// another player after the ability has left the stack does not block
// the table (ADR 0018 §6), so it must not hold the boundary either. If
// it did, play would go on with state-based actions switched off.
func TestANonBlockingPromptDoesNotHoldTheSweep(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := bearForPause(g, me.ID)
	pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "hurt, then tax", func(g *Game, it *StackItem) error {
		markLethal(g, bear)
		return g.QueuePayUnlessForEffect(them.ID, uuid.Nil, "{1}", "Pay {1}?", nil)
	})
	resolveUntilPrompt(t, g)

	if findCard(g, bear) != nil {
		t.Error("the sweep was held behind a prompt that does not block the table")
	}
	if g.ResolutionPaused() {
		t.Error("ResolutionPaused() = true behind a non-blocking prompt")
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("PassPriority behind a non-blocking prompt: %v", err)
	}
}

// TestTriggerPlacementPromptsAreNotPartOfTheResolution pins
// choiceBelongsToResolution: a prompt about putting a triggered
// ability on the stack is answered at the boundary this change holds,
// so it must not hold that boundary itself (#809's order).
func TestTriggerPlacementPromptsAreNotPartOfTheResolution(t *testing.T) {
	cases := []struct {
		name string
		c    PendingChoice
		want bool
	}{
		{"trigger order", PendingChoice{Kind: PendingChoiceTriggerOrder}, false},
		{"loop shortcut", PendingChoice{Kind: PendingChoiceLoopShortcut}, false},
		{"optional trigger", PendingChoice{Kind: PendingChoiceTriggerPrompt, triggerResume: &triggerResumeFrame{}}, false},
		{"trigger target", PendingChoice{Kind: PendingChoicePickTarget, pickTargetResume: &pickTargetFrame{}}, false},
		{"trigger mode", PendingChoice{Kind: PendingChoiceModePick, modePickResume: &modePickFrame{}}, false},
		{"copy target", PendingChoice{Kind: PendingChoicePickTarget, copyResume: &copyFrame{}}, true},
		{"replacement order", PendingChoice{Kind: PendingChoiceReplacementOrder}, true},
		{"confirm", PendingChoice{Kind: PendingChoiceConfirm}, true},
		{"choose cards", PendingChoice{Kind: PendingChoiceChooseCards}, true},
		{"scry", PendingChoice{Kind: PendingChoiceScry}, true},
	}
	for _, tc := range cases {
		if got := choiceBelongsToResolution(&tc.c); got != tc.want {
			t.Errorf("%s: choiceBelongsToResolution = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestAPromptQueuedOutsideAResolutionIsNotStamped: the stamp is about
// resolutions. A prompt queued at any other time holds nothing.
func TestAPromptQueuedOutsideAResolutionIsNotStamped(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueConfirmForEffect(ConfirmPrompt{Chooser: me.ID, Question: "?"})
	})
	if g.PendingChoices[0].midResolution {
		t.Error("a prompt queued outside a resolution was stamped")
	}
	if g.ResolutionPaused() {
		t.Error("ResolutionPaused() = true with no resolution open")
	}
}

// TestAPausedResolutionReleasesWhenItsChooserLeaves is the departure
// half of ADR 0018 §6, for every prompt kind the gate classifies. A
// resolution paused on a prompt addressed to a player who then
// concedes must never be left holding for a seat that has gone: the
// departure table drops the prompt (and the sweep runs) or reassigns it
// (and a live seat owes the answer).
//
// The prompt is about an enchantment the resolving player controls, so
// the kinds CR 800.4g reassigns have an object and an inheritor, and
// both columns of the table are exercised.
func TestAPausedResolutionReleasesWhenItsChooserLeaves(t *testing.T) {
	released, reassigned := 0, 0
	for _, kind := range ClassifiedChoiceKinds() {
		t.Run(string(kind), func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			me, leaver := g.Seats[0], g.Seats[2]
			bear := bearForPause(g, me.ID)
			source := pushBattlefieldForTest(g, me.ID, "Asking Enchantment", "Enchantment", "")
			pushAbilityItem(g, StackItemTriggered, me.ID, source, "hurt, then ask", func(g *Game, it *StackItem) error {
				markLethal(g, bear)
				g.QueueChoiceForEffect(PendingChoice{
					Kind:       kind,
					Chooser:    leaver.ID,
					FromPlayer: me.ID,
					Source:     source,
					Count:      1,
					Reason:     "asked of the seat that leaves",
				})
				return nil
			})
			resolveUntilPrompt(t, g)

			if err := g.Concede(leaver.ID); err != nil {
				t.Fatalf("Concede: %v", err)
			}
			if c := g.pausedResolutionChoiceLocked(); c != nil && g.resolutionOpen {
				if p := g.playerByIDLocked(c.Chooser); p == nil || p.Eliminated {
					t.Fatalf("the resolution is held for %s, whose chooser has left", c.Kind)
				}
				reassigned++
				return
			}
			if findCard(g, bear) != nil {
				t.Error("the prompt was dropped with its chooser, but the sweep it was holding never ran")
			}
			released++
		})
	}
	if released == 0 || reassigned == 0 {
		t.Errorf("released=%d reassigned=%d; both departure columns should be exercised", released, reassigned)
	}
}

// TestUndoAcrossAPausedResolutionRestoresThePause: Clone carries the
// open resolution and the stamp, so an undone answer is paused again.
func TestUndoAcrossAPausedResolutionRestoresThePause(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := earthbendTestLand(t, g, me, "Forest")
	registerPausingCounterBoard(g, CounterPlusOne)
	pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "earthbend 3", func(g *Game, it *StackItem) error {
		return g.EarthbendForEffect(it.Controller, it.SourceCardID, land, 3)
	})
	resolveUntilPrompt(t, g)

	snap := g.Clone()
	answerCounterOrder(t, g)
	g.RestoreFrom(snap)

	if !g.ResolutionPaused() {
		t.Fatal("the restored game is not paused on the resolution's prompt")
	}
	if findCard(g, land) == nil {
		t.Fatal("the land is gone in the restored game")
	}
	answerCounterOrder(t, g)
	if got := counterOn(g, land, CounterPlusOne); got != 7 {
		t.Errorf("counters after the replayed answer = %d, want 7", got)
	}
}
