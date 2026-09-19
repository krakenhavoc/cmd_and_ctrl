package effects

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exit_payout_guard_test.go is the #911 / #993 half of the lint
// life_continuation_guard_test.go is the #793 / #807 half of: a source
// scan for a mistake `go vet` cannot see and a behaviour test will not
// catch on an ordinary board.
//
// THE SHAPE. An EXIT opens the CR 614 window, so it can be cancelled
// ("cards in graveyards can't be exiled"), redirected (a commander card
// taking CR 903.9's offer) or merely PAUSED on a prompt. The
// fire-and-forget forms — `ExileTarget` with no `Then`,
// `g.ExileCardForEffect`, and the same pair for every other verb —
// return nil for all three, so a card that pays out on the next line
// pays out for an exit that did not happen. That is what Cling to Dust,
// Scavenging Ooze and Deluge of the Dead did: read the card's type into
// a local, exile, then gate the life / counter / Zombie on the local.
//
// WHAT IS FLAGGED is precisely that: a CONDITION, after a
// fire-and-forget exit, that reads a local the function assigned BEFORE
// the exit. A gate on a pre-exit snapshot is a clause ABOUT THE CARD,
// and a clause about the card has to know whether the card actually
// went.
//
// #993 WIDENED IT FROM EXILE TO EVERY EXIT. §5t scoped the original
// scan to exile because that was the verb #911 settled, and named the
// six sites a wider table would find; widened, it finds a SEVENTH —
// Chain of Vapor, which is Boomerang Basics with a different second
// sentence. All seven are answered in ADR 0013 §5v: two moved onto a
// bounce continuation for the ORDER alone (Boomerang Basics, Chain of
// Vapor), one onto a sacrifice continuation for the ANSWER (Ruthless
// Technomancer, the only changed outcome), two were mills whose
// hand-rolled loop is now the primitive's own run (Consuming
// Aberration, Hermit Druid), one was a false positive that stopped
// being one when SacrificeChoice grew a real continuation, and one is
// allowlisted below beside §5t's.
//
// The verbs are the ones with a pausable exit: exile, destroy,
// sacrifice, bounce, tuck, mill, a discard (§5g) and a graveyard
// arrival (§5q). The last two find nothing in the catalog today and are
// in the tables anyway — a guard rail is for the site that has not been
// written yet.
//
// WHAT IS NOT FLAGGED, on purpose:
//
//   - a clause that is not a gate. "Exile target creature. Its
//     controller gains life equal to its power" reads a snapshot too,
//     but it happens either way (ADR 0013 §5m item 5), so a pre-exit
//     local used as a VALUE is fine. Only conditions are read.
//   - `if err != nil`. The error is assigned by the exit, not before
//     it, and `err` is excluded by name as well.
//   - an exit with a `Then`. That is the answer, not the problem.
//   - the PROMPT-driven sacrifices, `g.PlayerSacrificesForEffect` and
//     `g.EachPlayerSacrificesForEffect`. They queue a question and
//     return how many seats were asked; nothing has left the
//     battlefield when they return, and no continuation form exists to
//     point a card at — the sacrifice happens when the player answers,
//     on a prompt that carries no tail. That is a different seam with a
//     different answer (Rise of the Witch-king declares it as a caveat,
//     and #1019 tracks it), and a lint whose message names a fix that
//     does not exist is worse than no lint.
//
// FALSE POSITIVES ARE EXPECTED AND CHEAP, exactly as they are in the
// life guard: add the "<file>:<line>" to the allowlist with the reason
// it is not the bug. An unexplained entry is how a lint stops meaning
// anything, and TestExitPayoutAllowlistIsLive fails on an entry that no
// longer names a finding, so the list cannot quietly outlive the code
// it excuses.
//
// WHY A SECOND SCANNER rather than a third row in the life guard's
// table: the life pattern keys on a SELECTOR field read (`p.Life`)
// anywhere after the call, and this one keys on a LOCAL read inside a
// condition, with the call itself qualified by the absence of a struct
// field. The two have no body in common but the glob.

// exitPayoutAllowlist exempts one "<file>:<line>" finding, with the
// reason it is not the bug.
var exitPayoutAllowlist = map[string]string{
	"solitude.go:74": "Solitude's `if power <= 0` is a guard against gaining zero life, not a gate " +
		"on the exile: \"its controller gains life equal to its power\" is an unconditional clause " +
		"about a player, which ADR 0013 §5m item 5 declared ungated and §5t left ungated.",
	"dark_confidant.go:91": "Dark Confidant's `if life == 0` is Solitude's guard with a different " +
		"verb: \"you lose life equal to its mana value\" is an unconditional clause about a PLAYER, " +
		"so it is not gated on the card reaching a hand, and the zero case returns early only to " +
		"avoid asking the engine for a life change of nothing. The mana value is read before the " +
		"move because that is the last moment the card is guaranteed findable (CR 608.2h), which " +
		"is the ungated half of the line ADR 0013 §5t draws.",
}

// exitVerb names one pausable exit and the continuation form a card
// that reads its outcome must use. The message a finding prints is
// built from this, so a lint that fires always names the fix.
type exitVerb struct {
	kind string
	fix  string
}

var (
	exileVerb = exitVerb{"exile", "ExileTarget.Then, ExileThenIfItWas (the \"if it was a <type> card\" family) " +
		"or g.ExileCardThenForEffect / g.ExileCardsThenForEffect, and gate the clause on `exiled`"}
	destroyVerb = exitVerb{"destroy", "DestroyAllMatching.Then or g.DestroyPermanentsThenForEffect " +
		"(a set of one is fine), and gate the clause on the `destroyed` list"}
	sacrificeVerb = exitVerb{"sacrifice", "SacrificePermanent.Then or g.SacrificeThenForEffect / " +
		"g.SacrificeAllThenForEffect, and gate the clause on `sacrificed`"}
	bounceVerb = exitVerb{"bounce", "BounceToHand.Then, ReturnAllToHand.Then or " +
		"g.BounceToHandThenForEffect / g.BounceCardsToHandThenForEffect, and gate the clause on `bounced`"}
	tuckVerb = exitVerb{"tuck", "g.TuckToLibraryThenForEffect / g.TuckCardsToLibraryThenForEffect, " +
		"and gate the clause on `tucked`"}
	millVerb = exitVerb{"mill", "MillToZone.Then or g.MillToZoneThenForEffect, and gate the clause " +
		"on the `milled` list"}
	discardVerb = exitVerb{"discard", "there is no continuation form for a discard yet — a discard is " +
		"an exit (ADR 0013 §5g) and DiscardRandomForEffect returns before a paused leg lands, so " +
		"write the wrapper beside g.ExileCardThenForEffect rather than reading back on the next line"}
	graveyardVerb = exitVerb{"graveyard arrival", "there is no continuation form for a graveyard " +
		"arrival yet — g.PutIntoGraveyardForEffect rides routeCardToZoneLocked and returns before a " +
		"paused leg lands, so write the wrapper beside g.ExileCardThenForEffect " +
		"(routeAllThenLocked(zoneRoute{Dst: ZoneGraveyard}, …) is the body) rather than reading back " +
		"on the next line"}
)

// exitStartsTheClock names the fire-and-forget engine calls. The `Then`
// forms are deliberately absent — they are the answer.
var exitStartsTheClock = map[string]exitVerb{
	"ExileCardForEffect":               exileVerb,
	"ExileCardsForEffect":              exileVerb,
	"ExileCardWithPermissionForEffect": exileVerb,
	"ExileTopWithPermissionForEffect":  exileVerb,
	"ExileTopFaceDownForEffect":        exileVerb,
	"DestroyPermanentForEffect":        destroyVerb,
	"DestroyPermanentsForEffect":       destroyVerb,
	"SacrificePermanentForEffect":      sacrificeVerb,
	"SacrificeAllForEffect":            sacrificeVerb,
	"BounceToHandForEffect":            bounceVerb,
	"BounceCardsToHandForEffect":       bounceVerb,
	"TuckToLibraryForEffect":           tuckVerb,
	"TuckToLibraryAtDepthForEffect":    tuckVerb,
	"MillNForEffect":                   millVerb,
	"MillToZoneForEffect":              millVerb,
	"DiscardRandomForEffect":           discardVerb,
	"PutIntoGraveyardForEffect":        graveyardVerb,
}

// exitPrimitives are the catalog structs whose `.Apply(ctx)` reaches
// one of those, but ONLY when the literal declares no `Then`.
var exitPrimitives = map[string]exitVerb{
	"ExileTarget":            exileVerb,
	"ExileAllMatching":       exileVerb,
	"ExileWithPermission":    exileVerb,
	"ExileTopWithPermission": exileVerb,
	"ExileTopFaceDown":       exileVerb,
	"DestroyTarget":          destroyVerb,
	"DestroyAllMatching":     destroyVerb,
	"SacrificePermanent":     sacrificeVerb,
	"BounceToHand":           bounceVerb,
	"BounceAllMatching":      bounceVerb,
	"ReturnAllToHand":        bounceVerb,
	"MillCards":              millVerb,
	"MillToZone":             millVerb,
	"DiscardCards":           discardVerb,
}

func TestNoExitPayoutWithoutAContinuation(t *testing.T) {
	findings, scanned := scanCatalogForExitPayouts(t, exitPayoutAllowlist)
	if scanned == 0 {
		t.Fatal("scanned no catalog files — the glob is wrong, and a lint that reads nothing passes everything")
	}
	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("a clause is gated on a value read BEFORE a fire-and-forget exit, so it pays out for "+
		"a move the CR 614 window cancelled, redirected or has not answered yet (#911, #993, "+
		"ADR 0013 §5t / §5v).\nIf this condition is fine, add it to exitPayoutAllowlist with the "+
		"reason.\n  %s", strings.Join(findings, "\n  "))
}

// TestExitPayoutAllowlistIsLive keeps the allowlist honest in both
// directions. An entry with no reason is an exemption nobody can audit,
// and an entry that no longer names a finding is a line that outlived
// the code it excused — the next author reads it as a rule about a site
// that has already been fixed, or moved, or deleted.
//
// #1003's TestEveryReplacementEventKindIsSwitchedOn is the standard:
// the list is enforced, not promised.
func TestExitPayoutAllowlistIsLive(t *testing.T) {
	unfiltered, _ := scanCatalogForExitPayouts(t, nil)
	live := map[string]bool{}
	for _, f := range unfiltered {
		live[strings.SplitN(f, " ", 2)[0]] = true
	}
	for where, reason := range exitPayoutAllowlist {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("exitPayoutAllowlist[%q] has no reason — an unexplained entry is how a lint "+
				"stops meaning anything", where)
		}
		if !live[where] {
			t.Errorf("exitPayoutAllowlist[%q] excuses nothing: the scanner reports no finding "+
				"there. The site was fixed, renamed or moved — delete the entry, or move it to the "+
				"line the finding is on now.", where)
		}
	}
}

// scanCatalogForExitPayouts parses every non-test file in the catalog
// package and reports the findings, plus how many files it read.
// `allow` is consulted per site; pass nil to see every finding.
func scanCatalogForExitPayouts(t *testing.T, allow map[string]string) ([]string, int) {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	seen := map[string]bool{}
	var findings []string
	scanned := 0

	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		file, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		scanned++
		for _, f := range exitPayoutsIn(fset, name, file, allow) {
			// A nested func literal is walked both on its own and as
			// part of its enclosing declaration, so one mistake can be
			// reported twice. It is one mistake.
			if seen[f] {
				continue
			}
			seen[f] = true
			findings = append(findings, f)
		}
	}
	return findings, scanned
}

// exitPayoutsIn reports every gate-on-a-pre-exit-local inside file.
// It walks function declarations and function literals alike, because
// most of the catalog's bodies are literals on a Spec.
func exitPayoutsIn(fset *token.FileSet, name string, file *ast.File, allow map[string]string) []string {
	var out []string
	ast.Inspect(file, func(n ast.Node) bool {
		var body *ast.BlockStmt
		switch fn := n.(type) {
		case *ast.FuncDecl:
			body = fn.Body
		case *ast.FuncLit:
			body = fn.Body
		default:
			return true
		}
		if body == nil {
			return true
		}
		out = append(out, exitPayoutsInBody(fset, name, body, allow)...)
		return true
	})
	return out
}

// exitCall is one fire-and-forget exit found in a body, with the verb
// it took — the verb is what lets a finding name the continuation.
type exitCall struct {
	call *ast.CallExpr
	verb exitVerb
}

func exitPayoutsInBody(fset *token.FileSet, name string, body *ast.BlockStmt, allow map[string]string) []string {
	var exits []exitCall
	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if verb, ok := fireAndForgetExit(call); ok {
				exits = append(exits, exitCall{call, verb})
			}
		}
		return true
	})
	if len(exits) == 0 {
		return nil
	}
	// The locals this function assigned before its FIRST exit. A
	// later one is covered too: anything assigned before the first is
	// assigned before all of them.
	pre := map[string]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || as.End() > exits[0].call.Pos() {
			return true
		}
		for _, lhs := range as.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" && id.Name != "err" {
				pre[id.Name] = true
			}
		}
		return true
	})
	if len(pre) == 0 {
		return nil
	}

	var out []string
	ast.Inspect(body, func(n ast.Node) bool {
		var cond ast.Expr
		switch s := n.(type) {
		case *ast.IfStmt:
			cond = s.Cond
		case *ast.SwitchStmt:
			cond = s.Tag
		default:
			return true
		}
		if cond == nil {
			return true
		}
		for _, e := range exits {
			if n.Pos() <= e.call.End() {
				continue
			}
			read := ""
			ast.Inspect(cond, func(k ast.Node) bool {
				if id, ok := k.(*ast.Ident); ok && pre[id.Name] {
					read = id.Name
				}
				return true
			})
			if read == "" {
				return true
			}
			where := name + ":" + strconv.Itoa(fset.Position(n.Pos()).Line)
			if _, ok := allow[where]; ok {
				return true
			}
			out = append(out, where+" gates on `"+read+"`, read before the "+e.verb.kind+" at "+
				name+":"+strconv.Itoa(fset.Position(e.call.Pos()).Line)+" — use "+e.verb.fix)
			return true
		}
		return true
	})
	return out
}

// fireAndForgetExit reports whether call is an exit that cannot tell
// its caller what landed — a direct `g.ExileCardForEffect(…)` and its
// siblings, or a `Primitive{…}.Apply(ctx)` whose literal declares no
// `Then` — and which verb it is.
func fireAndForgetExit(call *ast.CallExpr) (exitVerb, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return exitVerb{}, false
	}
	if verb, ok := exitStartsTheClock[sel.Sel.Name]; ok {
		return verb, true
	}
	if sel.Sel.Name != "Apply" {
		return exitVerb{}, false
	}
	// Both spellings: `ExileTarget{…}.Apply(ctx)` and the
	// parenthesised form gofmt insists on when the literal is the
	// whole statement.
	recv := sel.X
	for {
		paren, ok := recv.(*ast.ParenExpr)
		if !ok {
			break
		}
		recv = paren.X
	}
	lit, ok := recv.(*ast.CompositeLit)
	if !ok {
		return exitVerb{}, false
	}
	ident, ok := lit.Type.(*ast.Ident)
	if !ok {
		return exitVerb{}, false
	}
	verb, ok := exitPrimitives[ident.Name]
	if !ok {
		return exitVerb{}, false
	}
	for _, e := range lit.Elts {
		kv, ok := e.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if k, ok := kv.Key.(*ast.Ident); ok && k.Name == "Then" {
			return exitVerb{}, false
		}
	}
	return verb, true
}

// TestExitPayoutGuardCatchesTheShape proves the lint is a lint and not
// a function that returns nil. A catalog with no findings is the
// expected state, which means the scan could rot into a no-op — a wrong
// glob, a renamed entry point, a `Then` check that stopped matching —
// and nothing would say so.
//
// One synthetic offender PER VERB, because #993's widening is a table
// and a table is exactly the thing that rots one row at a time: an
// entry point renamed in the engine leaves its verb silently unguarded,
// and only a fixture that fires through that verb says so. Each is
// paired with the continuation the failure message names, which must
// stay quiet, and the two shapes that were never the bug round it out.
func TestExitPayoutGuardCatchesTheShape(t *testing.T) {
	const before = `package effects

func ooze(ctx *Context, id, source uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	wasCreature := c.IsCreature()
	if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if !wasCreature {
		return nil
	}
	return AddCounter{Target: source, Kind: "+1/+1", N: 1}.Apply(ctx)
}
`
	const after = `package effects

func ooze(ctx *Context, id, source uuid.UUID) error {
	return ExileThenIfItWas{
		Target: id,
		Was:    WasCreatureCard,
		Then: func(ctx *Context) error {
			return AddCounter{Target: source, Kind: "+1/+1", N: 1}.Apply(ctx)
		},
	}.Apply(ctx)
}
`
	const throughTheEngineCall = `package effects

func ooze(ctx *Context, id uuid.UUID) error {
	c, _ := ctx.Game.LookupCardForEffect(id)
	wasCreature := c.IsCreature()
	if err := ctx.Game.ExileCardForEffect(id); err != nil {
		return err
	}
	if wasCreature {
		return GainLife{Amount: 1}.Apply(ctx)
	}
	return nil
}
`
	const unconditionalClause = `package effects

func plowshares(ctx *Context, id uuid.UUID) error {
	card, _ := ctx.Game.LookupCardForEffect(id)
	power := card.CurrentPower()
	if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: card.Controller, Amount: power}.Apply(ctx)
}
`
	// --- one offender and one answer per widened verb -------------

	const destroyOffender = `package effects

func wrath(ctx *Context, id uuid.UUID) error {
	c, _ := ctx.Game.LookupCardForEffect(id)
	wasArtifact := c.IsArtifact()
	if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if wasArtifact {
		return DrawCards{N: 1}.Apply(ctx)
	}
	return nil
}
`
	const destroyAnswer = `package effects

func wrath(ctx *Context, id uuid.UUID) error {
	return ctx.Game.DestroyPermanentsThenForEffect([]uuid.UUID{id}, func(g *game.Game, destroyed []uuid.UUID) error {
		if len(destroyed) == 0 {
			return nil
		}
		return DrawCards{N: 1}.Apply(ctx)
	})
}
`
	const sacrificeOffender = `package effects

func technomancer(ctx *Context, id uuid.UUID) error {
	victim, _ := ctx.Game.LookupCardForEffect(id)
	power := victim.CurrentPower()
	if err := (SacrificePermanent{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if power <= 0 {
		return nil
	}
	return CreateToken{Template: TreasureToken(), N: power}.Apply(ctx)
}
`
	const sacrificeAnswer = `package effects

func technomancer(ctx *Context, id uuid.UUID) error {
	victim, _ := ctx.Game.LookupCardForEffect(id)
	power := victim.CurrentPower()
	return SacrificePermanent{
		Target: id,
		Then: func(ctx *Context, sacrificed bool) error {
			if !sacrificed || power <= 0 {
				return nil
			}
			return CreateToken{Template: TreasureToken(), N: power}.Apply(ctx)
		},
	}.Apply(ctx)
}
`
	const bounceOffender = `package effects

func boomerang(ctx *Context, id uuid.UUID) error {
	yours, ok := controllerOfTarget(ctx, id)
	if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if !ok || yours != ctx.Controller() {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
}
`
	const bounceAnswer = `package effects

func boomerang(ctx *Context, id uuid.UUID) error {
	yours, ok := controllerOfTarget(ctx, id)
	drawer := ctx.Controller()
	return BounceToHand{
		Target: id,
		Then: func(ctx *Context, _ bool) error {
			if !ok || yours != drawer {
				return nil
			}
			return DrawCards{Player: drawer, N: 1}.Apply(ctx)
		},
	}.Apply(ctx)
}
`
	const tuckOffender = `package effects

func chaosWarp(ctx *Context, id uuid.UUID) error {
	c, _ := ctx.Game.LookupCardForEffect(id)
	wasPermanent := c.IsPermanent()
	if err := ctx.Game.TuckToLibraryForEffect(id, false); err != nil {
		return err
	}
	if wasPermanent {
		return DrawCards{N: 1}.Apply(ctx)
	}
	return nil
}
`
	const tuckAnswer = `package effects

func chaosWarp(ctx *Context, id uuid.UUID) error {
	c, _ := ctx.Game.LookupCardForEffect(id)
	wasPermanent := c.IsPermanent()
	return ctx.Game.TuckToLibraryThenForEffect(id, game.TuckOptions{}, func(g *game.Game, tucked bool) error {
		if !tucked || !wasPermanent {
			return nil
		}
		return DrawCards{N: 1}.Apply(ctx)
	})
}
`
	const millOffender = `package effects

func untilALand(ctx *Context, player uuid.UUID) error {
	p := ctx.PlayerByID(player)
	top := p.Library.Cards[len(p.Library.Cards)-1]
	if err := (MillCards{Player: player, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if top.IsLand() {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
}
`
	const millAnswer = `package effects

func untilALand(ctx *Context, player uuid.UUID) error {
	return MillToZone{
		Player: player,
		Until:  func(c game.Card) bool { return c.IsLand() },
		Then: func(ctx *Context, milled []uuid.UUID) error {
			return DrawCards{N: len(milled)}.Apply(ctx)
		},
	}.Apply(ctx)
}
`
	const graveyardOffender = `package effects

func genesisWave(ctx *Context, id uuid.UUID) error {
	c, _ := ctx.Game.LookupCardForEffect(id)
	wasCreature := c.IsCreature()
	if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
		return err
	}
	if wasCreature {
		return DrawCards{N: 1}.Apply(ctx)
	}
	return nil
}
`
	const discardOffender = `package effects

func rummage(ctx *Context, player uuid.UUID) error {
	p := ctx.PlayerByID(player)
	held := p.Hand.Size()
	if err := (DiscardCards{Player: player, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if held > 0 {
		return DrawCards{N: 1}.Apply(ctx)
	}
	return nil
}
`
	for _, tc := range []struct {
		name string
		src  string
		want int
	}{
		{"the pre-#911 read-back", before, 1},
		{"the same mistake through the engine call", throughTheEngineCall, 1},
		{"the ExileThenIfItWas continuation", after, 0},
		{"an unconditional clause that only READS a snapshot", unconditionalClause, 0},

		{"destroy: a gate on a pre-destroy local", destroyOffender, 1},
		{"destroy: DestroyPermanentsThenForEffect", destroyAnswer, 0},
		{"sacrifice: a gate on a pre-sacrifice local", sacrificeOffender, 1},
		{"sacrifice: SacrificePermanent.Then", sacrificeAnswer, 0},
		{"bounce: a gate on a pre-bounce local", bounceOffender, 1},
		{"bounce: BounceToHand.Then", bounceAnswer, 0},
		{"tuck: a gate on a pre-tuck local", tuckOffender, 1},
		{"tuck: TuckToLibraryThenForEffect", tuckAnswer, 0},
		{"mill: a gate on a pre-mill local", millOffender, 1},
		{"mill: MillToZone.Then", millAnswer, 0},
		{"a graveyard arrival: a gate on a pre-move local", graveyardOffender, 1},
		{"a discard: a gate on a pre-discard local", discardOffender, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", tc.src, 0)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			got := exitPayoutsIn(fset, "fixture.go", file, nil)
			if len(got) != tc.want {
				t.Errorf("findings = %d, want %d — the lint has stopped seeing the shape it exists for.\n  %s",
					len(got), tc.want, strings.Join(got, "\n  "))
			}
		})
	}
}

// TestEveryExitEntryPointStillExists is the other half of the drift
// guard, and the one the header comment promised before this test
// existed: the tables are STRINGS, so an entry point renamed in game/
// or a primitive renamed in this package leaves a dead row that matches
// nothing and guards nothing, and every test here keeps passing.
//
// The self-test's per-verb fixtures catch a verb that has lost ALL its
// entry points. This catches the one row of eight that went stale,
// which is the shape a rename actually takes.
func TestEveryExitEntryPointStillExists(t *testing.T) {
	gameType := reflect.TypeOf(&game.Game{})
	for name, verb := range exitStartsTheClock {
		if _, ok := gameType.MethodByName(name); !ok {
			t.Errorf("exitStartsTheClock names (*game.Game).%s for the %q exit, and no such method "+
				"exists. It was renamed or removed, so that row now matches nothing — point it at "+
				"the new name, or delete it if the exit is gone.", name, verb.kind)
		}
	}

	declared := declaredTypesInCatalog(t)
	for name, verb := range exitPrimitives {
		if !declared[name] {
			t.Errorf("exitPrimitives names the catalog type %s for the %q exit, and no such type is "+
				"declared in this package. It was renamed or removed, so that row now matches "+
				"nothing — point it at the new name, or delete it if the primitive is gone.",
				name, verb.kind)
		}
	}
}

// declaredTypesInCatalog is the set of type names this package declares.
func declaredTypesInCatalog(t *testing.T) map[string]bool {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	out := map[string]bool{}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					out[ts.Name.Name] = true
				}
			}
		}
	}
	if len(out) == 0 {
		t.Fatal("found no type declarations in the catalog — the glob is wrong, and a check that " +
			"reads nothing passes everything")
	}
	return out
}

// TestEveryExitVerbIsReachableFromTheTables is the widening's own
// drift guard. A verb declared above but named by no entry point is a
// row that guards nothing — the shape #1003 calls "enforced, not
// promised" — and it is exactly what a rename in game/ leaves behind.
func TestEveryExitVerbIsReachableFromTheTables(t *testing.T) {
	declared := map[string]bool{}
	for _, v := range []exitVerb{
		exileVerb, destroyVerb, sacrificeVerb, bounceVerb,
		tuckVerb, millVerb, discardVerb, graveyardVerb,
	} {
		if strings.TrimSpace(v.fix) == "" {
			t.Errorf("the %q verb names no fix — a lint that fires without naming the continuation "+
				"sends the next author looking for one", v.kind)
		}
		declared[v.kind] = true
	}
	reached := map[string]bool{}
	for _, v := range exitStartsTheClock {
		reached[v.kind] = true
	}
	for _, v := range exitPrimitives {
		reached[v.kind] = true
	}
	for kind := range declared {
		if !reached[kind] {
			t.Errorf("the %q verb is in no table, so nothing can ever flag it — an entry point was "+
				"renamed, or the row was never wired up", kind)
		}
	}
}
