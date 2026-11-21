package testsms

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// next_spell_promise_test.go — #1852: the five cards that promise
// something about "the next spell you cast", through the real catalog.
// The engine contract (consumption, expiry, a countered spell, restore
// points) is pinned in game/next_spell_promise_test.go and the
// enumerator's agreement in legal/next_spell_promise_test.go.

const (
	nsSavageSummoningOracle   = "30c09562-0bca-4382-ae84-aca00e0eb89b"
	nsQuickenOracle           = "cf188dd3-1927-481b-b8c0-93b1222dbf53"
	nsHardenedBerserkerOracle = "2894efc6-ee78-4353-b98a-8005a2451139"
	nsSaheeliGiftedOracle     = "0e084342-31ff-47a5-9fd8-5b385add7a1a"
	nsKazaOracle              = "0b8517c0-0441-4fd9-99ec-2274e57c3cb6"
)

func TestSavageSummoningGivesFlashUncounterableAndACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Savage Summoning", "Instant", nsSavageSummoningOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.NsPromises(me)
	if len(got) != 1 || !got[0].Flash || !got[0].CantBeCountered || got[0].Counters != 1 || !got[0].Filter.CreatureOnly {
		t.Fatalf("the promise: %+v", got)
	}

	// An instant leaves it; the first creature spell spends it, in a
	// step where a creature spell could not otherwise be cast.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	nonCreature := effectstest.NsHandCard(me, "Test Instant", "Instant", "")
	if err := g.CastSpell(me.ID, nonCreature.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.NsPromises(me)) != 1 {
		t.Fatal("a noncreature spell does not spend the promise")
	}
	bear := effectstest.NsHandCard(me, "Flash Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("a creature spell is castable at instant speed under the promise: %v", err)
	}
	if len(effectstest.NsPromises(me)) != 0 {
		t.Fatal("the creature spell spent the promise")
	}
	var uncounterable bool
	g.WithWriteLock(func() { uncounterable = g.SpellCantBeCounteredForEffect(bear.InstanceID) })
	if !uncounterable {
		t.Error("that spell can't be countered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear.InstanceID, "+1/+1"); n != 1 {
		t.Errorf("the creature entered with %d +1/+1 counters, want 1", n)
	}
	later := effectstest.NsHandCard(me, "Later Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, later.InstanceID, game.CastSpellParams{}); err == nil {
		t.Error("the flash was one creature spell's")
	}
}

func TestSavageSummoningCantBeCounteredItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastCatalogSpell(t, g, "Savage Summoning", "Instant", nsSavageSummoningOracle, nil)
	var uncounterable bool
	g.WithWriteLock(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
	if !uncounterable {
		t.Error("This spell can't be countered.")
	}
}

func TestQuickenGivesTheNextSorceryFlashAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Quicken", "Instant", nsQuickenOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Quicken draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	if got := effectstest.NsPromises(me); len(got) != 1 || !got[0].Filter.SorceryOnly || !got[0].Flash {
		t.Fatalf("the promise: %+v", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	bear := effectstest.NsHandCard(me, "Late Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err == nil {
		t.Fatal("a creature spell gets no flash from a sorcery promise")
	}
	sorcery := effectstest.NsHandCard(me, "Test Sorcery", "Sorcery", "")
	if err := g.CastSpell(me.ID, sorcery.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("the next sorcery is castable at instant speed: %v", err)
	}
	if len(effectstest.NsPromises(me)) != 0 {
		t.Error("and it spends the promise")
	}
}

func TestHardenedBerserkerDiscountsTheNextSpellAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	berserker := effectstest.B12Push(g, me.ID, "Hardened Berserker", "Creature — Human Berserker", nsHardenedBerserkerOracle, 3, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(berserker, opp.ID); err != nil {
		t.Fatal(err)
	}
	// The attack declaration is locked in as the step ends, and the
	// trigger goes on the stack.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.NsPromises(me)) != 1 {
		t.Fatalf("the trigger made no promise: %+v", me.Statics)
	}
	spell := game.Card{Name: "Test Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID}
	if got := effectstest.NsPrice(t, g, me, spell); got.Generic != 1 {
		t.Errorf("the next spell costs {1} less: %+v", got)
	}
	if got := effectstest.NsPrice(t, g, opp, spell); got.Generic != 2 {
		t.Errorf("an opponent's spell is not discounted: %+v", got)
	}
}

func TestSaheeliTheGiftedAffinityPricesTheNextSpellByYourArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saheeli := effectstest.PushWalkerForTest(g, me.ID, "Saheeli, the Gifted", nsSaheeliGiftedOracle, 4)
	effectstest.PushPermanentForTest(g, me.ID, "Sol Ring", "", "Artifact")
	effectstest.PushPermanentForTest(g, me.ID, "Mind Stone", "", "Artifact")
	effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, saheeli, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.NsPromises(me)
	if len(got) != 1 || got[0].AffinityFor != "Artifact" {
		t.Fatalf("the promise: %+v", got)
	}
	// Two artifacts (the Bear is not one): {4}{R} costs {2}{R}, of any kind.
	for _, tc := range []struct{ typeLine, cost string }{{"Sorcery", "{4}{R}"}, {"Creature — Bear", "{4}{R}"}} {
		spell := game.Card{Name: "Test Spell", TypeLine: tc.typeLine, ManaCost: tc.cost, Owner: me.ID}
		if p := effectstest.NsPrice(t, g, me, spell); p.Generic != 2 {
			t.Errorf("%s: affinity for two artifacts: %+v", tc.typeLine, p)
		}
	}
	// An artifact that arrives before the cast counts too ("as you cast it").
	effectstest.PushPermanentForTest(g, me.ID, "Signet", "", "Artifact")
	if p := effectstest.NsPrice(t, g, me, game.Card{Name: "Test Spell", TypeLine: "Sorcery", ManaCost: "{4}{R}", Owner: me.ID}); p.Generic != 1 {
		t.Errorf("three artifacts: %+v", p)
	}
	// An opponent's artifacts and an opponent's casts are not yours.
	if p := effectstest.NsPrice(t, g, g.Seats[(seat+1)%len(g.Seats)], game.Card{Name: "Theirs", TypeLine: "Sorcery", ManaCost: "{4}{R}"}); p.Generic != 4 {
		t.Errorf("an opponent's spell is not discounted: %+v", p)
	}
}

func TestSaheeliTheGiftedPlusOneMakesAServo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saheeli := effectstest.PushWalkerForTest(g, me.ID, "Saheeli, the Gifted", nsSaheeliGiftedOracle, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, saheeli, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Servo" && c.Controller == me.ID && c.IsArtifact() && c.IsCreature() {
			found = true
		}
	}
	if !found {
		t.Error("no Servo token")
	}
}

func TestSaheeliTheGiftedUltimateCopiesEachArtifactWithHasteAndExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saheeli := effectstest.PushWalkerForTest(g, me.ID, "Saheeli, the Gifted", nsSaheeliGiftedOracle, 7)
	effectstest.PushPermanentForTest(g, me.ID, "Sol Ring", "", "Artifact")
	effectstest.PushPermanentForTest(g, me.ID, "Mind Stone", "", "Artifact")
	effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	effectstest.AdvanceToMainOf(t, g, seat)
	count := func(name string) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == me.ID {
				n++
			}
		}
		return n
	}
	if err := g.ActivateCatalogAbility(me.ID, saheeli, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if count("Sol Ring") != 2 || count("Mind Stone") != 2 || count("Bear") != 1 {
		t.Fatalf("one copy of each artifact and none of the Bear: sol %d, stone %d, bear %d", count("Sol Ring"), count("Mind Stone"), count("Bear"))
	}
	for _, c := range g.Battlefield.Cards {
		if effects.IsToken(c) && c.Name != "" && !game.HasKeyword(&c, "haste") {
			t.Errorf("token %s lacks haste", c.Name)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if count("Sol Ring") != 1 || count("Mind Stone") != 1 {
		t.Errorf("the copies are exiled at the next end step: sol %d, stone %d", count("Sol Ring"), count("Mind Stone"))
	}
}

func TestKazaDiscountsTheNextInstantOrSorceryByHerWizards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kaza := effectstest.PushCatalogPermanent(g, me.ID, "Kaza, Roil Chaser", "Legendary Creature — Human Wizard", nsKazaOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Other Wizard", "Creature — Human Wizard", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Not A Wizard", "Creature — Bear", "", false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, kaza, 0, game.ActivateAbilityParams{})
	got := effectstest.NsPromises(me)
	if len(got) != 1 || got[0].Reduce != 2 || !got[0].Filter.InstantOrSorceryOnly {
		t.Fatalf("X is the Wizards counted on resolution: %+v", got)
	}
	// A Wizard that arrives afterwards changes nothing.
	effectstest.PushCatalogPermanent(g, me.ID, "Late Wizard", "Creature — Human Wizard", "", false)
	bolt := game.Card{Name: "Test Bolt", TypeLine: "Instant", ManaCost: "{3}{R}", Owner: me.ID}
	if got := effectstest.NsPrice(t, g, me, bolt); got.Generic != 1 {
		t.Errorf("{3}{R} less {2}: %+v", got)
	}
	bear := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Owner: me.ID}
	if got := effectstest.NsPrice(t, g, me, bear); got.Generic != 3 {
		t.Errorf("a creature spell is not discounted: %+v", got)
	}
}
