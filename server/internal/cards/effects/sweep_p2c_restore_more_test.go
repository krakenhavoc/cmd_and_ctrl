package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_p2c_restore_more_test.go — group C's restore tests that ride on a
// larger board (the villainous-choice table, the double-faced trigger table).

// The Master, Gallifrey's End: the controller's tie-break among the opponents
// with the most life is a keyed player choice carrying the exiled card, so the
// chosen opponent faces the villainous choice over it on the restored table.
func TestSweepP2CMasterTieBreakIsARestorePoint(t *testing.T) {
	g, me, robot := masterBoard(t)
	g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 40, 40, 39

	villainDestroy(g, robot)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	const key = "option-pick/the-master-gallifreys-end-opponent"
	sweepP2COptionKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweepP2COptionKey(t, g, me.ID, key)
	answerChoosePlayer(t, g, me.ID, g.Seats[2])
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[2].ID), 0)
	if g.Seats[2].Life != 36 || g.Seats[1].Life != 40 {
		t.Errorf("lives = %d / %d, want the chosen tied opponent at 36 and the other at 40", g.Seats[2].Life, g.Seats[1].Life)
	}
}

// Orvar, the All-Form: the choice among the permanents a spell targeted carries
// their IDs and epochs as data, so the copy is of the one picked on the
// restored table.
func TestSweepP2COrvarCopyChoiceIsARestorePoint(t *testing.T) {
	d := newDBOTable(t)
	apaPush(d.g, d.caster.ID, d.caster.ID, apaCreature("Orvar, the All-Form", orvarOracle, 3, 3))
	bear := b12Creature(d.g, d.caster.ID, "My Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(d.g, d.caster.ID, "My Wolf", "Creature — Wolf", 2, 2)

	castCatalogSpell(t, d.g, "Giant Growth", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: wolf}})
	passPriorityAroundTable(t, d.g)
	const key = "option-pick/orvar-the-all-form-copy"
	pick := sweepP2COptionKey(t, d.g, d.caster.ID, key)
	wolfIndex := -1
	for i, o := range pick.PickOptions {
		if o.Label == "My Wolf" {
			wolfIndex = i
		}
	}
	if wolfIndex < 0 {
		t.Fatalf("the Wolf is not offered: %+v", pick.PickOptions)
	}

	g := restoreRoundTrip(t, d.g, true)
	caster := g.Seats[0]
	sweepP2COptionKey(t, g, caster.ID, key)
	answerOptionPick(t, g, caster.ID, wolfIndex)
	passPriorityAroundTable(t, g)
	wolves, bears := 0, 0
	for _, c := range g.BattlefieldCardsForEffect() {
		switch c.Name {
		case "My Wolf":
			wolves++
		case "My Bear":
			bears++
		}
	}
	if wolves != 2 || bears != 1 {
		t.Errorf("Wolves = %d, Bears = %d, want the picked Wolf copied (2) and the Bear not (1)", wolves, bears)
	}
}
