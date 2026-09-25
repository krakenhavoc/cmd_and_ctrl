package effects

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// duration_grant_cards_test.go — ADR 0093 PR 4 (#1584): abilities a
// RESOLVING spell or ability grants for a duration, on real cards. The
// grant is a ScopedEffect record with a grantAbilities mod; every test
// asserts what a player sees — the creature coming back, the Treasure,
// the bounce, the Saga's mana — not merely that a record was written.

const (
	feignDeathOracle       = "f718e507-296b-4f22-842b-5fb91322069b"
	fakeYourOwnDeathOracle = "ad01df89-29fe-44c7-a133-91425f8ff09c"
	retractionHelixOracle  = "2707f9f4-2b80-47ac-af0a-99fc53af94bf"
	malakirRebirthOracle   = "a731e87b-8d99-4b64-8ee3-8e540d652366"
	urzasSagaOracle        = "4c6a0c30-b547-4eff-8ff4-0ca25803c076"
)

func dgCardRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// dgCastOn casts an instant from the active seat's hand at `target`
// and resolves it.
func dgCastOn(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, name, "Instant", oracle, dgCardRef(target))
	passPriorityAroundTable(t, g)
}

func dgTreasuresOf(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == controller {
			n++
		}
	}
	return n
}

// Feign Death: the creature dies and comes back tapped with a +1/+1
// counter — and, being a new object the grant never named (CR 400.7),
// the second death is final.
func TestFeignDeathReturnsTheCreatureOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	dgCastOn(t, g, "Feign Death", feignDeathOracle, bear)
	if grantedTriggerCount(g, bear) != 1 {
		t.Fatal("the target has Feign Death's granted dies trigger")
	}

	b18Kill(t, g, bear)
	back := findBattlefieldCardByID(g, bear)
	if back == nil {
		t.Fatal("the creature died and did not come back")
	}
	if !back.Tapped {
		t.Error("it returns tapped")
	}
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if grantedTriggerCount(g, bear) != 0 {
		t.Error("the returned creature is a new object without the grant")
	}

	b18Kill(t, g, bear)
	if findBattlefieldCardByID(g, bear) != nil || !inGraveyardOf(g, me.ID, bear) {
		t.Error("the second death is final")
	}
}

// "Until end of turn": after the cleanup step the creature no longer
// has the ability, and the record is gone.
func TestFeignDeathEndsAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	dgCastOn(t, g, "Feign Death", feignDeathOracle, bear)
	advancePastCleanupForTest(t, g)
	if grantedTriggerCount(g, bear) != 0 {
		t.Error("the grant outlived the turn")
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d scoped effects outlived their duration", n)
	}
	b18Kill(t, g, bear)
	if findBattlefieldCardByID(g, bear) != nil {
		t.Error("after the turn, the creature stays dead")
	}
}

// A later "loses all abilities" on the recipient takes the grant
// (CR 613.6): the grant is the creature's ability, not the spell's.
func TestFeignDeathIsLostToALaterAbilityRemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	dgCastOn(t, g, "Feign Death", feignDeathOracle, bear)
	pushAuraOnLand(t, g, me.ID, "Darksteel Mutation", darksteelMutationOracle, bear)
	if grantedTriggerCount(g, bear) != 0 {
		t.Fatal("a later ability removal takes the granted trigger")
	}
}

// The payoff of making it data: the table is a full restore point
// while the grant lives, and in the restored game the creature still
// comes back.
func TestFeignDeathSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	dgCastOn(t, g, "Feign Death", feignDeathOracle, bear)

	restored := restoreRoundTrip(t, g, true)
	if grantedTriggerCount(restored, bear) != 1 {
		t.Fatal("restored: the creature lost the granted trigger")
	}
	b18Kill(t, restored, bear)
	back := findBattlefieldCardByID(restored, bear)
	if back == nil {
		t.Fatal("restored: the creature died and did not come back")
	}
	if !back.Tapped || counterCount(restored, bear, game.CounterPlusOne) != 1 {
		t.Error("restored: it returns tapped with a +1/+1 counter")
	}
}

// Fake Your Own Death: +2/+0 now, and on death the creature returns and
// its CONTROLLER — the trigger's "you" — makes a Treasure.
func TestFakeYourOwnDeathBoostsReturnsAndMakesATreasureForTheController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	dgCastOn(t, g, "Fake Your Own Death", fakeYourOwnDeathOracle, theirs)
	if got := effectivePower(t, g, theirs); got != 4 {
		t.Errorf("power = %d, want 2 + 2", got)
	}

	b18Kill(t, g, theirs)
	back := findBattlefieldCardByID(g, theirs)
	if back == nil || !back.Tapped {
		t.Fatal("the creature returns tapped")
	}
	if back.Controller != opp.ID {
		t.Error("it returns under its owner's control")
	}
	if dgTreasuresOf(g, opp.ID) != 1 || dgTreasuresOf(g, me.ID) != 0 {
		t.Errorf("Treasures: theirs %d, mine %d — the trigger's controller makes it",
			dgTreasuresOf(g, opp.ID), dgTreasuresOf(g, me.ID))
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("the returned creature is a new object without the +2/+0: power %d", got)
	}
}

// Retraction Helix: the creature taps to bounce a nonland permanent.
// The ability is the creature's, activated by its controller, and gone
// after the turn.
func TestRetractionHelixGrantsABounce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	dgCastOn(t, g, "Retraction Helix", retractionHelixOracle, mine)

	idx, ref := grantedActivatedIndex(t, g, mine)
	if idx < 0 {
		t.Fatal("the target has the granted bounce")
	}
	// ADR 0093 Decision 8: the table sees what the creature can now do,
	// and which spell gave it. (The unfiltered view: these fixture
	// permanents were pushed straight onto the battlefield, so no seat is
	// recorded as knowing them.)
	found := false
	for _, c := range protocol.ViewOfGameFor(g, "").Battlefield.Cards {
		if c.InstanceID != mine.String() {
			continue
		}
		for _, ga := range c.GrantedAbilities {
			if ga.Text == "{T}: Return target nonland permanent to its owner's hand." && ga.SourceName == "Retraction Helix" {
				found = true
			}
		}
	}
	if !found {
		t.Error("the wire does not show the granted bounce from Retraction Helix")
	}
	b16Activate(t, g, me.ID, mine, idx, game.ActivateAbilityParams{Ref: ref, Targets: dgCardRef(theirs)})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, theirs) != nil {
		t.Error("the target nonland permanent is bounced")
	}
	if !b16Tapped(t, g, mine) {
		t.Error("{T}: the bounce taps the creature that has it")
	}

	advancePastCleanupForTest(t, g)
	if idx, _ := grantedActivatedIndex(t, g, mine); idx >= 0 {
		t.Error("the granted bounce outlived the turn")
	}
}

// Retraction Helix's bounce is "target NONLAND permanent".
func TestRetractionHelixCannotBounceALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	land := gaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	dgCastOn(t, g, "Retraction Helix", retractionHelixOracle, mine)
	idx, ref := grantedActivatedIndex(t, g, mine)
	if idx < 0 {
		t.Fatal("setup: the target has the granted bounce")
	}
	if err := g.ActivateCatalogAbility(me.ID, mine, idx,
		game.ActivateAbilityParams{Ref: ref, Targets: dgCardRef(land)}); err == nil {
		t.Error("a land is not a legal target")
	}
}

// Malakir Rebirth: 2 life, and the creature returns tapped with no
// counter.
func TestMalakirRebirthCostsTwoLifeAndReturnsTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	life := me.Life
	// The real modal DFC, cast as its front face.
	card := mdfcCard(me.ID, malakirRebirthOracle, game.LayoutModalDFC,
		game.Face{Name: "Malakir Rebirth", TypeLine: "Instant", ManaCost: "{B}", Colors: []string{"B"}},
		game.Face{Name: "Malakir Mire", TypeLine: "Land"},
	)
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Targets: dgCardRef(bear)}); err != nil {
		t.Fatalf("CastSpell Malakir Rebirth: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d", me.Life, life-2)
	}
	b18Kill(t, g, bear)
	back := findBattlefieldCardByID(g, bear)
	if back == nil || !back.Tapped {
		t.Fatal("the creature returns tapped")
	}
	if got := counterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters = %d, want none", got)
	}
}

// Urza's Saga: chapter I gives the Saga a {C} ability, chapter II the
// Construct ability, chapter III fetches an artifact with mana cost {1}
// (and not a {2} one), and the grants end with the Saga.
func TestUrzasSagaGrantsItselfAbilitiesThenFetches(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := castCatalogSpell(t, g, "Urza's Saga", "Enchantment Land — Urza's Saga", urzasSagaOracle, nil)
	passPriorityAroundTable(t, g)
	if loreCountersOn(g, saga) != 1 {
		t.Fatalf("setup: the Saga entered with %d lore counters", loreCountersOn(g, saga))
	}

	// I — {T}: Add {C}.
	if grantedManaRowsOf(g, saga) != 1 {
		t.Fatal("chapter I: the Saga has the granted {C} ability")
	}
	if err := tapGrantedMana(t, g, me.ID, saga, "C"); err != nil {
		t.Fatalf("tapping the Saga for {C}: %v", err)
	}
	if !b16Tapped(t, g, saga) {
		t.Error("the granted {T} taps the Saga")
	}

	// II — the Construct ability, beside the {C}.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	idx, ref := grantedActivatedIndex(t, g, saga)
	if idx < 0 {
		t.Fatal("chapter II: the Saga has the granted Construct ability")
	}
	if grantedManaRowsOf(g, saga) != 1 {
		t.Error("chapter II keeps chapter I's ability")
	}
	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, saga, idx, game.ActivateAbilityParams{Ref: ref})
	passPriorityAroundTable(t, g)
	if got := countBattlefieldByName(g, "Construct"); got != 1 {
		t.Fatalf("Constructs = %d, want 1", got)
	}

	// III — search for mana cost {0} or {1}.
	pushLibraryCardForTest(me, game.Card{Name: "Two-Drop Rock", TypeLine: "Artifact", ManaCost: "{2}"})
	pushLibraryCardForTest(me, game.Card{Name: "One-Drop Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("chapter III: no search prompt")
	}
	if searchOptionNamed(g, c, "Two-Drop Rock") != uuid.Nil {
		t.Error("a {2} artifact is not findable")
	}
	answerSearchNamed(t, g, me.ID, "One-Drop Rock")
	passPriorityAroundTable(t, g)
	if countBattlefieldByName(g, "One-Drop Rock") != 1 {
		t.Error("chapter III puts the {1} artifact onto the battlefield")
	}
	if findBattlefieldCardByID(g, saga) != nil {
		t.Fatal("the Saga is sacrificed after chapter III")
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d grant records outlived the Saga they were pinned to", n)
	}
}

// A Saga holding its two indefinite grants is a restore point, and the
// restored Saga still taps for {C}.
func TestUrzasSagaGrantsSurviveARestore(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	saga := castCatalogSpell(t, g, "Urza's Saga", "Enchantment Land — Urza's Saga", urzasSagaOracle, nil)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	restored := restoreRoundTrip(t, g, true)
	if grantedManaRowsOf(restored, saga) != 1 {
		t.Error("restored: the Saga lost its {C} ability")
	}
	if idx, _ := grantedActivatedIndex(t, restored, saga); idx < 0 {
		t.Error("restored: the Saga lost its Construct ability")
	}
	rme := restored.Seats[seat]
	if err := tapGrantedMana(t, restored, rme.ID, saga, "C"); err != nil {
		t.Errorf("restored: tapping the Saga for {C}: %v", err)
	}
}

// GrantAbilitiesFor refuses a bundle nobody registered, and one with a
// Static slot (ADR 0093 Decision 10).
func TestGrantAbilitiesForRefusesAnUnknownOrStaticBundle(t *testing.T) {
	if err := checkGrantMods([]game.Mod{game.GrantAbilitiesMod("nobody/registered-this")}); err == nil {
		t.Error("an unregistered bundle was accepted")
	}
	if err := checkGrantMods([]game.Mod{game.GrantAbilitiesMod(feignDeathReturn)}); err != nil {
		t.Errorf("Feign Death's bundle was refused: %v", err)
	}
	if err := checkGrantMods([]game.Mod{{Kind: game.ModGrantAbilities}}); err == nil {
		t.Error("a grant naming nothing was accepted")
	}
	// Phantasmal Image's copy grant has a Triggered slot only; find any
	// registered bundle with a Static slot to prove the second refusal.
	for key, def := range defs {
		if strings.HasPrefix(key, game.GrantKeyPrefix) && len(def.Static) > 0 {
			if err := checkGrantMods([]game.Mod{game.GrantAbilitiesMod(key)}); err == nil {
				t.Errorf("the static bundle %s was accepted", key)
			}
			break
		}
	}
}

// TestEveryDurationGrantKeyResolves is the build-time half of the
// registration guard (the twin of TestEveryGrantKeyResolves, which
// walks the statics): a resolving effect's bundle key lives in a
// closure no boot-time walk can read, so this scans the catalog's
// source for every GrantAbilitiesFor{Keys: …} literal and every
// game.GrantAbilitiesMod(…) call, resolves each key — a string literal
// or a package-level constant — and holds it to a registered bundle
// without a Static slot. A key built at runtime is its caller's risk,
// and Apply refuses it at resolution.
func TestEveryDurationGrantKeyResolves(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("glob: %v (%d files)", err, len(files))
	}
	fset := token.NewFileSet()
	var parsed []*ast.File
	consts := map[string]string{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		af, err := parser.ParseFile(fset, f, src, 0)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		parsed = append(parsed, af)
		for _, d := range af.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, spec := range gd.Specs {
				vs := spec.(*ast.ValueSpec)
				for i, name := range vs.Names {
					if i < len(vs.Values) {
						if lit, ok := vs.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							consts[name.Name], _ = strconv.Unquote(lit.Value)
						}
					}
				}
			}
		}
	}
	resolve := func(e ast.Expr) (string, bool) {
		switch v := e.(type) {
		case *ast.BasicLit:
			s, err := strconv.Unquote(v.Value)
			return s, err == nil
		case *ast.Ident:
			s, ok := consts[v.Name]
			return s, ok
		}
		return "", false
	}
	seen := 0
	check := func(e ast.Expr) {
		key, ok := resolve(e)
		if !ok {
			return
		}
		seen++
		if p := durationGrantKeyProblem(key); p != "" {
			t.Errorf("%s: %s", fset.Position(e.Pos()), p)
		}
	}
	for _, af := range parsed {
		ast.Inspect(af, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.CallExpr:
				if sel, ok := v.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "GrantAbilitiesMod" {
					for _, a := range v.Args {
						check(a)
					}
				}
			case *ast.CompositeLit:
				if id, ok := v.Type.(*ast.Ident); !ok || id.Name != "GrantAbilitiesFor" {
					return true
				}
				for _, el := range v.Elts {
					kv, ok := el.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					if k, ok := kv.Key.(*ast.Ident); !ok || k.Name != "Keys" {
						continue
					}
					if lit, ok := kv.Value.(*ast.CompositeLit); ok {
						for _, e := range lit.Elts {
							check(e)
						}
					}
				}
			}
			return true
		})
	}
	if seen == 0 {
		t.Fatal("no duration-grant keys found; the scan is broken")
	}
}
