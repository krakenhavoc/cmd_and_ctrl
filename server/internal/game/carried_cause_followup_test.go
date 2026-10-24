package game

import (
	"testing"

	"github.com/google/uuid"
)

// carried_cause_followup_test.go — #4281. A keyed pay-unless runs its
// branch under the cause it was queued with (#4279). A prompt that branch
// queues — a confirm, an option pick, a choose-cards pick — is answered
// later still, with the table moved on, so it carries that cause too, and
// the discard or sacrifice its answer queues next names the spell that
// asked and meets the "can't cause you to discard" gate.

// followupDiscard is what every follow-up's answer does: the payer
// discards one of two cards, by their own choice.
func followupDiscard(g *Game, who, source uuid.UUID) error {
	g.QueueDiscardChoiceForEffect(DiscardPrompt{Player: who, Source: source, N: 1})
	return nil
}

var (
	followupConfirm = RegisterConfirmThen(testEffectKeyPrefix+"followup-confirm", func(g *Game, r ConfirmAnswer) error {
		return followupDiscard(g, r.Chooser, r.Source)
	})
	followupOption = RegisterOptionPickThen(testEffectKeyPrefix+"followup-option", func(g *Game, r OptionPicked) error {
		return followupDiscard(g, r.Chooser, r.Source)
	})
	followupCards = RegisterChooseCardsThen(testEffectKeyPrefix+"followup-cards", func(g *Game, r CardsChosen) error {
		return followupDiscard(g, r.Chooser, r.Source)
	})

	// The pay-unless decline queues the follow-up. It reads only what the
	// prompt hands it.
	followupViaConfirm = RegisterPayUnlessElse(testEffectKeyPrefix+"followup-via-confirm", func(g *Game, r PayUnlessAnswer) error {
		_, err := g.QueueConfirmKeyedForEffect(ConfirmPrompt{Chooser: r.Chooser, Source: r.Source, Question: "Discard a card?"}, followupConfirm, nil)
		return err
	})
	followupViaOption = RegisterPayUnlessElse(testEffectKeyPrefix+"followup-via-option", func(g *Game, r PayUnlessAnswer) error {
		g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser: r.Chooser, Source: r.Source, Question: "Discard a card?",
			Options: []ChoiceOption{{Label: "Yes"}, {Label: "No"}}, ThenKey: followupOption,
		})
		return nil
	})
	followupViaCards = RegisterPayUnlessElse(testEffectKeyPrefix+"followup-via-cards", func(g *Game, r PayUnlessAnswer) error {
		g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser: r.Chooser, Source: r.Source, Question: "Pick a card",
			Cards: r.Carry, Min: 1, Max: 1, Zone: ZoneHand, ThenKey: followupCards,
		})
		return nil
	})
)

// openFollowupChain has seat 1's spell resolve and ask seat 0 to pay {2}
// or run `via`, closes the resolution and walks a step on. Seat 0 holds
// two cards. Returns the table, the asking item, the prompt and the hand.
func openFollowupChain(t *testing.T, via PayUnlessElse) (g *Game, item *StackItem, promptID uuid.UUID, hand []uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	payer, caster := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		h := NewCard("Spare", payer.ID)
		payer.Hand.PushTop(h)
		hand = append(hand, h.InstanceID)
	}
	item = &StackItem{ID: uuid.New(), Controller: caster.ID, SourceCardID: uuid.New()}
	g.WithWriteLock(func() {
		g.beginResolvingLocked(item)
		g.beginResolutionLocked()
		if err := g.QueuePayUnlessKeyedForEffect(payer.ID, item.SourceCardID, "{2}", "Pay {2}?", via, hand); err != nil {
			t.Fatalf("QueuePayUnlessKeyedForEffect: %v", err)
		}
		g.endResolutionLocked()
		g.holdForOpenResolutionLocked()
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePayUnless {
			promptID = c.ID
		}
	}
	if promptID == uuid.Nil {
		t.Fatal("no pay-unless prompt queued")
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	g.WithWriteLock(func() {
		if g.resolutionOpen || g.resolving != nil {
			t.Fatalf("the resolution is still open after the step advance (open=%v)", g.resolutionOpen)
		}
	})
	return g, item, promptID, hand
}

// openPromptOfKind returns the open prompt of a kind, or nil.
func openPromptOfKind(g *Game, kind PendingChoiceKind) *PendingChoice {
	var out *PendingChoice
	g.WithWriteLock(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == kind {
				out = c
			}
		}
	})
	return out
}

// followupAnswers answers the follow-up a pay-unless decline queued, by
// kind, and returns once the discard prompt it queues is open.
var followupAnswers = []struct {
	name string
	via  PayUnlessElse
	kind PendingChoiceKind
	ask  func(t *testing.T, g *Game, c *PendingChoice) error
}{
	{"confirm", followupViaConfirm, PendingChoiceConfirm, func(_ *testing.T, g *Game, c *PendingChoice) error {
		return g.ResolveConfirm(c.ID, c.Chooser, true)
	}},
	{"option pick", followupViaOption, PendingChoiceOptionPick, func(_ *testing.T, g *Game, c *PendingChoice) error {
		return g.ResolveOptionPick(c.ID, c.Chooser, 0)
	}},
	{"choose cards", followupViaCards, PendingChoiceChooseCards, func(_ *testing.T, g *Game, c *PendingChoice) error {
		return g.ResolveChooseCards(c.ID, c.Chooser, c.ChooseCards[:1])
	}},
}

// The follow-up prompt, answered after the resolution closed, still
// leads to a discard that names the asking spell.
func TestFollowUpPromptOfAKeyedBranchKeepsTheCause(t *testing.T) {
	for _, tc := range followupAnswers {
		t.Run(tc.name, func(t *testing.T) {
			g, item, id, hand := openFollowupChain(t, tc.via)
			payer := g.Seats[0]
			if err := g.ResolvePayUnless(id, payer.ID, false); err != nil {
				t.Fatalf("ResolvePayUnless: %v", err)
			}
			follow := openPromptOfKind(g, tc.kind)
			if follow == nil {
				t.Fatalf("the decline queued no %s prompt", tc.kind)
			}
			if follow.PayCauseController != item.Controller || follow.PayCauseItem != item.ID {
				t.Fatalf("follow-up carries cause (%s, %s), want the asking item (%s, %s)",
					follow.PayCauseController, follow.PayCauseItem, item.Controller, item.ID)
			}
			if err := tc.ask(t, g, follow); err != nil {
				t.Fatalf("answering the %s: %v", tc.name, err)
			}
			disc := openPromptOfKind(g, PendingChoiceChooseCards)
			if disc == nil {
				t.Fatal("the follow-up's answer queued no discard prompt")
			}
			if err := g.ResolveChooseCards(disc.ID, payer.ID, hand[:1]); err != nil {
				t.Fatalf("answering the discard: %v", err)
			}
			if payer.Hand.Contains(hand[0]) {
				t.Fatal("the discard did not happen")
			}
			ev := discardEventOf(g, payer.ID)
			if ev == nil {
				t.Fatal("no discard event")
			}
			if ev.Cause != MoveCauseEffect || ev.CauseController != item.Controller || ev.CauseItem != item.ID {
				t.Errorf("discard cause = (%q, %s, %s), want the asking spell (effect, %s, %s)",
					ev.Cause, ev.CauseController, ev.CauseItem, item.Controller, item.ID)
			}
			if !DiscardCausedByOpponentOf(*ev, payer.ID) {
				t.Error("the discard is not read as an opponent's spell causing it")
			}
			g.WithWriteLock(func() {
				if g.carriedCause != nil {
					t.Error("the carried cause outlived the answer")
				}
			})
		})
	}
}

// With the confirm saved and restored in between, the discard still
// names the asking spell.
func TestFollowUpPromptCauseSurvivesARestore(t *testing.T) {
	g, item, id, hand := openFollowupChain(t, followupViaConfirm)
	payer := g.Seats[0]
	if err := g.ResolvePayUnless(id, payer.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	_, g = roundTrip(t, g)
	payer = g.Seats[0]
	follow := openPromptOfKind(g, PendingChoiceConfirm)
	if follow == nil {
		t.Fatal("the restored table lost the confirm")
	}
	if err := g.ResolveConfirm(follow.ID, payer.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	disc := openPromptOfKind(g, PendingChoiceChooseCards)
	if disc == nil {
		t.Fatal("no discard prompt after the restore")
	}
	if err := g.ResolveChooseCards(disc.ID, payer.ID, hand[:1]); err != nil {
		t.Fatalf("answering the discard: %v", err)
	}
	ev := discardEventOf(g, payer.ID)
	if ev == nil || ev.CauseController != item.Controller || ev.CauseItem != item.ID {
		t.Fatalf("restored follow-up's discard event = %+v, want the asking spell", ev)
	}
}

// "Spells and abilities your opponents control can't cause you to
// discard" stops the follow-up's discard too, with the slot long gone.
func TestFollowUpPromptIsStillGatedByACantCauseDiscard(t *testing.T) {
	withProtection(t)
	g, _, id, hand := openFollowupChain(t, followupViaConfirm)
	payer := g.Seats[0]
	shield := NewCard("Shield", payer.ID)
	shield.TypeLine = "Enchantment"
	shield.OracleID = protectedKey
	g.Battlefield.PushTop(shield)

	if err := g.ResolvePayUnless(id, payer.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	follow := openPromptOfKind(g, PendingChoiceConfirm)
	if follow == nil {
		t.Fatal("no confirm queued")
	}
	if err := g.ResolveConfirm(follow.ID, payer.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	for _, h := range hand {
		if !payer.Hand.Contains(h) {
			t.Fatal("an opponent's follow-up discarded a protected player's card")
		}
	}
	if openPromptOfKind(g, PendingChoiceChooseCards) != nil {
		t.Error("a discard prompt was queued for a player the gate protects")
	}
}

// effectCauseControllerLocked reads the resolving slot only while a
// resolution is open (the gates share resolutionCauseLocked's reading).
func TestEffectCauseControllerIgnoresAClosedResolution(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		g.resolving = &resolvingItem{item: &StackItem{ID: uuid.New(), Controller: opp.ID}}
		g.resolutionOpen = false
		if got := g.effectCauseControllerLocked(); got != uuid.Nil {
			t.Errorf("a closed resolution's slot gave cause controller %s", got)
		}
		g.resolutionOpen = true
		if got := g.effectCauseControllerLocked(); got != opp.ID {
			t.Errorf("an open resolution gave cause controller %s, want %s", got, opp.ID)
		}
	})
}
