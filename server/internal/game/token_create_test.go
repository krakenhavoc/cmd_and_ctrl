package game

import (
	"testing"

	"github.com/google/uuid"
)

// token_create_test.go — #762, the engine half. The catalog half (the
// real Parallel Lives, Anointed Procession, Academy Manufactor,
// Doubling Season and Mondrak) lives in
// cards/effects/token_creation_test.go.
//
// What is here is the engine's own notion of a creation: one event per
// INSTRUCTION, the entry every created token then takes, and what both
// halves do when they have to pause.

// goblinToken is the plain 1/1 every test below creates.
func goblinToken() Card {
	return Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1}
}

// clueToken is a second KIND, so a kind-rewriting replacement has
// something to rewrite to.
func clueToken() Card {
	return Card{Name: "Clue", TypeLine: "Token Artifact — Clue"}
}

// tokenDoubler is a Parallel-Lives-shaped replacement: every creation
// under its own controller makes twice as many.
func tokenDoubler(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventTokenCreated},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
			return ev.Kind == RepEventCreateTokens && src != nil && ev.TokenController == src.Controller
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.MultiplyTokens(2)
			return nil
		},
		Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}

// tokenKindSwapper is an Academy-Manufactor-shaped replacement: it
// does not change how many tokens are made, it changes which.
func tokenKindSwapper(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventTokenCreated},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
			return ev.Kind == RepEventCreateTokens && src != nil && ev.TokenController == src.Controller
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.ReplaceTokenKinds(clueToken())
			return nil
		},
		Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}

// entersTappedForOthers is a Kismet-shaped entry replacement: any
// permanent entering under somebody else's control arrives tapped.
func entersTappedForOthers(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, g *Game, src *Card) bool {
			if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || src == nil {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.Controller != src.Controller
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}

// entersWithACounter is a Renata-shaped entry replacement.
func entersWithACounter(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, g *Game, src *Card) bool {
			if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || src == nil {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.Controller == src.Controller && entering.IsCreature()
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.AddCounterAtETB("+1/+1", 1)
			return nil
		},
		Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}

// tokensNamed counts battlefield permanents by name.
func tokensNamed(g *Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// createGoblins runs one creation instruction for n Goblins under
// controller, with an optional continuation.
func createGoblins(t *testing.T, g *Game, controller uuid.UUID, n int, then func(*Game, []uuid.UUID) error) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokensThenForEffect(TokenCreation{
			Controller: controller,
			Groups:     []TokenGroup{{Template: goblinToken(), Count: n}},
		}, then); err != nil {
			t.Fatalf("CreateTokensThenForEffect: %v", err)
		}
	})
}

// --- the creation event itself ----------------------------------------

// TestOneDoublerDoublesTheWholeInstruction — CR 701.7b: "create three
// tokens" is ONE event, so a doubler makes six, in one batch, with no
// prompt.
func TestOneDoublerDoublesTheWholeInstruction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"parallel-lives": {tokenDoubler("double tokens")},
	})
	pushReplacementSource(g, "parallel-lives", me)

	var created []uuid.UUID
	createGoblins(t, g, me, 3, func(_ *Game, ids []uuid.UUID) error { created = ids; return nil })

	if len(g.PendingChoices) != 0 {
		t.Fatalf("one applicable replacement queued %d prompt(s)", len(g.PendingChoices))
	}
	if got := tokensNamed(g, "Goblin"); got != 6 {
		t.Errorf("Goblins = %d, want 6 (three doubled once)", got)
	}
	if len(created) != 6 {
		t.Errorf("the continuation was handed %d ids, want 6", len(created))
	}
	// One batch: every token was created by the same instruction, so
	// the events share an event batch (CR 603.2c).
	var batches = map[uint64]bool{}
	for _, ev := range g.Events {
		if ev.Kind == EventTokenCreated {
			batches[ev.Batch] = true
		}
	}
	if len(batches) != 1 {
		t.Errorf("token_created events spread over %d batches, want 1", len(batches))
	}
}

// TestTwoIdenticalDoublersQuadrupleWithNoPrompt — #792's identical
// window, on the creation event: two Anointed Processions are one
// declared effect on two objects, so there is nothing to order.
func TestTwoIdenticalDoublersQuadrupleWithNoPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"anointed-procession": {tokenDoubler("double tokens")},
	})
	pushReplacementSource(g, "anointed-procession", me)
	pushReplacementSource(g, "anointed-procession", me)

	createGoblins(t, g, me, 1, nil)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("two copies of one effect queued %d prompt(s), want none", len(g.PendingChoices))
	}
	if got := tokensNamed(g, "Goblin"); got != 4 {
		t.Errorf("Goblins = %d, want 4 (×2 then ×2)", got)
	}
}

// TestTwoDifferentCreationReplacementsPromptAndBothOrdersLand — a
// doubler and a kind-swapper are DIFFERENT declared effects, so CR 616
// gives the creation's controller the ordering, and the two orders
// really do produce different boards.
func TestTwoDifferentCreationReplacementsPromptAndBothOrdersLand(t *testing.T) {
	seed := func(t *testing.T) (*Game, uuid.UUID, uuid.UUID, uuid.UUID) {
		t.Helper()
		g := newActiveGame(t)
		me := g.Seats[0].ID
		stubCatalogReplacements(t, map[string][]ReplacementEffect{
			"doubler": {tokenDoubler("double tokens")},
			"swapper": {tokenKindSwapper("a Clue instead")},
		})
		dbl := pushReplacementSource(g, "doubler", me)
		swp := pushReplacementSource(g, "swapper", me)
		return g, me, dbl, swp
	}

	t.Run("double first, then swap the kind", func(t *testing.T) {
		g, me, dbl, swp := seed(t)
		createGoblins(t, g, me, 1, nil)
		prompt := onlyReplacementOrderPrompt(t, g, me)
		first := replacementEffectFromSource(t, g, prompt.ReplacementEffectIDs, dbl)
		second := replacementEffectFromSource(t, g, prompt.ReplacementEffectIDs, swp)
		if err := g.ResolveReplacementOrder(prompt.ID, me, []ReplacementEffectID{first, second}); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		if got := tokensNamed(g, "Clue"); got != 2 {
			t.Errorf("Clues = %d, want 2 (doubled to two Goblins, then both become Clues)", got)
		}
		if got := tokensNamed(g, "Goblin"); got != 0 {
			t.Errorf("Goblins = %d, want 0", got)
		}
	})

	t.Run("swap the kind first, then double", func(t *testing.T) {
		g, me, dbl, swp := seed(t)
		createGoblins(t, g, me, 1, nil)
		prompt := onlyReplacementOrderPrompt(t, g, me)
		first := replacementEffectFromSource(t, g, prompt.ReplacementEffectIDs, swp)
		second := replacementEffectFromSource(t, g, prompt.ReplacementEffectIDs, dbl)
		if err := g.ResolveReplacementOrder(prompt.ID, me, []ReplacementEffectID{first, second}); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		if got := tokensNamed(g, "Clue"); got != 2 {
			t.Errorf("Clues = %d, want 2 (one Clue, then doubled)", got)
		}
	})
}

// TestACancelledCreationMakesNothingAndStillRunsItsThen — CR 614.10
// with a null replacement. The caller's continuation is a terminal
// outcome, not a promise of tokens.
func TestACancelledCreationMakesNothingAndStillRunsItsThen(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"no-tokens": {{
			Watches: []EventKind{EventTokenCreated},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCreateTokens
			},
			Replace:    func(ev *ReplacementEvent, _ *Game, _ *Card) error { ev.Cancel(); return nil },
			Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
			Label:      "no tokens",
		}},
	})
	pushReplacementSource(g, "no-tokens", me)

	ran := 0
	var handed []uuid.UUID
	createGoblins(t, g, me, 2, func(_ *Game, ids []uuid.UUID) error { ran++; handed = ids; return nil })

	if got := tokensNamed(g, "Goblin"); got != 0 {
		t.Errorf("Goblins = %d, want 0 (the creation was cancelled)", got)
	}
	if ran != 1 || len(handed) != 0 {
		t.Errorf(`"then" ran %d time(s) with %d ids, want 1 and 0`, ran, len(handed))
	}
}

// --- the entry every created token then takes -------------------------

// TestACreatedTokenTakesTheEntryPipeline — the second half of #762.
// An enters-tapped replacement and an enters-with-counters replacement
// both reach a token, because a token now enters the way every other
// permanent does.
func TestACreatedTokenTakesTheEntryPipeline(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"kismet": {entersTappedForOthers("enters tapped")},
		"renata": {entersWithACounter("enters with a counter")},
	})
	pushReplacementSource(g, "kismet", me)
	pushReplacementSource(g, "renata", opp)

	createGoblins(t, g, opp, 1, nil)

	// Two DIFFERENT entry replacements apply to the same entering
	// token, so CR 616 asks the token's controller for the order —
	// on the ENTRY, exactly as it would for a creature entering from a
	// hand. Answering it finishes the entry.
	prompt := onlyReplacementOrderPrompt(t, g, opp)
	if err := g.ResolveReplacementOrder(prompt.ID, opp, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}

	var tok *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Goblin" {
			tok = &g.Battlefield.Cards[i]
		}
	}
	if tok == nil {
		t.Fatal("no Goblin token was created")
	}
	if !tok.Tapped {
		t.Error("the token entered untapped — the enters-tapped replacement did not see it")
	}
	if got := tok.Counters["+1/+1"]; got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 — the enters-with-counters replacement did not see it", got)
	}
}

// TestACreationsOwnTappedClauseAndAReplacementBothLand — "create a
// TAPPED token" is the creation's clause and the replacement adds its
// own on top; they settle in one field and neither can lose the other.
func TestACreationsOwnTappedClauseSurvivesTheWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		if err := g.CreateTokensThenForEffect(TokenCreation{
			Controller: me,
			Groups:     []TokenGroup{{Template: goblinToken(), Count: 1, Entry: TokenEntryOptions{Tapped: true}}},
		}, nil); err != nil {
			t.Fatalf("CreateTokensThenForEffect: %v", err)
		}
	})
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && !c.Tapped {
			t.Error(`"create a tapped token" entered untapped`)
		}
	}
}

// TestTokensCreatedAttackingAreAttacking — CR 506.3c: put onto the
// battlefield attacking, never declared, so no EventAttack is
// announced and the token really is attacking.
func TestTokensCreatedAttackingAreAttacking(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.CreateTokensAttackingForEffect(me, goblinToken(), 2, opp); err != nil {
			t.Fatalf("CreateTokensAttackingForEffect: %v", err)
		}
	})
	attacking := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.AttackingTarget == opp {
			attacking++
		}
	}
	if attacking != 2 {
		t.Errorf("attacking Goblins = %d, want 2", attacking)
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventAttack {
			t.Error("a token PUT onto the battlefield attacking must not announce an attack declaration (CR 506.3c)")
		}
	}
}

// TestATokenEntryThatPausesFinishesAfterTheAnswer — two DIFFERENT
// enters-tapped effects on one entering token is a CR 616 ordering
// prompt on the ENTRY, not the creation. Nothing is on the battlefield
// while it is open, and answering finishes the whole batch.
func TestATokenEntryThatPausesFinishesAfterTheAnswer(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"kismet": {entersTappedForOthers("Kismet: enters tapped")},
		"thalia": {entersTappedForOthers("Thalia: enters tapped")},
	})
	pushReplacementSource(g, "kismet", me)
	pushReplacementSource(g, "thalia", me)

	ran := 0
	var handed []uuid.UUID
	createGoblins(t, g, opp, 2, func(_ *Game, ids []uuid.UUID) error { ran++; handed = ids; return nil })

	if got := tokensNamed(g, "Goblin"); got != 0 {
		t.Fatalf("Goblins on the battlefield = %d while the entry prompt is open, want 0", got)
	}
	if ran != 0 {
		t.Fatalf(`"then" ran while the entry prompt was open`)
	}
	prompt := onlyReplacementOrderPrompt(t, g, opp)
	if err := g.ResolveReplacementOrder(prompt.ID, opp, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	// The second token's entry asks the same question again.
	if len(g.PendingChoices) == 1 {
		second := onlyReplacementOrderPrompt(t, g, opp)
		if err := g.ResolveReplacementOrder(second.ID, opp, second.ReplacementEffectIDs); err != nil {
			t.Fatalf("ResolveReplacementOrder (second token): %v", err)
		}
	}
	if got := tokensNamed(g, "Goblin"); got != 2 {
		t.Errorf("Goblins = %d after both answers, want 2 — the rest of the batch was dropped", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && !c.Tapped {
			t.Error("a token whose entry paused on the ordering prompt entered untapped")
		}
	}
	if ran != 1 || len(handed) != 2 {
		t.Errorf(`"then" ran %d time(s) with %d ids, want 1 and 2`, ran, len(handed))
	}
}

// TestUndoAcrossATokenCreationPauseReplaysTheSameWay — the batch and
// its continuation live on an unserialisable frame, so the undo stack
// has to rewind a game sitting on one and have the replay come out the
// same. Two rewinds, because they fail differently: before the
// creation (the prompt and the tokens both go) and into the open
// prompt (answering again must make the same tokens, once).
func TestUndoAcrossATokenCreationPauseReplaysTheSameWay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"doubler": {tokenDoubler("double tokens")},
		"swapper": {tokenKindSwapper("a Clue instead")},
	})
	dbl := pushReplacementSource(g, "doubler", me)
	_ = pushReplacementSource(g, "swapper", me)

	ran := 0
	make1 := func() {
		t.Helper()
		ran = 0
		createGoblins(t, g, me, 1, func(_ *Game, _ []uuid.UUID) error { ran++; return nil })
	}

	beforeCreation := g.Clone()
	make1()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the CR 616 ordering prompt", len(g.PendingChoices))
	}
	g.WithWriteLock(func() { g.RestoreFrom(beforeCreation) })
	if len(g.PendingChoices) != 0 || tokensNamed(g, "Clue") != 0 || tokensNamed(g, "Goblin") != 0 {
		t.Fatal("the rewind to before the creation left a prompt or a token behind")
	}

	make1()
	promptOpen := g.Clone()
	answer := func() {
		t.Helper()
		prompt := onlyReplacementOrderPrompt(t, g, me)
		first := replacementEffectFromSource(t, g, prompt.ReplacementEffectIDs, dbl)
		rest := []ReplacementEffectID{first}
		for _, id := range prompt.ReplacementEffectIDs {
			if id != first {
				rest = append(rest, id)
			}
		}
		if err := g.ResolveReplacementOrder(prompt.ID, me, rest); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
	}
	answer()
	if got := tokensNamed(g, "Clue"); got != 2 || ran != 1 {
		t.Fatalf("first run: %d Clues and %d continuation runs, want 2 and 1", got, ran)
	}

	ran = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if tokensNamed(g, "Clue") != 0 {
		t.Fatal("the rewind into the open prompt left the tokens on the battlefield")
	}
	answer()
	if got := tokensNamed(g, "Clue"); got != 2 || ran != 1 {
		t.Errorf("replay: %d Clues and %d continuation runs, want 2 and 1", got, ran)
	}
}

// --- shared prompt helpers --------------------------------------------

// onlyReplacementOrderPrompt asserts the game is sitting on exactly one
// CR 616 ordering prompt addressed to `chooser`, and returns it.
func onlyReplacementOrderPrompt(t *testing.T, g *Game, chooser uuid.UUID) *PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want exactly one", len(g.PendingChoices))
	}
	pc := g.PendingChoices[0]
	if pc.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", pc.Kind, PendingChoiceReplacementOrder)
	}
	if pc.Chooser != chooser {
		t.Fatalf("chooser = %s, want %s", pc.Chooser, chooser)
	}
	return pc
}

// replacementEffectFromSource picks the prompt entry contributed by a
// particular permanent, without hard-coding the ID-minting scheme.
func replacementEffectFromSource(t *testing.T, g *Game, ids []ReplacementEffectID, cardID uuid.UUID) ReplacementEffectID {
	t.Helper()
	for _, id := range ids {
		if _, src := g.ReplacementOptionMetaForEffect(id); src == cardID {
			return id
		}
	}
	t.Fatalf("no prompt entry for source card %s", cardID)
	return 0
}
