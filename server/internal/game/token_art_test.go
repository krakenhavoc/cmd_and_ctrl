package game

import (
	"testing"
)

// token_art_test.go — ADR 0078 (#1115): the engine half of "resolve a
// Scryfall token printing at runtime and stamp it". The matching rule
// itself lives in internal/cards/tokenart and is tested there against
// a hermetic pool; this file pins the two things that live in
// `game` — the hook is wired at the one creation path, and a
// pre-stamped ScryfallID (a token COPY) is never overwritten.

// withTokenArtResolver swaps TokenArtResolver for the duration of a
// test and restores whatever was there before (nil in every other
// game-package test, by ADR 0078 decision 6).
func withTokenArtResolver(t *testing.T, fn func(req TokenArtRequest) string) {
	t.Helper()
	prev := TokenArtResolver
	TokenArtResolver = fn
	t.Cleanup(func() { TokenArtResolver = prev })
}

func TestMintedTokenIsStampedFromTheResolver(t *testing.T) {
	const resolvedID = "11111111-1111-1111-1111-111111111111"
	calls := 0
	withTokenArtResolver(t, func(req TokenArtRequest) string {
		calls++
		if req.Template.Name != "Goblin" {
			t.Errorf("resolver saw template %q, want Goblin", req.Template.Name)
		}
		return resolvedID
	})

	g := newActiveGame(t)
	me := g.Seats[0]
	createGoblins(t, g, me.ID, 1, nil)

	if calls != 1 {
		t.Fatalf("resolver called %d times, want 1", calls)
	}
	tok, ok := findCardByName(g, "Goblin")
	if !ok {
		t.Fatal("no Goblin token on the battlefield")
	}
	if tok.ScryfallID != resolvedID {
		t.Errorf("ScryfallID = %q, want %q", tok.ScryfallID, resolvedID)
	}
	if !tok.TokenArtOnly {
		t.Error("TokenArtOnly = false, want true — this id names ART, not identity")
	}
}

func TestMintedTokenWithNoMatchKeepsTextFallback(t *testing.T) {
	withTokenArtResolver(t, func(TokenArtRequest) string { return "" })

	g := newActiveGame(t)
	me := g.Seats[0]
	createGoblins(t, g, me.ID, 1, nil)

	tok, ok := findCardByName(g, "Goblin")
	if !ok {
		t.Fatal("no Goblin token on the battlefield")
	}
	if tok.ScryfallID != "" {
		t.Errorf("ScryfallID = %q, want \"\" — a miss keeps today's text rendering", tok.ScryfallID)
	}
	if tok.TokenArtOnly {
		t.Error("TokenArtOnly = true on a miss; it must only be set alongside a real stamp")
	}
}

func TestNilResolverLeavesTokenArtEmpty(t *testing.T) {
	withTokenArtResolver(t, nil)

	g := newActiveGame(t)
	me := g.Seats[0]
	createGoblins(t, g, me.ID, 1, nil)

	tok, ok := findCardByName(g, "Goblin")
	if !ok {
		t.Fatal("no Goblin token on the battlefield")
	}
	if tok.ScryfallID != "" {
		t.Errorf("ScryfallID = %q, want \"\" with no resolver wired", tok.ScryfallID)
	}
}

// TestTokenCopyKeepsItsCopiedArtNotTheResolvers is Luke's decision for
// a token COPY (CreateTokenCopy / TokenCopyTemplate, ADR 0078
// "Interaction with ADR 0034"): the template already carries the
// copied card's real printing id, and mintTokenLocked must never run
// the generic resolver over it and replace it with a different
// printing.
func TestTokenCopyKeepsItsCopiedArtNotTheResolvers(t *testing.T) {
	const copiedArt = "22222222-2222-2222-2222-222222222222"
	calls := 0
	withTokenArtResolver(t, func(TokenArtRequest) string {
		calls++
		return "33333333-3333-3333-3333-333333333333" // a different, wrong id
	})

	g := newActiveGame(t)
	me := g.Seats[0]
	copyTemplate := Card{
		Name:       "Clone Target",
		TypeLine:   "Token Creature — Human",
		ScryfallID: copiedArt, // as TokenCopyTemplate stamps it (CR 707.2)
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokensThenForEffect(TokenCreation{
			Controller: me.ID,
			Groups:     []TokenGroup{{Template: copyTemplate, Count: 1}},
		}, nil); err != nil {
			t.Fatalf("CreateTokensThenForEffect: %v", err)
		}
	})

	if calls != 0 {
		t.Errorf("resolver called %d times, want 0 — a copy's ScryfallID must not be re-resolved", calls)
	}
	tok, ok := findCardByName(g, "Clone Target")
	if !ok {
		t.Fatal("no Clone Target token on the battlefield")
	}
	if tok.ScryfallID != copiedArt {
		t.Errorf("ScryfallID = %q, want the copied printing %q", tok.ScryfallID, copiedArt)
	}
	if tok.TokenArtOnly {
		t.Error("TokenArtOnly = true on a copy; its ScryfallID is a real copied identity, not a bare art stamp")
	}
}

// findCardByName is tokensNamed's cousin: the actual card, not just a
// count, so its stamped fields can be asserted on.
func findCardByName(g *Game, name string) (Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return Card{}, false
}

// TestToughnessIsKnownIgnoresAnArtOnlyStamp is the ADR 0078 amendment,
// pinned as a direct unit test rather than only through the whole-game
// tests above: a token whose ScryfallID is a bare art stamp
// (TokenArtOnly) must NOT read as "a printing stands behind this
// object" — a live 0/0 Construct token must not start dying to
// CR 704.5f just because it got a picture.
func TestToughnessIsKnownIgnoresAnArtOnlyStamp(t *testing.T) {
	artOnly := Card{
		Name: "Construct", TypeLine: "Token Artifact Creature — Construct",
		ScryfallID: "11111111-1111-1111-1111-111111111111", TokenArtOnly: true,
	}
	if artOnly.ToughnessIsKnown() {
		t.Error("ToughnessIsKnown = true on an art-only 0/0 token, want false")
	}

	// The same ScryfallID, WITHOUT TokenArtOnly (a token COPY's real
	// copied identity — TestTokenCopyOfAPrintedZeroZeroDiesOnArrival's
	// exact shape) still reads as known.
	copied := artOnly
	copied.TokenArtOnly = false
	if !copied.ToughnessIsKnown() {
		t.Error("ToughnessIsKnown = false on a copied 0/0 printing, want true")
	}
}

// TestFromScryfallPrintingIgnoresAnArtOnlyStamp is fromScryfallPrinting's
// half of the same amendment.
func TestFromScryfallPrintingIgnoresAnArtOnlyStamp(t *testing.T) {
	artOnly := Card{ScryfallID: "11111111-1111-1111-1111-111111111111", TokenArtOnly: true}
	if artOnly.fromScryfallPrinting() {
		t.Error("fromScryfallPrinting = true on an art-only stamp, want false")
	}

	copied := artOnly
	copied.TokenArtOnly = false
	if !copied.fromScryfallPrinting() {
		t.Error("fromScryfallPrinting = false on a copied printing, want true")
	}

	// The OracleID half is untouched by any of this.
	withOracle := Card{OracleID: "some-oracle-id"}
	if !withOracle.fromScryfallPrinting() {
		t.Error("fromScryfallPrinting = false with a non-empty OracleID, want true")
	}
}
