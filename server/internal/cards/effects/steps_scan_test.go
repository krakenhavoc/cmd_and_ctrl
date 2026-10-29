package effects

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// steps_scan_test.go — the scanner behind #4155: which hand-written
// resolution bodies are actually wrong under CR 608.2c.
//
// A body is wrong when an EARLIER statement can open a prompt in the
// middle of resolution (a scry or surveil, a search, a choose-cards, a
// CR 616 ordering on damage, counters, tokens, life, a draw or a mill, a
// damage prevention's follow-up question) and a LATER statement of the
// same sequence has an observable effect: the later statement runs
// before the answer. A body whose earlier statements cannot ask does
// not matter, and neither does one that is already ResolveSteps,
// ModeSteps or a row's own Do (those park).
//
// It is a source scan, in the family of the read-back guard
// (life_continuation_guard_test.go), and it asserts nothing: it writes a
// ranked list, to be read by a person. Run it from server/:
//
//	CMDCTRL_SCRYFALL_DUMP=…/default-cards.json \
//	  go test ./internal/cards/effects/ -run TestScanStepsBodies -scan-steps=/path/out.md
//
// Without -scan-steps the test skips. The dump is only for the EDHREC
// rank that orders the list (popular cards first); without it every
// rank is 0 and the list is ordered by name.
//
// WHAT COUNTS. The sets below are data. A call "asks" if it is an
// effect primitive's Apply or a game method (…ForEffect) in the
// asking list; a call "acts" if it asks or is any other effect
// primitive or mutating game method. A same-package helper function is
// as asking or acting as its own body, followed transitively, so
// `return damageToEachOpponent(g, item, 2)` is read through. A nested
// func literal (a Then continuation, a cleanup) is not part of its
// statement: it runs after its own prompt.
//
// FALSE POSITIVES ARE EXPECTED. A first statement that asks only on a
// board with two replacement effects is on the list: that is the
// bug (Doubling Season and Hardened Scales, then the next statement),
// and it is why the list is long. "shares" names the variables an
// earlier statement declares that a later one reads; a piece cannot
// pass those through a Go variable (the doc in docs/adding-cards.md),
// so the card needs them re-read from the game before it converts.

var scanStepsOut = flag.String("scan-steps", "",
	"write the #4155 scan (hand-written bodies whose later statement can outrun an earlier prompt) to this file")

// stepsAskApply are effect primitives whose Apply can open a prompt.
var stepsAskApply = setOf(
	"DealDamage", "DamageController", "DamageEachOpponent", "DrawCards", "GainLife", "GainLifeEqualToSubject",
	"EachOpponentLosesLife", "TargetPlayerLosesLife", "PayLifeAmount", "AddCounter",
	"CreateToken", "CreateTokenAdvanced", "CreateTokenCopy", "CreateTokensForEach", "CreateXByXToken",
	"CreateRoleToken", "CreateTokensAttackingYourChoice", "MillCards", "MillToZone", "Scry", "Surveil",
	"SearchLibrary", "TakeFromLibraryToHand", "LookAtTop", "LookAtLibraryThenPlace", "Amass", "Bolster",
	"Incubate", "Populate", "Proliferate", "Explores", "ConniveX", "ConniveEach", "Discover", "DiscardCards",
	"ChooseNumber", "ChoosePlayer", "ChoosePermanents", "PickOption", "PileSplit", "MayChoice", "MayPay",
	"MayPayX", "MayPayEnergy", "SacrificeChoice", "EachPlayerSacrifices", "Vote", "RevealPick",
	"ChooseFromRevealedHand", "PutFromHandOntoBattlefield", "PutFromLibraryOntoBattlefield",
	"PutInLibraryInAnyOrder", "PutFromHandOnTopInAnyOrder", "Recruit", "Manifest", "ManifestDread", "ManifestDreadTimes",
	"OfferGraveyardCast", "CopyAndCast", "Earthbend", "RevealUntilThenPutOntoBattlefield", "PayUnless",
	"AnyPlayerPays", "AnyPlayerPaysForEach", "ReflexiveTrigger", "FaceVillainousChoice", "VentureIntoTheDungeon",
	"EachPlayerTakesFromLibrary", "TakeFromOutside", "TakeFromExileOrOutside", "PutIntoLibraryAtDepthOrBottom",
)

// stepsAskGame matches game methods that can open a prompt.
var stepsAskGame = regexp.MustCompile(`^(Queue[A-Za-z]*|Scry[A-Za-z]*|Surveil[A-Za-z]*|SearchLibrary[A-Za-z]*|` +
	`Discard[A-Za-z]*|EachPlayerDiscards[A-Za-z]*|PlayerDiscards[A-Za-z]*|PlayersDiscard[A-Za-z]*|` +
	`Draw[A-Za-z]*|Mill[A-Za-z]*|ChangePlayerLife[A-Za-z]*|ChangeLifeAfter[A-Za-z]*|PayLife[A-Za-z]*|` +
	`DealDamage[A-Za-z]*|AddCounter[A-Za-z]*|AddPlayerCounter[A-Za-z]*|CreateToken[A-Za-z]*|Amass[A-Za-z]*|` +
	`Explore[A-Za-z]*|Proliferate[A-Za-z]*|Choose[A-Za-z]*|RevealPick[A-Za-z]*|RunVote[A-Za-z]*|` +
	`StartVote[A-Za-z]*|RunOptionPick[A-Za-z]*|Sacrifice[A-Za-z]*|EachPlayerSacrifices[A-Za-z]*|` +
	`PlayerSacrifices[A-Za-z]*|PlayersSacrifice[A-Za-z]*|Venture[A-Za-z]*|Discover[A-Za-z]*|Cascade[A-Za-z]*|` +
	`Manifest[A-Za-z]*|LoseLifeEach[A-Za-z]*|Offer[A-Za-z]*|FreeCast[A-Za-z]*|Put(Cards|From)[A-Za-z]*|` +
	`TakeFromLibrary[A-Za-z]*|LookAtTop[A-Za-z]*|Earthbend[A-Za-z]*|Awaken[A-Za-z]*|PermanentsPicked[A-Za-z]*|` +
	`ChooseCards[A-Za-z]*|PileSplit[A-Za-z]*|Reveal(Cards|Top)[A-Za-z]*|Suspect[A-Za-z]*|Vote[A-Za-z]*|` +
	`Teamwork[A-Za-z]*|PreventNextDamage[A-Za-z]*|CopyCardsAndOffer[A-Za-z]*)ForEffect$`)

// stepsActGame matches the other game methods with an observable effect.
var stepsActGame = regexp.MustCompile(`^(Destroy|Exile|Return|Tuck|Tap|Untap|Attach|Unattach|Gain|Grant|Register|` +
	`Regenerate|Transform|PhaseOut|Counter|Prevent|Redirect|Set|Skip|Become|Schedule|Goad|Flip|Roll|Shuffle|` +
	`WinTheGame|LoseTheGame|Monstrosity|Exchange|Copy|Remove|Take|AddMana|Lose|Heal|Turn|Bounce|Fight|Convert|` +
	`Harness|Meld|Mark|Lock|Unlock|Solve|Taunt|Swap|Ring|Day|Night)[A-Za-z]*ForEffect$`)

func setOf(names ...string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	return m
}

// stepsFinding is one wrong-looking body.
type stepsFinding struct {
	Rank   int
	Name   string
	Oracle string
	File   string
	Line   int
	Site   string // OnResolve, Effect, ModeDoing, Then, …
	Flat   bool   // the pair is in the body's own top-level sequence
	Ask    int    // line of the earlier statement that can ask
	Why    string // what in it asks
	Later  int    // line of the later statement with an effect
	Shares []string
}

// bodyFacts are the (asks, acts) of one piece of source.
type bodyFacts struct {
	asks, acts bool
	why        string
}

func (a *bodyFacts) merge(b bodyFacts) {
	if b.asks && !a.asks {
		a.why = b.why
	}
	a.asks = a.asks || b.asks
	a.acts = a.acts || b.acts
}

// stepsScanner carries the package's helper functions, so a call into a
// helper is read through.
type stepsScanner struct {
	helpers map[string]*ast.FuncDecl
	memo    map[string]*bodyFacts
	busy    map[string]bool
}

// factsOf is the facts of a node, skipping nested func literals.
func (s *stepsScanner) factsOf(n ast.Node) bodyFacts {
	var f bodyFacts
	ast.Inspect(n, func(x ast.Node) bool {
		switch x := x.(type) {
		case *ast.FuncLit:
			return false
		case *ast.CallExpr:
			f.merge(s.callFacts(x))
		}
		return true
	})
	return f
}

func (s *stepsScanner) callFacts(c *ast.CallExpr) bodyFacts {
	switch fn := c.Fun.(type) {
	case *ast.SelectorExpr:
		name := fn.Sel.Name
		if name == "Apply" {
			recv := receiverTypeName(fn.X)
			// A Do/ResolveSteps value or a step built elsewhere: unknown,
			// an act that does not ask.
			if stepsAskApply[recv] {
				return bodyFacts{asks: true, acts: true, why: recv + ".Apply"}
			}
			return bodyFacts{acts: true}
		}
		if stepsAskGame.MatchString(name) {
			return bodyFacts{asks: true, acts: true, why: name}
		}
		if stepsActGame.MatchString(name) {
			return bodyFacts{acts: true}
		}
		// A same-package helper called as a method on the context is
		// not resolvable; a plain function is below.
	case *ast.Ident:
		if h := s.helperFacts(fn.Name); h != nil {
			return *h
		}
	}
	return bodyFacts{}
}

// receiverTypeName is the type name of `T{…}` or `(T{…})`, or "".
func receiverTypeName(x ast.Expr) string {
	for {
		switch e := x.(type) {
		case *ast.ParenExpr:
			x = e.X
		case *ast.CompositeLit:
			switch t := e.Type.(type) {
			case *ast.Ident:
				return t.Name
			case *ast.SelectorExpr:
				return t.Sel.Name
			}
			return ""
		default:
			return ""
		}
	}
}

func (s *stepsScanner) helperFacts(name string) *bodyFacts {
	if f, ok := s.memo[name]; ok {
		return f
	}
	d := s.helpers[name]
	if d == nil || d.Body == nil || s.busy[name] {
		return nil
	}
	s.busy[name] = true
	f := s.factsOf(d.Body)
	delete(s.busy, name)
	if f.asks {
		f.why = name + "→" + f.why
	}
	s.memo[name] = &f
	return &f
}

// scanStepsBodies is the scan itself: every hand-written resolution body in
// the catalog whose earlier statement can ask and whose later statement has
// an effect, and the name of every row the scan read. ranks (EDHREC, by
// oracle id and lower-cased name) may be nil.
func scanStepsBodies(t *testing.T, ranks map[string]int) ([]stepsFinding, map[string]bool) {
	t.Helper()
	fset := token.NewFileSet()
	stepsPos = fset
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	parsed := map[string]*ast.File{}
	sc := &stepsScanner{helpers: map[string]*ast.FuncDecl{}, memo: map[string]*bodyFacts{}, busy: map[string]bool{}}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		parsed[name] = f
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil {
				sc.helpers[fd.Name.Name] = fd
			}
		}
	}
	var out []stepsFinding
	seen := map[string]bool{}
	names := map[string]bool{}
	for _, name := range files {
		f := parsed[name]
		if f == nil || !fileRegisters(f) {
			continue
		}
		specs := specsIn(f, fset)
		for _, sp := range specs {
			names[sp.name] = true
		}
		visit := func(site string, body *ast.BlockStmt, pos token.Pos) {
			for _, fd := range scanSequence(sc, body, true) {
				fd.File, fd.Site = name, site
				fd.Line = fset.Position(pos).Line
				card := cardFor(specs, pos)
				fd.Name, fd.Oracle = card.name, card.oracle
				fd.Rank = ranks[card.oracle]
				if fd.Rank == 0 {
					fd.Rank = ranks[strings.ToLower(card.name)]
				}
				key := fmt.Sprintf("%s:%d:%d", name, fd.Line, fd.Ask)
				if !seen[key] {
					seen[key] = true
					out = append(out, fd)
				}
			}
		}
		var stack []ast.Node
		ast.Inspect(f, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return true
			}
			stack = append(stack, n)
			switch n := n.(type) {
			case *ast.FuncLit:
				if returnsError(n.Type) {
					visit(siteOf(stack), n.Body, n.Pos())
				}
			}
			return true
		})
	}
	return out, names
}

func TestScanStepsBodies(t *testing.T) {
	if *scanStepsOut == "" {
		t.Skip("set -scan-steps=<file> to write the #4155 scan")
	}
	out, _ := scanStepsBodies(t, loadEDHRecRanks(t))
	sort.Slice(out, func(i, j int) bool {
		ri, rj := rankKey(out[i].Rank), rankKey(out[j].Rank)
		if ri != rj {
			return ri < rj
		}
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Line < out[j].Line
	})
	var b strings.Builder
	fmt.Fprintf(&b, "# #4155 scan: %d bodies where a later statement can outrun an earlier prompt\n\n", len(out))
	cards := map[string]bool{}
	for _, f := range out {
		cards[f.Oracle+f.Name] = true
	}
	fmt.Fprintf(&b, "%d cards. Ordered by EDHREC rank (0 = unranked, last). `flat` = the pair is in the body's own sequence.\n\n", len(cards))
	b.WriteString("| EDHREC | Card | Site | Where | Earlier (asks) | Later | Shares | Skipped because |\n|---|---|---|---|---|---|---|---|\n")
	for _, f := range out {
		shape := "flat"
		if !f.Flat {
			shape = "nested"
		}
		fmt.Fprintf(&b, "| %d | %s | %s %s | %s:%d | L%d %s | L%d | %s | %s |\n", f.Rank, f.Name, f.Site, shape,
			f.File, f.Line, f.Ask, f.Why, f.Later, strings.Join(f.Shares, ","), stepsSkipped[f.Name])
	}
	if err := os.WriteFile(*scanStepsOut, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d findings over %d cards -> %s", len(out), len(cards), *scanStepsOut)
}

func rankKey(r int) int {
	if r == 0 {
		return 1 << 30
	}
	return r
}

func returnsError(ft *ast.FuncType) bool {
	if ft.Results == nil || len(ft.Results.List) != 1 {
		return false
	}
	id, ok := ft.Results.List[0].Type.(*ast.Ident)
	return ok && id.Name == "error"
}

// fileRegisters reports whether the file declares a catalog row.
func fileRegisters(f *ast.File) bool {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "Register" {
				found = true
			}
		}
		return !found
	})
	return found
}

// siteOf names where a func literal sits: the field it initialises or
// the constructor it is passed to.
func siteOf(stack []ast.Node) string {
	if len(stack) < 2 {
		return "?"
	}
	switch p := stack[len(stack)-2].(type) {
	case *ast.KeyValueExpr:
		if id, ok := p.Key.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.CallExpr:
		switch fn := p.Fun.(type) {
		case *ast.Ident:
			return fn.Name
		case *ast.SelectorExpr:
			return fn.Sel.Name
		}
	case *ast.ReturnStmt:
		return "return"
	case *ast.AssignStmt:
		return "assigned"
	}
	return "literal"
}

type specAt struct {
	pos          token.Pos
	oracle, name string
}

func specsIn(f *ast.File, fset *token.FileSet) []specAt {
	var out []specAt
	// A row may name itself by a string constant of the file (`const name =
	// "Summon: Ixion"`, then `Name: name`).
	consts := map[string]string{}
	ast.Inspect(f, func(n ast.Node) bool {
		if vs, ok := n.(*ast.ValueSpec); ok && len(vs.Names) == len(vs.Values) {
			for i, id := range vs.Names {
				if lit, ok := vs.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if v, err := strconv.Unquote(lit.Value); err == nil {
						consts[id.Name] = v
					}
				}
			}
		}
		return true
	})
	ast.Inspect(f, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		var s specAt
		for _, e := range cl.Elts {
			kv, ok := e.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, _ := kv.Key.(*ast.Ident)
			if key == nil {
				continue
			}
			var v string
			switch val := kv.Value.(type) {
			case *ast.BasicLit:
				if val.Kind != token.STRING {
					continue
				}
				v, _ = strconv.Unquote(val.Value)
			case *ast.Ident:
				c, ok := consts[val.Name]
				if !ok || key.Name != "Name" {
					continue
				}
				v = c
			default:
				continue
			}
			switch key.Name {
			case "OracleID":
				s.oracle = v
			case "Name":
				s.name = v
			}
		}
		if s.oracle != "" || s.name != "" {
			s.pos = cl.Pos()
			out = append(out, s)
		}
		return true
	})
	return out
}

// cardFor picks the row a body belongs to: the nearest one declared
// before it in the file, or the first.
func cardFor(specs []specAt, pos token.Pos) specAt {
	if len(specs) == 0 {
		return specAt{}
	}
	best := specs[0]
	for _, s := range specs {
		if s.pos <= pos {
			best = s
		}
	}
	return best
}

// scanSequence finds the pairs in a block's own statement list, then
// descends into the blocks inside it (a nested sequence is reported
// non-flat).
func scanSequence(sc *stepsScanner, body *ast.BlockStmt, flat bool) []stepsFinding {
	var out []stepsFinding
	stmts := body.List
	facts := make([]bodyFacts, len(stmts))
	for i, st := range stmts {
		facts[i] = sc.factsOf(st)
	}
	for i := range stmts {
		if !facts[i].asks {
			continue
		}
		for j := i + 1; j < len(stmts); j++ {
			if !facts[j].acts {
				continue
			}
			out = append(out, stepsFinding{
				Flat: flat, Ask: lineOf(stmts[i]), Why: facts[i].why, Later: lineOf(stmts[j]),
				Shares: sharedNames(stmts[i], stmts[j:]),
			})
			// One finding per body sequence: the first pair is the proof.
			return out
		}
	}
	for _, st := range stmts {
		ast.Inspect(st, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncLit:
				return false
			case *ast.BlockStmt:
				if n != body {
					if clauseList(n) {
						// A switch's clauses are exclusive arms, not a
						// sequence (#4162): each is scanned on its own.
						return true
					}
					out = append(out, scanSequence(sc, n, false)...)
					return false
				}
			case *ast.CaseClause:
				out = append(out, scanSequence(sc, &ast.BlockStmt{List: n.Body}, false)...)
				return false
			case *ast.CommClause:
				out = append(out, scanSequence(sc, &ast.BlockStmt{List: n.Body}, false)...)
				return false
			}
			return true
		})
	}
	return out
}

// clauseList reports whether a block is a switch's or a select's body:
// a list of clauses, of which one runs.
func clauseList(b *ast.BlockStmt) bool {
	if len(b.List) == 0 {
		return false
	}
	for _, st := range b.List {
		switch st.(type) {
		case *ast.CaseClause, *ast.CommClause:
		default:
			return false
		}
	}
	return true
}

var stepsPos *token.FileSet

func lineOf(n ast.Node) int {
	if stepsPos == nil {
		return int(n.Pos())
	}
	return stepsPos.Position(n.Pos()).Line
}

// sharedNames are the identifiers the earlier statement declares (other
// than err and _) that a later statement reads.
func sharedNames(first ast.Stmt, later []ast.Stmt) []string {
	decl := map[string]bool{}
	ast.Inspect(first, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				for _, l := range n.Lhs {
					if id, ok := l.(*ast.Ident); ok && id.Name != "_" && id.Name != "err" {
						decl[id.Name] = true
					}
				}
			}
		}
		return true
	})
	if len(decl) == 0 {
		return nil
	}
	used := map[string]bool{}
	for _, st := range later {
		ast.Inspect(st, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && decl[id.Name] {
				used[id.Name] = true
			}
			return true
		})
	}
	var names []string
	for n := range used {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// loadEDHRecRanks reads edhrec_rank per oracle id from the Scryfall dump
// (CMDCTRL_SCRYFALL_DUMP). A missing dump gives an empty map.
func loadEDHRecRanks(t *testing.T) map[string]int {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Log("CMDCTRL_SCRYFALL_DUMP is unset: every EDHREC rank reads 0")
		return nil
	}
	fh, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	dec := json.NewDecoder(fh)
	if _, err := dec.Token(); err != nil {
		t.Fatal(err)
	}
	ranks := map[string]int{}
	for dec.More() {
		var c struct {
			OracleID string `json:"oracle_id"`
			Name     string `json:"name"`
			Rank     int    `json:"edhrec_rank"`
		}
		if err := dec.Decode(&c); err != nil {
			t.Fatal(err)
		}
		if c.Rank <= 0 {
			continue
		}
		// Keyed by oracle id and by lower-cased name: a two-faced card
		// has no top-level oracle id, and a row may name its id by a constant.
		for _, k := range []string{c.OracleID, strings.ToLower(c.Name)} {
			if k != "" && (ranks[k] == 0 || c.Rank < ranks[k]) {
				ranks[k] = c.Rank
			}
		}
	}
	return ranks
}

// stepsConverted are the cards #4155's first batch converted: each body of
// theirs that has an asking statement before an acting one is now a step
// sequence (ResolveSteps, ModeSteps or a row's Do), so the scan finds none
// of them. A card re-flattened into one function fails here with the line.
var stepsConverted = setOf(
	"Frantic Search", "Vampiric Tutor", "Black Market Connections", "Deadly Dispute", "Return of the Wildspeaker",
	"Night's Whisper", "The Great Henge", "Sensei's Divining Top", "Malakir Rebirth", "Growth Spiral",
	"Rishkar's Expertise", "Takenuma, Abandoned Mire", "Inspiring Call", "Ragavan, Nimble Pilferer",
	"Loran of the Third Path", "Shamanic Revelation", "Snakeskin Veil", "Eldritch Evolution",
	"Read the Bones", "Veil of Summer", "Imperial Seal", "Sword of Feast and Famine", "Bitterblossom",
	"Greater Good", "Archivist of Oghma", "Bloodchief Ascension", "Castle Locthwain", "Yawgmoth, Thran Physician",
	"Buster Sword", "Serum Visions", "Pull from Tomorrow", "Tezzeret's Gambit", "Moldervine Reclamation",
	"Sythis, Harvest's Hand", "Secret Rendezvous", "Prismari Command", "Glissa Sunslayer", "Peer into the Abyss",
	"Last March of the Ents", "Lumra, Bellow of the Woods", "Slip Out the Back", "Grand Crescendo",
	"Plumb the Forbidden", "Vampiric Rites", "Dreadhorde Invasion", "Lightning Helix", "Consume Spirit",
	// Batch 2.
	"Felidar Retreat", "Unbreakable Formation", "Sea Gate Restoration", "Revitalizing Repast", "Seize the Spoils",
	"Elspeth, Storm Slayer", "Breach the Multiverse", "Descent into Avernus", "Kappa Cannoneer",
	"Setessan Champion", "Demolition Field", "Insatiable Avarice", "Arachnogenesis",
	"Kambal, Consul of Allocation", "Finale of Revelation", "Crime Novelist", "Ancient Silver Dragon",
	"Hydroid Krasis", "Uro, Titan of Nature's Wrath", "Sunscorch Regent", "Inspiring Overseer",
	"Gwenna, Eyes of Gaea", "Smuggler's Share", "Ugin, Eye of the Storms", "Painful Truths", "Flame of Anor",
	"Uthros Research Craft", "Blue Sun's Zenith", "Thought Scour", "Collector's Vault", "Undead Augur",
	"Korvold, Fae-Cursed King", "Bitterbloom Bearer", "Ob Nixilis, the Fallen", "White Sun's Twilight",
	"Skrelv's Hive", "Eureka Moment", "Liliana, Death's Majesty", "Drown in Dreams", "Galadhrim Ambush",
	"Rise of the Eldrazi", "Grapple with the Past", "Riot Control", "Circle of Power", "Greedy Freebooter",
	"Consuming Corruption", "Forge of Heroes", "Valgavoth, Harrower of Souls", "Persistent Constrictor",
	"Flumph", "Lazotep Plating", "Momentous Fall", "Wan Shi Tong, Librarian", "Wheel of Fortune",
	"Expropriate", "Will of the Abzan", "Zack Fair",
	// #4162/#4163: continuations written as steps (#4162) and what one step
	// hands the next (#4163). Currency Converter's pair was two arms of one
	// if, now one instruction.
	"Arcane Denial", "Braids, Arisen Nightmare", "Forgotten Ancient", "Twilight Prophet", "Scavenging Ooze",
	"Seize the Spotlight", "Disciple of Freyalise", "Disciple of Bolas", "Replicating Ring",
	"Bonehoard Dracosaur", "Midnight Clock", "Currency Converter", "Court of Ambition", "Transpose",
	"Sylvan Library", "Martial Coup", "Windfall", "Dark Deal", "Tolarian Winds", "Cut a Deal",
	// Batch 3.
	"Finale of Glory", "Lifestream's Blessing", "Beacon of Immortality", "Farmer Cotton",
	"Xantcha, Sleeper Agent", "Ravenous Squirrel", "Damnable Pact", "Extract from Darkness",
	"Jace, the Mind Sculptor", "Lord of the Nazgûl", "Bloodgift Demon", "Ambition's Cost",
	"Tidus, Yuna's Guardian", "Flick a Coin", "Voyage Home", "Reckoner's Bargain",
	"Teferi, Hero of Dominaria", "Sorin Markov", "Tyvar Kell", "Dark Prophecy", "Sword of Body and Mind",
	"Moogles' Valor", "Thoughtseize", "Liliana's Triumph", "Compulsive Research", "Cloud, Ex-SOLDIER",
	"Dream Fracture", "Pawpatch Formation", "Agonasaur Rex", "Fevered Visions", "Nightshade Harvester",
	"Moria Scavenger", "Rush of Inspiration", "The Wandering Emperor", "Lord Skitter's Blessing",
	"The Speed Demon", "Curse of Fool's Wisdom", "Great Intelligence's Plan", "Aquitect's Will",
	"Dreamdew Entrancer", "Siegfried, Famed Swordsman", "Davros, Dalek Creator", "Kiora, the Crashing Wave",
	"Phlage, Titan of Fire's Fury", "Sorin the Mirthless", "Derelict Attic // Widow's Walk",
	"Rejoin the Fight", "Meat Locker // Drowned Diner", "Dr. Eggman", "Contaminated Drink",
	"Twisted Sewer-Witch",
	// Batch 4.
	"Birthday Escape", "Revitalize", "Abzan Charm", "Ardbert, Warrior of Darkness", "Summon: Magus Sisters",
	"Summon: Ixion", "Change of Fortune", "Sauron, the Dark Lord", "Ox of Agonas", "Ogre-Head Helm",
	"Ajani Unrelenting", "Behold the Beyond", "Umaro, Raging Yeti", "Collective Defiance", "Snort",
	"Ruin Grinder", "Sail into the West", "Sycorax Commander", "Ill-Gotten Gains", "Will of the Sultai",
	"Gluntch, the Bestower", "Dawn of a New Age", "The Gitrog, Ravenous Ride", "Tivit, Seller of Secrets",
	"Locke, Treasure Hunter", "Summon: Kujata", "Glimmer of Genius", "Illusion of Choice",
	"Brutal Hordechief", "Dromoka's Command", "Mog, Moogle Warrior", "Starving Revenant",
	"Círdan the Shipwright", "Trap the Trespassers", "Fleeting Flight", "Let's Play a Game",
	"Emissary Green", "Peema Trailblazer", "Sorin, Grim Nemesis", "Hunted by The Family",
	"Smite the Deathless", "Estinien Varlineau", "Krav, the Unredeemed", "Depopulate",
	"Pendant of Prosperity", "Bloatfly Swarm", "Agent of the Shadow Thieves", "Shadow, Mysterious Assassin",
	"Stern Lesson", "Crushing Disappointment",
	// Batch 5.
	"Narset of the Ancient Way", "Truth or Consequences", "Noosegraf Mob", "Blight Titan",
	"Synchronized Charge", "Read the Runes", "Erestor of the Council", "Sunder the Gateway",
	"Clan Crafter", "Delirium", "Collective Brutality", "Kick in the Door",
	"Jace, Multiverse Architect", "Lich Lord of Unx", "Jeskai Revelation", "You Find a Cursed Idol",
	"Travel Through Caradhras", "Promise of Power", "Skeletal Scrying", "Severed Strands",
	"Scurry of Gremlins", "The Torment of Gollum", "Survivor's Med Kit", "Explosion of Riches",
	"Niv-Mizzet, Ghost Counsel", "Boommobile", "Fog of War", "Soul-Shackled Zombie",
	"Messenger Jays", "Glint Weaver", "Fatehold Charm", "Damocles Base, Sword of Kang",
	"Aggressive Negotiations", "Ingris Stingerquill", "Draconic Intervention", "Zygon Infiltrator",
	"Kimoyo Beads", "Bushmeat Poacher", "Lieutenants of the Guard", "Aerid Konstrari",
	"Pyrrhic Blast", "Torch the Tower", "Eyes in the Skies",
	// Batch 6.
	"Midnight Oil", "Necromancer's Stockpile", "Agonizing Remorse", "Way of the Paradox",
	"Simulacrum", "Curse of the Werefox", "Oketra's Last Mercy", "Flame On!",
	"Thought Erasure", "Humiliate", "Vigorbloom Charm", "Glint-Sleeve Siphoner",
	"Ego Drain", "Orchard Elemental", "Favored Hoplite", "Ebony Charm",
	"Narset's Rebuke", "Repulsor Blast", "Sorin, Vampire Lord", "Way of the Healer",
	"Toll of the Invasion", "Stubborn Burrowfiend", "Vraska's Final Mercy", "The Crowd Goes Wild",
	"Sphinx's Approach", "Harsh Scrutiny", "Eyes of Gitaxias",
	"Render Speechless", "Live Fast", "Face Yourself", "Lightning Storm",
	"Splitting the Powerstone", "Gix's Caress", "Hungering Puppetbeast", "Scion of Vitu-Ghazi",
	"Eliminate the Impossible", "Konstrari Charm", "Rise of the Deathbringer", "Cease",
	"Angelic Favor", "Hired Hexblade", "Proft, Consulting Detective", "Zealot il-Vec",
	"Inspired Idea", "Horncaller's Chant", "Primal Witchstalker",
	// Batch 7.
	"Meditate", "Surgical Precision", "Flames of the Blood Hand", "Edgar, Ancient Bloodlord",
	"Command the Stage", "No Admittance", "Scorching Lava", "Coursers' Accord",
	"Timeline Inquiry", "Cut In", "Searing Barb", "Hexhaven Battalion",
	"Violent Echoes", "Burning Wish", "Brainbite", "Rhonas's Last Stand",
	"Memory Theft", "Tam's Resistance", "Living Wish", "Break Under Pressure",
	"Illuminate", "Cunning Wish", "Death Wish", "Absolving Lammasu",
	"Diplomacy of the Wastes", "Glittering Wish", "Spiritual Focus", "Stinging Vitriol",
	"Ribbons of Night", "Awaken the Inferno", "Pillar Tombs of Aku", "Teyo, Diamondblade Mage",
	"Fateshaper Aspirant", "Golden Wish", "Mangara's Blessing", "Arc of Fortune",
	"Arcee, Acrobatic Coupe", "Arcee, Sharpshooter", "Curse of Leeches", "Cyclonus, Cybertronian Fighter",
	"Flamewar, Brash Veteran", "Initiate of Blood", "Profane Insight", "Restorative Burst",
	"The Great Synthesis", "The Legend of Yangchen", "Wretched Bonemass",
	// #4228: the four bodies that carried Go state between prompts.
	"Rhys, the Evermore", "Cecily, Haunted Mage", "The Ur-Sphinx", "Flamewar, Streetwise Operative",
)

// stepsSkipped are the cards the scan flags that are not converted, each
// with the reason, so the list of what is left shrinks honestly (#4155). A
// reason is one of a few kinds:
//
//   - exclusive arms: the two statements are in arms of one switch or if that
//     cannot both run, so nothing outruns anything;
//   - no prompt: the earlier statement only reveals, or picks at random, and
//     asks nothing;
//   - one batch: both statements are inside one simultaneous damage or draw
//     instance, which is a single event (CR 120.2);
//   - disjoint players: the two statements act on different players.
//
// A continuation (a Then, a callback) is written as RegisterThenSteps
// (#4162), and a count or ids a later step needs goes through the step
// record (#4163), so neither is a reason any more. The arms of one
// switch are not scanned as a sequence (clauseList), so a switch needs
// no entry either.
//
// TestSkippedBodiesAreStillFlagged fails if an entry is stale.
var stepsSkipped = map[string]string{
	"Herald's Horn":              "no prompt: a reveal, then the bounce, inside a MayChoice's OnYes",
	"Geier Reach Sanitarium":     "each player draws and then discards in turn: one loop over the players",
	"Chandra's Ignition":         "one batch: the damage to each creature and opponent is one instance",
	"Tibalt's Trickery":          "no prompt: the number is picked at random",
	"Tribute to the World Tree":  "exclusive arms: a draw, or the counters",
	"Plaguecrafter":              "disjoint players: the discards are for players with nothing to sacrifice",
	"Krenko, Tin Street Kingpin": "exclusive arms: Krenko left, or the counter's Then",
	"Sheoldred's Edict":          "exclusive arms of one mode switch",
	"Painful Quandary":           "exclusive arms: the life loss or the question",
	"Genesis Wave":               "no prompt: a reveal asks nothing, so the cards it names need no step record",
	"Coiling Oracle":             "no prompt: a reveal asks nothing, so the card needs no step record",
	"Mogis, God of Slaughter":    "exclusive arms: the sacrifice, or the damage",
	"Animist's Awakening":        "no prompt: a reveal asks nothing, so the cards need no step record",
	"Thrasios, Triton Hero":      "no prompt: a reveal inside the scry's Then",
	"Fangs of Kalonia":           "exclusive arms: overloaded, or the target",
	"Cosmos Elixir":              "exclusive arms: a draw, or the life",
	// Batch 3.
	"Urza's Bauble":           "no prompt: a random pick from a hand, and a private look",
	"Deadbridge Chant":        "no prompt: the card is picked at random",
	"Twist Reality":           "exclusive arms of one mode switch",
	"Path of Discovery":       "exclusive arms: the explore by object, or by id",
	"Vexing Radgull":          "exclusive arms: the rad counters, or the proliferate",
	"Shelinda, Yevon Acolyte": "exclusive arms: the counter goes on the smaller of the two",
	"Ravenous Tyrannosaurus":  "one batch: the damage and its excess to the controller are one instruction (CR 120.4a)",
	"Fear, Fire, Foes!":       "one batch: one damage instance",
	"Dragonspark Reactor":     "one batch: one damage instance",
	"Hell to Pay":             "carried: the Treasures are the excess over lethal damage, a count only the damage step knows (#4163)",
	"Contest of Claws":        "carried: the discover number is the excess over lethal damage, a count only the damage step knows (#4163)",
	"Awaken the Erstwhile":    "carried: each player's Zombies are the cards they discarded (#4163)",
	"Decaying Time Loop":      "carried: the number of cards discarded is the number drawn (#4163)",
	"Lagomos, Hand of Hatred": "carried: the delayed sacrifice names the token just made (#4163)",
	// Batch 4.
	"Selvala's Stampede":         "exclusive arms: no wild votes returns at once, else the reveal's Then puts the free permanents",
	"Caustic Bronco":             "exclusive arms: each opponent loses the life, or you do",
	"Break Down the Door":        "exclusive arms: manifest dread, or the exile",
	"Yuna's Whistle":             "no prompt: the reflexive trigger goes on the stack after the spell finishes (CR 603.12)",
	"Tamiyo, Collector of Tales": "no prompt: a reveal asks nothing, and the cards it names are the next step's input",
	"Majestic Genesis":           "no prompt: a reveal asks nothing, and the cards it names are the put's input",
	"Coercive Portal":            "exclusive arms: the draw, or the sacrifice and the destruction",
	"Diviner's Portent":          "exclusive arms: the draw, or the scry's Then",
	"Tyrant's Choice":            "exclusive arms: the sacrifices, or the life loss",
	// Batch 5.
	"Galadriel, Elven-Queen":    "exclusive arms: the draw, or the Ring tempts and the counter",
	"Vault 11: Voter's Dilemma": "exclusive arms: every player draws, or the voted creatures are destroyed",
	"Fatal Grudge":              "exclusive arms: the draw, or the sacrifices and the draw in their Then",
	"Earth's Mightiest Heroes":  "no prompt: a reveal asks nothing, and the cards it names are the put's input",
	// Batch 6.
	"Surreal Memoir":    "no prompt: the card is picked at random",
	"Burn the Accursed": "one batch: the damage to the creature and to its controller are one instance (the exile rider after it is a step)",
	"Devour Intellect":  "exclusive arms: the choice from the revealed hand, or the discard",
	"Earthen Arms":      "exclusive arms: with no target awaken at once, else the counters' Then awakens",
	// Batch 7.
	"Repeat Offender":        "exclusive arms: the counter, or the suspect",
	"Gruesome Discovery":     "exclusive arms: the reveal and choice, or the plain discard",
	"Desperate Futurescribe": "exclusive arms: the counter, or the +1/+1 boost",
}

func TestSkippedBodiesAreStillFlagged(t *testing.T) {
	findings, names := scanStepsBodies(t, nil)
	flagged := map[string]bool{}
	for _, f := range findings {
		flagged[f.Name] = true
	}
	for name, why := range stepsSkipped {
		switch {
		case !names[name]:
			t.Errorf("%q is in stepsSkipped but no catalog row has that name", name)
		case stepsConverted[name]:
			t.Errorf("%q is in stepsSkipped and in stepsConverted", name)
		case !flagged[name]:
			t.Errorf("%q is in stepsSkipped (%s) but the scan no longer flags it: remove it", name, why)
		}
	}
}

func TestConvertedBodiesStayConverted(t *testing.T) {
	findings, names := scanStepsBodies(t, nil)
	for name := range stepsConverted {
		if !names[name] {
			t.Errorf("%q is in stepsConverted but no catalog row has that name", name)
		}
	}
	for _, f := range findings {
		if stepsConverted[f.Name] {
			t.Errorf("%s (%s:%d, %s): L%d (%s) can open a prompt and L%d runs before its answer — write the body as ResolveSteps, ModeSteps or Do (docs/adding-cards.md, CR 608.2c, #4155)",
				f.Name, f.File, f.Line, f.Site, f.Ask, f.Why, f.Later)
		}
	}
}
