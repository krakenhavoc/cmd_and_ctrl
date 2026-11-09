package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdrGenericCost is the generic part of a spell's total cost as the
// cast would price it for `me`.
func cmdrGenericCost(t *testing.T, g *game.Game, me *game.Player, name, typeLine, manaCost string) int {
	t.Helper()
	id := handCardFull(me, name, typeLine, manaCost, "", []string{"W"})
	var card game.Card
	for _, c := range me.Hand.Cards {
		if c.InstanceID == id {
			card = c
		}
	}
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast %s: %v", name, err)
	}
	return price.Total.Generic
}

// --- Starnheim Aspirant ----------------------------------------------------

func TestCmdrStarnheimAspirantDiscountsYourAngelSpellsByTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	if got := cmdrGenericCost(t, g, me, "Big Angel", "Creature — Angel", "{4}{W}{W}"); got != 4 {
		t.Fatalf("without the Aspirant: %d generic, want 4", got)
	}
	cmdrPush(g, me.ID, "Starnheim Aspirant", "Creature — Human Cleric", cmdrStarnheimAspirantOracle, 2, 2)
	if got := cmdrGenericCost(t, g, me, "Big Angel", "Creature — Angel", "{4}{W}{W}"); got != 2 {
		t.Errorf("Angel spell: %d generic, want 2", got)
	}
	if got := cmdrGenericCost(t, g, me, "Pure Angel", "Creature — Angel", "{W}{W}"); got != 0 {
		t.Errorf("coloured-only Angel: %d generic, want 0 (it cannot reduce coloured mana)", got)
	}
	if got := cmdrGenericCost(t, g, me, "Bear", "Creature — Bear", "{4}{W}"); got != 4 {
		t.Errorf("non-Angel: %d generic, want 4", got)
	}
	if got := cmdrGenericCost(t, g, me, "Angel Maker", "Sorcery", "{4}{W}"); got != 4 {
		t.Errorf("a sorcery that makes Angels: %d generic, want 4", got)
	}
	// An opponent's Aspirant discounts only its controller's spells.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	toMainForCost(t, g2)
	cmdrPush(g2, opp.ID, "Starnheim Aspirant", "Creature — Human Cleric", cmdrStarnheimAspirantOracle, 2, 2)
	if got := cmdrGenericCost(t, g2, me2, "Big Angel", "Creature — Angel", "{4}{W}{W}"); got != 4 {
		t.Errorf("an opponent's Aspirant discounted my Angel: %d generic", got)
	}
}

// --- Herald of War ----------------------------------------------------------

func TestCmdrHeraldOfWarGrowsOnAttackAndDiscountsAngelsAndHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	herald := cmdrPush(g, me.ID, "Herald of War", "Creature — Angel", cmdrHeraldOfWarOracle, 3, 3)
	toMainForCost(t, g)
	if got := cmdrGenericCost(t, g, me, "Big Angel", "Creature — Angel", "{4}{W}{W}"); got != 4 {
		t.Fatalf("no counters: %d generic, want 4", got)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(herald, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, herald, game.CounterPlusOne); got != 1 {
		t.Fatalf("attacked: %d counters, want 1", got)
	}
	// More counters, from anywhere, count too.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(herald, game.CounterPlusOne, 1) })
	if got := cmdrGenericCost(t, g, me, "Big Angel", "Creature — Angel", "{4}{W}{W}"); got != 2 {
		t.Errorf("two counters, Angel: %d generic, want 2", got)
	}
	if got := cmdrGenericCost(t, g, me, "Soldier", "Creature — Human Soldier", "{3}{W}"); got != 1 {
		t.Errorf("two counters, Human: %d generic, want 1", got)
	}
	if got := cmdrGenericCost(t, g, me, "Angel Human", "Creature — Angel Human", "{3}{W}"); got != 1 {
		t.Errorf("a spell that is both: %d generic, want 1 (discounted once per counter, not twice)", got)
	}
	if got := cmdrGenericCost(t, g, me, "Bear", "Creature — Bear", "{3}{W}"); got != 3 {
		t.Errorf("a Bear: %d generic, want 3", got)
	}
	if got := cmdrGenericCost(t, g, me, "Cheap Angel", "Creature — Angel", "{1}{W}"); got != 0 {
		t.Errorf("a cheap Angel: %d generic, want 0 (stops at zero)", got)
	}
}

func TestCmdrHeraldOfWarDoesNotDiscountAnOpponentsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herald := cmdrPush(g, opp.ID, "Herald of War", "Creature — Angel", cmdrHeraldOfWarOracle, 3, 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(herald, game.CounterPlusOne, 2) })
	toMainForCost(t, g)
	if got := cmdrGenericCost(t, g, me, "Big Angel", "Creature — Angel", "{4}{W}{W}"); got != 4 {
		t.Errorf("an opponent's Herald discounted my Angel: %d generic", got)
	}
}

// --- Serra Avenger ------------------------------------------------------------

func TestCmdrSerraAvengerCannotBeCastInYourFirstThreeTurns(t *testing.T) {
	for turns, wantOK := range map[int]bool{1: false, 2: false, 3: false, 4: true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		me.TurnsBegun = turns
		toMainForCost(t, g)
		id := handCardFull(me, "Serra Avenger", "Creature — Angel", "{W}{W}", cmdrSerraAvengerOracle, []string{"W"})
		floatForTest(g, me, "WW")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if wantOK && err != nil {
			t.Errorf("turn %d: cast refused: %v", turns, err)
		}
		if !wantOK && err == nil {
			t.Errorf("turn %d: cast allowed", turns)
		}
	}
}

func TestCmdrSerraAvengerCanBeCastOnSomeoneElsesEarlyTurn(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	other.TurnsBegun = 0
	active.TurnsBegun = 1
	if cmdrB1IsYourFirstThreeTurns(g, other.ID) {
		t.Error("it is not the other player's turn")
	}
	if !cmdrB1IsYourFirstThreeTurns(g, active.ID) {
		t.Error("it is the active player's first turn")
	}
}

// --- Seraph of the Sword --------------------------------------------------------

func TestCmdrSeraphOfTheSwordPreventsCombatDamageToItselfOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	attacker := pushVanillaCreature(g, me.ID, "Big Attacker", 5, 5)
	seraph := cmdrPush(g, opp.ID, "Seraph of the Sword", "Creature — Angel", cmdrSeraphOfTheSwordOracle, 3, 3)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(seraph, attacker); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, seraph)
	if !ok {
		t.Fatal("the Seraph died to combat damage")
	}
	if c.DamageMarked != 0 {
		t.Errorf("the Seraph has %d damage marked", c.DamageMarked)
	}
	// Other damage is not prevented.
	b27Damage(g, attacker, seraph, 3)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, seraph); ok {
		t.Error("3 noncombat damage did not kill the 3/3 Seraph")
	}
}

// --- Sunblast Angel -----------------------------------------------------------

func TestCmdrSunblastAngelDestroysEveryTappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tappedMine := cmdrPush(g, me.ID, "Tapped Mine", "Creature — Bear", "", 2, 2)
	tappedTheirs := cmdrPush(g, opp.ID, "Tapped Theirs", "Creature — Bear", "", 2, 2)
	untapped := cmdrPush(g, opp.ID, "Untapped", "Creature — Bear", "", 2, 2)
	tappedRock := cmdrPush(g, opp.ID, "Tapped Rock", "Artifact", "", 0, 0)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{tappedMine, tappedTheirs, tappedRock} {
			_ = g.TapTargetForEffect(id)
		}
	})
	angel := castAndResolveCreature(t, g, "Sunblast Angel", "Creature — Angel", cmdrSunblastAngelOracle)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{tappedMine, tappedTheirs} {
		if _, ok := battlefieldCard(g, id); ok {
			t.Error("a tapped creature survived")
		}
	}
	for name, id := range map[string]uuid.UUID{"untapped creature": untapped, "tapped artifact": tappedRock, "the Angel": angel} {
		if _, ok := battlefieldCard(g, id); !ok {
			t.Errorf("the %s was destroyed", name)
		}
	}
}

// --- Defy Death ------------------------------------------------------------------

func TestCmdrDefyDeathReturnsACreatureAndCountersAnAngel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	angel := me.Graveyard
	angelID := uuid.New()
	angel.PushTop(game.Card{InstanceID: angelID, Name: "Dead Angel", TypeLine: "Creature — Angel", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Defy Death", "Sorcery", cmdrDefyDeathOracle, []game.TargetRef{{Kind: game.TargetCard, ID: angelID}})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, angelID, game.CounterPlusOne); got != 2 {
		t.Errorf("an Angel came back with %d counters, want 2", got)
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	bearID := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bearID, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Defy Death", "Sorcery", cmdrDefyDeathOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bearID}})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bearID, game.CounterPlusOne); got != 0 {
		t.Errorf("a Bear came back with %d counters, want 0", got)
	}
}

func TestCmdrDefyDeathRefusesAnOpponentsGraveyardAndNonCreatures(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Angel", TypeLine: "Creature — Angel", Owner: opp.ID, Controller: opp.ID})
	if err := castCatalogSpellErr(t, g, "Defy Death", "Sorcery", cmdrDefyDeathOracle, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("Defy Death targeted a card in an opponent's graveyard")
	}
	g = newCatalogGame(t)
	me := g.Seats[0]
	rock := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if err := castCatalogSpellErr(t, g, "Defy Death", "Sorcery", cmdrDefyDeathOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Error("Defy Death targeted an artifact card")
	}
}

// --- Invoke the Divine --------------------------------------------------------------

func TestCmdrInvokeTheDivineDestroysAnArtifactOrEnchantmentAndGainsFour(t *testing.T) {
	for _, typeLine := range []string{"Artifact", "Enchantment"} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		target := pushTypedCard(g, opp.ID, "Target", typeLine, "{2}")
		life := me.Life
		castCatalogSpell(t, g, "Invoke the Divine", "Instant", cmdrInvokeTheDivineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
		passPriorityAroundTable(t, g)
		if _, ok := battlefieldCard(g, target); ok {
			t.Errorf("the %s survived", typeLine)
		}
		if got := me.Life - life; got != 4 {
			t.Errorf("gained %d life, want 4", got)
		}
	}
}

func TestCmdrInvokeTheDivineRefusesACreatureAndGainsNothingWhenItFizzles(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}")
	if err := castCatalogSpellErr(t, g, "Invoke the Divine", "Instant", cmdrInvokeTheDivineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("Invoke the Divine targeted a creature")
	}

	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushTypedCard(g, opp.ID, "Rock", "Artifact", "{2}")
	life := me.Life
	castCatalogSpell(t, g, "Invoke the Divine", "Instant", cmdrInvokeTheDivineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	b27Kill(g, rock) // in response
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a fizzled Invoke the Divine gained %d life", me.Life-life)
	}
}
