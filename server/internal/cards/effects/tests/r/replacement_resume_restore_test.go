package testsr

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- dredge ------------------------------------------------------------

// A dredge offer on the first draw of "draw a card, then discard a card":
// the restored offer dredges (a "yes") or draws (a "no"), and the discard
// the draw's continuation owes opens afterwards either way.
func TestADredgeOfferIsARestorePoint(t *testing.T) {
	for _, dredge := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		p := g.Seats[1]
		imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
		libBefore, handBefore := p.Library.Size(), p.Hand.Size()
		effectstest.StartLoot(g, p, 1, 1)
		if effectstest.OptionalOffer(g, p) == nil {
			t.Fatal("dredge was not offered")
		}

		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDraw)
		p = g.Seats[1]
		offer := effectstest.OptionalOffer(g, p)
		if offer == nil || offer.DredgeOffer() != 5 || offer.Source != imp {
			t.Fatalf("the restored offer is %+v, want the Imp's dredge 5", offer)
		}
		if effectstest.DiscardPromptFor(g, p) != nil {
			t.Fatal("the restored table opened the discard before the draw")
		}
		effectstest.AnswerOptional(t, g, p, dredge)

		if dredge {
			if !effectstest.ZoneHas(p.Hand, imp) || p.Library.Size() != libBefore-5 || p.Hand.Size() != handBefore+1 {
				t.Errorf("dredged: Imp in hand %v, library %d (want %d), hand %d (want %d)",
					effectstest.ZoneHas(p.Hand, imp), p.Library.Size(), libBefore-5, p.Hand.Size(), handBefore+1)
			}
		} else if effectstest.ZoneHas(p.Hand, imp) || p.Library.Size() != libBefore-1 || p.Hand.Size() != handBefore+1 {
			t.Errorf("declined: library %d (want %d), hand %d (want %d)", p.Library.Size(), libBefore-1, p.Hand.Size(), handBefore+1)
		}
		if effectstest.DiscardPromptFor(g, p) == nil {
			t.Errorf("dredge %v: the discard the draw owed did not open after the restored answer", dredge)
		}
	}
}

// --- a CR 616.1 ordering prompt ----------------------------------------

// Thought Reflection and a dredge card both apply to one draw, so the
// drawer orders them. Restored at the ordering prompt, and again at the
// first descendant's dredge offer (the second draw rides the frame as
// data), the table finishes both draws as the live one does.
func TestAnOrderingPromptOverADrawIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle)
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	thug := effectstest.GyCardWithOracle(p, "Golgari Thug", "Creature — Human Warrior", effectstest.GolgariThugOracle)
	before, libBefore := p.Hand.Size(), p.Library.Size()
	effectstest.DrawN(g, p, 1)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, p.ID) == nil {
		t.Fatal("no ordering prompt")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDraw)
	p = g.Seats[1]
	effectstest.OrderFirst(t, g, p, "Thought Reflection")

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDraw)
	p = g.Seats[1]
	effectstest.AnswerOptional(t, g, p, true)
	effectstest.AnswerOptional(t, g, p, true)
	if !p.Hand.Contains(imp) || !p.Hand.Contains(thug) || p.Hand.Size() != before+2 || p.Library.Size() != libBefore-9 {
		t.Fatalf("each expanded draw dredges its card: hand %d (want %d), library %d (want %d)",
			p.Hand.Size(), before+2, p.Library.Size(), libBefore-9)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("questions left open: %+v", g.PendingChoices)
	}
}

// The restored descendant's offer, declined: the descendant draws, and
// Thought Reflection, already applied to it before the restore (CR
// 614.5), does not double it again. Hand +2 and library -2 is the two
// draws the doubler made; a lost mark would double the declined one.
func TestARestoredDrawKeepsItsOncePerEventMarks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle)
	effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	before, libBefore := p.Hand.Size(), p.Library.Size()
	effectstest.DrawN(g, p, 1)
	effectstest.OrderFirst(t, g, p, "Thought Reflection")

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDraw)
	p = g.Seats[1]
	effectstest.AnswerOptional(t, g, p, false)
	effectstest.AnswerOptional(t, g, p, false)
	if p.Hand.Size() != before+2 || p.Library.Size() != libBefore-2 {
		t.Errorf("hand %d (want %d), library %d (want %d)", p.Hand.Size(), before+2, p.Library.Size(), libBefore-2)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("questions left open: %+v", g.PendingChoices)
	}
}

// The object serials a dredge offer's ID packs (#3482) survive the
// restore. The Imp is gathered first, on a declined draw, and Thought
// Reflection after it; a restored game that minted serials afresh would
// walk the battlefield first and give the Imp another one, and the
// restored ordering prompt would name an effect nothing gathers.
func TestARestoredFrameKeepsItsObjectSerials(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	imp := effectstest.GyCardWithOracle(p, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	effectstest.DrawN(g, p, 1)
	effectstest.AnswerOptional(t, g, p, false)
	effectstest.BfCardWithOracle(g, p, "Thought Reflection", "Enchantment", effectstest.ThoughtReflectionOracle)
	before, libBefore := p.Hand.Size(), p.Library.Size()
	effectstest.DrawN(g, p, 1)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDraw)
	p = g.Seats[1]
	order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, p.ID)
	if order == nil {
		t.Fatal("the restored table has no ordering prompt")
	}
	for _, id := range order.ReplacementEffectIDs {
		if label, _ := g.ReplacementOptionMetaForEffect(id); label == "" {
			t.Errorf("the restored prompt lists effect %d, which names nothing", id)
		}
	}
	// Dredge first, and its "may" taken: the Imp comes back and nothing
	// is drawn, so Thought Reflection has nothing left to replace.
	effectstest.OrderFirst(t, g, p, "Dredge")
	effectstest.AnswerOptional(t, g, p, true)
	if !p.Hand.Contains(imp) || p.Hand.Size() != before+1 || p.Library.Size() != libBefore-5 {
		t.Errorf("dredge first: Imp in hand %v, hand %d (want %d), library %d (want %d)",
			p.Hand.Contains(imp), p.Hand.Size(), before+1, p.Library.Size(), libBefore-5)
	}
}

// --- Retriever Phoenix ---------------------------------------------------

// The Phoenix's "return it instead of learning?" is a restore point. A
// restored "yes" returns the Phoenix and learns nothing; a restored "no"
// learns, whose first question is the discard.
func TestRetrieverPhoenixQuestionIsARestorePoint(t *testing.T) {
	for _, ret := range []bool{true, false} {
		g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.LessonCard("Environmental Sciences"))
		me := g.Seats[0]
		phoenix := gyPhoenix(me)
		effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if learnOffer(g, me) == nil {
			t.Fatal("no offer")
		}

		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeKeywordAction)
		me = g.Seats[0]
		offer := learnOffer(g, me)
		if offer == nil || offer.Source != phoenix {
			t.Fatalf("the restored offer is %+v, want the Phoenix's", offer)
		}
		effectstest.AnswerOptional(t, g, me, ret)
		effectstest.PassPriorityAroundTable(t, g)

		if ret {
			if !phoenixInPlay(g, phoenix) || effectstest.ZoneHas(me.Graveyard, phoenix) {
				t.Error("a restored yes did not return the Phoenix")
			}
			if effectstest.LearnDiscardPrompt(g, me) != nil || effectstest.LearnLessonPrompt(g, me) != nil {
				t.Error("a replaced learn asked its own questions")
			}
		} else {
			if phoenixInPlay(g, phoenix) {
				t.Error("a restored no returned the Phoenix")
			}
			if effectstest.LearnDiscardPrompt(g, me) == nil {
				t.Error("a restored no did not learn")
			}
		}
	}
}

// --- riot and unleash ----------------------------------------------------

// Riot's question on a creature spell's entry is a restore point: the
// paused entry carries the resolving spell, and the restored answer
// lands the creature with what was chosen.
func TestARiotQuestionIsARestorePoint(t *testing.T) {
	for _, counter := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Zhur-Taa Goblin", TypeLine: "Creature — Goblin Berserker",
			Power: 2, Toughness: 2, Keywords: []string{game.KeywordRiot}, Owner: me.ID, Controller: me.ID,
		})
		effectstest.AdvanceToMain(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.PendingOfKind(g, game.PendingChoiceEntryRiot) == nil {
			t.Fatal("riot was not asked")
		}

		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMove)
		if g.Battlefield.Contains(id) {
			t.Fatal("the restored table landed the creature before the answer")
		}
		effectstest.AnswerRiotFor(t, g, counter)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(id) {
			t.Fatal("the restored answer did not land the creature")
		}
		haste, n := effectstest.HasKeywordOnBattlefield(t, g, id, "haste"), effectstest.CounterCount(g, id, game.CounterPlusOne)
		if counter && (n != 1 || haste) {
			t.Errorf("counter answer: %d counters, haste %v; want one counter and no haste", n, haste)
		}
		if !counter && (n != 0 || !haste) {
			t.Errorf("haste answer: %d counters, haste %v; want haste alone", n, haste)
		}
	}
}

// Unleash, granted by Tesak to a Dog cast under it, asks the ordinary
// "may"; a restored "yes" puts the counter on and the Dog can't block.
func TestAnUnleashQuestionIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RiotPush(g, me.ID, "Tesak, Judith's Hellhound", "Legendary Creature — Elemental Dog", unleashTesakOracle, 3, 3, 0)
	dog := effectstest.CastCatalogSpell(t, g, "Hound", "Creature — Dog", "", nil)
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMove)
	answerUnleash(t, g, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, dog, game.CounterPlusOne) != 1 {
		t.Fatalf("the restored unleash: %d counters, want 1", effectstest.CounterCount(g, dog, game.CounterPlusOne))
	}
	if c := effectstest.FindBattlefieldCardByID(g, dog); !game.Restricted(c, game.CantBlock) {
		t.Error("an unleashed Dog with its counter can block")
	}
}

// --- the keys a frame names ----------------------------------------------

// Every "then discard" question a caller hands drawThenDiscard is
// registered at init, so a freshly booted binary knows the key a paused
// draw's frame names.
func TestEveryDrawThenDiscardQuestionIsRegisteredAtInit(t *testing.T) {
	files, err := effectstest.FlatGlob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	calls := 0
	for _, path := range files {
		if filepath.Ext(path) != ".go" || len(path) > 8 && path[len(path)-8:] == "_test.go" {
			continue
		}
		f, err := effectstest.FlatParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "DrawThenDiscard" || len(call.Args) == 0 {
				return true
			}
			calls++
			lit, ok := call.Args[len(call.Args)-1].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				t.Errorf("%s: DrawThenDiscard's question is not a string literal, so it cannot be registered at init", fset.Position(call.Pos()))
				return true
			}
			q, _ := strconv.Unquote(lit.Value)
			if !slices.Contains(effects.DrawThenDiscardQuestions, q) || !game.KnownDrawThen("discard:"+q) {
				t.Errorf("%s: the question %q is not in DrawThenDiscardQuestions", fset.Position(call.Pos()), q)
			}
			return true
		})
	}
	if calls == 0 {
		t.Fatal("found no DrawThenDiscard calls; the scan is broken")
	}
}
