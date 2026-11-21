package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_frame_restore_test.go — #4361 S14a: real catalog cards whose
// triggers pause in their announcement, restored strictly mid-prompt
// through JSON, and answered on the restored table.

// Miracle's "you may reveal it" (CR 702.94a) is asked before the stack:
// the one production "you may" that holds a trigger frame. A table
// waiting on it is a restore point, and the restored "yes" still
// reveals and grants the miracle cast.
func TestMiracleRevealPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.PushPermanentForTest(g, g.Seats[2].ID, "Bear", "", "Creature — Bear")
	term := effectstest.DrawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", effectstest.TerminusOracle)
	if effectstest.MiraclePrompt(g, g.Seats[1].ID) == nil {
		t.Fatal("no miracle prompt to restore across")
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me := g.Seats[1]
	effectstest.RevealAndResolve(t, g, me)
	if !effectstest.HasMiracleGrant(me, term) {
		t.Fatal("the restored reveal granted no miracle cast")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, term, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: game.AltCostKeyMiracle,
	}); err != nil {
		t.Fatalf("the miracle cast after the restore: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("Terminus left the creature on the battlefield")
	}
}

// Charming Prince's "choose one" (CR 603.3c), restored strictly while
// the mode is open. The life bullet resolves on the restored table.
func TestCharmingPrinceModePickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastAndResolveCreature(t, g, "Charming Prince", "Creature — Human Noble", effectstest.CharmingPrinceOracle)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, g.Seats[0].ID) == nil {
		t.Fatal("no mode_pick to restore across")
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me := g.Seats[0]
	life := me.Life
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	// The exile bullet has no other creature to target, so it was never
	// offered (CR 603.3d).
	if c == nil || len(c.ModeOptionIndex) != 2 {
		t.Fatalf("the restored mode pick is %+v, want the two untargeted bullets", c)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick(life) after the restore: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life after the restored gain-3 bullet: %d, want %d", me.Life, life+3)
	}
}

// The targeted bullet, after the restore: the answer continues into the
// CR 603.3d target walk from the rebuilt row, and the exile happens.
func TestCharmingPrinceTargetedModeAfterARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pal := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pal", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastAndResolveCreature(t, g, "Charming Prince", "Creature — Human Noble", effectstest.CharmingPrinceOracle)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("the restored table has no mode_pick")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("ResolveModePick(exile) after the restore: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, pal)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pal) || !g.Exile.Contains(pal) {
		t.Error("the chosen creature should be in exile after the restored exile bullet")
	}
}
