package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander wave 2, batch 9 (#5368): Vampires — lords, tokens, Vampire Demon.

const (
	c2b09MartyrOfDusk      = "794aaf8a-6d8e-41fb-acbe-a2b73d65d9d9"
	c2b09ElendasHierophant = "cb99c675-adc5-4dad-8be8-524f5f921213"
	c2b09MarchOfCanonized  = "9eb427ab-1483-4c40-8cb2-fd463e1e5fe0"
)

// c2b09Vampires counts the controller's Vampire tokens with lifelink.
func c2b09Vampires(g *game.Game, owner uuid.UUID) int {
	n, _ := countTokensNamed(g, owner, "Vampire")
	return n
}

func TestC2B09MartyrOfDuskLeavesALifelinkVampire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	martyr := b12Push(g, me.ID, "Martyr of Dusk", "Creature — Vampire Soldier", c2b09MartyrOfDusk, 2, 1)
	b15Destroy(t, g, martyr)
	if got := c2b09Vampires(g, me.ID); got != 1 {
		t.Fatalf("Vampire tokens %d, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vampire" {
			if c.Power != 1 || c.Toughness != 1 {
				t.Errorf("token is %d/%d, want 1/1", c.Power, c.Toughness)
			}
			if !p1g2HasKeyword(g, c.InstanceID, "lifelink") {
				t.Errorf("token has no lifelink")
			}
		}
	}
}

func TestC2B09ElendasHierophantGrowsOnLifeGainAndLeavesItsPowerInVampires(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	h := b12Push(g, me.ID, "Elenda's Hierophant", "Creature — Vampire Cleric", c2b09ElendasHierophant, 1, 1)
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(h, me.ID, 3) })
		passPriorityAroundTable(t, g)
	}
	if got := countersOn(g, h, game.CounterPlusOne); got != 2 {
		t.Fatalf("counters %d, want 2 (one per life-gain event)", got)
	}
	b15Destroy(t, g, h)
	if got := c2b09Vampires(g, me.ID); got != 3 {
		t.Errorf("Vampire tokens %d, want 3 (its last-known power)", got)
	}
}

func TestC2B09ElendasHierophantIgnoresOpponentsLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h := b12Push(g, me.ID, "Elenda's Hierophant", "Creature — Vampire Cleric", c2b09ElendasHierophant, 1, 1)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(h, opp.ID, 3) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, h, game.CounterPlusOne); got != 0 {
		t.Errorf("counters %d, want 0", got)
	}
}

func TestC2B09MarchOfTheCanonizedMakesXVampires(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "March of the Canonized", "Enchantment", c2b09MarchOfCanonized, "{X}{W}{W}", 4, nil)
	passPriorityAroundTable(t, g)
	if got := c2b09Vampires(g, me.ID); got != 4 {
		t.Fatalf("Vampire tokens %d, want 4", got)
	}
}

func TestC2B09MarchOfTheCanonizedUpkeepNeedsSevenDevotion(t *testing.T) {
	// The March itself is {X}{W}{W}: two white symbols. The pad adds the rest.
	for _, tc := range []struct {
		pad  string
		want int
	}{{"{W}{B}{B}{B}", 0}, {"{W}{B}{B}{B}{B}", 1}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "March of the Canonized", TypeLine: "Enchantment", OracleID: c2b09MarchOfCanonized,
			ManaCost: "{X}{W}{W}", Owner: me.ID, Controller: me.ID,
		})
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Devotion Pad", TypeLine: "Enchantment",
			ManaCost: tc.pad, Owner: me.ID, Controller: me.ID,
		})
		advanceToUpkeepOfSeat(t, g, 0)
		passPriorityAroundTable(t, g)
		got := 0
		var demon *game.Card
		for i, c := range g.Battlefield.Cards {
			if c.Name == "Vampire Demon" && c.Controller == me.ID {
				got++
				demon = &g.Battlefield.Cards[i]
			}
		}
		if got != tc.want {
			t.Fatalf("pad %s: Vampire Demons %d, want %d", tc.pad, got, tc.want)
		}
		if demon != nil && (demon.Power != 4 || demon.Toughness != 3) {
			t.Errorf("Demon is %d/%d, want 4/3", demon.Power, demon.Toughness)
		}
	}
}

const (
	c2b09Bloodtracker    = "fcf9ced2-867d-4bc5-b7b6-6153ed2cac97"
	c2b09MasterOfDarkRit = "a65250f6-55e2-47ed-a355-d690047db6f2"
)

func TestC2B09BloodtrackerPaysLifeForCountersAndDrawsThemWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bt := b12Push(g, me.ID, "Bloodtracker", "Creature — Vampire Wizard", c2b09Bloodtracker, 2, 2)
	life := me.Life
	for i := 0; i < 3; i++ {
		b06AddMana(me, "B")
		if err := g.ActivateCatalogAbility(me.ID, bt, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}
	if got := countersOn(g, bt, game.CounterPlusOne); got != 3 {
		t.Fatalf("counters %d, want 3", got)
	}
	if me.Life != life-6 {
		t.Errorf("life %d, want %d", me.Life, life-6)
	}
	hand := me.Hand.Size()
	b15Destroy(t, g, bt)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3 (one per counter it had)", got)
	}
}

func TestC2B09BloodtrackerNeedsTheManaAndTheLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bt := b12Push(g, me.ID, "Bloodtracker", "Creature — Vampire Wizard", c2b09Bloodtracker, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, bt, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	b06AddMana(me, "B")
	me.Life = 1
	if err := g.ActivateCatalogAbility(me.ID, bt, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("paid 2 life with 1 life")
	}
}

func TestC2B09MasterOfDarkRitesMakesBBBForVampireClericDemonSpells(t *testing.T) {
	cast := func(typeLine string) error {
		g := newCatalogGame(t)
		me := g.Seats[0]
		master := b12Push(g, me.ID, "Master of Dark Rites", "Creature — Vampire Cleric", c2b09MasterOfDarkRit, 1, 1)
		fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		if err := g.ActivateManaAbility(me.ID, master, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		if me.ManaPool.Total() != 3 {
			t.Fatalf("pool %d, want 3", me.ManaPool.Total())
		}
		if _, ok := battlefieldCard(g, fodder); ok {
			t.Fatal("the sacrificed creature is still on the battlefield")
		}
		if c, _ := battlefieldCard(g, master); !c.Tapped {
			t.Fatal("the Master did not tap")
		}
		id := putInHand(me, game.Card{Name: "Spell", TypeLine: typeLine, ManaCost: "{B}{B}{B}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	for _, tl := range []string{"Creature — Vampire Knight", "Creature — Human Cleric", "Creature — Demon"} {
		if err := cast(tl); err != nil {
			t.Errorf("%s: %v", tl, err)
		}
	}
	for _, tl := range []string{"Creature — Human Warrior", "Sorcery", "Artifact"} {
		if err := cast(tl); err == nil {
			t.Errorf("%s: the restricted mana paid for it", tl)
		}
	}
}

func TestC2B09MasterOfDarkRitesCannotSacrificeItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	master := b12Push(g, me.ID, "Master of Dark Rites", "Creature — Vampire Cleric", c2b09MasterOfDarkRit, 1, 1)
	if err := g.ActivateManaAbility(me.ID, master, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{master}}); err == nil {
		t.Fatal("sacrificed itself for 'another creature'")
	}
}

const (
	c2b09Vona        = "b1c56e23-eae6-47a1-a1cd-3a2d01621768"
	c2b09Wrath       = "9aebd891-c8c5-4309-a07b-6cab1b204172"
	c2b09Choir       = "90c62b27-027a-44a3-8d21-3c62701b18a8"
	c2b09Order       = "0cb41c3f-5e6a-4c63-80b5-c9ab84b4e4bf"
	c2b09Promise     = "4ef4b4f9-47cd-4862-8cbb-b6e624d6e6fc"
	c2b09Clavileno   = "2b62a716-7831-4c76-8f2d-b99b04e33350"
	c2b09VonaAbility = 0
)

func TestC2B09VonaDestroysANonlandPermanentOnYourTurnForSevenLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	vona := b12Push(g, me.ID, "Vona, Butcher of Magan", "Legendary Creature — Vampire Knight", c2b09Vona, 4, 4)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	land := pushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	target := func(id uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
	}
	if err := g.ActivateCatalogAbility(me.ID, vona, c2b09VonaAbility, target(land)); err == nil {
		t.Fatal("targeted a land")
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, vona, c2b09VonaAbility, target(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived")
	}
	if me.Life != life-7 {
		t.Errorf("life %d, want %d", me.Life, life-7)
	}
	if c, _ := battlefieldCard(g, vona); !c.Tapped {
		t.Error("Vona did not tap")
	}
}

func TestC2B09VonaNeedsSevenLifeAndYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	vona := b12Push(g, me.ID, "Vona, Butcher of Magan", "Legendary Creature — Vampire Knight", c2b09Vona, 4, 4)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}
	me.Life = 6
	if err := g.ActivateCatalogAbility(me.ID, vona, c2b09VonaAbility, params); err == nil {
		t.Fatal("paid 7 life from 6")
	}
	me.Life = 40
	for g.Turn.ActiveSeat == 0 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, vona, c2b09VonaAbility, params); err == nil {
		t.Fatal("activated on an opponent's turn")
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("the victim died anyway")
	}
}

func TestC2B09OliviasWrathShrinksNonVampiresByYourVampires(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "Vamp A", "Creature — Vampire", 2, 2)
	b12Creature(g, me.ID, "Vamp B", "Creature — Vampire Knight", 2, 2)
	b12Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 2, 2) // not counted, and spared
	small := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Olivia's Wrath", "Sorcery", c2b09Wrath, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("a 2/2 non-Vampire survived -2/-2")
	}
	if p, tough := ptOf(t, g, big); p != 3 || tough != 3 {
		t.Errorf("Big Bear is %d/%d, want 3/3", p, tough)
	}
	if got := onBattlefieldNamed(g, "Vamp A") + onBattlefieldNamed(g, "Vamp B") + onBattlefieldNamed(g, "Their Vampire"); got != 3 {
		t.Errorf("%d Vampires left, want all 3", got)
	}
}

func TestC2B09OliviasWrathWithNoVampiresDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Olivia's Wrath", "Sorcery", c2b09Wrath, nil)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("Bear power %d, want 2", p)
	}
}

func TestC2B09RedemptionChoirCovenReturnsACheapPermanentOnEnterAndAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "One", "Creature — Bear", 1, 1)
	b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	cheap := pushGraveyardPermanent(me, "Cheap Bear", "Creature — Bear", "{1}{G}")
	dear := pushGraveyardPermanent(me, "Serra Angel", "Creature — Angel", "{3}{W}{W}")
	spell := pushGraveyardPermanent(me, "Shock", "Instant", "{R}")
	choir := c2b09CastChoir(t, g, me)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt: coven should hold with powers 1, 2 and 3")
	}
	if !hasID(prompt.PickTargetCards, cheap) || hasID(prompt.PickTargetCards, dear) || hasID(prompt.PickTargetCards, spell) {
		t.Fatalf("offered %v, want only the mana value 2 permanent card", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the enters trigger did not return the card")
	}
	// The attack half.
	again := pushGraveyardPermanent(me, "Another Bear", "Creature — Bear", "{2}")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	declareAttack(t, g, opp.ID, choir)
	pickCard(t, g, me.ID, again)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(again) {
		t.Error("the attack trigger did not return the card")
	}
}

func TestC2B09RedemptionChoirWithoutThreeDifferentPowersDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Three", "Creature — Bear", 3, 3)
	b12Creature(g, me.ID, "Three Again", "Creature — Bear", 3, 3)
	cheap := pushGraveyardPermanent(me, "Cheap Bear", "Creature — Bear", "{1}{G}")
	c2b09CastChoir(t, g, me)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("coven triggered with only powers 3 and 3")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("a card returned without coven")
	}
}

// c2b09CastChoir casts a 3/3 Redemption Choir and lets it resolve, stopping at
// the first prompt its enters trigger asks.
func c2b09CastChoir(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := avB1Cast(t, g, me, game.Card{Name: "Redemption Choir", TypeLine: "Creature — Vampire Cleric",
		OracleID: c2b09Choir, Power: 3, Toughness: 3}, game.CastSpellParams{})
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	return id
}

func TestC2B09OrderOfSacredDuskConvokesAndGrantsExaltedToOtherVampires(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var tappers []uuid.UUID
	for i := 0; i < 7; i++ {
		tappers = append(tappers, pushTapCostPermanent(g, me.ID, "Soldier", "Creature — Soldier", "W"))
	}
	black := pushTapCostPermanent(g, me.ID, "Cutpurse", "Creature — Human Rogue", "B")
	if _, err := castWithTapParams(t, g, "Order of Sacred Dusk", "Creature — Vampire Knight", "{6}{W}{B}", c2b09Order,
		game.CastSpellParams{Strict: true, TapIDs: tappers}); err == nil {
		t.Fatal("cast {6}{W}{B} by convoking only white creatures")
	}
	for _, id := range tappers {
		if tapCostTapped(g, id) {
			t.Fatal("a refused cast tapped a creature")
		}
	}
	id, err := castWithTapParams(t, g, "Order of Sacred Dusk", "Creature — Vampire Knight", "{6}{W}{B}", c2b09Order,
		game.CastSpellParams{Strict: true, TapIDs: append(tappers, black)})
	if err != nil {
		t.Fatalf("convoke cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Order never resolved")
	}
	for _, kw := range []string{"flying", "lifelink", "haste", "exalted"} {
		if !p1g2HasKeyword(g, id, kw) {
			t.Errorf("the Order has no %s", kw)
		}
	}
	// Another Vampire gets exalted; an opponent's, and a non-Vampire, do not.
	mine := b12Creature(g, me.ID, "My Vampire", "Creature — Vampire", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 2, 2)
	human := b12Creature(g, me.ID, "My Human", "Creature — Human", 2, 2)
	if !p1g2HasKeyword(g, mine, "exalted") {
		t.Error("another Vampire you control lacks exalted")
	}
	if p1g2HasKeyword(g, theirs, "exalted") {
		t.Error("an opponent's Vampire has exalted")
	}
	if p1g2HasKeyword(g, human, "exalted") {
		t.Error("a non-Vampire has exalted")
	}
}

func TestC2B09OrderOfSacredDuskAddsAnExaltedInstancePerOtherVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Order of Sacred Dusk",
		TypeLine: "Creature — Vampire Knight", OracleID: c2b09Order, Power: 5, Toughness: 5,
		Keywords: []string{"flying", "lifelink", "haste", "exalted"}, Owner: me.ID, Controller: me.ID})
	human := b12Creature(g, me.ID, "My Human", "Creature — Human", 2, 2)
	// The Order's own exalted: one instance.
	declareAttack(t, g, opp.ID, human)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, human); p != 3 {
		t.Fatalf("with the Order alone the lone attacker has power %d, want 3", p)
	}
}

func TestC2B09OrderOfSacredDuskWithAnotherVampireMakesTwoExaltedTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Order of Sacred Dusk",
		TypeLine: "Creature — Vampire Knight", OracleID: c2b09Order, Power: 5, Toughness: 5,
		Keywords: []string{"flying", "lifelink", "haste", "exalted"}, Owner: me.ID, Controller: me.ID})
	b12Creature(g, me.ID, "Other Vampire", "Creature — Vampire", 1, 1)
	human := b12Creature(g, me.ID, "My Human", "Creature — Human", 2, 2)
	declareAttack(t, g, opp.ID, human)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, human); p != 4 {
		t.Fatalf("the lone attacker has power %d, want 4 (two exalted instances)", p)
	}
}

func TestC2B09PromiseOfAclazotzSacrificesANonDemonToPopulate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Promise of Aclazotz", "Enchantment", c2b09Promise, 0, 0)
	token := b12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	_ = token
	demon := b12Creature(g, me.ID, "A Demon", "Creature — Demon", 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil {
		t.Fatal("no sacrifice prompt")
	}
	for _, c := range prompt.ChooseCards {
		if c == demon {
			t.Error("a Demon was offered")
		}
	}
	answerOwnPermanents(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 2 {
		t.Errorf("Spirit tokens %d, want 2 (populate)", n)
	}
}

func TestC2B09PromiseOfAclazotzMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Promise of Aclazotz", "Enchantment", c2b09Promise, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was sacrificed though declined")
	}
}

func TestC2B09FoulRebirthSacrificesANonDemonForAVampireDemon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	demon := b12Creature(g, me.ID, "A Demon", "Creature — Demon", 4, 4)
	castCatalogSpell(t, g, "Foul Rebirth", "Sorcery — Adventure", c2b09Promise+"#1", nil)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if !g.Battlefield.Contains(demon) {
		t.Error("the Demon was sacrificed")
	}
	if n, _ := countTokensNamed(g, me.ID, "Vampire Demon"); n != 1 {
		t.Errorf("Vampire Demon tokens %d, want 1", n)
	}
}

func TestC2B09FoulRebirthWithOnlyDemonsMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "A Demon", "Creature — Demon", 4, 4)
	castCatalogSpell(t, g, "Foul Rebirth", "Sorcery — Adventure", c2b09Promise+"#1", nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Vampire Demon"); n != 0 {
		t.Errorf("Vampire Demon tokens %d, want 0", n)
	}
}

func TestC2B09ClavilenoMakesAnAttackingVampireADemonThatLeavesOneBehind(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Clavileño, First of the Blessed", "Legendary Creature — Vampire Cleric", c2b09Clavileno, 2, 2)
	vamp := b12Creature(g, me.ID, "My Vampire", "Creature — Vampire", 2, 2)
	declareAttack(t, g, opp.ID, vamp)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, vamp)
	}
	passPriorityAroundTable(t, g)
	card := avB1Card(t, g, vamp)
	if !card.HasSubtype("Demon") || !card.HasSubtype("Vampire") {
		t.Fatalf("types %q, want Vampire and Demon", card.TypeLine)
	}
	hand := me.Hand.Size()
	b15Destroy(t, g, vamp)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	n, tapped := countTokensNamed(g, me.ID, "Vampire Demon")
	if n != 1 || tapped != 1 {
		t.Errorf("Vampire Demon tokens %d (%d tapped), want one tapped", n, tapped)
	}
}

func TestC2B09ClavilenoIgnoresDemonsAndNonVampires(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Clavileño, First of the Blessed", "Legendary Creature — Vampire Cleric", c2b09Clavileno, 2, 2)
	demon := b12Creature(g, me.ID, "Vampire Demon", "Creature — Vampire Demon", 4, 3)
	human := b12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	declareAttack(t, g, opp.ID, demon, human)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("offered a target though no attacking Vampire is a non-Demon")
	}
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	b15Destroy(t, g, demon)
	if me.Hand.Size() != hand {
		t.Error("a Demon that was never granted the trigger drew a card")
	}
}
