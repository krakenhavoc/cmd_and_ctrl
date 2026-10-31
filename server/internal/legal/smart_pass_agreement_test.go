package legal_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// smart_pass_agreement_test.go is the server half of ADR 0009's
// agreement fixture for smart auto-pass (the amendment of 2026-10-30,
// #4734), built the way agreement_test.go builds the timing one.
//
// The browser's smart auto-pass (client/src/lib/autopassDecision.ts) and
// the MCP seat's port of it at the default settings
// (internal/mcpseat/smart.go) read one server verdict per move,
// answers_now, but each keeps its own copy of the windows and the
// precedence. This fixture keeps the two from drifting. Each frame is a
// real filtered wire frame with its legal moves, the answers_now each
// cast, activation and mana move must carry, and the smart pass verdict
// at the default settings, "stop" or "pass", declared by hand. This file
// asserts the moves' answers_now; client/src/lib/smartPassAgreement.test.ts
// (slice 3) and mcpseat's smart_test.go (slice 4) assert the verdict.
// Either side drifting fails a test, and neither is the other's oracle.
// mcpseat may not import internal/game (its imports_test.go), so the
// frames are written here.
//
// The verdict is the one the readers give once they read answers_now:
// a move with a verdict is a response exactly when it is true (a cast an
// instant, an activation or a mana move an ability), a move without one
// is classed as before, and with the stack empty a matched move holds in
// any window, not only combat and an opponent's end step (rule 7's new
// clause). The default settings are Smart, the default "Stop for"
// categories (no untargeted abilities, no value instants on an
// opponent's item), the default stops grid (the seat's own main phases
// and combat declarations, nothing on an opponent's turn) with "only
// when I can do something" on, no pins, no bluffs, autoPassOwnStack on.
//
// Regenerate with:
//
//	go test ./internal/legal/ -run TestSmartPassAgreement -update

const smartPassFixturePath = "testdata/smart_pass_agreement.json"

// smartPassMove is one source and kind of a frame's cast, activation and
// mana moves, and the answers_now every one of them carries: true,
// false, or null for a move that keeps today's class.
type smartPassMove struct {
	Source     string `json:"source"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	AnswersNow *bool  `json:"answers_now"`
	Why        string `json:"why"`
}

// smartPassFrame is one window, the frame its viewer receives, and what
// both readers must make of it.
type smartPassFrame struct {
	Name    string            `json:"name"`
	Note    string            `json:"note"`
	Viewer  string            `json:"viewer"`
	View    protocol.GameView `json:"view"`
	Moves   []smartPassMove   `json:"moves"`
	Verdict string            `json:"verdict"`
	Why     string            `json:"why"`
}

// declared is a frame's hand-written answers_now for some sources, with
// why; every other cast, activation and mana move is declared null.
type declared map[uuid.UUID]struct {
	v   *bool
	why string
}

// frameOf projects a game into a fixture frame. Every cast, activation
// and mana move in the viewer's frame is listed once per source and
// kind, with its declared verdict.
func frameOf(t *testing.T, g *game.Game, viewer *game.Player, name, note string, want declared, stop bool, why string) smartPassFrame {
	t.Helper()
	view := protocol.ViewOfGameFor(g, viewer.ID.String())
	f := smartPassFrame{Name: name, Note: note, Viewer: viewer.ID.String(), View: view, Verdict: "pass", Why: why}
	if stop {
		f.Verdict = "stop"
	}
	type sk struct {
		source uuid.UUID
		kind   legal.Kind
	}
	seen := map[sk]bool{}
	for _, m := range view.LegalMoves {
		switch m.Kind {
		case legal.KindCast, legal.KindActivate, legal.KindMana:
		default:
			continue
		}
		k := sk{m.Source, m.Kind}
		if seen[k] {
			continue
		}
		seen[k] = true
		d := want[m.Source]
		why := d.why
		if why == "" {
			why = "declares nothing it answers, or only value: classed as before"
		}
		f.Moves = append(f.Moves, smartPassMove{
			Source: m.Source.String(), Name: cardNameOn(g, m.Source), Kind: string(m.Kind),
			AnswersNow: d.v, Why: why,
		})
	}
	for src := range want {
		found := false
		for _, m := range f.Moves {
			found = found || m.Source == src.String()
		}
		if !found {
			t.Fatalf("%s: %s offers no move: %v", name, cardNameOn(g, src), labels(view.LegalMoves))
		}
	}
	return f
}

// cardNameOn is the name of the card with this id on the battlefield or
// in a hand.
func cardNameOn(g *game.Game, id uuid.UUID) string {
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
			return c.Name
		}
	}
	for _, p := range g.Seats {
		for i := range p.Hand.Cards {
			if c := &p.Hand.Cards[i]; c.InstanceID == id {
				return c.Name
			}
		}
	}
	return ""
}

// buildSmartPassFrames is the fixture: the #4734 regression's frames,
// then a few windows it does not cover, on small boards.
func buildSmartPassFrames(t *testing.T) []smartPassFrame {
	t.Helper()
	col := fmColumn(t)
	var out []smartPassFrame
	for _, h := range heldFrames3376(t) {
		want := declared{
			h.fm: {h.fmWant(col), h.fmWhy},
			h.rc: {verdict(h.rcMatches), h.rcWhy},
		}
		if col == 0 {
			want[h.fm] = struct {
				v   *bool
				why string
			}{nil, "Flawless Maneuver declares nothing: classed as before, an instant"}
		}
		out = append(out, frameOf(t, h.g, h.claude, h.name, h.note, want, h.stop[col], h.stopWhy[col]))
	}

	// small is the light table of held_answers_3376a15c_test.go at turn 1
	// of `active`, with nothing on it but a creature for the seat
	// (Claude) and one for an opponent who is not the active player.
	type small struct {
		g                   *game.Game
		seat, active, other *game.Player
		bear, otherBear     uuid.UUID
	}
	newSmall := func(active string) *small {
		tb := newTable3376(t, active)
		s := &small{g: tb.g, seat: tb.claude, active: tb.by[active]}
		for _, n := range seats3376 {
			if n != active && n != "Claude" {
				s.other = tb.by[n]
				break
			}
		}
		s.bear = battlefieldCard(s.g, s.seat, creature("Seat Bear", "{1}{G}", 2, 2))
		s.otherBear = battlefieldCard(s.g, s.other, creature("Other Bear", "{1}{G}", 2, 2))
		return s
	}
	firstStrike := func(s *small, step game.Step) uuid.UUID {
		row := battlefieldCard(s.g, s.seat, declaredRelic(game.AnswerSurvive))
		a := battlefieldCard(s.g, s.active, game.Card{
			Name: "First Striker", TypeLine: "Creature — Knight", ManaCost: "{W}",
			Power: 1, Toughness: 1, Keywords: []string{"first strike"},
		})
		advanceTo(t, s.g, game.StepDeclareAttackers)
		if err := s.g.DeclareAttacker(a, s.seat.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, s.g, game.StepDeclareBlockers)
		if err := s.g.DeclareBlocker(s.bear, a); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		if err := s.g.FinishBlocks(s.seat.ID); err != nil {
			t.Fatalf("FinishBlocks: %v", err)
		}
		advanceTo(t, s.g, step)
		givePriority(s.g, s.seat)
		return row
	}
	{
		s := newSmall("krakenhavoc")
		row := firstStrike(s, game.StepFirstStrikeDamage)
		out = append(out, frameOf(t, s.g, s.seat, "first_strike_step_your_blocker_survive",
			"an opponent's first striker is blocked by the seat's creature; the first-strike damage step, the stack empty; the seat holds a survive row",
			declared{row: {verdict(true), "a creature of the seat's is blocking where combat damage is still to come"}},
			true, "a matched answer with the stack empty holds in any window (rule 7's new clause), here a step with no key window today"))
	}
	{
		s := newSmall("krakenhavoc")
		row := firstStrike(s, game.StepCombatDamage)
		out = append(out, frameOf(t, s.g, s.seat, "regular_damage_step_your_blocker_survive",
			"the same combat in the regular combat damage step: damage is dealt before anyone gets priority",
			declared{row: {verdict(false), "the regular combat damage step matches nothing (CR 117.5, 704.5g)"}},
			false, "nothing the seat holds matches"))
	}
	{
		s := newSmall("krakenhavoc")
		row := battlefieldCard(s.g, s.seat, declaredRelic(game.AnswerRestrict))
		advanceTo(t, s.g, game.StepDraw)
		givePriority(s.g, s.seat)
		out = append(out, frameOf(t, s.g, s.seat, "opponent_draw_step_restrict",
			"an opponent's draw step, the stack empty; the seat holds a restrict row",
			declared{row: {verdict(false), "restrict matches an opponent's upkeep, not their draw step"}},
			false, "nothing the seat holds matches"))
	}
	for _, victim := range []string{"yours", "another player's"} {
		s := newSmall("krakenhavoc")
		row := battlefieldCard(s.g, s.seat, declaredRelic(game.AnswerSurvive))
		advanceTo(t, s.g, game.StepPrecombatMain)
		target, name, matches := s.bear, "opponent_murder_on_your_creature_survive", true
		if victim != "yours" {
			target, name, matches = s.otherBear, "opponent_murder_on_other_creature_survive", false
		}
		pushSpellAW(s.g, s.active, game.Card{Name: "Murder", TypeLine: "Instant", ManaCost: "{1}{B}{B}", OracleID: oracleMurder},
			game.TargetRef{Kind: game.TargetCard, ID: target})
		givePriority(s.g, s.seat)
		why := "the stack threatens nothing of the seat's"
		verdictWhy := "nothing the seat holds answers the opponent's item"
		if matches {
			why = "an opponent's spell targets a creature of the seat's"
			verdictWhy = "a matched answer to an opponent's item is a response (rule 6)"
		}
		out = append(out, frameOf(t, s.g, s.seat, name,
			"an opponent's Murder aimed at "+victim+" creature; the seat holds a survive row",
			declared{row: {verdict(matches), why}}, matches, verdictWhy))
	}
	{
		s := newSmall("krakenhavoc")
		battlefieldCard(s.g, s.seat, basic("Island", "Island"))
		opt := handCard(s.seat, game.Card{Name: "Opt", TypeLine: "Instant", ManaCost: "{U}", OracleID: oracleOptVO})
		advanceTo(t, s.g, game.StepPrecombatMain)
		pushSpellAW(s.g, s.active, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
		givePriority(s.g, s.seat)
		out = append(out, frameOf(t, s.g, s.seat, "value_instant_on_opponent_spell",
			"an opponent's creature spell on the stack; the seat holds Opt and an Island",
			declared{opt: {nil, "declared exactly value: value_only, and no verdict (#2877)"}},
			false, "a value instant is no response to an opponent's item by default (#2877), and the mana is none"))
	}
	{
		s := newSmall("krakenhavoc")
		row := battlefieldCard(s.g, s.seat, game.Card{
			Name: "Instance Relic", TypeLine: "Artifact",
			ActivatedAbilities: []game.ActivatedAbilityShape{{
				Label:  "Do something.",
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}},
		})
		advanceTo(t, s.g, game.StepEnd)
		givePriority(s.g, s.seat)
		out = append(out, frameOf(t, s.g, s.seat, "undeclared_row_at_opponent_end_step",
			"an opponent's end step, the stack empty; the seat holds a run-time row that declares nothing",
			declared{row: {nil, "declares nothing: interacts, classed as before"}},
			true, "an undeclared row is an ability as before, and an opponent's end step is a key window"))
	}
	return out
}

// TestSmartPassAgreement asserts every frame's moves carry the declared
// answers_now, and writes the fixture the readers' tests read.
func TestSmartPassAgreement(t *testing.T) {
	frames := buildSmartPassFrames(t)
	for _, f := range frames {
		t.Run(f.Name, func(t *testing.T) {
			if len(f.Moves) == 0 {
				t.Fatal("the frame declares no move")
			}
			if f.Verdict != "stop" && f.Verdict != "pass" {
				t.Fatalf("verdict %q is neither stop nor pass", f.Verdict)
			}
			for _, want := range f.Moves {
				n := 0
				for _, m := range f.View.LegalMoves {
					if m.Source.String() != want.Source || string(m.Kind) != want.Kind {
						continue
					}
					n++
					if verdictString(m.AnswersNow) != verdictString(want.AnswersNow) {
						t.Errorf("%s %s: answers_now %s, declared %s (%s)", want.Name, want.Kind,
							verdictString(m.AnswersNow), verdictString(want.AnswersNow), want.Why)
					}
				}
				if n == 0 {
					t.Errorf("%s %s: no such move in the frame", want.Name, want.Kind)
				}
			}
		})
	}
	writeSmartPassFixture(t, frames)
}

// writeSmartPassFixture compares the generated fixture against the
// committed one, or rewrites it under -update, as writeFixture does.
func writeSmartPassFixture(t *testing.T, frames []smartPassFrame) {
	t.Helper()
	body, err := json.MarshalIndent(frames, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	body = append(canonicaliseIDs(body), '\n')
	if *updateFixture {
		if err := os.WriteFile(smartPassFixturePath, body, 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		t.Logf("wrote %s (%d frames, %d B)", smartPassFixturePath, len(frames), len(body))
		return
	}
	have, err := os.ReadFile(smartPassFixturePath)
	if err != nil {
		t.Fatalf("read %s (regenerate with `go test ./internal/legal/ -run TestSmartPassAgreement -update`): %v", smartPassFixturePath, err)
	}
	if string(have) != string(body) {
		t.Errorf("%s is stale: the frames the smart pass readers' tests read no longer match what this package produces.\n"+
			"Regenerate with: go test ./internal/legal/ -run TestSmartPassAgreement -update", smartPassFixturePath)
	}
}

// TestSmartPassAgreementFixtureIsSeatSafe re-checks, on the committed
// bytes, that no frame carries another seat's move list.
func TestSmartPassAgreementFixtureIsSeatSafe(t *testing.T) {
	body, err := os.ReadFile(smartPassFixturePath)
	if err != nil {
		t.Skipf("no fixture yet: %v", err)
	}
	var frames []smartPassFrame
	if err := json.Unmarshal(body, &frames); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	if len(frames) == 0 {
		t.Fatal("fixture is empty")
	}
	for _, f := range frames {
		for _, m := range f.View.LegalMoves {
			if m.Player.String() != f.Viewer {
				t.Errorf("%s: frame for %s carries a move belonging to %s", f.Name, f.Viewer, m.Player)
			}
		}
	}
}
