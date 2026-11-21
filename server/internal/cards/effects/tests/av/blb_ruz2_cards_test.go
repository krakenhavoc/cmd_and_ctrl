package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- Reptilian Recruiter ------------------------------------------------------

// blbRUZ2RecruiterSteals casts the Recruiter, aims its entry trigger at
// target and resolves it.
func blbRUZ2RecruiterSteals(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	recruiter := effectstest.BlbGLCastCreature(t, g, "Reptilian Recruiter", "Creature — Lizard Warrior", blbRUZ2RecruiterOracle, 4, 2, nil)
	effectstest.AnswerPickTarget(t, g, target)
	effectstest.PassPriorityAroundTable(t, g)
	return recruiter
}

func TestBlbRUZ2ReptilianRecruiterStealsASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	small := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.FindBattlefieldCardForTest(g, small).Tapped = true
	recruiter := blbRUZ2RecruiterSteals(t, g, small)
	if !effectstest.HasKeywordOnBattlefield(t, g, recruiter, "trample") {
		t.Error("the Recruiter lacks trample")
	}
	c, ok := effectstest.BattlefieldCard(g, small)
	if !ok || c.Controller != me.ID {
		t.Fatal("a power-2 creature should change control")
	}
	if c.Tapped {
		t.Error("the stolen creature should be untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, small, "haste") {
		t.Error("the stolen creature should have haste")
	}
	effectstest.EndTurn(t, g)
	if effectstest.ControllerOf(t, g, small) != opp.ID {
		t.Error("control should return at end of turn")
	}
}

func TestBlbRUZ2ReptilianRecruiterNeedsAnotherLizardForABigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	big := effectstest.BlbB1Push(g, opp.ID, "Their Ox", "Creature — Ox", "", 3, 3)
	blbRUZ2RecruiterSteals(t, g, big)
	if effectstest.ControllerOf(t, g, big) != opp.ID {
		t.Error("power 3 and no other Lizard: the creature should not change control")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Lizard", "Creature — Lizard", "", 1, 1)
	theirLizard := effectstest.BlbB1Push(g, opp.ID, "Their Lizard", "Creature — Lizard", "", 1, 1)
	big = effectstest.BlbB1Push(g, opp.ID, "Their Ox", "Creature — Ox", "", 3, 3)
	blbRUZ2RecruiterSteals(t, g, big)
	if effectstest.ControllerOf(t, g, big) != me.ID {
		t.Error("with another Lizard you control, the power-3 creature should change control")
	}
	if effectstest.ControllerOf(t, g, theirLizard) != opp.ID {
		t.Error("an untargeted creature changed control")
	}
}

func TestBlbRUZ2ReptilianRecruiterIgnoresAnOpponentsLizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, opp.ID, "Their Lizard", "Creature — Lizard", "", 1, 1)
	big := effectstest.BlbB1Push(g, opp.ID, "Their Ox", "Creature — Ox", "", 3, 3)
	blbRUZ2RecruiterSteals(t, g, big)
	if effectstest.ControllerOf(t, g, big) != opp.ID {
		t.Error("a Lizard an opponent controls is not one you control")
	}
}

// --- Ruthless Negotiation -----------------------------------------------------

func TestBlbRUZ2RuthlessNegotiationExilesTheOpponentsPick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	keep := effectstest.HandCardForTest(opp, "Keep", "Instant", "")
	lose := effectstest.HandCardForTest(opp, "Lose", "Sorcery", "")
	effectstest.CastCatalogSpell(t, g, "Ruthless Negotiation", "Sorcery", blbRUZ2NegotiationOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the target opponent should be asked which card to exile")
	}
	if pick.ChooseMin != 1 {
		t.Errorf("the opponent must exile a card: min %d", pick.ChooseMin)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, lose)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(keep) || opp.Hand.Contains(lose) || !g.Exile.Contains(lose) {
		t.Error("the picked card should be exiled and the other kept")
	}
	if me.Hand.Size() != hand {
		t.Errorf("cast from hand: my hand is %d, want %d (no draw)", me.Hand.Size(), hand)
	}
}

func TestBlbRUZ2RuthlessNegotiationFlashbackAlsoDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	lose := effectstest.HandCardForTest(opp, "Lose", "Sorcery", "")
	id := effectstest.SeedGraveyardCard(t, g, "Ruthless Negotiation", "Sorcery", blbRUZ2NegotiationOracle)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, lose)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Ruthless Negotiation", "Sorcery", blbRUZ2NegotiationOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Error("you are not a legal target")
	}
}

// --- Valley Questcaller -------------------------------------------------------

func TestBlbRUZ2ValleyQuestcallerPumpsItsFourTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	questcaller := effectstest.BlbB1Push(g, me.ID, "Valley Questcaller", "Creature — Rabbit Warrior", blbRUZ2QuestcallerOracle, 2, 3)
	rabbit := effectstest.BlbB1Push(g, me.ID, "Rabbit", "Creature — Rabbit", "", 1, 1)
	mouse := effectstest.BlbB1Push(g, me.ID, "Mouse", "Creature — Mouse Soldier", "", 1, 1)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirBat := effectstest.BlbB1Push(g, opp.ID, "Their Bat", "Creature — Bat", "", 1, 1)
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
		if p, tough := effectstest.PtOf(t, g, c.id); p != c.p || tough != c.t {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.p, c.t)
		}
	}
}

func TestBlbRUZ2ValleyQuestcallerScriesWhenItsKinEnter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.BlbB1Push(g, me.ID, "Valley Questcaller", "Creature — Rabbit Warrior", blbRUZ2QuestcallerOracle, 2, 3)
	effectstest.BlbGLCastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Fatal("a Bear entering should not scry")
	}
	effectstest.BlbGLCastCreature(t, g, "Bird", "Creature — Bird", "", 1, 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("a Bird entering should scry 1, got %+v", c)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

// --- Warren Warleader ---------------------------------------------------------

// blbRUZ2WarleaderAttacks attacks with the Warleader and a Bear and
// chooses the trigger's mode.
func blbRUZ2WarleaderAttacks(t *testing.T, g *game.Game, mode int) (warleader, bear uuid.UUID) {
	t.Helper()
	me, opp, _ := effectstest.BlbB1Seats(g)
	warleader = effectstest.BlbB1Push(g, me.ID, "Warren Warleader", "Creature — Rabbit Knight", blbRUZ2WarleaderOracle, 4, 4)
	bear = effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, warleader, bear)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	return warleader, bear
}

func TestBlbRUZ2WarrenWarleaderPumpsAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	idle := effectstest.BlbB1Push(g, me.ID, "Idle Bear", "Creature — Bear", "", 2, 2)
	warleader, bear := blbRUZ2WarleaderAttacks(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, warleader); p != 5 || tough != 5 {
		t.Errorf("the attacking Warleader is %d/%d, want 5/5", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the attacking Bear is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, idle); p != 2 || tough != 2 {
		t.Errorf("a creature that did not attack is %d/%d, want 2/2", p, tough)
	}
}

func TestBlbRUZ2WarrenWarleaderMakesAnAttackingRabbit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	blbRUZ2WarleaderAttacks(t, g, 0)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	tokens := effectstest.BattlefieldIDsNamed(g, "Rabbit")
	if len(tokens) != 1 {
		t.Fatalf("%d Rabbit tokens, want 1", len(tokens))
	}
	tok := effectstest.CardByID(g, tokens[0])
	if !tok.Tapped || tok.AttackingTarget != opp.ID || tok.Controller != me.ID {
		t.Errorf("want a tapped Rabbit attacking the defender: %+v", tok)
	}
	if p, tough := effectstest.PtOf(t, g, tokens[0]); p != 1 || tough != 1 {
		t.Errorf("the Rabbit is %d/%d, want 1/1", p, tough)
	}
}

func TestBlbRUZ2WarrenWarleaderHasOffspring(t *testing.T) {
	spec, _ := effects.Lookup(blbRUZ2WarleaderOracle)
	if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != effects.OffspringKey || spec.OptionalCosts[0].ManaCost != "{2}" {
		t.Fatalf("optional costs %+v, want Offspring {2}", spec.OptionalCosts)
	}
}

// --- Wick's Patrol ------------------------------------------------------------

func TestBlbRUZ2WicksPatrolShrinksByTheGreatestManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibraryOfCost(g, me, 4, "{1}{B}")
	big := game.NewCard("Big Spell", me.ID)
	big.TypeLine = "Sorcery"
	big.ManaCost = "{3}{B}{B}"
	me.Graveyard.PushTop(big)
	mine := effectstest.BlbB1Push(g, me.ID, "My Ox", "Creature — Ox", "", 6, 6)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 6, 6)
	effectstest.BlbGLCastCreature(t, g, "Wick's Patrol", "Creature — Rat Warlock", blbRUZ2PatrolOracle, 5, 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 4 {
		t.Fatalf("the graveyard holds %d cards, want 4 (three milled)", got)
	}
	pick := effectstest.PickTargetPrompt(t, g)
	for _, id := range pick.PickTargetCards {
		if id == mine {
			t.Fatal("a creature you control is not a legal target")
		}
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1 (-5/-5)", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, mine); p != 6 || tough != 6 {
		t.Errorf("my creature is %d/%d, want 6/6", p, tough)
	}
}

func TestBlbRUZ2WicksPatrolNeedsThreeCardsToMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibraryOfCost(g, me, 2, "{5}")
	effectstest.BlbB1Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 6, 6)
	effectstest.BlbGLCastCreature(t, g, "Wick's Patrol", "Creature — Rat Warlock", blbRUZ2PatrolOracle, 5, 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 2 {
		t.Errorf("the graveyard holds %d cards, want the 2 the library had", got)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("fewer than three cards milled: the reflexive trigger should not happen")
	}
}

// --- Zoraline, Cosmos Caller --------------------------------------------------

// blbRUZ2Settle answers trigger order, Zoraline's offer (paying when pay)
// and the return target, and passes priority until the stack is empty.
func blbRUZ2Settle(t *testing.T, g *game.Game, me *game.Player, pay bool, target uuid.UUID) {
	t.Helper()
	for i := 0; i < 12; i++ {
		effectstest.BlbGLSettleTriggerOrder(g, me.ID)
		switch {
		case effectstest.HasPayUnlessFor(g, me.ID):
			if pay {
				effectstest.B06AddMana(me, "W", "B")
			}
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
		case effectstest.LatestPickTarget(g, me.ID) != nil:
			effectstest.AnswerPickTarget(t, g, target)
		case effectstest.StackFullyEmpty(g):
			return
		default:
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	t.Fatal("the stack did not settle")
}

func TestBlbRUZ2ZoralineGainsPerBatAndReturnsWithFinality(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	zoraline := effectstest.BlbB1Push(g, me.ID, "Zoraline, Cosmos Caller", "Legendary Creature — Bat Cleric", blbRUZ2ZoralineOracle, 3, 3)
	bat := effectstest.BlbB1Push(g, me.ID, "Bat", "Creature — Bat", "", 1, 1)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, zoraline, kw) {
			t.Errorf("Zoraline lacks %s", kw)
		}
	}
	small := game.NewCard("Small Relic", me.ID)
	small.TypeLine = "Artifact"
	small.ManaCost = "{2}{B}"
	me.Graveyard.PushTop(small)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, zoraline, bat, bear)
	blbRUZ2Settle(t, g, me, true, small.InstanceID)
	if got := effectstest.LifeOf(g, me.ID); got != life+2-2 {
		t.Errorf("life %d, want %d (two Bats attacked, 2 life paid)", got, life)
	}
	if !g.Battlefield.Contains(small.InstanceID) {
		t.Fatal("the mana value 3 card should return")
	}
	if n := effectstest.CountersOn(g, small.InstanceID, game.CounterFinality); n != 1 {
		t.Errorf("%d finality counters, want 1", n)
	}
}

func TestBlbRUZ2ZoralineOffersOnlySmallNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	zoraline := effectstest.BlbB1Push(g, me.ID, "Zoraline, Cosmos Caller", "Legendary Creature — Bat Cleric", blbRUZ2ZoralineOracle, 3, 3)
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
	effectstest.DeclareAttack(t, g, opp.ID, zoraline)
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		effectstest.BlbGLSettleTriggerOrder(g, me.ID)
		if effectstest.HasPayUnlessFor(g, me.ID) {
			effectstest.B06AddMana(me, "W", "B")
			effectstest.AnswerPayUnless(t, g, me.ID, true)
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("paying should create the reflexive trigger")
	}
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != small.InstanceID {
		t.Errorf("offered %v, want only the mana value 3 artifact", pick.PickTargetCards)
	}
}

func TestBlbRUZ2ZoralineDecliningReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	zoraline := effectstest.BlbB1Push(g, me.ID, "Zoraline, Cosmos Caller", "Legendary Creature — Bat Cleric", blbRUZ2ZoralineOracle, 3, 3)
	small := game.NewCard("Small Relic", me.ID)
	small.TypeLine = "Artifact"
	small.ManaCost = "{1}"
	me.Graveyard.PushTop(small)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, zoraline)
	blbRUZ2Settle(t, g, me, false, small.InstanceID)
	if g.Battlefield.Contains(small.InstanceID) {
		t.Error("declining the payment returned a card")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d (Zoraline is a Bat, nothing paid)", got, life+1)
	}
}
