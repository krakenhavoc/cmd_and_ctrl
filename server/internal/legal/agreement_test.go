package legal_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// agreement_test.go is half of a cross-language contract test.
//
// S31 sub-PR 2 deletes `client/src/lib/timing.ts`'s reimplementation
// of the cast-timing rules and replaces it with a lookup over the
// `legal_moves` this package produces. Deleting a second
// implementation is only safe if you have first proved the two agree,
// and "agree" has to mean something stronger than "the new one
// matches itself".
//
// So the fixture below carries a THIRD statement of the truth: each
// scenario declares, by hand, what should be castable and why. This
// file asserts the Go enumerator matches that declaration; the
// generated testdata/timing_agreement.json carries the real filtered
// wire frames to client/src/lib/timingAgreement.test.ts, which
// asserts canCastFromHand matches the same declaration. Either side
// drifting fails a test, and neither side is the other's oracle.
//
// Regenerate with:
//
//	go test ./internal/legal/ -run TestTimingAgreement -update
var updateFixture = flag.Bool("update", false, "rewrite testdata/timing_agreement.json")

const fixturePath = "testdata/timing_agreement.json"

// expectation is one hand or command-zone card and the verdict both
// implementations must reach about it. Reason is documentation for
// whoever is staring at a failure — neither side asserts on it.
type expectation struct {
	InstanceID string `json:"instance_id"`
	Name       string `json:"name"`
	Castable   bool   `json:"castable"`
	Why        string `json:"why"`
}

// scenario is one game state, one viewer, the wire frame that viewer
// receives, and the declared verdicts.
type scenario struct {
	Name   string            `json:"name"`
	Note   string            `json:"note"`
	Viewer string            `json:"viewer"`
	View   protocol.GameView `json:"view"`
	Expect []expectation     `json:"expect"`
}

// --- fixture construction -----------------------------------------

// duel builds a two-seat game past its mulligan window with empty
// hands, ready for a scenario to stock. Two seats rather than four
// only because the fixture ships as JSON and a four-player frame is
// ~25 KB of board nobody is asserting on.
func duel(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := game.NewGame()
	for i := range 2 {
		// Ten filler is the smallest deck that survives Start's
		// seven-card opening hand plus a turn-1 draw, and it keeps
		// the committed fixture small: the viewer's OWN library is
		// unredacted, so every card left in it is wire weight in
		// every scenario.
		deck := make([]game.Card, 0, 11)
		deck = append(deck, game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil))
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat %d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(21, 23))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
		p.Hand.Cards = nil
	}
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%2]
	return g, active, other
}

func bolt() game.Card {
	return game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt}
}

func rites() game.Card {
	return game.Card{Name: "Village Rites", TypeLine: "Instant", ManaCost: "{B}", OracleID: oracleVillageRites}
}

func sorcery(name, cost string) game.Card {
	return game.Card{Name: name, TypeLine: "Sorcery", ManaCost: cost}
}

// lands drops n untapped basics of one type onto the battlefield.
func lands(g *game.Game, p *game.Player, name, sub string, n int) {
	for range n {
		battlefieldCard(g, p, basic(name, sub))
	}
}

// buildScenarios is the table. Each entry states its own truth in the
// Expect list; nothing here reads the enumerator to decide what it
// should say.
func buildScenarios(t *testing.T) []scenario {
	t.Helper()
	var out []scenario

	// 1. The sorcery-speed window OPEN: active seat, precombat main,
	//    empty stack. Everything the seat can pay for is playable.
	{
		g, active, _ := duel(t)
		lands(g, active, "Mountain", "Mountain", 2)
		lands(g, active, "Swamp", "Swamp", 1)
		battlefieldCard(g, active, creature("Sacrificial Bear", "{1}{G}", 2, 2))
		land := handCard(active, basic("Forest", "Forest"))
		bolts := handCard(active, bolt())
		sorc := handCard(active, sorcery("Divination", "{2}{U}"))
		body := handCard(active, creature("Grizzly Bears", "{1}{G}", 2, 2))
		advanceTo(t, g, game.StepPrecombatMain)
		out = append(out, scenario{
			Name: "sorcery_window_open",
			Note: "active seat, precombat main, empty stack, land drop unused",
			View: protocol.ViewOfGameFor(g, active.ID.String()), Viewer: active.ID.String(),
			Expect: []expectation{
				{land.String(), "Forest", true, "CR 305: main phase, empty stack, your turn, land drop unused"},
				{bolts.String(), "Lightning Bolt", true, "instant, and {R} is affordable off two Mountains"},
				{sorc.String(), "Divination", false, "sorcery timing is fine; {2}{U} is not payable with no blue source"},
				{body.String(), "Grizzly Bears", false, "{1}{G} is not payable with no green source"},
			},
		})
	}

	// 2. The same seat with the mana to pay. Separated from scenario 1
	//    so a failure tells you whether timing or affordability broke.
	{
		g, active, _ := duel(t)
		lands(g, active, "Forest", "Forest", 4)
		sorc := handCard(active, sorcery("Overrun", "{2}{G}"))
		body := handCard(active, creature("Grizzly Bears", "{1}{G}", 2, 2))
		advanceTo(t, g, game.StepPrecombatMain)
		out = append(out, scenario{
			Name: "sorcery_window_open_and_affordable",
			Note: "four Forests: the sorcery and the creature are both payable",
			View: protocol.ViewOfGameFor(g, active.ID.String()), Viewer: active.ID.String(),
			Expect: []expectation{
				{sorc.String(), "Overrun", true, "CR 307.1 window open and {2}{G} payable"},
				{body.String(), "Grizzly Bears", true, "same window, {1}{G} payable"},
			},
		})
	}

	// 3. The sorcery-speed window SHUT by the step. This is the whole
	//    point of the exercise: instants stay live in combat,
	//    everything else does not.
	{
		g, active, _ := duel(t)
		lands(g, active, "Mountain", "Mountain", 4)
		land := handCard(active, basic("Forest", "Forest"))
		bolts := handCard(active, bolt())
		sorc := handCard(active, sorcery("Act of Treason", "{2}{R}"))
		advanceTo(t, g, game.StepDeclareAttackers)
		out = append(out, scenario{
			Name: "combat_step_instant_only",
			Note: "declare attackers: CR 307.1 shuts, CR 305 shuts, instants stay open",
			View: protocol.ViewOfGameFor(g, active.ID.String()), Viewer: active.ID.String(),
			Expect: []expectation{
				{land.String(), "Forest", false, "CR 305: lands are a main-phase-only turn-based action"},
				{bolts.String(), "Lightning Bolt", true, "CR 304.1: an instant may be cast whenever you hold priority"},
				{sorc.String(), "Act of Treason", false, "CR 307.1: sorcery timing needs a main phase"},
			},
		})
	}

	// 4. The window shut by a non-empty stack, in the seat's own main
	//    phase. Same verdict, different reason — and the reason the
	//    old TypeScript checked stack emptiness separately.
	{
		g, active, other := duel(t)
		lands(g, active, "Mountain", "Mountain", 4)
		land := handCard(active, basic("Forest", "Forest"))
		bolts := handCard(active, bolt())
		sorc := handCard(active, sorcery("Act of Treason", "{2}{R}"))
		advanceTo(t, g, game.StepPrecombatMain)
		// Park something on the stack the cheap way: a filler card
		// from the opponent's hand pushed onto the shared stack zone.
		stackFiller := game.NewCard("Opposing Spell", uuid.Nil)
		stackFiller.InstanceID = uuid.New()
		stackFiller.Owner = other.ID
		stackFiller.Controller = other.ID
		g.Stack.PushTop(stackFiller)
		out = append(out, scenario{
			Name: "stack_not_empty",
			Note: "own main phase, but something is on the stack",
			View: protocol.ViewOfGameFor(g, active.ID.String()), Viewer: active.ID.String(),
			Expect: []expectation{
				{land.String(), "Forest", false, "CR 305: land drops need an empty stack"},
				{bolts.String(), "Lightning Bolt", true, "instants are exactly the response the window exists for"},
				{sorc.String(), "Act of Treason", false, "CR 307.1: sorcery timing needs an empty stack"},
			},
		})
	}

	// 5. No priority at all. Every verdict is false and none of them
	//    is about the card.
	{
		g, active, other := duel(t)
		lands(g, other, "Mountain", "Mountain", 4)
		land := handCard(other, basic("Forest", "Forest"))
		bolts := handCard(other, bolt())
		advanceTo(t, g, game.StepPrecombatMain)
		_ = active
		out = append(out, scenario{
			Name: "no_priority",
			Note: "the non-active seat during the active seat's main phase",
			View: protocol.ViewOfGameFor(g, other.ID.String()), Viewer: other.ID.String(),
			Expect: []expectation{
				{land.String(), "Forest", false, "not your turn and not your priority"},
				{bolts.String(), "Lightning Bolt", false, "CR 117.1: you cannot cast without priority"},
			},
		})
	}

	// 6. The land drop already spent. The OLD TypeScript had no idea
	//    — it checked the window and said yes — so this scenario is
	//    one of the two that only start passing after the port.
	{
		g, active, _ := duel(t)
		lands(g, active, "Mountain", "Mountain", 2)
		first := handCard(active, basic("Forest", "Forest"))
		advanceTo(t, g, game.StepPrecombatMain)
		if g.LandsPlayedThisTurn == nil {
			g.LandsPlayedThisTurn = map[uuid.UUID]int{}
		}
		g.LandsPlayedThisTurn[active.ID] = 1
		out = append(out, scenario{
			Name: "land_drop_spent",
			Note: "CR 305.2: one land per turn, and this seat has had theirs",
			View: protocol.ViewOfGameFor(g, active.ID.String()), Viewer: active.ID.String(),
			Expect: []expectation{
				{first.String(), "Forest", false, "CR 305.2: the land drop is spent"},
			},
		})
	}

	// 7. An additional cost with nothing to pay it with. This branch
	//    reads server-stamped fields on both sides and SURVIVES the
	//    port — it is here to prove the port did not break what it
	//    was supposed to keep.
	{
		g, active, _ := duel(t)
		lands(g, active, "Swamp", "Swamp", 2)
		emptyBoard := handCard(active, rites())
		advanceTo(t, g, game.StepPrecombatMain)
		out = append(out, scenario{
			Name: "additional_cost_unpayable",
			Note: "Village Rites with no creature to sacrifice (CR 601.2h)",
			View: protocol.ViewOfGameFor(g, active.ID.String()), Viewer: active.ID.String(),
			Expect: []expectation{
				{emptyBoard.String(), "Village Rites", false, "nothing on the board to sacrifice"},
			},
		})
	}

	// 8. The same card with the cost payable.
	{
		g, active, _ := duel(t)
		lands(g, active, "Swamp", "Swamp", 2)
		battlefieldCard(g, active, creature("Sacrificial Bear", "{1}{G}", 2, 2))
		payable := handCard(active, rites())
		advanceTo(t, g, game.StepPrecombatMain)
		out = append(out, scenario{
			Name: "additional_cost_payable",
			Note: "Village Rites with a creature on the board",
			View: protocol.ViewOfGameFor(g, active.ID.String()), Viewer: active.ID.String(),
			Expect: []expectation{
				{payable.String(), "Village Rites", true, "a creature is available to sacrifice"},
			},
		})
	}

	return out
}

// --- direction one: the Go enumerator ------------------------------

// TestTimingAgreement asserts the enumerator reaches every declared
// verdict, and writes the fixture the TypeScript half reads.
func TestTimingAgreement(t *testing.T) {
	scenarios := buildScenarios(t)

	for _, sc := range scenarios {
		t.Run(sc.Name, func(t *testing.T) {
			offered := map[string]bool{}
			for _, m := range sc.View.LegalMoves {
				if m.Kind == legal.KindCast || m.Kind == legal.KindLand {
					offered[m.Source.String()] = true
				}
			}
			for _, e := range sc.Expect {
				if got := offered[e.InstanceID]; got != e.Castable {
					t.Errorf("%s: enumerator says castable=%v, scenario declares %v (%s)",
						e.Name, got, e.Castable, e.Why)
				}
			}
		})
	}

	writeFixture(t, scenarios)
}

// uuidRE matches a canonical UUID anywhere in the fixture bytes.
var uuidRE = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// canonicaliseIDs rewrites every UUID in the fixture to a stable
// placeholder numbered by first appearance.
//
// Without this the fixture could never be diffed or staleness-checked:
// game.NewGame and every card instance mint a fresh random UUID, so
// two runs of an otherwise identical scenario produce byte-different
// JSON and the committed file would look stale on every CI run.
//
// Nothing downstream cares about the values. Both halves of the
// agreement test join on ID *equality* — expectation to hand card,
// move source to instance — and the placeholders preserve that
// exactly. They stay valid v4-shaped UUIDs so the Go side can still
// unmarshal the frame into protocol.GameView.
//
// The remaining ordering IS deterministic: the deck shuffle runs off
// a fixed PCG seed and every scenario builds its board in a fixed
// order, so first-appearance numbering is stable run to run.
func canonicaliseIDs(body []byte) []byte {
	seen := map[string]string{}
	return uuidRE.ReplaceAllFunc(body, func(m []byte) []byte {
		key := string(m)
		if repl, ok := seen[key]; ok {
			return []byte(repl)
		}
		if key == "00000000-0000-0000-0000-000000000000" {
			// The nil UUID is a real value (a move with no source
			// card); it must stay itself, not become id #7.
			seen[key] = key
			return m
		}
		repl := fmt.Sprintf("%08x-0000-4000-8000-000000000000", len(seen)+1)
		seen[key] = repl
		return []byte(repl)
	})
}

// writeFixture compares the generated fixture against the committed
// one, or rewrites it under -update. The comparison is on the JSON
// bytes (with IDs canonicalised): if a view field changes shape, the
// TypeScript half needs to see the new shape, and a silent skew
// between the two is the exact failure this whole file exists to
// prevent.
func writeFixture(t *testing.T, scenarios []scenario) {
	t.Helper()
	body, err := json.MarshalIndent(scenarios, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	body = canonicaliseIDs(body)
	body = append(body, '\n')

	if *updateFixture {
		if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(fixturePath, body, 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		t.Logf("wrote %s (%d scenarios, %d B)", fixturePath, len(scenarios), len(body))
		return
	}

	have, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read %s (regenerate with `go test ./internal/legal/ -run TestTimingAgreement -update`): %v", fixturePath, err)
	}
	if string(have) != string(body) {
		t.Errorf("%s is stale — the wire frames the TypeScript agreement test reads no longer match what this package produces.\n"+
			"Regenerate with: go test ./internal/legal/ -run TestTimingAgreement -update", fixturePath)
	}
}

// TestTimingAgreementFixtureIsSeatSafe re-checks, on the exact bytes
// the client test consumes, that no scenario's frame carries another
// seat's move list. The fixture is committed to the repo, so a leak
// here would be a leak in version control.
func TestTimingAgreementFixtureIsSeatSafe(t *testing.T) {
	body, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Skipf("no fixture yet: %v", err)
	}
	var scenarios []scenario
	if err := json.Unmarshal(body, &scenarios); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	if len(scenarios) == 0 {
		t.Fatal("fixture is empty")
	}
	for _, sc := range scenarios {
		for _, m := range sc.View.LegalMoves {
			if m.Player.String() != sc.Viewer {
				t.Errorf("%s: frame for %s carries a move belonging to %s", sc.Name, sc.Viewer, m.Player)
			}
		}
	}
}
