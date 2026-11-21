package testst

import (
	"testing"

	"github.com/google/uuid"

	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The number prompt is the Do's first step, so the Do parks its second
// behind it. The answer runs the continuation, whose scry asks: the
// continuation parks ABOVE the Do (#4162, its rest is the earlier
// instruction), and the number chosen waits in the step record (#4163).
// After the scry the draw happens, then the Do's look, which sees the
// cards drawn. Live, and on a table restored while the scry is open.
func TestThenStepsWaitAndResumeBeforeTheRestOfTheEffect(t *testing.T) {
	for _, restored := range []bool{false, true} {
		effectstest.ThenStepsProbeLook = -1
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		ids := effectstest.SeedLibrary(me, "Top", "Second", "Third")
		hand := me.Hand.Size()
		probe := effectstest.PushCatalogPermanent(g, me.ID, "Then Steps Probe", "Artifact", effectstest.ThenStepsProbeOracle, false)
		if err := effectstest.ActivateOn(t, g, me.ID, probe, 0, uuid.Nil); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.ThenStepsProbeLook != -1 {
			t.Fatal("the Do's second step ran with the number prompt open")
		}
		effectstest.AnswerNumber(t, g, me.ID, 2)
		if effectstest.ScryChoiceFor(g, me.ID) == nil {
			t.Fatal("the continuation's scry is not open")
		}
		if me.Hand.Size() != hand || effectstest.ThenStepsProbeLook != -1 {
			t.Fatalf("hand %d, look %d with the scry open; want %d, -1: a later step ran before the scry",
				me.Hand.Size(), effectstest.ThenStepsProbeLook, hand)
		}
		if restored {
			g = effectstest.RestoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		c := effectstest.ScryChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("restored=%v: the scry is not open", restored)
		}
		// Bottom Top, keep nothing on top.
		if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{ids[0]}, nil); err != nil {
			t.Fatal(err)
		}
		if me.Hand.Size() != hand+2 || !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[2]) {
			var names []string
			for _, c := range me.Hand.Cards {
				names = append(names, c.Name)
			}
			t.Fatalf("restored=%v: hand %d (Second %v, Third %v) %v; want %d, both drawn after the scry bottomed Top",
				restored, me.Hand.Size(), me.Hand.Contains(ids[1]), me.Hand.Contains(ids[2]), names, hand+2)
		}
		if effectstest.ThenStepsProbeLook != hand+2 {
			t.Fatalf("restored=%v: the Do's last step saw hand %d, want %d: it ran before the continuation's draw (-1: never)",
				restored, effectstest.ThenStepsProbeLook, hand+2)
		}
	}
}

// Seize the Spotlight with three fortunes: "you draw a card and create a
// Treasure token" three times. The first draw offers dredge (Stinkweed
// Imp), and the Treasure waits for the answer; the rest of the loop runs
// from the record of fortunes still owed.
func TestSeizeTheSpotlightTreasureWaitsForDredge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	imp := effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	effectstest.CastCatalogSpell(t, g, "Seize the Spotlight", "Sorcery", "3d1b09a3-f151-4218-974b-02bb6159247a", nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	for _, p := range g.Seats[1:] {
		effectstest.AnswerOptionPick(t, g, p.ID, 1) // fortune
	}
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered on the first fortune's draw")
	}
	if n, _ := effectstest.E2Treasures(g, me.ID); n != 0 {
		t.Fatalf("%d Treasures with dredge open, want 0: the Treasure was made before the answer", n)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[0]
		effectstest.AnswerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) {
			t.Fatal("dredge did not return the Imp")
		}
		if n, _ := effectstest.E2Treasures(table, p.ID); n != 3 {
			t.Fatalf("%d Treasures after the answer, want 3", n)
		}
		// The Imp instead of the first card, and the two other fortunes' cards.
		if p.Hand.Size() != hand+3 {
			t.Fatalf("hand %d, want %d", p.Hand.Size(), hand+3)
		}
	}
}

// Scavenging Ooze eats a creature card with Doubling Season and Hardened
// Scales out: the counter waits on the CR 616 order, and the life gain,
// the exile callback's second instruction, waits behind it.
func TestScavengingOozeLifeGainWaitsForTheCounterOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	ooze := effectstest.PushOoze(g, me)
	prey := effectstest.Batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	life := me.Life
	effectstest.OozeEats(t, g, me, ooze, prey)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if me.Life != life {
		t.Fatalf("life %d with the order open, want %d: the gain ran before the counter", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		effectstest.AnswerOrderBySource(t, table, ds, hs)
		if got := effectstest.B12Counter(t, table, ooze, "+1/+1"); got != 3 {
			t.Fatalf("+1/+1 counters %d, want 3 ((1×2)+1, Doubling Season first)", got)
		}
		if p := table.Seats[0]; p.Life != life+1 {
			t.Fatalf("life %d after the order, want %d", p.Life, life+1)
		}
	}
}

// Martial Coup for X=5 with Doubling Season and Ojer Taq out: the
// Soldiers wait on the CR 616 order, so the sweep waits too, and it
// spares every Soldier the instruction made — read from the event log
// after the step's record, not from what the creation returned before
// the answer.
func TestMartialCoupSparesTheSoldiersMadeAfterTheTokenOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	ojer := effectstest.SeedReplacementPermanent(g, cardso.OjerTaqOracleID, "Ojer Taq, Deepest Foundation", me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.B12PlayFromHand(t, g, "Martial Coup", "Sorcery", effectstest.B12MartialCoupOracle, game.CastSpellParams{XValue: 5})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("the sweep ran with the token order open")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		effectstest.AnswerOrderBySource(t, table, ds, ojer)
		if n := effectstest.CountBattlefieldNamed(table, table.Seats[0].ID, "Soldier"); n != 30 {
			t.Fatalf("%d Soldiers survive, want all 30 the instruction made (5×2×3)", n)
		}
		if table.Battlefield.Contains(theirs) {
			t.Fatal("the other creature survived the sweep")
		}
	}
}

// Cut a Deal: the first opponent's draw offers dredge. The caster's draw
// waits for the answer, and counts the opponents who drew from the event
// log after the step's record: a dredge is not a draw.
func TestCutADealCountsTheDrawsAfterTheDredgeAnswer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, first := g.Seats[0], g.Seats[1]
	effectstest.GyCardWithOracle(first, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	effectstest.CastCatalogSpell(t, g, "Cut a Deal", "Sorcery", effectstest.B08CutADealOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, first) == nil {
		t.Fatal("dredge is not offered to the first opponent")
	}
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d with dredge open, want %d: the caster drew before the answer", me.Hand.Size(), hand)
	}
	undo := g.Clone()
	for _, c := range []struct {
		table  *game.Game
		dredge bool
		want   int
	}{{g, true, 2}, {undo, false, 3}} {
		effectstest.AnswerOptional(t, c.table, c.table.Seats[1], c.dredge)
		if got := c.table.Seats[0].Hand.Size() - hand; got != c.want {
			t.Fatalf("dredge=%v: the caster drew %d, want %d", c.dredge, got, c.want)
		}
	}
}
