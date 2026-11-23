package testsds

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsex "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ex"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Feign Death: the creature dies and comes back tapped with a +1/+1
// counter — and, being a new object the grant never named (CR 400.7),
// the second death is final.
func TestFeignDeathReturnsTheCreatureOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.DgCastOn(t, g, "Feign Death", effectstest.FeignDeathOracle, bear)
	if effectstest.GrantedTriggerCount(g, bear) != 1 {
		t.Fatal("the target has Feign Death's granted dies trigger")
	}

	effectstest.B18Kill(t, g, bear)
	back := effectstest.FindBattlefieldCardByID(g, bear)
	if back == nil {
		t.Fatal("the creature died and did not come back")
	}
	if !back.Tapped {
		t.Error("it returns tapped")
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if effectstest.GrantedTriggerCount(g, bear) != 0 {
		t.Error("the returned creature is a new object without the grant")
	}

	effectstest.B18Kill(t, g, bear)
	if effectstest.FindBattlefieldCardByID(g, bear) != nil || !effectstest.InGraveyardOf(g, me.ID, bear) {
		t.Error("the second death is final")
	}
}

// ADR 0115: a commander under Feign Death dies like any creature, and
// its granted "when this creature dies, return it" waits while CR 903.9a
// asks its owner (CR 704.3: the state-based action is performed before
// the trigger goes on the stack). On a "no" the trigger finds the card
// in the graveyard and returns it. On a "yes" the card has gone to the
// command zone, a new object (CR 400.7), and the trigger, which looks
// only in the first zone the card went to (CR 603.6c), returns nothing.
func TestFeignDeathOnACommanderFollowsItsOwnersAnswer(t *testing.T) {
	for _, home := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		cmd := effectstest.B16Creature(g, me.ID, "Commander", "Legendary Creature — Bear", 2, 2, "G")
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmd {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
		effectstest.DgCastOn(t, g, "Feign Death", effectstest.FeignDeathOracle, cmd)

		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cmd) })
		if !effectstest.InGraveyardOf(g, me.ID, cmd) {
			t.Fatalf("answer %v: the commander did not die", home)
		}
		if c := effectstest.CommanderReturnPromptFor(g, me.ID); c == nil {
			t.Fatalf("answer %v: no CR 903.9a question", home)
		}
		if len(g.StackMeta) != 0 {
			t.Fatalf("answer %v: the dies trigger went on the stack before the CR 903.9a answer", home)
		}
		effectstest.AnswerCommanderReturn(t, g, me.ID, home)
		effectstest.PassPriorityAroundTable(t, g)

		back := effectstest.FindBattlefieldCardByID(g, cmd)
		if home {
			if back != nil {
				t.Error("yes: Feign Death pulled the commander back out of the command zone")
			}
			if !me.Command.Contains(cmd) {
				t.Error("yes: the commander is not in the command zone")
			}
			continue
		}
		if back == nil {
			t.Error("no: Feign Death did not return the commander from the graveyard")
		} else if !back.Tapped {
			t.Error("no: it returns tapped")
		}
	}
}

// "Until end of turn": after the cleanup step the creature no longer
// has the ability, and the record is gone.
func TestFeignDeathEndsAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.DgCastOn(t, g, "Feign Death", effectstest.FeignDeathOracle, bear)
	effectstest.AdvancePastCleanupForTest(t, g)
	if effectstest.GrantedTriggerCount(g, bear) != 0 {
		t.Error("the grant outlived the turn")
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d scoped effects outlived their duration", n)
	}
	effectstest.B18Kill(t, g, bear)
	if effectstest.FindBattlefieldCardByID(g, bear) != nil {
		t.Error("after the turn, the creature stays dead")
	}
}

// A later "loses all abilities" on the recipient takes the grant
// (CR 613.6): the grant is the creature's ability, not the spell's.
func TestFeignDeathIsLostToALaterAbilityRemoval(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.DgCastOn(t, g, "Feign Death", effectstest.FeignDeathOracle, bear)
	effectstest.PushAuraOnLand(t, g, me.ID, "Darksteel Mutation", effectstest.DarksteelMutationOracle, bear)
	if effectstest.GrantedTriggerCount(g, bear) != 0 {
		t.Fatal("a later ability removal takes the granted trigger")
	}
}

// The payoff of making it data: the table is a full restore point
// while the grant lives, and in the restored game the creature still
// comes back.
func TestFeignDeathSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.DgCastOn(t, g, "Feign Death", effectstest.FeignDeathOracle, bear)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	if effectstest.GrantedTriggerCount(restored, bear) != 1 {
		t.Fatal("restored: the creature lost the granted trigger")
	}
	effectstest.B18Kill(t, restored, bear)
	back := effectstest.FindBattlefieldCardByID(restored, bear)
	if back == nil {
		t.Fatal("restored: the creature died and did not come back")
	}
	if !back.Tapped || effectstest.CounterCount(restored, bear, game.CounterPlusOne) != 1 {
		t.Error("restored: it returns tapped with a +1/+1 counter")
	}
}

// Fake Your Own Death: +2/+0 now, and on death the creature returns and
// its CONTROLLER — the trigger's "you" — makes a Treasure.
func TestFakeYourOwnDeathBoostsReturnsAndMakesATreasureForTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.DgCastOn(t, g, "Fake Your Own Death", effectstest.FakeYourOwnDeathOracle, theirs)
	if got := effectstest.EffectivePower(t, g, theirs); got != 4 {
		t.Errorf("power = %d, want 2 + 2", got)
	}

	effectstest.B18Kill(t, g, theirs)
	back := effectstest.FindBattlefieldCardByID(g, theirs)
	if back == nil || !back.Tapped {
		t.Fatal("the creature returns tapped")
	}
	if back.Controller != opp.ID {
		t.Error("it returns under its owner's control")
	}
	if effectstest.DgTreasuresOf(g, opp.ID) != 1 || effectstest.DgTreasuresOf(g, me.ID) != 0 {
		t.Errorf("Treasures: theirs %d, mine %d — the trigger's controller makes it",
			effectstest.DgTreasuresOf(g, opp.ID), effectstest.DgTreasuresOf(g, me.ID))
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("the returned creature is a new object without the +2/+0: power %d", got)
	}
}

// Retraction Helix: the creature taps to bounce a nonland permanent.
// The ability is the creature's, activated by its controller, and gone
// after the turn.
func TestRetractionHelixGrantsABounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.DgCastOn(t, g, "Retraction Helix", effectstest.RetractionHelixOracle, mine)

	idx, ref := effectstest.GrantedActivatedIndex(t, g, mine)
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
	effectstest.B16Activate(t, g, me.ID, mine, idx, game.ActivateAbilityParams{Ref: ref, Targets: effectstest.DgCardRef(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardByID(g, theirs) != nil {
		t.Error("the target nonland permanent is bounced")
	}
	if !effectstest.B16Tapped(t, g, mine) {
		t.Error("{T}: the bounce taps the creature that has it")
	}

	effectstest.AdvancePastCleanupForTest(t, g)
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, mine); idx >= 0 {
		t.Error("the granted bounce outlived the turn")
	}
}

// Retraction Helix's bounce is "target NONLAND permanent".
func TestRetractionHelixCannotBounceALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	land := effectstest.GaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	effectstest.DgCastOn(t, g, "Retraction Helix", effectstest.RetractionHelixOracle, mine)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, mine)
	if idx < 0 {
		t.Fatal("setup: the target has the granted bounce")
	}
	if err := g.ActivateCatalogAbility(me.ID, mine, idx,
		game.ActivateAbilityParams{Ref: ref, Targets: effectstest.DgCardRef(land)}); err == nil {
		t.Error("a land is not a legal target")
	}
}

// Malakir Rebirth: 2 life, and the creature returns tapped with no
// counter.
func TestMalakirRebirthCostsTwoLifeAndReturnsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	life := me.Life
	// The real modal DFC, cast as its front face.
	card := effectstest.MdfcCard(me.ID, effectstest.MalakirRebirthOracle, game.LayoutModalDFC,
		game.Face{Name: "Malakir Rebirth", TypeLine: "Instant", ManaCost: "{B}", Colors: []string{"B"}},
		game.Face{Name: "Malakir Mire", TypeLine: "Land"},
	)
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Targets: effectstest.DgCardRef(bear)}); err != nil {
		t.Fatalf("CastSpell Malakir Rebirth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d", me.Life, life-2)
	}
	effectstest.B18Kill(t, g, bear)
	back := effectstest.FindBattlefieldCardByID(g, bear)
	if back == nil || !back.Tapped {
		t.Fatal("the creature returns tapped")
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters = %d, want none", got)
	}
}

// Urza's Saga: chapter I gives the Saga a {C} ability, chapter II the
// Construct ability, chapter III fetches an artifact with mana cost {1}
// (and not a {2} one), and the grants end with the Saga.
func TestUrzasSagaGrantsItselfAbilitiesThenFetches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := effectstest.CastCatalogSpell(t, g, "Urza's Saga", "Enchantment Land — Urza's Saga", effectstest.UrzasSagaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoreCountersOn(g, saga) != 1 {
		t.Fatalf("setup: the Saga entered with %d lore counters", effectstest.LoreCountersOn(g, saga))
	}

	// I — {T}: Add {C}.
	if effectstest.GrantedManaRowsOf(g, saga) != 1 {
		t.Fatal("chapter I: the Saga has the granted {C} ability")
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, saga, "C"); err != nil {
		t.Fatalf("tapping the Saga for {C}: %v", err)
	}
	if !effectstest.B16Tapped(t, g, saga) {
		t.Error("the granted {T} taps the Saga")
	}

	// II — the Construct ability, beside the {C}.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, saga)
	if idx < 0 {
		t.Fatal("chapter II: the Saga has the granted Construct ability")
	}
	if effectstest.GrantedManaRowsOf(g, saga) != 1 {
		t.Error("chapter II keeps chapter I's ability")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, saga, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldByName(g, "Construct"); got != 1 {
		t.Fatalf("Constructs = %d, want 1", got)
	}

	// III — search for mana cost {0} or {1}.
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Two-Drop Rock", TypeLine: "Artifact", ManaCost: "{2}"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "One-Drop Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("chapter III: no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Two-Drop Rock") != uuid.Nil {
		t.Error("a {2} artifact is not findable")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "One-Drop Rock")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldByName(g, "One-Drop Rock") != 1 {
		t.Error("chapter III puts the {1} artifact onto the battlefield")
	}
	if effectstest.FindBattlefieldCardByID(g, saga) != nil {
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	saga := effectstest.CastCatalogSpell(t, g, "Urza's Saga", "Enchantment Land — Urza's Saga", effectstest.UrzasSagaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	if effectstest.GrantedManaRowsOf(restored, saga) != 1 {
		t.Error("restored: the Saga lost its {C} ability")
	}
	if idx, _ := effectstest.GrantedActivatedIndex(t, restored, saga); idx < 0 {
		t.Error("restored: the Saga lost its Construct ability")
	}
	rme := restored.Seats[seat]
	if err := effectstest.TapGrantedMana(t, restored, rme.ID, saga, "C"); err != nil {
		t.Errorf("restored: tapping the Saga for {C}: %v", err)
	}
}

// GrantAbilitiesFor refuses a bundle nobody registered, and one with a
// static a layer-6 grant cannot give (ADR 0093 Decision 10, amended
// 2026-10-10: only a layer-7c one).
func TestGrantAbilitiesForRefusesAnUnknownOrStaticBundle(t *testing.T) {
	if err := effects.CheckGrantMods([]game.Mod{game.GrantAbilitiesMod("nobody/registered-this")}); err == nil {
		t.Error("an unregistered bundle was accepted")
	}
	if err := effects.CheckGrantMods([]game.Mod{game.GrantAbilitiesMod(cardsex.FeignDeathReturn)}); err != nil {
		t.Errorf("Feign Death's bundle was refused: %v", err)
	}
	if err := effects.CheckGrantMods([]game.Mod{{Kind: game.ModGrantAbilities}}); err == nil {
		t.Error("a grant naming nothing was accepted")
	}
	// Phantasmal Image's copy grant has a Triggered slot only; find any
	// registered bundle with a static no layer-6 grant can give to prove
	// the second refusal.
	for key, def := range effects.Defs {
		if strings.HasPrefix(key, game.GrantKeyPrefix) && effects.LayerGrantStaticProblem(def) != "" {
			if err := effects.CheckGrantMods([]game.Mod{game.GrantAbilitiesMod(key)}); err == nil {
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
// whose statics a layer-6 grant can give. A key built at runtime is its caller's risk,
// and Apply refuses it at resolution.
func TestEveryDurationGrantKeyResolves(t *testing.T) {
	files, err := effectstest.FlatGlob("*.go")
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
		src, err := effectstest.FlatReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		af, err := effectstest.FlatParseFile(fset, f, src, 0)
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
		if p := effects.DurationGrantKeyProblem(key); p != "" {
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
