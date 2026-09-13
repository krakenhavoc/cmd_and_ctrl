package aiseat_test

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// The tier table is the API's promise: all four names exist from the
// first PR so the picker and the wire never reshape, and exactly the
// buildable ones report available.
func TestTierCatalogDeclaresAllFourAndOffersOne(t *testing.T) {
	tiers := aiseat.Tiers()
	if len(tiers) != 4 {
		t.Fatalf("tiers: %+v", tiers)
	}
	want := []aiseat.Tier{aiseat.TierRandom, aiseat.TierHeuristic, aiseat.TierAssisted, aiseat.TierStrong}
	for i, ti := range tiers {
		if ti.Tier != want[i] {
			t.Errorf("tier %d = %q, want %q", i, ti.Tier, want[i])
		}
		if ti.Label == "" || ti.Description == "" {
			t.Errorf("tier %q has no picker copy: %+v", ti.Tier, ti)
		}
	}
	if avail := aiseat.AvailableTiers(); len(avail) != 1 || avail[0] != aiseat.TierRandom {
		t.Errorf("available: %v, want [random]", avail)
	}
}

// An unavailable tier must fail cleanly. Silently handing back a
// RandomPolicy for "strong" is the failure mode this guards.
func TestNewPolicyRefusesRatherThanDowngrades(t *testing.T) {
	p, err := aiseat.NewPolicy(aiseat.TierRandom)
	if err != nil || p == nil || p.Name() != "random" {
		t.Fatalf("random: %v %v", p, err)
	}
	for _, tier := range []aiseat.Tier{aiseat.TierHeuristic, aiseat.TierAssisted, aiseat.TierStrong} {
		got, err := aiseat.NewPolicy(tier)
		if !errors.Is(err, aiseat.ErrTierUnavailable) {
			t.Errorf("%s: err = %v, want ErrTierUnavailable", tier, err)
		}
		if got != nil {
			t.Errorf("%s: returned a policy %q — an unavailable tier must not fall back", tier, got.Name())
		}
	}
	if _, err := aiseat.NewPolicy("galaxy-brain"); !errors.Is(err, aiseat.ErrUnknownTier) {
		t.Errorf("unknown tier: err = %v, want ErrUnknownTier", err)
	}
	if _, ok := aiseat.LookupTier("galaxy-brain"); ok {
		t.Error("LookupTier accepted a name that is not a tier")
	}
}

// The placeholder deck source is the contract sub-PR 5 replaces: a
// named lookup returning decklist TEXT the existing upload path can
// consume.
func TestPlaceholderDeckSourceRoundTrips(t *testing.T) {
	src := aiseat.PlaceholderDecks()
	list := src.List()
	if len(list) == 0 {
		t.Fatal("placeholder source lists no decks")
	}
	for _, info := range list {
		if info.ID == "" || info.Name == "" {
			t.Errorf("deck has no identity: %+v", info)
		}
		got, text, ok := src.Decklist(info.ID)
		if !ok {
			t.Fatalf("Decklist(%q) not found although List() offered it", info.ID)
		}
		if got.ID != info.ID {
			t.Errorf("Decklist(%q) returned %+v", info.ID, got)
		}
		// deck.ParseText's shape, which is what makes this consumable
		// by the same pipeline a human's upload takes.
		if !strings.Contains(text, "Commander:") || !strings.Contains(text, "Mainboard:") {
			t.Errorf("deck %q is not a parseable decklist:\n%s", info.ID, text)
		}
	}
	if _, _, ok := src.Decklist("no-such-deck"); ok {
		t.Error("Decklist accepted an unknown ID")
	}
}

// A seat whose tier this build cannot play must not be quietly played
// at random — it is skipped, loudly. This is the restore case: a
// persisted game seated at a tier a rolled-back binary no longer has.
func TestStartBotsSkipsSeatsWithNoPolicy(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := lobby.NewLobby(mgr)
	host := aiseat.NewManagerWithConfig(nil, aiseat.Config{}, log)
	l.SetBotHost(host)

	meta, _ := l.Create("Rolled back")
	if _, _, err := l.AddBot(meta.ID, "Good", string(aiseat.TierRandom), "", "d", monoRedDeck(uuid.Nil)); err != nil {
		t.Fatal(err)
	}
	// AddBot does not police the tier — the HTTP layer does — so this
	// stands in for a seat persisted by a future binary.
	if _, _, err := l.AddBot(meta.ID, "Future", string(aiseat.TierStrong), "", "d", monoRedDeck(uuid.Nil)); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	defer host.Shutdown()
	if got := len(host.Runners(meta.ID)); got != 1 {
		t.Errorf("runners: %d, want 1 (the unavailable tier must be skipped, not downgraded)", got)
	}
}
