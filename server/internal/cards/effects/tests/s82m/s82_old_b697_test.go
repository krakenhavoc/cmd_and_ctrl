package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b697_test.go — Mercadian Masques, "Instants and sorceries" (S82 OLD
// batch 697, #5619): each card driven through a real cast.

const (
	oldB697ShovingMatch = "1ef07ae5-3b04-45c8-a080-852b84da4fef"
	oldB697Moment       = "43af5041-a4bd-4b61-b1a1-bdea916afc59"
	oldB697ClearLand    = "4604c729-8d2d-4b2b-a95c-5245563637e6"
	oldB697Tectonic     = "4a3c6d0e-6d77-4434-95f4-737879de026f"
	oldB697Extortion    = "8841fdb9-0565-4ef5-814c-293c14242741"
	oldB697Wail         = "af32ce11-c026-4135-b1e3-10c6aa15c72c"
	oldB697Touch        = "b5f9920d-7af5-4876-8823-6efbf2b4d9b9"
	oldB697Generation   = "d5b021e3-a0c1-471b-9088-a76816806354"
)

func oldB697PlayerTarget(p *game.Player) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: p.ID}}
}

func TestS82OldB697AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB697ShovingMatch, oldB697Moment, oldB697ClearLand, oldB697Tectonic,
		oldB697Extortion, oldB697Wail, oldB697Touch, oldB697Generation} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB697TurnSteps(t *testing.T, g *game.Game) []game.Step {
	t.Helper()
	seat := g.Turn.ActiveSeat
	var since uint64
	for _, ev := range g.Events {
		since = max(since, ev.Seq)
	}
	for i := 0; i < 64 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	var out []game.Step
	for _, ev := range g.Events {
		if ev.Seq > since && ev.Kind == game.EventStepBegan {
			out = append(out, ev.Step)
		}
	}
	return out
}

func oldB697HasCombat(steps []game.Step) bool {
	return slices.ContainsFunc(steps, func(s game.Step) bool { return game.PhaseOf(s) == game.PhaseCombat })
}

func TestS82OldB697MomentOfSilenceSkipsThisTurnsCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Moment of Silence", "Instant", oldB697Moment, oldB697PlayerTarget(me))
	effectstest.PassPriorityAroundTable(t, g)
	if n := g.NextCombatPhaseSkipsForLocked(me.ID); n != 1 {
		t.Fatalf("the target owes %d combat skips, want 1", n)
	}
	steps := oldB697TurnSteps(t, g)
	if oldB697HasCombat(steps) {
		t.Errorf("this turn still had combat: %v", steps)
	}
	if !slices.Contains(steps, game.StepPostcombatMain) {
		t.Errorf("the turn lost its second main phase: %v", steps)
	}
}

func TestS82OldB697MomentOfSilenceOnAnotherPlayerLastsOnlyThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Moment of Silence", "Instant", oldB697Moment, oldB697PlayerTarget(opp))
	effectstest.PassPriorityAroundTable(t, g)
	steps := oldB697TurnSteps(t, g)
	if !oldB697HasCombat(steps) {
		t.Errorf("the caster's own combat was skipped: %v", steps)
	}
	if n := g.NextCombatPhaseSkipsForLocked(opp.ID); n != 0 {
		t.Errorf("the skip outlived the turn: %d owed", n)
	}
	_ = me
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if steps := oldB697TurnSteps(t, g); !oldB697HasCombat(steps) {
		t.Errorf("the target's next turn had no combat: %v", steps)
	}
}

func TestS82OldB697MomentOfSilenceAfterCombatDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.CastCatalogSpell(t, g, "Moment of Silence", "Instant", oldB697Moment, oldB697PlayerTarget(me))
	effectstest.PassPriorityAroundTable(t, g)
	oldB697TurnSteps(t, g)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	if steps := oldB697TurnSteps(t, g); !oldB697HasCombat(steps) {
		t.Errorf("the next turn had no combat: %v", steps)
	}
}

func TestS82OldB697SpectersWailDiscardsOneAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	for opp.Hand.Size() > 0 {
		opp.Hand.PopTop()
	}
	opp.Hand.PushTop(game.NewCard("A", opp.ID))
	opp.Hand.PushTop(game.NewCard("B", opp.ID))
	effectstest.CastCatalogSpell(t, g, "Specter's Wail", "Sorcery", oldB697Wail, oldB697PlayerTarget(opp))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 1 || opp.Graveyard.Size() != 1 {
		t.Errorf("hand %d graveyard %d, want 1 and 1", opp.Hand.Size(), opp.Graveyard.Size())
	}
	if err := g.CastSpell(g.Seats[0].ID, uuid.New(), game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast a card that is not in hand")
	}
}

func TestS82OldB697ExtortionDiscardsUpToTwoChosenCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for opp.Hand.Size() > 0 {
		opp.Hand.PopTop()
	}
	keep := effectstest.HandSpell(opp, "Keep", "Sorcery", "")
	a := effectstest.HandSpell(opp, "Toss A", "Sorcery", "")
	b := effectstest.HandSpell(opp, "Toss B", "Sorcery", "")
	effectstest.CastCatalogSpell(t, g, "Extortion", "Sorcery", oldB697Extortion, oldB697PlayerTarget(opp))
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if len(pick.DiscardOptions) != 3 || pick.FromPlayer != opp.ID {
		t.Fatalf("pick offers %d cards from %v, want the target's 3", len(pick.DiscardOptions), pick.FromPlayer)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{a, b, keep}); err == nil {
		t.Fatal("three cards were accepted for 'up to two'")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolvePendingChoice: %v", err)
	}
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) || !opp.Hand.Contains(keep) {
		t.Errorf("graveyard has a=%v b=%v, hand has keep=%v", opp.Graveyard.Contains(a), opp.Graveyard.Contains(b), opp.Hand.Contains(keep))
	}
}

func TestS82OldB697SpontaneousGenerationMakesASaprolingPerCardInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for me.Hand.Size() > 0 {
		me.Hand.PopTop()
	}
	for i := 0; i < 3; i++ {
		effectstest.HandSpell(me, "Filler", "Sorcery", "")
	}
	effectstest.CastCatalogSpell(t, g, "Spontaneous Generation", "Sorcery", oldB697Generation, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Saproling" && c.Controller == me.ID && c.Power == 1 && c.Toughness == 1 {
			n++
		}
	}
	if n != 3 {
		t.Errorf("%d Saprolings, want 3 (the spell is not counted)", n)
	}
}

func TestS82OldB697KarnsTouchAnimatesAnArtifactForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{4}"})
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Karn's Touch", "Instant", oldB697Touch, effectstest.DgCardRef(bear)); err == nil {
		t.Error("targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Karn's Touch", "Instant", oldB697Touch, effectstest.DgCardRef(rock))
	effectstest.PassPriorityAroundTable(t, g)
	check := func(when string, want bool) {
		t.Helper()
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		c := effectstest.FindBattlefieldCardForTest(g, rock)
		if c == nil {
			t.Fatalf("%s: the rock is gone", when)
		}
		eff := c.Effective()
		if c.IsCreature() != want || !c.IsArtifact() {
			t.Fatalf("%s: creature %v artifact %v, want creature %v", when, c.IsCreature(), c.IsArtifact(), want)
		}
		if want && (eff.Power != 4 || eff.Toughness != 4) {
			t.Errorf("%s: %d/%d, want 4/4 (its mana value)", when, eff.Power, eff.Toughness)
		}
	}
	check("as it resolves", true)
	effectstest.AdvancePastCleanupForTest(t, g)
	check("after the turn", false)
}

func TestS82OldB697ShovingMatchGrantsEveryCreatureATap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Shoving Match", "Instant", oldB697ShovingMatch, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if idx, _ := effectstest.GrantedActivatedIndex(t, g, id); idx < 0 {
			t.Fatal("a creature has no granted tap ability")
		}
	}
	late := effectstest.B16Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2, "G")
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, late); idx >= 0 {
		t.Error("a creature that entered afterwards got the ability (CR 611.2c)")
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, mine)
	effectstest.UntapForTest(g, mine)
	effectstest.B16Activate(t, g, me.ID, mine, idx, game.ActivateAbilityParams{Ref: ref, Targets: effectstest.DgCardRef(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, theirs); !tapped {
		t.Error("the target creature was not tapped")
	}
	if !effectstest.B16Tapped(t, g, mine) {
		t.Error("the {T} cost did not tap the granted creature")
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, mine); idx >= 0 {
		t.Error("the granted ability outlived the turn")
	}
}

func TestS82OldB697TectonicBreakMakesEveryPlayerSacrificeXLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	lands := make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		for j := 0; j < 3; j++ {
			lands[i] = append(lands[i], effectstest.PushTestPermanent(g, p.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}))
		}
	}
	bear := effectstest.B16Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastXSpell(t, g, "Tectonic Break", "Sorcery", oldB697Tectonic, "{X}{R}{R}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	active := g.Turn.ActiveSeat
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		if effectstest.SacrificeChoiceFor(g, p.ID) == nil {
			t.Fatalf("seat %s was not asked for lands", p.Name)
		}
		effectstest.AnswerSacrifice(t, g, p.ID, lands[seat][0], lands[seat][1])
	}
	for i := range g.Seats {
		if g.Battlefield.Contains(lands[i][0]) || g.Battlefield.Contains(lands[i][1]) || !g.Battlefield.Contains(lands[i][2]) {
			t.Errorf("seat %d: the wrong lands were sacrificed", i)
		}
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was sacrificed")
	}
}

func TestS82OldB697ClearTheLandPutsLandsInTappedAndExilesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	landsIn := make([][]uuid.UUID, len(g.Seats))
	restOut := make([][]uuid.UUID, len(g.Seats))
	sixth := make([]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		sixth[i] = effectstest.SeedLibraryTop(p, "Sixth", "Land", "")
		for j := 0; j < 2; j++ {
			restOut[i] = append(restOut[i], effectstest.SeedLibraryTop(p, "Bear", "Creature — Bear", "{1}{G}"))
		}
		for j := 0; j < 3; j++ {
			landsIn[i] = append(landsIn[i], effectstest.SeedLibraryTop(p, "Forest", "Basic Land — Forest", ""))
		}
	}
	effectstest.CastCatalogSpell(t, g, "Clear the Land", "Sorcery", oldB697ClearLand, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		for _, id := range landsIn[i] {
			c, ok := effectstest.BattlefieldCard(g, id)
			if !ok || !c.Tapped || c.Controller != p.ID {
				t.Errorf("seat %d: land on battlefield=%v tapped=%v, want it in tapped", i, ok, c.Tapped)
			}
		}
		for _, id := range restOut[i] {
			if g.Battlefield.Contains(id) || !g.Exile.Contains(id) {
				t.Errorf("seat %d: a nonland card was not exiled", i)
			}
		}
		if !p.Library.Contains(sixth[i]) {
			t.Errorf("seat %d: the sixth card left the library", i)
		}
	}
}
