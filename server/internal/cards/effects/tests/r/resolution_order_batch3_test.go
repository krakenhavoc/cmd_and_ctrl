package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "D1", "D2", "D3", "D4", "D5", "D6")
	imp := effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	jace := effectstest.PushCatalogWalker(g, me.ID, "Jace, the Mind Sculptor", effectstest.JaceTheMindSculptorOracle, 3)
	effectstest.B16Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("the put-back prompt is open with dredge open: it ran before the answer")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		effectstest.AnswerOptional(t, table, p, true)
		pick := effectstest.LatestChooseCardsFor(table, p.ID)
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
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := effectstest.KharnPush(g, me.ID, me.ID)
		sorin := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Sorin Markov", "Legendary Planeswalker — Sorin", effectstest.B39SorinMarkovOracle, 4)
		life := me.Life
		if err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{
			Targets: effectstest.CardRefs(kharn),
		}); err != nil {
			t.Fatalf("restored=%v: activate: %v", restored, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatalf("restored=%v: Khârn's question is not open", restored)
		}
		if me.Life != life {
			t.Fatalf("restored=%v: life %d with the question open, want %d: the gain ran before the answer", restored, me.Life, life)
		}
		if restored {
			g = effectstest.RestoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		effectstest.KharnAnswer(t, g, me.ID, third.ID)
		if me.Life != life+2 {
			t.Fatalf("restored=%v: life %d after the answer, want %d", restored, me.Life, life+2)
		}
	}
}

// Derelict Attic — "When you unlock this door, you draw two cards and you lose
// 2 life." A Room door's trigger is a row's Do: the life loss waits for the
// dredge offer the first draw makes.
func TestDerelictAtticLifeLossWaitsForDredge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	room := effectstest.RoomCardC(me.ID, effectstest.DerelictAtticOracle, "Derelict Attic", "{2}{B}", "Widow's Walk", "{3}{B}", "B")
	effectstest.CastRoomC(t, g, me.ID, room, 1)
	effectstest.SettleOrdering(t, g, me.ID)
	imp := effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	life := me.Life
	effectstest.UnlockDoorC(t, g, me.ID, room.InstanceID, game.DoorLeft)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[0]
		effectstest.AnswerOptional(t, table, p, true)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Tidus, Yuna's Guardian", "Legendary Creature — Human Warrior", effectstest.TdTidusOracle, 3, 3)
	from := effectstest.TdCreature(g, me.ID, "From", 2, 2)
	to := effectstest.TdCreature(g, me.ID, "To", 2, 2)
	effectstest.TdAddCounter(g, me.ID, from, game.CounterPlusOne, 1)

	// Tidus's own combat trigger moves the counter, so Cheer has a creature with one.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, from)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, to)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	imp := effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	effectstest.AttackWith(t, g, opp.ID, to)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if effectstest.ProliferatePrompt(g) != nil {
		t.Fatal("the proliferate prompt is open with dredge open: it ran before the answer")
	}
	effectstest.AnswerOptional(t, g, me, true)
	if !me.Hand.Contains(imp) {
		t.Fatal("dredge did not return the Imp")
	}
	if effectstest.ProliferatePrompt(g) == nil {
		t.Fatal("the proliferate prompt did not open after the answer")
	}
	effectstest.AnswerProliferate(t, g, to)
	if got := effectstest.CountersOn(g, to, game.CounterPlusOne); got != 2 {
		t.Fatalf("the creature has %d counters after the proliferate, want 2", got)
	}
}

// Pawpatch Formation's third mode — "Draw a card. Create a Food token." A
// modal spell whose OnResolve is ResolveSteps, each step guarded by its mode:
// the Food waits for the dredge offer the draw makes, and the other two modes
// do nothing in these steps.
func TestPawpatchFormationFoodWaitsForDredge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	imp := effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	foods := func(table *game.Game) int {
		n := 0
		for _, c := range table.Battlefield.Cards {
			if c.Name == "Food" && c.Controller == me.ID {
				n++
			}
		}
		return n
	}
	effectstest.CastModal(t, g, "Pawpatch Formation", "Instant", pawpatchFormationOracle, []int{2}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if n := foods(g); n != 0 {
		t.Fatalf("%d Foods with dredge open, want 0: the token was made before the answer", n)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		effectstest.AnswerOptional(t, table, p, true)
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
		g := effectstest.NewCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())
		life := caster.Life
		effectstest.CastCatalogSpell(t, g, "Thoughtseize", "Sorcery", effectstest.ThoughtseizeOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.OpenPick(t, g)
		if caster.Life != life {
			t.Fatalf("restored=%v: life %d with the pick open, want %d: the loss ran before the answer", restored, caster.Life, life)
		}
		if restored {
			g = effectstest.RestoreRoundTrip(t, g, true)
			caster, victim = g.Seats[0], g.Seats[1]
			pick = effectstest.OpenPick(t, g)
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
