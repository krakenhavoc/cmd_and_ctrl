package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// sorcery_speed_stack_test.go — the enumerator half of #1352. CR
// 307.1 / CR 117.1a: sorcery timing needs an empty stack, and a
// triggered ability on the stack makes it not empty exactly as a
// spell does (CR 405.1).
//
// The enumerator was not the permissive side for everything: its own
// land-play check read StackMeta, but casts, sorcery-speed activations
// and plot all ask the engine's timing reads, which asked a gate that
// looked at the stack zone alone. So a bot holding priority over its
// own upkeep-style trigger was offered a sorcery, an equip and a plot
// the rules forbid. This pins every one of them against the ONE read
// now, and the #544 half by dispatching everything that IS offered.

const oracleSorcerySpeedPlot = "legal-sorcery-speed-plot-oracle"

// pushTriggerOntoStack puts a triggered ability on the stack the way
// resolution_pause_test.go does: a StackMeta entry and no card, which
// is what a trigger is.
func pushTriggerOntoStack(g *game.Game, owner *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		g.StackMeta[id] = &game.StackItem{
			ID: id, Kind: game.StackItemTriggered,
			Controller: owner.ID, Owner: owner.ID,
			SourceCardID: uuid.New(),
			Label:        "Test trigger — nothing happens",
			Effect:       func(*game.Game, *game.StackItem) error { return nil },
		}
	})
	return id
}

func TestEnumeratorOffersNoSorcerySpeedMoveOverATrigger(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withSpecialActions(t, oracleSorcerySpeedPlot, []game.SpecialAction{{
		Kind: game.SpecialActionPlot, Cost: "{1}", Label: "Plot {1}",
	}})

	ritual := handCard(active, sorcery("Test Sorcery", "{1}"))
	bear := handCard(active, creature("Grizzly Bears", "{1}", 2, 2))
	forest := handCard(active, basic("Forest", "Forest"))
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	plotter := handCard(active, game.Card{
		Name: "Test Plotter", TypeLine: "Creature — Djinn", ManaCost: "{4}{U}",
		Power: 4, Toughness: 3, OracleID: oracleSorcerySpeedPlot,
	})
	battlefieldCard(g, active, game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	sword := battlefieldCard(g, active, game.Card{
		Name: "Bonesplitter", TypeLine: "Artifact — Equipment", OracleID: oracleBonesplitter,
	})
	lands(g, active, "Mountain", "Mountain", 3)
	advanceTo(t, g, game.StepPrecombatMain)

	type row struct {
		name   string
		offers func([]legal.Move) bool
	}
	sorcerySpeed := []row{
		{"the sorcery", func(m []legal.Move) bool { return len(castMovesFor(m, ritual)) > 0 }},
		{"the creature", func(m []legal.Move) bool { return len(castMovesFor(m, bear)) > 0 }},
		{"the land play", func(m []legal.Move) bool { return len(movesOfKindFor(m, legal.KindLand, forest)) > 0 }},
		{"equip (activate only as a sorcery)", func(m []legal.Move) bool { return len(activationsOf(m, sword)) > 0 }},
		{"plot", func(m []legal.Move) bool { return len(specialActionsOf(m, plotter)) > 0 }},
	}

	// Setup: with the stack empty, every one of them is on the list —
	// so an absence below is the trigger's doing, not the fixture's.
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	for _, r := range sorcerySpeed {
		if !r.offers(moves) {
			t.Fatalf("setup: %s is not offered with an empty stack: %v", r.name, labels(moves))
		}
	}

	pushTriggerOntoStack(g, active)
	moves = legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	for _, r := range sorcerySpeed {
		if r.offers(moves) {
			t.Errorf("%s is offered over a trigger on the stack (CR 307.1)", r.name)
		}
	}
	if len(castMovesFor(moves, bolt)) == 0 {
		t.Errorf("Lightning Bolt is instant-speed and must still be offered over a trigger: %v", labels(moves))
	}
	if !hasLabel(moves, "Pass priority") {
		t.Errorf("the seat holding priority over a trigger must be able to pass: %v", labels(moves))
	}
}
