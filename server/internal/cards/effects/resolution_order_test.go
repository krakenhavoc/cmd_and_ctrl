package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_order_test.go — #4019 and #4070 (CR 608.2c): an instruction
// after one that asks something waits for the answer, wherever the
// instructions are written — a spell's OnResolve (ResolveSteps), a modal
// bullet (ModeSteps), a Do inside an intervening "if" (InterveningIf), an
// If inside either — and a table restored while the question is open
// still runs them after the answer. A modal item's next bullet also
// waits for a prevention announcement its earlier bullet's damage owes.

const demonicPactOracle = "19a2f0a0-9e68-4982-a5f5-b77d805befd7"

// Electrolyze — "deals 2 damage divided as you choose among one or two
// targets. Draw a card." — with all 2 on Khârn the Betrayer, whose
// prevention asks its controller for an opponent (#2040). The draw is
// the next instruction, so it waits for the answer, live and on a table
// restored while the question is open.
func TestElectrolyzeDrawWaitsForKharnsQuestion(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := kharnPush(g, me.ID, me.ID)
		d1658Cast(t, g, "Electrolyze", "Instant", electrolyzeOracle, game.CastSpellParams{
			Targets: cardRefs(kharn), Distribution: map[uuid.UUID]int{kharn: 2},
		})
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatal("Khârn's question is not open")
		}
		if me.Hand.Size() != hand {
			t.Fatalf("restored=%v: hand %d with the question open, want %d: the draw ran before the answer", restored, me.Hand.Size(), hand)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		kharnAnswer(t, g, me.ID, third.ID)
		if me.Hand.Size() != hand+1 {
			t.Fatalf("restored=%v: hand %d after the answer, want %d", restored, me.Hand.Size(), hand+1)
		}
		if got := kharnController(t, g, kharn); got != third.ID {
			t.Fatalf("restored=%v: Khârn's controller %s, want %s", restored, got, third.ID)
		}
	}
}

// Tune the Narrative — "Draw a card. You get {E}{E}." — with Stinkweed
// Imp in the graveyard: the draw offers dredge (CR 702.52), and the
// energy waits for the answer, live and on an undo taken while dredge is
// open. (The dredge offer resumes through a closure, so that table is
// not a restore point; the parked step adds nothing to the census.)
func TestTuneTheNarrativeEnergyWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	castCatalogSpell(t, g, "Tune the Narrative", "Instant", eaTune, nil)
	energy := energyOf(me)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if energyOf(me) != energy {
		t.Fatalf("energy %d with dredge open, want %d: the energy came before the draw", energyOf(me), energy)
	}
	if n := g.CaptureSnapshot().Continuations.ChoiceResumeFrames; n != 1 {
		t.Fatalf("census counts %d resume frames, want 1 (the dredge offer alone)", n)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		answerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || energyOf(p) != energy+2 {
			t.Fatalf("Imp in hand %v, energy %d; want true, %d", p.Hand.Contains(imp), energyOf(p), energy+2)
		}
	}
}

// Demonic Pact's first bullet — "This enchantment deals 4 damage to any
// target and you gain 4 life" — aimed at Khârn the Betrayer: the life
// gain is the same bullet's next instruction, so it waits for Khârn's
// question, live and across a strict restore taken while it is open.
func TestDemonicPactLifeGainWaitsForKharnsQuestion(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, third := g.Seats[seat], g.Seats[(seat+2)%len(g.Seats)]
		pushPermanentForTest(g, me.ID, "Demonic Pact", demonicPactOracle, "Enchantment")
		advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
		advanceToUpkeepOf(t, g, seat)
		// Khârn arrives now: it must attack each combat if able, which
		// would stop the walk to the upkeep.
		kharn := kharnPush(g, me.ID, me.ID)
		chooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 0)
		pickCard(t, g, me.ID, kharn)
		life := me.Life
		passPriorityAroundTable(t, g)
		if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatal("Khârn's question is not open")
		}
		if me.Life != life {
			t.Fatalf("restored=%v: life %d with the question open, want %d: the gain ran before the answer", restored, me.Life, life)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			me = g.Seats[seat]
		}
		kharnAnswer(t, g, me.ID, third.ID)
		if me.Life != life+4 {
			t.Fatalf("restored=%v: life %d after the answer, want %d", restored, me.Life, life+4)
		}
		if got := kharnController(t, g, kharn); got != third.ID {
			t.Fatalf("restored=%v: Khârn's controller %s, want %s", restored, got, third.ID)
		}
	}
}

// Imoen, Mystic Trickster — "if you have the initiative, draw a card.
// Draw another card if you've completed a dungeon." — is a Do inside an
// intervening "if" (InterveningIf). With Stinkweed Imp in the graveyard
// the first draw offers dredge, and the second draw waits for the
// answer: it then draws a real card, because the Imp is in the hand.
// Live and on an undo taken while dredge is open (not a restore point,
// for Tune the Narrative's reason).
func TestImoenSecondDrawWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Imoen, Mystic Trickster", imoenOracle, "Legendary Creature — Human Rogue Wizard")
	initCompleteADungeon(me)
	initVenture(t, g, me.ID)
	advanceToEndStepOf(t, g, 0)
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	hand, library := me.Hand.Size(), me.Library.Size()
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Hand.Size() != hand || me.Library.Size() != library {
		t.Fatalf("hand %d, library %d with dredge open; want %d, %d: the second draw ran before the answer",
			me.Hand.Size(), me.Library.Size(), hand, library)
	}
	offers := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement {
			offers++
		}
	}
	if offers != 1 {
		t.Fatalf("%d dredge offers open, want 1: the second draw began before the first was answered", offers)
	}
	if n := g.CaptureSnapshot().Continuations.ChoiceResumeFrames; n != 1 {
		t.Fatalf("census counts %d resume frames, want 1 (the dredge offer alone)", n)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[0]
		answerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || p.Hand.Size() != hand+2 || p.Library.Size() != library-6 {
			t.Fatalf("Imp in hand %v, hand %d, library %d; want true, %d, %d (dredge 5, then one draw)",
				p.Hand.Contains(imp), p.Hand.Size(), p.Library.Size(), hand+2, library-6)
		}
	}
}

// Probes for the two shapes no catalog card writes yet: an If inside a
// spell's ResolveSteps, and a bullet's steps followed by a later bullet.
// Plus #4070's modal spell: the first bullet's one damage instance hits
// Khârn and a player under Selfless Squire's shield, and the second
// bullet looks at the Squire's triggers.
const (
	resolveStepsProbeOracle = "test-4019-resolve-steps-probe"
	modeStepsProbeOracle    = "test-4019-mode-steps-probe"
	modeSquireProbeOracle   = "test-4070-mode-squire-probe"
)

// modeStepsSawHand records the hand size as the probe's second bullet
// begins (-1: not run).
var modeStepsSawHand = -1

func init() {
	Register(Spec{
		OracleID:  resolveStepsProbeOracle,
		Name:      "Resolve Steps Probe",
		OnResolve: ResolveSteps(If{Cond: doWalkAlways{}, Then: []Applier{Scry{N: 1}, DrawCards{N: 1}}}, GainLife{Amount: 1}),
	})
	Register(Spec{
		OracleID: modeStepsProbeOracle,
		Name:     "Mode Steps Probe",
		Modes: ChooseN("Choose two", 2, 2,
			ModeSteps("Scry 1, then draw a card.", nil, Scry{N: 1}, DrawCards{N: 1}),
			ModeSteps("Look at your hand.", nil, Instruction(func(ctx *Context) error {
				for _, p := range ctx.Game.Seats {
					if p.ID == ctx.Controller() {
						modeStepsSawHand = p.Hand.Size()
					}
				}
				return nil
			})),
		),
	})
	Register(Spec{
		OracleID: modeSquireProbeOracle,
		Name:     "Mode Squire Probe",
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("1 damage to target creature and 2 damage to you.", TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, _ int) error { return doWalkProbeKharnAndYou{}.Apply(ctx) }),
			ModeDoing("Look at the Squire.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error { return doWalkProbeSeeSquire{}.Apply(ctx) }),
		),
	})
}

// castProbe casts a test card for the active seat with the given modes
// and targets.
func castProbe(t *testing.T, g *game.Game, name, oracle string, modes []int, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle,
		Owner: active.ID, Controller: active.ID})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
}

// An If inside a spell's ResolveSteps parks one level down: the draw
// inside it and the life gain after it both wait for the scry, across a
// strict restore, and the draw takes the card the scry kept on top.
func TestResolveStepsInsideAnIfWaitForTheScry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedLibrary(me, "Top", "Second")
	castProbe(t, g, "Resolve Steps Probe", resolveStepsProbeOracle, nil, nil)
	hand, life := me.Hand.Size(), me.Life
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatalf("hand %d, life %d with the scry open; want %d, %d", me.Hand.Size(), me.Life, hand, life)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("restored: the scry is not open")
	}
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{ids[0]}, nil); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(ids[1]) || me.Hand.Size() != hand+1 || me.Life != life+1 {
		t.Fatalf("hand %d (Second %v), life %d; want %d (true), %d", me.Hand.Size(), me.Hand.Contains(ids[1]), me.Life, hand+1, life+1)
	}
}

// A bullet's later step runs before the next bullet: the first bullet's
// draw waits for its scry, and the second bullet, which looks at the
// hand as it begins, sees the drawn card. Live and across a strict
// restore taken while the scry is open.
func TestModeStepsRestOfABulletRunsBeforeTheNextBullet(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		ids := seedLibrary(me, "Top", "Second")
		modeStepsSawHand = -1
		castProbe(t, g, "Mode Steps Probe", modeStepsProbeOracle, []int{0, 1}, nil)
		hand := me.Hand.Size()
		passPriorityAroundTable(t, g)
		if me.Hand.Size() != hand || modeStepsSawHand != -1 {
			t.Fatalf("restored=%v: hand %d, second bullet saw %d with the scry open; want %d and not run",
				restored, me.Hand.Size(), modeStepsSawHand, hand)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		c := scryChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("the scry is not open")
		}
		if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{ids[0]}, nil); err != nil {
			t.Fatal(err)
		}
		if !me.Hand.Contains(ids[1]) || modeStepsSawHand != hand+1 {
			t.Fatalf("restored=%v: Second in hand %v, second bullet saw %d; want true and %d", restored, me.Hand.Contains(ids[1]), modeStepsSawHand, hand+1)
		}
	}
}

// #4070: the first bullet's one damage instance hits Khârn the Betrayer
// (who asks) and its controller under Selfless Squire's shield (whose
// CR 615.13 announcement waits behind the question). The second bullet
// waits too, and when it runs the announcement has been made, because
// it belongs to the earlier bullet's damage (CR 615.5, CR 608.2c). Live
// and across a strict restore taken while the question is open.
func TestModalNextBulletWaitsForTheSquiresAnnouncement(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		me, third := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := kharnPush(g, me.ID, me.ID)
		doWalkSquire = castSquire(t, g)
		doWalkSquireSeen = -1
		life := me.Life
		castProbe(t, g, "Mode Squire Probe", modeSquireProbeOracle, []int{0, 1},
			[]game.TargetRef{{Kind: game.TargetCard, ID: kharn, Mode: 0}})
		passPriorityAroundTable(t, g)
		if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatal("Khârn's question is not open")
		}
		if n := squireTriggers(g, doWalkSquire); n != 0 || doWalkSquireSeen != -1 {
			t.Fatalf("%d Squire triggers and the second bullet saw %d with the question open; want 0 and not run", n, doWalkSquireSeen)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		kharnAnswer(t, g, me.ID, third.ID)
		if doWalkSquireSeen != 1 {
			t.Fatalf("restored=%v: the second bullet saw %d Squire triggers, want 1 — the announcement first", restored, doWalkSquireSeen)
		}
		settleOrderingTriggers(t, g)
		if n := countersOn(g, doWalkSquire, game.CounterPlusOne); n != 2 || me.Life != life {
			t.Fatalf("restored=%v: Squire counters %d (want 2), life %d (want %d)", restored, n, me.Life, life)
		}
		if got := kharnController(t, g, kharn); got != third.ID {
			t.Fatalf("restored=%v: Khârn's controller %s, want %s", restored, got, third.ID)
		}
	}
}
