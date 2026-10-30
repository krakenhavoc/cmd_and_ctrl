package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a6_cards_test.go — Marvel Super Heroes, half A, sixth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshA6MoonstoneOracle   = "aeeb50f2-8e02-4d40-a1dd-28b1c5c85c7e"
	mshA6IncursionOracle   = "d3678fce-6897-4f2c-bf50-243facb6581b"
	mshA6NickFuryOracle    = "62916107-14c0-4516-9fed-a032b1a80e5d"
	mshA6NinjaOracle       = "c38f7447-e646-438d-81bd-fe2a71c8a79c"
	mshA6OriginOracle      = "d7f3178d-7b7c-4346-af50-a7b0206bee22"
	mshA6PatriotOracle     = "1df7a09f-fbc8-4c2a-a903-bab9e680eda7"
	mshA6PhotonBlastOracle = "a5424e33-ff4e-4c82-8aa9-9d25ced7830a"
	mshA6BrokerOracle      = "2b3d76e7-faaa-4e6f-9e21-a88dd719247e"
	mshA6PunchOracle       = "454c4e8b-e417-4381-a931-6a11356ef64e"
	mshA6QuakeOracle       = "3966c94b-e597-4b3f-a4f5-001d52fff9fe"
	mshA6RaftOracle        = "60754d4b-2add-4928-8d30-8ce9c0f54f91"
	mshA6RedGuardianOracle = "e85567e8-1102-4d44-a15b-d5f1fdef8465"
	mshA6RecruitOracle     = "4e60c38c-cbf5-4ff7-8fff-634f7e3086cc"
)

func TestMshA6EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mshA6MoonstoneOracle:   "Moonstone, Harsh Mistress",
		mshA6IncursionOracle:   "Multiversal Incursion",
		mshA6NickFuryOracle:    "Nick Fury, Agent of S.H.I.E.L.D.",
		mshA6NinjaOracle:       "Ninja of the Hand",
		mshA6OriginOracle:      "Origin of the Avengers",
		mshA6PatriotOracle:     "Patriot, Shield Wielder",
		mshA6PhotonBlastOracle: "Photon Blast Barrage",
		mshA6BrokerOracle:      "Powerful Broker",
		mshA6PunchOracle:       "Punishing Punch",
		mshA6QuakeOracle:       "Quake, Agent of S.H.I.E.L.D.",
		mshA6RaftOracle:        "Raft Security Officer",
		mshA6RedGuardianOracle: "Red Guardian, Super-Soldier",
		mshA6RecruitOracle:     "Red Room Recruit",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestMshA6KeywordsAreDeclared(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct{ name, typeLine, oracle, keyword string }{
		{"Moonstone, Harsh Mistress", "Legendary Creature — Human Doctor Villain", mshA6MoonstoneOracle, "flying"},
		{"Ninja of the Hand", "Creature — Human Ninja Villain", mshA6NinjaOracle, "deathtouch"},
		{"Red Guardian, Super-Soldier", "Legendary Creature — Human Soldier Villain", mshA6RedGuardianOracle, "flash"},
	} {
		id := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		if !mshA3HasKeyword(g, id, tc.keyword) {
			t.Errorf("%s has no %s", tc.name, tc.keyword)
		}
	}
}

// mshA6Battlefield lists the battlefield permanents with a name.
func mshA6Battlefield(g *game.Game, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				out = append(out, c)
			}
		}
	})
	return out
}

// mshA6Graveyard puts a creature card straight into a graveyard.
func mshA6Graveyard(g *game.Game, p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Owner: p.ID, Controller: p.ID})
	})
	return id
}

// mshA6Discard makes `p` discard the given card from hand.
func mshA6Discard(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p.ID, Source: uuid.Nil, N: 1})
	})
	answerDiscard(t, g, p.ID, card)
}

// --- Moonstone, Harsh Mistress -------------------------------------------------------------

const mshA6MoonstoneType = "Legendary Creature — Human Doctor Villain"

func TestMshA6MoonstoneExilesADiscardedCardAndLetsYouPlayItThroughYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Moonstone, Harsh Mistress", mshA6MoonstoneType, mshA6MoonstoneOracle, false)
	advanceToMain(t, g)
	card := handCardForTest(me, "Spare Spell", "Sorcery", "")
	mshA6Discard(t, g, me, card)
	assistCardSettle(t, g, me.ID, nil, true)
	if z := b12ZoneOf(g, card); z != game.ZoneExile {
		t.Fatalf("the discarded card is in %v, want exile", z)
	}
	if !b12ExiledPlayableBy(t, g, card, me.ID) {
		t.Fatal("the exiled card is not playable by its discarder")
	}
	// Through the rest of this turn and all of the next one of mine.
	powerUpPassTo(t, g, 0)
	if !b12ExiledPlayableBy(t, g, card, me.ID) {
		t.Error("the permission ended before the end of my next turn")
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if b12ExiledPlayableBy(t, g, card, me.ID) {
		t.Error("the permission outlived my next turn")
	}
}

func TestMshA6MoonstoneDeclinedLeavesTheCardInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Moonstone, Harsh Mistress", mshA6MoonstoneType, mshA6MoonstoneOracle, false)
	advanceToMain(t, g)
	card := handCardForTest(me, "Spare Spell", "Sorcery", "")
	mshA6Discard(t, g, me, card)
	assistCardSettle(t, g, me.ID, nil, false)
	if z := b12ZoneOf(g, card); z != game.ZoneGraveyard {
		t.Errorf("a declined card is in %v, want the graveyard", z)
	}
}

func TestMshA6MoonstoneIgnoresAnOpponentsDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Moonstone, Harsh Mistress", mshA6MoonstoneType, mshA6MoonstoneOracle, false)
	advanceToMain(t, g)
	card := handCardForTest(opp, "Their Spell", "Sorcery", "")
	mshA6Discard(t, g, opp, card)
	assistCardSettle(t, g, me.ID, nil, true)
	if z := b12ZoneOf(g, card); z != game.ZoneGraveyard {
		t.Errorf("an opponent's discarded card is in %v, want their graveyard", z)
	}
}

// --- Ninja of the Hand -----------------------------------------------------------------------

func TestMshA6NinjaPowerUpMakesEachOpponentDiscardAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	spare := handCardForTest(opp, "Their Spell", "Sorcery", "")
	ninja := castPowerUpCreature(t, g, me, "Ninja of the Hand", "Creature — Human Ninja Villain", "{2}{B}", "{C}{C}{B}", mshA6NinjaOracle, 2)
	if got := chargedPowerUp(t, g, me.ID, ninja); got != "{2}" {
		t.Fatalf("the turn it entered: %q, want {2} ({4}{B} less {2}{B})", got)
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if discardOwed(g, opp.ID) != 1 {
		t.Fatalf("opponent owes %d discards, want 1", discardOwed(g, opp.ID))
	}
	answerDiscard(t, g, opp.ID, spare)
	if got := plusOnes(g, ninja); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

// --- Red Room Recruit ------------------------------------------------------------------------

func TestMshA6RedRoomRecruitConnivesWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := handCardForTest(me, "Spare Spell", "Sorcery", "")
	before := me.Hand.Size()
	id := castHoldCreature(t, g, "Red Room Recruit", "Creature — Human Spy Villain", mshA6RecruitOracle, 1, 2)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand %d, want %d (the draw, before the discard)", got, before+1)
	}
	answerDiscard(t, g, me.ID, spell)
	if n := hobA8Counters(t, g, id); n != 1 {
		t.Errorf("%d +1/+1 counters after discarding a nonland card, want 1", n)
	}
}

func TestMshA6RedRoomRecruitGetsNoCounterForALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := handCardForTest(me, "Spare Forest", "Basic Land — Forest", "")
	id := castHoldCreature(t, g, "Red Room Recruit", "Creature — Human Spy Villain", mshA6RecruitOracle, 1, 2)
	answerDiscard(t, g, me.ID, forest)
	if n := hobA8Counters(t, g, id); n != 0 {
		t.Errorf("%d +1/+1 counters after discarding a land, want 0", n)
	}
}

// --- Patriot, Shield Wielder -----------------------------------------------------------------

func TestMshA6PatriotPumpsAndShieldsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	patriot := pushCatalogPermanent(g, me.ID, "Patriot, Shield Wielder", "Legendary Creature — Human Hero", mshA6PatriotOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, patriot, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power = %d, want 4", got)
	}
	if !mshA3HasKeyword(g, bear, "hexproof") {
		t.Error("the target did not gain hexproof")
	}
	if !b16Tapped(t, g, patriot) {
		t.Error("Patriot did not tap for the cost")
	}
}

func TestMshA6PatriotCannotTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	patriot := pushCatalogPermanent(g, me.ID, "Patriot, Shield Wielder", "Legendary Creature — Human Hero", mshA6PatriotOracle, false)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, patriot, 0, game.ActivateAbilityParams{Targets: cardRefs(patriot)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("targeting himself: err = %v, want ErrIllegalTarget", err)
	}
}

func TestMshA6PatriotCannotTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	patriot := pushCatalogPermanent(g, me.ID, "Patriot, Shield Wielder", "Legendary Creature — Human Hero", mshA6PatriotOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, patriot, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("targeting an opponent's creature: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Raft Security Officer -------------------------------------------------------------------

// mshA6AddMana puts mana in the pool for a strict-gate payment.
func mshA6AddMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect(%s): %v", mana, err)
	}
}

func mshA6PriceOf(t *testing.T, g *game.Game, who, src uuid.UUID, targets []game.TargetRef) string {
	t.Helper()
	p, err := g.PriceActivation(who, src, 0, targets)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	return p.Total.String()
}

func TestMshA6RaftSecurityOfficerCostsOneLessAgainstASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	officer := pushCatalogPermanent(g, me.ID, "Raft Security Officer", "Creature — Human Soldier", mshA6RaftOracle, false)
	small := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	if got := mshA6PriceOf(t, g, me.ID, officer, cardRefs(small)); got != "{1}" {
		t.Errorf("against power 3: %q, want {1}", got)
	}
	if got := mshA6PriceOf(t, g, me.ID, officer, cardRefs(big)); got != "{2}" {
		t.Errorf("against power 4: %q, want {2}", got)
	}
}

func TestMshA6RaftSecurityOfficerTapsTheTargetForTheReducedCost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	officer := pushCatalogPermanent(g, me.ID, "Raft Security Officer", "Creature — Human Soldier", mshA6RaftOracle, false)
	small := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 3, 3)
	mshA6AddMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, officer, 0, game.ActivateAbilityParams{Targets: cardRefs(small), Strict: true}); err != nil {
		t.Fatalf("activate for {1}: %v", err)
	}
	resolveTopOfStack(t, g)
	if !b16Tapped(t, g, small) {
		t.Error("the target is untapped")
	}
}

func TestMshA6RaftSecurityOfficerCannotPayOneAgainstABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	officer := pushCatalogPermanent(g, me.ID, "Raft Security Officer", "Creature — Human Soldier", mshA6RaftOracle, false)
	big := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	mshA6AddMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, officer, 0, game.ActivateAbilityParams{Targets: cardRefs(big), Strict: true}); err == nil {
		t.Fatal("a power-4 target was tapped for {1}")
	}
}

// --- Quake, Agent of S.H.I.E.L.D. ---------------------------------------------------------------

func TestMshA6QuakeTapsATargetCreatureWhenYouCastANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Quake, Agent of S.H.I.E.L.D.", "Legendary Creature — Inhuman Spy Hero", mshA6QuakeOracle, false)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castNoncreature(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	if !b16Tapped(t, g, victim) {
		t.Error("the targeted creature is untapped")
	}
}

func TestMshA6QuakeTapsALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Quake, Agent of S.H.I.E.L.D.", "Legendary Creature — Inhuman Spy Hero", mshA6QuakeOracle, false)
	land := pushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	castNoncreature(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: land}, true)
	if !b16Tapped(t, g, land) {
		t.Error("the targeted land is untapped")
	}
}

func TestMshA6QuakeIgnoresACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Quake, Agent of S.H.I.E.L.D.", "Legendary Creature — Inhuman Spy Hero", mshA6QuakeOracle, false)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Filler Bear", "Creature — Bear", "", nil)
	assistCardSettle(t, g, me.ID, nil, true)
	if b16Tapped(t, g, victim) {
		t.Error("a creature spell triggered Quake")
	}
}

// --- Punishing Punch -------------------------------------------------------------------------

func TestMshA6PunishingPunchDealsTwiceThePowerOneSidedly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 2, 6)
	floatMana(t, g, me, "{C}{C}{G}")
	if _, err := castFromHandErr(g, me, "Punishing Punch", "Instant", "{2}{G}", mshA6PunchOracle, game.CastSpellParams{Targets: cardRefs(mine, theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTopOfStack(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("a 2/6 survived 6 damage from a 3-power creature")
	}
	if c := b12Card(t, g, mine); c.DamageMarked != 0 {
		t.Errorf("the source took %d damage back; the damage is one-sided", c.DamageMarked)
	}
}

func TestMshA6PunishingPunchCostsTwoLessWithTwoCreatureCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 2, 6)
	price := func() int {
		return hobB3Price(t, g, me, "Punishing Punch", "Instant", "{2}{G}", mine, theirs)
	}
	if got := price(); got != 3 {
		t.Errorf("empty graveyard: %d, want 3", got)
	}
	mshA6Graveyard(g, me, "Dead Bear")
	if got := price(); got != 3 {
		t.Errorf("one creature card: %d, want 3", got)
	}
	mshA6Graveyard(g, me, "Dead Wolf")
	if got := price(); got != 1 {
		t.Errorf("two creature cards: %d, want 1", got)
	}
}

func TestMshA6PunishingPunchCannotTargetYourOwnCreatureAsTheVictim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	other := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{G}")
	if _, err := castFromHandErr(g, me, "Punishing Punch", "Instant", "{2}{G}", mshA6PunchOracle, game.CastSpellParams{Targets: cardRefs(mine, other)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("err = %v, want ErrIllegalTarget", err)
	}
}

// --- Red Guardian, Super-Soldier -------------------------------------------------------------

const mshA6RedGuardianType = "Legendary Creature — Human Soldier Villain"

func mshA6RedGuardianCard() game.Card {
	return game.Card{Name: "Red Guardian, Super-Soldier", TypeLine: mshA6RedGuardianType, OracleID: mshA6RedGuardianOracle, Power: 2, Toughness: 2}
}

func TestMshA6RedGuardianDestroysAnOpponentsCreatureThatDealtDamageThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	biter := b12Creature(g, opp.ID, "Their Biter", "Creature — Bear", 2, 2)
	idle := b12Creature(g, opp.ID, "Their Idler", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(biter, me.ID, 1) })
	castFromHand(t, g, mshA6RedGuardianCard(), game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: biter}, true)
	if g.Battlefield.Contains(biter) {
		t.Error("the creature that dealt damage survived")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("a creature that dealt no damage was destroyed")
	}
}

func TestMshA6RedGuardianCountsDamageDealtToACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	biter := b12Creature(g, opp.ID, "Their Biter", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 5)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(biter, mine, 2) })
	castFromHand(t, g, mshA6RedGuardianCard(), game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: biter}, true)
	if g.Battlefield.Contains(biter) {
		t.Error("the creature that damaged a creature survived")
	}
}

func TestMshA6RedGuardianNeedsADamageDealer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	idle := b12Creature(g, opp.ID, "Their Idler", "Creature — Bear", 2, 2)
	castFromHand(t, g, mshA6RedGuardianCard(), game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	if !g.Battlefield.Contains(idle) {
		t.Error("a creature that dealt no damage was destroyed")
	}
}

func TestMshA6RedGuardianIgnoresYourOwnDamageDealer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "My Biter", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, g.Seats[1].ID, 1) })
	castFromHand(t, g, mshA6RedGuardianCard(), game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was destroyed")
	}
}

// --- Photon Blast Barrage --------------------------------------------------------------------

// mshA6RetargetCopies settles the stack, answering each copy's re-target
// offer with the next of `picks` (then the last again) and counting offers.
func mshA6RetargetCopies(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) int {
	t.Helper()
	seen := 0
	for i := 0; i < 64; i++ {
		if p := latestPickTarget(g, chooser); p != nil {
			pick := picks[len(picks)-1]
			if seen < len(picks) {
				pick = picks[seen]
			}
			seen++
			if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: pick}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return seen
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
	return seen
}

func TestMshA6PhotonBlastBarrageCopiesItXTimesAndEachCopyHitsItsOwnTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	a := b12Creature(g, opp.ID, "Big Target", "Creature — Giant", 2, 5)
	b := b12Creature(g, opp.ID, "Small Target", "Creature — Bear", 2, 2)
	if _, err := hobB3Cast(t, g, me, "Photon Blast Barrage", "Sorcery", "{X}{R}{R}", "{C}{C}{R}{R}", game.CastSpellParams{XValue: 2, Targets: cardRefs(a)}); err != nil {
		t.Fatalf("cast with X=2: %v", err)
	}
	if got := mshA6RetargetCopies(t, g, me.ID, b, b); got != 2 {
		t.Errorf("%d re-target offers, want 2 (one per copy)", got)
	}
	if g.Battlefield.Contains(b) {
		t.Error("two copies of 1 damage did not kill the 2/2 they were aimed at")
	}
	if c := b12Card(t, g, a); c.DamageMarked != 1 {
		t.Errorf("the original's target has %d damage, want 1", c.DamageMarked)
	}
}

func TestMshA6PhotonBlastBarrageWithXZeroMakesNoCopies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	a := b12Creature(g, opp.ID, "Big Target", "Creature — Giant", 2, 5)
	if _, err := hobB3Cast(t, g, me, "Photon Blast Barrage", "Sorcery", "{X}{R}{R}", "{R}{R}", game.CastSpellParams{XValue: 0, Targets: cardRefs(a)}); err != nil {
		t.Fatalf("cast with X=0: %v", err)
	}
	if got := mshA6RetargetCopies(t, g, me.ID, a); got != 0 {
		t.Errorf("%d re-target offers with X=0, want none", got)
	}
	if c := b12Card(t, g, a); c.DamageMarked != 1 {
		t.Errorf("target has %d damage, want 1", c.DamageMarked)
	}
}

// --- Powerful Broker -------------------------------------------------------------------------

func TestMshA6PowerfulBrokerGivesAnotherCounterOfEachKindOnAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	broker := pushCatalogPermanent(g, me.ID, "Powerful Broker", "Creature — Human Villain", mshA6BrokerOracle, false)
	target := b12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(target, game.CounterPlusOne, 3)
		_ = g.AddCounterForEffect(target, "charge", 2)
	})
	b16Activate(t, g, me.ID, broker, 0, game.ActivateAbilityParams{Targets: cardRefs(target)})
	if got := plusOnes(g, target); got != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", got)
	}
	if got := countersOn(g, target, "charge"); got != 3 {
		t.Errorf("charge counters = %d, want 3", got)
	}
}

func TestMshA6PowerfulBrokerGivesAPlayerAnotherCounterOfEachKind(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	broker := pushCatalogPermanent(g, me.ID, "Powerful Broker", "Creature — Human Villain", mshA6BrokerOracle, false)
	g.WithWriteLock(func() {
		_ = g.AddPlayerCounterForEffect(opp.ID, "poison", 2)
		_ = g.AddPlayerCounterForEffect(opp.ID, "energy", 1)
	})
	b16Activate(t, g, me.ID, broker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if got := opp.Counters["poison"]; got != 3 {
		t.Errorf("poison counters = %d, want 3", got)
	}
	if got := opp.Counters["energy"]; got != 2 {
		t.Errorf("energy counters = %d, want 2", got)
	}
}

func TestMshA6PowerfulBrokerDoesNothingToAPermanentWithoutCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	broker := pushCatalogPermanent(g, me.ID, "Powerful Broker", "Creature — Human Villain", mshA6BrokerOracle, false)
	target := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, broker, 0, game.ActivateAbilityParams{Targets: cardRefs(target)})
	if c := b12Card(t, g, target); len(c.Counters) != 0 {
		t.Errorf("a counterless permanent got %v", c.Counters)
	}
}

func TestMshA6PowerfulBrokerIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	broker := pushCatalogPermanent(g, me.ID, "Powerful Broker", "Creature — Human Villain", mshA6BrokerOracle, false)
	target := b12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(target, game.CounterPlusOne, 1) })
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, broker, 0, game.ActivateAbilityParams{Targets: cardRefs(target)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Origin of the Avengers ------------------------------------------------------------------

const mshA6OriginType = "Enchantment — Saga"

func mshA6HandCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID, Power: 1, Toughness: 1})
	return id
}

// mshA6OriginAtChapter casts the Saga (chapter I scries; the scry is
// answered) and walks my turns until `chapter` is on the stack. Earlier
// chapters resolve on the way, so the hand must hold no legal Hero for
// chapter II when `chapter` is III.
func mshA6OriginAtChapter(t *testing.T, g *game.Game, me *game.Player, chapter int) uuid.UUID {
	t.Helper()
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	id := castCatalogSpell(t, g, "Origin of the Avengers", mshA6OriginType, mshA6OriginOracle, nil)
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("chapter I did not scry")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	for ch := 2; ch <= chapter; ch++ {
		advanceToPrecombatMainOf(t, g, 0)
		if ch < chapter {
			passPriorityAroundTable(t, g)
		}
	}
	return id
}

func TestMshA6OriginChapterOneScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	castCatalogSpell(t, g, "Origin of the Avengers", mshA6OriginType, mshA6OriginOracle, nil)
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("chapter I did not scry")
	}
	if len(scry.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(scry.ScryCards))
	}
}

func TestMshA6OriginChapterTwoPutsAHeroFromHandOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	low := mshA6HandCard(me, "Cheap Hero", "Creature — Human Hero", "{1}{W}")
	high := mshA6HandCard(me, "Pricey Hero", "Creature — Human Hero", "{2}{W}{W}")
	villain := mshA6HandCard(me, "Cheap Villain", "Creature — Human Villain", "{1}{B}")
	mshA6OriginAtChapter(t, g, me, 2)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("chapter II offered nothing")
	}
	if !hasID(pick.ChooseCards, low) || hasID(pick.ChooseCards, high) || hasID(pick.ChooseCards, villain) {
		t.Errorf("candidates %v: want only the cheap Hero", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, low)
	if !g.Battlefield.Contains(low) {
		t.Error("the Hero did not enter")
	}
	if got := me.Hand.Size(); got != hand-1 {
		t.Errorf("hand %d, want %d: no draw when a card was put", got, hand-1)
	}
}

func TestMshA6OriginChapterTwoDrawsWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	low := mshA6HandCard(me, "Cheap Hero", "Creature — Human Hero", "{1}{W}")
	mshA6OriginAtChapter(t, g, me, 2)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	if g.Battlefield.Contains(low) {
		t.Error("a declined Hero entered")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: declining draws a card", got, hand+1)
	}
}

func TestMshA6OriginChapterTwoDrawsWithNoHeroToPut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mshA6HandCard(me, "Pricey Hero", "Creature — Human Hero", "{2}{W}{W}")
	mshA6OriginAtChapter(t, g, me, 2)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a prompt opened with no legal Hero in hand")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: no Hero to put draws a card", got, hand+1)
	}
}

func TestMshA6OriginChapterThreePutsACounterOnEachCreatureYouControlThenIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mine2 := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	saga := mshA6OriginAtChapter(t, g, me, 3)
	passPriorityAroundTable(t, g)
	if plusOnes(g, mine) != 1 || plusOnes(g, mine2) != 1 {
		t.Errorf("my creatures have %d and %d counters, want 1 each", plusOnes(g, mine), plusOnes(g, mine2))
	}
	if plusOnes(g, theirs) != 0 {
		t.Error("an opponent's creature got a counter")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- Multiversal Incursion -------------------------------------------------------------------

func TestMshA6MultiversalIncursionCopiesEachNontokenCreatureWithoutLegendary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := b12Creature(g, me.ID, "Lone Legend", "Legendary Creature — Human Hero", 3, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/1 black Villain with menace"), 1)
	})
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tokensBefore := 0
	for _, c := range mshA6Battlefield(g, "Villain") {
		if c.IsToken() {
			tokensBefore++
		}
	}
	castCatalogSpell(t, g, "Multiversal Incursion", "Sorcery", mshA6IncursionOracle, nil)
	passPriorityAroundTable(t, g)
	legends := mshA6Battlefield(g, "Lone Legend")
	if len(legends) != 2 {
		t.Fatalf("%d Lone Legends, want 2 (the original and a token copy)", len(legends))
	}
	for _, c := range legends {
		if c.IsToken() && (strings.Contains(c.TypeLine, "Legendary") || c.Power != 3) {
			t.Errorf("the token copy is %q %d/%d, want a non-legendary 3/3", c.TypeLine, c.Power, c.Toughness)
		}
	}
	if !g.Battlefield.Contains(legend) || !g.Battlefield.Contains(bear) {
		t.Error("an original left the battlefield")
	}
	if got := len(mshA6Battlefield(g, "My Bear")); got != 2 {
		t.Errorf("%d My Bears, want 2", got)
	}
	tokensAfter := 0
	for _, c := range mshA6Battlefield(g, "Villain") {
		if c.IsToken() {
			tokensAfter++
		}
	}
	if tokensAfter != tokensBefore {
		t.Errorf("a token creature was copied: %d -> %d", tokensBefore, tokensAfter)
	}
	if got := len(mshA6Battlefield(g, "Their Bear")); got != 1 {
		t.Errorf("an opponent's creature was copied: %d Their Bears", got)
	}
}

// --- Nick Fury, Agent of S.H.I.E.L.D. --------------------------------------------------------

const mshA6NickFuryType = "Legendary Creature — Human Spy Hero"

// mshA6NickFurySetup casts Nick Fury, and stacks the top seven of the
// library with a Hero, an Equipment, a Vehicle and four Bears, returning
// their ids.
func mshA6NickFurySetup(t *testing.T, g *game.Game, me *game.Player) (nick, hero, equip, vehicle uuid.UUID, bears []uuid.UUID) {
	t.Helper()
	powerUpMain(t, g)
	g.WithWriteLock(func() {
		for i := 0; i < 4; i++ {
			bears = append(bears, seedLibraryTop(me, "Library Bear", "Creature — Bear", "{1}{G}"))
		}
		vehicle = seedLibraryTop(me, "Library Vehicle", "Artifact — Vehicle", "{3}")
		equip = seedLibraryTop(me, "Library Equipment", "Artifact — Equipment", "{1}")
		hero = seedLibraryTop(me, "Library Hero", "Creature — Human Hero", "{2}{W}")
	})
	nick = castPowerUpCreature(t, g, me, "Nick Fury, Agent of S.H.I.E.L.D.", mshA6NickFuryType, "{W}", "{W}", mshA6NickFuryOracle, 2)
	return nick, hero, equip, vehicle, bears
}

func TestMshA6NickFuryPowerUpGrowsAndPutsAHeroEquipmentOrVehicleOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nick, hero, equip, vehicle, bears := mshA6NickFurySetup(t, g, me)
	if got := chargedPowerUp(t, g, me.ID, nick); got != "{U}{B}{R}{G}" {
		t.Fatalf("the turn he entered: %q, want {U}{B}{R}{G} (all five less his {W})", got)
	}
	libBefore := me.Library.Size()
	floatMana(t, g, me, "{U}{B}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, nick, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, nick); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick offered")
	}
	for _, id := range []uuid.UUID{hero, equip, vehicle} {
		if !hasID(pick.ChooseCards, id) {
			t.Errorf("%v was not offered", id)
		}
	}
	for _, id := range bears {
		if hasID(pick.ChooseCards, id) {
			t.Error("a Bear was offered")
		}
	}
	if pick.ChooseMax != 1 || pick.ChooseMin != 0 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, equip)
	if !g.Battlefield.Contains(equip) {
		t.Error("the Equipment did not enter")
	}
	if got := me.Library.Size(); got != libBefore-1 {
		t.Errorf("library %d, want %d", got, libBefore-1)
	}
	// The six left are the bottom six of the library.
	rest := map[uuid.UUID]bool{hero: true, vehicle: true}
	for _, id := range bears {
		rest[id] = true
	}
	for i := 0; i < 6; i++ {
		if !rest[me.Library.Cards[i].InstanceID] {
			t.Errorf("library card %d from the bottom is not one of the six left", i)
		}
	}
	// Used once only.
	floatMana(t, g, me, "{W}{U}{B}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, nick, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

func TestMshA6NickFuryMayPutNothingAndAllSevenGoToTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nick, hero, equip, vehicle, bears := mshA6NickFurySetup(t, g, me)
	floatMana(t, g, me, "{U}{B}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, nick, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	libBefore := me.Library.Size()
	answerChooseCards(t, g, me.ID)
	if got := me.Library.Size(); got != libBefore {
		t.Errorf("library %d, want %d", got, libBefore)
	}
	all := map[uuid.UUID]bool{hero: true, equip: true, vehicle: true}
	for _, id := range bears {
		all[id] = true
	}
	for i := 0; i < 7; i++ {
		if !all[me.Library.Cards[i].InstanceID] {
			t.Errorf("library card %d from the bottom is not one of the seven looked at", i)
		}
	}
}

func mshA6TransformingHero(p *game.Player) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Layout: game.LayoutTransform, Owner: p.ID, Controller: p.ID,
		Faces: []game.Face{
			{Name: "Day Hero", TypeLine: "Creature — Human Hero", ManaCost: "{2}{W}", Power: 2, Toughness: 2},
			{Name: "Night Hero", TypeLine: "Creature — Hero Spirit", Power: 4, Toughness: 4},
		},
	}
	c.SetFace(0)
	p.Library.PushTop(c)
	return c.InstanceID
}

func TestMshA6NickFuryMayTransformADoubleFacedCardThatEnters(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		nick, _, _, _, bears := mshA6NickFurySetup(t, g, me)
		var dfc uuid.UUID
		g.WithWriteLock(func() { dfc = mshA6TransformingHero(me) })
		floatMana(t, g, me, "{U}{B}{R}{G}")
		if err := g.ActivateCatalogAbility(me.ID, nick, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("power up: %v", err)
		}
		resolveTopOfStack(t, g)
		answerChooseCards(t, g, me.ID, dfc)
		answerConfirmFor(t, g, me, yes)
		if !g.Battlefield.Contains(dfc) {
			t.Fatalf("yes=%v: the double-faced card did not enter", yes)
		}
		wantFace := 0
		if yes {
			wantFace = 1
		}
		if got := b12Card(t, g, dfc).ActiveFace; got != wantFace {
			t.Errorf("yes=%v: active face %d, want %d", yes, got, wantFace)
		}
		// Either way the rest went to the bottom afterwards.
		if z := b12ZoneOf(g, bears[0]); z != game.ZoneLibrary {
			t.Errorf("yes=%v: a looked-at Bear is in %v", yes, z)
		}
		if got := me.Library.Cards[0].Name; got != "Library Bear" && got != "Library Vehicle" && got != "Library Equipment" {
			t.Errorf("yes=%v: the bottom card is %q, want one of the rest", yes, got)
		}
	}
}
