package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_order_batch3_test.go — #4155 (CR 608.2c): the third batch of
// hand-written bodies written as steps. Batches 1 and 2 covered a spell's
// ResolveSteps, a row's Do on a triggered or activated ability, ModeSteps and
// SpreeModeSteps. These add the shapes batch 3 introduced, each with a real
// card:
//
//   - a loyalty ability's Do on a dredge offer (Jace, the Mind Sculptor's 0);
//   - a loyalty ability's Do on Khârn's question, live and across a restore
//     (Sorin Markov's +2);
//   - a Room door's trigger Do on a dredge offer (Derelict Attic);
//   - a triggered ability whose Do is an If, with a once-a-turn mark set
//     before the draw (Tidus, Yuna's Guardian's Cheer);
//   - a modal spell's ResolveSteps, each step guarded by its mode (Pawpatch
//     Formation);
//   - ResolveSteps on a revealed-hand choice, live and across a restore
//     (Thoughtseize's life loss).
//
// A dredge offer resumes through a closure, so the table that holds one is not
// a restore point; an undo (a clone) stands in for the restore there, as in
// resolution_order_batch1_test.go. Khârn's question and the revealed-hand pick
// are data.

const pawpatchFormationOracle = "fa99cf82-ebf1-4526-ab9c-b24eb2970f2a"

// Jace, the Mind Sculptor's 0 — "Draw three cards, then put two cards from
// your hand on top of your library in any order." The put waits for the dredge
// offer the first draw makes, and only then asks which two cards to put back.
func TestJaceZeroPutBackWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "D1", "D2", "D3", "D4", "D5", "D6")
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	jace := pushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", jaceTheMindSculptorOracle, 3)
	b16Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("the put-back prompt is open with dredge open: it ran before the answer")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		answerOptional(t, table, p, true)
		pick := latestChooseCardsFor(table, p.ID)
		if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
			t.Fatalf("after the answer the put-back prompt is %+v, want one for exactly two cards", pick)
		}
		if !p.Hand.Contains(imp) {
			t.Fatal("dredge did not return the Imp")
		}
	}
}

// Sorin Markov's +2 — "deals 2 damage to any target and you gain 2 life" —
// aimed at Khârn the Betrayer, whose prevention asks its controller for an
// opponent (#2040). The life gain is the next instruction, so it waits for the
// answer, live and on a table restored while the question is open.
func TestSorinMarkovLifeGainWaitsForKharnsQuestion(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		toMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := kharnPush(g, me.ID, me.ID)
		sorin := pushCatalogWalkerWithLine(g, me.ID, "Sorin Markov", "Legendary Planeswalker — Sorin", b39SorinMarkovOracle, 4)
		life := me.Life
		if err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{
			Targets: cardRefs(kharn),
		}); err != nil {
			t.Fatalf("restored=%v: activate: %v", restored, err)
		}
		passPriorityAroundTable(t, g)
		if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatalf("restored=%v: Khârn's question is not open", restored)
		}
		if me.Life != life {
			t.Fatalf("restored=%v: life %d with the question open, want %d: the gain ran before the answer", restored, me.Life, life)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		kharnAnswer(t, g, me.ID, third.ID)
		if me.Life != life+2 {
			t.Fatalf("restored=%v: life %d after the answer, want %d", restored, me.Life, life+2)
		}
	}
}

// Derelict Attic — "When you unlock this door, you draw two cards and you lose
// 2 life." A Room door's trigger is a row's Do: the life loss waits for the
// dredge offer the first draw makes.
func TestDerelictAtticLifeLossWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	room := roomCardC(me.ID, derelictAtticOracle, "Derelict Attic", "{2}{B}", "Widow's Walk", "{3}{B}", "B")
	castRoomC(t, g, me.ID, room, 1)
	settleOrdering(t, g, me.ID)
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := me.Life
	unlockDoorC(t, g, me.ID, room.InstanceID, game.DoorLeft)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[0]
		answerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || p.Life != life-2 {
			t.Fatalf("Imp in hand %v, life %d; want true, %d", p.Hand.Contains(imp), p.Life, life-2)
		}
	}
}

// Tidus, Yuna's Guardian's Cheer — "draw a card and proliferate", once each
// turn. The row's Do is an If: the once-a-turn check is made as the ability
// begins, the turn's mark is set, and the proliferate waits for the dredge
// offer the draw makes.
func TestTidusCheerProliferateWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tidus, Yuna's Guardian", "Legendary Creature — Human Warrior", tdTidusOracle, 3, 3)
	from := tdCreature(g, me.ID, "From", 2, 2)
	to := tdCreature(g, me.ID, "To", 2, 2)
	tdAddCounter(g, me.ID, from, game.CounterPlusOne, 1)

	// Tidus's own combat trigger moves the counter, so Cheer has a creature with one.
	advanceTo(t, g, game.StepBeginCombat)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, from)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, to)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	attackWith(t, g, opp.ID, to)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if proliferatePrompt(g) != nil {
		t.Fatal("the proliferate prompt is open with dredge open: it ran before the answer")
	}
	answerOptional(t, g, me, true)
	if !me.Hand.Contains(imp) {
		t.Fatal("dredge did not return the Imp")
	}
	if proliferatePrompt(g) == nil {
		t.Fatal("the proliferate prompt did not open after the answer")
	}
	answerProliferate(t, g, to)
	if got := countersOn(g, to, game.CounterPlusOne); got != 2 {
		t.Fatalf("the creature has %d counters after the proliferate, want 2", got)
	}
}

// Pawpatch Formation's third mode — "Draw a card. Create a Food token." A
// modal spell whose OnResolve is ResolveSteps, each step guarded by its mode:
// the Food waits for the dredge offer the draw makes, and the other two modes
// do nothing in these steps.
func TestPawpatchFormationFoodWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	foods := func(table *game.Game) int {
		n := 0
		for _, c := range table.Battlefield.Cards {
			if c.Name == "Food" && c.Controller == me.ID {
				n++
			}
		}
		return n
	}
	castModal(t, g, "Pawpatch Formation", "Instant", pawpatchFormationOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if n := foods(g); n != 0 {
		t.Fatalf("%d Foods with dredge open, want 0: the token was made before the answer", n)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		answerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || foods(table) != 1 {
			t.Fatalf("Imp in hand %v, Foods %d; want true, 1", p.Hand.Contains(imp), foods(table))
		}
	}
}

// Thoughtseize — "Target player reveals their hand. You choose a nonland card
// from it. Exile that card. You lose 2 life." The life loss is the instruction
// after the choice, so it waits for the pick, live and on a table restored
// while the pick is open.
func TestThoughtseizeLifeLossWaitsForThePick(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		ids := revealHand(victim, rhForest(), rhBolt())
		life := caster.Life
		castCatalogSpell(t, g, "Thoughtseize", "Sorcery", thoughtseizeOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
		passPriorityAroundTable(t, g)
		pick := openPick(t, g)
		if caster.Life != life {
			t.Fatalf("restored=%v: life %d with the pick open, want %d: the loss ran before the answer", restored, caster.Life, life)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			caster, victim = g.Seats[0], g.Seats[1]
			pick = openPick(t, g)
		}
		if err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{ids[1]}); err != nil {
			t.Fatalf("restored=%v: taking the Bolt: %v", restored, err)
		}
		// The actions layer settles the resolution after every answer; the life loss
		// behind the pick runs there.
		g.SettleResolution()
		if caster.Life != life-2 {
			t.Fatalf("restored=%v: life %d after the pick, want %d", restored, caster.Life, life-2)
		}
		if !victim.Graveyard.Contains(ids[1]) {
			t.Fatalf("restored=%v: the Bolt was not discarded", restored)
		}
	}
}
