package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// obeka_splitter_of_seconds_test.go — #5405: Obeka's combat damage adds
// that many upkeep steps after the combat phase, each a beginning phase
// with only its upkeep (CR 500.10), and "at the beginning of your
// upkeep" fires in every one.

const (
	obekaSplitterOracle  = "1d15a06f-5471-48f2-8c4a-9b0886bef5af"
	obekaPhyrexianArena  = "ee579a32-a048-4335-b966-231ba731cdea"
	obekaTestCreatureTyp = "Legendary Creature — Ogre Warlock"
)

// obekaUpkeepCount counts the upkeep steps still to come this turn.
func obekaUpkeepCount(g *game.Game) int {
	var upcoming []game.PlannedStep
	g.ReadSnapshot(func() { upcoming = g.UpcomingStepsForEffect() })
	n := 0
	for _, p := range upcoming {
		if p.Step == game.StepUpkeep {
			n++
		}
	}
	return n
}

// obekaRunAddedUpkeep steps from wherever the cursor is into the next
// step, which must be an added upkeep of the same turn, resolves the
// Arena trigger there and checks it drew and cost 1 life.
func obekaRunAddedUpkeep(t *testing.T, g *game.Game, seq int, me *game.Player, which string) {
	t.Helper()
	life, hand := me.Life, me.Hand.Size()
	next := stepOnce(t, g)
	if next.Seq != seq || next.Step != game.StepUpkeep || next.Phase != game.PhaseBeginning {
		t.Fatalf("%s: cursor at %+v, want an added upkeep of turn %d", which, next, seq)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Errorf("%s: priority is with seat %d, want the active player (CR 503.1)", which, g.Turn.PriorityHolder)
	}
	if arena := findBattlefieldByName(g, "Phyrexian Arena"); triggerOnStack(g, arena) == nil {
		t.Fatalf("%s: the \"at the beginning of your upkeep\" trigger is not on the stack", which)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-1 || me.Hand.Size() != hand+1 {
		t.Errorf("%s: life %d→%d, hand %d→%d; want the Arena's draw and 1 life, and no draw step",
			which, life, me.Life, hand, me.Hand.Size())
	}
}

func TestObekaCombatDamageGivesThatManyUpkeepSteps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	obeka := pushDiesCreatureForTest(g, me.ID, "Obeka, Splitter of Seconds", obekaSplitterOracle, obekaTestCreatureTyp, 3, 5)
	pushPermanentForTest(g, me.ID, "Phyrexian Arena", obekaPhyrexianArena, "Enchantment")
	tapped := pushVanillaCreature(g, me.ID, "Tapped Thing", 1, 1)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })

	oppLife := opp.Life
	attackWith(t, g, opp.ID, obeka)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Fatalf("setup: Obeka dealt %d combat damage, want 3", oppLife-opp.Life)
	}
	if n := obekaUpkeepCount(g); n != 3 {
		t.Fatalf("after 3 combat damage the turn has %d upkeep steps still to come, want 3", n)
	}
	seq := g.Turn.Seq
	if next := stepOnce(t, g); next.Step != game.StepEndCombat {
		t.Fatalf("after combat damage: %+v, want the end of combat step first", next)
	}

	obekaRunAddedUpkeep(t, g, seq, me, "first added upkeep")
	if !isTapped(g, tapped) {
		t.Errorf("an added upkeep's untap step happened: CR 500.10 skips it")
	}

	// The rest of the queue is on the restore point: the turn plan.
	g = restoreThroughJSON(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	if n := obekaUpkeepCount(g); n != 2 {
		t.Fatalf("after a restore %d upkeep steps are still to come, want 2", n)
	}
	obekaRunAddedUpkeep(t, g, seq, me, "second added upkeep (restored)")
	obekaRunAddedUpkeep(t, g, seq, me, "third added upkeep (restored)")

	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain || next.Seq != seq {
		t.Errorf("after the added upkeeps: %+v, want this turn's postcombat main phase", next)
	}
}
