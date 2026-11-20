package game

import (
	"testing"

	"github.com/google/uuid"
)

// spell_copy_not_created_test.go — #4994, the owner's 2026-11-19 decision
// (ADR 0043's 2026-11-20 amendment). CR 111.13: "A copy of a permanent
// spell becomes a token as it resolves. ... The token is not 'created' for
// the purposes of any replacement effects or triggered abilities that
// refer to creating a token." CR 608.3f extends that to "any rules or
// effects". So the token takes no RepEventCreateTokens window and is not
// announced as EventTokenCreated, but it still enters: entry replacements
// apply, EventETB fires, and a pause in its entry is a restore point.

// castBearsAndCopyIt casts a {0} Grizzly Bears for `me`, copies it under
// `me`'s control (CR 707.10) and leaves both on the stack, the copy on top.
func castBearsAndCopyIt(t *testing.T, g *Game, me *Player) uuid.UUID {
	t.Helper()
	bears := typedCard(me.ID, "Grizzly Bears", "Creature — Bear")
	bears.Power, bears.Toughness = 2, 2
	g.WithWriteLock(func() {
		me.Hand.PushTop(bears)
		g.markCardKnownInZoneLocked(me.Hand, bears.InstanceID)
	})
	if err := g.CastSpell(me.ID, bears.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast Grizzly Bears: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(bears.InstanceID, me.ID, false, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
	return bears.InstanceID
}

// bearTokens is the Grizzly Bears tokens on the battlefield.
func bearTokens(g *Game) []Card {
	var out []Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Grizzly Bears" && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

// copyTokenArrival counts, for the token `id`, the events that announce it
// as created and as a move from no zone, and its EventETB.
func copyTokenArrival(g *Game, id uuid.UUID) (created, became, etb int) {
	for _, ev := range g.Events {
		if ev.CardID != id {
			continue
		}
		switch {
		case ev.Kind == EventTokenCreated:
			created++
		case ev.Kind == EventZoneMove && ev.BecameToken && ev.OldZone == "" && ev.NewZone == ZoneBattlefield:
			became++
		case ev.Kind == EventETB:
			etb++
		}
	}
	return created, became, etb
}

// A doubler under the copy's controller does not double the token, and
// is not even asked about; an entry replacement still applies to it, and
// it is announced as a move from no zone, never as a creation, so the
// tokens-created tally does not count it.
func TestTheTokenACopyBecomesIsNotCreated(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"parallel-lives": {tokenDoubler("double tokens")},
		"renata":         {entersWithACounter("enters with a counter")},
	})
	pushReplacementSource(g, "parallel-lives", me.ID)
	pushReplacementSource(g, "renata", me.ID)

	original := castBearsAndCopyIt(t, g, me)
	passUntilStackEmpty(t, g)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts open, want none", len(g.PendingChoices))
	}
	toks := bearTokens(g)
	if len(toks) != 1 {
		t.Fatalf("%d token copies, want 1: the doubler must not see a token that is not created (CR 111.13)", len(toks))
	}
	if !g.Battlefield.Contains(original) {
		t.Error("the original Grizzly Bears did not resolve")
	}
	tok := toks[0]
	if tok.Counters[CounterPlusOne] != 1 {
		t.Errorf("the token has %d +1/+1 counters, want 1: it still takes the ordinary entry window", tok.Counters[CounterPlusOne])
	}
	created, became, etb := copyTokenArrival(g, tok.InstanceID)
	if created != 0 {
		t.Errorf("%d EventTokenCreated for the copy's token, want none (CR 111.13)", created)
	}
	if became != 1 {
		t.Errorf("%d became-token moves onto the battlefield, want 1", became)
	}
	if etb != 1 {
		t.Errorf("%d EventETB for the token, want 1: its enters abilities still trigger", etb)
	}
	if n := g.TurnTally.Players[me.ID].TokensCreated; n != 0 {
		t.Errorf("tokens created this turn = %d, want 0 (CR 608.3f)", n)
	}
}

// The token's entry pauses on a CR 616 order prompt (two enters-tapped
// effects an opponent controls). The table is a restore point there, and
// the restored answer still puts one token onto the battlefield that is
// not announced as created: the run's NotCreated rode the disk.
func TestTheCopysTokenIsStillNotCreatedAfterARestoreMidEntry(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"kismet":         {entersTappedForOthers("Kismet: enters tapped")},
		"thalia":         {entersTappedForOthers("Thalia: enters tapped")},
		"parallel-lives": {tokenDoubler("double tokens")},
	})
	pushReplacementSource(g, "kismet", opp.ID)
	pushReplacementSource(g, "thalia", opp.ID)
	pushReplacementSource(g, "parallel-lives", me.ID)

	castBearsAndCopyIt(t, g, me)
	passUntilPrompt(t, g)
	onlyReplacementOrderPrompt(t, g, me.ID)
	if n := len(bearTokens(g)); n != 0 {
		t.Fatalf("%d tokens landed before the order was answered", n)
	}

	framed := false
	for _, c := range g.CaptureSnapshot().PendingChoices {
		framed = framed || c.PickThen == ReplacementResumeMoveTail
	}
	if !framed {
		t.Fatalf("the open prompt is not written as the entry's frame %q", ReplacementResumeMoveTail)
	}
	g = strictRestore(t, g)
	me = g.Seats[0]
	answerOnlyOrder(t, g, me.ID)
	toks := bearTokens(g)
	if len(toks) != 1 {
		t.Fatalf("%d token copies after the restored answer, want 1", len(toks))
	}
	tok := toks[0]
	if !tok.Tapped {
		t.Error("the restored token entered untapped: the answered entry replacements did not apply")
	}
	created, became, _ := copyTokenArrival(g, tok.InstanceID)
	if created != 0 || became != 1 {
		t.Errorf("after the restore: %d creations and %d became-token moves, want 0 and 1 (CR 111.13)", created, became)
	}
	if n := g.TurnTally.Players[me.ID].TokensCreated; n != 0 {
		t.Errorf("tokens created this turn = %d, want 0 (CR 608.3f)", n)
	}
}
