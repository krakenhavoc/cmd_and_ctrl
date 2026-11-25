package testss82v

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scuttlemutt_test.go — Scuttlemutt (#6146): "{T}: Target creature becomes
// the color or colors of your choice until end of turn."

const scuttlemuttOracle = "564c785b-c2fb-4857-a5bf-a14ae2cf6c93"

func TestScuttlemuttIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(scuttlemuttOracle)
	if !ok || spec.Name != "Scuttlemutt" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Scuttlemutt spec = %+v, %v", spec, ok)
	}
}

// CR 105.3: the chosen colours replace the old ones; the creature is then
// white and blue for a protection-from-blue check, and it is red again after
// the cleanup step (CR 514.2).
func TestScuttlemuttMakesACreatureTheChosenColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mutt := effectstest.B27Push(g, me.ID, "Scuttlemutt", "Artifact Creature — Scarecrow", scuttlemuttOracle, "{3}", 2, 2)
	ogre := effectstest.B27Push(g, me.ID, "Ogre", "Creature — Ogre", "", "{2}{R}", 3, 3, "R")
	if err := effectstest.InvB1Activate(g, me.ID, mutt, 0, effectstest.CardTarget(ogre)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PendingOfKind(g, game.PendingChoiceColor)
	if c == nil || !c.ColorMulti {
		t.Fatalf("prompt = %+v, want a multi colour pick", c)
	}
	if err := g.ResolveColorChoice(c.ID, me.ID, ""); err == nil {
		t.Fatal("an empty set of colours was accepted")
	}
	effectstest.AnswerColor(t, g, me.ID, "UW")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB9Colors(t, g, ogre); got != "WU" {
		t.Fatalf("colors = %q, want white and blue only", got)
	}
	layered := effectstest.LayeredCard(t, g, ogre)
	fromBlue, _ := game.ParseProtectionQuality(game.ProtectionFromColor("U"))
	fromRed, _ := game.ParseProtectionQuality(game.ProtectionFromColor("R"))
	if !fromBlue.Matches(game.SourceCharacteristics(&layered)) {
		t.Error("the white-and-blue creature is not blue for protection from blue")
	}
	if fromRed.Matches(game.SourceCharacteristics(&layered)) {
		t.Error("the creature is still red for protection from red")
	}
	effectstest.RunCleanup(t, g)
	if got := effectstest.TmtB9Colors(t, g, ogre); got != "R" {
		t.Errorf("colors after cleanup = %q, want red again", got)
	}
}

func TestScuttlemuttTargetsOnlyCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mutt := effectstest.B27Push(g, me.ID, "Scuttlemutt", "Artifact Creature — Scarecrow", scuttlemuttOracle, "{3}", 2, 2)
	rock := effectstest.B27Push(g, me.ID, "Rock", "Artifact", "", "{1}", 0, 0)
	if err := effectstest.InvB1Activate(g, me.ID, mutt, 0, effectstest.CardTarget(rock)...); err == nil {
		t.Fatal("Scuttlemutt targeted a noncreature artifact")
	}
}

// A target gone by resolution: no prompt, nothing changes (CR 608.2b).
func TestScuttlemuttDoesNothingWhenTheTargetLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mutt := effectstest.B27Push(g, me.ID, "Scuttlemutt", "Artifact Creature — Scarecrow", scuttlemuttOracle, "{3}", 2, 2)
	ogre := effectstest.B27Push(g, me.ID, "Ogre", "Creature — Ogre", "", "{2}{R}", 3, 3, "R")
	if err := effectstest.InvB1Activate(g, me.ID, mutt, 0, effectstest.CardTarget(ogre)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.LeaveBattlefield(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.PendingOfKind(g, game.PendingChoiceColor); c != nil {
		t.Errorf("a colour prompt opened for a target that left: %+v", c)
	}
}
