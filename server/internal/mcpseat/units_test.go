package mcpseat

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestParseInvite(t *testing.T) {
	id := uuid.NewString()
	for _, tc := range []struct {
		raw    string
		ok     bool
		kind   inviteKind
		origin string
	}{
		{"https://cmd-dev.labxp.io/#/games/" + id + "/join?t=abc", true, inviteJoin, "https://cmd-dev.labxp.io"},
		{"http://127.0.0.1:5173/#/games/" + id + "/reclaim?t=tick", true, inviteReclaim, "http://127.0.0.1:5173"},
		{"https://cmd.labxp.io/games/" + id + "/join?t=abc", true, inviteJoin, "https://cmd.labxp.io"},
		{"https://cmd.labxp.io/#/games/" + id + "/join", false, 0, ""},
		{"https://cmd.labxp.io/#/games/not-a-uuid/join?t=abc", false, 0, ""},
		{"javascript:alert(1)", false, 0, ""},
		{"https://cmd.labxp.io/#/games/" + id + "/spectate?t=abc", false, 0, ""},
	} {
		inv, err := parseInvite(tc.raw)
		if (err == nil) != tc.ok {
			t.Errorf("%s: err = %v", tc.raw, err)
			continue
		}
		if tc.ok && (inv.Kind != tc.kind || inv.Origin != tc.origin || inv.GameID.String() != id) {
			t.Errorf("%s: %+v", tc.raw, inv)
		}
	}
}

func TestOriginAllowlist(t *testing.T) {
	def, err := NewOriginAllowlist(nil)
	if err != nil {
		t.Fatal(err)
	}
	for raw, want := range map[string]bool{
		"https://cmd.labxp.io":      true,
		"https://cmd-dev.labxp.io":  true,
		"http://localhost:8080":     true,
		"http://127.0.0.1:41234":    true,
		"http://cmd.labxp.io":       false, // wrong scheme
		"https://cmd.labxp.io:8443": false, // wrong port
		"https://evil.labxp.io":     false,
		"https://localhost:8080":    false,
	} {
		u, _ := url.Parse(raw)
		if got := def.Allows(u); got != want {
			t.Errorf("default allows %s = %v, want %v", raw, got, want)
		}
	}
	only, err := NewOriginAllowlist([]string{"https://cmd-dev.labxp.io/"})
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse("https://cmd.labxp.io")
	if only.Allows(u) {
		t.Error("an explicit list must replace the defaults")
	}
	for _, bad := range []string{"cmd.labxp.io", "ftp://x", "https://x/path", "https://user@x"} {
		if _, err := NewOriginAllowlist([]string{bad}); err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
}

func TestStateFileRefusesLoosePermissions(t *testing.T) {
	root := filepath.Join(t.TempDir(), "state")
	st := stateStore{root: root}
	ss := &savedSession{Origin: "https://cmd-dev.labxp.io", GameID: uuid.New(), PlayerID: uuid.New(), Token: "tok-123456789", ExpiresAt: time.Now().Add(time.Hour)}
	if err := st.save(ss); err != nil {
		t.Fatal(err)
	}
	got, err := st.load(ss.Origin, ss.GameID)
	if err != nil || got == nil || got.Token != ss.Token {
		t.Fatalf("load = %+v, %v", got, err)
	}
	path := st.pathFor(ss.Origin, ss.GameID)
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := st.load(ss.Origin, ss.GameID); !errors.Is(err, errInsecureState) {
		t.Errorf("a 0644 token file was read: %v", err)
	}
	_ = os.Chmod(path, 0o600)
	if err := os.Chmod(st.dirFor(ss.Origin), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := st.load(ss.Origin, ss.GameID); !errors.Is(err, errInsecureState) {
		t.Errorf("a 0755 directory was used: %v", err)
	}
	_ = os.Chmod(st.dirFor(ss.Origin), 0o700)
	if err := st.remove(ss.Origin, ss.GameID); err != nil {
		t.Fatal(err)
	}
	if got, err := st.load(ss.Origin, ss.GameID); got != nil || err != nil {
		t.Errorf("after remove: %+v %v", got, err)
	}
}

func TestAgentClientName(t *testing.T) {
	for in, want := range map[string]string{
		"claude-code":                 "claude-code",
		"Claude Code":                 "claude-code",
		"codex_cli.v2":                "codex_cli.v2",
		"":                            "unknown",
		"!!!":                         "unknown",
		strings.Repeat("abcdefgh", 6): strings.Repeat("abcdefgh", 4),
		"Ünïcode/../../etc":           "ncode....etc",
	} {
		if got := agentClientName(in); got != want {
			t.Errorf("agentClientName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseAbsorbOnlyNamesLayerARules(t *testing.T) {
	got, err := ParseAbsorb("forced,mana-only,coin-call")
	if err != nil || len(got) != 3 || got[rules.RuleSameLand] {
		t.Fatalf("default: %v %v", got, err)
	}
	if got, err := ParseAbsorb("none"); err != nil || len(got) != 0 {
		t.Errorf("none: %v %v", got, err)
	}
	if got, err := ParseAbsorb("forced,same-land"); err != nil || !got[rules.RuleSameLand] {
		t.Errorf("same-land can be switched on: %v %v", got, err)
	}
	for _, bad := range []string{"escalate", "no-moves", "heuristic", "everything"} {
		if _, err := ParseAbsorb(bad); err == nil {
			t.Errorf("--absorb=%s was accepted; nothing outside Layer A may be added", bad)
		}
	}
}

func TestDecideAutomaticallyFollowsTheTable(t *testing.T) {
	me := uuid.New()
	v := &protocol.GameView{State: "active", Seats: []protocol.PlayerView{{ID: me.String(), Seat: 0}, {ID: uuid.NewString(), Seat: 1}},
		Turn: protocol.TurnView{ActiveSeat: 1}}
	pass := legal.Move{Type: legal.TypePassPriority, Kind: legal.KindPass, Label: "Pass"}
	mana := legal.Move{Type: legal.TypeActivateManaAbility, Kind: legal.KindMana, Label: "Tap"}
	land := legal.Move{Type: "play_land", Kind: legal.KindLand, Label: "Play Forest"}
	cast := legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: "Cast"}
	attack := legal.Move{Type: legal.TypeDeclareAttacker, Kind: legal.KindAttack, Label: "Attack"}
	def, _ := ParseAbsorb(strings.Join(DefaultAbsorb, ","))

	for _, tc := range []struct {
		name      string
		moves     []legal.Move
		absorb    map[string]bool
		passUntil bool
		wantAuto  bool
		wantRule  string
	}{
		{"forced", []legal.Move{pass}, def, false, true, rules.RuleForced},
		{"mana-only", []legal.Move{pass, mana}, def, false, true, rules.RuleManaOnly},
		{"same-land escalates by default", []legal.Move{pass, land, land}, def, false, false, ""},
		{"same-land when switched on", []legal.Move{pass, land, land}, map[string]bool{rules.RuleSameLand: true}, false, true, rules.RuleSameLand},
		{"a removed rule escalates", []legal.Move{pass}, map[string]bool{}, false, false, ""},
		{"a real choice", []legal.Move{pass, cast}, def, false, false, ""},
		{"pass_until on another's turn", []legal.Move{pass, cast}, def, true, true, rulePassUntil},
		{"pass_until never answers a declaration", []legal.Move{attack, attack}, def, true, false, ""},
	} {
		ans, ok := decideAutomatically(v, me, tc.moves, tc.absorb, tc.passUntil)
		if ok != tc.wantAuto || (ok && ans.rule != tc.wantRule) {
			t.Errorf("%s: auto=%v rule=%q, want %v %q", tc.name, ok, ans.rule, tc.wantAuto, tc.wantRule)
		}
	}

	stacked := *v
	stacked.StackItems = []protocol.StackItemView{{ID: "x"}}
	if _, ok := decideAutomatically(&stacked, me, []legal.Move{pass, cast}, def, true); ok {
		t.Error("pass_until passed with something on the stack")
	}
	own := *v
	own.Turn.ActiveSeat = 0
	if _, ok := decideAutomatically(&own, me, []legal.Move{pass, cast}, def, true); ok {
		t.Error("pass_until passed on the seat's own turn")
	}
	loop := *v
	loop.LoopNotice = &protocol.LoopNoticeView{Label: "x", Count: 3}
	if _, ok := decideAutomatically(&loop, me, []legal.Move{pass}, def, true); ok {
		t.Error("a pass went out under the loop notice")
	}
}

// bigBoard is a table far past the compact budget.
func bigBoard(me string) *protocol.GameView {
	v := &protocol.GameView{State: "active", Turn: protocol.TurnView{Number: 9, Step: "main1"}}
	for s := 0; s < 4; s++ {
		id := me
		if s > 0 {
			id = uuid.NewString()
		}
		p := protocol.PlayerView{ID: id, Name: fmt.Sprintf("Player %d", s), Seat: s, Life: 30}
		for i := 0; i < 60; i++ {
			p.Graveyard.Cards = append(p.Graveyard.Cards, protocol.CardView{InstanceID: uuid.NewString(), Name: fmt.Sprintf("Long Graveyard Card Name Number %d", i)})
		}
		p.Graveyard.Count = 60
		v.Seats = append(v.Seats, p)
		for i := 0; i < 40; i++ {
			v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{
				InstanceID: uuid.NewString(), Controller: id, Name: fmt.Sprintf("Enormous Creature With A Long Name %d", i),
				TypeLine: "Creature — Beast", Power: i, Toughness: i, Abilities: []string{"flying", "trample", "vigilance"},
				Counters: map[string]int{"+1/+1": 2},
			})
		}
	}
	return v
}

func TestTheCompactBoardHoldsItsBudgetAndTheMovesAreNeverCut(t *testing.T) {
	me := uuid.NewString()
	v := bigBoard(me)
	var logLines, chat []string
	for i := 0; i < 30; i++ {
		logLines = append(logLines, untrusted(fmt.Sprintf("Player 1 cast a very long spell name number %d at Player 2", i), maxLogLen))
		chat = append(chat, chatLine(protocol.ChatPayload{AuthorName: "Bob", Text: strings.Repeat("blah ", 50)}))
	}
	board := compactBoard(v, me, logLines, chat)
	if len(board) > budgetCompact {
		t.Fatalf("compact board is %d bytes, budget %d", len(board), budgetCompact)
	}
	if !strings.Contains(board, "(YOU)") {
		t.Errorf("the seat's own line was cut:\n%s", board)
	}
	full := fullBoard(v, me)
	if len(full) > budgetFull {
		t.Fatalf("full board is %d bytes, budget %d", len(full), budgetFull)
	}

	w := &window{token: "w0.1"}
	for i := 0; i < 400; i++ {
		w.moves = append(w.moves, wireMove{Move: legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: fmt.Sprintf("Cast thing %d", i)}})
	}
	moves := renderMoves(v, w, "")
	if !strings.Contains(moves, "  399: Cast thing 399") {
		t.Error("the move list was cut")
	}
}

func TestTheUnimplementedNoteIsSpelledOut(t *testing.T) {
	me := uuid.NewString()
	v := &protocol.GameView{State: "active", Seats: []protocol.PlayerView{{ID: me, Name: "Agent"}}}
	v.Battlefield.Cards = []protocol.CardView{{InstanceID: "a", Controller: me, Name: "Odd Card", Unimplemented: true}}
	board := compactBoard(v, me, nil, nil)
	if !strings.Contains(board, "unimplemented: the engine does not run this card's text") {
		t.Errorf("no unimplemented note:\n%s", board)
	}
}
