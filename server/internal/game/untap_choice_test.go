package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// untap_choice_test.go — #826 / ADR 0070. CR 502.3's first sentence:
// the active player determines which permanents they control untap.

func withCatalogUntapCaps(t *testing.T, fn func(string) []UntapCap) {
	t.Helper()
	prev := CatalogUntapCaps
	CatalogUntapCaps = fn
	t.Cleanup(func() { CatalogUntapCaps = prev })
}

func withCatalogUntapOptOuts(t *testing.T, fn func(string) []UntapOptOut) {
	t.Helper()
	prev := CatalogUntapOptOuts
	CatalogUntapOptOuts = fn
	t.Cleanup(func() { CatalogUntapOptOuts = prev })
}

// capOverLands is a Winter Orb: "no more than N lands", live only
// while the source is untapped.
func capOverLands(key string, n int) func(string) []UntapCap {
	return func(k string) []UntapCap {
		if k != key {
			return nil
		}
		return []UntapCap{{
			Max:     n,
			Label:   "no more than N lands",
			Applies: func(_ *Game, source *Card, _ uuid.UUID) bool { return !source.Tapped },
			Counts:  func(_ *Game, _ *Card, target *Card) bool { return target.IsLand() },
		}}
	}
}

// optOutSelf is a Rust Tick: "you may choose not to untap this".
func optOutSelf(key string) func(string) []UntapOptOut {
	return func(k string) []UntapOptOut {
		if k != key {
			return nil
		}
		return []UntapOptOut{{
			Label: "you may choose not to untap this",
			Optional: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
		}}
	}
}

func openUntapChoice(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceUntapChoice {
			return c
		}
	}
	t.Fatal("no untap_choice prompt is open")
	return nil
}

// A board with no cap and no opt-out never asks. The commonest untap
// step in every game must not grow a click.
func TestUntapStepDoesNotAskWhenNothingIsInQuestion(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	a := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	b := pushTappedPermanent(g, seat, "Bear", "", "Creature", true)

	var paused bool
	g.WithWriteLock(func() { paused = g.performUntapStepLocked(0) })

	if paused {
		t.Fatal("the step paused with nothing to decide")
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a prompt was queued: %d", len(g.PendingChoices))
	}
	if cardByIDForUntapTest(g, a).Tapped || cardByIDForUntapTest(g, b).Tapped {
		t.Error("everything untapped")
	}
}

// A cap that does not BIND is not a decision. One tapped land under
// Winter Orb untaps with no prompt.
func TestUntapCapDoesNotAskWhenItDoesNotBind(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const orb = "winter-orb"
	land := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	pushTappedPermanent(g, seat, "Orb", orb, "Artifact", false)
	pushTappedPermanent(g, seat, "Bear", "", "Creature", true)
	withCatalogUntapCaps(t, capOverLands(orb, 1))

	var paused bool
	g.WithWriteLock(func() { paused = g.performUntapStepLocked(0) })

	if paused || len(g.PendingChoices) != 0 {
		t.Fatalf("one tapped land under a cap of one is not a decision (paused=%v, prompts=%d)",
			paused, len(g.PendingChoices))
	}
	if cardByIDForUntapTest(g, land).Tapped {
		t.Error("the one land untapped")
	}
}

// The Winter Orb shape: two tapped lands, a cap of one. The step
// pauses with NOTHING untapped (CR 502.3 untaps them simultaneously),
// offers only the lands, and the whole set untaps on the answer.
func TestUntapCapAsksAndUntapsTheWholeSetOnTheAnswer(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const orb = "winter-orb"
	land1 := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	land2 := pushTappedPermanent(g, seat, "Island", "", "Land", true)
	bear := pushTappedPermanent(g, seat, "Bear", "", "Creature", true)
	pushTappedPermanent(g, seat, "Orb", orb, "Artifact", false)
	withCatalogUntapCaps(t, capOverLands(orb, 1))

	var paused bool
	g.WithWriteLock(func() { paused = g.performUntapStepLocked(0) })
	if !paused {
		t.Fatal("the step did not pause")
	}
	prompt := openUntapChoice(t, g)
	if prompt.Chooser != seat {
		t.Errorf("chooser = %v, want the active player", prompt.Chooser)
	}
	if prompt.ChooseMin != 1 || prompt.ChooseMax != 1 {
		t.Errorf("bounds = %d..%d, want 1..1", prompt.ChooseMin, prompt.ChooseMax)
	}
	if len(prompt.ChooseCards) != 2 {
		t.Fatalf("candidates = %d, want the two lands only", len(prompt.ChooseCards))
	}
	for _, id := range prompt.ChooseCards {
		if id == bear {
			t.Error("a permanent no cap counts was offered")
		}
	}
	// Nothing untapped yet — the determination is not finished.
	if !cardByIDForUntapTest(g, bear).Tapped {
		t.Error("the step untapped a permanent before the determination was complete")
	}

	if err := g.ResolveUntapChoice(prompt.ID, seat, []uuid.UUID{land2}); err != nil {
		t.Fatalf("ResolveUntapChoice: %v", err)
	}
	if cardByIDForUntapTest(g, land2).Tapped {
		t.Error("the chosen land untapped")
	}
	if !cardByIDForUntapTest(g, land1).Tapped {
		t.Error("the land that was not chosen stays tapped")
	}
	if cardByIDForUntapTest(g, bear).Tapped {
		t.Error("the creature no cap counts untapped with the rest")
	}
}

// A cap whose source is TAPPED is not live: "as long as this artifact
// is untapped" is read at the instant the step asks.
func TestUntapCapReadsItsConditionAtTheStep(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const orb = "winter-orb"
	pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	pushTappedPermanent(g, seat, "Island", "", "Land", true)
	pushTappedPermanent(g, seat, "Orb", orb, "Artifact", true) // tapped
	withCatalogUntapCaps(t, capOverLands(orb, 1))

	var paused bool
	g.WithWriteLock(func() { paused = g.performUntapStepLocked(0) })
	if paused {
		t.Fatal("a tapped Winter Orb capped the step")
	}
}

// Caps COMPOSE. Winter Orb (lands ≤ 1) and Static Orb (permanents ≤ 2)
// over the same board: a legal set is two permanents of which at most
// one is a land, and the solver refuses everything else — including
// the short answer, because untapping is not optional.
func TestUntapCapsComposeAndRefuseIllegalSets(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const (
		winter = "winter-orb"
		static = "static-orb"
	)
	land1 := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	land2 := pushTappedPermanent(g, seat, "Island", "", "Land", true)
	bear := pushTappedPermanent(g, seat, "Bear", "", "Creature", true)
	ox := pushTappedPermanent(g, seat, "Ox", "", "Creature", true)
	pushTappedPermanent(g, seat, "Winter", winter, "Artifact", false)
	pushTappedPermanent(g, seat, "Static", static, "Artifact", false)
	withCatalogUntapCaps(t, func(k string) []UntapCap {
		if c := capOverLands(winter, 1)(k); c != nil {
			return c
		}
		if k != static {
			return nil
		}
		return []UntapCap{{
			Max:     2,
			Label:   "no more than two permanents",
			Applies: func(_ *Game, source *Card, _ uuid.UUID) bool { return !source.Tapped },
		}}
	})

	g.WithWriteLock(func() { g.performUntapStepLocked(0) })
	prompt := openUntapChoice(t, g)
	if prompt.ChooseMin != 1 || prompt.ChooseMax != 2 {
		t.Errorf("bounds = %d..%d, want 1..2", prompt.ChooseMin, prompt.ChooseMax)
	}
	if len(prompt.ChooseCards) != 4 {
		t.Fatalf("candidates = %d, want the four tapped permanents", len(prompt.ChooseCards))
	}

	for _, tc := range []struct {
		name  string
		picks []uuid.UUID
		want  bool
	}{
		{"one land and one creature", []uuid.UUID{land1, bear}, true},
		{"two creatures", []uuid.UUID{bear, ox}, true},
		{"two lands breaks Winter Orb", []uuid.UUID{land1, land2}, false},
		{"one permanent wastes a mandatory untap", []uuid.UUID{bear}, false},
	} {
		got := false
		g.ReadSnapshot(func() { got = g.ChooseCardsPickLegalLocked(prompt, tc.picks) })
		if got != tc.want {
			t.Errorf("%s: legal = %v, want %v", tc.name, got, tc.want)
		}
	}

	if err := g.ResolveUntapChoice(prompt.ID, seat, []uuid.UUID{land1, land2}); !errors.Is(err, ErrChoiceSetRejected) {
		t.Fatalf("an illegal set was accepted: %v", err)
	}
	if err := g.ResolveUntapChoice(prompt.ID, seat, []uuid.UUID{land1, bear}); err != nil {
		t.Fatalf("a legal set was refused: %v", err)
	}
	if cardByIDForUntapTest(g, land1).Tapped || cardByIDForUntapTest(g, bear).Tapped {
		t.Error("the chosen permanents untapped")
	}
	if !cardByIDForUntapTest(g, land2).Tapped || !cardByIDForUntapTest(g, ox).Tapped {
		t.Error("the rest stay tapped")
	}
}

// "You may choose not to untap this" — the floor is zero, and the
// empty answer is a real one.
func TestUntapOptOutCanDeclineAndCanAccept(t *testing.T) {
	const tick = "rust-tick"
	for _, tc := range []struct {
		name        string
		pick        bool
		wantsTapped bool
	}{
		{"declined", false, true},
		{"accepted", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			seat := g.Seats[0].ID
			id := pushTappedPermanent(g, seat, "Tick", tick, "Artifact Creature", true)
			other := pushTappedPermanent(g, seat, "Bear", "", "Creature", true)
			withCatalogUntapOptOuts(t, optOutSelf(tick))

			g.WithWriteLock(func() { g.performUntapStepLocked(0) })
			prompt := openUntapChoice(t, g)
			if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
				t.Fatalf("bounds = %d..%d, want 0..1", prompt.ChooseMin, prompt.ChooseMax)
			}
			if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != id {
				t.Fatalf("candidates = %v, want only the opt-out permanent", prompt.ChooseCards)
			}
			var picks []uuid.UUID
			if tc.pick {
				picks = []uuid.UUID{id}
			}
			if err := g.ResolveUntapChoice(prompt.ID, seat, picks); err != nil {
				t.Fatalf("ResolveUntapChoice: %v", err)
			}
			if cardByIDForUntapTest(g, id).Tapped != tc.wantsTapped {
				t.Errorf("tapped = %v, want %v", cardByIDForUntapTest(g, id).Tapped, tc.wantsTapped)
			}
			if cardByIDForUntapTest(g, other).Tapped {
				t.Error("the permanent that was never in question untapped either way")
			}
		})
	}
}

// A cap and an opt-out in one prompt: the floor counts the mandatory
// half only, and the opt-out permanent is exempt from maximality.
func TestUntapCapAndOptOutShareOnePrompt(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const (
		orb  = "winter-orb"
		tick = "rust-tick"
	)
	land1 := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	pushTappedPermanent(g, seat, "Island", "", "Land", true)
	tickID := pushTappedPermanent(g, seat, "Tick", tick, "Artifact Creature", true)
	pushTappedPermanent(g, seat, "Orb", orb, "Artifact", false)
	withCatalogUntapCaps(t, capOverLands(orb, 1))
	withCatalogUntapOptOuts(t, optOutSelf(tick))

	g.WithWriteLock(func() { g.performUntapStepLocked(0) })
	prompt := openUntapChoice(t, g)
	if prompt.ChooseMin != 1 || prompt.ChooseMax != 2 {
		t.Errorf("bounds = %d..%d, want 1..2", prompt.ChooseMin, prompt.ChooseMax)
	}
	if len(prompt.ChooseCards) != 3 {
		t.Fatalf("candidates = %d, want two lands and the Tick", len(prompt.ChooseCards))
	}
	for _, tc := range []struct {
		name  string
		picks []uuid.UUID
		want  bool
	}{
		{"one land, Tick left tapped", []uuid.UUID{land1}, true},
		{"one land and the Tick", []uuid.UUID{land1, tickID}, true},
		{"only the Tick wastes the land untap", []uuid.UUID{tickID}, false},
	} {
		got := false
		g.ReadSnapshot(func() { got = g.ChooseCardsPickLegalLocked(prompt, tc.picks) })
		if got != tc.want {
			t.Errorf("%s: legal = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// ADR 0058 still has the first word. A restricted permanent and a
// marked one are not offered and do not untap; a chosen permanent with
// a stun counter loses the counter instead (CR 122.1d).
func TestUntapChoiceNeverOffersWhatCannotUntap(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const (
		orb   = "winter-orb"
		vault = "mana-vault"
	)
	free := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	stunned := pushTappedPermanent(g, seat, "Island", "", "Land", true)
	marked := pushTappedPermanent(g, seat, "Swamp", "", "Land", true)
	restricted := pushTappedPermanent(g, seat, "Vault", vault, "Artifact Land", true)
	pushTappedPermanent(g, seat, "Orb", orb, "Artifact", false)
	withCatalogUntapCaps(t, capOverLands(orb, 1))
	withCatalogUntapStepRestrictions(t, func(k string) []UntapStepRestriction {
		if k != vault {
			return nil
		}
		return []UntapStepRestriction{{Restricts: func(target *Card, _ *Game, source *Card) bool {
			return target.InstanceID == source.InstanceID
		}}}
	})
	g.WithWriteLock(func() {
		_ = g.SkipNextUntapForEffect(marked, uuid.Nil)
		cardByIDForUntapTest(g, stunned).Counters = map[string]int{CounterStun: 1}
	})

	g.WithWriteLock(func() { g.performUntapStepLocked(0) })
	prompt := openUntapChoice(t, g)
	for _, id := range prompt.ChooseCards {
		if id == marked || id == restricted {
			t.Error("a permanent that cannot untap was offered as one of the N choices")
		}
	}
	if len(prompt.ChooseCards) != 2 {
		t.Fatalf("candidates = %d, want the two lands that can untap", len(prompt.ChooseCards))
	}

	if err := g.ResolveUntapChoice(prompt.ID, seat, []uuid.UUID{stunned}); err != nil {
		t.Fatalf("ResolveUntapChoice: %v", err)
	}
	if c := cardByIDForUntapTest(g, stunned); !c.Tapped || c.Counters[CounterStun] != 0 {
		t.Errorf("a stunned permanent spends the counter and stays tapped: tapped=%v counters=%v",
			c.Tapped, c.Counters)
	}
	if !cardByIDForUntapTest(g, free).Tapped {
		t.Error("the land that was not chosen stays tapped")
	}
	if !cardByIDForUntapTest(g, restricted).Tapped || !cardByIDForUntapTest(g, marked).Tapped {
		t.Error("restricted and marked permanents stay tapped")
	}
}

// A cap is about the ACTIVE player's own determination. "During their
// untap steps" does not reach a Seedborn Muse untap on somebody else's
// turn (ADR 0070 Decision 4, ADR 0058 Decision 1).
func TestUntapCapDoesNotReachAnotherPlayersPermissionUntap(t *testing.T) {
	g := newActiveGame(t)
	active, other := g.Seats[0].ID, g.Seats[1].ID
	const (
		orb  = "winter-orb"
		muse = "seedborn"
	)
	theirs1 := pushTappedPermanent(g, other, "Their Forest", "", "Land", true)
	theirs2 := pushTappedPermanent(g, other, "Their Island", "", "Land", true)
	pushTappedPermanent(g, other, "Muse", muse, "Creature", false)
	pushTappedPermanent(g, active, "Orb", orb, "Artifact", false)
	withCatalogUntapCaps(t, capOverLands(orb, 1))
	withCatalogUntapStepPermissions(t, func(k string) []UntapStepPermission {
		if k != muse {
			return nil
		}
		return []UntapStepPermission{{
			AppliesTo: func(_ *Game, source *Card, activePlayer uuid.UUID) bool {
				return activePlayer != source.Controller
			},
			Untaps: func(_ *Game, source *Card, target *Card) bool {
				return target.Controller == source.Controller
			},
		}}
	})

	var paused bool
	g.WithWriteLock(func() { paused = g.performUntapStepLocked(0) })
	if paused {
		t.Fatal("the active player's cap asked about another player's permission untap")
	}
	if cardByIDForUntapTest(g, theirs1).Tapped || cardByIDForUntapTest(g, theirs2).Tapped {
		t.Error("both of the other player's lands untapped, uncapped")
	}
}

// The pause is a real pause: the step-entry hook stops at untap, the
// table cannot walk past the prompt, and the answer carries the cursor
// on to the upkeep in the same call.
func TestUntapChoiceStopsTheStepEntryAndTheAnswerResumesIt(t *testing.T) {
	g := newActiveGame(t)
	const orb = "winter-orb"
	seat1 := g.Seats[1].ID
	land1 := pushTappedPermanent(g, seat1, "Forest", "", "Land", true)
	pushTappedPermanent(g, seat1, "Island", "", "Land", true)
	pushTappedPermanent(g, seat1, "Orb", orb, "Artifact", false)
	withCatalogUntapCaps(t, capOverLands(orb, 1))

	advanceUntil(t, g, 30, func() bool { return len(g.PendingChoices) > 0 })
	prompt := openUntapChoice(t, g)
	if g.Turn.Step != StepUntap || g.Turn.ActiveSeat != 1 {
		t.Fatalf("cursor = seat %d %s, want seat 1 untap", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("advance_step past the prompt: %v, want ErrChoicePending", err)
	}

	if err := g.ResolveUntapChoice(prompt.ID, seat1, []uuid.UUID{land1}); err != nil {
		t.Fatalf("ResolveUntapChoice: %v", err)
	}
	if g.Turn.Step != StepUpkeep || g.Turn.ActiveSeat != 1 {
		t.Fatalf("after the answer the cursor is at seat %d %s, want seat 1 upkeep",
			g.Turn.ActiveSeat, g.Turn.Step)
	}
	if cardByIDForUntapTest(g, land1).Tapped {
		t.Error("the chosen land untapped")
	}
}

// The prompt holds the restore point, and it is counted as a
// continuation rather than silently dropped (ADR 0041).
func TestPausedUntapStepIsNotARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const orb = "winter-orb"
	pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	pushTappedPermanent(g, seat, "Island", "", "Land", true)
	pushTappedPermanent(g, seat, "Orb", orb, "Artifact", false)
	withCatalogUntapCaps(t, capOverLands(orb, 1))
	g.WithWriteLock(func() { g.performUntapStepLocked(0) })

	snap := g.CaptureSnapshot()
	if snap.Restorable() {
		t.Fatal("a paused untap step wrote a restore point")
	}
	if snap.Continuations.ChoiceResumeFrames == 0 {
		t.Error("the paused step's continuation is not counted in the census")
	}
}

// Only the chooser may answer, and only through the right verb.
func TestUntapChoiceRefusesTheWrongSeatAndTheWrongKind(t *testing.T) {
	g := newActiveGame(t)
	seat, other := g.Seats[0].ID, g.Seats[1].ID
	const orb = "winter-orb"
	land := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	pushTappedPermanent(g, seat, "Island", "", "Land", true)
	pushTappedPermanent(g, seat, "Orb", orb, "Artifact", false)
	withCatalogUntapCaps(t, capOverLands(orb, 1))
	g.WithWriteLock(func() { g.performUntapStepLocked(0) })
	prompt := openUntapChoice(t, g)

	if err := g.ResolveUntapChoice(prompt.ID, other, []uuid.UUID{land}); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("another seat answered: %v", err)
	}
	// The kind check is what stops a client answering an untap
	// determination through the chained-choice verb.
	if err := g.ResolveChooseCards(prompt.ID, seat, []uuid.UUID{land}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("the choose_cards verb answered an untap_choice: %v", err)
	}
}

// Undo restores a CLONE and the continuation has to resolve against
// that one (the StackItem.Effect contract every frame in the queue
// follows). The prompt's plan is captured data and its IDs are IDs, so
// a cloned game answers its own copy of the question.
func TestUntapChoiceSurvivesAClone(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const orb = "winter-orb"
	land1 := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	land2 := pushTappedPermanent(g, seat, "Island", "", "Land", true)
	pushTappedPermanent(g, seat, "Orb", orb, "Artifact", false)
	withCatalogUntapCaps(t, capOverLands(orb, 1))
	g.WithWriteLock(func() { g.performUntapStepLocked(0) })

	clone := g.Clone()
	prompt := openUntapChoice(t, clone)
	if err := clone.ResolveUntapChoice(prompt.ID, seat, []uuid.UUID{land1}); err != nil {
		t.Fatalf("answering the clone's prompt: %v", err)
	}
	if cardByIDForUntapTest(clone, land1).Tapped {
		t.Error("the clone's chosen land untapped")
	}
	if !cardByIDForUntapTest(clone, land2).Tapped {
		t.Error("the clone's other land stays tapped")
	}
	// The original is untouched, prompt included — which is what makes
	// "undo the answer, answer again" land where answering once would.
	if !cardByIDForUntapTest(g, land1).Tapped || len(g.PendingChoices) != 1 {
		t.Fatal("answering the clone changed the original")
	}
	if err := g.ResolveUntapChoice(prompt.ID, seat, []uuid.UUID{land2}); err != nil {
		t.Fatalf("answering the original after the clone: %v", err)
	}
	if cardByIDForUntapTest(g, land2).Tapped || !cardByIDForUntapTest(g, land1).Tapped {
		t.Error("the original answered independently of the clone")
	}
}

// CR 800.4a/h: the active player leaving during their own paused untap
// step takes the question and the permanents with them, and the turn
// ends. Dropping the prompt must not leave the cursor parked on a step
// nobody can finish (#902's table classifies this kind `false`).
func TestUntapChoiceDroppedWhenItsChooserLeaves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	const orb = "winter-orb"
	seat1 := g.Seats[1].ID
	pushTappedPermanent(g, seat1, "Forest", "", "Land", true)
	pushTappedPermanent(g, seat1, "Island", "", "Land", true)
	pushTappedPermanent(g, seat1, "Orb", orb, "Artifact", false)
	withCatalogUntapCaps(t, capOverLands(orb, 1))

	advanceUntil(t, g, 30, func() bool { return len(g.PendingChoices) > 0 })
	openUntapChoice(t, g)
	if g.Turn.Step != StepUntap || g.Turn.ActiveSeat != 1 {
		t.Fatalf("cursor = seat %d %s, want seat 1 untap", g.Turn.ActiveSeat, g.Turn.Step)
	}

	if err := g.Concede(seat1); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceUntapChoice {
			t.Fatal("the departed player's untap determination is still queued")
		}
	}
	if g.Turn.ActiveSeat == 1 {
		t.Fatalf("the turn did not move on: still seat 1 at %s", g.Turn.Step)
	}
	if g.Turn.Step == StepUntap {
		t.Fatalf("the cursor is parked on an untap step nobody can finish (seat %d)", g.Turn.ActiveSeat)
	}
}
