package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_spell_b_test.go — the Reality Fracture instants and
// sorceries of slice fra-spell-b.

const (
	rfbPredictivePreparationsOracle = "d06a5642-861b-4b0d-9ea5-07046b7e1d37"
	rfbProphesiedEndOracle          = "45f4d057-5134-40a2-9590-ddba73a65582"
	rfbRecursiveRecruitmentOracle   = "b73e35b0-b8c9-471a-bdff-f7dc8f453d7a"
	rfbRestoreWithEmpathyOracle     = "cb1e697f-6689-4f0b-8725-61d51432b8dd"
	rfbReturnToTheLightRealmsOracle = "217e76ca-ea94-417c-b5b0-82e2302eeee8"
	rfbRiseOfTheDeathbringerOracle  = "57ea9f63-8267-4303-a3dd-9ded220294dc"
	rfbSomethingWorthSavingOracle   = "46cf0e3e-3c3e-4eb8-a71a-4b4b87744ced"
	rfbSphinxsApproachOracle        = "48b1ac27-430f-4fd9-a71b-6ad125c57fe3"
	rfbStingerquillCharmOracle      = "987aadac-d6a6-4e50-9468-9e8e53cc5529"
	rfbStingingVitriolOracle        = "624cf822-0b4e-4f0c-9893-a53cb094f616"
	rfbSurgicalPrecisionOracle      = "0caccec3-4942-499a-9434-0c94389afcf8"
	rfbTethermagesAdvantageOracle   = "b4f89885-eeb9-46c7-9820-9758b624ca24"
	rfbTheorixCharmOracle           = "cf527bbd-e898-4aa9-909d-daec2f4b62ad"
	rfbTwinnedVisionOracle          = "fa1cc78f-26bc-4d5d-973b-7cb7d7490973"
	rfbTwistedFatesOracle           = "3ce30a19-e716-478f-be89-e2d95753cfed"
	rfbVigorbloomCharmOracle        = "826f17d1-3806-4579-8b1b-2082b7b03d69"
)

func rfbModalSorcery(t *testing.T, g *game.Game, name, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	toMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := handCardFull(me, name, "Sorcery", "", oracle, nil)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func rfbPlayerTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func TestPredictivePreparationsCountersOneOrTwoAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := seedCreature(g, "A", me.ID)
	b := seedCreature(g, "B", me.ID)
	castCatalogSpell(t, g, "Predictive Preparations", "Sorcery", rfbPredictivePreparationsOracle, append(rfCardTarget(a), rfCardTarget(b)...))
	passPriorityAroundTable(t, g)
	if rfCounters(t, g, a) != 1 || rfCounters(t, g, b) != 1 {
		t.Errorf("counters = %d, %d, want 1 each", rfCounters(t, g, a), rfCounters(t, g, b))
	}

	g2 := newCatalogGame(t)
	c := seedCreature(g2, "C", g2.Seats[0].ID)
	id := seedGraveyardCard(t, g2, "Predictive Preparations", "Sorcery", rfbPredictivePreparationsOracle)
	if err := g2.CastSpell(g2.Seats[0].ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: rfCardTarget(c)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if rfCounters(t, g2, c) != 1 || !g2.Exile.Contains(id) {
		t.Errorf("flashback should counter and exile (counters %d)", rfCounters(t, g2, c))
	}
}

func TestPredictivePreparationsRefusesThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var ts []game.TargetRef
	for _, n := range []string{"A", "B", "C"} {
		ts = append(ts, rfCardTarget(seedCreature(g, n, me.ID))...)
	}
	if err := castCatalogSpellErr(t, g, "Predictive Preparations", "Sorcery", rfbPredictivePreparationsOracle, ts); err == nil {
		t.Error("three targets must be refused")
	}
}

func TestProphesiedEndDrawsOnlyIfNotAttacking(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	idle := seedCreature(g, "Idle", opp.ID)
	before := opp.Hand.Size()
	castCatalogSpell(t, g, "Prophesied End", "Instant", rfbProphesiedEndOracle, rfCardTarget(idle))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(idle) {
		t.Error("the creature should be destroyed")
	}
	if got := opp.Hand.Size(); got != before+1 {
		t.Errorf("controller hand = %d, want %d (drew for a non-attacker)", got, before+1)
	}

	g2 := newCatalogGame(t)
	opp2 := g2.Seats[1]
	attacker := seedCreature(g2, "Attacker", opp2.ID)
	g2.WithWriteLock(func() {
		for i := range g2.Battlefield.Cards {
			if g2.Battlefield.Cards[i].InstanceID == attacker {
				g2.Battlefield.Cards[i].AttackingTarget = g2.Seats[0].ID
			}
		}
	})
	before2 := opp2.Hand.Size()
	castCatalogSpell(t, g2, "Prophesied End", "Instant", rfbProphesiedEndOracle, rfCardTarget(attacker))
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(attacker) {
		t.Error("the attacker should be destroyed")
	}
	if got := opp2.Hand.Size(); got != before2 {
		t.Errorf("controller hand = %d, want %d (no draw for an attacker)", got, before2)
	}
}

func TestRecursiveRecruitmentMakesTwoCadetsAndScalesOnFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Recursive Recruitment", "Sorcery", rfbRecursiveRecruitmentOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Cadet"); got != 2 {
		t.Fatalf("Cadets = %d, want 2", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	for i := 0; i < 6; i++ {
		pushGraveyardCardForTest(me2, "Filler")
	}
	id := seedGraveyardCard(t, g2, "Recursive Recruitment", "Sorcery", rfbRecursiveRecruitmentOracle)
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g2)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	perm := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "")
	life := me.Life
	castCatalogSpell(t, g, "Restore with Empathy", "Instant", rfbRestoreWithEmpathyOracle, rfCardTarget(perm))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(perm) {
		t.Error("the permanent card should return to hand")
	}
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d", me.Life, life+4)
	}
}

func TestRestoreWithEmpathyRefusesAnInstantCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := pushGraveyardCardTyped(me, "Old Bolt", "Instant")
	if err := castCatalogSpellErr(t, g, "Restore with Empathy", "Instant", rfbRestoreWithEmpathyOracle, rfCardTarget(spell)); err == nil {
		t.Error("an instant card is not a permanent card")
	}
}

func TestReturnToTheLightRealmsReturnsNonlandPermanentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "")
	rock := pushGraveyardPermanent(me, "Rock", "Artifact", "")
	land := pushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	bolt := pushGraveyardCardTyped(me, "Bolt", "Instant")
	castCatalogSpell(t, g, "Return to the Light Realms", "Sorcery", rfbReturnToTheLightRealmsOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("the creature and the artifact should return")
	}
	if g.Battlefield.Contains(land) || g.Battlefield.Contains(bolt) {
		t.Error("a land card and an instant card must stay in the graveyard")
	}
}

func TestRiseOfTheDeathbringerDrawsAndLosesLifeByGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	seedCreature(g, "Small", me.ID)
	hand, life := me.Hand.Size(), me.Life
	if err := rfCastModal(t, g, "Rise of the Deathbringer", rfbRiseOfTheDeathbringerOracle, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand = %d, want %d (draw 3)", got, hand+3)
	}
	if me.Life != life-3 {
		t.Errorf("life = %d, want %d", me.Life, life-3)
	}
}

func TestRiseOfTheDeathbringerMinusThreeHitsEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	small := seedCreature(g, "Small", g.Seats[0].ID)
	big := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", Power: 5, Toughness: 5,
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := rfCastModal(t, g, "Rise of the Deathbringer", rfbRiseOfTheDeathbringerOracle, []int{1}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 should die")
	}
	if p, tough := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the 5/5 = %d/%d, want 2/2", p, tough)
	}
}

func rfbCastSomethingWorthSaving(t *testing.T, g *game.Game, lib ...game.Card) []uuid.UUID {
	t.Helper()
	toMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := mm2Library(me, lib...)
	id := handCardFull(me, "Something Worth Saving", "Instant", "", rfbSomethingWorthSavingOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	return ids
}

func TestSomethingWorthSavingMillsFourAndTakesAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	ids := rfbCastSomethingWorthSaving(t, g,
		mm2Spell("Spell A"), game.Card{Name: "Prize", TypeLine: "Artifact"}, mm2Spell("Spell B"), mm2Spell("Spell C"), mm2Spell("Deep"))
	answerChooseCards(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	rfbCastSomethingWorthSaving(t, g, mm2Spell("A"), mm2Spell("B"), mm2Spell("C"), mm2Spell("D"))
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no pick should be offered when nothing milled is a permanent")
	}
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}
}

func TestSphinxsApproachExilesFiveAndFetchesASphinx(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sphinx := pushLibraryCardForTest(me, game.Card{Name: "Sphinx of Testing", TypeLine: "Creature — Sphinx", Power: 3, Toughness: 3})
	var copies []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Sphinx's Approach", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		copies = append(copies, id)
	}
	spell := castCatalogSpell(t, g, "Sphinx's Approach", "Instant", rfbSphinxsApproachOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, copies...)
	if pick := searchChoiceFor(g, me.ID); pick != nil {
		answerSearchByID(t, g, me.ID, sphinx)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Sphinx's Approach", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	spell := castCatalogSpell(t, g, "Sphinx's Approach", "Instant", rfbSphinxsApproachOracle, nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no question should be asked with only three copies")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the spell should go to the graveyard as usual")
	}
}

func TestSphinxsApproachDeclinedWhenNotExactlyFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var copies []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Sphinx's Approach", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		copies = append(copies, id)
	}
	castCatalogSpell(t, g, "Sphinx's Approach", "Instant", rfbSphinxsApproachOracle, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	if err := rfCastModal(t, g, "Stingerquill Charm", rfbStingerquillCharmOracle, []int{0}, rfbPlayerTarget(opp.ID)); err != nil {
		t.Fatalf("mode 0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life = %d, want %d", opp.Life, life-3)
	}

	g2 := newCatalogGame(t)
	c := seedCreature(g2, "Bear", g2.Seats[0].ID)
	if err := rfCastModal(t, g2, "Stingerquill Charm", rfbStingerquillCharmOracle, []int{1}, rfCardTarget(c)); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	passPriorityAroundTable(t, g2)
	ab := effectiveAbilities(t, g2, c)
	if !hasAbility(ab, "first strike") || !hasAbility(ab, "deathtouch") {
		t.Errorf("abilities = %v, want first strike and deathtouch", ab)
	}

	g3 := newCatalogGame(t)
	if err := rfCastModal(t, g3, "Stingerquill Charm", rfbStingerquillCharmOracle, []int{2}, nil); err != nil {
		t.Fatalf("mode 2: %v", err)
	}
	passPriorityAroundTable(t, g3)
	cadet := findBattlefieldByName(g3, "Cadet")
	if cadet == uuid.Nil {
		t.Fatal("no Cadet")
	}
	if !hasAbility(effectiveAbilities(t, g3, cadet), "haste") {
		t.Error("the Cadet should have haste this turn")
	}
}

func TestStingingVitriolDamagesAndTakesANonland(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	land := handCardFull(opp, "Forest", "Basic Land — Forest", "", "", nil)
	spell := handCardFull(opp, "Prize", "Sorcery", "", "", nil)
	life := opp.Life
	castCatalogSpell(t, g, "Stinging Vitriol", "Sorcery", rfbStingingVitriolOracle, rfbPlayerTarget(opp.ID))
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	if err := castCatalogSpellErr(t, g, "Stinging Vitriol", "Sorcery", rfbStingingVitriolOracle, rfbPlayerTarget(g.Seats[0].ID)); err == nil {
		t.Error("only an opponent is a legal target")
	}
}

func TestSurgicalPrecisionModes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := seedCreature(g, "Small", g.Seats[1].ID)
	big := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", Power: 4, Toughness: 4,
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := rfbModalSorcery(t, g, "Surgical Precision", rfbSurgicalPrecisionOracle, []int{0}, rfCardTarget(small)); err == nil {
		t.Error("a 2/2 is not a legal target")
	}
	life := me.Life
	if err := rfbModalSorcery(t, g, "Surgical Precision", rfbSurgicalPrecisionOracle, []int{0}, rfCardTarget(big)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4/4 should be destroyed")
	}
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d", me.Life, life+1)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	hand, life2 := me2.Hand.Size(), me2.Life
	if err := rfbModalSorcery(t, g2, "Surgical Precision", rfbSurgicalPrecisionOracle, []int{1}, nil); err != nil {
		t.Fatalf("cast mode 1: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (drew one)", me2.Hand.Size(), hand+1)
	}
	if me2.Life != life2+2 {
		t.Errorf("life = %d, want %d", me2.Life, life2+2)
	}
}

func TestTethermagesAdvantagePumpsGivesReachAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	c := seedCreature(g, "Bear", g.Seats[0].ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == c {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	castCatalogSpell(t, g, "Tethermage's Advantage", "Instant", rfbTethermagesAdvantageOracle, rfCardTarget(c))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, c), effectiveToughness(t, g, c); p != 4 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tough)
	}
	if !hasAbility(effectiveAbilities(t, g, c), "reach") {
		t.Error("should gain reach")
	}
	if card, _ := g.LookupCardForEffect(c); card.Tapped {
		t.Error("the creature should be untapped")
	}
}

func TestTheorixCharmModes(t *testing.T) {
	// -2/-2
	g := newCatalogGame(t)
	c := seedCreature(g, "Bear", g.Seats[1].ID)
	if err := rfCastModal(t, g, "Theorix Charm", rfbTheorixCharmOracle, []int{1}, rfCardTarget(c)); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c) {
		t.Error("a 2/2 given -2/-2 should die")
	}

	// mill three, then draw
	g2 := newCatalogGame(t)
	me := g2.Seats[0]
	lib, hand := me.Library.Size(), me.Hand.Size()
	if err := rfCastModal(t, g2, "Theorix Charm", rfbTheorixCharmOracle, []int{2}, nil); err != nil {
		t.Fatalf("mode 2: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me.Library.Size() != lib-4 {
		t.Errorf("library = %d, want %d (three milled, one drawn)", me.Library.Size(), lib-4)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (drew one)", me.Hand.Size(), hand+1)
	}
}

func TestTheorixCharmCounterIsRefusedOnACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	bear := handCardFull(me, "Bear Cub", "Creature — Bear", "", "", []string{"G"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	if err := rfCastModal(t, g, "Theorix Charm", rfbTheorixCharmOracle, []int{0}, rfCardTarget(bear)); err == nil {
		t.Error("a creature spell is not a legal target")
	}
}

func TestTheorixCharmCountersAnUnpaidNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	spell := handCardFull(me, "Plain Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := rfCastModal(t, g, "Theorix Charm", rfbTheorixCharmOracle, []int{0}, rfCardTarget(spell)); err != nil {
		t.Fatalf("cast charm: %v", err)
	}
	passUntilTaxed(t, g, me.ID)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) {
		t.Error("declining the tax should counter the spell")
	}
}

func TestTwinnedVisionDrawsOneFromHandTwoFromFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Twinned Vision", "Instant", rfbTwinnedVisionOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (one drawn net of the cast)", me.Hand.Size(), hand+1)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	fodder := handCardFull(me2, "Fodder", "Sorcery", "", "", nil)
	id := seedGraveyardCard(t, g2, "Twinned Vision", "Instant", rfbTwinnedVisionOracle)
	hand2 := me2.Hand.Size()
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", AltCostIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand2-1+2 {
		t.Errorf("hand = %d, want %d (discard one, draw two)", me2.Hand.Size(), hand2+1)
	}
	if !me2.Graveyard.Contains(fodder) || !g2.Exile.Contains(id) {
		t.Error("the discarded card goes to the graveyard and the flashed-back card to exile")
	}
}

func TestTwinnedVisionFlashbackNeedsACardToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	id := seedGraveyardCard(t, g, "Twinned Vision", "Instant", rfbTwinnedVisionOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err == nil {
		t.Error("flashback without naming a card to discard must be refused")
	}
}

func TestTwistedFatesDestroysAndCountersTheChosenPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	mine := seedCreature(g, "Mine", g.Seats[0].ID)
	mine2 := seedCreature(g, "Mine2", g.Seats[0].ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	castCatalogSpell(t, g, "Twisted Fates", "Sorcery", rfbTwistedFatesOracle,
		append(rfCardTarget(rock), rfbPlayerTarget(g.Seats[0].ID)...))
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := castCatalogSpellErr(t, g, "Twisted Fates", "Sorcery", rfbTwistedFatesOracle,
		append(rfCardTarget(land), rfbPlayerTarget(g.Seats[0].ID)...)); err == nil {
		t.Error("a land is not a legal target")
	}
}

func TestVigorbloomCharmModes(t *testing.T) {
	g := newCatalogGame(t)
	c := seedCreature(g, "Bear", g.Seats[0].ID)
	if err := rfCastModal(t, g, "Vigorbloom Charm", rfbVigorbloomCharmOracle, []int{0}, rfCardTarget(c)); err != nil {
		t.Fatalf("mode 0: %v", err)
	}
	passPriorityAroundTable(t, g)
	ab := effectiveAbilities(t, g, c)
	if !hasAbility(ab, "hexproof") || !hasAbility(ab, "indestructible") {
		t.Errorf("abilities = %v", ab)
	}

	g2 := newCatalogGame(t)
	me := g2.Seats[0]
	life := me.Life
	if err := rfCastModal(t, g2, "Vigorbloom Charm", rfbVigorbloomCharmOracle, []int{1}, nil); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
}

func TestVigorbloomCharmCountersThenFights(t *testing.T) {
	g := newCatalogGame(t)
	mine := seedCreature(g, "Mine", g.Seats[0].ID)
	theirs := seedCreature(g, "Theirs", g.Seats[1].ID)
	targets := append(rfCardTarget(mine), rfCardTarget(theirs)...)
	if err := rfCastModal(t, g, "Vigorbloom Charm", rfbVigorbloomCharmOracle, []int{2}, targets); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := rfPT(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("mine = %d/%d, want 3/3", p, tough)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3-power fighter should kill the 2/2 (the counter lands first)")
	}
}

func TestVigorbloomCharmRefusesFightingYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	a := seedCreature(g, "A", g.Seats[0].ID)
	b := seedCreature(g, "B", g.Seats[0].ID)
	if err := rfCastModal(t, g, "Vigorbloom Charm", rfbVigorbloomCharmOracle, []int{2}, append(rfCardTarget(a), rfCardTarget(b)...)); err == nil {
		t.Error("the second creature must be an opponent's")
	}
}
