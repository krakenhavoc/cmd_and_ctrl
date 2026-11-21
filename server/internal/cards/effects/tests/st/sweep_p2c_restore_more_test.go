package testsst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_p2c_restore_more_test.go — group C's restore tests that ride on a
// larger board (the villainous-choice table, the double-faced trigger table).

// The Master, Gallifrey's End: the controller's tie-break among the opponents
// with the most life is a keyed player choice carrying the exiled card, so the
// chosen opponent faces the villainous choice over it on the restored table.
func TestSweepP2CMasterTieBreakIsARestorePoint(t *testing.T) {
	g, me, robot := effectstest.MasterBoard(t)
	g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 40, 40, 39

	effectstest.VillainDestroy(g, robot)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	const key = "option-pick/the-master-gallifreys-end-opponent"
	sweepP2COptionKey(t, g, me.ID, key)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweepP2COptionKey(t, g, me.ID, key)
	effectstest.AnswerChoosePlayer(t, g, me.ID, g.Seats[2])
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[2].ID), 0)
	if g.Seats[2].Life != 36 || g.Seats[1].Life != 40 {
		t.Errorf("lives = %d / %d, want the chosen tied opponent at 36 and the other at 40", g.Seats[2].Life, g.Seats[1].Life)
	}
}

// Orvar, the All-Form: the choice among the permanents a spell targeted carries
// their IDs and epochs as data, so the copy is of the one picked on the
// restored table.
func TestSweepP2COrvarCopyChoiceIsARestorePoint(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	effectstest.ApaPush(d.G, d.Caster.ID, d.Caster.ID, effectstest.ApaCreature("Orvar, the All-Form", effectstest.OrvarOracle, 3, 3))
	bear := effectstest.B12Creature(d.G, d.Caster.ID, "My Bear", "Creature — Bear", 2, 2)
	wolf := effectstest.B12Creature(d.G, d.Caster.ID, "My Wolf", "Creature — Wolf", 2, 2)

	effectstest.CastCatalogSpell(t, d.G, "Giant Growth", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: wolf}})
	effectstest.PassPriorityAroundTable(t, d.G)
	const key = "option-pick/orvar-the-all-form-copy"
	pick := sweepP2COptionKey(t, d.G, d.Caster.ID, key)
	wolfIndex := -1
	for i, o := range pick.PickOptions {
		if o.Label == "My Wolf" {
			wolfIndex = i
		}
	}
	if wolfIndex < 0 {
		t.Fatalf("the Wolf is not offered: %+v", pick.PickOptions)
	}

	g := effectstest.RestoreRoundTrip(t, d.G, true)
	caster := g.Seats[0]
	sweepP2COptionKey(t, g, caster.ID, key)
	effectstest.AnswerOptionPick(t, g, caster.ID, wolfIndex)
	effectstest.PassPriorityAroundTable(t, g)
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
