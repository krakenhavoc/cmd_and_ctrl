package testso

import (
	"testing"

	"github.com/google/uuid"

	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaslem's combat-damage trigger reveals that many cards; the pick is
// a creature and/or a land, and the rest goes to the bottom.
func TestOjerKaslemRevealsAndPutsACreatureAndALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	god := effectstest.ImportToBattlefield(t, g, effectstest.OjerKaslemRow(), me)
	// Pushed to the top one after another, so the LAST pushed is the top.
	var ids []uuid.UUID
	for _, tl := range []string{"Basic Land — Island", "Basic Land — Forest", "Creature — Ox", "Creature — Elf", "Sorcery"} {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: tl, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	island, forest, ox, elf, sorcery := ids[0], ids[1], ids[2], ids[3], ids[4]

	effectstest.DealCombatDamageToPlayer(g, god, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick prompt after the reveal")
	}
	// Four revealed: sorcery, elf, ox, forest. The island stays put.
	for _, id := range []uuid.UUID{elf, ox, forest} {
		if !effectstest.ContainsID(c.ChooseCards, id) {
			t.Errorf("a revealed creature or land was not offered")
		}
	}
	if effectstest.ContainsID(c.ChooseCards, sorcery) || effectstest.ContainsID(c.ChooseCards, island) {
		t.Error("a sorcery or an unrevealed card was offered")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{elf, ox}); err == nil {
		t.Error("two creatures were accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, elf, forest)
	if !g.Battlefield.Contains(elf) || !g.Battlefield.Contains(forest) {
		t.Error("the picked creature and land did not enter")
	}
	if g.Battlefield.Contains(ox) || g.Battlefield.Contains(sorcery) {
		t.Error("an unpicked card entered")
	}
}

// Kaslem returns as Temple of Cultivation, which transforms only with
// ten or more permanents.
func TestOjerKaslemTempleTransformsOnlyWithTenPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := effectstest.ImportToBattlefield(t, g, effectstest.OjerKaslemRow(), me)
	effectstest.OjerKill(t, g, god)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	temple := effectstest.OjerOnBattlefield(g, cardso.OjerKaslemOracleID)
	if temple == nil || temple.Name != "Temple of Cultivation" || !temple.Tapped {
		t.Fatalf("returned as %+v", temple)
	}
	id := temple.InstanceID
	g.WithWriteLock(func() { effectstest.OjerOnBattlefield(g, cardso.OjerKaslemOracleID).Tapped = false })
	effectstest.FillPoolColored(me, "G", 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("transformed without ten permanents")
	}
	for i := 0; i < 9; i++ {
		effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	}
	effectstest.S58p6Activate(t, g, me.ID, id, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.OjerOnBattlefield(g, cardso.OjerKaslemOracleID)
	if back == nil || back.ActiveFace != 0 {
		t.Fatalf("after the transform: %+v, want the God again", back)
	}
}
