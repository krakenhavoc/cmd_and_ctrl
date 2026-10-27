package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

type orderedPermanentCall struct {
	runs int
	got  []PromptedPicks
}

// beginOrderedPermanentChoices starts a keyed ordered collection whose
// continuation records each answer in out. (#4361 S7b: the closure
// collector, ChoosePermanentsInOrderThenForEffect, is gone; these tests ask
// the keyed one, ChoosePermanentsInOrderKeyedForEffect.)
func beginOrderedPermanentChoices(t *testing.T, g *Game, legs []ChooseCardsPrompt, out *orderedPermanentCall) error {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		// The asking card is the first leg's Source, as the closure
		// collector read it.
		var source uuid.UUID
		if len(legs) > 0 {
			source = legs[0].Source
		}
		err = g.ChoosePermanentsInOrderKeyedForEffect(legs, testRunHookThen, source, runHook(func(_ *Game, r RunDone) error {
			out.runs++
			out.got = append(out.got, r.Picks())
			return nil
		}))
	})
	return err
}

func orderedPermanentPrompt(g *Game, chooser uuid.UUID) *PendingChoice {
	for _, choice := range g.PendingChoices {
		if choice != nil && choice.Kind == PendingChoiceChooseCards && choice.Chooser == chooser {
			return choice
		}
	}
	return nil
}

func orderedPermanentLeg(chooser uuid.UUID, cards []uuid.UUID, min, max int) ChooseCardsPrompt {
	return ChooseCardsPrompt{Chooser: chooser, Cards: cards, Min: min, Max: max, Zone: ZoneBattlefield, Question: "Choose a permanent"}
}

func answerOrderedPermanent(t *testing.T, g *Game, chooser uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	prompt := orderedPermanentPrompt(g, chooser)
	if prompt == nil {
		t.Fatalf("seat %s has no ordered permanent prompt", chooser)
	}
	if err := g.ResolveChooseCards(prompt.ID, chooser, cards); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

func movePermanentAwayAndBack(t *testing.T, g *Game, owner uuid.UUID, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: owner}, id); err != nil {
		t.Fatalf("move permanent to graveyard: %v", err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneGraveyard, Owner: owner}, ZoneRef{Kind: ZoneBattlefield}, id); err != nil {
		t.Fatalf("return permanent to battlefield: %v", err)
	}
}

func TestOrderedPermanentChoicesAskOneSeatAtATimeAndFinishAfterTheLastPick(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	aCard, bCard := pushBear(g, a.ID), pushBear(g, b.ID)
	out := &orderedPermanentCall{}
	if err := beginOrderedPermanentChoices(t, g, []ChooseCardsPrompt{
		orderedPermanentLeg(a.ID, []uuid.UUID{aCard}, 0, 1),
		orderedPermanentLeg(b.ID, []uuid.UUID{bCard}, 0, 1),
	}, out); err != nil {
		t.Fatalf("start choices: %v", err)
	}
	if orderedPermanentPrompt(g, a.ID) == nil || orderedPermanentPrompt(g, b.ID) != nil {
		t.Fatal("only the first seat should have a prompt")
	}
	if out.runs != 0 {
		t.Fatalf("tail ran %d times before any answer", out.runs)
	}
	answerOrderedPermanent(t, g, a.ID, aCard)
	second := orderedPermanentPrompt(g, b.ID)
	if second == nil || orderedPermanentPrompt(g, a.ID) != nil {
		t.Fatal("answering the first seat should open only the second prompt")
	}
	if out.runs != 0 {
		t.Fatalf("tail ran %d times while the second prompt remained open", out.runs)
	}
	earlier := second.EarlierPermanentChoicesForEffect()
	if len(earlier) != 1 || earlier[0].Seat != a.ID || len(earlier[0].Cards) != 1 || earlier[0].Cards[0] != aCard {
		t.Fatalf("earlier choices = %+v, want %s choosing %s", earlier, a.Name, aCard)
	}
	answerOrderedPermanent(t, g, b.ID, bCard)
	if out.runs != 1 || len(out.got) != 1 {
		t.Fatalf("tail ran %d times, want exactly once after the final pick", out.runs)
	}
	if len(out.got[0]) != 2 || out.got[0][0].Seat != a.ID || out.got[0][1].Seat != b.ID || out.got[0].Count() != 2 {
		t.Fatalf("final picks = %+v, want both seats in supplied order", out.got[0])
	}
}

func TestOrderedPermanentChoicesAutoAcceptForcedSetsAndSkipEmptyLegs(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, empty, b := g.Seats[0], g.Seats[1], g.Seats[2]
	aCard, bCard := pushBear(g, a.ID), pushBear(g, b.ID)
	var unavailable uuid.UUID
	out := &orderedPermanentCall{}
	if err := beginOrderedPermanentChoices(t, g, []ChooseCardsPrompt{
		orderedPermanentLeg(a.ID, []uuid.UUID{aCard}, 1, 1),
		orderedPermanentLeg(empty.ID, []uuid.UUID{unavailable}, 0, 1),
		orderedPermanentLeg(b.ID, []uuid.UUID{bCard}, 0, 1),
	}, out); err != nil {
		t.Fatalf("start choices: %v", err)
	}
	if orderedPermanentPrompt(g, a.ID) != nil || orderedPermanentPrompt(g, empty.ID) != nil || orderedPermanentPrompt(g, b.ID) == nil {
		t.Fatal("forced and empty legs should be settled inline; only the final optional leg should prompt")
	}
	if out.runs != 0 {
		t.Fatal("tail ran before the final prompted choice")
	}
	answerOrderedPermanent(t, g, b.ID, bCard)
	if out.runs != 1 || out.got[0].Count() != 2 || out.got[0][0].Seat != a.ID || out.got[0][1].Seat != b.ID {
		t.Fatalf("tail result = %d runs, %+v; want forced A and chosen B, without empty seat", out.runs, out.got)
	}

	emptyOut := &orderedPermanentCall{}
	if err := beginOrderedPermanentChoices(t, g, []ChooseCardsPrompt{
		orderedPermanentLeg(empty.ID, nil, 1, 1),
	}, emptyOut); err != nil {
		t.Fatalf("start empty choices: %v", err)
	}
	if emptyOut.runs != 1 || len(emptyOut.got) != 1 || len(emptyOut.got[0]) != 0 {
		t.Fatalf("all-empty run = %d, %+v; want one inline empty tail", emptyOut.runs, emptyOut.got)
	}
}

func TestOrderedPermanentChoicesRejectsANonBattlefieldZoneOrADuplicate(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := pushBear(g, me.ID)
	for name, mutate := range map[string]func(*ChooseCardsPrompt){
		"non-battlefield zone": func(p *ChooseCardsPrompt) { p.Zone = ZoneGraveyard },
		"duplicate candidate":  func(p *ChooseCardsPrompt) { p.Cards = append(p.Cards, card) },
	} {
		t.Run(name, func(t *testing.T) {
			leg := orderedPermanentLeg(me.ID, []uuid.UUID{card}, 0, 1)
			mutate(&leg)
			out := &orderedPermanentCall{}
			if err := beginOrderedPermanentChoices(t, g, []ChooseCardsPrompt{leg}, out); !errors.Is(err, ErrInvalidParam) {
				t.Fatalf("helper error = %v, want ErrInvalidParam", err)
			}
			if len(g.PendingChoices) != 0 || out.runs != 0 {
				t.Fatalf("invalid leg created prompt/tail: prompts=%d tail=%d", len(g.PendingChoices), out.runs)
			}
		})
	}
}

func TestOrderedPermanentChoicesAdvancesWhenChooserConcedes(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b, caster := g.Seats[0], g.Seats[1], g.Seats[2]
	aCard, bCard := pushBear(g, a.ID), pushBear(g, b.ID)
	aOther, bOther := pushBear(g, a.ID), pushBear(g, b.ID)
	source := departureTestSource(g, caster.ID, "Liliana, Dreadhorde General")
	first := orderedPermanentLeg(a.ID, []uuid.UUID{aCard, aOther}, 1, 1)
	first.Source = source // the source survives the affected player's concession
	out := &orderedPermanentCall{}
	if err := beginOrderedPermanentChoices(t, g, []ChooseCardsPrompt{
		first,
		orderedPermanentLeg(b.ID, []uuid.UUID{bCard, bOther}, 1, 1),
	}, out); err != nil {
		t.Fatalf("start choices: %v", err)
	}
	if err := g.Concede(a.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if orderedPermanentPrompt(g, b.ID) == nil || out.runs != 0 {
		t.Fatal("conceding chooser should settle its leg and advance to the next prompt")
	}
	answerOrderedPermanent(t, g, b.ID, bCard)
	if out.runs != 1 || out.got[0].Picked(a.ID) || !out.got[0].Picked(b.ID) {
		t.Fatalf("post-concession tail = %d, %+v; want only B's pick", out.runs, out.got)
	}
}

func TestOrderedPermanentChoicesDropStaleObjectEpochsFromCandidatesAndReceipts(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	aCard, bCard := pushBear(g, a.ID), pushBear(g, b.ID)
	aOther, bOther := pushBear(g, a.ID), pushBear(g, b.ID)
	out := &orderedPermanentCall{}
	if err := beginOrderedPermanentChoices(t, g, []ChooseCardsPrompt{
		orderedPermanentLeg(a.ID, []uuid.UUID{aCard, aOther}, 1, 1),
		orderedPermanentLeg(b.ID, []uuid.UUID{bCard, bOther}, 1, 1),
	}, out); err != nil {
		t.Fatalf("start choices: %v", err)
	}
	answerOrderedPermanent(t, g, a.ID, aCard)
	movePermanentAwayAndBack(t, g, a.ID, aCard)
	answerOrderedPermanent(t, g, b.ID, bCard)
	if out.runs != 1 || out.got[0].Count() != 1 || !out.got[0].Picked(b.ID) || out.got[0].Picked(a.ID) {
		t.Fatalf("tail receipts = %d, %+v; returned object should not count as A's original", out.runs, out.got)
	}

	// A prompt's original candidate is equally epoch-bound. A card that
	// leaves and returns after the prompt opens is not the offered object.
	candidate, alternative := pushBear(g, a.ID), pushBear(g, a.ID)
	staleOut := &orderedPermanentCall{}
	if err := beginOrderedPermanentChoices(t, g, []ChooseCardsPrompt{
		orderedPermanentLeg(a.ID, []uuid.UUID{candidate, alternative}, 0, 1),
	}, staleOut); err != nil {
		t.Fatalf("start stale-candidate case: %v", err)
	}
	prompt := orderedPermanentPrompt(g, a.ID)
	if prompt == nil {
		t.Fatal("expected optional prompt")
	}
	movePermanentAwayAndBack(t, g, a.ID, candidate)
	if err := g.ResolveChooseCards(prompt.ID, a.ID, []uuid.UUID{candidate}); err == nil {
		t.Fatal("stale candidate answer was accepted")
	}
	if orderedPermanentPrompt(g, a.ID) == nil || staleOut.runs != 0 {
		t.Fatal("refusing a stale object should leave the prompt open without running the tail")
	}
}

func TestEarlierPermanentChoiceGetterReturnsFreshCopies(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	aCard, bCard := pushBear(g, a.ID), pushBear(g, b.ID)
	aOther := pushBear(g, a.ID)
	out := &orderedPermanentCall{}
	if err := beginOrderedPermanentChoices(t, g, []ChooseCardsPrompt{
		orderedPermanentLeg(a.ID, []uuid.UUID{aCard, aOther}, 1, 1),
		orderedPermanentLeg(b.ID, []uuid.UUID{bCard}, 0, 1),
	}, out); err != nil {
		t.Fatalf("start choices: %v", err)
	}
	answerOrderedPermanent(t, g, a.ID, aCard)
	prompt := orderedPermanentPrompt(g, b.ID)
	first := prompt.EarlierPermanentChoicesForEffect()
	first[0].Seat = uuid.Nil
	first[0].Cards[0] = uuid.Nil
	first[0].Cards = append(first[0].Cards, uuid.New())
	second := prompt.EarlierPermanentChoicesForEffect()
	if len(second) != 1 || second[0].Seat != a.ID || len(second[0].Cards) != 1 || second[0].Cards[0] != aCard {
		t.Fatalf("mutating getter result changed stored choices: %+v", second)
	}
}

func TestOrderedPermanentChoicesCloneUndoReplaysTailIndependently(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	aCard, bCard := pushBear(g, a.ID), pushBear(g, b.ID)
	aOther, bOther := pushBear(g, a.ID), pushBear(g, b.ID)
	out := &orderedPermanentCall{}
	if err := beginOrderedPermanentChoices(t, g, []ChooseCardsPrompt{
		orderedPermanentLeg(a.ID, []uuid.UUID{aCard, aOther}, 1, 1),
		orderedPermanentLeg(b.ID, []uuid.UUID{bCard, bOther}, 1, 1),
	}, out); err != nil {
		t.Fatalf("start choices: %v", err)
	}
	answerOrderedPermanent(t, g, a.ID, aCard)
	var midway *Game
	g.WithWriteLock(func() { midway = g.cloneLocked() })
	answerOrderedPermanent(t, g, b.ID, bCard)
	if out.runs != 1 || out.got[0].Count() != 2 {
		t.Fatalf("first branch = %d, %+v", out.runs, out.got)
	}
	g.WithWriteLock(func() { g.RestoreFrom(midway) })
	if out.runs != 1 {
		t.Fatalf("restoring halfway reran tail: %d", out.runs)
	}
	answerOrderedPermanent(t, g, b.ID, bCard)
	if out.runs != 2 || len(out.got) != 2 || out.got[1].Count() != 2 || out.got[1].By(a.ID)[0] != aCard {
		t.Fatalf("replayed branch = %d, %+v; each branch should finish once with its own receipt history", out.runs, out.got)
	}
}
