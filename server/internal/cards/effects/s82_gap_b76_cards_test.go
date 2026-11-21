package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b76_cards_test.go — S82 GAP batch 76 (#5948), the Ravnica
// Remastered white-blue and white creatures that were missing. Each test plays
// the catalog entry through the engine.

const (
	gapMesmeristOracle      = "f507f9ba-a7e9-4e45-86dc-8c60a72a6234"
	gapAzoriusGuildmageOr   = "16f61f69-cdde-4b6c-b855-c74db9038ceb"
	gapJudgesFamiliarOracle = "22b7975c-37a5-4895-b209-bdeeec0c6aea"
	gapSphinxOracle         = "3bdd3090-4503-4382-9a11-3e1cfcdffd71"
	gapArmoryGuardOracle    = "96c67950-df59-430c-94fd-234086c6c8f2"
	gapMakeshiftOracle      = "8b6e2feb-9e8f-4912-ada0-1614502d9813"
)

func TestGapB76CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gapMesmeristOracle: "Vedalken Mesmerist", gapAzoriusGuildmageOr: "Azorius Guildmage",
		gapJudgesFamiliarOracle: "Judge's Familiar", gapSphinxOracle: "Sphinx of New Prahv",
		gapArmoryGuardOracle: "Armory Guard", gapMakeshiftOracle: "Makeshift Battalion", gapSkyHussarOracle: "Sky Hussar",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v (registered %v)", name, spec, ok)
		}
	}
}

// gapActivatedItemOf is the id of the activated ability of `source` waiting on
// the stack.
func gapActivatedItemOf(g *game.Game, source uuid.UUID) uuid.UUID {
	for id, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated && item.SourceCardID == source {
			return id
		}
	}
	return uuid.Nil
}

// --- Vedalken Mesmerist -----------------------------------------------------------

func TestGapB76VedalkenMesmeristShrinksAnOpposingCreatureWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	mesmerist := b12Push(g, me.ID, "Vedalken Mesmerist", "Creature — Vedalken Wizard", gapMesmeristOracle, 2, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, mesmerist)
	if p := latestPickTarget(g, me.ID); p != nil && slices.Contains(p.PickTargetCards, mine) {
		t.Error("a creature the controller controls was offered")
	}
	answerTriggerTargets(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 3 {
		t.Errorf("the target is %d/%d, want 1/3", p, tough)
	}
}

// --- Azorius Guildmage ----------------------------------------------------------------

func TestGapB76AzoriusGuildmageTapsATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := b12Push(g, me.ID, "Azorius Guildmage", "Creature — Vedalken Wizard", gapAzoriusGuildmageOr, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	addPool(t, g, me.ID, "{C}{C}{W}")
	p7Activate(t, g, me, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	if !tappedNow(g, bear) {
		t.Error("the target is untapped")
	}
	if tappedNow(g, gm) {
		t.Error("the Guildmage tapped itself: the cost has no {T}")
	}
}

func TestGapB76AzoriusGuildmageCountersAnActivatedAbilityButNotASpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := b12Push(g, me.ID, "Azorius Guildmage", "Creature — Vedalken Wizard", gapAzoriusGuildmageOr, 2, 2)
	jailer := b12Push(g, me.ID, "Akroan Jailer", "Creature — Human Soldier", "e3ad2c77-2a41-4a4a-a6bf-e542d0162b88", 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	addPool(t, g, me.ID, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, jailer, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate the Jailer: %v", err)
	}
	ability := gapActivatedItemOf(g, jailer)
	if ability == uuid.Nil {
		t.Fatal("the Jailer's ability is not on the stack")
	}
	addPool(t, g, me.ID, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(ability)}); err != nil {
		t.Fatalf("counter the activated ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tappedNow(g, bear) {
		t.Error("the countered ability still tapped its target")
	}
	spell := castCatalogSpell(t, g, "A Charm", "Instant", "", nil)
	addPool(t, g, me.ID, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(spell)}); err == nil {
		t.Error("countered a spell with an ability that counters abilities")
	}
}

// --- Judge's Familiar ----------------------------------------------------------------------

func TestGapB76JudgesFamiliarCountersAnInstantUnlessItsControllerPaysOne(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fund   bool
		wantGY bool
	}{{"unpaid", false, true}, {"paid", true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			jf := b12Push(g, me.ID, "Judge's Familiar", "Creature — Bird", gapJudgesFamiliarOracle, 1, 1)
			advanceToMain(t, g)
			spell := castCatalogSpell(t, g, "A Charm", "Instant", "", nil)
			if err := g.ActivateCatalogAbility(me.ID, jf, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(spell)}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(jf) {
				t.Fatal("the Familiar was not sacrificed as a cost")
			}
			if tc.fund {
				addPool(t, g, me.ID, "{C}")
			}
			passPriorityAroundTable(t, g)
			if c := answerPayUnless(t, g, me.ID, tc.fund); c == nil {
				t.Fatal("no pay-unless prompt")
			}
			passPriorityAroundTable(t, g)
			if got := p7Countered(g, spell); got != tc.wantGY {
				t.Errorf("countered = %v, want %v", got, tc.wantGY)
			}
		})
	}
}

func TestGapB76JudgesFamiliarCannotTargetACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jf := b12Push(g, me.ID, "Judge's Familiar", "Creature — Bird", gapJudgesFamiliarOracle, 1, 1)
	advanceToMain(t, g)
	spell := castCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, jf, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(spell)}); err == nil {
		t.Error("targeted a creature spell")
	}
	if !hasEffectiveKeyword(t, g, jf, "flying") {
		t.Error("no flying")
	}
}

// --- Sphinx of New Prahv --------------------------------------------------------------------------

func TestGapB76SphinxOfNewPrahvTaxesOpponentsSpellsThatTargetIt(t *testing.T) {
	g, me, opp := p7Table(t)
	sphinx := b12Push(g, me.ID, "Sphinx of New Prahv", "Creature — Sphinx", gapSphinxOracle, 4, 3)
	mine := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, sphinx, kw) {
			t.Errorf("no %s", kw)
		}
	}
	// It is the opponent's spell: make the opponent the caster, in their own main phase.
	for i := 0; i < len(g.Seats); i++ {
		if g.Turn.ActiveSeat == 1 {
			break
		}
		endTurn(t, g)
	}
	advanceToMain(t, g)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: lightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	flashIfFund(g, opp, "R")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: cardTarget(sphinx)}); err == nil {
		t.Fatal("cast a {R} spell at the Sphinx with only {R}")
	}
	flashIfFund(g, opp, "R", "C", "C")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: cardTarget(sphinx)}); err != nil {
		t.Fatalf("with {R}{C}{C}: %v", err)
	}
	// A spell aimed at another creature is not taxed.
	bolt2 := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt2, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: lightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	flashIfFund(g, opp, "R")
	if err := g.CastSpell(opp.ID, bolt2, game.CastSpellParams{Strict: true, Targets: cardTarget(mine)}); err != nil {
		t.Fatalf("an untaxed spell: %v", err)
	}
}

// --- Armory Guard ------------------------------------------------------------------------------------

func TestGapB76ArmoryGuardHasVigilanceOnlyWhileYouControlAGate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := b12Push(g, me.ID, "Armory Guard", "Creature — Giant Soldier", gapArmoryGuardOracle, 2, 5)
	advanceToMain(t, g)
	checkState(t, g)
	if hasEffectiveKeyword(t, g, guard, "vigilance") {
		t.Fatal("vigilance without a Gate")
	}
	b12Permanent(g, opp.ID, "Opposing Gate", "Land — Gate")
	checkState(t, g)
	if hasEffectiveKeyword(t, g, guard, "vigilance") {
		t.Fatal("an opponent's Gate gave vigilance")
	}
	gate := b12Permanent(g, me.ID, "Guildgate", "Land — Gate")
	checkState(t, g)
	if !hasEffectiveKeyword(t, g, guard, "vigilance") {
		t.Error("no vigilance with a Gate")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(gate) })
	checkState(t, g)
	if hasEffectiveKeyword(t, g, guard, "vigilance") {
		t.Error("vigilance after the Gate left")
	}
}

// --- Makeshift Battalion --------------------------------------------------------------------------------

func TestGapB76MakeshiftBattalionGrowsOnlyWithTwoOtherAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	mb := b12Push(g, me.ID, "Makeshift Battalion", "Creature — Human Soldier", gapMakeshiftOracle, 3, 2)
	advanceToMain(t, g)
	gtcAttackWith(t, g, mb, 1)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, mb, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters with one other attacker, want 0", n)
	}
	_ = opp
}

func TestGapB76MakeshiftBattalionPutsACounterOnItselfWithTwoOtherAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := bnsActive(g)
	mb := b12Push(g, me.ID, "Makeshift Battalion", "Creature — Human Soldier", gapMakeshiftOracle, 3, 2)
	advanceToMain(t, g)
	gtcAttackWith(t, g, mb, 2)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, mb, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
}

// --- Sky Hussar ---------------------------------------------------------------------------------------------

const gapSkyHussarOracle = "21b41dd6-bff7-4e3d-aa28-551a30194225"

// gapForecastTable has the active player in their upkeep with a Sky Hussar in
// hand and the given untapped creatures on the battlefield.
func gapForecastTable(t *testing.T, creatures ...[]string) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var ids []uuid.UUID
	for i, colors := range creatures {
		ids = append(ids, apaPush(g, me.ID, me.ID, game.Card{Name: "Pilot " + string(rune('A'+i)), TypeLine: "Creature — Human",
			Power: 1, Toughness: 1, Colors: colors}))
	}
	hussar := handCardForTest(me, "Sky Hussar", "Creature — Human Knight", gapSkyHussarOracle)
	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	return g, me, hussar, ids
}

func TestGapB76SkyHussarForecastDrawsACardForTwoWhiteOrBlueCreatures(t *testing.T) {
	g, me, hussar, ids := gapForecastTable(t, []string{"W"}, []string{"U"}, []string{"R"})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, hussar, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{ids[0], ids[2]}}); err == nil {
		t.Fatal("a red creature paid for forecast")
	}
	if err := g.ActivateCatalogAbility(me.ID, hussar, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{ids[0]}}); err == nil {
		t.Fatal("one creature paid for two")
	}
	p7Activate(t, g, me, hussar, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{ids[0], ids[1]}})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the card stays, one is drawn)", me.Hand.Size(), hand+1)
	}
	if z, _ := cardWhere(g, me, hussar); z != "hand" {
		t.Errorf("the Sky Hussar is in %s, want the hand (revealing moves nothing)", z)
	}
	if !tappedNow(g, ids[0]) || !tappedNow(g, ids[1]) || tappedNow(g, ids[2]) {
		t.Error("the two creatures that paid are tapped and the third is not")
	}
	// Once each turn.
	if err := g.ActivateCatalogAbility(me.ID, hussar, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{ids[0], ids[1]}}); err == nil {
		t.Error("forecast twice in one turn")
	}
}

func TestGapB76SkyHussarForecastRevealsTheCard(t *testing.T) {
	g, me, hussar, ids := gapForecastTable(t, []string{"W"}, []string{"U"})
	p7Activate(t, g, me, hussar, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{ids[0], ids[1]}})
	revealed := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && slices.Contains([]uuid.UUID{ev.CardID, ev.Source}, hussar) {
			revealed = true
		}
	}
	if !revealed {
		t.Error("no reveal event for the Sky Hussar")
	}
}

func TestGapB76SkyHussarForecastOnlyDuringYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := apaPush(g, me.ID, me.ID, game.Card{Name: "A", TypeLine: "Creature — Human", Power: 1, Toughness: 1, Colors: []string{"W"}})
	b := apaPush(g, me.ID, me.ID, game.Card{Name: "B", TypeLine: "Creature — Human", Power: 1, Toughness: 1, Colors: []string{"U"}})
	hussar := handCardForTest(me, "Sky Hussar", "Creature — Human Knight", gapSkyHussarOracle)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, hussar, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a, b}}); err == nil {
		t.Error("forecast in a main phase")
	}
	// An opponent's upkeep is not yours.
	advanceToUpkeepOfSeat(t, g, 1)
	if err := g.ActivateCatalogAbility(me.ID, hussar, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a, b}}); err == nil {
		t.Error("forecast in an opponent's upkeep")
	}
}

func TestGapB76SkyHussarFliesAndUntapsYourCreaturesWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tapped := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tapped", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Tapped: true})
	advanceToMain(t, g)
	hussar := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: hussar, Name: "Sky Hussar", TypeLine: "Creature — Human Knight", OracleID: gapSkyHussarOracle,
		Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, hussar, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tappedNow(g, tapped) {
		t.Error("the creature is still tapped")
	}
	if !hasEffectiveKeyword(t, g, hussar, "flying") {
		t.Error("no flying")
	}
}
