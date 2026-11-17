package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b09_cards_test.go — Dominaria United, S82 batch 9 (#5487): the
// kicker spells Choking Miasma, Fires of Victory, Joint Exploration,
// Protect the Negotiators, Stall for Time, Timely Interference, Tribute
// to Urborg and Urborg Repossession (CR 702.33).

const (
	dmuChokingMiasmaOracle    = "158aaeab-1646-4160-9615-bc8747620a03"
	dmuFiresOfVictoryOracle   = "d6afb1a9-c9ed-4a56-ba7d-700ab71f2e0d"
	dmuJointExplorationOracle = "9eba1789-6e58-4d78-8a0d-b54a8be76ec5"
	dmuProtectNegotiators     = "7ccd666a-0d56-4fe3-b16b-d035bf2ad420"
	dmuStallForTimeOracle     = "09a400bd-83f1-4fac-a218-64c9f041874f"
	dmuTimelyInterference     = "d64feb23-8259-4445-9b40-efc8b4466679"
	dmuTributeToUrborgOracle  = "8355d44d-cf03-4b19-b2dc-0ca3450dbf5c"
	dmuUrborgRepossession     = "0c0eeb01-62e7-4a7f-923d-a7b830e4c83b"
)

func TestDmuB09CardsAreRegisteredFullWithTheirKickers(t *testing.T) {
	for oracle, tc := range map[string]struct{ name, kicker string }{
		dmuChokingMiasmaOracle:    {"Choking Miasma", "{G}"},
		dmuFiresOfVictoryOracle:   {"Fires of Victory", "{2}{U}"},
		dmuJointExplorationOracle: {"Joint Exploration", "{G}"},
		dmuProtectNegotiators:     {"Protect the Negotiators", "{W}"},
		dmuStallForTimeOracle:     {"Stall for Time", "{1}{U}"},
		dmuTimelyInterference:     {"Timely Interference", "{1}{R}"},
		dmuTributeToUrborgOracle:  {"Tribute to Urborg", "{1}{U}"},
		dmuUrborgRepossession:     {"Urborg Repossession", "{1}{G}"},
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", tc.name)
			continue
		}
		if spec.Name != tc.name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", tc.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		oc := game.OptionalCostsFor(oracle)
		if len(oc) != 1 || oc[0].Key != game.KickerKey || oc[0].ManaCost != tc.kicker {
			t.Errorf("%s: optional costs %+v, want one kicker %s", tc.name, oc, tc.kicker)
		}
	}
}

// dmuEmptyHand clears a seat's opening hand so hand-size reads are exact.
func dmuEmptyHand(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { p.Hand.Cards = nil })
}

// dmuCast casts a catalog spell for free from the active seat's hand in a
// main phase, kicked or not.
func dmuCast(t *testing.T, g *game.Game, name, typeLine, oracle string, kicked bool, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := handCardForTest(me, name, typeLine, oracle)
	params := game.CastSpellParams{Targets: targets}
	if kicked {
		params.OptionalCosts = []int{0}
	}
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast %s (kicked %v): %v", name, kicked, err)
	}
	return id
}

// dmuCastPaid casts strictly with exactly the mana in the pool.
func dmuCastPaid(g *game.Game, me *game.Player, name, typeLine, cost, oracle string, kicked bool, targets []game.TargetRef) error {
	id := handCardOf(g, me, name, typeLine, cost, oracle)
	params := game.CastSpellParams{Strict: true, Targets: targets}
	if kicked {
		params.OptionalCosts = []int{0}
	}
	return g.CastSpell(me.ID, id, params)
}

// --- Choking Miasma ---------------------------------------------------

func TestDmuB09ChokingMiasmaUnkickedShrinksEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	dmuCast(t, g, "Choking Miasma", "Sorcery", dmuChokingMiasmaOracle, false, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("my creature is %d/%d, want 1/1", p, tough)
	}
	if findBattlefieldCardByID(g, theirs) != nil {
		t.Error("their 2/2 survived the -2/-2")
	}
	if pendingOfKind(g, game.PendingChoiceOwnPermanents) != nil {
		t.Error("an unkicked cast asked for a counter")
	}
}

func TestDmuB09ChokingMiasmaKickedCountersFirstThenShrinks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	dmuCast(t, g, "Choking Miasma", "Sorcery", dmuChokingMiasmaOracle, true, nil)
	passPriorityAroundTable(t, g)
	own := pendingOfKind(g, game.PendingChoiceOwnPermanents)
	if own == nil {
		t.Fatal("the kicked cast did not ask which creature gets the counter")
	}
	if err := g.ResolveOwnPermanents(own.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("choose creature: %v", err)
	}
	g.SettleResolution()
	if findBattlefieldCardByID(g, mine) == nil {
		t.Fatal("my 2/2 with a +1/+1 counter died to -2/-2: the counter must land first")
	}
	if got := countersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if p, tough := ptOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("my creature is %d/%d, want 1/1", p, tough)
	}
	if findBattlefieldCardByID(g, theirs) != nil {
		t.Error("their 2/2 survived the -2/-2")
	}
}

func TestDmuB09ChokingMiasmaKickedWithNoCreatureStillShrinksNothing(t *testing.T) {
	g := newCatalogGame(t)
	dmuCast(t, g, "Choking Miasma", "Sorcery", dmuChokingMiasmaOracle, true, nil)
	passPriorityAroundTable(t, g)
	if pendingOfKind(g, game.PendingChoiceOwnPermanents) != nil {
		t.Error("asked for a creature with none on the battlefield")
	}
}

func TestDmuB09ChokingMiasmaKickerNeedsItsGreenMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b06AddMana(me, "B", "B", "C")
	if err := dmuCastPaid(g, me, "Choking Miasma", "Sorcery", "{1}{B}{B}", dmuChokingMiasmaOracle, true, nil); err == nil {
		t.Error("a kicked cast was accepted without {G}")
	}
	if err := dmuCastPaid(g, me, "Choking Miasma", "Sorcery", "{1}{B}{B}", dmuChokingMiasmaOracle, false, nil); err != nil {
		t.Errorf("the unkicked cast was refused: %v", err)
	}
}

// --- Fires of Victory -------------------------------------------------

func TestDmuB09FiresOfVictoryDealsDamageEqualToYourHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dmuEmptyHand(g, me)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Beast", 9, 9)
	for i := 0; i < 3; i++ {
		handCardForTest(me, "Filler", "Sorcery", "")
	}
	dmuCast(t, g, "Fires of Victory", "Instant", dmuFiresOfVictoryOracle, false, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 3 {
		t.Errorf("damage = %d, want 3 (the hand, not counting the spell)", got)
	}
}

func TestDmuB09FiresOfVictoryKickedDrawsFirstSoTheCardCounts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dmuEmptyHand(g, me)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Beast", 9, 9)
	handCardForTest(me, "Filler", "Sorcery", "")
	seedLibrary(me, "Drawn")
	dmuCast(t, g, "Fires of Victory", "Instant", dmuFiresOfVictoryOracle, true, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 2 {
		t.Errorf("damage = %d, want 2 (filler plus the drawn card)", got)
	}
	if me.Hand.Size() != 2 {
		t.Errorf("hand = %d, want 2", me.Hand.Size())
	}
}

func TestDmuB09FiresOfVictoryRefusesAnIllegalTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	advanceToMain(t, g)
	b06AddMana(me, "R", "C")
	if err := dmuCastPaid(g, me, "Fires of Victory", "Instant", "{1}{R}", dmuFiresOfVictoryOracle, false, cardTarget(land)); err == nil {
		t.Error("Fires of Victory targeted a land")
	}
}

// --- Joint Exploration ------------------------------------------------

func TestDmuB09JointExplorationScriesThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dmuEmptyHand(g, me)
	seedLibrary(me, "One", "Two", "Three")
	dmuCast(t, g, "Joint Exploration", "Instant", dmuJointExplorationOracle, false, nil)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	g.SettleResolution()
	if me.Hand.Size() != 1 {
		t.Errorf("hand = %d, want 1 drawn card", me.Hand.Size())
	}
	if pendingOfKind(g, game.PendingChoiceChooseCards) != nil {
		t.Error("an unkicked cast offered a land")
	}
}

func TestDmuB09JointExplorationKickedMayPutALandFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dmuEmptyHand(g, me)
	seedLibrary(me, "One", "Two", "Three")
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	rock := handCardForTest(me, "Sol Ring", "Artifact", "")
	dmuCast(t, g, "Joint Exploration", "Instant", dmuJointExplorationOracle, true, nil)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	g.SettleResolution()
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the kicked cast did not offer a land from hand")
	}
	for _, id := range pick.ChooseCards {
		if id == rock {
			t.Error("a non-land was offered")
		}
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not enter the battlefield")
	}
	if c, ok := battlefieldCard(g, land); ok && c.Tapped {
		t.Error("the land entered tapped")
	}
}

// --- Protect the Negotiators -------------------------------------------

func TestDmuB09ProtectTheNegotiatorsTaxesOneForEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	dmuCast(t, g, "Protect the Negotiators", "Instant", dmuProtectNegotiators, false, cardTarget(victim))
	passPriorityAroundTable(t, g)
	c := pendingOfKind(g, game.PendingChoicePayUnless)
	if c == nil {
		t.Fatalf("no pay-unless prompt: %+v", g.PendingChoices)
	}
	if c.PayCost != "{2}" {
		t.Errorf("the prompt asks for %q, want {2}", c.PayCost)
	}
}

func TestDmuB09ProtectTheNegotiatorsKickedMakesASoldierThatCounts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	dmuCast(t, g, "Protect the Negotiators", "Instant", dmuProtectNegotiators, true, cardTarget(victim))
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("Soldier tokens = %d, want 1", n)
	}
	c := pendingOfKind(g, game.PendingChoicePayUnless)
	if c == nil || c.PayCost != "{2}" {
		t.Fatalf("pay-unless = %+v, want {2}: the creature and the new token", c)
	}
}

func TestDmuB09ProtectTheNegotiatorsMakesNoTokenUnkicked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	dmuCast(t, g, "Protect the Negotiators", "Instant", dmuProtectNegotiators, false, cardTarget(victim))
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Soldier"); n != 0 {
		t.Errorf("Soldier tokens = %d, want 0", n)
	}
}

func TestDmuB09ProtectTheNegotiatorsNeedsASpellTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "U", "C")
	if err := dmuCastPaid(g, me, "Protect the Negotiators", "Instant", "{1}{U}", dmuProtectNegotiators, false, cardTarget(bear)); err == nil {
		t.Error("Protect the Negotiators targeted a creature")
	}
}

// --- Stall for Time -----------------------------------------------------

func TestDmuB09StallForTimeTapsUpToTwoAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dmuEmptyHand(g, me)
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	seedLibrary(me, "Drawn")
	dmuCast(t, g, "Stall for Time", "Instant", dmuStallForTimeOracle, false, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		tapped, stun := championsOfTheShoalStunned(t, g, id)
		if !tapped || stun != 0 {
			t.Errorf("creature tapped %v stun %d, want tapped with no counter", tapped, stun)
		}
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand = %d, want 1", me.Hand.Size())
	}
}

func TestDmuB09StallForTimeKickedStunsEachOfThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dmuEmptyHand(g, me)
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	seedLibrary(me, "Drawn")
	dmuCast(t, g, "Stall for Time", "Instant", dmuStallForTimeOracle, true, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		tapped, stun := championsOfTheShoalStunned(t, g, id)
		if !tapped || stun != 1 {
			t.Errorf("creature tapped %v stun %d, want tapped with one stun counter", tapped, stun)
		}
	}
	if tapped, stun := championsOfTheShoalStunned(t, g, c); tapped || stun != 0 {
		t.Errorf("the bystander is tapped %v stun %d", tapped, stun)
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand = %d, want 1", me.Hand.Size())
	}
}

func TestDmuB09StallForTimeWithNoTargetStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dmuEmptyHand(g, me)
	seedLibrary(me, "Drawn")
	dmuCast(t, g, "Stall for Time", "Instant", dmuStallForTimeOracle, true, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 {
		t.Errorf("hand = %d, want 1", me.Hand.Size())
	}
}

func TestDmuB09StallForTimeRefusesThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := []uuid.UUID{
		b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2),
		b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2),
		b12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2),
	}
	advanceToMain(t, g)
	b06AddMana(me, "W", "C", "C")
	if err := dmuCastPaid(g, me, "Stall for Time", "Instant", "{2}{W}", dmuStallForTimeOracle, false, cardRefs(ids...)); err == nil {
		t.Error("Stall for Time accepted three targets")
	}
}

// --- Timely Interference --------------------------------------------------

func TestDmuB09TimelyInterferenceShrinksPowerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dmuEmptyHand(g, me)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	seedLibrary(me, "Drawn")
	dmuCast(t, g, "Timely Interference", "Instant", dmuTimelyInterference, false, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 3 {
		t.Errorf("creature is %d/%d, want 2/3", p, tough)
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand = %d, want 1", me.Hand.Size())
	}
	if mkmB1MustBlock(t, g, theirs) {
		t.Error("an unkicked cast made the creature block")
	}
}

func TestDmuB09TimelyInterferenceKickedForcesABlock(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	dmuCast(t, g, "Timely Interference", "Instant", dmuTimelyInterference, true, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, theirs); p != 2 {
		t.Errorf("power = %d, want 2", p)
	}
	if !mkmB1MustBlock(t, g, theirs) {
		t.Error("the kicked cast did not require the creature to block")
	}
}

// --- Tribute to Urborg -------------------------------------------------------

func TestDmuB09TributeToUrborgUnkickedIsMinusTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushGraveyardCardTyped(me, "Spell", "Instant")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 5, 5)
	dmuCast(t, g, "Tribute to Urborg", "Instant", dmuTributeToUrborgOracle, false, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("creature is %d/%d, want 3/3", p, tough)
	}
}

func TestDmuB09TributeToUrborgKickedCountsInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushGraveyardCardTyped(me, "Spell", "Instant")
	pushGraveyardCardTyped(me, "Other Spell", "Sorcery")
	pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	pushGraveyardCardTyped(opp, "Their Spell", "Instant")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 9, 9)
	dmuCast(t, g, "Tribute to Urborg", "Instant", dmuTributeToUrborgOracle, true, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, theirs); p != 5 || tough != 5 {
		t.Errorf("creature is %d/%d, want 5/5 (-2/-2, then -1/-1 for each of my two spells)", p, tough)
	}
}

func TestDmuB09TributeToUrborgKickedKillsWhenTheGraveyardIsDeepEnough(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushGraveyardCardTyped(me, "Spell", "Instant")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	dmuCast(t, g, "Tribute to Urborg", "Instant", dmuTributeToUrborgOracle, true, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, theirs) != nil {
		t.Error("a 3/3 survived -3/-3")
	}
}

// --- Urborg Repossession -----------------------------------------------------

func TestDmuB09UrborgRepossessionReturnsTheCreatureAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	land := pushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	life := lifeOf(g, me.ID)
	dmuCast(t, g, "Urborg Repossession", "Sorcery", dmuUrborgRepossession, false, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || me.Hand.Contains(land) {
		t.Errorf("hand has bear %v land %v, want the bear only", me.Hand.Contains(bear), me.Hand.Contains(land))
	}
	if got := lifeOf(g, me.ID); got != life+2 {
		t.Errorf("life = %d, want %d", got, life+2)
	}
}

func TestDmuB09UrborgRepossessionKickedReturnsAnotherPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	land := pushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	spell := pushGraveyardCardTyped(me, "Bolt", "Instant")
	life := lifeOf(g, me.ID)
	dmuCast(t, g, "Urborg Repossession", "Sorcery", dmuUrborgRepossession, true, cardRefs(bear, land))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(land) || me.Hand.Contains(spell) {
		t.Errorf("hand has bear %v land %v spell %v", me.Hand.Contains(bear), me.Hand.Contains(land), me.Hand.Contains(spell))
	}
	if got := lifeOf(g, me.ID); got != life+2 {
		t.Errorf("life = %d, want %d", got, life+2)
	}
}

func TestDmuB09UrborgRepossessionRefusesBadKickedTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	spell := pushGraveyardCardTyped(me, "Bolt", "Instant")
	advanceToMain(t, g)
	b06AddMana(me, "B", "G", "C")
	if err := dmuCastPaid(g, me, "Urborg Repossession", "Sorcery", "{B}", dmuUrborgRepossession, true, cardRefs(bear, spell)); err == nil {
		t.Error("the second target was an instant card")
	}
	if err := dmuCastPaid(g, me, "Urborg Repossession", "Sorcery", "{B}", dmuUrborgRepossession, true, cardRefs(bear, bear)); err == nil {
		t.Error("the same card was named twice")
	}
	if err := dmuCastPaid(g, me, "Urborg Repossession", "Sorcery", "{B}", dmuUrborgRepossession, false, cardRefs(bear, bear)); err == nil {
		t.Error("an unkicked cast named two cards")
	}
}
