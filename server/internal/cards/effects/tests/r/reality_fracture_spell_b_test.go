package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func rfbModalSorcery(t *testing.T, g *game.Game, name, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.HandCardFull(me, name, "Sorcery", "", oracle, nil)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func rfbPlayerTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func TestPredictivePreparationsCountersOneOrTwoAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.SeedCreature(g, "A", me.ID)
	b := effectstest.SeedCreature(g, "B", me.ID)
	effectstest.CastCatalogSpell(t, g, "Predictive Preparations", "Sorcery", effectstest.RfbPredictivePreparationsOracle, append(effectstest.RfCardTarget(a), effectstest.RfCardTarget(b)...))
	effectstest.PassPriorityAroundTable(t, g)
	if rfCounters(t, g, a) != 1 || rfCounters(t, g, b) != 1 {
		t.Errorf("counters = %d, %d, want 1 each", rfCounters(t, g, a), rfCounters(t, g, b))
	}

	g2 := effectstest.NewCatalogGame(t)
	c := effectstest.SeedCreature(g2, "C", g2.Seats[0].ID)
	id := effectstest.SeedGraveyardCard(t, g2, "Predictive Preparations", "Sorcery", effectstest.RfbPredictivePreparationsOracle)
	if err := g2.CastSpell(g2.Seats[0].ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: effectstest.RfCardTarget(c)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if rfCounters(t, g2, c) != 1 || !g2.Exile.Contains(id) {
		t.Errorf("flashback should counter and exile (counters %d)", rfCounters(t, g2, c))
	}
}

func TestPredictivePreparationsRefusesThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var ts []game.TargetRef
	for _, n := range []string{"A", "B", "C"} {
		ts = append(ts, effectstest.RfCardTarget(effectstest.SeedCreature(g, n, me.ID))...)
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Predictive Preparations", "Sorcery", effectstest.RfbPredictivePreparationsOracle, ts); err == nil {
		t.Error("three targets must be refused")
	}
}

func TestProphesiedEndDrawsOnlyIfNotAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	idle := effectstest.SeedCreature(g, "Idle", opp.ID)
	before := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Prophesied End", "Instant", effectstest.RfbProphesiedEndOracle, effectstest.RfCardTarget(idle))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(idle) {
		t.Error("the creature should be destroyed")
	}
	if got := opp.Hand.Size(); got != before+1 {
		t.Errorf("controller hand = %d, want %d (drew for a non-attacker)", got, before+1)
	}

	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	attacker := effectstest.SeedCreature(g2, "Attacker", opp2.ID)
	g2.WithWriteLock(func() {
		for i := range g2.Battlefield.Cards {
			if g2.Battlefield.Cards[i].InstanceID == attacker {
				g2.Battlefield.Cards[i].AttackingTarget = g2.Seats[0].ID
			}
		}
	})
	before2 := opp2.Hand.Size()
	effectstest.CastCatalogSpell(t, g2, "Prophesied End", "Instant", effectstest.RfbProphesiedEndOracle, effectstest.RfCardTarget(attacker))
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(attacker) {
		t.Error("the attacker should be destroyed")
	}
	if got := opp2.Hand.Size(); got != before2 {
		t.Errorf("controller hand = %d, want %d (no draw for an attacker)", got, before2)
	}
}

func TestRecursiveRecruitmentMakesTwoCadetsAndScalesOnFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Recursive Recruitment", "Sorcery", effectstest.RfbRecursiveRecruitmentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Cadet"); got != 2 {
		t.Fatalf("Cadets = %d, want 2", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	for i := 0; i < 6; i++ {
		effectstest.PushGraveyardCardForTest(me2, "Filler")
	}
	id := effectstest.SeedGraveyardCard(t, g2, "Recursive Recruitment", "Sorcery", effectstest.RfbRecursiveRecruitmentOracle)
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	n := 0
	for _, c := range g2.Battlefield.Cards {
		if c.Name == "Cadet" && c.Controller == me2.ID {
			n++
			// six fillers (+ the cast card is exiled, not counted) = two counters
			if got := c.CurrentPower(); got != 4 {
				t.Errorf("flashed-back Cadet power = %d, want 4", got)
			}
		}
	}
	if n != 2 {
		t.Errorf("Cadets = %d, want 2", n)
	}
}

func TestRestoreWithEmpathyReturnsAPermanentAndGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	perm := effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Restore with Empathy", "Instant", effectstest.RfbRestoreWithEmpathyOracle, effectstest.RfCardTarget(perm))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(perm) {
		t.Error("the permanent card should return to hand")
	}
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d", me.Life, life+4)
	}
}

func TestRestoreWithEmpathyRefusesAnInstantCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.PushGraveyardCardTyped(me, "Old Bolt", "Instant")
	if err := effectstest.CastCatalogSpellErr(t, g, "Restore with Empathy", "Instant", effectstest.RfbRestoreWithEmpathyOracle, effectstest.RfCardTarget(spell)); err == nil {
		t.Error("an instant card is not a permanent card")
	}
}

func TestReturnToTheLightRealmsReturnsNonlandPermanentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "")
	rock := effectstest.PushGraveyardPermanent(me, "Rock", "Artifact", "")
	land := effectstest.PushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	bolt := effectstest.PushGraveyardCardTyped(me, "Bolt", "Instant")
	effectstest.CastCatalogSpell(t, g, "Return to the Light Realms", "Sorcery", effectstest.RfbReturnToTheLightRealmsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("the creature and the artifact should return")
	}
	if g.Battlefield.Contains(land) || g.Battlefield.Contains(bolt) {
		t.Error("a land card and an instant card must stay in the graveyard")
	}
}

func TestRiseOfTheDeathbringerDrawsAndLosesLifeByGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.SeedCreature(g, "Small", me.ID)
	hand, life := me.Hand.Size(), me.Life
	if err := effectstest.RfCastModal(t, g, "Rise of the Deathbringer", effectstest.RfbRiseOfTheDeathbringerOracle, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand = %d, want %d (draw 3)", got, hand+3)
	}
	if me.Life != life-3 {
		t.Errorf("life = %d, want %d", me.Life, life-3)
	}
}

func TestRiseOfTheDeathbringerMinusThreeHitsEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	small := effectstest.SeedCreature(g, "Small", g.Seats[0].ID)
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", Power: 5, Toughness: 5,
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := effectstest.RfCastModal(t, g, "Rise of the Deathbringer", effectstest.RfbRiseOfTheDeathbringerOracle, []int{1}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 should die")
	}
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the 5/5 = %d/%d, want 2/2", p, tough)
	}
}

func TestSomethingWorthSavingMillsFourAndTakesAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	ids := effectstest.RfbCastSomethingWorthSaving(t, g,
		effectstest.Mm2Spell("Spell A"), game.Card{Name: "Prize", TypeLine: "Artifact"}, effectstest.Mm2Spell("Spell B"), effectstest.Mm2Spell("Spell C"), effectstest.Mm2Spell("Deep"))
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the permanent card should be in hand")
	}
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}
	if me.Library.Size() != 1 || !me.Graveyard.Contains(ids[0]) || me.Graveyard.Contains(ids[4]) {
		t.Errorf("exactly the top four should be milled (library %d)", me.Library.Size())
	}
}

func TestSomethingWorthSavingWithNoPermanentStillGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.RfbCastSomethingWorthSaving(t, g, effectstest.Mm2Spell("A"), effectstest.Mm2Spell("B"), effectstest.Mm2Spell("C"), effectstest.Mm2Spell("D"))
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no pick should be offered when nothing milled is a permanent")
	}
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}
}

func TestSphinxsApproachExilesFiveAndFetchesASphinx(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sphinx := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Sphinx of Testing", TypeLine: "Creature — Sphinx", Power: 3, Toughness: 3})
	var copies []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Sphinx's Approach", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		copies = append(copies, id)
	}
	spell := effectstest.CastCatalogSpell(t, g, "Sphinx's Approach", "Instant", effectstest.RfbSphinxsApproachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, copies...)
	if pick := effectstest.SearchChoiceFor(g, me.ID); pick != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, sphinx)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sphinx) {
		t.Error("the Sphinx should be on the battlefield")
	}
	for _, id := range append(copies, spell) {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s should be exiled", id)
		}
	}
}

func TestSphinxsApproachAsksNothingWithFewerThanFourCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Sphinx's Approach", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	spell := effectstest.CastCatalogSpell(t, g, "Sphinx's Approach", "Instant", effectstest.RfbSphinxsApproachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no question should be asked with only three copies")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the spell should go to the graveyard as usual")
	}
}

func TestSphinxsApproachDeclinedWhenNotExactlyFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var copies []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Sphinx's Approach", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		copies = append(copies, id)
	}
	effectstest.CastCatalogSpell(t, g, "Sphinx's Approach", "Instant", effectstest.RfbSphinxsApproachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("expected a prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, copies[:2]); err == nil {
		t.Error("two copies is not a legal answer")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	for _, id := range copies {
		if !me.Graveyard.Contains(id) {
			t.Error("declining leaves the graveyard alone")
		}
	}
}

func TestStingerquillCharmModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	if err := effectstest.RfCastModal(t, g, "Stingerquill Charm", effectstest.RfbStingerquillCharmOracle, []int{0}, rfbPlayerTarget(opp.ID)); err != nil {
		t.Fatalf("mode 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life = %d, want %d", opp.Life, life-3)
	}

	g2 := effectstest.NewCatalogGame(t)
	c := effectstest.SeedCreature(g2, "Bear", g2.Seats[0].ID)
	if err := effectstest.RfCastModal(t, g2, "Stingerquill Charm", effectstest.RfbStingerquillCharmOracle, []int{1}, effectstest.RfCardTarget(c)); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	ab := effectstest.EffectiveAbilities(t, g2, c)
	if !effectstest.HasAbility(ab, "first strike") || !effectstest.HasAbility(ab, "deathtouch") {
		t.Errorf("abilities = %v, want first strike and deathtouch", ab)
	}

	g3 := effectstest.NewCatalogGame(t)
	if err := effectstest.RfCastModal(t, g3, "Stingerquill Charm", effectstest.RfbStingerquillCharmOracle, []int{2}, nil); err != nil {
		t.Fatalf("mode 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g3)
	cadet := effectstest.FindBattlefieldByName(g3, "Cadet")
	if cadet == uuid.Nil {
		t.Fatal("no Cadet")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g3, cadet), "haste") {
		t.Error("the Cadet should have haste this turn")
	}
}

func TestStingingVitriolDamagesAndTakesANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	land := effectstest.HandCardFull(opp, "Forest", "Basic Land — Forest", "", "", nil)
	spell := effectstest.HandCardFull(opp, "Prize", "Sorcery", "", "", nil)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Stinging Vitriol", "Sorcery", effectstest.RfbStingingVitriolOracle, rfbPlayerTarget(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life = %d, want %d", opp.Life, life-2)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("expected a revealed-hand pick, got %d choices", len(g.PendingChoices))
	}
	if err := g.ResolvePendingChoice(g.PendingChoices[0].ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Error("a land is not a legal pick")
	}
	if err := g.ResolvePendingChoice(g.PendingChoices[0].ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !opp.Graveyard.Contains(spell) || !opp.Hand.Contains(land) {
		t.Error("the nonland card should be discarded and the land kept")
	}
}

func TestStingingVitriolRefusesYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := effectstest.CastCatalogSpellErr(t, g, "Stinging Vitriol", "Sorcery", effectstest.RfbStingingVitriolOracle, rfbPlayerTarget(g.Seats[0].ID)); err == nil {
		t.Error("only an opponent is a legal target")
	}
}

func TestSurgicalPrecisionModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.SeedCreature(g, "Small", g.Seats[1].ID)
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", Power: 4, Toughness: 4,
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := rfbModalSorcery(t, g, "Surgical Precision", effectstest.RfbSurgicalPrecisionOracle, []int{0}, effectstest.RfCardTarget(small)); err == nil {
		t.Error("a 2/2 is not a legal target")
	}
	life := me.Life
	if err := rfbModalSorcery(t, g, "Surgical Precision", effectstest.RfbSurgicalPrecisionOracle, []int{0}, effectstest.RfCardTarget(big)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4/4 should be destroyed")
	}
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	hand, life2 := me2.Hand.Size(), me2.Life
	if err := rfbModalSorcery(t, g2, "Surgical Precision", effectstest.RfbSurgicalPrecisionOracle, []int{1}, nil); err != nil {
		t.Fatalf("cast mode 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (drew one)", me2.Hand.Size(), hand+1)
	}
	if me2.Life != life2+2 {
		t.Errorf("life = %d, want %d", me2.Life, life2+2)
	}
}

func TestTethermagesAdvantagePumpsGivesReachAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	c := effectstest.SeedCreature(g, "Bear", g.Seats[0].ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == c {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Tethermage's Advantage", "Instant", effectstest.RfbTethermagesAdvantageOracle, effectstest.RfCardTarget(c))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, c), effectstest.EffectiveToughness(t, g, c); p != 4 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, c), "reach") {
		t.Error("should gain reach")
	}
	if card, _ := g.LookupCardForEffect(c); card.Tapped {
		t.Error("the creature should be untapped")
	}
}

func TestTheorixCharmModes(t *testing.T) {
	// -2/-2
	g := effectstest.NewCatalogGame(t)
	c := effectstest.SeedCreature(g, "Bear", g.Seats[1].ID)
	if err := effectstest.RfCastModal(t, g, "Theorix Charm", effectstest.RfbTheorixCharmOracle, []int{1}, effectstest.RfCardTarget(c)); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c) {
		t.Error("a 2/2 given -2/-2 should die")
	}

	// mill three, then draw
	g2 := effectstest.NewCatalogGame(t)
	me := g2.Seats[0]
	lib, hand := me.Library.Size(), me.Hand.Size()
	if err := effectstest.RfCastModal(t, g2, "Theorix Charm", effectstest.RfbTheorixCharmOracle, []int{2}, nil); err != nil {
		t.Fatalf("mode 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if me.Library.Size() != lib-4 {
		t.Errorf("library = %d, want %d (three milled, one drawn)", me.Library.Size(), lib-4)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (drew one)", me.Hand.Size(), hand+1)
	}
}

func TestTheorixCharmCounterIsRefusedOnACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	bear := effectstest.HandCardFull(me, "Bear Cub", "Creature — Bear", "", "", []string{"G"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	if err := effectstest.RfCastModal(t, g, "Theorix Charm", effectstest.RfbTheorixCharmOracle, []int{0}, effectstest.RfCardTarget(bear)); err == nil {
		t.Error("a creature spell is not a legal target")
	}
}

func TestTheorixCharmCountersAnUnpaidNoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	spell := effectstest.HandCardFull(me, "Plain Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := effectstest.RfCastModal(t, g, "Theorix Charm", effectstest.RfbTheorixCharmOracle, []int{0}, effectstest.RfCardTarget(spell)); err != nil {
		t.Fatalf("cast charm: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) {
		t.Error("declining the tax should counter the spell")
	}
}

func TestTwinnedVisionDrawsOneFromHandTwoFromFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Twinned Vision", "Instant", effectstest.RfbTwinnedVisionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (one drawn net of the cast)", me.Hand.Size(), hand+1)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	fodder := effectstest.HandCardFull(me2, "Fodder", "Sorcery", "", "", nil)
	id := effectstest.SeedGraveyardCard(t, g2, "Twinned Vision", "Instant", effectstest.RfbTwinnedVisionOracle)
	hand2 := me2.Hand.Size()
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", AltCostIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand2-1+2 {
		t.Errorf("hand = %d, want %d (discard one, draw two)", me2.Hand.Size(), hand2+1)
	}
	if !me2.Graveyard.Contains(fodder) || !g2.Exile.Contains(id) {
		t.Error("the discarded card goes to the graveyard and the flashed-back card to exile")
	}
}

func TestTwinnedVisionFlashbackNeedsACardToDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	id := effectstest.SeedGraveyardCard(t, g, "Twinned Vision", "Instant", effectstest.RfbTwinnedVisionOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err == nil {
		t.Error("flashback without naming a card to discard must be refused")
	}
}

func TestTwistedFatesDestroysAndCountersTheChosenPlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	mine := effectstest.SeedCreature(g, "Mine", g.Seats[0].ID)
	mine2 := effectstest.SeedCreature(g, "Mine2", g.Seats[0].ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	effectstest.CastCatalogSpell(t, g, "Twisted Fates", "Sorcery", effectstest.RfbTwistedFatesOracle,
		append(effectstest.RfCardTarget(rock), rfbPlayerTarget(g.Seats[0].ID)...))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
	if rfCounters(t, g, mine) != 1 || rfCounters(t, g, mine2) != 1 {
		t.Error("each creature the chosen player controls gets a counter")
	}
	if rfCounters(t, g, theirs) != 0 {
		t.Error("another player's creature gets none")
	}
}

func TestTwistedFatesRefusesALandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Twisted Fates", "Sorcery", effectstest.RfbTwistedFatesOracle,
		append(effectstest.RfCardTarget(land), rfbPlayerTarget(g.Seats[0].ID)...)); err == nil {
		t.Error("a land is not a legal target")
	}
}

func TestVigorbloomCharmModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	c := effectstest.SeedCreature(g, "Bear", g.Seats[0].ID)
	if err := effectstest.RfCastModal(t, g, "Vigorbloom Charm", effectstest.RfbVigorbloomCharmOracle, []int{0}, effectstest.RfCardTarget(c)); err != nil {
		t.Fatalf("mode 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ab := effectstest.EffectiveAbilities(t, g, c)
	if !effectstest.HasAbility(ab, "hexproof") || !effectstest.HasAbility(ab, "indestructible") {
		t.Errorf("abilities = %v", ab)
	}

	g2 := effectstest.NewCatalogGame(t)
	me := g2.Seats[0]
	life := me.Life
	if err := effectstest.RfCastModal(t, g2, "Vigorbloom Charm", effectstest.RfbVigorbloomCharmOracle, []int{1}, nil); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
}

func TestVigorbloomCharmCountersThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := effectstest.SeedCreature(g, "Mine", g.Seats[0].ID)
	theirs := effectstest.SeedCreature(g, "Theirs", g.Seats[1].ID)
	targets := append(effectstest.RfCardTarget(mine), effectstest.RfCardTarget(theirs)...)
	if err := effectstest.RfCastModal(t, g, "Vigorbloom Charm", effectstest.RfbVigorbloomCharmOracle, []int{2}, targets); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.RfPT(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("mine = %d/%d, want 3/3", p, tough)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3-power fighter should kill the 2/2 (the counter lands first)")
	}
}

func TestVigorbloomCharmRefusesFightingYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.SeedCreature(g, "A", g.Seats[0].ID)
	b := effectstest.SeedCreature(g, "B", g.Seats[0].ID)
	if err := effectstest.RfCastModal(t, g, "Vigorbloom Charm", effectstest.RfbVigorbloomCharmOracle, []int{2}, append(effectstest.RfCardTarget(a), effectstest.RfCardTarget(b)...)); err == nil {
		t.Error("the second creature must be an opponent's")
	}
}
