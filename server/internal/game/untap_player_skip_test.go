package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// skipPlayersNextUntap records a player-wide skip from no source object.
func skipPlayersNextUntap(t *testing.T, g *Game, player uuid.UUID, anyController bool, types ...string) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.SkipNextUntapStepForPlayerForEffect(player, uuid.Nil, types, anyController, "probe")
	})
	if !ok {
		t.Fatal("SkipNextUntapStepForPlayerForEffect recorded nothing")
	}
}

func playerUntapSkipRecords(g *Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == ModSkipNextUntapStep {
				n++
			}
		}
	}
	return n
}

// TestAPlayerWideSkipReachesALandPlayedLater is #3995's core (CR 611.2c):
// "lands target player controls don't untap during their next untap
// step" is a rule read at that step, so a land that player gets after
// the effect began is held too. It holds for that one step only.
func TestAPlayerWideSkipReachesALandPlayedLater(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	before := pushTappedPermanent(g, p.ID, "Old Land", "", "Land — Island", true)
	skipPlayersNextUntap(t, g, p.ID, false, "land")
	later := pushTappedPermanent(g, p.ID, "New Land", "", "Land — Island", true)
	rock := pushTappedPermanent(g, p.ID, "Rock", "", "Artifact", true)

	passTurnTo(t, g, 0)
	if !findCard(g, before).Tapped || !findCard(g, later).Tapped {
		t.Fatal("both lands sit out the next untap step, the later one included")
	}
	if findCard(g, rock).Tapped {
		t.Fatal("an artifact is not a land: it untaps")
	}
	if n := playerUntapSkipRecords(g); n != 0 {
		t.Fatalf("%d skip records left after the step, want 0 (used up)", n)
	}
	passTurnTo(t, g, 0)
	if findCard(g, before).Tapped || findCard(g, later).Tapped {
		t.Fatal("the following untap step untaps them")
	}
}

// TestAPlayerWideSkipWaitsForThatPlayersStep: another player's untap
// step neither holds their own permanents nor uses the skip up.
func TestAPlayerWideSkipWaitsForThatPlayersStep(t *testing.T) {
	g := newActiveGame(t)
	p, q := g.Seats[0], g.Seats[1]
	theirs := pushTappedPermanent(g, q.ID, "Their Bear", "", "Creature — Bear", true)
	mine := pushTappedPermanent(g, p.ID, "My Bear", "", "Creature — Bear", true)
	skipPlayersNextUntap(t, g, p.ID, false, "creature")

	passTurnTo(t, g, 1)
	if findCard(g, theirs).Tapped {
		t.Fatal("seat 1's creature untaps in seat 1's step")
	}
	if n := playerUntapSkipRecords(g); n != 1 {
		t.Fatalf("%d skip records after another player's step, want 1", n)
	}
	passTurnTo(t, g, 0)
	if !findCard(g, mine).Tapped {
		t.Fatal("seat 0's creature sits out seat 0's next untap step")
	}
}

// TestAPlayerWideSkipFollowsControlAtTheStep: "creatures target player
// controls" is read as the step begins. A creature the player gave away
// untaps in its new controller's step; one they took is held.
func TestAPlayerWideSkipFollowsControlAtTheStep(t *testing.T) {
	g := newActiveGame(t)
	p, q := g.Seats[0], g.Seats[1]
	given := pushTappedPermanent(g, p.ID, "Given", "", "Creature — Bear", true)
	taken := pushTappedPermanent(g, q.ID, "Taken", "", "Creature — Bear", true)
	skipPlayersNextUntap(t, g, p.ID, false, "creature")
	g.WithWriteLock(func() {
		cardByIDForUntapTest(g, given).Controller = q.ID
		cardByIDForUntapTest(g, taken).Controller = p.ID
	})
	passTurnTo(t, g, 1)
	if findCard(g, given).Tapped {
		t.Fatal("the creature seat 0 gave away untaps in its new controller's step")
	}
	passTurnTo(t, g, 0)
	if !findCard(g, taken).Tapped {
		t.Fatal("the creature seat 0 took is held in seat 0's step")
	}
}

// TestAnyControllerHoldsAPermissionUntap is Blinding Beam's bare
// "Creatures don't untap during target player's next untap step": a
// creature another player controls that a Seedborn-style permission
// would untap in that step stays tapped. The controlled form ("creatures
// target player controls") lets it untap.
func TestAnyControllerHoldsAPermissionUntap(t *testing.T) {
	const oracle = "test-seedborn-probe"
	withCatalogUntapStepPermissions(t, func(id string) []UntapStepPermission {
		if id != oracle {
			return nil
		}
		return []UntapStepPermission{{
			Label: "probe: untap all permanents you control during each other player's untap step",
			AppliesTo: func(_ *Game, source *Card, activePlayer uuid.UUID) bool {
				return activePlayer != source.Controller
			},
			Untaps: func(_ *Game, source, target *Card) bool {
				return target.Controller == source.Controller
			},
		}}
	})
	for _, tc := range []struct {
		name          string
		anyController bool
		wantTapped    bool
	}{
		{"creatures target player controls", false, false},
		{"creatures", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p, q := g.Seats[0], g.Seats[1]
			pushTappedPermanent(g, q.ID, "Muse", oracle, "Creature — Spirit", false)
			bear := pushTappedPermanent(g, q.ID, "Bear", "", "Creature — Bear", true)
			skipPlayersNextUntap(t, g, p.ID, tc.anyController, "creature")
			// Seat 1's own step comes first and untaps the bear; tap it
			// again so seat 0's step is the one that decides.
			passTurnTo(t, g, 1)
			g.WithWriteLock(func() { cardByIDForUntapTest(g, bear).Tapped = true })
			passTurnTo(t, g, 0)
			if got := findCard(g, bear).Tapped; got != tc.wantTapped {
				t.Fatalf("bear tapped = %v after seat 0's step, want %v", got, tc.wantTapped)
			}
		})
	}
}

// TestAPlayerWideSkipWaitsPastASkippedUntapStep: a skipped untap step
// (Stasis, CR 614.10a) never happens, so it uses nothing up; the skip
// holds at the next untap step that does.
func TestAPlayerWideSkipWaitsPastASkippedUntapStep(t *testing.T) {
	const oracle = "test-stasis-probe"
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		oracle: {{
			Watches: []EventKind{EventStepTransition},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventStepTransition && ev.StepTransitionStep == StepUntap
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Cancel()
				return nil
			},
			PureCancel: true,
			Label:      "probe: players skip their untap steps",
		}},
	})
	g := newActiveGame(t)
	p, q := g.Seats[0], g.Seats[1]
	stasis := pushTappedPermanent(g, q.ID, "Probe", oracle, "Enchantment", false)
	land := pushTappedPermanent(g, p.ID, "Land", "", "Land", true)
	skipPlayersNextUntap(t, g, p.ID, false, "land")

	passTurnTo(t, g, 0)
	if n := playerUntapSkipRecords(g); n != 1 {
		t.Fatalf("%d skip records after a skipped untap step, want 1", n)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stasis {
				g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
				break
			}
		}
	})
	passTurnTo(t, g, 0)
	if !findCard(g, land).Tapped {
		t.Fatal("the first untap step that happens holds the land")
	}
	if n := playerUntapSkipRecords(g); n != 0 {
		t.Fatalf("%d skip records left, want 0", n)
	}
}

// TestPlayerUntapSkipsCoveringNamesOnlyTheController: the badge and the
// bots read the skip on the permanents the named player controls.
func TestPlayerUntapSkipsCoveringNamesOnlyTheController(t *testing.T) {
	g := newActiveGame(t)
	p, q := g.Seats[0], g.Seats[1]
	mine := pushTappedPermanent(g, p.ID, "Mine", "", "Creature — Bear", true)
	theirs := pushTappedPermanent(g, q.ID, "Theirs", "", "Creature — Bear", true)
	land := pushTappedPermanent(g, p.ID, "Land", "", "Land", true)
	skipPlayersNextUntap(t, g, p.ID, true, "creature")
	skipPlayersNextUntap(t, g, p.ID, true, "creature")
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if got := g.PlayerUntapSkipsCoveringLocked(cardByIDForUntapTest(g, mine)); len(got) != 1 || got[0] != p.ID {
			t.Errorf("mine: %v, want [seat 0] once", got)
		}
		if got := g.PlayerUntapSkipsCoveringLocked(cardByIDForUntapTest(g, theirs)); len(got) != 0 {
			t.Errorf("theirs: %v, want none", got)
		}
		if got := g.PlayerUntapSkipsCoveringLocked(cardByIDForUntapTest(g, land)); len(got) != 0 {
			t.Errorf("land: %v, want none", got)
		}
	})
}

// TestAPlayerWideSkipForAPlayerWhoHasLeftWritesNothing (CR 800.4a).
func TestAPlayerWideSkipForAPlayerWhoHasLeftWritesNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Seats[2].Eliminated = true
	g.WithWriteLock(func() {
		if g.SkipNextUntapStepForPlayerForEffect(g.Seats[2].ID, uuid.Nil, []string{"land"}, false, "probe") {
			t.Error("a skip was recorded for a player who has left")
		}
		if g.SkipNextUntapStepForPlayerForEffect(g.Seats[1].ID, uuid.Nil, nil, false, "probe") {
			t.Error("a skip naming no type was recorded")
		}
	})
	if n := playerUntapSkipRecords(g); n != 0 {
		t.Fatalf("%d records, want 0", n)
	}
}

// TestPlayerUntapSkipsSurviveASnapshot: the record is plain data on the
// registry, so a restore point keeps it and it holds after the restore.
func TestPlayerUntapSkipsSurviveASnapshot(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	land := pushTappedPermanent(g, p.ID, "Land", "", "Land", true)
	skipPlayersNextUntap(t, g, p.ID, false, "land")
	_, restored := roundTrip(t, g)
	if n := playerUntapSkipRecords(restored); n != 1 {
		t.Fatalf("restored game has %d skip records, want 1", n)
	}
	passTurnTo(t, restored, 0)
	if !findCard(restored, land).Tapped {
		t.Fatal("the restored skip holds the land")
	}
}

// TestASnapshotNamingSkipNextUntapStepIsRefusedWithoutIt: an older
// binary refuses the file rather than untap what the card held.
func TestASnapshotNamingSkipNextUntapStepIsRefusedWithoutIt(t *testing.T) {
	g := newActiveGame(t)
	skipPlayersNextUntap(t, g, g.Seats[0].ID, false, "land")
	snap := g.CaptureSnapshot()

	spec, ok := modKinds[ModSkipNextUntapStep]
	if !ok {
		t.Fatal("skipNextUntapStep is not in the mod-kind table")
	}
	delete(modKinds, ModSkipNextUntapStep)
	defer func() { modKinds[ModSkipNextUntapStep] = spec }()

	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore without skipNextUntapStep = %v, want ErrUnknownEffectKey", err)
	}
}

// TestASkipNextUntapStepModThatNamesNothingIsRefused: registration
// panics, and restore refuses, a mod with no player or no permanent
// type, and anyController on another kind.
func TestASkipNextUntapStepModThatNamesNothingIsRefused(t *testing.T) {
	for name, m := range map[string]Mod{
		"no player":            {Kind: ModSkipNextUntapStep, Types: []string{"land"}},
		"no type":              {Kind: ModSkipNextUntapStep, Player: uuid.New()},
		"not a permanent type": {Kind: ModSkipNextUntapStep, Player: uuid.New(), Types: []string{"instant"}},
		"capitalised type":     {Kind: ModSkipNextUntapStep, Player: uuid.New(), Types: []string{"Land"}},
		"on another kind":      {Kind: ModSkipTurns, Player: uuid.New(), Amount: 1, AnyController: true},
	} {
		if skipNextUntapStepModProblem(m) == "" {
			t.Errorf("%s: accepted", name)
		}
	}
	if p := skipNextUntapStepModProblem(Mod{Kind: ModSkipNextUntapStep, Player: uuid.New(), Types: []string{"creature", "land"}, AnyController: true}); p != "" {
		t.Errorf("a sound mod refused: %s", p)
	}

	g := newActiveGame(t)
	skipPlayersNextUntap(t, g, g.Seats[0].ID, false, "land")
	snap := g.CaptureSnapshot()
	snap.ScopedEffects[0].Mods[0].Types = nil
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("a restore point naming no type = %v, want ErrUnknownEffectKey", err)
	}

	defer func() {
		if recover() == nil {
			t.Error("registering a skipNextUntapStep mod with no player did not panic")
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeGame, uuid.Nil,
			[]Mod{{Kind: ModSkipNextUntapStep, Types: []string{"land"}}}, IndefiniteDuration(), "probe")
	})
}
