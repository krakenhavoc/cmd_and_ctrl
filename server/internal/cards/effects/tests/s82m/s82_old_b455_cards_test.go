package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b455_cards_test.go — Commander 2011, token makers and graveyard
// recursion, S82 older sets, batch 455 (#5574). Each test plays the real
// catalog entry.

const (
	oldB455Catapult  = "56d13879-2dc5-4d84-918b-4bd1a5fc2a33"
	oldB455Dragons   = "88e912e3-0548-4a4a-94c1-c804093ca1b0"
	oldB455Spawn     = "afc59d95-d85c-4ab8-b6c5-c78b608b53b7"
	oldB455Hedge     = "b89b285d-62d3-4013-b881-23f0f236337d"
	oldB455Riku      = "f697c78d-7e4f-4320-bfc6-2a25e6d7dc94"
	oldB455Nucklavee = "09c2d504-f2e1-4799-b948-c7b9502d5088"
	oldB455Feast     = "0a35df9d-80a0-4b6f-91d5-dad980af47e1"
	oldB455Fire      = "1890ef1b-38e5-432d-b040-f33cea7814e7"
	oldB455Rebirth   = "522297b1-493a-4c62-a82a-3017c44183e8"
	oldB455Bear      = "Creature — Bear"
)

func TestOldB455EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB455Catapult:  "Acorn Catapult",
		oldB455Dragons:   "Death by Dragons",
		oldB455Spawn:     "Spawnwrithe",
		oldB455Hedge:     "Gwyllion Hedge-Mage",
		oldB455Riku:      "Riku of Two Reflections",
		oldB455Nucklavee: "Nucklavee",
		oldB455Feast:     "Footbottom Feast",
		oldB455Fire:      "Punishing Fire",
		oldB455Rebirth:   "Vengeful Rebirth",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB455Count(g *game.Game, owner uuid.UUID, name string, token bool) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name && c.IsToken() == token {
			n++
		}
	}
	return n
}

func oldB455Plain(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
}

// --- Acorn Catapult ---------------------------------------------------------

func TestOldB455AcornCatapultPingsAndGivesTheTargetsControllerASquirrel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	catapult := effectstest.B12Push(g, me.ID, "Acorn Catapult", "Artifact", oldB455Catapult, 0, 0)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB455Bear, 2, 2)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.P7Activate(t, g, me, catapult, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(victim)})
	if got := oldB455Count(g, opp.ID, "Squirrel", true); got != 1 {
		t.Errorf("the creature's controller has %d Squirrels, want 1", got)
	}
	if got := oldB455Count(g, me.ID, "Squirrel", true); got != 0 {
		t.Errorf("the activator got %d Squirrels", got)
	}
	if got := effectstest.Pr6Marked(g, victim); got != 1 {
		t.Errorf("the target took %d damage, want 1", got)
	}
}

func TestOldB455AcornCatapultAtAPlayerGivesThatPlayerASquirrel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	catapult := effectstest.B12Push(g, me.ID, "Acorn Catapult", "Artifact", oldB455Catapult, 0, 0)
	life := opp.Life
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.P7Activate(t, g, me, catapult, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, opp.Life)
	}
	if got := oldB455Count(g, opp.ID, "Squirrel", true); got != 1 {
		t.Errorf("the player has %d Squirrels, want 1", got)
	}
}

func TestOldB455AcornCatapultNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	catapult := effectstest.B12Push(g, me.ID, "Acorn Catapult", "Artifact", oldB455Catapult, 0, 0)
	err := g.ActivateCatalogAbility(me.ID, catapult, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if err == nil {
		t.Error("the ability was activated with no mana to pay {1}")
	}
}

// --- Death by Dragons --------------------------------------------------------

func TestOldB455DeathByDragonsGivesEveryoneButTheTargetADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Death by Dragons", "Sorcery", oldB455Dragons, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 1
		if p.ID == opp.ID {
			want = 0
		}
		if got := oldB455Count(g, p.ID, "Dragon", true); got != want {
			t.Errorf("%s has %d Dragons, want %d", p.Name, got, want)
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dragon" && (c.CurrentPower() != 5 || c.CurrentToughness() != 5 || !game.HasKeyword(&c, "flying")) {
			t.Errorf("the Dragon is %d/%d, flying %v", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"))
		}
	}
	_ = me
}

// --- Spawnwrithe -------------------------------------------------------------

func TestOldB455SpawnwritheCopiesItselfOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	worm := effectstest.B12Push(g, me.ID, "Spawnwrithe", "Creature — Elemental", oldB455Spawn, 2, 2)
	effectstest.DealCombatDamageToPlayer(g, worm, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB455Count(g, me.ID, "Spawnwrithe", true); got != 1 {
		t.Fatalf("%d Spawnwrithe tokens, want 1", got)
	}
	if got := oldB455Count(g, me.ID, "Spawnwrithe", false); got != 1 {
		t.Errorf("%d Spawnwrithe cards, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spawnwrithe" && c.IsToken() && !game.HasKeyword(&c, "trample") {
			t.Error("the token copy has no trample")
		}
	}
}

// --- Gwyllion Hedge-Mage ------------------------------------------------------

func oldB455Lands(g *game.Game, owner uuid.UUID, n int, typeLine string) {
	for i := 0; i < n; i++ {
		oldB455Plain(g, owner, "Land", typeLine)
	}
}

func TestOldB455GwyllionMakesAKithkinOnlyWithTwoPlains(t *testing.T) {
	for _, tc := range []struct {
		plains int
		want   int
	}{{1, 0}, {2, 1}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		oldB455Lands(g, me.ID, tc.plains, "Basic Land — Plains")
		effectstest.EnterFromHand(t, g, me.ID, "Gwyllion Hedge-Mage", "Creature — Hag Wizard", oldB455Hedge)
		effectstest.PassPriorityAroundTable(t, g)
		if tc.want > 0 {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if got := oldB455Count(g, me.ID, "Kithkin Soldier", true); got != tc.want {
			t.Errorf("%d Plains: %d Kithkin Soldiers, want %d", tc.plains, got, tc.want)
		}
	}
}

func TestOldB455GwyllionPutsAMinusCounterOnlyWithTwoSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB455Lands(g, me.ID, 2, "Basic Land — Swamp")
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB455Bear, 2, 2)
	effectstest.EnterFromHand(t, g, me.ID, "Gwyllion Hedge-Mage", "Creature — Hag Wizard", oldB455Hedge)
	effectstest.PickCard(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, foe); !ok || c.Counters[game.CounterMinusOne] != 1 {
		t.Errorf("the target has %v counters, want one -1/-1", c.Counters)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	oldB455Lands(g2, me2.ID, 1, "Basic Land — Swamp")
	foe2 := effectstest.B12Creature(g2, opp2.ID, "Their Bear", oldB455Bear, 2, 2)
	effectstest.EnterFromHand(t, g2, me2.ID, "Gwyllion Hedge-Mage", "Creature — Hag Wizard", oldB455Hedge)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.LatestPickTarget(g2, me2.ID) != nil {
		t.Error("asked for a target with a single Swamp")
	}
	if c, _ := effectstest.BattlefieldCard(g2, foe2); c.Counters[game.CounterMinusOne] != 0 {
		t.Error("a counter was put on the creature with a single Swamp")
	}
}

// --- Nucklavee ----------------------------------------------------------------

func TestOldB455NucklaveeReturnsARedSorceryAndABlueInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	redSorcery := effectstest.PushGraveyardMaterial(me, game.Card{InstanceID: uuid.New(), Name: "Red Sorcery", TypeLine: "Sorcery", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	redInstant := effectstest.PushGraveyardMaterial(me, game.Card{InstanceID: uuid.New(), Name: "Red Instant", TypeLine: "Instant", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	blueInstant := effectstest.PushGraveyardMaterial(me, game.Card{InstanceID: uuid.New(), Name: "Blue Instant", TypeLine: "Instant", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	blueSorcery := effectstest.PushGraveyardMaterial(me, game.Card{InstanceID: uuid.New(), Name: "Blue Sorcery", TypeLine: "Sorcery", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	effectstest.EnterFromHand(t, g, me.ID, "Nucklavee", "Creature — Beast", oldB455Nucklavee)
	for i := 0; i < 2; i++ {
		pick := effectstest.LatestPickTarget(g, me.ID)
		if pick == nil {
			t.Fatalf("no target prompt for trigger %d", i+1)
		}
		for _, wrong := range []uuid.UUID{redInstant, blueSorcery} {
			if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: wrong}); err == nil {
				t.Fatalf("trigger %d accepted a card that is neither a red sorcery nor a blue instant", i+1)
			}
		}
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: redSorcery}); err != nil {
			effectstest.PickCard(t, g, me.ID, blueInstant)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if effectstest.LatestTriggerPrompt(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(redSorcery) || !me.Hand.Contains(blueInstant) {
		t.Error("the red sorcery and the blue instant did not both return")
	}
	if me.Hand.Contains(redInstant) || me.Hand.Contains(blueSorcery) || !me.Graveyard.Contains(redInstant) || !me.Graveyard.Contains(blueSorcery) {
		t.Error("a card of the wrong colour or type moved")
	}
}

// --- Footbottom Feast -----------------------------------------------------------

func TestOldB455FootbottomFeastPutsCreaturesOnTopThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardCreature(g, me.ID, "Grave Bear", "{1}{G}")
	b := effectstest.PushGraveyardCreature(g, me.ID, "Grave Wolf", "{2}{G}")
	spell := effectstest.PushGraveyardCardTyped(me, "Grave Spell", "Sorcery")
	handBefore := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Footbottom Feast", "Instant", oldB455Feast, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(a) && me.Graveyard.Contains(b) {
		t.Fatal("neither creature left the graveyard")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("a noncreature card was moved")
	}
	if len(me.Hand.Cards) != handBefore+1 {
		t.Errorf("hand went %d -> %d, want one card drawn", handBefore, len(me.Hand.Cards))
	}
	// The draw takes one of the two tucked cards; the other is now on top.
	drawn := 0
	for _, id := range []uuid.UUID{a, b} {
		if me.Hand.Contains(id) {
			drawn++
		}
	}
	if drawn != 1 {
		t.Errorf("%d of the tucked creatures were drawn, want exactly the top one", drawn)
	}
}

func TestOldB455FootbottomFeastWithNoTargetsStillDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	handBefore := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Footbottom Feast", "Instant", oldB455Feast, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != handBefore+1 {
		t.Errorf("hand went %d -> %d, want one card drawn", handBefore, len(me.Hand.Cards))
	}
}

// --- Punishing Fire ------------------------------------------------------------

func TestOldB455PunishingFireDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Punishing Fire", "Instant", oldB455Fire, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, opp.Life)
	}
}

func TestOldB455PunishingFireReturnsWhenAnOpponentGainsLifeAndYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := effectstest.SeedGraveyardCard(t, g, "Punishing Fire", "Instant", oldB455Fire)
	effectstest.AfrB07Gain(g, opp, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "R")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) || me.Graveyard.Contains(card) {
		t.Error("paying {R} did not return Punishing Fire to hand")
	}
}

func TestOldB455PunishingFireStaysWhenDeclinedOrWhenYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := effectstest.SeedGraveyardCard(t, g, "Punishing Fire", "Instant", oldB455Fire)
	effectstest.AfrB07Gain(g, me, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("asked to pay when its owner gained life")
	}
	effectstest.AfrB07Gain(g, opp, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "R")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(card) {
		t.Error("declining moved the card out of the graveyard")
	}
}

// --- Vengeful Rebirth -----------------------------------------------------------

func TestOldB455VengefulRebirthReturnsANonlandAndDealsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := effectstest.PushGraveyardCreature(g, me.ID, "Grave Giant", "{3}{G}")
	life := opp.Life
	spell := effectstest.CastCatalogSpell(t, g, "Vengeful Rebirth", "Sorcery", oldB455Rebirth,
		[]game.TargetRef{{Kind: game.TargetCard, ID: card}, {Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the card did not return to hand")
	}
	if opp.Life != life-4 {
		t.Errorf("life %d -> %d, want -4 (the card's mana value)", life, opp.Life)
	}
	if me.Graveyard.Contains(spell) {
		t.Error("Vengeful Rebirth went to the graveyard instead of exile")
	}
	exiled := false
	for _, c := range g.Exile.Cards {
		if c.InstanceID == spell {
			exiled = true
		}
	}
	if !exiled {
		t.Error("Vengeful Rebirth is not in exile")
	}
}

func TestOldB455VengefulRebirthDealsNothingForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushGraveyardPermanent(me, "Grave Forest", "Basic Land — Forest", "")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Vengeful Rebirth", "Sorcery", oldB455Rebirth,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}, {Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land did not return to hand")
	}
	if opp.Life != life {
		t.Errorf("life %d -> %d, want no damage for a land", life, opp.Life)
	}
}

func TestOldB455VengefulRebirthRefusesATargetOutsideYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.PushGraveyardCreature(g, opp.ID, "Their Giant", "{3}{G}")
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Vengeful Rebirth", "Sorcery", oldB455Rebirth,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetPlayer, ID: opp.ID}})
	if err == nil {
		t.Error("a card from an opponent's graveyard was accepted")
	}
}

// --- Riku of Two Reflections ------------------------------------------------------

func TestOldB455RikuCopiesAnInstantWhenYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Riku of Two Reflections", "Legendary Creature — Human Wizard", oldB455Riku, 2, 2)
	life := opp.Life
	effectstest.CastFromSeat(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{U}{R}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("retarget: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opp.Life != life-6 {
		t.Errorf("life %d -> %d, want -6 (Bolt and its copy)", life, opp.Life)
	}
}

func TestOldB455RikuMakesNoCopyWhenYouDeclineToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Riku of Two Reflections", "Legendary Creature — Human Wizard", oldB455Riku, 2, 2)
	life := opp.Life
	effectstest.CastFromSeat(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{U}{R}")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3 (no copy)", life, opp.Life)
	}
}

func TestOldB455RikuCopiesAnEnteringNontokenCreatureWhenYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Riku of Two Reflections", "Legendary Creature — Human Wizard", oldB455Riku, 2, 2)
	effectstest.EnterFromHand(t, g, me.ID, "Plain Bear", oldB455Bear, "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{G}{U}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB455Count(g, me.ID, "Plain Bear", true); got != 1 {
		t.Errorf("%d token copies of the Bear, want 1", got)
	}
	// The token entering is a token: no further question.
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a token entering asked for {G}{U} again")
	}
}
