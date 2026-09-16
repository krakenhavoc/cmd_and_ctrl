package game

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

// snapshot_test.go is the round-trip proof for game persistence.
//
// The central test is a PROPERTY, not a field checklist:
//
//	capture(restore(decode(encode(capture(g))))) == capture(g)
//
// i.e. a game, snapshotted, serialised to JSON, parsed back, restored
// into a fresh *Game, and snapshotted again, produces a snapshot
// identical to the first. Phrasing it that way means a field added to
// the snapshot is covered the moment it is added, with no test edit —
// and a field DROPPED between capture and restore fails immediately,
// because the second capture will not contain it.
//
// What the property deliberately cannot see is a field that exists on
// *Game but on neither side of the snapshot. That blind spot is what
// snapshot_drift_test.go closes.

// newRestorableGame returns a started 2-player game whose random
// source is both deterministic (fixed PCG seed) and persistable.
//
// The shared newActiveGame helper starts from a caller-built
// *rand.Rand, which by construction cannot have its stream position
// read back out, so every game it makes is flagged unpersistable.
// StartWithSource is the shape that gets both.
func newRestorableGame(t *testing.T) *Game {
	t.Helper()
	g := NewGame()
	for i := 0; i < 2; i++ {
		name := "P1"
		cmdr := "Commander 1"
		if i == 1 {
			name, cmdr = "P2", "Commander 2"
		}
		if _, err := g.AddPlayer(name, buildTestDeck(cmdr)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(rand.NewPCG(7, 11)); err != nil {
		t.Fatalf("StartWithSource: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// enrich drives the game into a state that touches as much of the
// snapshot surface as a continuation-free board can: permanents with
// counters / damage / tap state, a spell on the stack, a queued
// delayed trigger, exile with a play permission, a paused choice, a
// vote, promises, mana in pools, per-turn tallies and an event log.
//
// Everything here is deliberately closure-FREE. A board carrying live
// continuations is not a restore point (see ContinuationCensus), and
// asserting an exact round-trip on one would be asserting the wrong
// thing.
func enrich(t *testing.T, g *Game) {
	t.Helper()
	p0, p1 := g.Seats[0], g.Seats[1]

	g.WithWriteLock(func() {
		// --- battlefield: two permanents, one per seat ------------
		for i, p := range g.Seats {
			c := Card{
				InstanceID:           uuid.New(),
				Name:                 "Test Permanent",
				OracleID:             "oracle-permanent",
				TypeLine:             "Creature — Test",
				Power:                2,
				Toughness:            3,
				ManaCost:             "{1}{G}",
				Colors:               []string{"G"},
				ColorIdentity:        []string{"G"},
				ProducedMana:         []string{"G"},
				Keywords:             []string{"trample", "vigilance"},
				Owner:                p.ID,
				Controller:           p.ID,
				Tapped:               i == 0,
				BattleX:              float64(i) * 1.5,
				BattleY:              2.25,
				Counters:             map[string]int{"+1/+1": 2, "loyalty": 0},
				DamageMarked:         1,
				EnteredBattlefieldAt: int64(100 + i),
				SummonedThisTurn:     i == 1,
				LostLastCounter:      i == 1,
				VariableToughness:    i == 0,
				KnownBy:              map[uuid.UUID]bool{p0.ID: true, p1.ID: true},
			}
			g.Battlefield.PushTop(c)
		}

		// --- exile, carrying a play permission --------------------
		g.Exile.PushTop(Card{
			InstanceID: uuid.New(),
			Name:       "Exiled Card",
			OracleID:   "oracle-exiled",
			Owner:      p1.ID,
			Controller: p0.ID,
			ExilePlay:  ExilePlayPermission{Player: p0.ID, UntilTurn: 3},
			KnownBy:    map[uuid.UUID]bool{p0.ID: true},
		})

		// --- a spell on the stack, plus its StackMeta -------------
		// Kind spell with a nil Effect: spells dispatch through
		// EffectResolver by oracle ID at resolution, so a spell item
		// carries no closure and keeps the board a restore point.
		spellID := uuid.New()
		g.Stack.PushTop(Card{
			InstanceID: spellID,
			Name:       "Test Instant",
			OracleID:   "oracle-instant",
			TypeLine:   "Instant",
			Owner:      p0.ID,
			Controller: p0.ID,
		})
		g.StackMeta = map[uuid.UUID]*StackItem{
			spellID: {
				ID:           spellID,
				Kind:         StackItemSpell,
				Controller:   p0.ID,
				Owner:        p0.ID,
				SourceCardID: spellID,
				Label:        "Test Instant",
				Targets:      []TargetRef{{Kind: TargetPlayer, ID: p1.ID}},
				Modes:        []int{0, 2},
				XValue:       3,
				Distribution: map[uuid.UUID]int{p1.ID: 3},
				CastFromZone: ZoneHand,
				AltCost:      "flashback",
				Seq:          42,
			},
		}

		// --- a delayed trigger with no closure --------------------
		// Constructed directly rather than through
		// ScheduleDelayedTriggerForEffect, which (correctly) drops a
		// trigger with no Effect. The data half is what the snapshot
		// carries; the Effect half is exactly what phase 3 has to
		// make data-driven.
		g.DelayedTriggers = []*DelayedTrigger{{
			ID:                 uuid.New(),
			Controller:         p0.ID,
			SourceCardID:       spellID,
			Label:              "return the exiled creature",
			At:                 StepEnd,
			ControllerTurnOnly: true,
			CreatedTurn:        1,
			Cards:              []uuid.UUID{uuid.New()},
		}}

		// --- a paused choice carrying data but no continuation ----
		g.PendingChoices = []*PendingChoice{{
			ID:                uuid.New(),
			Kind:              PendingChoiceDiscardFromHand,
			Chooser:           p1.ID,
			FromPlayer:        p0.ID,
			Count:             2,
			Source:            spellID,
			Reason:            "discard two cards",
			ColorOptions:      []string{"W", "U"},
			PickTargetPlayers: []uuid.UUID{p0.ID},
			PickTargetCards:   []uuid.UUID{spellID},
			PickTargetMin:     1,
			PickTargetMax:     2,
			SacrificeOptions:  []uuid.UUID{spellID},
			ScryCards:         []uuid.UUID{spellID},
			TriggerOrderIDs:   []uuid.UUID{spellID},
			PayCost:           "{2}",
			SearchCards:       []uuid.UUID{spellID},
			SearchMax:         1,
			NoLegalTarget:     true,
			DamageAssignment: &DamageAssignmentFrame{
				AttackerID:       spellID,
				BlockerIDs:       []uuid.UUID{spellID},
				AttackerPower:    4,
				AllowTrample:     true,
				HasDeathtouch:    true,
				SourceController: p0.ID,
			},
		}}

		// --- politics --------------------------------------------
		g.Promises = map[PromiseKey]int{
			{From: p0.ID, To: p1.ID}: 2,
			{From: p1.ID, To: p0.ID}: 1,
		}
		g.Vote = &Vote{
			ID:        uuid.New(),
			Topic:     "who takes the initiative",
			Options:   []string{"P1", "P2"},
			Initiator: p0.ID,
			Ballots:   map[uuid.UUID]int{p0.ID: 0, p1.ID: 1},
		}
		g.Monarch = p0.ID
		g.Initiative = p1.ID

		// --- per-turn bookkeeping --------------------------------
		g.LoyaltyActivatedThisTurn = map[uuid.UUID]bool{spellID: true}
		g.SpellsCastThisTurn = map[uuid.UUID]CastTally{p0.ID: {Total: 3, Noncreature: 2}}
		g.LandsPlayedThisTurn = map[uuid.UUID]int{p0.ID: 1}
		g.DiscardPending = map[uuid.UUID]int{p1.ID: 2}
		g.SplitSecondActive = true
		g.StartingSeat = 1

		// --- CR 603.10 LKI ---------------------------------------
		g.lastKnownBattlefield = map[uuid.UUID]Characteristic{
			spellID: {
				Power: 2, Toughness: 2,
				Types:    []string{"Creature"},
				Subtypes: []string{"Test"},
				Colors:   []string{"G"},
				Name:     "Test Permanent",
			},
		}

		// --- player-scoped state ----------------------------------
		p0.Life = 34
		p0.Poison = 3
		p0.Energy = 2
		p0.Counters = map[string]int{"experience": 1}
		p0.MaxHandSize = 9
		p0.CommanderDamage = map[uuid.UUID]int{p1.ID: 7}
		p0.CommanderCasts = map[uuid.UUID]int{p0.ID: 2}
		p0.LifeHistory = []LifeChange{
			{Delta: -6, NewTotal: 34, At: time.Unix(1700000000, 0).UTC()},
		}
		p0.ManaPool = ManaPool{
			{Color: "G", Source: spellID, Restrictions: []string{"spend only on creature spells"}},
			{Color: "C"},
		}
		p1.Eliminated = false
		p1.LosesAtNextSBA = true
		p1.DisplayName = "Player Two"
		p1.DiscordID = "1234567890"
		p1.DiscordAvatarHash = "abcdef"

		// --- event log -------------------------------------------
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: p0.ID, Amount: 1})
		g.EmitEvent(Event{Kind: EventCast, Actor: p0.ID, Source: spellID, Label: "Test Instant"})
	})
}

// roundTrip runs a snapshot through JSON and back into a *Game.
func roundTrip(t *testing.T, g *Game) (*GameSnapshot, *Game) {
	t.Helper()
	snap := g.CaptureSnapshot()

	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}

	restored, err := decoded.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	return snap, restored
}

// TestSnapshotRoundTripIsExact is the headline property. A rich but
// continuation-free game, put through capture → JSON → restore →
// capture, must produce an identical snapshot.
func TestSnapshotRoundTripIsExact(t *testing.T) {
	g := newRestorableGame(t)
	enrich(t, g)

	before, restored := roundTrip(t, g)
	after := restored.CaptureSnapshot()

	// Two fields are expected to differ, both deliberately:
	//
	//  - TakenAt is wall-clock metadata about the capture, not game
	//    state.
	//  - LayerVersion is advanced by one on restore, exactly as
	//    RestoreFrom does after an undo, because every restored Card
	//    dropped its `effective` characteristic cache and the next
	//    read must recompute instead of serving a stale one. Assert
	//    the bump rather than ignoring it, then normalise.
	if after.LayerVersion != before.LayerVersion+1 {
		t.Errorf("restored layerVersion = %d, want %d (one past the captured value, to force a recompute)",
			after.LayerVersion, before.LayerVersion+1)
	}
	before.TakenAt, after.TakenAt = time.Time{}, time.Time{}
	after.LayerVersion = before.LayerVersion

	if !reflect.DeepEqual(before, after) {
		t.Errorf("snapshot did not survive the round trip.\nbefore: %s\nafter:  %s",
			mustJSON(t, before), mustJSON(t, after))
	}
}

// TestSnapshotRoundTripKeepsGameUsable proves the restored game is a
// working *Game and not just a matching blob: it has the listeners
// and built-in replacements a fresh NewGame installs (they are
// process-lifetime singletons the new BINARY owns, so they are
// rebuilt rather than deserialised), and it accepts a mutation.
func TestSnapshotRoundTripKeepsGameUsable(t *testing.T) {
	g := newRestorableGame(t)
	enrich(t, g)
	_, restored := roundTrip(t, g)

	if n := len(restored.Listeners); n != len(g.Listeners) {
		t.Errorf("restored game has %d listeners, want %d — they are rebuilt by NewGame, not serialised", n, len(g.Listeners))
	}
	if n := len(restored.BuiltinReplacements); n != 1 {
		t.Errorf("restored game has %d built-in replacements, want 1 (commander zone)", n)
	}
	// The layer engine must be stale so the first read recomputes
	// against the restored board rather than serving a cache that
	// was never carried over.
	if lv, rv := restored.layerVersion.Load(), restored.lastResolvedVersion.Load(); lv <= rv {
		t.Errorf("restored layerVersion=%d lastResolvedVersion=%d; want stale", lv, rv)
	}
	before := restored.LayerRecomputeCountForTest()
	restored.ReadSnapshot(func() {})
	if restored.LayerRecomputeCountForTest() == before {
		t.Error("first read of a restored game did not recompute layers")
	}
	// And it still mutates.
	if err := restored.DrawCard(restored.Seats[0].ID); err != nil {
		t.Errorf("restored game rejected DrawCard: %v", err)
	}
}

// TestSnapshotResumesTheRandomStream is the fairness assertion behind
// Game.rngState. A restored game must continue the stream it was in,
// not start a new one — otherwise a deploy silently re-deals every
// library from a different sequence.
func TestSnapshotResumesTheRandomStream(t *testing.T) {
	g := newRestorableGame(t)
	_, restored := roundTrip(t, g)

	// Draw the same number of values from each. Identical output
	// means the restored game is at the same position in the same
	// stream.
	var live, back [8]uint64
	g.WithWriteLock(func() {
		for i := range live {
			live[i] = g.rng.Uint64()
		}
	})
	restored.WithWriteLock(func() {
		for i := range back {
			back[i] = restored.rng.Uint64()
		}
	})
	if live != back {
		t.Errorf("restored game draws a different random stream:\nlive:     %v\nrestored: %v", live, back)
	}
}

// TestSnapshotRNGIsNotReproducibleAcrossGames guards the other half
// of fairness: two games started the production way must not share a
// stream. A constant or clock-derived seed would make library orders
// predictable from the outside.
func TestSnapshotRNGIsNotReproducibleAcrossGames(t *testing.T) {
	mk := func() [4]uint64 {
		g := NewGame()
		for i := 0; i < 2; i++ {
			if _, err := g.AddPlayer("P", buildTestDeck("C")); err != nil {
				t.Fatalf("AddPlayer: %v", err)
			}
		}
		if err := g.Start(nil); err != nil {
			t.Fatalf("Start: %v", err)
		}
		var out [4]uint64
		g.WithWriteLock(func() {
			for i := range out {
				out[i] = g.rng.Uint64()
			}
		})
		return out
	}
	if a, b := mk(), mk(); a == b {
		t.Errorf("two production-started games share a random stream: %v", a)
	}
}

// TestExternalRNGIsCensusedNotFaked covers the case the engine cannot
// persist: a caller-supplied *rand.Rand whose Source math/rand/v2
// keeps private. The snapshot must SAY it lost the stream rather than
// quietly reseeding and claiming a faithful restore.
func TestExternalRNGIsCensusedNotFaked(t *testing.T) {
	g := newActiveGame(t) // starts from rand.New(...) — external
	snap := g.CaptureSnapshot()

	if snap.RNG.Kind != rngKindExternal {
		t.Errorf("rng kind = %q, want %q", snap.RNG.Kind, rngKindExternal)
	}
	if !snap.Continuations.UnpersistableRNG {
		t.Error("census did not flag the unreadable random source")
	}
	if snap.Restorable() {
		t.Error("a game whose random stream was lost must not count as a restore point")
	}
	// It must still restore on the lenient path, with a real source.
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("lenient restore: %v", err)
	}
	if restored.rng == nil || restored.rngState == nil {
		t.Error("restored game must get a fresh persistable source, not a nil one")
	}
}

// TestCensusCountsEveryContinuationKind walks each closure-bearing
// slot and proves the census notices it. This is the test that keeps
// "we'll roll back to a clean boundary" honest: if a future change
// adds a continuation the census does not count, a snapshot holding
// it would wrongly pass as a restore point.
func TestCensusCountsEveryContinuationKind(t *testing.T) {
	noop := func(g *Game, item *StackItem) error { return nil }

	cases := []struct {
		name   string
		set    func(g *Game)
		expect func(c ContinuationCensus) int
	}{
		{
			name: "stack item effect",
			set: func(g *Game) {
				g.StackMeta = map[uuid.UUID]*StackItem{
					{}: {ID: uuid.New(), Kind: StackItemActivated, Label: "ability", Effect: noop},
				}
			},
			expect: func(c ContinuationCensus) int { return c.StackEffects },
		},
		{
			name: "pending trigger effect",
			set: func(g *Game) {
				g.PendingTriggers = []*StackItem{
					{ID: uuid.New(), Kind: StackItemTriggered, Label: "trigger", Effect: noop},
				}
			},
			expect: func(c ContinuationCensus) int { return c.StackEffects },
		},
		{
			name: "ability target spec",
			set: func(g *Game) {
				g.StackMeta = map[uuid.UUID]*StackItem{
					{}: {ID: uuid.New(), Kind: StackItemActivated, targetSpec: &TargetSpec{}},
				}
			},
			expect: func(c ContinuationCensus) int { return c.StackTargetSpecs },
		},
		{
			name: "delayed trigger effect",
			set: func(g *Game) {
				g.DelayedTriggers = []*DelayedTrigger{
					{ID: uuid.New(), Label: "at end step", At: StepEnd, Effect: noop},
				}
			},
			expect: func(c ContinuationCensus) int { return c.DelayedTriggerEffects },
		},
		{
			name: "pending choice resume frame",
			set: func(g *Game) {
				g.PendingChoices = []*PendingChoice{
					{ID: uuid.New(), Kind: PendingChoiceScry, scryResume: func(*Game) error { return nil }},
				}
			},
			expect: func(c ContinuationCensus) int { return c.ChoiceResumeFrames },
		},
		{
			name: "turn-scoped static",
			set: func(g *Game) {
				g.TurnScopedStatics = []ScopedStatic{{Label: "Giant Growth +3/+3"}}
			},
			expect: func(c ContinuationCensus) int { return c.TurnScopedStatics },
		},
		{
			name: "turn-scoped replacement",
			set: func(g *Game) {
				g.TurnScopedReplacements = []ReplacementEffect{{Label: "Fog"}}
			},
			expect: func(c ContinuationCensus) int { return c.TurnScopedReplacements },
		},
		{
			name: "intrinsic abilities with no oracle id",
			set: func(g *Game) {
				g.Battlefield.PushTop(Card{
					InstanceID:    uuid.New(),
					Name:          "Treasure",
					TypeLine:      "Token Artifact — Treasure",
					ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{C}"}},
				})
			},
			expect: func(c ContinuationCensus) int { return c.IntrinsicAbilityCards },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newRestorableGame(t)
			if !g.CaptureSnapshot().Restorable() {
				t.Fatal("setup: baseline game is already not a restore point")
			}
			g.WithWriteLock(func() { tc.set(g) })

			snap := g.CaptureSnapshot()
			if n := tc.expect(snap.Continuations); n != 1 {
				t.Errorf("census counted %d, want 1; census = %+v", n, snap.Continuations)
			}
			if snap.Restorable() {
				t.Error("snapshot with a live continuation must not be a restore point")
			}
			if len(snap.Continuations.Labels) == 0 {
				t.Error("census gave the operator no label to act on")
			}
			if _, err := snap.RestoreStrict(); !errors.Is(err, ErrSnapshotNotRestorable) {
				t.Errorf("RestoreStrict error = %v, want ErrSnapshotNotRestorable", err)
			}
			// The lenient path still works — a diagnostic tool wants
			// the game even when a production restore would refuse.
			if _, err := snap.Restore(); err != nil {
				t.Errorf("lenient Restore refused a censused snapshot: %v", err)
			}
		})
	}
}

// TestSpellTargetSpecIsRederived proves the half of the closure
// problem that DOES have an answer. A spell's target clause was
// looked up from the catalog by oracle ID when it was cast, so the
// new binary can look it up again — no census entry, and the restored
// item gets its spec back.
func TestSpellTargetSpecIsRederived(t *testing.T) {
	spec := &TargetSpec{Mode: "creature"}
	prev := CatalogTargetSpec
	CatalogTargetSpec = func(oracleID string) *TargetSpec {
		if oracleID == "oracle-bolt" {
			return spec
		}
		return nil
	}
	t.Cleanup(func() { CatalogTargetSpec = prev })

	g := newRestorableGame(t)
	spellID := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(Card{
			InstanceID: spellID,
			Name:       "Lightning Bolt",
			OracleID:   "oracle-bolt",
			TypeLine:   "Instant",
			Owner:      g.Seats[0].ID,
			Controller: g.Seats[0].ID,
		})
		g.StackMeta = map[uuid.UUID]*StackItem{spellID: {
			ID:           spellID,
			Kind:         StackItemSpell,
			Controller:   g.Seats[0].ID,
			SourceCardID: spellID,
			Label:        "Lightning Bolt",
			targetSpec:   spec,
		}}
	})

	snap := g.CaptureSnapshot()
	if snap.Continuations.StackTargetSpecs != 0 {
		t.Errorf("a spell's catalog-owned target spec was censused: %+v", snap.Continuations)
	}
	if !snap.Restorable() {
		t.Fatalf("spell on the stack should still be a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if got := restored.StackMeta[spellID].targetSpec; got == nil {
		t.Error("restore did not re-derive the spell's target spec from the catalog")
	} else if got.Mode != "creature" {
		t.Errorf("re-derived spec = %+v, want Mode=creature", got)
	}
}

// TestTokenCopyAbilitiesAreRederived covers CR 707.2: a token COPY
// carries the copied card's oracle ID, so its ability closures come
// back from the catalog and it does not block a restore — unlike a
// true token (Treasure, Food), which has no printing to look up.
func TestTokenCopyAbilitiesAreRederived(t *testing.T) {
	prev := CatalogActivatedAbilities
	CatalogActivatedAbilities = func(oracleID string) []ActivatedAbilityShape {
		if oracleID != "oracle-copied" {
			return nil
		}
		return []ActivatedAbilityShape{{
			Label:  "{T}: draw a card",
			Effect: func(*Game, *StackItem) error { return nil },
		}}
	}
	t.Cleanup(func() { CatalogActivatedAbilities = prev })

	g := newRestorableGame(t)
	tokenID := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{
			InstanceID: tokenID,
			Name:       "Copied Creature",
			OracleID:   "oracle-copied",
			TypeLine:   "Token Creature — Test",
			Owner:      g.Seats[0].ID,
			Controller: g.Seats[0].ID,
			ActivatedAbilities: []ActivatedAbilityShape{{
				Label:  "{T}: draw a card",
				Effect: func(*Game, *StackItem) error { return nil },
			}},
		})
	})

	snap := g.CaptureSnapshot()
	if snap.Continuations.IntrinsicAbilityCards != 0 {
		t.Errorf("token copy with an oracle id was censused: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	c := findCardInZone(restored.Battlefield, tokenID)
	if c == nil {
		t.Fatal("restored battlefield lost the token copy")
	}
	if len(c.ActivatedAbilities) != 1 {
		t.Fatalf("token copy has %d activated abilities, want 1", len(c.ActivatedAbilities))
	}
	if c.ActivatedAbilities[0].Effect == nil {
		t.Error("re-derived ability has no effect closure")
	}
}

// TestSchemaSkewIsRefusedNotGuessed is the version-skew policy in
// test form. A file from a newer server, or one older than any
// migration we still carry, is refused — never partially applied.
func TestSchemaSkewIsRefusedNotGuessed(t *testing.T) {
	g := newRestorableGame(t)
	snap := g.CaptureSnapshot()

	t.Run("too new", func(t *testing.T) {
		s := *snap
		s.Schema = SnapshotSchemaVersion + 1
		if _, err := s.Restore(); !errors.Is(err, ErrSchemaTooNew) {
			t.Errorf("err = %v, want ErrSchemaTooNew", err)
		}
		if _, err := s.RestoreStrict(); !errors.Is(err, ErrSchemaTooNew) {
			t.Errorf("strict err = %v, want ErrSchemaTooNew", err)
		}
	})
	t.Run("too old", func(t *testing.T) {
		s := *snap
		s.Schema = minRestorableSchema - 1
		if _, err := s.Restore(); !errors.Is(err, ErrSchemaUnsupported) {
			t.Errorf("err = %v, want ErrSchemaUnsupported", err)
		}
	})
	t.Run("current", func(t *testing.T) {
		if snap.Schema != SnapshotSchemaVersion {
			t.Fatalf("capture stamped schema %d, want %d", snap.Schema, SnapshotSchemaVersion)
		}
		if _, err := snap.Restore(); err != nil {
			t.Errorf("current schema refused: %v", err)
		}
	})
}

// TestSnapshotIsByteStable proves two captures of one unchanged game
// serialise identically. Map iteration order is unspecified in Go, so
// without the sorts in captureSnapshotLocked a restore point would
// churn on disk and be undiffable.
func TestSnapshotIsByteStable(t *testing.T) {
	g := newRestorableGame(t)
	enrich(t, g)

	first := mustJSON(t, zeroTakenAt(g.CaptureSnapshot()))
	for i := 0; i < 20; i++ {
		if got := mustJSON(t, zeroTakenAt(g.CaptureSnapshot())); got != first {
			t.Fatalf("capture %d differs from the first:\n%s\n%s", i, first, got)
		}
	}
}

// TestSnapshotDoesNotAliasTheLiveGame is the clone.go lesson applied
// to the snapshot: mutating the game after a capture must not reach
// the captured copy.
func TestSnapshotDoesNotAliasTheLiveGame(t *testing.T) {
	g := newRestorableGame(t)
	enrich(t, g)
	snap := g.CaptureSnapshot()

	before := mustJSON(t, zeroTakenAt(snap))

	g.WithWriteLock(func() {
		g.Seats[0].Life = 1
		g.Seats[0].ManaPool = nil
		g.Seats[0].Counters["experience"] = 99
		g.Battlefield.Cards[0].Counters["+1/+1"] = 99
		g.Battlefield.Cards[0].KnownBy = nil
		g.Promises = nil
		g.Vote = nil
		g.Events = nil
		g.lastKnownBattlefield = nil
		g.StackMeta = nil
	})

	if after := mustJSON(t, zeroTakenAt(snap)); after != before {
		t.Error("post-capture mutation leaked into the snapshot (aliased state)")
	}
}

func zeroTakenAt(s *GameSnapshot) *GameSnapshot {
	s.TakenAt = time.Time{}
	return s
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}
