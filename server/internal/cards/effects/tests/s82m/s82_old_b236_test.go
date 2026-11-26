package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b236_test.go — Hour of Devastation, batch 236 (#5534): each test plays
// the real catalog entry through the real cast, activation and trigger paths.

const (
	oldB236DoomOracle     = "6ded956d-30ed-49ca-8653-cbaaae676b0c"
	oldB236HunterOracle   = "8c7436fa-4802-405b-ba45-4e14f14a3d8e"
	oldB236TormentOracle  = "9e9a7f0b-ffba-43d6-892a-95fd0b8b39f8"
	oldB236MagmarothOID   = "ac5c2069-0706-4b85-a811-a522db016e3e"
	oldB236AmmitOracle    = "c2f6ddc2-86bc-43da-9bdd-0c2af79c62fd"
	oldB236BanewhipOracle = "d46c397c-e772-4a29-bbd2-1d7b47a461e8"
	oldB236SpiderOracle   = "f7289e62-ce5e-4f11-8d64-9fb62531b578"
)

func TestOldB236CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB236DoomOracle:     "Imminent Doom",
		oldB236HunterOracle:   "Tenacious Hunter",
		oldB236TormentOracle:  "Torment of Venom",
		oldB236MagmarothOID:   "Magmaroth",
		oldB236AmmitOracle:    "Ammit Eternal",
		oldB236BanewhipOracle: "Banewhip Punisher",
		oldB236SpiderOracle:   "Obelisk Spider",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// oldB236PutCounter puts one -1/-1 counter on `target` as an effect controlled
// by `placer`.
func oldB236PutCounter(g *game.Game, target, placer uuid.UUID) {
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{ID: uuid.New(), SourceCardID: target, Controller: placer})
		_ = effects.AddCounter{Target: target, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
	})
}

func oldB236Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB236Counters(g *game.Game, id uuid.UUID, kind string) int {
	return effectstest.CountersOn(g, id, kind)
}

// --- Imminent Doom ---

func TestOldB236ImminentDoomEntersWithADoomCounterAndFiresOnMatchingSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doom := effectstest.CastCatalogSpell(t, g, "Imminent Doom", "Enchantment", oldB236DoomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, doom, "doom"); got != 1 {
		t.Fatalf("doom counters on entering = %d, want 1", got)
	}
	// A spell with mana value 1 matches the one counter.
	effectstest.FloatMana(t, g, me, "{C}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "One Drop", TypeLine: "Instant", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("the opponent is at %d, want 39 (1 damage)", opp.Life)
	}
	if got := oldB236Counters(g, doom, "doom"); got != 2 {
		t.Errorf("doom counters after firing = %d, want 2", got)
	}
}

func TestOldB236ImminentDoomIgnoresASpellOfAnotherManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	doom := effectstest.CastCatalogSpell(t, g, "Imminent Doom", "Enchantment", oldB236DoomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Two Drop", TypeLine: "Instant", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a mana value 2 spell triggered Imminent Doom with one counter")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, doom, "doom"); got != 1 {
		t.Errorf("doom counters = %d, want 1", got)
	}
}

// --- Tenacious Hunter ---

func TestOldB236TenaciousHunterHasVigilanceAndDeathtouchWhileAnyCreatureHasAMinusCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := effectstest.B12Push(g, me.ID, "Tenacious Hunter", "Creature — Crocodile", oldB236HunterOracle, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	for _, kw := range []string{"vigilance", "deathtouch"} {
		if effectstest.HasEffectiveKeyword(t, g, hunter, kw) {
			t.Fatalf("the Hunter has %s with no -1/-1 counter on the battlefield", kw)
		}
	}
	oldB236PutCounter(g, bear, opp.ID)
	g.RunStateChecksForTest()
	for _, kw := range []string{"vigilance", "deathtouch"} {
		if !effectstest.HasEffectiveKeyword(t, g, hunter, kw) {
			t.Errorf("the Hunter lacks %s while a creature has a -1/-1 counter", kw)
		}
	}
}

// --- Torment of Venom ---

func oldB236Torment(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Torment of Venom", "Instant", oldB236TormentOracle, oldB236Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	_ = me
	return g, me, opp, bear
}

func TestOldB236TormentOfVenomPutsThreeCountersAndLetsThemLoseTheLife(t *testing.T) {
	g, _, opp, bear := oldB236Torment(t)
	if got := oldB236Counters(g, bear, game.CounterMinusOne); got != 3 {
		t.Fatalf("-1/-1 counters = %d, want 3", got)
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 37 {
		t.Errorf("life = %d, want 37", opp.Life)
	}
}

func TestOldB236TormentOfVenomLetsThemDiscardInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = me
	bear := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	card := effectstest.HandCard(opp, "Spare", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Torment of Venom", "Instant", oldB236TormentOracle, oldB236Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	// Offers: lose life, discard (no other nonland permanent to sacrifice).
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	effectstest.AnswerDiscard(t, g, opp.ID, card)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(card) {
		t.Error("the card was not discarded")
	}
	if opp.Life != 40 {
		t.Errorf("life = %d, want 40 (paid with a discard)", opp.Life)
	}
}

func TestOldB236TormentOfVenomLetsThemSacrificeAnotherNonlandPermanentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	rock := effectstest.B12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	land := effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Torment of Venom", "Instant", oldB236TormentOracle, oldB236Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	// Offers: lose life, sacrifice (an empty hand has nothing to discard).
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	c := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if effectstest.HasID(c.ChooseCards, bear) || effectstest.HasID(c.ChooseCards, land) {
		t.Errorf("the creature or a land was offered for sacrifice: %v", c.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if opp.Life != 40 {
		t.Errorf("life = %d, want 40", opp.Life)
	}
}

// --- Magmaroth ---

func TestOldB236MagmarothGainsACounterEachUpkeepAndLosesOneForANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mag := effectstest.B12Push(g, me.ID, "Magmaroth", "Creature — Elemental", oldB236MagmarothOID, 5, 5)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, mag, game.CounterMinusOne); got != 1 {
		t.Fatalf("-1/-1 counters after your upkeep = %d, want 1", got)
	}
	effectstest.AdvanceToMain(t, g)
	// A creature spell leaves the counter; a noncreature spell removes it.
	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, mag, game.CounterMinusOne); got != 1 {
		t.Errorf("a creature spell changed the counters to %d", got)
	}
	effectstest.CastCatalogSpell(t, g, "A Trick", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, mag, game.CounterMinusOne); got != 0 {
		t.Errorf("a noncreature spell left %d counters, want 0", got)
	}
}

// --- Ammit Eternal ---

func TestOldB236AmmitEternalGrowsSmallerWhenAnOpponentCastsAndResetsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ammit := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Ammit Eternal", OracleID: oldB236AmmitOracle,
		TypeLine: "Creature — Zombie Crocodile Demon", Power: 5, Toughness: 5})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.CastCatalogSpell(t, g, "Their Trick", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 0 })
	if got := oldB236Counters(g, ammit, game.CounterMinusOne); got != 1 {
		t.Fatalf("-1/-1 counters after an opponent's spell = %d, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "My Trick", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, ammit, game.CounterMinusOne); got != 1 {
		t.Errorf("your own spell changed the counters to %d", got)
	}
	effectstest.AttackWith(t, g, opp.ID, ammit)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, ammit, game.CounterMinusOne); got != 0 {
		t.Errorf("combat damage to a player left %d counters, want 0", got)
	}
}

// --- Banewhip Punisher ---

func TestOldB236BanewhipPunisherMayPutACounterThenSacrificeToDestroyIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 4, 4)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Ogre", 4, 4)
	punisher := effectstest.CastHoldCreature(t, g, "Banewhip Punisher", "Creature — Human Warrior", oldB236BanewhipOracle, 2, 2)
	_ = punisher
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, victim, game.CounterMinusOne); got != 1 {
		t.Fatalf("-1/-1 counters on the target = %d, want 1", got)
	}
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == punisher {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, punisher, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB236Card(bystander)}); err == nil {
		t.Fatal("a creature with no -1/-1 counter was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, punisher, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB236Card(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature with a -1/-1 counter survived")
	}
	if g.Battlefield.Contains(punisher) {
		t.Error("the Punisher was not sacrificed")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("another creature was destroyed")
	}
}

func TestOldB236BanewhipPunisherDeclinedPutsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 4, 4)
	effectstest.CastHoldCreature(t, g, "Banewhip Punisher", "Creature — Human Warrior", oldB236BanewhipOracle, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, victim, game.CounterMinusOne); got != 0 {
		t.Errorf("declining still put %d counters", got)
	}
}

// --- Obelisk Spider ---

func TestOldB236ObeliskSpiderCountersWhatItDamagesAndDrainsForEachCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Obelisk Spider", OracleID: oldB236SpiderOracle,
		TypeLine: "Creature — Spider", Power: 1, Toughness: 4, Keywords: []string{"reach"}})
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.Pr7bAttack(t, g, opp.ID, spider)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: spider})
	myLife, theirLife := me.Life, opp.Life
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB236Counters(g, wall, game.CounterMinusOne); got != 1 {
		t.Fatalf("-1/-1 counters on the damaged creature = %d, want 1", got)
	}
	if opp.Life != theirLife-1 || me.Life != myLife+1 {
		t.Errorf("drain: you %d -> %d, they %d -> %d; want +1 and -1", myLife, me.Life, theirLife, opp.Life)
	}
}

func TestOldB236ObeliskSpiderIgnoresCountersYouDidNotPut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Obelisk Spider", OracleID: oldB236SpiderOracle,
		TypeLine: "Creature — Spider", Power: 1, Toughness: 4, Keywords: []string{"reach"}})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	myLife, theirLife := me.Life, opp.Life
	oldB236PutCounter(g, bear, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != theirLife {
		t.Errorf("a counter nobody of yours put drained: you %d -> %d, they %d -> %d", myLife, me.Life, theirLife, opp.Life)
	}
}
