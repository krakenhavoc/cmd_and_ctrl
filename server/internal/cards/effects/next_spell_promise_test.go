package effects

import (
	"testing"

	"github.com/google/uuid"

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

func nsPromises(p *game.Player) []game.NextSpellPromise {
	var out []game.NextSpellPromise
	for _, s := range p.Statics {
		if s.NextSpell.Active {
			out = append(out, s.NextSpell)
		}
	}
	return out
}

// nsHandCard seeds a card into a seat's hand for a direct CastSpell.
func nsHandCard(p *game.Player, name, typeLine, cost string) game.Card {
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID, Power: 2, Toughness: 2}
	p.Hand.PushTop(c)
	return c
}

// nsPrice prices `card` for `p` through the one CR 601.2f pass every
// cast uses.
func nsPrice(t *testing.T, g *game.Game, p *game.Player, card game.Card) game.ParsedCost {
	t.Helper()
	base, err := game.ParseCost(card.ManaCost)
	if err != nil {
		t.Fatal(err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: card, Controller: p.ID, FromZone: game.ZoneHand})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestSavageSummoningGivesFlashUncounterableAndACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Savage Summoning", "Instant", nsSavageSummoningOracle, nil)
	passPriorityAroundTable(t, g)
	got := nsPromises(me)
	if len(got) != 1 || !got[0].Flash || !got[0].CantBeCountered || got[0].Counters != 1 || !got[0].Filter.CreatureOnly {
		t.Fatalf("the promise: %+v", got)
	}

	// An instant leaves it; the first creature spell spends it, in a
	// step where a creature spell could not otherwise be cast.
	advanceTo(t, g, game.StepEnd)
	nonCreature := nsHandCard(me, "Test Instant", "Instant", "")
	if err := g.CastSpell(me.ID, nonCreature.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if len(nsPromises(me)) != 1 {
		t.Fatal("a noncreature spell does not spend the promise")
	}
	bear := nsHandCard(me, "Flash Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("a creature spell is castable at instant speed under the promise: %v", err)
	}
	if len(nsPromises(me)) != 0 {
		t.Fatal("the creature spell spent the promise")
	}
	var uncounterable bool
	g.WithWriteLock(func() { uncounterable = g.SpellCantBeCounteredForEffect(bear.InstanceID) })
	if !uncounterable {
		t.Error("that spell can't be countered")
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear.InstanceID, "+1/+1"); n != 1 {
		t.Errorf("the creature entered with %d +1/+1 counters, want 1", n)
	}
	later := nsHandCard(me, "Later Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, later.InstanceID, game.CastSpellParams{}); err == nil {
		t.Error("the flash was one creature spell's")
	}
}

func TestSavageSummoningCantBeCounteredItself(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Savage Summoning", "Instant", nsSavageSummoningOracle, nil)
	var uncounterable bool
	g.WithWriteLock(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
	if !uncounterable {
		t.Error("This spell can't be countered.")
	}
}

func TestQuickenGivesTheNextSorceryFlashAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Quicken", "Instant", nsQuickenOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Quicken draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	if got := nsPromises(me); len(got) != 1 || !got[0].Filter.SorceryOnly || !got[0].Flash {
		t.Fatalf("the promise: %+v", got)
	}
	advanceTo(t, g, game.StepEnd)
	bear := nsHandCard(me, "Late Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err == nil {
		t.Fatal("a creature spell gets no flash from a sorcery promise")
	}
	sorcery := nsHandCard(me, "Test Sorcery", "Sorcery", "")
	if err := g.CastSpell(me.ID, sorcery.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("the next sorcery is castable at instant speed: %v", err)
	}
	if len(nsPromises(me)) != 0 {
		t.Error("and it spends the promise")
	}
}

func TestHardenedBerserkerDiscountsTheNextSpellAfterAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	berserker := b12Push(g, me.ID, "Hardened Berserker", "Creature — Human Berserker", nsHardenedBerserkerOracle, 3, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(berserker, opp.ID); err != nil {
		t.Fatal(err)
	}
	// The attack declaration is locked in as the step ends, and the
	// trigger goes on the stack.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if len(nsPromises(me)) != 1 {
		t.Fatalf("the trigger made no promise: %+v", me.Statics)
	}
	spell := game.Card{Name: "Test Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID}
	if got := nsPrice(t, g, me, spell); got.Generic != 1 {
		t.Errorf("the next spell costs {1} less: %+v", got)
	}
	if got := nsPrice(t, g, opp, spell); got.Generic != 2 {
		t.Errorf("an opponent's spell is not discounted: %+v", got)
	}
}

func TestSaheeliTheGiftedAffinityPricesTheNextSpellByYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saheeli := pushWalkerForTest(g, me.ID, "Saheeli, the Gifted", nsSaheeliGiftedOracle, 4)
	pushPermanentForTest(g, me.ID, "Sol Ring", "", "Artifact")
	pushPermanentForTest(g, me.ID, "Mind Stone", "", "Artifact")
	pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, saheeli, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	got := nsPromises(me)
	if len(got) != 1 || got[0].AffinityFor != "Artifact" {
		t.Fatalf("the promise: %+v", got)
	}
	// Two artifacts (the Bear is not one): {4}{R} costs {2}{R}, of any kind.
	for _, tc := range []struct{ typeLine, cost string }{{"Sorcery", "{4}{R}"}, {"Creature — Bear", "{4}{R}"}} {
		spell := game.Card{Name: "Test Spell", TypeLine: tc.typeLine, ManaCost: tc.cost, Owner: me.ID}
		if p := nsPrice(t, g, me, spell); p.Generic != 2 {
			t.Errorf("%s: affinity for two artifacts: %+v", tc.typeLine, p)
		}
	}
	// An artifact that arrives before the cast counts too ("as you cast it").
	pushPermanentForTest(g, me.ID, "Signet", "", "Artifact")
	if p := nsPrice(t, g, me, game.Card{Name: "Test Spell", TypeLine: "Sorcery", ManaCost: "{4}{R}", Owner: me.ID}); p.Generic != 1 {
		t.Errorf("three artifacts: %+v", p)
	}
	// An opponent's artifacts and an opponent's casts are not yours.
	if p := nsPrice(t, g, g.Seats[(seat+1)%len(g.Seats)], game.Card{Name: "Theirs", TypeLine: "Sorcery", ManaCost: "{4}{R}"}); p.Generic != 4 {
		t.Errorf("an opponent's spell is not discounted: %+v", p)
	}
}

func TestSaheeliTheGiftedPlusOneMakesAServo(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saheeli := pushWalkerForTest(g, me.ID, "Saheeli, the Gifted", nsSaheeliGiftedOracle, 4)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, saheeli, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saheeli := pushWalkerForTest(g, me.ID, "Saheeli, the Gifted", nsSaheeliGiftedOracle, 7)
	pushPermanentForTest(g, me.ID, "Sol Ring", "", "Artifact")
	pushPermanentForTest(g, me.ID, "Mind Stone", "", "Artifact")
	pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	advanceToMainOf(t, g, seat)
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
	passPriorityAroundTable(t, g)
	if count("Sol Ring") != 2 || count("Mind Stone") != 2 || count("Bear") != 1 {
		t.Fatalf("one copy of each artifact and none of the Bear: sol %d, stone %d, bear %d", count("Sol Ring"), count("Mind Stone"), count("Bear"))
	}
	for _, c := range g.Battlefield.Cards {
		if IsToken(c) && c.Name != "" && !game.HasKeyword(&c, "haste") {
			t.Errorf("token %s lacks haste", c.Name)
		}
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if count("Sol Ring") != 1 || count("Mind Stone") != 1 {
		t.Errorf("the copies are exiled at the next end step: sol %d, stone %d", count("Sol Ring"), count("Mind Stone"))
	}
}

func TestKazaDiscountsTheNextInstantOrSorceryByHerWizards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kaza := pushCatalogPermanent(g, me.ID, "Kaza, Roil Chaser", "Legendary Creature — Human Wizard", nsKazaOracle, false)
	pushCatalogPermanent(g, me.ID, "Other Wizard", "Creature — Human Wizard", "", false)
	pushCatalogPermanent(g, me.ID, "Not A Wizard", "Creature — Bear", "", false)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, kaza, 0, game.ActivateAbilityParams{})
	got := nsPromises(me)
	if len(got) != 1 || got[0].Reduce != 2 || !got[0].Filter.InstantOrSorceryOnly {
		t.Fatalf("X is the Wizards counted on resolution: %+v", got)
	}
	// A Wizard that arrives afterwards changes nothing.
	pushCatalogPermanent(g, me.ID, "Late Wizard", "Creature — Human Wizard", "", false)
	bolt := game.Card{Name: "Test Bolt", TypeLine: "Instant", ManaCost: "{3}{R}", Owner: me.ID}
	if got := nsPrice(t, g, me, bolt); got.Generic != 1 {
		t.Errorf("{3}{R} less {2}: %+v", got)
	}
	bear := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Owner: me.ID}
	if got := nsPrice(t, g, me, bear); got.Generic != 3 {
		t.Errorf("a creature spell is not discounted: %+v", got)
	}
}
