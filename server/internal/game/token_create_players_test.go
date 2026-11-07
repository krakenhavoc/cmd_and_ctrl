package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// token_create_players_test.go — #4867 (ADR 0061 amendment 2026-11-06):
// "each player creates a token" opens one creation window per player and
// enters every player's tokens in one simultaneous entry (CR 101.4,
// CR 608.2f, CR 603.6a); a tagged group's tokens are told to the
// creation's keyed tail (TokenGroup.Tag, TokenTags).

func TestTokensCreatedByEachPlayerEnterTogether(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	before := tctBattlefieldIDs(g)
	soldier := Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}
	var specs []TokenCreation
	want := map[uuid.UUID]int{}
	for i, s := range g.Seats[1:] {
		specs = append(specs, TokenCreation{Controller: s.ID, Groups: []TokenGroup{{Template: soldier, Count: i + 1}}})
		want[s.ID] = i + 1
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokensForPlayersTogetherForEffect(specs); err != nil {
			t.Fatalf("CreateTokensForPlayersTogetherForEffect: %v", err)
		}
	})
	tokens := tctNewTokens(g, before)
	if len(tokens) != 6 {
		t.Fatalf("%d tokens entered, want 1+2+3", len(tokens))
	}
	got := map[uuid.UUID]int{}
	for _, tok := range tokens {
		got[tok.Controller]++
		if tok.Owner != tok.Controller {
			t.Errorf("a token is owned by %v and controlled by %v: a token's owner is the player who created it", tok.Owner, tok.Controller)
		}
		if tok.EntryOrdinal != tokens[0].EntryOrdinal {
			t.Error("the players' tokens did not enter in one event")
		}
	}
	for p, n := range want {
		if got[p] != n {
			t.Errorf("player %v has %d tokens, want %d", p, got[p], n)
		}
	}
	if got[g.Seats[0].ID] != 0 {
		t.Error("a player with no creation got a token")
	}
	if len(g.enteringTokens) != 0 {
		t.Errorf("%d tokens are still staged", len(g.enteringTokens))
	}
}

// The door makes no attacking or blocking tokens: no "each player creates"
// card has them, and a batch member has no field for what a token blocks.
func TestTokensCreatedByEachPlayerRefuseAttackingAndBlocking(t *testing.T) {
	g := newActiveGame(t)
	soldier := Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}
	for _, spec := range []TokenCreation{
		{Controller: g.Seats[0].ID, Attacking: g.Seats[1].ID, Groups: []TokenGroup{{Template: soldier, Count: 1}}},
		{Controller: g.Seats[0].ID, Groups: []TokenGroup{{Template: soldier, Count: 1, Blocking: uuid.New()}}},
		{Controller: g.Seats[0].ID, Groups: []TokenGroup{{Template: soldier, Count: 1, AttackOpponent: g.Seats[1].ID}}},
	} {
		var err error
		g.WithWriteLock(func() { err = g.CreateTokensForPlayersTogetherForEffect([]TokenCreation{spec}) })
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("err = %v, want ErrInvalidParam", err)
		}
	}
}

// withTokenTags appends one (token, tag) pair per tagged token after the
// caller's carry, and TokenTags reads them back past it. Untagged tokens and
// a zero tail are told nothing.
func TestTokenTagsRideTheTailAfterTheCallersCarry(t *testing.T) {
	own := uuid.New()
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	oppA, oppB := uuid.New(), uuid.New()
	batch := []stagedToken{
		{card: Card{InstanceID: a}, tag: oppA},
		{card: Card{InstanceID: b}},
		{card: Card{InstanceID: c}, tag: oppB},
	}
	tail := KeyedTail{Ref: TailRef{key: "test/tags"}, Carry: []uuid.UUID{own}}
	got := withTokenTags(tail, batch)
	if len(tail.Carry) != 1 {
		t.Fatal("withTokenTags wrote through the caller's Carry")
	}
	if len(got.Carry) != 5 || got.Carry[0] != own {
		t.Fatalf("carry = %v, want the caller's entry then two pairs", got.Carry)
	}
	tags := TokenTags(TailOutcome{Tail: got}, 1)
	if len(tags) != 2 || tags[a] != oppA || tags[c] != oppB {
		t.Fatalf("tags = %v", tags)
	}
	if _, ok := tags[b]; ok {
		t.Error("an untagged token was told a tag")
	}
	if z := withTokenTags(KeyedTail{}, batch); !z.IsZero() || len(z.Carry) != 0 {
		t.Error("a creation with no keyed tail was given one")
	}
	if u := withTokenTags(tail, batch[1:2]); len(u.Carry) != 1 {
		t.Error("a creation with no tagged group changed its tail")
	}
}
