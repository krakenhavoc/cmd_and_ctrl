package effects

import (
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// token_text_guard_test.go is #4005's lint, beside #3872's
// (text_word_guard_test.go): the two classes of printed word ADR 0160's
// slice 3 left unread.
//
// A TOKEN DESCRIPTION ("create a 2/2 black Zombie creature token") is
// words of the creating object's text (CR 612.2), so its colours and
// types follow a Sleight of Mind or an Artificial Evolution. CreateToken
// and CreateTokenAdvanced read their template through the effect's text
// themselves. A card file that hands a template to the game's creation
// methods (CreateTokenForEffect and its kin, or a game.TokenGroup) names
// what the template is, with one of tokenTemplateDoors:
//
//   - ctx.TextToken(t), or reader.TokenThroughText(t): a description,
//     read through the text;
//   - copiedToken(t): a copy of an object, copiable values (CR 707.2);
//   - ruleDefinedToken(t): a token a keyword action defines (living
//     weapon's Germ, incubate's Incubator), no word of the text.
//
// A variable is fine when every assignment to it in the same function is
// one of those calls. Anything else is a finding.
//
// A CAST-PERMISSION CONDITION that names a creature type (Robber of the
// Rich's "a Rogue") is fixed as the effect that grants it resolves, so a
// card file reads the word through its text there:
// game.PermissionCondition{Subtype: ctx.TextSubtype("Rogue")}. A literal
// is a finding. A PERMISSION FILTER's creature type is read through the
// granting permanent's text by the engine as the permission is derived
// (stampStandingPermissionLocked, the granted alternative costs), so a
// literal is how a standing permission names its word, and only the files
// in permissionFilterDerived may hold one.

// tokenCreationArgs are the game's token-creation methods, with the
// index of the template argument.
var tokenCreationArgs = map[string]int{
	"CreateTokenForEffect":           1,
	"CreateTokensForEffect":          1,
	"CreateTokensAttackingForEffect": 1,
	"CreateTokensBlockingForEffect":  1,
}

// tokenTemplateDoors are the calls that say what a template is.
var tokenTemplateDoors = map[string]bool{
	"TextToken":        true,
	"TokenThroughText": true,
	"copiedToken":      true,
	"ruleDefinedToken": true,
}

// permissionFilterDerived are the files whose PermissionFilter names a
// creature type the engine reads through the granting permanent's text
// as it derives the permission, each with where.
var permissionFilterDerived = map[string]string{
	"gisa_and_geralf.go":          "standing cast permission (stampStandingPermissionLocked)",
	"conspicuous_snoop.go":        "standing play permission (stampStandingPermissionLocked)",
	"deeproot_historian.go":       "standing retrace permission (stampStandingPermissionLocked)",
	"granted_alternative_cost.go": "Rooftop Storm's granted alternative cost (grantedAlternativeCostsLocked)",
	"blur_of_heroism.go":          "standing cast timing (castTimingVerdictLocked)",
	"rivaz_of_the_claw.go":        "standing cast permission (stampStandingPermissionLocked)",
}

func TestTokenDescriptionsReadThroughText(t *testing.T) {
	fset := token.NewFileSet()
	files := parseCatalogSources(t, fset)
	if len(files) == 0 {
		t.Fatal("scanned no catalog files — the glob is wrong, and a lint that reads nothing passes everything")
	}

	var findings []string
	seen := map[string]int{}
	usedDerived := map[string]bool{}
	for file, tree := range files {
		// Each top-level declaration — a func, or a var holding a
		// registered continuation — is the scope its variables are read
		// in.
		for _, decl := range tree.Decls {
			assigned := assignmentsIn(decl)
			check := func(what string, at token.Pos, tmpl ast.Expr) {
				seen[what]++
				if templateNamed(tmpl, assigned) {
					return
				}
				findings = append(findings, file+":"+strconv.Itoa(fset.Position(at).Line)+" "+what)
			}
			ast.Inspect(decl, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.CallExpr:
					if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
						if i, ok := tokenCreationArgs[sel.Sel.Name]; ok && i < len(n.Args) {
							check(sel.Sel.Name, n.Pos(), n.Args[i])
						}
					}
				case *ast.CompositeLit:
					if isGameType(n.Type, "TokenGroup") {
						checkTokenGroup(n, check)
					}
					if arr, ok := n.Type.(*ast.ArrayType); ok && isGameType(arr.Elt, "TokenGroup") {
						for _, el := range n.Elts {
							if lit, ok := el.(*ast.CompositeLit); ok && lit.Type == nil {
								checkTokenGroup(lit, check)
							}
						}
					}
				}
				return true
			})
		}
		ast.Inspect(tree, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			switch {
			case isGameType(lit.Type, "PermissionCondition"):
				if fieldValue(lit, "Subtype") != nil {
					seen["PermissionCondition"]++
				}
				for _, word := range creatureTypeLiterals(fieldValue(lit, "Subtype")) {
					findings = append(findings, file+":"+strconv.Itoa(fset.Position(lit.Pos()).Line)+
						" PermissionCondition{Subtype: "+strconv.Quote(word)+"}")
				}
			case isGameType(lit.Type, "PermissionFilter"):
				words := append(creatureTypeLiterals(fieldValue(lit, "CreatureType")),
					creatureTypeLiterals(fieldValue(lit, "CreatureTypesAny"))...)
				for _, word := range words {
					seen["PermissionFilter"]++
					if _, ok := permissionFilterDerived[file]; ok {
						usedDerived[file] = true
						continue
					}
					findings = append(findings, file+":"+strconv.Itoa(fset.Position(lit.Pos()).Line)+
						" PermissionFilter naming "+strconv.Quote(word))
				}
			}
			return true
		})
	}

	// The control: a scan that stopped matching would pass every card.
	for name := range tokenCreationArgs {
		if seen[name] == 0 {
			t.Errorf("no catalog file calls .%s — drop it from tokenCreationArgs", name)
		}
	}
	if seen["TokenGroup"] == 0 {
		t.Error("no catalog file builds a game.TokenGroup — the literal scan is matching nothing")
	}
	if seen["PermissionCondition"] == 0 {
		t.Error("no catalog file sets a PermissionCondition's Subtype — the literal scan is matching nothing")
	}
	if seen["PermissionFilter"] == 0 {
		t.Error("no catalog file names a creature type in a PermissionFilter — the literal scan is matching nothing")
	}
	for file := range permissionFilterDerived {
		if !usedDerived[file] {
			t.Errorf("%s no longer names a creature type in a PermissionFilter: drop it from permissionFilterDerived", file)
		}
	}

	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("%d token template(s) or permission word(s) not read through the text whose words they are (#4005, CR 612.2).\n"+
		"Hand a token description to the game as ctx.TextToken(t) (or reader.TokenThroughText(t)), a copy as "+
		"copiedToken(t), a keyword action's token as ruleDefinedToken(t); read a permission condition's creature "+
		"type with ctx.TextSubtype (ADR 0160's 2026-10-16 slice-4 follow-up).\n  %s",
		len(findings), strings.Join(findings, "\n  "))
}

// checkTokenGroup checks a game.TokenGroup literal's Template.
func checkTokenGroup(lit *ast.CompositeLit, check func(string, token.Pos, ast.Expr)) {
	if tmpl := fieldValue(lit, "Template"); tmpl != nil {
		check("TokenGroup", lit.Pos(), tmpl)
	}
}

// templateNamed reports whether a template expression is a door call, or
// a variable every assignment of which (in the same function) is one.
func templateNamed(e ast.Expr, assigned map[string][]ast.Expr) bool {
	if isDoorCall(e) {
		return true
	}
	id, ok := e.(*ast.Ident)
	if !ok {
		return false
	}
	rhs := assigned[id.Name]
	if len(rhs) == 0 {
		return false
	}
	for _, r := range rhs {
		if !isDoorCall(r) {
			return false
		}
	}
	return true
}

func isDoorCall(e ast.Expr) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return false
	}
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return tokenTemplateDoors[fn.Name]
	case *ast.SelectorExpr:
		return tokenTemplateDoors[fn.Sel.Name]
	}
	return false
}

// assignmentsIn maps each variable name to the right-hand sides assigned
// to it anywhere in scope (one-to-one assignments only; a multi-value one
// records nothing, so a variable set only that way is never a door).
func assignmentsIn(scope ast.Node) map[string][]ast.Expr {
	out := map[string][]ast.Expr{}
	ast.Inspect(scope, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, l := range as.Lhs {
			id, ok := l.(*ast.Ident)
			if !ok {
				continue
			}
			if len(as.Rhs) == len(as.Lhs) {
				out[id.Name] = append(out[id.Name], as.Rhs[i])
			} else {
				out[id.Name] = append(out[id.Name], &ast.BadExpr{})
			}
		}
		return true
	})
	return out
}

// isGameType reports whether e is the type game.<name>.
func isGameType(e ast.Expr, name string) bool {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "game"
}

// fieldValue is the value of a keyed field in a composite literal, or nil.
func fieldValue(lit *ast.CompositeLit, key string) ast.Expr {
	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if id, ok := kv.Key.(*ast.Ident); ok && id.Name == key {
			return kv.Value
		}
	}
	return nil
}

// creatureTypeLiterals are the creature-type string literals in e: e
// itself, or the elements of an array literal.
func creatureTypeLiterals(e ast.Expr) []string {
	var exprs []ast.Expr
	switch v := e.(type) {
	case nil:
		return nil
	case *ast.CompositeLit:
		exprs = v.Elts
	default:
		exprs = []ast.Expr{v}
	}
	var out []string
	for _, x := range exprs {
		lit, ok := x.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		word, err := strconv.Unquote(lit.Value)
		if err == nil && isTextWordArgument(word) {
			out = append(out, word)
		}
	}
	return out
}
