package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_ruz2_cards_test.go — the second batch of Bloomburrow cards starting
// with R or U–Z (#4467). Each test plays the real catalog entry and
// asserts what its text says.

const (
	blbRUZ2RecruiterOracle   = "a12a66b5-ea4a-4a58-b0f0-13152598b7e6"
	blbRUZ2NegotiationOracle = "a82ea8f2-e576-46cd-aa08-9c98d6029db4"
	blbRUZ2QuestcallerOracle = "81c509b2-3164-47e9-94b3-bdba0487d2f1"
	blbRUZ2WarleaderOracle   = "a7e7059e-83f8-4124-9cc0-64ba150b5d96"
	blbRUZ2PatrolOracle      = "29423a9a-31a6-4605-8b49-a75895038a3a"
	blbRUZ2ZoralineOracle    = "5a775f50-20f8-4a8f-868c-8907305bfad6"
)

func TestBlbRUZ2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbRUZ2RecruiterOracle:   "Reptilian Recruiter",
		blbRUZ2NegotiationOracle: "Ruthless Negotiation",
		blbRUZ2QuestcallerOracle: "Valley Questcaller",
		blbRUZ2WarleaderOracle:   "Warren Warleader",
		blbRUZ2PatrolOracle:      "Wick's Patrol",
		blbRUZ2ZoralineOracle:    "Zoraline, Cosmos Caller",
	} {
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

// --- Reptilian Recruiter ------------------------------------------------------

// blbRUZ2RecruiterSteals casts the Recruiter, aims its entry trigger at
// target and resolves it.
func blbRUZ2RecruiterSteals(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	recruiter := blbGLCastCreature(t, g, "Reptilian Recruiter", "Creature — Lizard Warrior", blbRUZ2RecruiterOracle, 4, 2, nil)
	answerPickTarget(t, g, target)
	passPriorityAroundTable(t, g)
	return recruiter
}

func TestBlbRUZ2ReptilianRecruiterStealsASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	small := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	findBattlefieldCardForTest(g, small).Tapped = true
	recruiter := blbRUZ2RecruiterSteals(t, g, small)
	if !hasKeywordOnBattlefield(t, g, recruiter, "trample") {
		t.Error("the Recruiter lacks trample")
	}
	c, ok := battlefieldCard(g, small)
	if !ok || c.Controller != me.ID {
		t.Fatal("a power-2 creature should change control")
	}
	if c.Tapped {
		t.Error("the stolen creature should be untapped")
	}
	if !hasEffectiveKeyword(t, g, small, "haste") {
		t.Error("the stolen creature should have haste")
	}
	endTurn(t, g)
	if controllerOf(t, g, small) != opp.ID {
		t.Error("control should return at end of turn")
	}
}

func TestBlbRUZ2ReptilianRecruiterNeedsAnotherLizardForABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	big := blbB1Push(g, opp.ID, "Their Ox", "Creature — Ox", "", 3, 3)
	blbRUZ2RecruiterSteals(t, g, big)
	if controllerOf(t, g, big) != opp.ID {
		t.Error("power 3 and no other Lizard: the creature should not change control")
	}

	g = newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Lizard", "Creature — Lizard", "", 1, 1)
	theirLizard := blbB1Push(g, opp.ID, "Their Lizard", "Creature — Lizard", "", 1, 1)
	big = blbB1Push(g, opp.ID, "Their Ox", "Creature — Ox", "", 3, 3)
	blbRUZ2RecruiterSteals(t, g, big)
	if controllerOf(t, g, big) != me.ID {
		t.Error("with another Lizard you control, the power-3 creature should change control")
	}
	if controllerOf(t, g, theirLizard) != opp.ID {
		t.Error("an untargeted creature changed control")
	}
}

func TestBlbRUZ2ReptilianRecruiterIgnoresAnOpponentsLizard(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	blbB1Push(g, opp.ID, "Their Lizard", "Creature — Lizard", "", 1, 1)
	big := blbB1Push(g, opp.ID, "Their Ox", "Creature — Ox", "", 3, 3)
	blbRUZ2RecruiterSteals(t, g, big)
	if controllerOf(t, g, big) != opp.ID {
		t.Error("a Lizard an opponent controls is not one you control")
	}
}

// --- Ruthless Negotiation -----------------------------------------------------

func TestBlbRUZ2RuthlessNegotiationExilesTheOpponentsPick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B")
	keep := handCardForTest(opp, "Keep", "Instant", "")
	lose := handCardForTest(opp, "Lose", "Sorcery", "")
	castCatalogSpell(t, g, "Ruthless Negotiation", "Sorcery", blbRUZ2NegotiationOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the target opponent should be asked which card to exile")
	}
	if pick.ChooseMin != 1 {
		t.Errorf("the opponent must exile a card: min %d", pick.ChooseMin)
	}
	answerChooseCards(t, g, opp.ID, lose)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(keep) || opp.Hand.Contains(lose) || !g.Exile.Contains(lose) {
		t.Error("the picked card should be exiled and the other kept")
	}
	if me.Hand.Size() != hand {
		t.Errorf("cast from hand: my hand is %d, want %d (no draw)", me.Hand.Size(), hand)
	}
}

func TestBlbRUZ2RuthlessNegotiationFlashbackAlsoDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B")
	lose := handCardForTest(opp, "Lose", "Sorcery", "")
	id := seedGraveyardCard(t, g, "Ruthless Negotiation", "Sorcery", blbRUZ2NegotiationOracle)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, lose)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(lose) {
		t.Error("the opponent's card should be exiled")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("cast from a graveyard: my hand is %d, want %d", me.Hand.Size(), hand+1)
	}
	if me.Graveyard.Contains(id) || !g.Exile.Contains(id) {
		t.Error("the flashed-back card should be exiled")
	}
}

func TestBlbRUZ2RuthlessNegotiationTargetsOnlyAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	if err := castCatalogSpellErr(t, g, "Ruthless Negotiation", "Sorcery", blbRUZ2NegotiationOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Error("you are not a legal target")
	}
}

// --- Valley Questcaller -------------------------------------------------------

func TestBlbRUZ2ValleyQuestcallerPumpsItsFourTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	questcaller := blbB1Push(g, me.ID, "Valley Questcaller", "Creature — Rabbit Warrior", blbRUZ2QuestcallerOracle, 2, 3)
	rabbit := blbB1Push(g, me.ID, "Rabbit", "Creature — Rabbit", "", 1, 1)
	mouse := blbB1Push(g, me.ID, "Mouse", "Creature — Mouse Soldier", "", 1, 1)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirBat := blbB1Push(g, opp.ID, "Their Bat", "Creature — Bat", "", 1, 1)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		p, t int
	}{
		{"the Questcaller itself", questcaller, 2, 3},
		{"my Rabbit", rabbit, 2, 2},
		{"my Mouse", mouse, 2, 2},
		{"my Bear", bear, 2, 2},
		{"an opponent's Bat", theirBat, 1, 1},
	} {
		if p, tough := ptOf(t, g, c.id); p != c.p || tough != c.t {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.p, c.t)
		}
	}
}

func TestBlbRUZ2ValleyQuestcallerScriesWhenItsKinEnter(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	blbB1Push(g, me.ID, "Valley Questcaller", "Creature — Rabbit Warrior", blbRUZ2QuestcallerOracle, 2, 3)
	blbGLCastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2, nil)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Fatal("a Bear entering should not scry")
	}
	blbGLCastCreature(t, g, "Bird", "Creature — Bird", "", 1, 1, nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("a Bird entering should scry 1, got %+v", c)
	}
	answerScryKeepAll(t, g, me.ID)
}

// --- Warren Warleader ---------------------------------------------------------

// blbRUZ2WarleaderAttacks attacks with the Warleader and a Bear and
// chooses the trigger's mode.
func blbRUZ2WarleaderAttacks(t *testing.T, g *game.Game, mode int) (warleader, bear uuid.UUID) {
	t.Helper()
	me, opp, _ := blbB1Seats(g)
	warleader = blbB1Push(g, me.ID, "Warren Warleader", "Creature — Rabbit Knight", blbRUZ2WarleaderOracle, 4, 4)
	bear = blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, warleader, bear)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	return warleader, bear
}

func TestBlbRUZ2WarrenWarleaderPumpsAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	idle := blbB1Push(g, me.ID, "Idle Bear", "Creature — Bear", "", 2, 2)
	warleader, bear := blbRUZ2WarleaderAttacks(t, g, 1)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, warleader); p != 5 || tough != 5 {
		t.Errorf("the attacking Warleader is %d/%d, want 5/5", p, tough)
	}
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the attacking Bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, idle); p != 2 || tough != 2 {
		t.Errorf("a creature that did not attack is %d/%d, want 2/2", p, tough)
	}
}

func TestBlbRUZ2WarrenWarleaderMakesAnAttackingRabbit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbRUZ2WarleaderAttacks(t, g, 0)
	passAndSettleAttackSplit(t, g, me.ID)
	tokens := battlefieldIDsNamed(g, "Rabbit")
	if len(tokens) != 1 {
		t.Fatalf("%d Rabbit tokens, want 1", len(tokens))
	}
	tok := cardByID(g, tokens[0])
	if !tok.Tapped || tok.AttackingTarget != opp.ID || tok.Controller != me.ID {
		t.Errorf("want a tapped Rabbit attacking the defender: %+v", tok)
	}
	if p, tough := ptOf(t, g, tokens[0]); p != 1 || tough != 1 {
		t.Errorf("the Rabbit is %d/%d, want 1/1", p, tough)
	}
}

func TestBlbRUZ2WarrenWarleaderHasOffspring(t *testing.T) {
	spec, _ := Lookup(blbRUZ2WarleaderOracle)
	if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != OffspringKey || spec.OptionalCosts[0].ManaCost != "{2}" {
		t.Fatalf("optional costs %+v, want Offspring {2}", spec.OptionalCosts)
	}
}

// --- Wick's Patrol ------------------------------------------------------------

func TestBlbRUZ2WicksPatrolShrinksByTheGreatestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibraryOfCost(g, me, 4, "{1}{B}")
	big := game.NewCard("Big Spell", me.ID)
	big.TypeLine = "Sorcery"
	big.ManaCost = "{3}{B}{B}"
	me.Graveyard.PushTop(big)
	mine := blbB1Push(g, me.ID, "My Ox", "Creature — Ox", "", 6, 6)
	theirs := blbB1Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 6, 6)
	blbGLCastCreature(t, g, "Wick's Patrol", "Creature — Rat Warlock", blbRUZ2PatrolOracle, 5, 3, nil)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 4 {
		t.Fatalf("the graveyard holds %d cards, want 4 (three milled)", got)
	}
	pick := pickTargetPrompt(t, g)
	for _, id := range pick.PickTargetCards {
		if id == mine {
			t.Fatal("a creature you control is not a legal target")
		}
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1 (-5/-5)", p, tough)
	}
	if p, tough := ptOf(t, g, mine); p != 6 || tough != 6 {
		t.Errorf("my creature is %d/%d, want 6/6", p, tough)
	}
}

func TestBlbRUZ2WicksPatrolNeedsThreeCardsToMill(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibraryOfCost(g, me, 2, "{5}")
	blbB1Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 6, 6)
	blbGLCastCreature(t, g, "Wick's Patrol", "Creature — Rat Warlock", blbRUZ2PatrolOracle, 5, 3, nil)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 2 {
		t.Errorf("the graveyard holds %d cards, want the 2 the library had", got)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("fewer than three cards milled: the reflexive trigger should not happen")
	}
}

// --- Zoraline, Cosmos Caller --------------------------------------------------

// blbRUZ2Settle answers trigger order, Zoraline's offer (paying when pay)
// and the return target, and passes priority until the stack is empty.
func blbRUZ2Settle(t *testing.T, g *game.Game, me *game.Player, pay bool, target uuid.UUID) {
	t.Helper()
	for i := 0; i < 12; i++ {
		blbGLSettleTriggerOrder(g, me.ID)
		switch {
		case hasPayUnlessFor(g, me.ID):
			if pay {
				b06AddMana(me, "W", "B")
			}
			answerPayUnless(t, g, me.ID, pay)
		case latestPickTarget(g, me.ID) != nil:
			answerPickTarget(t, g, target)
		case stackFullyEmpty(g):
			return
		default:
			passPriorityAroundTable(t, g)
		}
	}
	t.Fatal("the stack did not settle")
}

func TestBlbRUZ2ZoralineGainsPerBatAndReturnsWithFinality(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	zoraline := blbB1Push(g, me.ID, "Zoraline, Cosmos Caller", "Legendary Creature — Bat Cleric", blbRUZ2ZoralineOracle, 3, 3)
	bat := blbB1Push(g, me.ID, "Bat", "Creature — Bat", "", 1, 1)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasKeywordOnBattlefield(t, g, zoraline, kw) {
			t.Errorf("Zoraline lacks %s", kw)
		}
	}
	small := game.NewCard("Small Relic", me.ID)
	small.TypeLine = "Artifact"
	small.ManaCost = "{2}{B}"
	me.Graveyard.PushTop(small)
	life := lifeOf(g, me.ID)
	declareAttack(t, g, opp.ID, zoraline, bat, bear)
	blbRUZ2Settle(t, g, me, true, small.InstanceID)
	if got := lifeOf(g, me.ID); got != life+2-2 {
		t.Errorf("life %d, want %d (two Bats attacked, 2 life paid)", got, life)
	}
	if !g.Battlefield.Contains(small.InstanceID) {
		t.Fatal("the mana value 3 card should return")
	}
	if n := countersOn(g, small.InstanceID, game.CounterFinality); n != 1 {
		t.Errorf("%d finality counters, want 1", n)
	}
}

func TestBlbRUZ2ZoralineOffersOnlySmallNonlandPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	zoraline := blbB1Push(g, me.ID, "Zoraline, Cosmos Caller", "Legendary Creature — Bat Cleric", blbRUZ2ZoralineOracle, 3, 3)
	small := game.NewCard("Small Relic", me.ID)
	small.TypeLine = "Artifact"
	small.ManaCost = "{2}{B}"
	big := game.NewCard("Big Relic", me.ID)
	big.TypeLine = "Artifact"
	big.ManaCost = "{4}"
	land := game.NewCard("Old Land", me.ID)
	land.TypeLine = "Land"
	spell := game.NewCard("Old Spell", me.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{B}"
	for _, c := range []game.Card{small, big, land, spell} {
		me.Graveyard.PushTop(c)
	}
	declareAttack(t, g, opp.ID, zoraline)
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		blbGLSettleTriggerOrder(g, me.ID)
		if hasPayUnlessFor(g, me.ID) {
			b06AddMana(me, "W", "B")
			answerPayUnless(t, g, me.ID, true)
			continue
		}
		passPriorityAroundTable(t, g)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("paying should create the reflexive trigger")
	}
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != small.InstanceID {
		t.Errorf("offered %v, want only the mana value 3 artifact", pick.PickTargetCards)
	}
}

func TestBlbRUZ2ZoralineDecliningReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	zoraline := blbB1Push(g, me.ID, "Zoraline, Cosmos Caller", "Legendary Creature — Bat Cleric", blbRUZ2ZoralineOracle, 3, 3)
	small := game.NewCard("Small Relic", me.ID)
	small.TypeLine = "Artifact"
	small.ManaCost = "{1}"
	me.Graveyard.PushTop(small)
	life := lifeOf(g, me.ID)
	declareAttack(t, g, opp.ID, zoraline)
	blbRUZ2Settle(t, g, me, false, small.InstanceID)
	if g.Battlefield.Contains(small.InstanceID) {
		t.Error("declining the payment returned a card")
	}
	if got := lifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d (Zoraline is a Bat, nothing paid)", got, life+1)
	}
}
