package game

import (
	"testing"

	"github.com/google/uuid"
)

// token_create_together_test.go — #4867 (ADR 0061 amendment 2026-11-05):
// the tokens one creation instruction makes are created at the same time
// (CR 701.7a, CR 608.2f), so every creation of more than one token enters
// them in ONE simultaneous entry, whatever the caller asked for. One
// entry is one entry ordinal (ADR 0109 §8). The batch keeps the per-token
// clauses a token entering on its own has, such as the host an Aura token
// is created attached to. Tokens created blocking are the one exception:
// they still enter one at a time, and each still blocks.

// tctNewTokens lists the battlefield cards that are not in before.
func tctNewTokens(g *Game, before map[uuid.UUID]bool) []Card {
	var out []Card
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] {
			out = append(out, c)
		}
	}
	return out
}

// tctBattlefieldIDs is the set of IDs on the battlefield now.
func tctBattlefieldIDs(g *Game) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		out[c.InstanceID] = true
	}
	return out
}

func TestCreatingSeveralTokensEntersThemTogether(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	before := tctBattlefieldIDs(g)

	var ids []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		ids, err = g.CreateTokensForEffect(me, Card{
			Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1,
		}, 3, TokenEntryOptions{Tapped: true})
		if err != nil {
			t.Fatalf("CreateTokensForEffect: %v", err)
		}
	})
	if len(ids) != 3 {
		t.Fatalf("the door reported %d tokens, want 3", len(ids))
	}
	tokens := tctNewTokens(g, before)
	if len(tokens) != 3 {
		t.Fatalf("%d tokens entered, want 3", len(tokens))
	}
	ordinal := tokens[0].EntryOrdinal
	for _, tok := range tokens {
		if tok.EntryOrdinal != ordinal || ordinal == 0 {
			t.Errorf("token entry ordinals %d and %d: one creation is one simultaneous entry (CR 608.2f)", tok.EntryOrdinal, ordinal)
		}
		if !tok.Tapped {
			t.Error("a token lost the creation's own \"tapped\" clause")
		}
	}

	// Two creations are two entries.
	before = tctBattlefieldIDs(g)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me, Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}, 1); err != nil {
			t.Fatal(err)
		}
	})
	next := tctNewTokens(g, before)
	if len(next) != 1 || next[0].EntryOrdinal == ordinal {
		t.Errorf("a later creation shares the first one's entry: %+v", next)
	}
}

// Brimaz's Cat Soldier, doubled: both tokens are created blocking the
// attacker, the attacker becomes blocked once (CR 509.3c), and neither
// token was declared as a blocker (CR 509.4). A paused simultaneous entry
// has no field for what a token blocks, so this creation keeps entering
// its tokens one at a time (ADR 0061 amendment 2026-11-05).
func TestSeveralTokensCreatedBlockingAllBlock(t *testing.T) {
	g := newActiveGame(t)
	attackerA, _, _, _ := blockSetup(t, g)
	defender := g.Seats[1].ID
	if err := g.FinishBlocks(defender); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	before := tctBattlefieldIDs(g)

	g.WithWriteLock(func() {
		if err := g.CreateTokensBlockingForEffect(defender, Card{
			Name: "Cat Soldier", TypeLine: "Token Creature — Cat Soldier", Power: 1, Toughness: 1,
		}, 2, attackerA); err != nil {
			t.Fatalf("CreateTokensBlockingForEffect: %v", err)
		}
	})
	tokens := tctNewTokens(g, before)
	if len(tokens) != 2 {
		t.Fatalf("%d tokens entered, want 2", len(tokens))
	}
	for _, tok := range tokens {
		if tok.BlockingTarget != attackerA {
			t.Errorf("a token is blocking %s, want %s", tok.BlockingTarget, attackerA)
		}
		if n := len(blockDeclEvents(g, EventBlock, tok.InstanceID)); n != 0 {
			t.Errorf("a token was declared as a blocker (CR 509.4): %d block events", n)
		}
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, attackerA)); n != 1 {
		t.Errorf("the attacker became blocked %d times, want once (CR 509.3c)", n)
	}
}

// An Aura token created attached to a named host (a copy of an Aura
// spell, #3158), doubled: both tokens enter attached to it.
func TestSeveralAuraTokensCreatedAttachedAllAttach(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	host := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	before := tctBattlefieldIDs(g)

	g.WithWriteLock(func() {
		if err := g.CreateTokensKeyedForEffect(TokenCreation{
			Controller: me,
			Groups: []TokenGroup{{
				Template: Card{Name: "Test Aura", TypeLine: "Token Enchantment — Aura"},
				Count:    2,
				Entry:    TokenEntryOptions{AttachedTo: TargetRef{Kind: TargetCard, ID: host}},
			}},
		}, KeyedTail{}); err != nil {
			t.Fatalf("CreateTokensKeyedForEffect: %v", err)
		}
	})
	tokens := tctNewTokens(g, before)
	if len(tokens) != 2 {
		t.Fatalf("%d tokens entered, want 2", len(tokens))
	}
	for _, tok := range tokens {
		if tok.AttachedTo.ID != host {
			t.Errorf("an Aura token is attached to %+v, want the host %s", tok.AttachedTo, host)
		}
	}
	if tokens[0].EntryOrdinal != tokens[1].EntryOrdinal {
		t.Error("the two tokens did not enter together")
	}
}
