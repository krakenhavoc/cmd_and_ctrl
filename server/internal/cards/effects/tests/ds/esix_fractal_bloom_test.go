package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const esixOracle = "9d22960b-babc-4cf3-b228-d32e13bc6014"

func TestEsixFractalBloomIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(esixOracle)
	if !ok || spec.Name != "Esix, Fractal Bloom" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// esixSetup parks Esix under me, a tapped Golem of mine with two +1/+1
// counters, and an opponent's Ogre.
func esixSetup(t *testing.T, g *game.Game, me, opp uuid.UUID) (esix, golem, ogre uuid.UUID) {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	esix = effectstest.PushCatalogPermanent(g, me, "Esix, Fractal Bloom", "Legendary Creature — Fractal", esixOracle, false)
	golem = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Host Golem", TypeLine: "Artifact Creature — Golem",
		Power: 3, Toughness: 3, Owner: me, Controller: me,
	})
	ogre = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Ogre", TypeLine: "Creature — Ogre",
		Power: 4, Toughness: 2, Owner: opp, Controller: opp,
	})
	effectstest.AddCountersForTest(g, golem, game.CounterPlusOne, 2)
	effectstest.TapForTest(t, g, golem)
	return esix, golem, ogre
}

// esixPrompt is the open copy_target prompt, which must be the token
// kind and owed by `chooser`.
func esixPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	p := effectstest.PendingOfKind(g, game.PendingChoiceCopyTarget)
	if p == nil {
		t.Fatalf("no copy_target prompt: %+v", g.PendingChoices)
	}
	if p.Chooser != chooser || p.CopyUse != game.CopyUseTokens {
		t.Fatalf("prompt chooser=%v use=%q, want %v and %q", p.Chooser, p.CopyUse, chooser, game.CopyUseTokens)
	}
	return p
}

// The first creation of your turn may become that many copies of a
// creature you choose, an opponent's included and never Esix: plain
// copies, without the original's counters or tapped state. The second
// creation of the turn is made as printed.
func TestEsixFractalBloomTurnsTheFirstCreationIntoCopiesOfAChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	esix, golem, ogre := esixSetup(t, g, me.ID, opp.ID)

	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 2)
	p := esixPrompt(t, g, me.ID)
	has := func(id uuid.UUID) bool {
		for _, o := range p.CopyOptions {
			if o == id {
				return true
			}
		}
		return false
	}
	if has(esix) || !has(golem) || !has(ogre) {
		t.Fatalf("options = %v; want the Golem and the Ogre and not Esix", p.CopyOptions)
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 0 {
		t.Fatalf("Treasures = %d before the answer, want none", got)
	}
	if err := g.ResolveCopyTarget(p.ID, me.ID, golem); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 0 {
		t.Errorf("Treasures = %d, want none", got)
	}
	var copies []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Host Golem" && c.InstanceID != golem {
				copies = append(copies, c)
			}
		}
	})
	if len(copies) != 2 {
		t.Fatalf("%d copies of the Golem, want 2", len(copies))
	}
	for _, c := range copies {
		if !effects.IsToken(c) || c.Tapped || c.Counters[game.CounterPlusOne] != 0 || c.Controller != me.ID {
			t.Errorf("a copy is token=%v tapped=%v with %d counters under %v; want an untapped token of mine without counters",
				effects.IsToken(c), c.Tapped, c.Counters[game.CounterPlusOne], c.Controller)
		}
		if lc := effectstest.LayeredCard(t, g, c.InstanceID); lc.CurrentPower() != 3 || lc.CurrentToughness() != 3 {
			t.Errorf("a copy is %d/%d, want the printed 3/3", lc.CurrentPower(), lc.CurrentToughness())
		}
	}

	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	if p := effectstest.PendingOfKind(g, game.PendingChoiceCopyTarget); p != nil {
		t.Fatal("the second creation of the turn was offered the swap")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
}

// An opponent's creature can be the one copied, and the copies are yours.
func TestEsixFractalBloomCopiesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	_, _, ogre := esixSetup(t, g, me.ID, opp.ID)

	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	p := esixPrompt(t, g, me.ID)
	if err := g.ResolveCopyTarget(p.ID, me.ID, ogre); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	mine := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Their Ogre" && c.InstanceID != ogre && c.Controller == me.ID && effects.IsToken(c) {
				mine++
			}
		}
	})
	if mine != 1 || effectstest.OnBattlefieldNamed(g, "Treasure") != 0 {
		t.Fatalf("Ogre copies of mine = %d, Treasures = %d; want 1 and 0", mine, effectstest.OnBattlefieldNamed(g, "Treasure"))
	}
}

// Choosing no creature makes the tokens as printed and uses up the
// turn's first time; so does a pick that is not a candidate (Esix).
// Another player's creation is not yours.
func TestEsixFractalBloomDecliningAndOtherPlayersTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	esix, _, _ := esixSetup(t, g, me.ID, opp.ID)

	effectstest.MakeTokens(t, g, opp.ID, effects.TreasureToken(), 1)
	if p := effectstest.PendingOfKind(g, game.PendingChoiceCopyTarget); p != nil {
		t.Fatal("an opponent's creation was offered the swap")
	}
	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	p := esixPrompt(t, g, me.ID)
	if err := g.ResolveCopyTarget(p.ID, opp.ID, uuid.Nil); err == nil {
		t.Fatal("an opponent answered Esix's prompt")
	}
	if err := g.ResolveCopyTarget(p.ID, me.ID, esix); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	if got := effectstest.OnBattlefieldNamed(g, "Esix, Fractal Bloom"); got != 1 {
		t.Fatalf("Esixes = %d; Esix copied itself", got)
	}
	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	if p := effectstest.PendingOfKind(g, game.PendingChoiceCopyTarget); p != nil {
		t.Fatal("declining did not use up the turn's first time")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 3 {
		t.Errorf("Treasures = %d, want 3", got)
	}
}

// "During each of your turns": a creation of yours in another player's
// turn is never offered the swap.
func TestEsixFractalBloomOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active, other := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	esixSetup(t, g, other.ID, active.ID)

	effectstest.MakeTokens(t, g, other.ID, effects.TreasureToken(), 1)
	if p := effectstest.PendingOfKind(g, game.PendingChoiceCopyTarget); p != nil {
		t.Fatal("Esix offered the swap in another player's turn")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
}

// A table paused on Esix's question is a restore point: the restored
// table answers it and makes the copies.
func TestEsixFractalBloomOpenQuestionIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	_, golem, _ := esixSetup(t, g, me.ID, opp.ID)

	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 2)
	esixPrompt(t, g, me.ID)
	restored := effectstest.RestoreRoundTrip(t, g, true)
	p := esixPrompt(t, restored, me.ID)
	if err := restored.ResolveCopyTarget(p.ID, me.ID, golem); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	if golems, treasures := effectstest.OnBattlefieldNamed(restored, "Host Golem"), effectstest.OnBattlefieldNamed(restored, "Treasure"); golems != 3 || treasures != 0 {
		t.Fatalf("restored: Golems = %d, Treasures = %d; want 3 and 0", golems, treasures)
	}
}
