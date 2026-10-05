package game

// agent_badge_test.go: ADR 0122 §7, the agent badge on the engine's
// seat. It is set once, at join, and nothing takes it off.

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"math/rand/v2"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func agentTestGame(t *testing.T) (*Game, *Player, *Player) {
	t.Helper()
	g := NewGame()
	var seats []*Player
	for i := range 2 {
		deck := []Card{NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		p, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck)
		if err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
		seats = append(seats, p)
	}
	return g, seats[0], seats[1]
}

func TestSetAgentIsOneWayAndNeverABot(t *testing.T) {
	g, human, agent := agentTestGame(t)

	if err := g.SetAgent(agent.ID, "claude-code"); err != nil {
		t.Fatalf("SetAgent: %v", err)
	}
	// A second declaration does not relabel the seat.
	if err := g.SetAgent(agent.ID, "codex"); err != nil {
		t.Fatalf("SetAgent again: %v", err)
	}
	if p := g.PlayerByID(agent.ID); !p.Agent || p.AgentClient != "claude-code" {
		t.Errorf("agent seat = Agent %v / %q, want true / claude-code", p.Agent, p.AgentClient)
	}
	if p := g.PlayerByID(human.ID); p.Agent || p.AgentClient != "" {
		t.Errorf("the other seat is badged: %v / %q", p.Agent, p.AgentClient)
	}

	// A seat is a bot, an agent or neither.
	if err := g.SetBot(agent.ID, "random", ""); !errors.Is(err, ErrSeatKindTaken) {
		t.Errorf("SetBot on an agent seat: %v, want ErrSeatKindTaken", err)
	}
	if err := g.SetBot(human.ID, "random", ""); err != nil {
		t.Fatalf("SetBot: %v", err)
	}
	if err := g.SetAgent(human.ID, "claude-code"); !errors.Is(err, ErrSeatKindTaken) {
		t.Errorf("SetAgent on a bot seat: %v, want ErrSeatKindTaken", err)
	}
	if err := g.SetAgent(uuid.New(), "x"); !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("SetAgent on no seat: %v, want ErrPlayerNotFound", err)
	}

	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := g.SetAgent(agent.ID, "x"); !errors.Is(err, ErrGameNotInLobby) {
		t.Errorf("SetAgent after the start: %v, want ErrGameNotInLobby", err)
	}
}

// TestAgentBadgeSurvivesCloneAndRestore: the two copies of a seat the
// engine makes, an undo checkpoint and a restore point, keep the badge.
func TestAgentBadgeSurvivesCloneAndRestore(t *testing.T) {
	g, human, agent := agentTestGame(t)
	if err := g.SetAgent(agent.ID, "codex"); err != nil {
		t.Fatalf("SetAgent: %v", err)
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 4))); err != nil {
		t.Fatalf("Start: %v", err)
	}

	check := func(label string, gg *Game) {
		t.Helper()
		if p := gg.PlayerByID(agent.ID); p == nil || !p.Agent || p.AgentClient != "codex" {
			t.Errorf("%s: agent seat = %+v", label, p)
		}
		if p := gg.PlayerByID(human.ID); p == nil || p.Agent || p.AgentClient != "" {
			t.Errorf("%s: human seat badged", label)
		}
	}

	check("clone", g.Clone())

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatal("snapshot not restorable")
	}
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	check("restore", restored)

	// The keys are omitted on a seat that is not an agent's, so a file
	// without agents is byte-for-byte what it was before ADR 0122.
	for _, s := range snap.Seats {
		if s.ID == human.ID && (s.Agent || s.AgentClient != "") {
			t.Errorf("human seat snapshot carries the badge")
		}
	}
}

// agentBadgeFields are the names the badge goes by: game.Player's and
// the snapshot's Agent / AgentClient, and the wire's IsAgent (PlayerView
// and lobby.SeatInfo; AgentClient is shared).
var agentBadgeFields = map[string]bool{"Agent": true, "AgentClient": true, "IsAgent": true}

// agentBadgeWriters is every function in the server that writes the
// badge, and how many writes each makes. Each is either THE one place
// the badge is set (game.SetAgent, and lobby.join stamping the same
// declaration on the SeatInfo), or a copy of a badge that already
// exists from one representation of the seat to another. None may
// write a constant false or "": that is checked separately below.
//
// A new writer fails TestAgentBadgeHasNoClearingWriter until it is
// listed here, and listing it is the review question: does this code
// copy the badge, or could it take it off?
var agentBadgeWriters = map[string]int{
	"internal/game/mutations.go:SetAgent":      2, // the one setter: Agent = true, AgentClient = client
	"internal/game/clone.go:clonePlayer":       2, // undo checkpoint copy
	"internal/game/snapshot.go:snapshotPlayer": 2, // restore point capture
	"internal/game/snapshot.go:restorePlayer":  2, // restore point read
	"internal/protocol/view.go:viewOfPlayer":   2, // PlayerView projection
	"internal/lobby/lobby.go:join":             2, // SeatInfo at join, from the same declaration
	// SeatInfo on restore: read from game.Player, or, only when the
	// engine has no such player, from seats.agent_client when it is
	// set (ADR 0124 §6). Both onto a fresh SeatInfo, both true / a
	// non-empty client.
	"internal/lobby/persist.go:loadEntry": 4,
	// seats.agent_client from SeatInfo, for an agent seat only; the
	// normalised name is never empty (ADR 0124 §6).
	"internal/lobby/persist.go:seatRecords": 1,
	// seats.agent_client read into a fresh SeatRecord (ADR 0124 §6).
	"internal/lobby/store_sql.go:LoadGame": 1,
	// the admin views' copy, onto a fresh LiveSeat for an agent seat
	// only (ADR 0124 §5).
	"internal/lobby/live_tables.go:LiveTables": 1,
	// The admin views' copies (ADR 0124 §3.3), each a read of a badge
	// that exists onto a fresh value: LiveSeat into adminview's overlay,
	// seats.agent_client into a fresh SeatRow, and a row's or memory's
	// client onto a fresh served Seat, only when it is set.
	"internal/lobby/admin_views.go:adminLiveTables": 1,
	"internal/adminview/store_sql.go:attachSeats":   1,
	"internal/adminview/merge.go:rowSeat":           1,
	"internal/adminview/merge.go:overlaySeat":       1,
	"internal/adminview/merge.go:liveSeat":          1,
}

// TestAgentBadgeHasNoClearingWriter pins ADR 0122 §7's "the badge
// cannot be removed": no route, action, setting, reclaim, restore or
// admin path takes it off a seat. It walks every non-test Go file in
// the server and finds each write of a badge field, as an assignment
// (x.Agent = …) or a composite-literal key (Agent: …), and fails on:
//
//   - a write in a function not in agentBadgeWriters, or a different
//     number of writes than listed;
//   - any write of a constant false or "" anywhere, listed or not.
//
// The behavioural half, the lobby routes that could plausibly clear it
// (host hand-over, Discord linking, reclaim, restore), is
// TestAgentBadgeCannotBeCleared in internal/lobby.
func TestAgentBadgeHasNoClearingWriter(t *testing.T) {
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()
	got := map[string]int{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "testdata", "vendor", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			site := rel + ":" + fn.Name.Name
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.AssignStmt:
					for i, lhs := range n.Lhs {
						sel, ok := lhs.(*ast.SelectorExpr)
						if !ok || !agentBadgeFields[sel.Sel.Name] {
							continue
						}
						got[site]++
						if i < len(n.Rhs) && isZeroConstant(n.Rhs[i]) {
							t.Errorf("%s: %s.%s is written a zero value; the agent badge never comes off (ADR 0122 §7)",
								fset.Position(n.Pos()), exprString(sel.X), sel.Sel.Name)
						}
					}
				case *ast.KeyValueExpr:
					key, ok := n.Key.(*ast.Ident)
					if !ok || !agentBadgeFields[key.Name] {
						return true
					}
					got[site]++
					if isZeroConstant(n.Value) {
						t.Errorf("%s: %s: is written a zero value; the agent badge never comes off (ADR 0122 §7)",
							fset.Position(n.Pos()), key.Name)
					}
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	var sites []string
	for s := range got {
		sites = append(sites, s)
	}
	sort.Strings(sites)
	for _, s := range sites {
		want, listed := agentBadgeWriters[s]
		switch {
		case !listed:
			t.Errorf("%s writes the agent badge %d time(s) and is not in agentBadgeWriters: if it copies the badge, list it; if it could clear it, it must not exist (ADR 0122 §7)", s, got[s])
		case got[s] != want:
			t.Errorf("%s writes the agent badge %d time(s), agentBadgeWriters says %d", s, got[s], want)
		}
	}
	// Control: the setter and the copies are really there, so the walk
	// is matching something. A rename that hid them would otherwise
	// pass this test vacuously.
	for s, want := range agentBadgeWriters {
		if _, seen := got[s]; !seen {
			t.Errorf("%s is listed with %d write(s) but the walk found none: renamed or moved? Update agentBadgeWriters", s, want)
		}
	}
}

// isZeroConstant reports whether e is the literal false or "".
func isZeroConstant(e ast.Expr) bool {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name == "false"
	case *ast.BasicLit:
		return v.Kind == token.STRING && (v.Value == `""` || v.Value == "``")
	case *ast.ParenExpr:
		return isZeroConstant(v.X)
	}
	return false
}

func exprString(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return exprString(v.X) + "." + v.Sel.Name
	case *ast.IndexExpr:
		return exprString(v.X) + "[…]"
	}
	return "…"
}
