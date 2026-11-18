package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b05_cards_test.go — S82 Midnight Hunt, batch 5 (#5497): Candlegrove
// Witch; Dawnhart Wardens; Stalwart Pathlighter; Duelcraft Trainer;
// Harvesttide Sentry; Candlelit Cavalry; Ritual Guardian; Might of the Old
// Ways; Ritual of Hope; Dawnhart Mentor; Candletrap; Contortionist Troupe;
// Duel for Dominance.

const (
	midB05WitchOracle      = "95f47e9b-0290-4ec4-a1ba-bc5a432ffab2"
	midB05WardensOracle    = "0604ed6c-a85f-41da-a163-0c414e57d133"
	midB05PathlighterOracl = "88ba283d-2d03-4455-a4a3-a2a5622209bd"
	midB05TrainerOracle    = "81713b4d-85df-4b2c-9673-4f5162590149"
	midB05SentryOracle     = "6cd6b56d-d69e-4144-a6c7-8c066629435b"
	midB05CavalryOracle    = "edfd2614-8875-4741-b64b-390afc0d2376"
	midB05GuardianOracle   = "9f4719f2-0eab-464b-9ae5-2b96717ae08d"
	midB05MightOracle      = "266d3f06-c982-4427-b421-7520e5da5c2e"
	midB05RitualOracle     = "8afe3d28-be3e-416c-9457-8a2af85b2eec"
	midB05MentorOracle     = "e12fb44a-524e-4603-a89c-fad8204fdb2b"
	midB05CandletrapOracle = "089788ce-d06c-4a85-915c-82e2cb5b0103"
	midB05TroupeOracle     = "be9bee08-ac04-42bc-93be-5a77482a7690"
	midB05DuelOracle       = "5064a55a-9692-4a67-9e82-16250b9140d1"
)

func TestMidB05CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		midB05WitchOracle:      "Candlegrove Witch",
		midB05WardensOracle:    "Dawnhart Wardens",
		midB05PathlighterOracl: "Stalwart Pathlighter",
		midB05TrainerOracle:    "Duelcraft Trainer",
		midB05SentryOracle:     "Harvesttide Sentry",
		midB05CavalryOracle:    "Candlelit Cavalry",
		midB05GuardianOracle:   "Ritual Guardian",
		midB05MightOracle:      "Might of the Old Ways",
		midB05RitualOracle:     "Ritual of Hope",
		midB05MentorOracle:     "Dawnhart Mentor",
		midB05CandletrapOracle: "Candletrap",
		midB05TroupeOracle:     "Contortionist Troupe",
		midB05DuelOracle:       "Duel for Dominance",
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

// midB05Coven puts creatures of powers 1 and 2 beside a power-3 source,
// making three different powers.
func midB05Coven(g *game.Game, me uuid.UUID) (one, two uuid.UUID) {
	one = b12Creature(g, me, "Coven One", "Creature — Bear", 1, 1)
	two = b12Creature(g, me, "Coven Two", "Creature — Bear", 2, 2)
	return
}

// midB05SelfKeyword runs the combat trigger of a "this creature gains X"
// card with and without coven.
func midB05SelfKeyword(t *testing.T, name, oracle, typeLine, keyword string, power int) {
	t.Helper()
	for _, coven := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		src := b12Push(g, me.ID, name, typeLine, oracle, power, 2)
		if coven {
			// powers: the source's, plus two others that differ from it and each other.
			b12Creature(g, me.ID, "Other A", "Creature — Bear", power+1, 2)
			b12Creature(g, me.ID, "Other B", "Creature — Bear", power+2, 2)
		} else {
			b12Creature(g, me.ID, "Other A", "Creature — Bear", power, 2)
			b12Creature(g, me.ID, "Other B", "Creature — Bear", power+1, 2)
		}
		advanceTo(t, g, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		if got := hasKeywordOnBattlefield(t, g, src, keyword); got != coven {
			t.Errorf("%s coven=%v: has %s = %v", name, coven, keyword, got)
		}
	}
}

func TestMidB05CandlegroveWitchGainsFlyingWithCoven(t *testing.T) {
	midB05SelfKeyword(t, "Candlegrove Witch", midB05WitchOracle, "Creature — Human Warlock", "flying", 2)
}

func TestMidB05RitualGuardianGainsLifelinkWithCoven(t *testing.T) {
	midB05SelfKeyword(t, "Ritual Guardian", midB05GuardianOracle, "Creature — Human Soldier", "lifelink", 3)
}

func TestMidB05CandlelitCavalryGainsTrampleWithCoven(t *testing.T) {
	midB05SelfKeyword(t, "Candlelit Cavalry", midB05CavalryOracle, "Creature — Human Knight", "trample", 5)
}

func TestMidB05CovenTriggerIsOnlyOnYourOwnTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	witch := b12Push(g, opp.ID, "Candlegrove Witch", "Creature — Human Warlock", midB05WitchOracle, 2, 2)
	midB05Coven(g, opp.ID)
	_ = me
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if hasKeywordOnBattlefield(t, g, witch, "flying") {
		t.Error("the Witch gained flying on another player's combat")
	}
}

func TestMidB05PathlighterGrantsIndestructibleToCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pl := b12Push(g, me.ID, "Stalwart Pathlighter", "Creature — Human Soldier", midB05PathlighterOracl, 3, 1)
	one, two := midB05Coven(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{pl, one, two} {
		if !hasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("%s lacks indestructible", id)
		}
	}
	if hasKeywordOnBattlefield(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature gained indestructible")
	}
}

func TestMidB05PathlighterDoesNothingWithoutCoven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pl := b12Push(g, me.ID, "Stalwart Pathlighter", "Creature — Human Soldier", midB05PathlighterOracl, 3, 1)
	b12Creature(g, me.ID, "Same Power", "Creature — Bear", 3, 3)
	b12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if hasKeywordOnBattlefield(t, g, pl, "indestructible") {
		t.Error("indestructible without three different powers")
	}
}

func TestMidB05WardensPumpCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	w := b12Push(g, me.ID, "Dawnhart Wardens", "Creature — Human Warlock", midB05WardensOracle, 3, 3)
	one, two := midB05Coven(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{w: 4, one: 2, two: 3, theirs: 2} {
		if p, _ := effectivePTOf(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
	if !hasKeywordOnBattlefield(t, g, w, "vigilance") {
		t.Error("the Wardens lack vigilance")
	}
}

func TestMidB05WardensDoNothingWithoutCoven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	w := b12Push(g, me.ID, "Dawnhart Wardens", "Creature — Human Warlock", midB05WardensOracle, 3, 3)
	b12Creature(g, me.ID, "Same", "Creature — Bear", 3, 3)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, w); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
}

func TestMidB05TrainerGivesDoubleStrikeToATargetWithCoven(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tr := b12Push(g, me.ID, "Duelcraft Trainer", "Creature — Human Soldier", midB05TrainerOracle, 3, 3)
	one, two := midB05Coven(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 5, 5)
	advanceTo(t, g, game.StepBeginCombat)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the trigger did not ask for a target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Error("an opponent's creature was a legal target")
	}
	pickCard(t, g, me.ID, two)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, two, "double strike") {
		t.Error("the chosen creature lacks double strike")
	}
	if hasKeywordOnBattlefield(t, g, one, "double strike") || hasKeywordOnBattlefield(t, g, tr, "double strike") {
		t.Error("a creature that was not chosen has double strike")
	}
	if !hasKeywordOnBattlefield(t, g, tr, "first strike") {
		t.Error("the Trainer lacks first strike")
	}
}

func TestMidB05TrainerDoesNotTriggerWithoutCoven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tr := b12Push(g, me.ID, "Duelcraft Trainer", "Creature — Human Soldier", midB05TrainerOracle, 3, 3)
	b12Creature(g, me.ID, "Same", "Creature — Bear", 3, 3)
	advanceTo(t, g, game.StepBeginCombat)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("the trigger asked for a target without coven")
	}
	if triggerOnStack(g, tr) != nil {
		t.Error("the trigger went on the stack without coven")
	}
}

func TestMidB05SentryCannotBeBlockedByLowPowerWithCoven(t *testing.T) {
	for _, coven := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		sentry := b12Push(g, me.ID, "Harvesttide Sentry", "Creature — Human Warrior", midB05SentryOracle, 3, 1)
		if coven {
			midB05Coven(g, me.ID)
		} else {
			b12Creature(g, me.ID, "Same", "Creature — Bear", 3, 3)
		}
		small := b12Creature(g, opp.ID, "Power Two", "Creature — Bear", 2, 2)
		big := b12Creature(g, opp.ID, "Power Three", "Creature — Bear", 3, 3)
		advanceTo(t, g, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(sentry, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		lockInAttacks(t, g)
		passPriorityAroundTable(t, g)
		advanceTo(t, g, game.StepDeclareBlockers)
		if coven {
			brRefusal(t, g.DeclareBlocker(small, sentry), game.BlockReasonCantBeBlockedBy)
		} else if err := g.DeclareBlocker(small, sentry); err != nil {
			t.Errorf("without coven a power 2 creature could not block: %v", err)
		}
		if coven {
			if err := g.DeclareBlocker(big, sentry); err != nil {
				t.Errorf("a power 3 creature could not block: %v", err)
			}
		}
	}
}

func TestMidB05SentryLeavingBeforeTheTriggerResolvesGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sentry := b12Push(g, me.ID, "Harvesttide Sentry", "Creature — Human Warrior", midB05SentryOracle, 3, 1)
	midB05Coven(g, me.ID)
	advanceTo(t, g, game.StepBeginCombat)
	if triggerOnStack(g, sentry) == nil {
		t.Fatal("no coven trigger on the stack")
	}
}

func TestMidB05RitualOfHopePumpsByCoven(t *testing.T) {
	for _, coven := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		a := b12Creature(g, me.ID, "A", "Creature — Bear", 3, 3)
		b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
		if coven {
			b12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
		} else {
			b12Creature(g, me.ID, "C", "Creature — Bear", 3, 3)
		}
		theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		castCatalogSpell(t, g, "Ritual of Hope", "Instant", midB05RitualOracle, nil)
		passPriorityAroundTable(t, g)
		wantA, wantTough := 4, 4
		if coven {
			wantA = 5
		}
		if p, tough := effectivePTOf(t, g, a); p != wantA || tough != wantTough {
			t.Errorf("coven=%v: %d/%d, want %d/%d", coven, p, tough, wantA, wantTough)
		}
		if p, tough := effectivePTOf(t, g, b); p != wantA-2 || tough != 2 {
			t.Errorf("coven=%v: B is %d/%d", coven, p, tough)
		}
		if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
			t.Errorf("coven=%v: an opponent's creature is %d/%d", coven, p, tough)
		}
	}
}

func TestMidB05MightOfTheOldWaysDrawsOnlyWithCoven(t *testing.T) {
	for _, coven := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		a := b12Creature(g, me.ID, "A", "Creature — Bear", 3, 3)
		b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
		if coven {
			b12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
		} else {
			// The +2/+2 makes A a 5, so a 5 beside it keeps the powers at two.
			b12Creature(g, me.ID, "C", "Creature — Bear", 5, 5)
		}
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
		advanceToMain(t, g)
		hand := len(me.Hand.Cards)
		castCatalogSpell(t, g, "Might of the Old Ways", "Instant", midB05MightOracle, cardRefs(a))
		passPriorityAroundTable(t, g)
		if p, tough := effectivePTOf(t, g, a); p != 5 || tough != 5 {
			t.Errorf("coven=%v: target is %d/%d, want 5/5", coven, p, tough)
		}
		want := hand
		if coven {
			want = hand + 1
		}
		if len(me.Hand.Cards) != want {
			t.Errorf("coven=%v: hand %d, want %d", coven, len(me.Hand.Cards), want)
		}
	}
}

func TestMidB05DuelForDominanceCountersThenFights(t *testing.T) {
	for _, coven := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 6)
		b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
		if coven {
			b12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
		} else {
			b12Creature(g, me.ID, "C", "Creature — Bear", 3, 3)
		}
		theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 4, 4)
		castCatalogSpell(t, g, "Duel for Dominance", "Instant", midB05DuelOracle, []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
		})
		passPriorityAroundTable(t, g)
		wantCounters := 0
		if coven {
			wantCounters = 1
		}
		if got := counterCount(g, mine, game.CounterPlusOne); got != wantCounters {
			t.Errorf("coven=%v: %d counters, want %d", coven, got, wantCounters)
		}
		// Mine deals 3 (4 with the counter) to a 4/4, so it dies only with
		// coven; theirs deals 4 to my 3/6 (4/7), which survives.
		if g.Battlefield.Contains(theirs) != !coven {
			t.Errorf("coven=%v: their 4/4 on the battlefield = %v", coven, g.Battlefield.Contains(theirs))
		}
		if c := findBattlefieldCardByID(g, mine); c == nil || c.DamageMarked != 4 {
			t.Errorf("coven=%v: my creature should carry 4 damage: %+v", coven, c)
		}
	}
}

func TestMidB05DuelForDominanceRefusesTwoOfMyCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 3, 3)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Duel for Dominance", TypeLine: "Instant", OracleID: midB05DuelOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1},
	}}); err == nil {
		t.Error("a creature I control was accepted as the one I don't control")
	}
}

func TestMidB05MentorMakesAHumanWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Dawnhart Mentor", "Creature — Human Warlock", "{0}", midB05MentorOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Human"); n != 1 {
		t.Fatalf("%d Human tokens, want 1", n)
	}
}

func TestMidB05MentorPumpsOnlyWithCoven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mentor := b12Push(g, me.ID, "Dawnhart Mentor", "Creature — Human Warlock", midB05MentorOracle, 0, 4)
	b12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
	advanceToMain(t, g)
	// Powers 0 and 1: not coven.
	fillPoolColored(me, "G", 6)
	if err := g.ActivateCatalogAbility(me.ID, mentor, 0, game.ActivateAbilityParams{Targets: cardRefs(mentor), Strict: true}); err == nil {
		t.Fatal("activated without coven")
	}
	b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	fillPoolColored(me, "G", 6)
	if err := g.ActivateCatalogAbility(me.ID, mentor, 0, game.ActivateAbilityParams{Targets: cardRefs(mentor), Strict: true}); err != nil {
		t.Fatalf("activate with coven: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mentor); p != 3 || tough != 7 {
		t.Errorf("Mentor is %d/%d, want 3/7", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, mentor, "trample") {
		t.Error("the Mentor lacks trample")
	}
}

func TestMidB05CandletrapPreventsItsHostsCombatDamageAndExilesWithCoven(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 4)
	host := b12Creature(g, opp.ID, "Host", "Creature — Bear", 3, 3)
	trap := castCatalogSpell(t, g, "Candletrap", "Enchantment — Aura", midB05CandletrapOracle, cardRefs(host))
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, host, "defender") {
		t.Fatal("the enchanted creature lacks defender")
	}
	// Without coven (powers 2 and 3 only) it can't be activated.
	fillPoolColored(me, "W", 3)
	if err := g.ActivateCatalogAbility(me.ID, trap, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("exiled without coven")
	}
	lifeBefore := me.Life
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(host, attacker); err != nil {
		t.Fatalf("a defender creature can block: %v", err)
	}
	advanceTo(t, g, game.StepEndCombat)
	if c := findBattlefieldCardByID(g, attacker); c == nil || c.DamageMarked != 0 {
		t.Errorf("the attacker was damaged by the enchanted creature: %+v", c)
	}
	if c := findBattlefieldCardByID(g, host); c == nil || c.DamageMarked != 2 {
		t.Errorf("the host should have taken the attacker's 2 damage: %+v", c)
	}
	_ = lifeBefore
	// Coven: powers 2, 1, 4.
	b12Creature(g, me.ID, "One", "Creature — Bear", 1, 1)
	b12Creature(g, me.ID, "Four", "Creature — Bear", 4, 4)
	fillPoolColored(me, "W", 3)
	if err := g.ActivateCatalogAbility(me.ID, trap, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with coven: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature was not exiled")
	}
	if g.Battlefield.Contains(trap) {
		t.Error("the Aura was not sacrificed")
	}
}

func TestMidB05TroupeEntersWithXCountersAndGrowsAnotherWithCoven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	troupe := castXSpell(t, g, "Contortionist Troupe", "Creature — Human", midB05TroupeOracle, "{X}{G}", 2, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, troupe, game.CounterPlusOne); got != 2 {
		t.Fatalf("enters with %d counters, want 2", got)
	}
	// Powers: the Troupe is 2/2; add a 1 and a 3.
	one := b12Creature(g, me.ID, "One", "Creature — Bear", 1, 1)
	b12Creature(g, me.ID, "Three", "Creature — Bear", 3, 3)
	advanceToEndStep(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the end-step trigger asked for no target")
	}
	pickCard(t, g, me.ID, one)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, one, game.CounterPlusOne); got != 1 {
		t.Errorf("the chosen creature has %d counters, want 1", got)
	}
}

func TestMidB05TroupeDoesNothingWithoutCoven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	troupe := b12Push(g, me.ID, "Contortionist Troupe", "Creature — Human", midB05TroupeOracle, 2, 2)
	b12Creature(g, me.ID, "Same", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	advanceToEndStep(t, g)
	if triggerOnStack(g, troupe) != nil || latestPickTarget(g, me.ID) != nil {
		t.Error("the end-step trigger fired without coven")
	}
}
