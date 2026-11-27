package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jud_old_b624_test.go — S82 OLD batch 624 (#5611): Judgment, activated
// abilities. Each test plays the real catalog card through the engine.

const (
	oldB624Wayfarer    = "1729317d-ccfb-4f27-84c6-20b36b2826c9"
	oldB624Wurm        = "31370fdf-c666-40d2-a7ff-3ba1dfe350cd"
	oldB624Inkshrouder = "a3add0d7-2610-4b4e-9d43-3f04a5cd8259"
	oldB624Brushhopper = "bdecf59c-f8bd-4003-9058-685b70c08837"
	oldB624Druid       = "c2b4f43a-6609-43ca-ad0d-897b6f8cb22f"
	oldB624MirrorWall  = "c4ab446a-97e4-46a5-89a8-180a0bf3447c"
)

func TestOldB624EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB624Wayfarer: "Krosan Wayfarer", oldB624Wurm: "Tunneler Wurm", oldB624Inkshrouder: "Cephalid Inkshrouder",
		oldB624Brushhopper: "Anurid Brushhopper", oldB624Druid: "Harvester Druid", oldB624MirrorWall: "Mirror Wall",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB624Hand(p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, effectstest.HandCardForTest(p, "Spare", "Sorcery", ""))
	}
	return ids
}

func TestOldB624KrosanWayfarerSacrificesForALandFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	w := effectstest.BlbB1Push(g, me.ID, "Krosan Wayfarer", "Creature — Human Druid", oldB624Wayfarer, 1, 1)
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	spell := effectstest.HandCardForTest(me, "Bolt", "Instant", "")
	effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(w) {
		t.Fatal("the Wayfarer was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		for _, c := range pick.ChooseCards {
			if c == spell {
				t.Error("a nonland card was offered")
			}
		}
		effectstest.AnswerChooseCards(t, g, me.ID, land)
	} else {
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.AnswerChooseCards(t, g, me.ID, land)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not enter the battlefield")
	}
	if !me.Hand.Contains(spell) {
		t.Error("the spell left the hand")
	}
}

func TestOldB624TunnelerWurmRegeneratesForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	wurm := effectstest.BlbB1Push(g, me.ID, "Tunneler Wurm", "Creature — Wurm", oldB624Wurm, 6, 6)
	hand := oldB624Hand(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, wurm, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without naming a card to discard")
	}
	effectstest.B16Activate(t, g, me.ID, wurm, 0, game.ActivateAbilityParams{DiscardIDs: hand[:1]})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(hand[0]) || !me.Graveyard.Contains(hand[0]) {
		t.Error("the discard was not paid")
	}
	effectstest.P1Destroy(t, g, wurm)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(wurm) {
		t.Fatal("the regeneration shield did not save the Wurm")
	}
	if tapped := effectstest.B16Tapped(t, g, wurm); !tapped {
		t.Error("a regenerated creature should be tapped")
	}
}

func TestOldB624CephalidInkshrouderGainsShroudAndCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	ink := effectstest.B12Push(g, me.ID, "Cephalid Inkshrouder", "Creature — Octopus", oldB624Inkshrouder, 2, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	hand := oldB624Hand(me, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, ink, "shroud") {
		t.Fatal("shroud before the ability")
	}
	effectstest.B16Activate(t, g, me.ID, ink, 0, game.ActivateAbilityParams{DiscardIDs: hand})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, ink, "shroud") {
		t.Error("no shroud after the ability")
	}
	effectstest.BrAttack(t, g, ink)
	if err := g.DeclareBlocker(blocker, ink); err == nil {
		t.Error("the Inkshrouder was blocked")
	}
}

func TestOldB624AnuridBrushhopperBlinksAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	hop := effectstest.B12Push(g, me.ID, "Anurid Brushhopper", "Creature — Frog Beast", oldB624Brushhopper, 3, 4)
	hand := oldB624Hand(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, hop, 0, game.ActivateAbilityParams{DiscardIDs: hand[:1]}); err == nil {
		t.Fatal("one discard paid for a two-card cost")
	}
	effectstest.B16Activate(t, g, me.ID, hop, 0, game.ActivateAbilityParams{DiscardIDs: hand[:2]})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hop) {
		t.Fatal("the Brushhopper stayed on the battlefield")
	}
	if me.Hand.Size() < 1 || !me.Graveyard.Contains(hand[0]) || !me.Graveyard.Contains(hand[1]) {
		t.Error("the two discards were not paid")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oldB624Brushhopper && c.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("the Brushhopper did not return at the beginning of the next end step")
	}
}

func TestOldB624HarvesterDruidMakesAColorYourLandsCouldProduce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	druid := effectstest.B12Push(g, me.ID, "Harvester Druid", "Creature — Human Druid", oldB624Druid, 1, 1)
	spec, _ := effects.Lookup(oldB624Druid)
	if len(spec.ManaAbilities) != 1 || !spec.ManaAbilities[0].DerivesFromOtherSources || !spec.ManaAbilities[0].DerivedColorsOnly {
		t.Fatalf("mana abilities: %+v", spec.ManaAbilities)
	}
	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the Druid: %v", err)
	}
	if got := effectstest.PoolColours(me); got["G"] != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("pool %v, want one green", got)
	}
}

func TestOldB624HarvesterDruidMakesNothingWithoutLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	// An opponent's land does not count.
	effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	druid := effectstest.B12Push(g, me.ID, "Harvester Druid", "Creature — Human Druid", oldB624Druid, 1, 1)
	_ = g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{})
	if me.ManaPool.Total() != 0 {
		t.Errorf("the pool holds %d mana with no lands of mine, want 0", me.ManaPool.Total())
	}
}

func TestOldB624MirrorWallCanAttackOnlyAfterThePayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	wall := effectstest.B12Push(g, me.ID, "Mirror Wall", "Creature — Wall", oldB624MirrorWall, 3, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, wall, "defender") {
		t.Fatal("no defender")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Fatal("a defender attacked")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2, _ := effectstest.BlbB1Seats(g2)
	wall2 := effectstest.B12Push(g2, me2.ID, "Mirror Wall", "Creature — Wall", oldB624MirrorWall, 3, 4)
	effectstest.B06AddMana(me2, "W")
	effectstest.B16Activate(t, g2, me2.ID, wall2, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(wall2, opp2.ID); err != nil {
		t.Errorf("the paid-for Wall could not attack: %v", err)
	}
}
