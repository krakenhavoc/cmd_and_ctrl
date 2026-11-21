package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fra_prepare_b_surveil_test.go — the three fra-prepare-b cards that
// lean on Peer Review / Soul Tether and surveil.

const (
	fraPrudentFateseer  = "63f67baa-3f37-41b8-9ee6-3c37f828917d"
	fraSemesterForeseer = "21b8d59f-a90d-44ff-a2e3-9f439ac3e14a"
	fraWoodworkProdigy  = "f3ed157f-a09b-405f-aec4-0e7a81d80fc8"
)

func fraBKeepSurveil(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	ch := effectstest.SurveilChoiceFor(g, p.ID)
	if ch == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(ch.ID, p.ID, ch.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

func TestWoodworkProdigyPreparesAtUpkeepAndSoulTetherMakesAHeartwood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := fraPrepB(me.ID, fraWoodworkProdigy, nil, "Woodwork Prodigy", "Creature — Cat Druid", 3, 3, "Soul Tether", "Sorcery")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("entered prepared; only the upkeep trigger prepares it")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("not prepared after its controller's upkeep")
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FraBCastCopy(t, g, me, "Soul Tether")
	if effectstest.FindBattlefieldByName(g, "Heartwood") == uuid.Nil {
		t.Fatal("no Heartwood token")
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

func TestSemesterForeseerSurveilsOnEntryAndPeerReviewMakesACadet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.SeedLibrary(me, "Fodder", "Keeper", "Deep")
	c := fraPrepB(me.ID, fraSemesterForeseer, nil, "Semester Foreseer", "Creature — Human Wizard", 3, 4, "Peer Review", "Sorcery")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("did not enter prepared")
	}
	fraBKeepSurveil(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FraBCastCopy(t, g, me, "Peer Review")
	cadet := effectstest.FindBattlefieldByName(g, "Cadet")
	if cadet == uuid.Nil {
		t.Fatal("no Cadet token")
	}
	if effectstest.EffectivePower(t, g, cadet) != 2 || effectstest.EffectiveToughness(t, g, cadet) != 2 {
		t.Errorf("Cadet is %d/%d, want 2/2", effectstest.EffectivePower(t, g, cadet), effectstest.EffectiveToughness(t, g, cadet))
	}
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("Peer Review did not surveil after making the token")
	}
}

func TestPrudentFateseerPumpsOnlyOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	c := fraPrepB(me.ID, fraPrudentFateseer, nil, "Prudent Fateseer", "Creature — Dwarf Wizard", 1, 4, "Peer Review", "Sorcery")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("did not enter prepared")
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Theirs", 2, 2)

	// First surveil of the turn: Peer Review's.
	effectstest.FraBCastCopy(t, g, me, "Peer Review")
	fraBKeepSurveil(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Fatalf("bear power after the first surveil = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature got the pump: power %d", got)
	}

	// Second surveil in the same turn: Semester Foreseer's entry.
	sem := fraPrepB(me.ID, fraSemesterForeseer, nil, "Semester Foreseer", "Creature — Human Wizard", 3, 4, "Peer Review", "Sorcery")
	effectstest.CastFromHandAndResolve(t, g, me, sem)
	fraBKeepSurveil(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("bear power after the second surveil = %d, want 3 (once each turn)", got)
	}
}
