package protocol

import (
	"go/ast"
	"go/parser"
	gotoken "go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_event_kind_gate_test.go — #984. The public log's coverage of the
// engine's event kinds is one `switch ev.Kind` in projectEvent, and
// nothing makes a new kind pass through it.
//
// The bug that produced this file: `EventColorChosen` (#742),
// `EventCreatureTypeChosen` (S26) and `EventPlayerChosen` (#1007) each
// record a CR 105.4 / CR 614.12 answer that is made OUT LOUD at the
// table, each landed with a card, a view field and a test — and none
// of the three reached the log, so "Ian chose green for Coldsteel
// Heart" was a fact the table never saw written down. Two of them sat
// unlogged for two sprints. Nothing failed: the default arm of a
// switch is a silence, and a silence that should be a line looks
// exactly like a silence that should be a silence.
//
// So this test reads the kinds and the switch straight out of the
// source — the mechanism `TestEveryReplacementEventKindIsSwitchedOn`
// (internal/game/replacement_kind_gate_test.go) and
// `TestEveryChoiceKindIsClassifiedAndEnumerated`
// (internal/legal/choice_gate_test.go) already use — and fails the
// moment an EventKind is declared that projectEvent neither narrates
// nor is listed below as a deliberate silence.
//
// The silence list is the point as much as the arms are. The log is
// the table's memory and a policy's history (ADR 0033 §4); "the log
// says nothing about this" is a decision, and a decision belongs in
// writing next to the reason for it.

// Shared reasons. A kind whose silence is its own argument gets its
// own string below instead.
const (
	silentAlreadyToldAsAZoneMove = "the LogZone entry for the same motion already says it — one fact, one line"
	silentStepSpine              = "the step is one LogStep line; these are the harvester's per-step hooks and would say it a second time"
	silentManaBookkeeping        = "mana is pool bookkeeping, read off the mana display and not the history; one turn emits dozens"
	silentServerDiagnostic       = "a server diagnostic, not something a player at the table observed"
	silentBoardStateIsVisible    = "the resulting board state is on the wire and the table reads it there; a line per change would bury the ones that matter"
	silentImpliedByAnotherLine   = "implied by a line the log already emits for the same action"
)

// silentEventKinds is every game.EventKind projectEvent deliberately
// drops, and why. Adding a kind here is a decision to keep the table
// uninformed about it; make it on purpose.
//
// Six of the entries this table shipped with were gaps rather than
// decisions — a control change and a counter landing are both things a
// player announces out loud — and #1021 turned those eight kinds into
// eight arms, which is what the list is FOR: they were visible in one
// place instead of being invisible in the default arm of a switch. The
// rows that remain are the ones whose silence survived being read.
var silentEventKinds = map[string]string{
	// --- already told, by another line -------------------------------
	"EventETB":            silentAlreadyToldAsAZoneMove,
	"EventLTB":            silentAlreadyToldAsAZoneMove,
	"EventMill":           silentAlreadyToldAsAZoneMove,
	"EventDiscardCard":    silentAlreadyToldAsAZoneMove,
	"EventBecomesTarget":  "the cast or activation line already named the spell; targeting is announce-time bookkeeping the stack view carries",
	"EventBecomesBlocked": "the LogBlock entry for the blocker is the same fact from the other side",
	// #1279. The blocks a declaration made are LogBlock lines; what
	// this adds is only that the defender has FINISHED, which the turn
	// cursor carries live (block_pending_seats / blocks_declared_seats)
	// and the next line — a block, a resolution, the next step — makes
	// obvious in the history. A "declares no blockers" line is a
	// reasonable follow-up; it needs a log kind and client copy.
	"EventBlockersDeclared": "the declaration's blocks are its LogBlock lines, and its completion is on the turn cursor (block_pending_seats / blocks_declared_seats)",
	// #1257: this reason was not true when it was written. The
	// ability's LogResolve carried no card and no label and rendered
	// as "a card resolved". It now names the ability by its stack
	// label and its source (projectAbilityItem), redacted with the
	// source — log_ability_resolve_test.go pins both halves.
	"EventTrigger": "a trigger reaching the stack is told by the LogResolve of the ability it becomes, which names it by its label and source (#1257)",
	// #1184: the same argument as the row above, for the other half of
	// the same announcement. An activation reaching the stack is told
	// by the LogResolve of the ability it becomes, and the stack view
	// carries it in the meantime; a second line at the announce would
	// say the same thing twice about one click. The kind exists so
	// TRIGGERS can watch an activation, not so the log can narrate
	// one.
	"EventActivateAbility": "an activation reaching the stack is told by the LogResolve of the ability it becomes, which names it by its label and source (#1257)",
	"EventKeywordAction":   silentImpliedByAnotherLine,

	// --- the step spine ----------------------------------------------
	"EventTurnBegan":          "the first EventStepBegan announces the same boundary to the table; this kind exists for engine consumers after per-turn resets",
	"EventStepTransition":     silentStepSpine,
	"EventBeginUpkeep":        silentStepSpine,
	"EventBeginDrawStep":      silentStepSpine,
	"EventBeginPrecombatMain": silentStepSpine,
	"EventBeginEndStep":       silentStepSpine,

	// --- mana --------------------------------------------------------
	"EventManaAdded":            silentManaBookkeeping,
	"EventManaSpent":            silentManaBookkeeping,
	"EventManaPoolEmptied":      silentManaBookkeeping,
	"EventManaAbilityActivated": silentManaBookkeeping,

	// --- diagnostics --------------------------------------------------
	"EventEffectError":             silentServerDiagnostic,
	"EventCostWarning":             silentServerDiagnostic,
	"EventLoopSuspected":           "GameView.loop_notice is the table's copy of this one, and it is a live control rather than history (#628)",
	"EventPendingChoiceDropped":    silentServerDiagnostic,
	"EventPendingChoiceReassigned": silentServerDiagnostic,
	// #1017 / #812 declared this silence when it added the kind, and
	// this row is that decision moved to where the test can hold it.
	// CR 701.3b: an attach that cannot happen simply doesn't happen,
	// and CR 608.2 / CR 702.6a keep the ability resolving — so NOTHING
	// changed, and a line saying so would be a line about a
	// non-event. The breadcrumb exists for a stall dump and the
	// catalog soak, which is exactly the posture the two rows above
	// take.
	"EventAttachSkipped": "CR 701.3b's attach that doesn't happen changes nothing a player could observe; the event is a breadcrumb for a stall dump, not a table fact (#812, #1017)",
	// ADR 0082 / #1194. Turning a permanent face up is loud and the
	// table is told — by the LogSpecialAction line for the CR 116.2g
	// action that did it, which carries the card and the price
	// ("Turn face up {1}{U}") and lands in the same frame as the
	// permanent's identity. turnFaceUpLocked is the ONE emitter of
	// this kind and it only ever runs from that action, so the two
	// are the same moment; a second line would say it twice. If a
	// second emitter ever appears — an effect that turns permanents
	// face up without a special action — this row stops being true
	// and the kind needs an arm.
	"EventTurnedFaceUp": "the CR 116.2g special action that turns a permanent face up is the only emitter, and its LogSpecialAction line already names the card and the price",
	// #1382. A card becomes plotted two ways and the table is told
	// both: the plot special action's LogSpecialAction line names the
	// card and the price ("Plot {3}{R}"), and an "it becomes plotted"
	// effect is the resolution of a spell or ability whose LogResolve
	// names it, with the LogZone line for the exile beside it. The
	// plotted state itself is on the wire as the card's cast
	// permission. The kind exists for "when this card becomes plotted"
	// to watch, not for the log to say a third time.
	"EventBecomesPlotted": "the plot special action's LogSpecialAction line, or the LogResolve of the effect that plotted it plus its LogZone exile, already tells the table (#1382)",

	// --- visible board state -------------------------------------------
	"EventTapCard":        silentBoardStateIsVisible,
	"EventUntapCard":      silentBoardStateIsVisible,
	"EventAttach":         silentBoardStateIsVisible,
	"EventUnattach":       silentBoardStateIsVisible,
	"EventCopyApplied":    silentBoardStateIsVisible,
	"EventCaseSolved":     silentBoardStateIsVisible,
	"EventHarnessed":      silentBoardStateIsVisible,
	"EventRegenerated":    silentImpliedByAnotherLine,
	"EventBattleDefeated": silentBoardStateIsVisible,

	// --- hidden-zone work ----------------------------------------------
	"EventSearchLibrary": "the number of matches is itself hidden information about a hidden zone (see the search_library prompt's redaction)",

	// --- waiting on the PR that writes the line ------------------------
	// The same posture as EventSettingsChanged above, and the same
	// shape of promise: the kind lands with the engine seam, the line
	// lands with the client PR that has somewhere to put it.
	//
	// ADR 0056 Decision 6 specifies it exactly — "Alice got 3 poison
	// counters (7/10)", one entry per placement with a positive delta,
	// the total in brackets — and puts it in PR 3 with the N/10 poison
	// chip and the damage-result suffixes it has to read alongside.
	// Writing half of it here would mean a LogKind and a protocol.ts
	// shape that PR 3 immediately rewrites.
	//
	// Until then PlayerView.poison carries the number to every viewer
	// on every frame, which is what the table has read since S10; what
	// is missing is the HISTORY, and that is what makes this a
	// temporary silence rather than a permanent one.
	"EventPlayerCounterPlaced": "TEMPORARY: ADR 0056 Decision 6 gives it the `poison` line (\"Alice got 3 poison counters (7/10)\") in the client PR; until then PlayerView.poison shows every viewer the current total",
}

// TestEveryEventKindIsNarratedOrDeliberatelySilent is the mechanism.
func TestEveryEventKindIsNarratedOrDeliberatelySilent(t *testing.T) {
	declared := declaredEventKinds(t)
	if len(declared) < 50 {
		t.Fatalf("found only %d EventKind constants in ../game — the scanner has stopped working", len(declared))
	}
	arms := narratedEventKinds(t)
	if len(arms) < 10 {
		t.Fatalf("found only %d arms in projectEvent — the scanner has stopped working", len(arms))
	}

	names := make([]string, 0, len(declared))
	for name := range declared {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if arms[name] {
			continue
		}
		reason, listed := silentEventKinds[name]
		if !listed {
			t.Errorf("game.%s (%q) has no arm in projectEvent and no entry in silentEventKinds: "+
				"either narrate it in server/internal/protocol/log.go or say in silentEventKinds why the table is not told (#984)",
				name, declared[name])
			continue
		}
		if strings.TrimSpace(reason) == "" {
			t.Errorf("game.%s (%q) is listed as a deliberate silence with an EMPTY reason — say why, or give it an arm",
				name, declared[name])
		}
	}

	// The reverse direction, twice over. A stale arm is dead code; a
	// stale silence is a reason nobody will ever read again, and both
	// are the only evidence that a deletion was incomplete.
	for name := range arms {
		if _, ok := declared[name]; !ok {
			t.Errorf("projectEvent has an arm for game.%s, which is not a declared EventKind", name)
		}
	}
	silent := make([]string, 0, len(silentEventKinds))
	for name := range silentEventKinds {
		silent = append(silent, name)
	}
	sort.Strings(silent)
	for _, name := range silent {
		if _, ok := declared[name]; !ok {
			t.Errorf("silentEventKinds explains game.%s, which is not a declared EventKind", name)
		}
		if arms[name] {
			t.Errorf("game.%s has BOTH an arm in projectEvent and an entry in silentEventKinds — one of the two is a leftover", name)
		}
	}
}

// TestTheThreeChosenValueKindsAreNarrated is #984 stated as a fact
// rather than as a mechanism: the three kinds the gate above was
// written for have arms, and would fail it if they lost them.
func TestTheThreeChosenValueKindsAreNarrated(t *testing.T) {
	arms := narratedEventKinds(t)
	for _, name := range []string{"EventColorChosen", "EventCreatureTypeChosen", "EventPlayerChosen"} {
		if !arms[name] {
			t.Errorf("projectEvent has no arm for game.%s — a chosen colour, type or player is never narrated (#984)", name)
		}
	}
}

// TestTheSixReadSilencesAreNarrated is the same statement for #1021:
// the six silences that read as gaps when the table was first written
// down are lines now, and losing one puts it back in the dark rather
// than failing the gate above (which a re-added silentEventKinds row
// would satisfy).
func TestTheSixReadSilencesAreNarrated(t *testing.T) {
	arms := narratedEventKinds(t)
	for _, name := range []string{
		"EventControlChanged",
		"EventSpecialAction",
		"EventCycle",
		"EventCounterPlaced",
		"EventScry",
		"EventSurveil",
		"EventSagaChapter",
		"EventClassLevel",
	} {
		if !arms[name] {
			t.Errorf("projectEvent has no arm for game.%s — one of #1021's six narrations is gone", name)
		}
		if _, listed := silentEventKinds[name]; listed {
			t.Errorf("game.%s is back in silentEventKinds; #1021 decided it is a line", name)
		}
	}
}

// declaredEventKinds parses ../game for every `const X EventKind =
// "..."` and returns constant name → value. Source scanning rather
// than reflection, because Go cannot enumerate the constants of a
// named string type at runtime and a hand-kept list here would be the
// second copy of the thing this file exists to stop drifting.
//
// `ReplacementEventKind` constants are a DIFFERENT type with its own
// gate (internal/game/replacement_kind_gate_test.go) and are excluded
// by the exact type-name match.
func declaredEventKinds(t *testing.T) map[string]game.EventKind {
	t.Helper()
	dir := filepath.Join("..", "game")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	fset := gotoken.NewFileSet()
	out := map[string]game.EventKind{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s/%s: %v", dir, name, err)
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != gotoken.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				ident, ok := vs.Type.(*ast.Ident)
				if !ok || ident.Name != "EventKind" {
					continue
				}
				for i, constName := range vs.Names {
					if i >= len(vs.Values) {
						continue
					}
					lit, ok := vs.Values[i].(*ast.BasicLit)
					if !ok || lit.Kind != gotoken.STRING {
						continue
					}
					value, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatalf("unquote %s = %s: %v", constName.Name, lit.Value, err)
					}
					out[constName.Name] = game.EventKind(value)
				}
			}
		}
	}
	return out
}

// narratedEventKinds parses log.go and returns the set of
// `game.EventXxx` names that appear as a case of projectEvent's
// `switch ev.Kind`. That switch IS the log's coverage, so reading it
// is reading the answer rather than a record of it.
func narratedEventKinds(t *testing.T) map[string]bool {
	t.Helper()
	fset := gotoken.NewFileSet()
	parsed, err := parser.ParseFile(fset, "log.go", nil, 0)
	if err != nil {
		t.Fatalf("parse log.go: %v", err)
	}
	for _, decl := range parsed.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name == nil || fd.Name.Name != "projectEvent" || fd.Body == nil {
			continue
		}
		out := map[string]bool{}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			sw, ok := n.(*ast.SwitchStmt)
			if !ok || !isEventKindTag(sw.Tag) {
				return true
			}
			for _, stmt := range sw.Body.List {
				clause, ok := stmt.(*ast.CaseClause)
				if !ok {
					continue
				}
				for _, expr := range clause.List {
					sel, ok := expr.(*ast.SelectorExpr)
					if !ok {
						continue
					}
					if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "game" {
						out[sel.Sel.Name] = true
					}
				}
			}
			return true
		})
		return out
	}
	t.Fatal("no func projectEvent in log.go")
	return nil
}

// isEventKindTag matches the `ev.Kind` projectEvent's switch keys on.
func isEventKindTag(tag ast.Expr) bool {
	sel, ok := tag.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Kind" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "ev"
}
