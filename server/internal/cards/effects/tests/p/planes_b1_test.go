package testsp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Coal Hill School --------------------------------------------------------

func TestCoalHillSchoolDrawsForAHistoricSpellCaster(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneCoalHillSchool)
	g, plane := effectstest.NewPlaneGame(t, effectstest.PlaneCoalHillSchool)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Test Bauble", "Artifact", "", nil)
	if got := effectstest.PlaneTriggerControllers(g, plane.InstanceID); len(got) != 1 || got[0] != me.ID {
		t.Fatalf("the historic cast triggered %v, want one item controlled by the planar controller", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a card was put in it, cast and one drawn, want %d", got, before+1)
	}
	effectstest.CastCatalogSpell(t, g, "Test Ritual", "Sorcery", "", nil)
	if n := effectstest.ItemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Errorf("a nonhistoric spell triggered the plane (%d items)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestCoalHillSchoolChaosReturnsAHistoricCard(t *testing.T) {
	g, plane := effectstest.NewPlaneGame(t, effectstest.PlaneCoalHillSchool)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Relic", "Legendary Artifact")
	bear := effectstest.PushGraveyardCardWithTypeLineID(me, "Bear", "Creature — Bear")
	rollBlank(t, g, me.ID)
	if n := effectstest.ItemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, effectstest.CardRef(bear)); err == nil {
		t.Fatal("a nonhistoric card was accepted as the target")
	}
	effectstest.AnswerPickTarget(t, g, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the historic card did not return to hand")
	}

	// Planeswalked away from, it is face down and does nothing.
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	rollChaos(t, g, me.ID)
	if n := effectstest.ItemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Error("a face-down plane's chaos ability triggered")
	}
}

// toNextUpkeep passes the turn and walks the next player's turn to its
// upkeep, where that player is the planar controller (CR 901.6). It
// returns the new active player.
func toNextUpkeep(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	effectstest.EndTurn(t, g)
	if g.Turn.Step != game.StepUpkeep {
		effectstest.AdvanceTo(t, g, game.StepUpkeep)
	}
	return g.Seats[g.Turn.ActiveSeat]
}

// advanceAnsweringPlayerTargets walks to `step`, answering every open
// player-target prompt with `target` and resolving what it put on the
// stack, and reports how many it answered.
func advanceAnsweringPlayerTargets(t *testing.T, g *game.Game, step game.Step, target uuid.UUID) int {
	t.Helper()
	asked := 0
	for i := 0; i < 60 && g.Turn.Step != step; i++ {
		if ch := effectstest.LatestChoiceOfKind(g, game.PendingChoicePickTarget); ch != nil {
			if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: target}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			asked++
			effectstest.PassPriorityAroundTable(t, g)
			continue
		}
		effectstest.FinishBlockDeclarations(t, g)
		if _, err := g.AdvanceStep(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("AdvanceStep toward %s: %v", step, err)
		}
	}
	if g.Turn.Step != step {
		t.Fatalf("never reached %s", step)
	}
	return asked
}

// --- Antarctic Research Base -------------------------------------------------

func TestAntarcticResearchBaseInvestigatesOnPlaneswalkAndUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneAntarcticBase)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneAntarcticBase)
	effectstest.AdvanceToMain(t, g)
	planeswalkToPlane(t, g, effectstest.PlaneAntarcticBase.Oracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Clue"); n != 1 {
		t.Fatalf("%d Clues after planeswalking to the plane, want 1", n)
	}
	next := toNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	clues := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" && c.Controller == next.ID {
			clues++
		}
	}
	if clues != 1 {
		t.Errorf("the new planar controller has %d Clues after their upkeep, want 1", clues)
	}
}

func TestAntarcticResearchBaseChaosCountsArtifactsAndMakesAPlant(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneAntarcticBase)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Bauble", "Artifact")
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.TmtB1Card(t, g, bear)
	if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 4 || tough != 4 {
		t.Errorf("the creature is %d/%d, want 4/4 (two artifacts)", p, tough)
	}
	if !c.HasSubtypeWord("Plant") || !c.HasSubtypeWord("Bear") {
		t.Errorf("the creature's subtypes lack Plant or Bear: %q", c.Effective().Subtypes)
	}
	effectstest.EndTurn(t, g)
	if !effectstest.TmtB1Card(t, g, bear).HasSubtypeWord("Plant") {
		t.Error("the Plant type ended with the turn; it has no duration")
	}
}

// --- Dalek Intensive Care ----------------------------------------------------

func TestDalekIntensiveCareExilesANonDalekForAHastyDalek(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneDalekCare)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneDalekCare)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	planeswalkToPlane(t, g, effectstest.PlaneDalekCare.Oracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	var dalek game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dalek" {
			dalek = c
		}
	}
	if dalek.InstanceID == uuid.Nil {
		t.Fatal("no Dalek was created")
	}
	if !effectstest.HasEffectiveKeyword(t, g, dalek.InstanceID, "haste") || !effectstest.HasEffectiveKeyword(t, g, dalek.InstanceID, "menace") {
		t.Error("the Dalek lacks haste or menace")
	}

	// The next upkeep: only the Dalek, so nothing is exiled and no Dalek made.
	_ = toNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Dalek"); n != 1 {
		t.Errorf("%d Daleks after an upkeep with no non-Dalek creature, want 1", n)
	}
}

func TestDalekIntensiveCareChaosBitesWithADalek(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneDalekCare)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	dalek := effectstest.B12Creature(g, me.ID, "Dalek", "Artifact Creature — Dalek", 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Human", 1, 3)
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, dalek)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the Dalek's 3 damage did not kill the 1/3")
	}
	if !g.Battlefield.Contains(dalek) {
		t.Error("the Dalek took damage back; the bite is one-sided")
	}
}

// --- Lake Silencio -----------------------------------------------------------

func TestLakeSilencioGivesEverySpellSplitSecond(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneLakeSilencio)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneLakeSilencio)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Test Ritual", "Sorcery", "", nil)
	if !g.SplitSecondActive {
		t.Fatal("a spell on the stack under Lake Silencio does not have split second")
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if g.SplitSecondActive {
		t.Error("split second outlasted planeswalking away from Lake Silencio")
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestLakeSilencioChaosDealsSixAndExilesTheDead(t *testing.T) {
	g, plane := effectstest.NewPlaneGame(t, effectstest.PlaneLakeSilencio)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Giant", 5, 5)
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, effectstest.CardRef(mine)); err == nil {
		t.Fatal("a creature the planar controller controls was accepted")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Error("the 5/5 dealt 6 damage was not exiled instead of dying")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventDealDamage && ev.Target == theirs && ev.Source != plane.InstanceID {
			t.Errorf("the damage source is %v, want Lake Silencio", ev.Source)
		}
	}
}

// --- Bowie Base One ----------------------------------------------------------

func TestBowieBaseOneGoadsACreatureOfThePlayerToTheLeft(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneBowieBaseOne)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneBowieBaseOne)
	effectstest.AdvanceToMain(t, g)
	left := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	across := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, left.ID, "Left Bear", "Creature — Bear", 2, 2)
	far := effectstest.B12Creature(g, across.ID, "Far Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, effectstest.CardRef(far)); err == nil {
		t.Fatal("a creature of a player not to the left was accepted")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.TmtB1Card(t, g, theirs); len(c.Goads) == 0 {
		t.Error("the creature was not goaded")
	}
}

func TestBowieBaseOneChaosGrantsIslandwalk(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneBowieBaseOne)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "islandwalk") {
		t.Fatal("no islandwalk")
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "islandwalk") {
		t.Error("islandwalk outlasted the turn")
	}
}

// --- City of the Daleks ------------------------------------------------------

func TestCityOfTheDaleksDrainsForArtifactsWhenYouAttack(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneCityOfTheDaleks)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneCityOfTheDaleks)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Bauble", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Ring", "Artifact")
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	asked := advanceAnsweringPlayerTargets(t, g, game.StepDeclareBlockers, opp.ID)
	if asked != 1 {
		t.Fatalf("the trigger asked for a target %d times, want once per attack", asked)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent lost %d life to the trigger, want 3 (once per attack, three artifacts)", got)
	}
}

func TestCityOfTheDaleksChaosMakesADalekForEachOpponent(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneCityOfTheDaleks)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	var daleks []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dalek" && c.Controller == me.ID {
			daleks = append(daleks, c.InstanceID)
		}
	}
	if len(daleks) != 3 {
		t.Fatalf("%d Daleks, want one for each of three opponents", len(daleks))
	}
	for _, id := range daleks {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") || !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
			t.Error("a Dalek lacks haste or menace")
		}
	}
	// Each Dalek attacks the opponent it was made for, if able.
	bound := map[uuid.UUID]uuid.UUID{}
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAttackPlayerIfAble && len(e.Affected) == 1 {
				bound[e.Affected[0].ID] = m.Player
			}
		}
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range daleks {
		opp, ok := bound[id]
		if !ok || opp == me.ID {
			t.Fatalf("Dalek %v is bound to %v", id, opp)
		}
		seen[opp] = true
	}
	if len(seen) != 3 {
		t.Fatalf("the Daleks are bound to %d opponents, want 3", len(seen))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range daleks {
		if err := g.DeclareAttacker(id, bound[id]); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceAnsweringPlayerTargets(t, g, game.StepEnd, effectstest.OpponentOf(g).ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range daleks {
		if g.Battlefield.Contains(id) {
			t.Error("a Dalek was not sacrificed at the end step")
		}
	}
}

// --- Kerblam! Warehouse ------------------------------------------------------

func TestKerblamWarehouseMakesATreasureOncePerCombatDamage(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneKerblamWarehouse)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneKerblamWarehouse)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

func TestKerblamWarehouseChaosGrantsTheFlipToNoncreatureArtifacts(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneKerblamWarehouse)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("an artifact creature was granted the ability")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, rock, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("the noncreature artifact cannot activate the granted ability: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact was not sacrificed to pay")
	}
	effectstest.PassPriorityAroundTable(t, g)
	flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if flip == nil {
		t.Fatal("no coin call")
	}
	if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	won := false
	for _, ev := range effectstest.RandomEvents(g, game.EventFlipCoin) {
		won = ev.Won
	}
	effectstest.PassPriorityAroundTable(t, g)
	want := life
	if won {
		want -= 3
	}
	if opp.Life != want {
		t.Errorf("won=%v: opponent at %d, want %d", won, opp.Life, want)
	}
}

// --- Mondassian Colony Ship --------------------------------------------------

func TestMondassianColonyShipPumpsAttackersForTheirKin(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneMondassianShip)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneMondassianShip)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	elf := effectstest.PushDiesCreatureForTest(g, me.ID, "Elf A", "", "Creature — Elf Warrior", 1, 1)
	effectstest.PushDiesCreatureForTest(g, me.ID, "Elf B", "", "Creature — Elf", 1, 1)
	effectstest.PushDiesCreatureForTest(g, me.ID, "Warrior", "", "Creature — Human Warrior", 1, 1)
	effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 1, 1)
	effectstest.PushDiesCreatureForTest(g, opp.ID, "Their Elf", "", "Creature — Elf", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(elf, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, elf); p != 3 || tough != 3 {
		t.Errorf("the attacking Elf Warrior is %d/%d, want 3/3 (one Elf, one Warrior of its controller's)", p, tough)
	}
}

func TestMondassianColonyShipChaosMakesACyberman(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneMondassianShip)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 7, 7)
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, giant)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.TmtB1Card(t, g, giant)
	if !c.FaceDownIsPermanent() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 ||
		!c.HasSubtypeWord("Cyberman") || !c.IsArtifact() {
		t.Errorf("the creature is not a face-down 2/2 Cyberman artifact creature: %q %d/%d", c.Effective().Subtypes, c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Besieged Viking Village -------------------------------------------------

func TestBesiegedVikingVillageGivesEveryCreatureBoast(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneVikingVillage)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneVikingVillage)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.BumpLayerVersionForTest()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("boast was activated by a creature that has not attacked")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the attacking creature cannot boast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.MkmMPPT(t, g, bear); p != 3 {
		t.Errorf("the creature has power %d after boasting, want 3", p)
	}
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("boast was activated twice in one turn")
	}
}

func TestBesiegedVikingVillageChaosPutsAnIndestructibleCounter(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneVikingVillage)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, effectstest.CardRef(idle)); err == nil {
		t.Fatal("a creature that did not attack was accepted")
	}
	effectstest.AnswerPickTarget(t, g, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Card(t, g, attacker).Counters[game.CounterIndestructible] != 1 {
		t.Error("no indestructible counter")
	}
}

// --- Gardens of Tranquil Repose ----------------------------------------------

func TestGardensOfTranquilReposeExilesTheDeadAndCountsThem(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneGardensOfRepose)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneGardensOfRepose)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(theirs)
		_ = g.DestroyPermanentForEffect(mine)
	})
	for i := 0; i < 4; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		for _, p := range []*game.Player{me, opp} {
			if effectstest.ScryChoiceFor(g, p.ID) != nil {
				effectstest.AnswerScryKeepAll(t, g, p.ID)
			}
		}
	}
	if !g.Exile.Contains(theirs) || !g.Exile.Contains(mine) {
		t.Fatal("the dead creatures were not exiled")
	}
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Dalek"); n != 3 {
		t.Errorf("%d Daleks, want 3 (one plus two cards exiled with the plane)", n)
	}
	// Planeswalked away from and back to, it is a new object.
	planeswalkToPlane(t, g, effectstest.PlaneGardensOfRepose.Oracle)
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Dalek"); n != 4 {
		t.Errorf("%d Daleks, want 4: the returned plane exiled nothing", n)
	}
}

// --- Hotel of Fears ----------------------------------------------------------

func TestHotelOfFearsExilesTheTopAndCostsItsManaValue(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneHotelOfFears)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneHotelOfFears)
	next := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := game.Card{InstanceID: uuid.New(), Name: "Top Spell", TypeLine: "Sorcery", ManaCost: "{2}{B}", Owner: next.ID, Controller: next.ID}
	next.Library.PushTop(top)
	life := next.Life
	if got := toNextUpkeep(t, g); got.ID != next.ID {
		t.Fatalf("the next turn is %s's", got.Name)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top.InstanceID) {
		t.Fatal("the top card was not exiled")
	}
	if got := life - next.Life; got != 3 {
		t.Errorf("lost %d life, want 3 (the mana value)", got)
	}
}

func TestHotelOfFearsChaosPraisesHim(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneHotelOfFears)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Knight", TypeLine: "Creature — Knight",
		ManaCost: "{W}{W}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, knight)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "W")
	effectstest.AnswerChooseCards(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.MkmMPPT(t, g, knight); p != 4 {
		t.Errorf("the Knight has power %d, want 4 (devotion to white 2)", p)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("another creature was not sacrificed")
	}
	if !g.Battlefield.Contains(knight) {
		t.Error("the target was sacrificed; it must be another creature")
	}
}

// --- Bad Wolf Bay ------------------------------------------------------------

func TestBadWolfBayExilesAtCombatAndReturnsAtTheEndStep(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneBadWolfBay)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneBadWolfBay)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.OpponentOf(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("the creature was not exiled")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	// A card returned from exile is a new object with a new instance.
	returned := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && c.Controller == opp.ID {
			returned = true
		}
	}
	if !returned {
		t.Error("the creature did not return under its owner's control at the end step")
	}
}

func TestBadWolfBayChaosStopsReturnsFromExileThenPlaneswalks(t *testing.T) {
	g, plane := effectstest.NewPlaneGame(t, effectstest.PlaneBadWolfBay)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.OpponentOf(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.FaceUpPlane(g, effectstest.PlaneBadWolfBay.Oracle); ok {
		t.Fatal("the chaos ability did not planeswalk away from Bad Wolf Bay")
	}
	if !g.PlanarDeck.Contains(plane.InstanceID) {
		t.Error("Bad Wolf Bay is not on the bottom of the planar deck")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Error("the creature returned from exile on a turn cards can't enter from exile")
	}
	// A token is not a card, and the rule ends with the turn.
	effectstest.EndTurn(t, g)
	g.WithWriteLock(func() {
		for i := range g.ScopedEffects {
			for _, m := range g.ScopedEffects[i].Mods {
				if m.Kind == game.ModCardsCantEnterFromExile {
					t.Error("the rule outlasted the turn")
				}
			}
		}
	})
}

// --- Aplan Mortarium ---------------------------------------------------------

func TestAplanMortariumExposureBuildsAcrossUpkeeps(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneAplanMortarium)
	g, plane := effectstest.NewPlaneGame(t, effectstest.PlaneAplanMortarium)
	first := toNextUpkeep(t, g)
	life := first.Life
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - first.Life; got != 1 {
		t.Fatalf("the first upkeep cost %d life, want 1", got)
	}
	second := toNextUpkeep(t, g)
	life = second.Life
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - second.Life; got != 2 {
		t.Errorf("the second upkeep cost %d life, want 2 (two exposure counters)", got)
	}
	up, _ := effectstest.FaceUpPlane(g, effectstest.PlaneAplanMortarium.Oracle)
	if up.Counters[effects.AplanMortariumExposure] != 2 {
		t.Errorf("the plane has %d exposure counters, want 2", up.Counters[effects.AplanMortariumExposure])
	}
	// Planeswalked away from, it is a new object with no counters.
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(second.ID) })
	for _, c := range g.PlanarDeckCards() {
		if c.InstanceID == plane.InstanceID && len(c.Counters) != 0 {
			t.Errorf("the face-down plane kept its counters: %v", c.Counters)
		}
	}
}

func TestAplanMortariumChaosAngelsStopBeingCreatures(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneAplanMortarium)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	var angels []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Alien Angel" && c.Controller == me.ID {
			angels = append(angels, c.InstanceID)
		}
	}
	if len(angels) != 2 {
		t.Fatalf("%d Alien Angels, want 2", len(angels))
	}
	a := angels[0]
	if !effectstest.HasEffectiveKeyword(t, g, a, "first strike") || !effectstest.HasEffectiveKeyword(t, g, a, "vigilance") {
		t.Error("the token lacks first strike or vigilance")
	}
	// The controller's own creature spell does nothing; an opponent's
	// makes each Angel a noncreature artifact until end of turn.
	effectstest.WeepingAngelOpponentCastsCreature(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, a), "Creature") {
		t.Fatal("the controller's own creature spell turned the token off")
	}
	effectstest.WeepingAngelOpponentCastsCreature(t, g, effectstest.OpponentOf(g))
	for i := 0; i < 4; i++ {
		p := effectstest.TriggerOrderPromptFor(g, me.ID)
		if p == nil {
			break
		}
		if err := g.ResolveTriggerOrder(p.ID, me.ID, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if types := effectstest.EffectiveTypes(t, g, a); effectstest.ContainsString(types, "Creature") || !effectstest.ContainsString(types, "Artifact") {
		t.Errorf("after an opponent's creature spell the token's types are %v; want an artifact that isn't a creature", types)
	}
	if effectstest.TmtB1Card(t, g, a).HasSubtypeWord("Angel") {
		t.Error("the noncreature token is still an Angel")
	}
	effectstest.EndTurn(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, a), "Creature") {
		t.Error("the token is still not a creature next turn")
	}
}

// --- Amy's Home --------------------------------------------------------------

func TestAmysHomeSuspendsACardFromHandForItsManaValue(t *testing.T) {
	requirePlaneRegisteredFull(t, effectstest.PlaneAmysHome)
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneAmysHome)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := game.Card{InstanceID: uuid.New(), Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{2}{U}", Owner: me.ID, Controller: me.ID}
	land := game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(spell)
	me.Hand.PushTop(land)
	planeswalkToPlane(t, g, effectstest.PlaneAmysHome.Oracle)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no choice of a card from hand")
	}
	if effectstest.ContainsID(ch.ChooseCards, land.InstanceID) {
		t.Error("a land was offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, spell.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(spell.InstanceID)
	if !ok || !g.Exile.Contains(spell.InstanceID) {
		t.Fatal("the card was not exiled")
	}
	if c.Counters[game.CounterTime] != 3 || !game.CardHasSuspend(c) {
		t.Errorf("the exiled card has %d time counters (suspend %v), want 3 and suspend", c.Counters[game.CounterTime], game.CardHasSuspend(c))
	}
	// "You may": choosing nothing exiles nothing.
	next := toNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if ch := effectstest.ChooseCardsChoiceFor(g, next.ID); ch != nil {
		hand := next.Hand.Size()
		effectstest.AnswerChooseCards(t, g, next.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if next.Hand.Size() != hand {
			t.Error("declining exiled a card anyway")
		}
	}
}

func TestAmysHomeChaosTimeTravels(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneAmysHome)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	suspended := game.Card{InstanceID: uuid.New(), Name: "Suspended Spell", TypeLine: "Sorcery", ManaCost: "{1}",
		Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() {
		g.Exile.PushTop(suspended)
		_ = g.GainSuspendForEffect(me.ID, suspended.InstanceID, 2)
	})
	clock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterTime: 1}})
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // the Clock: add one
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // the suspended card: remove one
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TmtB1Card(t, g, clock).Counters[game.CounterTime]; n != 2 {
		t.Errorf("the Clock has %d time counters, want 2", n)
	}
	c, _ := g.LookupCardForEffect(suspended.InstanceID)
	if n := c.Counters[game.CounterTime]; n != 1 {
		t.Errorf("the suspended card has %d time counters, want 1", n)
	}
}

// --- Restore points ----------------------------------------------------------

// Bad Wolf Bay's stored rule and Amy's Home's time-travel question are data:
// a table holding either survives a snapshot round trip and behaves the same.
func TestPlaneBatch1PausesAndRulesAreRestorePoints(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneBadWolfBay)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	found := false
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			found = found || m.Kind == game.ModCardsCantEnterFromExile
		}
	}
	if !found {
		t.Error("the restored table lost \"cards can't enter from exile this turn\"")
	}

	g, _ = effectstest.NewPlaneGame(t, effectstest.PlaneAmysHome)
	effectstest.AdvanceToMain(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	clock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterTime: 1}})
	rollChaos(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TmtB1Card(t, g, clock).Counters[game.CounterTime]; n != 0 {
		t.Errorf("after the restored time-travel question the Clock has %d time counters, want 0", n)
	}
}
