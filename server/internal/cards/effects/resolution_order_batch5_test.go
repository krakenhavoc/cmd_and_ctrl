package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_order_batch5_test.go — #4155 (CR 608.2c): the fifth batch of
// hand-written bodies written as steps. It adds a real card for each shape
// batch 5 converted, each with the prompt open:
//
//   - a spell's ResolveSteps on a CR 616 counter order, the later sentences
//     waiting (Kick in the Door);
//   - a spell's ResolveSteps whose X is read by two steps, on a dredge offer
//     (Skeletal Scrying);
//   - a mode's ModeSteps behind a dredge offer (Promise of Power);
//   - a planeswalker ability's Do, the second step opening a prompt of its own
//     once the dredge offers are answered (Jace, Multiverse Architect);
//   - a continuation run from a vote's result, the record carrying the tallies
//     (Truth or Consequences' draw before its damage).
//
// A dredge offer and a CR 616 order resume through closures, so a table that
// holds one is not a restore point; an undo (a clone) stands in for the restore
// there, as in resolution_order_batch1_test.go.

const promiseOfPowerOracle = "fde459b2-f45f-40d2-b328-58f65440f494"

// Kick in the Door — "Put a +1/+1 counter on target creature. It gains haste
// until end of turn and can't be blocked by Walls this turn. Venture into the
// dungeon." With Doubling Season and Hardened Scales out the counter waits on
// the CR 616 order, and the haste waits behind it.
func TestKickInTheDoorHasteWaitsForTheCounterOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	castCatalogSpell(t, g, "Kick in the Door", "Sorcery", kickInTheDoorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Fatal("the creature has haste with the order open: the keyword was granted before the counter")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		answerOrderBySource(t, table, ds, hs)
		if got := b12Counter(t, table, bear, "+1/+1"); got != 3 {
			t.Fatalf("+1/+1 counters %d, want 3 ((1×2)+1, Doubling Season first)", got)
		}
		if !hasKeywordOnBattlefield(t, table, bear, "haste") {
			t.Fatal("the creature did not gain haste after the order")
		}
	}
}

// Skeletal Scrying — "Draw X cards and lose X life." The life loss waits for
// the dredge offer the first draw makes, and reads the same X.
func TestSkeletalScryingLifeLossWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fuel := exileXFuel(me, "Sorcery", 3)
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := me.Life
	if _, err := castWithTapParams(t, g, "Skeletal Scrying", "Instant", "{X}{B}", skeletalScryingOracle,
		game.CastSpellParams{XValue: 2, ExileIDs: fuel[:2]}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	for optionalOffer(g, me) != nil {
		answerOptional(t, g, me, false)
	}
	g.SettleResolution()
	if me.Life != life-2 {
		t.Fatalf("life %d after the draws, want %d", me.Life, life-2)
	}
}

// Promise of Power's first mode — "You draw five cards and you lose 5 life."
// The life loss waits for the dredge offers the draws make.
func TestPromiseOfPowerLifeLossWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := me.Life
	castModal(t, g, "Promise of Power", "Sorcery", promiseOfPowerOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		for optionalOffer(table, p) != nil {
			answerOptional(t, table, p, false)
		}
		table.SettleResolution()
		if p.Life != life-5 {
			t.Fatalf("life %d after the draws, want %d", p.Life, life-5)
		}
	}
}

// Jace, Multiverse Architect's +1 — "Draw two cards, then put a card from your
// hand on the bottom of your library." The question waits for the dredge offers
// the draws make.
func TestJaceMultiverseArchitectBottomWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	jace := rfpbWalker(g, me.ID, "Jace, Multiverse Architect", "Legendary Planeswalker — Jace", rfpbJaceMultiverseOracle, 4)
	b16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("the bottom prompt is open with dredge open: it ran before the answer")
	}
	for optionalOffer(g, me) != nil {
		answerOptional(t, g, me, false)
	}
	g.SettleResolution()
	if latestChooseCardsFor(g, me.ID) == nil {
		t.Fatal("the bottom prompt did not open after the draws")
	}
}

// Truth or Consequences — "You draw cards equal to the number of truth votes.
// Then choose an opponent at random. Truth or Consequences deals 3 damage to
// that player for each consequences vote." The vote's continuation is steps:
// the damage waits for the dredge offer the draw makes, and reads the
// tallies from the step record.
func TestTruthOrConsequencesDamageWaitsForTheDrawsDredgeOffer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	var lives []int
	for _, p := range g.Seats {
		lives = append(lives, p.Life)
	}
	castTruthOrConsequences(t, g)
	// Every player's secret ballot: truth from me, consequences from the rest.
	for _, p := range voteOrder(g) {
		word := "consequences"
		if p.ID == me.ID {
			word = "truth"
		}
		voteFor(t, g, p.ID, word)
	}
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered on the draw")
	}
	for i, p := range g.Seats {
		if p.Life != lives[i] {
			t.Fatalf("seat %d life %d with dredge open, want %d: the damage ran before the answer", i, p.Life, lives[i])
		}
	}
	for optionalOffer(g, me) != nil {
		answerOptional(t, g, me, false)
	}
	g.SettleResolution()
	hurt := 0
	for i, p := range g.Seats {
		if p.Life < lives[i] {
			hurt += lives[i] - p.Life
		}
	}
	if hurt != 9 {
		t.Fatalf("%d damage dealt after the draw, want 9 (three consequences votes of 3)", hurt)
	}
}
