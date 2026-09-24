package lobby

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_test.go — ADR 0048 addendum §15, decision (c): the
// auto-tap preview sees a spell's own cost modifiers with no code of
// its own, because its modifier pass is the shared ApplyCostModifiers.
// These tests hold that claim to the real catalog cards, so a preview
// that went back to pricing the printed cost fails here rather than as
// a disabled "Auto-tap & cast" button on a castable Blasphemous Act.

const (
	blasphemousActOracle = "7a2484a9-04fd-41a0-8224-610c1c07ed10"
	fireballOracle       = "aa7714b0-2bfb-458a-8ebf-37ec2c53383e"
	gildedLotusOracle    = "9a02a9a7-39d9-4763-85d3-747a0540b60b"
)

type previewFixture struct {
	t      *testing.T
	lobby  *Lobby
	gameID uuid.UUID
	alice  uuid.UUID
	bob    uuid.UUID
	get    func(path string) *http.Response
}

type previewBody struct {
	OK      bool     `json:"ok"`
	Plan    []string `json:"plan"`
	Missing []string `json:"missing"`
	Cost    string   `json:"cost"`
}

// newPreviewFixture starts a two-seat game and mints Alice a player
// session for it.
func newPreviewFixture(t *testing.T) *previewFixture {
	t.Helper()
	srv, l, a := newTestHTTPStack(t)
	meta, err := l.Create("Preview")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("join Alice: %v", err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatalf("join Bob: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, alice, "dummy", fillerDeck("A", 12)); err != nil {
		t.Fatalf("SetDeck Alice: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, bob, "dummy", fillerDeck("B", 12)); err != nil {
		t.Fatalf("SetDeck Bob: %v", err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Production start rolls for first player. This fixture's "alice" is
	// the casting seat, so bind that role to the roll winner rather than
	// back to join order; the endpoint behavior under test is unchanged.
	started, err := l.LookupGame(meta.ID)
	if err != nil {
		t.Fatalf("LookupGame: %v", err)
	}
	if started.Turn.ActiveSeat == 1 {
		alice, bob = bob, alice
	}
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RolePlayer, GameID: meta.ID, PlayerID: alice, Name: "Alice",
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	return &previewFixture{
		t: t, lobby: l, gameID: meta.ID, alice: alice, bob: bob,
		get: func(path string) *http.Response { return doGet(t, srv, path, tok) },
	}
}

func (f *previewFixture) spawn(owner uuid.UUID, zone game.ZoneKind, c game.Card, n int) []uuid.UUID {
	f.t.Helper()
	ids, err := f.lobby.SpawnCards(f.gameID, owner, zone, c, n)
	if err != nil {
		f.t.Fatalf("spawn %d %s into %s: %v", n, c.Name, zone, err)
	}
	return ids
}

func (f *previewFixture) mountains(n int) {
	f.t.Helper()
	f.spawn(f.alice, game.ZoneBattlefield, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}, n)
}

func (f *previewFixture) preview(card uuid.UUID, x int) previewBody {
	f.t.Helper()
	path := "/games/" + f.gameID.String() + "/auto-tap-preview?card=" + card.String()
	if x > 0 {
		path += "&x=" + strconv.Itoa(x)
	}
	resp := f.get(path)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		f.t.Fatalf("GET %s: status %d", path, resp.StatusCode)
	}
	var body previewBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		f.t.Fatalf("decode preview: %v", err)
	}
	return body
}

// Blasphemous Act ({8}{R}, "{1} less for each creature on the
// battlefield") in hand with five creatures out costs {3}{R}. Four
// Mountains pay that and the preview says so; three do not, and the
// preview reports the REDUCED price as what is missing, not the
// printed one.
func TestAutoTapPreviewAppliesSelfCostReduction(t *testing.T) {
	f := newPreviewFixture(t)
	act := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Blasphemous Act", TypeLine: "Sorcery", ManaCost: "{8}{R}", OracleID: blasphemousActOracle,
	}, 1)[0]
	f.spawn(f.bob, game.ZoneBattlefield, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"}, 3)
	f.spawn(f.alice, game.ZoneBattlefield, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"}, 2)
	f.mountains(3)

	short := f.preview(act, 0)
	if short.OK {
		t.Fatalf("three Mountains for {3}{R}: preview ok with plan %v, want a miss", short.Plan)
	}
	// "missing" is read against the floating pool, which is empty, so
	// it lists the whole price the preview charged: four symbols for
	// {3}{R}, where the printed {8}{R} would list nine.
	if len(short.Missing) != 4 {
		t.Errorf("three Mountains for {3}{R}: missing %v, want the four symbols of {3}{R}", short.Missing)
	}

	f.mountains(1)
	got := f.preview(act, 0)
	if !got.OK || len(got.Plan) != 4 {
		t.Errorf("four Mountains for {3}{R}: ok=%v plan=%v missing=%v, want ok with four taps", got.OK, got.Plan, got.Missing)
	}
}

// Fireball ({X}{R}, "{1} more for each target beyond the first") at
// X = 3 previews at {3}{R}: the preview has no targets, so it prices the
// one-target cast, which is what the X picker's readout says it shows.
func TestAutoTapPreviewPricesFireballAtOneTarget(t *testing.T) {
	f := newPreviewFixture(t)
	fireball := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Fireball", TypeLine: "Sorcery", ManaCost: "{X}{R}", OracleID: fireballOracle,
	}, 1)[0]
	f.mountains(4)

	got := f.preview(fireball, 3)
	if !got.OK || len(got.Plan) != 4 {
		t.Fatalf("Fireball at X=3 with four Mountains: ok=%v plan=%v missing=%v, want ok with four taps ({3}{R})", got.OK, got.Plan, got.Missing)
	}
	if got := f.preview(fireball, 4); got.OK {
		t.Errorf("Fireball at X=4 with four Mountains: preview ok with plan %v, want a miss ({4}{R})", got.Plan)
	}
}

// #779: the preview reports a cast only a "N mana of any one color"
// source can fund as payable. Before the planner learned the shape, a
// board of Gilded Lotus + two Islands answered `ok: false` with
// `missing: [U U]` for {3}{U}{U} — the cast modal greyed out its
// "Auto-tap & cast" button on a board that pays.
func TestAutoTapPreviewPlansAGildedLotus(t *testing.T) {
	f := newPreviewFixture(t)
	whale := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Big Blue Thing", TypeLine: "Creature — Whale", ManaCost: "{3}{U}{U}",
	}, 1)[0]
	f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Gilded Lotus", TypeLine: "Artifact", OracleID: gildedLotusOracle,
	}, 1)
	f.spawn(f.alice, game.ZoneBattlefield, game.Card{Name: "Island", TypeLine: "Basic Land — Island"}, 2)

	got := f.preview(whale, 0)
	if !got.OK || len(got.Plan) != 3 {
		t.Fatalf("Gilded Lotus + 2 Islands for {3}{U}{U}: ok=%v plan=%v missing=%v, want ok with three taps",
			got.OK, got.Plan, got.Missing)
	}
}

// …and still says no to a cast the one pick cannot make: two colours
// out of one Lotus.
func TestAutoTapPreviewRefusesTwoColoursFromOneLotus(t *testing.T) {
	f := newPreviewFixture(t)
	bird := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Azorius Thing", TypeLine: "Creature — Bird", ManaCost: "{W}{U}",
	}, 1)[0]
	f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Gilded Lotus", TypeLine: "Artifact", OracleID: gildedLotusOracle,
	}, 1)

	if got := f.preview(bird, 0); got.OK {
		t.Errorf("{W}{U} off one Gilded Lotus: preview ok with plan %v, want a miss", got.Plan)
	}
}
