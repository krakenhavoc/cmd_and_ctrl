package testsms

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMshA6EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.MshA6MoonstoneOracle:   "Moonstone, Harsh Mistress",
		effectstest.MshA6IncursionOracle:   "Multiversal Incursion",
		effectstest.MshA6NickFuryOracle:    "Nick Fury, Agent of S.H.I.E.L.D.",
		effectstest.MshA6NinjaOracle:       "Ninja of the Hand",
		effectstest.MshA6OriginOracle:      "Origin of the Avengers",
		effectstest.MshA6PatriotOracle:     "Patriot, Shield Wielder",
		effectstest.MshA6BrokerOracle:      "Powerful Broker",
		effectstest.MshA6PunchOracle:       "Punishing Punch",
		effectstest.MshA6QuakeOracle:       "Quake, Agent of S.H.I.E.L.D.",
		effectstest.MshA6RaftOracle:        "Raft Security Officer",
		effectstest.MshA6RedGuardianOracle: "Red Guardian, Super-Soldier",
		effectstest.MshA6RecruitOracle:     "Red Room Recruit",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestMshA6KeywordsAreDeclared(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct{ name, typeLine, oracle, keyword string }{
		{"Moonstone, Harsh Mistress", "Legendary Creature — Human Doctor Villain", effectstest.MshA6MoonstoneOracle, "flying"},
		{"Ninja of the Hand", "Creature — Human Ninja Villain", effectstest.MshA6NinjaOracle, "deathtouch"},
		{"Red Guardian, Super-Soldier", "Legendary Creature — Human Soldier Villain", effectstest.MshA6RedGuardianOracle, "flash"},
	} {
		id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
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
	effectstest.AnswerDiscard(t, g, p.ID, card)
}

// --- Moonstone, Harsh Mistress -------------------------------------------------------------

const mshA6MoonstoneType = "Legendary Creature — Human Doctor Villain"

func TestMshA6MoonstoneExilesADiscardedCardAndLetsYouPlayItThroughYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Moonstone, Harsh Mistress", mshA6MoonstoneType, effectstest.MshA6MoonstoneOracle, false)
	effectstest.AdvanceToMain(t, g)
	card := effectstest.HandCardForTest(me, "Spare Spell", "Sorcery", "")
	mshA6Discard(t, g, me, card)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if z := effectstest.B12ZoneOf(g, card); z != game.ZoneExile {
		t.Fatalf("the discarded card is in %v, want exile", z)
	}
	if !effectstest.B12ExiledPlayableBy(t, g, card, me.ID) {
		t.Fatal("the exiled card is not playable by its discarder")
	}
	// Through the rest of this turn and all of the next one of mine.
	effectstest.PowerUpPassTo(t, g, 0)
	if !effectstest.B12ExiledPlayableBy(t, g, card, me.ID) {
		t.Error("the permission ended before the end of my next turn")
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if effectstest.B12ExiledPlayableBy(t, g, card, me.ID) {
		t.Error("the permission outlived my next turn")
	}
}

func TestMshA6MoonstoneDeclinedLeavesTheCardInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Moonstone, Harsh Mistress", mshA6MoonstoneType, effectstest.MshA6MoonstoneOracle, false)
	effectstest.AdvanceToMain(t, g)
	card := effectstest.HandCardForTest(me, "Spare Spell", "Sorcery", "")
	mshA6Discard(t, g, me, card)
	effectstest.AssistCardSettle(t, g, me.ID, nil, false)
	if z := effectstest.B12ZoneOf(g, card); z != game.ZoneGraveyard {
		t.Errorf("a declined card is in %v, want the graveyard", z)
	}
}

func TestMshA6MoonstoneIgnoresAnOpponentsDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Moonstone, Harsh Mistress", mshA6MoonstoneType, effectstest.MshA6MoonstoneOracle, false)
	effectstest.AdvanceToMain(t, g)
	card := effectstest.HandCardForTest(opp, "Their Spell", "Sorcery", "")
	mshA6Discard(t, g, opp, card)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if z := effectstest.B12ZoneOf(g, card); z != game.ZoneGraveyard {
		t.Errorf("an opponent's discarded card is in %v, want their graveyard", z)
	}
}

// --- Ninja of the Hand -----------------------------------------------------------------------

func TestMshA6NinjaPowerUpMakesEachOpponentDiscardAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	spare := effectstest.HandCardForTest(opp, "Their Spell", "Sorcery", "")
	ninja := effectstest.CastPowerUpCreature(t, g, me, "Ninja of the Hand", "Creature — Human Ninja Villain", "{2}{B}", "{C}{C}{B}", effectstest.MshA6NinjaOracle, 2)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, ninja); got != "{2}" {
		t.Fatalf("the turn it entered: %q, want {2} ({4}{B} less {2}{B})", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if effectstest.DiscardOwed(g, opp.ID) != 1 {
		t.Fatalf("opponent owes %d discards, want 1", effectstest.DiscardOwed(g, opp.ID))
	}
	effectstest.AnswerDiscard(t, g, opp.ID, spare)
	if got := effectstest.PlusOnes(g, ninja); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

// --- Red Room Recruit ------------------------------------------------------------------------

func TestMshA6RedRoomRecruitConnivesWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.HandCardForTest(me, "Spare Spell", "Sorcery", "")
	before := me.Hand.Size()
	id := effectstest.CastHoldCreature(t, g, "Red Room Recruit", "Creature — Human Spy Villain", effectstest.MshA6RecruitOracle, 1, 2)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand %d, want %d (the draw, before the discard)", got, before+1)
	}
	effectstest.AnswerDiscard(t, g, me.ID, spell)
	if n := effectstest.HobA8Counters(t, g, id); n != 1 {
		t.Errorf("%d +1/+1 counters after discarding a nonland card, want 1", n)
	}
}

func TestMshA6RedRoomRecruitGetsNoCounterForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.HandCardForTest(me, "Spare Forest", "Basic Land — Forest", "")
	id := effectstest.CastHoldCreature(t, g, "Red Room Recruit", "Creature — Human Spy Villain", effectstest.MshA6RecruitOracle, 1, 2)
	effectstest.AnswerDiscard(t, g, me.ID, forest)
	if n := effectstest.HobA8Counters(t, g, id); n != 0 {
		t.Errorf("%d +1/+1 counters after discarding a land, want 0", n)
	}
}

// --- Patriot, Shield Wielder -----------------------------------------------------------------

func TestMshA6PatriotPumpsAndShieldsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	patriot := effectstest.PushCatalogPermanent(g, me.ID, "Patriot, Shield Wielder", "Legendary Creature — Human Hero", effectstest.MshA6PatriotOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, patriot, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power = %d, want 4", got)
	}
	if !mshA3HasKeyword(g, bear, "hexproof") {
		t.Error("the target did not gain hexproof")
	}
	if !effectstest.B16Tapped(t, g, patriot) {
		t.Error("Patriot did not tap for the cost")
	}
}

func TestMshA6PatriotCannotTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	patriot := effectstest.PushCatalogPermanent(g, me.ID, "Patriot, Shield Wielder", "Legendary Creature — Human Hero", effectstest.MshA6PatriotOracle, false)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, patriot, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(patriot)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("targeting himself: err = %v, want ErrIllegalTarget", err)
	}
}

func TestMshA6PatriotCannotTargetAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	patriot := effectstest.PushCatalogPermanent(g, me.ID, "Patriot, Shield Wielder", "Legendary Creature — Human Hero", effectstest.MshA6PatriotOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, patriot, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); !errors.Is(err, game.ErrIllegalTarget) {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	officer := effectstest.PushCatalogPermanent(g, me.ID, "Raft Security Officer", "Creature — Human Soldier", effectstest.MshA6RaftOracle, false)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	if got := mshA6PriceOf(t, g, me.ID, officer, effectstest.CardRefs(small)); got != "{1}" {
		t.Errorf("against power 3: %q, want {1}", got)
	}
	if got := mshA6PriceOf(t, g, me.ID, officer, effectstest.CardRefs(big)); got != "{2}" {
		t.Errorf("against power 4: %q, want {2}", got)
	}
}

func TestMshA6RaftSecurityOfficerTapsTheTargetForTheReducedCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	officer := effectstest.PushCatalogPermanent(g, me.ID, "Raft Security Officer", "Creature — Human Soldier", effectstest.MshA6RaftOracle, false)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 3, 3)
	mshA6AddMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, officer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small), Strict: true}); err != nil {
		t.Fatalf("activate for {1}: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if !effectstest.B16Tapped(t, g, small) {
		t.Error("the target is untapped")
	}
}

func TestMshA6RaftSecurityOfficerCannotPayOneAgainstABigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	officer := effectstest.PushCatalogPermanent(g, me.ID, "Raft Security Officer", "Creature — Human Soldier", effectstest.MshA6RaftOracle, false)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	mshA6AddMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, officer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(big), Strict: true}); err == nil {
		t.Fatal("a power-4 target was tapped for {1}")
	}
}

// --- Quake, Agent of S.H.I.E.L.D. ---------------------------------------------------------------

func TestMshA6QuakeTapsATargetCreatureWhenYouCastANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Quake, Agent of S.H.I.E.L.D.", "Legendary Creature — Inhuman Spy Hero", effectstest.MshA6QuakeOracle, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastNoncreature(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("the targeted creature is untapped")
	}
}

func TestMshA6QuakeTapsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Quake, Agent of S.H.I.E.L.D.", "Legendary Creature — Inhuman Spy Hero", effectstest.MshA6QuakeOracle, false)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.CastNoncreature(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: land}, true)
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the targeted land is untapped")
	}
}

func TestMshA6QuakeIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Quake, Agent of S.H.I.E.L.D.", "Legendary Creature — Inhuman Spy Hero", effectstest.MshA6QuakeOracle, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Filler Bear", "Creature — Bear", "", nil)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if effectstest.B16Tapped(t, g, victim) {
		t.Error("a creature spell triggered Quake")
	}
}

// --- Punishing Punch -------------------------------------------------------------------------

func TestMshA6PunishingPunchDealsTwiceThePowerOneSidedly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 2, 6)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	if _, err := effectstest.CastFromHandErr(g, me, "Punishing Punch", "Instant", "{2}{G}", effectstest.MshA6PunchOracle, game.CastSpellParams{Targets: effectstest.CardRefs(mine, theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("a 2/6 survived 6 damage from a 3-power creature")
	}
	if c := effectstest.B12Card(t, g, mine); c.DamageMarked != 0 {
		t.Errorf("the source took %d damage back; the damage is one-sided", c.DamageMarked)
	}
}

func TestMshA6PunishingPunchCostsTwoLessWithTwoCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 2, 6)
	price := func() int {
		return effectstest.HobB3Price(t, g, me, "Punishing Punch", "Instant", "{2}{G}", mine, theirs)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	other := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{G}")
	if _, err := effectstest.CastFromHandErr(g, me, "Punishing Punch", "Instant", "{2}{G}", effectstest.MshA6PunchOracle, game.CastSpellParams{Targets: effectstest.CardRefs(mine, other)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("err = %v, want ErrIllegalTarget", err)
	}
}

// --- Red Guardian, Super-Soldier -------------------------------------------------------------

const mshA6RedGuardianType = "Legendary Creature — Human Soldier Villain"

func mshA6RedGuardianCard() game.Card {
	return game.Card{Name: "Red Guardian, Super-Soldier", TypeLine: mshA6RedGuardianType, OracleID: effectstest.MshA6RedGuardianOracle, Power: 2, Toughness: 2}
}

func TestMshA6RedGuardianDestroysAnOpponentsCreatureThatDealtDamageThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	biter := effectstest.B12Creature(g, opp.ID, "Their Biter", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Their Idler", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(biter, me.ID, 1) })
	effectstest.CastFromHand(t, g, mshA6RedGuardianCard(), game.CastSpellParams{})
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: biter}, true)
	if g.Battlefield.Contains(biter) {
		t.Error("the creature that dealt damage survived")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("a creature that dealt no damage was destroyed")
	}
}

func TestMshA6RedGuardianCountsDamageDealtToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	biter := effectstest.B12Creature(g, opp.ID, "Their Biter", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 5)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(biter, mine, 2) })
	effectstest.CastFromHand(t, g, mshA6RedGuardianCard(), game.CastSpellParams{})
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: biter}, true)
	if g.Battlefield.Contains(biter) {
		t.Error("the creature that damaged a creature survived")
	}
}

func TestMshA6RedGuardianNeedsADamageDealer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	idle := effectstest.B12Creature(g, opp.ID, "Their Idler", "Creature — Bear", 2, 2)
	effectstest.CastFromHand(t, g, mshA6RedGuardianCard(), game.CastSpellParams{})
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if !g.Battlefield.Contains(idle) {
		t.Error("a creature that dealt no damage was destroyed")
	}
}

func TestMshA6RedGuardianIgnoresYourOwnDamageDealer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Biter", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, g.Seats[1].ID, 1) })
	effectstest.CastFromHand(t, g, mshA6RedGuardianCard(), game.CastSpellParams{})
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was destroyed")
	}
}

// --- Powerful Broker -------------------------------------------------------------------------

func TestMshA6PowerfulBrokerGivesAnotherCounterOfEachKindOnAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	broker := effectstest.PushCatalogPermanent(g, me.ID, "Powerful Broker", "Creature — Human Villain", effectstest.MshA6BrokerOracle, false)
	target := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(target, game.CounterPlusOne, 3)
		_ = g.AddCounterForEffect(target, "charge", 2)
	})
	effectstest.B16Activate(t, g, me.ID, broker, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target)})
	if got := effectstest.PlusOnes(g, target); got != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", got)
	}
	if got := effectstest.CountersOn(g, target, "charge"); got != 3 {
		t.Errorf("charge counters = %d, want 3", got)
	}
}

func TestMshA6PowerfulBrokerGivesAPlayerAnotherCounterOfEachKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	broker := effectstest.PushCatalogPermanent(g, me.ID, "Powerful Broker", "Creature — Human Villain", effectstest.MshA6BrokerOracle, false)
	g.WithWriteLock(func() {
		_ = g.AddPlayerCounterForEffect(opp.ID, "poison", 2)
		_ = g.AddPlayerCounterForEffect(opp.ID, "energy", 1)
	})
	effectstest.B16Activate(t, g, me.ID, broker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if got := opp.Counters["poison"]; got != 3 {
		t.Errorf("poison counters = %d, want 3", got)
	}
	if got := opp.Counters["energy"]; got != 2 {
		t.Errorf("energy counters = %d, want 2", got)
	}
}

func TestMshA6PowerfulBrokerDoesNothingToAPermanentWithoutCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	broker := effectstest.PushCatalogPermanent(g, me.ID, "Powerful Broker", "Creature — Human Villain", effectstest.MshA6BrokerOracle, false)
	target := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, broker, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target)})
	if c := effectstest.B12Card(t, g, target); len(c.Counters) != 0 {
		t.Errorf("a counterless permanent got %v", c.Counters)
	}
}

func TestMshA6PowerfulBrokerIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	broker := effectstest.PushCatalogPermanent(g, me.ID, "Powerful Broker", "Creature — Human Villain", effectstest.MshA6BrokerOracle, false)
	target := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(target, game.CounterPlusOne, 1) })
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, broker, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Origin of the Avengers ------------------------------------------------------------------

const mshA6OriginType = "Enchantment — Saga"

// mshA6OriginAtChapter casts the Saga (chapter I scries; the scry is
// answered) and walks my turns until `chapter` is on the stack. Earlier
// chapters resolve on the way, so the hand must hold no legal Hero for
// chapter II when `chapter` is III.
func mshA6OriginAtChapter(t *testing.T, g *game.Game, me *game.Player, chapter int) uuid.UUID {
	t.Helper()
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	id := effectstest.CastCatalogSpell(t, g, "Origin of the Avengers", mshA6OriginType, effectstest.MshA6OriginOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("chapter I did not scry")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	for ch := 2; ch <= chapter; ch++ {
		effectstest.AdvanceToPrecombatMainOf(t, g, 0)
		if ch < chapter {
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	return id
}

func TestMshA6OriginChapterOneScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Origin of the Avengers", mshA6OriginType, effectstest.MshA6OriginOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("chapter I did not scry")
	}
	if len(scry.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(scry.ScryCards))
	}
}

func TestMshA6OriginChapterTwoPutsAHeroFromHandOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	low := effectstest.MshA6HandCard(me, "Cheap Hero", "Creature — Human Hero", "{1}{W}")
	high := effectstest.MshA6HandCard(me, "Pricey Hero", "Creature — Human Hero", "{2}{W}{W}")
	villain := effectstest.MshA6HandCard(me, "Cheap Villain", "Creature — Human Villain", "{1}{B}")
	mshA6OriginAtChapter(t, g, me, 2)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("chapter II offered nothing")
	}
	if !effectstest.HasID(pick.ChooseCards, low) || effectstest.HasID(pick.ChooseCards, high) || effectstest.HasID(pick.ChooseCards, villain) {
		t.Errorf("candidates %v: want only the cheap Hero", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, low)
	if !g.Battlefield.Contains(low) {
		t.Error("the Hero did not enter")
	}
	if got := me.Hand.Size(); got != hand-1 {
		t.Errorf("hand %d, want %d: no draw when a card was put", got, hand-1)
	}
}

func TestMshA6OriginChapterTwoDrawsWhenYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	low := effectstest.MshA6HandCard(me, "Cheap Hero", "Creature — Human Hero", "{1}{W}")
	mshA6OriginAtChapter(t, g, me, 2)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if g.Battlefield.Contains(low) {
		t.Error("a declined Hero entered")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: declining draws a card", got, hand+1)
	}
}

func TestMshA6OriginChapterTwoDrawsWithNoHeroToPut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.MshA6HandCard(me, "Pricey Hero", "Creature — Human Hero", "{2}{W}{W}")
	mshA6OriginAtChapter(t, g, me, 2)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a prompt opened with no legal Hero in hand")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: no Hero to put draws a card", got, hand+1)
	}
}

func TestMshA6OriginChapterThreePutsACounterOnEachCreatureYouControlThenIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mine2 := effectstest.B12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	saga := mshA6OriginAtChapter(t, g, me, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, mine) != 1 || effectstest.PlusOnes(g, mine2) != 1 {
		t.Errorf("my creatures have %d and %d counters, want 1 each", effectstest.PlusOnes(g, mine), effectstest.PlusOnes(g, mine2))
	}
	if effectstest.PlusOnes(g, theirs) != 0 {
		t.Error("an opponent's creature got a counter")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- Multiversal Incursion -------------------------------------------------------------------

func TestMshA6MultiversalIncursionCopiesEachNontokenCreatureWithoutLegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := effectstest.B12Creature(g, me.ID, "Lone Legend", "Legendary Creature — Human Hero", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/1 black Villain with menace"), 1)
	})
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tokensBefore := 0
	for _, c := range mshA6Battlefield(g, "Villain") {
		if c.IsToken() {
			tokensBefore++
		}
	}
	effectstest.CastCatalogSpell(t, g, "Multiversal Incursion", "Sorcery", effectstest.MshA6IncursionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PowerUpMain(t, g)
	g.WithWriteLock(func() {
		for i := 0; i < 4; i++ {
			bears = append(bears, effectstest.SeedLibraryTop(me, "Library Bear", "Creature — Bear", "{1}{G}"))
		}
		vehicle = effectstest.SeedLibraryTop(me, "Library Vehicle", "Artifact — Vehicle", "{3}")
		equip = effectstest.SeedLibraryTop(me, "Library Equipment", "Artifact — Equipment", "{1}")
		hero = effectstest.SeedLibraryTop(me, "Library Hero", "Creature — Human Hero", "{2}{W}")
	})
	nick = effectstest.CastPowerUpCreature(t, g, me, "Nick Fury, Agent of S.H.I.E.L.D.", mshA6NickFuryType, "{W}", "{W}", effectstest.MshA6NickFuryOracle, 2)
	return nick, hero, equip, vehicle, bears
}

func TestMshA6NickFuryPowerUpGrowsAndPutsAHeroEquipmentOrVehicleOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nick, hero, equip, vehicle, bears := mshA6NickFurySetup(t, g, me)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, nick); got != "{U}{B}{R}{G}" {
		t.Fatalf("the turn he entered: %q, want {U}{B}{R}{G} (all five less his {W})", got)
	}
	libBefore := me.Library.Size()
	effectstest.FloatMana(t, g, me, "{U}{B}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, nick, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, nick); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick offered")
	}
	for _, id := range []uuid.UUID{hero, equip, vehicle} {
		if !effectstest.HasID(pick.ChooseCards, id) {
			t.Errorf("%v was not offered", id)
		}
	}
	for _, id := range bears {
		if effectstest.HasID(pick.ChooseCards, id) {
			t.Error("a Bear was offered")
		}
	}
	if pick.ChooseMax != 1 || pick.ChooseMin != 0 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, equip)
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
	effectstest.FloatMana(t, g, me, "{W}{U}{B}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, nick, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

func TestMshA6NickFuryMayPutNothingAndAllSevenGoToTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nick, hero, equip, vehicle, bears := mshA6NickFurySetup(t, g, me)
	effectstest.FloatMana(t, g, me, "{U}{B}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, nick, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	libBefore := me.Library.Size()
	effectstest.AnswerChooseCards(t, g, me.ID)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		nick, _, _, _, bears := mshA6NickFurySetup(t, g, me)
		var dfc uuid.UUID
		g.WithWriteLock(func() { dfc = mshA6TransformingHero(me) })
		effectstest.FloatMana(t, g, me, "{U}{B}{R}{G}")
		if err := g.ActivateCatalogAbility(me.ID, nick, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("power up: %v", err)
		}
		effectstest.ResolveTopOfStack(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, dfc)
		effectstest.AnswerConfirmFor(t, g, me, yes)
		if !g.Battlefield.Contains(dfc) {
			t.Fatalf("yes=%v: the double-faced card did not enter", yes)
		}
		wantFace := 0
		if yes {
			wantFace = 1
		}
		if got := effectstest.B12Card(t, g, dfc).ActiveFace; got != wantFace {
			t.Errorf("yes=%v: active face %d, want %d", yes, got, wantFace)
		}
		// Either way the rest went to the bottom afterwards.
		if z := effectstest.B12ZoneOf(g, bears[0]); z != game.ZoneLibrary {
			t.Errorf("yes=%v: a looked-at Bear is in %v", yes, z)
		}
		if got := me.Library.Cards[0].Name; got != "Library Bear" && got != "Library Vehicle" && got != "Library Equipment" {
			t.Errorf("yes=%v: the bottom card is %q, want one of the rest", yes, got)
		}
	}
}
