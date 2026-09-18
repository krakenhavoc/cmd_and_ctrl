package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// leave_game_choices_test.go — #902, CR 800.4g/h/i/m: what happens to a
// choice a player still owed when they left the game, what an effect
// reads about them afterwards, and when a duration keyed to their next
// turn ends.
//
// EVERY TEST HERE NEEDS FOUR SEATS. ADR 0060 Decision 5: in a
// two-player game the departure ends the game, so none of CR 800.4a's
// consequences — the object sweep, and therefore the reassignment that
// rides on it — is ever observable. A two-seat version of any of these
// would pass for the wrong reason.

// --- helpers ---------------------------------------------------------

// departureTestSource puts a plain permanent on the battlefield under
// `owner`'s control and returns its instance ID. It is the OBJECT
// CR 800.4g asks about: the reassignment policy finds the prompt's
// controller by looking Source up in the zones, so a prompt built for
// these tests needs a real card behind it.
func departureTestSource(g *Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Enchantment",
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// handCardOf returns one card ID out of a seat's opening hand, which is
// a card that seat OWNS — so CR 800.4a takes it out of the game when
// they leave.
func handCardOf(t *testing.T, p *Player, n int) []uuid.UUID {
	t.Helper()
	if p.Hand == nil || len(p.Hand.Cards) < n {
		t.Fatalf("seat %s holds %d cards, need %d", p.Name, len(p.Hand.Cards), n)
	}
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.Hand.Cards[i].InstanceID)
	}
	return out
}

func findChoice(g *Game, id uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.ID == id {
			return c
		}
	}
	return nil
}

// lastChoiceEvent returns the most recent reassign / drop event for the
// departed chooser, or nil.
func lastChoiceEvent(g *Game, kind EventKind, actor uuid.UUID) *Event {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == kind && g.Events[i].Actor == actor {
			return &g.Events[i]
		}
	}
	return nil
}

// --- CR 800.4g: the reassignment ------------------------------------

// TestTriggerPromptOnAnotherPlayersPermanentIsReassigned is the
// headline, on the one path the trigger dispatcher really produces it:
// a CR 603.5 "you may" whose OptionalPrompt.Chooser points at somebody
// other than the source's controller (Edric, Spymaster of Trest is the
// catalog card of that shape).
//
// Seat 0 controls the permanent. Seat 1 owes the choice and concedes.
// CR 800.4g's second sentence — "if the original choice was to be made
// by an OPPONENT of the controller of the object, that player chooses
// another opponent if possible" — makes it seat 2's, not seat 0's.
func TestTriggerPromptOnAnotherPlayersPermanentIsReassigned(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	controller, leaver, next := g.Seats[0], g.Seats[1], g.Seats[2]
	sourceID := departureTestSource(g, controller.ID, "Edric, Spymaster of Trest")

	built := 0
	g.WithWriteLock(func() {
		source := Card{InstanceID: sourceID, Name: "Edric, Spymaster of Trest",
			Controller: controller.ID, Owner: controller.ID}
		g.dispatchTriggerLocked(Event{Kind: EventDealDamage, CardID: sourceID}, source, source.Effective(),
			TriggeredAbility{
				OptionalPrompt: &TriggerOptionalPrompt{
					Question: "Draw a card?",
					Chooser: func(Event, *Card, *Game) uuid.UUID {
						return leaver.ID
					},
				},
				Build: func(Event, *Card, Characteristic, *Game) *StackItem {
					built++
					return nil
				},
			})
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != leaver.ID {
		t.Fatalf("setup: want one prompt owed by the leaver, got %+v", g.PendingChoices)
	}
	choiceID := g.PendingChoices[0].ID

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	c := findChoice(g, choiceID)
	if c == nil {
		t.Fatalf("the prompt was dropped; CR 800.4g reassigns it")
	}
	if c.Chooser != next.ID {
		t.Fatalf("prompt went to %s, want seat 2 (%s) — the next OPPONENT of the object's controller in turn order",
			c.Chooser, next.ID)
	}
	ev := lastChoiceEvent(g, EventPendingChoiceReassigned, leaver.ID)
	if ev == nil {
		t.Fatalf("no EventPendingChoiceReassigned emitted")
	}
	if ev.Target != next.ID {
		t.Errorf("event Target = %s, want %s", ev.Target, next.ID)
	}
	if ev.Source != sourceID {
		t.Errorf("event Source = %s, want the object %s", ev.Source, sourceID)
	}
	if ev.Label != string(PendingChoiceTriggerPrompt) {
		t.Errorf("event Label = %q, want %q", ev.Label, PendingChoiceTriggerPrompt)
	}

	// The table is still held by the prompt — it has an answerer now,
	// not no answerer.
	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Errorf("PassPriority after the reassignment = %v, want ErrChoicePending", err)
	}
	// And the inheritor can answer it, which runs the SAME
	// continuation: the card did not change because the player
	// answering it did.
	if err := g.ResolveTriggerPrompt(choiceID, next.ID, true); err != nil {
		t.Fatalf("the inheritor could not answer the prompt: %v", err)
	}
	if built != 1 {
		t.Errorf("the trigger's Build ran %d times, want 1", built)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the prompt is still open after the inheritor answered: %+v", g.PendingChoices)
	}
	if err := g.PassPriority(); err != nil && errors.Is(err, ErrChoicePending) {
		t.Errorf("the table is still blocked after the prompt was answered: %v", err)
	}
}

// TestOnlyTheInheritorMayAnswerAReassignedPrompt — the reassignment is
// a real change of address, not a widening. The departed seat's ID is
// refused, and so is every seat that did not inherit it.
func TestOnlyTheInheritorMayAnswerAReassignedPrompt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	controller, leaver, next, last := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	sourceID := departureTestSource(g, controller.ID, "Edric, Spymaster of Trest")

	g.WithWriteLock(func() {
		source := Card{InstanceID: sourceID, Controller: controller.ID, Owner: controller.ID}
		g.dispatchTriggerLocked(Event{Kind: EventDealDamage, CardID: sourceID}, source, source.Effective(),
			TriggeredAbility{
				OptionalPrompt: &TriggerOptionalPrompt{
					Question: "Draw a card?",
					Chooser:  func(Event, *Card, *Game) uuid.UUID { return leaver.ID },
				},
				Build: func(Event, *Card, Characteristic, *Game) *StackItem { return nil },
			})
	})
	choiceID := g.PendingChoices[0].ID
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if c := findChoice(g, choiceID); c == nil || c.Chooser != next.ID {
		t.Fatalf("setup: the prompt did not reach seat 2")
	}
	for _, who := range []*Player{leaver, controller, last} {
		if err := g.ResolveTriggerPrompt(choiceID, who.ID, true); !errors.Is(err, ErrNotTheChooser) {
			t.Errorf("%s answered a prompt addressed to seat 2: %v", who.Name, err)
		}
	}
}

// TestPileSplitIsReassignedWhenTheSplitterLeaves is the other shape the
// engine really produces: Fact or Fiction's opponent-separates-the-piles
// prompt (a choose_cards over cards the SPLITTER does not own). The
// splitter leaves mid-split and the next opponent separates instead.
func TestPileSplitIsReassignedWhenTheSplitterLeaves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, splitter, next := g.Seats[0], g.Seats[1], g.Seats[2]
	sourceID := departureTestSource(g, caster.ID, "Fact or Fiction")
	revealed := handCardOf(t, caster, 3)

	var taken, left []uuid.UUID
	g.WithWriteLock(func() {
		g.QueuePileSplitForEffect(PileSplitPrompt{
			Splitter:      splitter.ID,
			Chooser:       caster.ID,
			Owner:         caster.ID,
			Source:        sourceID,
			SplitQuestion: "Separate those cards into two piles",
			PickQuestion:  "Take a pile",
			Cards:         revealed,
			Then: func(_ *Game, tk, lf []uuid.UUID) error {
				taken, left = tk, lf
				return nil
			},
		})
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceChooseCards {
		t.Fatalf("setup: want one choose_cards prompt, got %+v", g.PendingChoices)
	}
	choiceID := g.PendingChoices[0].ID

	if err := g.Concede(splitter.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := findChoice(g, choiceID)
	if c == nil {
		t.Fatalf("the split prompt was dropped; CR 800.4g reassigns it")
	}
	if c.Chooser != next.ID {
		t.Fatalf("the split went to %s, want seat 2 (%s)", c.Chooser, next.ID)
	}
	if len(c.ChooseCards) != len(revealed) {
		t.Errorf("the candidate list changed on reassignment: %d cards, want %d",
			len(c.ChooseCards), len(revealed))
	}

	// The inheritor splits, the caster picks, and the continuation the
	// card supplied runs untouched.
	if err := g.ResolveChooseCards(choiceID, next.ID, revealed[:1]); err != nil {
		t.Fatalf("the inheritor could not split: %v", err)
	}
	var pickID uuid.UUID
	for _, p := range g.PendingChoices {
		if p != nil && p.Kind == PendingChoiceOptionPick {
			pickID = p.ID
		}
	}
	if pickID == uuid.Nil {
		t.Fatalf("the chained pile-pick prompt was never queued: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionPick(pickID, caster.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if len(taken) != 1 || len(left) != len(revealed)-1 {
		t.Errorf("the card's continuation saw piles %v / %v", taken, left)
	}
}

// TestPickTargetIsReassignedAndItsTargetSetPruned covers the third
// reassignable kind at the seam rather than through the dispatcher.
// queuePickTargetLocked addresses a CR 603.3d target pick to the
// SOURCE'S CONTROLLER, and a controller who leaves takes the source
// with them (CR 800.4a), so the trigger harvester cannot produce an
// opponent-addressed pick_target today. The row in
// choiceReassignDecisions is a decision about the RULE — an object's
// choice that is not a cost is reassigned — and #918 already lets a
// card address any seat, so the prompt is built here the way such a
// card would build it.
//
// It also pins the prune: the departed player is not offered as a
// target to their inheritor, and neither are the permanents leaving
// the game with them.
func TestPickTargetIsReassignedAndItsTargetSetPruned(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	controller, leaver, next, last := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	sourceID := departureTestSource(g, controller.ID, "Opponent-directed trigger")
	theirs := departureTestSource(g, leaver.ID, "Grizzly Bears")
	survivor := departureTestSource(g, last.ID, "Grizzly Bears")

	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChoiceForEffect(PendingChoice{
			Kind:              PendingChoicePickTarget,
			Chooser:           leaver.ID,
			Count:             1,
			Source:            sourceID,
			Reason:            "choose a target",
			PickTargetPlayers: []uuid.UUID{leaver.ID, last.ID},
			PickTargetCards:   []uuid.UUID{theirs, survivor},
			PickTargetMin:     1,
			PickTargetMax:     1,
		})
	})
	if id == uuid.Nil {
		t.Fatalf("setup: the prompt was not queued")
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := findChoice(g, id)
	if c == nil {
		t.Fatalf("the prompt was dropped; a legal target survives the departure")
	}
	if c.Chooser != next.ID {
		t.Fatalf("prompt went to %s, want seat 2 (%s)", c.Chooser, next.ID)
	}
	for _, p := range c.PickTargetPlayers {
		if p == leaver.ID {
			t.Errorf("the departed player is still offered as a target")
		}
	}
	for _, card := range c.PickTargetCards {
		if card == theirs {
			t.Errorf("a permanent that left the game with the chooser is still offered as a target")
		}
	}
	if len(c.PickTargetCards) != 1 || c.PickTargetCards[0] != survivor {
		t.Errorf("PickTargetCards = %v, want just the surviving permanent", c.PickTargetCards)
	}
}

// --- CR 800.4g: what is NOT reassigned -------------------------------

// TestOwnMaterialOptionPickIsDroppedNotReassigned — Torment of
// Hailfire's "each opponent loses 3 life unless THAT PLAYER sacrifices
// a nonland permanent or discards a card". Every branch acts on the
// player being asked, and CR 800.4a took their permanents, their hand
// and their life out of the game a moment ago, so there is no answer
// that does anything. It is dropped, with #868's event, rather than
// handed to a living seat as a question about a player who is not
// there.
//
// The discriminator is structural and not per-card: FromPlayer is the
// engine's name for whose material a prompt is about, and here it is
// the chooser's own.
func TestOwnMaterialOptionPickIsDroppedNotReassigned(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver := g.Seats[0], g.Seats[1]
	sourceID := departureTestSource(g, caster.ID, "Torment of Hailfire")

	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser:  leaver.ID,
			Source:   sourceID,
			Question: "Torment of Hailfire",
			Options: []ChoiceOption{
				{Label: "Lose 3 life", LifeCost: 3},
				{Label: "Discard a card"},
			},
			Then: func(*Game, int) error { return nil },
		})
	})
	if id == uuid.Nil {
		t.Fatalf("setup: the option pick was not queued")
	}
	if got := findChoice(g, id).FromPlayer; got != leaver.ID {
		t.Fatalf("setup: FromPlayer = %s, want the chooser %s", got, leaver.ID)
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if c := findChoice(g, id); c != nil {
		t.Fatalf("an own-material option pick was reassigned to %s", c.Chooser)
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, leaver.ID) == nil {
		t.Errorf("no EventPendingChoiceDropped for the dropped own-material prompt")
	}
	if lastChoiceEvent(g, EventPendingChoiceReassigned, leaver.ID) != nil {
		t.Errorf("an own-material prompt emitted a reassignment event")
	}
}

// TestPayUnlessIsNeverReassigned is CR 800.4f, which is the rule
// immediately before the one this issue is about and says the opposite:
// "if an object requires a player who has left the game to pay a cost
// or choose whether to pay a cost, that cost is not paid". Nobody else
// is asked to pay a departed player's Rhystic tax.
func TestPayUnlessIsNeverReassigned(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	taxer, leaver := g.Seats[0], g.Seats[1]
	sourceID := departureTestSource(g, taxer.ID, "Rhystic Study")

	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(leaver.ID, sourceID, "{1}",
			"Rhystic Study — pay {1}?", nil); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePayUnless {
			t.Fatalf("a departed player's tax was handed to %s", c.Chooser)
		}
	}
}

// TestChoiceIsDroppedWhenItsObjectLeftFirst — every candidate gone.
// CR 800.4g reassigns a choice an OBJECT requires; when the object has
// itself left the game (its controller conceded earlier, so CR 800.4a
// took it), there is nothing left to require the choice and the prompt
// is dropped with the event #868 added.
func TestChoiceIsDroppedWhenItsObjectLeftFirst(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	controller, leaver := g.Seats[0], g.Seats[1]
	sourceID := departureTestSource(g, controller.ID, "Edric, Spymaster of Trest")

	g.WithWriteLock(func() {
		source := Card{InstanceID: sourceID, Controller: controller.ID, Owner: controller.ID}
		g.dispatchTriggerLocked(Event{Kind: EventDealDamage, CardID: sourceID}, source, source.Effective(),
			TriggeredAbility{
				OptionalPrompt: &TriggerOptionalPrompt{
					Question: "Draw a card?",
					Chooser:  func(Event, *Card, *Game) uuid.UUID { return leaver.ID },
				},
				Build: func(Event, *Card, Characteristic, *Game) *StackItem { return nil },
			})
	})
	choiceID := g.PendingChoices[0].ID

	// The object's controller leaves first: CR 800.4a takes the
	// permanent out of the game with them. The prompt is untouched —
	// its chooser is still at the table.
	if err := g.Concede(controller.ID); err != nil {
		t.Fatalf("Concede (controller): %v", err)
	}
	if c := findChoice(g, choiceID); c == nil || c.Chooser != leaver.ID {
		t.Fatalf("the prompt moved when somebody else left: %+v", g.PendingChoices)
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede (chooser): %v", err)
	}
	if c := findChoice(g, choiceID); c != nil {
		t.Fatalf("a prompt whose object had already left was reassigned to %s", c.Chooser)
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, leaver.ID) == nil {
		t.Errorf("no EventPendingChoiceDropped when there was nobody to inherit")
	}
	// And the table is free — the #864 symptom must not come back by
	// another door.
	if err := g.PassPriority(); errors.Is(err, ErrChoicePending) {
		t.Errorf("the table is blocked behind the dropped prompt: %v", err)
	}
}

// TestReassignmentPrunesCandidatesLeavingWithTheChooser — CR 800.4a
// runs in the same breath as the reassignment, so a candidate the
// departing player OWNS is a candidate that will not be there when the
// answer arrives (the staleness #701 taught the zone-change resume to
// recognise, one step earlier). It is pruned as the prompt changes
// hands.
func TestReassignmentPrunesCandidatesLeavingWithTheChooser(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver, next := g.Seats[0], g.Seats[1], g.Seats[2]
	sourceID := departureTestSource(g, caster.ID, "Fact or Fiction")
	mine := handCardOf(t, caster, 2)
	theirs := handCardOf(t, leaver, 2)

	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:    leaver.ID,
			FromPlayer: caster.ID,
			Source:     sourceID,
			Question:   "Separate those cards into two piles",
			Cards:      append(append([]uuid.UUID{}, mine...), theirs...),
			Then:       func(*Game, []uuid.UUID) error { return nil },
		})
	})
	if id == uuid.Nil {
		t.Fatalf("setup: the prompt was not queued")
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := findChoice(g, id)
	if c == nil {
		t.Fatalf("the prompt was dropped; two candidates survive the departure")
	}
	if c.Chooser != next.ID {
		t.Fatalf("prompt went to %s, want seat 2", c.Chooser)
	}
	if len(c.ChooseCards) != len(mine) {
		t.Fatalf("candidates after the prune = %d, want %d", len(c.ChooseCards), len(mine))
	}
	for _, id := range c.ChooseCards {
		for _, goneID := range theirs {
			if id == goneID {
				t.Errorf("a card that left the game with the chooser is still on the prompt")
			}
		}
	}
	if c.ChooseMax != len(mine) {
		t.Errorf("ChooseMax = %d after the prune, want %d — an enumerator that offers a bigger set than the resolver accepts is the #544 wedge",
			c.ChooseMax, len(mine))
	}
	// The bound is the whole point: the answer the inheritor gives has
	// to be one the resolver takes.
	if err := g.ResolveChooseCards(id, next.ID, c.ChooseCards); err != nil {
		t.Errorf("the inheritor's maximal answer was refused: %v", err)
	}
}

// TestPromptIsDroppedWhenEveryCandidateLeavesWithTheChooser — the same
// prune, taken to its end. A prompt with nothing left to ask is not
// moved: handing a seat a question with no answers is the #544 wedge.
func TestPromptIsDroppedWhenEveryCandidateLeavesWithTheChooser(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver := g.Seats[0], g.Seats[1]
	sourceID := departureTestSource(g, caster.ID, "Fact or Fiction")
	theirs := handCardOf(t, leaver, 3)

	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:    leaver.ID,
			FromPlayer: caster.ID,
			Source:     sourceID,
			Question:   "Separate those cards into two piles",
			Cards:      theirs,
			Min:        1,
			Then:       func(*Game, []uuid.UUID) error { return nil },
		})
	})
	if id == uuid.Nil {
		t.Fatalf("setup: the prompt was not queued")
	}
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if c := findChoice(g, id); c != nil {
		t.Fatalf("a prompt with no surviving candidates was handed to %s", c.Chooser)
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, leaver.ID) == nil {
		t.Errorf("no EventPendingChoiceDropped for the emptied prompt")
	}
}

// --- the classification table ---------------------------------------

// TestEveryChoiceKindHasAReassignmentDecision is the mechanism that
// keeps a NEW prompt kind from silently inheriting the wrong answer. It
// leans on choiceGateDecisions being complete, which
// TestEveryChoiceKindIsClassifiedAndEnumerated (internal/legal) already
// enforces against the declared constants.
func TestEveryChoiceKindHasAReassignmentDecision(t *testing.T) {
	classified := ClassifiedChoiceKinds()
	if len(classified) < 20 {
		t.Fatalf("the gate classifies only %d kinds — this test's premise has broken", len(classified))
	}
	for _, kind := range classified {
		if _, ok := choiceReassignDecisions[kind]; !ok {
			t.Errorf("%q has no row in choiceReassignDecisions — decide whether CR 800.4g reassigns it and say so in server/internal/game/leave_game.go (ADR 0060, 2026-09-18 amendment)", kind)
		}
	}
	gate := make(map[PendingChoiceKind]bool, len(classified))
	for _, kind := range classified {
		gate[kind] = true
	}
	for _, kind := range ReassignableChoiceKinds() {
		if !gate[kind] {
			t.Errorf("choiceReassignDecisions classifies %q, which no longer exists", kind)
		}
	}
}

// --- undo and the snapshot ------------------------------------------

// TestUndoAcrossAReassignmentReplaysIt — Chooser is plain data on a
// snapshotted choice, so an undo puts the prompt back in front of the
// player who left and the replayed departure moves it again. Nothing
// new is owed to clone.go or to the census: the reassignment writes
// only fields that were already carried.
func TestUndoAcrossAReassignmentReplaysIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	controller, leaver, next := g.Seats[0], g.Seats[1], g.Seats[2]
	sourceID := departureTestSource(g, controller.ID, "Edric, Spymaster of Trest")

	g.WithWriteLock(func() {
		source := Card{InstanceID: sourceID, Controller: controller.ID, Owner: controller.ID}
		g.dispatchTriggerLocked(Event{Kind: EventDealDamage, CardID: sourceID}, source, source.Effective(),
			TriggeredAbility{
				OptionalPrompt: &TriggerOptionalPrompt{
					Question: "Draw a card?",
					Chooser:  func(Event, *Card, *Game) uuid.UUID { return leaver.ID },
				},
				Build: func(Event, *Card, Characteristic, *Game) *StackItem { return nil },
			})
	})
	choiceID := g.PendingChoices[0].ID
	before := g.Clone()

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if c := findChoice(g, choiceID); c == nil || c.Chooser != next.ID {
		t.Fatalf("setup: the prompt did not reach seat 2")
	}

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if g.Seats[1].Eliminated {
		t.Fatal("the rewind puts the conceded player back at the table")
	}
	c := findChoice(g, choiceID)
	if c == nil || c.Chooser != g.Seats[1].ID {
		t.Fatalf("the rewind did not put the prompt back in front of the player who owed it: %+v", g.PendingChoices)
	}

	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede (replay): %v", err)
	}
	if c := findChoice(g, choiceID); c == nil || c.Chooser != g.Seats[2].ID {
		t.Fatalf("the replayed departure did not reassign the prompt again: %+v", g.PendingChoices)
	}
}

// TestReassignedChoiceSurvivesASnapshotRoundTrip — the restored game
// owes the prompt to the same seat, with the same pruned candidate set
// and the same bounds.
func TestReassignedChoiceSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver, next := g.Seats[0], g.Seats[1], g.Seats[2]
	sourceID := departureTestSource(g, caster.ID, "Fact or Fiction")
	mine := handCardOf(t, caster, 2)
	theirs := handCardOf(t, leaver, 1)

	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:    leaver.ID,
			FromPlayer: caster.ID,
			Source:     sourceID,
			Question:   "Separate those cards into two piles",
			Cards:      append(append([]uuid.UUID{}, mine...), theirs...),
			Then:       func(*Game, []uuid.UUID) error { return nil },
		})
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	live := findChoice(g, id)
	if live == nil || live.Chooser != next.ID {
		t.Fatalf("setup: the prompt did not reach seat 2")
	}

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	got := findChoice(restored, id)
	if got == nil {
		t.Fatalf("the reassigned prompt did not survive the round trip")
	}
	if got.Chooser != next.ID {
		t.Errorf("restored Chooser = %s, want %s", got.Chooser, next.ID)
	}
	if len(got.ChooseCards) != len(live.ChooseCards) || got.ChooseMax != live.ChooseMax {
		t.Errorf("restored candidates = %d/max %d, want %d/max %d",
			len(got.ChooseCards), got.ChooseMax, len(live.ChooseCards), live.ChooseMax)
	}
}

// --- CR 800.4i: last known information ------------------------------

// TestDepartedPlayersLifeTotalIsLastKnownInformation — CR 800.4i: "if
// an effect requires information about a specific player, the effect
// uses the current information about that player if they are still in
// the game; otherwise, the effect uses the LAST KNOWN INFORMATION about
// that player before they left the game."
//
// The engine gets this for free and this test is here to keep it that
// way: a departed seat stays in g.Seats with its life total, poison and
// commander damage as they were, so every effect that reads p.Life
// through PlayerByID reads last known information by construction.
// Nothing zeroes a conceding player's life, and nothing may start.
func TestDepartedPlayersLifeTotalIsLastKnownInformation(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	g.WithWriteLock(func() {
		leaver.Life = 37
		leaver.Poison = 3
	})

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	p := g.PlayerByID(leaver.ID)
	if p == nil {
		t.Fatalf("a departed player is no longer readable at all; CR 800.4i needs their last known information")
	}
	if p.Life != 37 {
		t.Errorf("departed player's life = %d, want their last known 37", p.Life)
	}
	if p.Poison != 3 {
		t.Errorf("departed player's poison = %d, want their last known 3", p.Poison)
	}
	if !p.Eliminated {
		t.Errorf("the departed seat is not marked eliminated")
	}
}

// TestActionsOfADepartedPlayerAreStillFindable — CR 800.4i's second
// sentence: "if an effect requires information from the game about
// actions players have taken, the effect can find actions that were
// taken by a player who has left the game." The event log is
// append-only and the per-turn tallies are keyed by player, so a
// departure erases neither.
func TestActionsOfADepartedPlayerAreStillFindable(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	castCardID := uuid.New()
	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]CastTally)
		}
		g.SpellsCastThisTurn[leaver.ID] = CastTally{Total: 2, Noncreature: 1}
		g.EmitEvent(Event{Kind: EventCast, Actor: leaver.ID, CardID: castCardID})
	})
	before := g.SpellsCastThisTurn[leaver.ID].Total

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if got := g.SpellsCastThisTurn[leaver.ID].Total; got != before || got == 0 {
		t.Errorf("spells cast by the departed player this turn = %d, want the %d they had cast", got, before)
	}
	found := false
	for _, ev := range g.Events {
		if ev.Kind == EventCast && ev.Actor == leaver.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("the departed player's cast is no longer in the event log")
	}
}

// TestDepartedPlayersObjectsAreGoneNotLastKnown pins the ONE place the
// engine deviates from CR 800.4i, so the deviation is a decision rather
// than a surprise. CR 800.4a is unconditional — the objects leave the
// game — and CR 800.4i then says an effect asking about that player
// uses last known information. The engine keeps the player's SCALARS
// (life, poison, counters: the test above) and does not keep a
// pre-departure copy of their BOARD, so "creatures that player
// controls" reads zero rather than what they had.
//
// Recorded in the ADR 0060 amendment as out of scope: closing it means
// a per-player LKI snapshot taken at the moment of departure, and no
// card in the catalog asks the question.
func TestDepartedPlayersObjectsAreGoneNotLastKnown(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	departureTestSource(g, leaver.ID, "Grizzly Bears")

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Owner == leaver.ID {
				t.Fatalf("CR 800.4a did not take the departed player's permanent out of the game")
			}
		}
	})
	// The documented consequence, stated as a test so a future change
	// that adds a board LKI snapshot has to come here and say so.
	count := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == leaver.ID {
				count++
			}
		}
	})
	if count != 0 {
		t.Errorf("permanents controlled by the departed player = %d, want 0 (CR 800.4a)", count)
	}
}

// --- CR 800.4m: "until that player's next turn" ----------------------

// TestUntilThatPlayersNextTurnWhenTheyLeaveOnTheirOwnTurn is CR 800.4m
// at the boundary ADR 0063 did not cover: the effect is created during
// the departing player's OWN turn, and that turn then ends early
// because its active player left (ADR 0059 Decision 6). The duration is
// stamped TurnsBegun+1, so it must last until the rotation next steps
// over their seat — a full round later — rather than ending with the
// turn it was made in or lasting for the rest of the game.
//
// 800.4m itself needed no code: #921 / ADR 0063 Decision 3 already bumps
// Player.TurnsBegun for a seat the rotation steps over
// (beginNextTurnLocked), which is what ends the effect at the right
// moment. See also
// TestUntilYourNextTurnEndsWhenADepartedPlayersTurnWouldHaveBegun.
func TestUntilThatPlayersNextTurnWhenTheyLeaveOnTheirOwnTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	active := g.Seats[0]
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: active seat is %d, want 0", g.Turn.ActiveSeat)
	}
	bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)

	var d Duration
	g.WithWriteLock(func() { d = g.UntilYourNextTurnDuration(active.ID) })
	registerDurationMarker(g, bear, d, "until the active player's next turn")

	// The active player concedes: their turn ends at once and seat 1's
	// begins. The effect is keyed to seat 0's NEXT turn, which is a
	// round away.
	if err := g.Concede(active.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := len(g.ScopedStatics); n != 1 {
		t.Fatalf("the effect ended with the turn it was made in (%d entries); CR 800.4m says it waits", n)
	}
	for seat := 2; seat <= 3; seat++ {
		advanceOneTurn(t, g)
		if g.Turn.ActiveSeat != seat {
			t.Fatalf("expected seat %d's turn, got %d", seat, g.Turn.ActiveSeat)
		}
		if n := len(g.ScopedStatics); n != 1 {
			t.Fatalf("the effect ended during seat %d's turn (%d entries)", seat, n)
		}
	}
	// Seat 0 has left, so the rotation steps over it — and that
	// never-taken turn is what ends the effect.
	advanceOneTurn(t, g)
	if g.Turn.ActiveSeat != 1 {
		t.Fatalf("expected the rotation to step over seat 0 onto seat 1, got %d", g.Turn.ActiveSeat)
	}
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("the departed player's turn would have begun and the effect is still live (%d entries)", n)
	}
}
