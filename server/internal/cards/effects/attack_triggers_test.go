package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_triggers_test.go — S22: "whenever ~ attacks" triggers on
// the new EventAttack. Seat 0 is the active player throughout; its
// creatures attack seat 1 (and seat 2 for the multi-defender case).
//
// Attack triggers land during the declare-attackers step, so these
// tests stop at StepDeclareAttackers rather than walking on to
// combat damage the way the S19 sub-PR 7 tests do — the whole point
// of the event is that the trigger is answerable before blockers.

const (
	hellriderOracle   = "f02557b0-f422-48cb-875e-2c814c39967d"
	sunTitanOracle    = "b2e950fb-cb7e-40a0-a311-5bbdd0477b29"
	krenkoKingpinOrcl = "e8065e1d-e937-4b56-8011-78f0d07328a0"
)

// declareAttack advances to the declare-attackers step and declares
// each creature against defender, leaving the cursor there.
func declareAttack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker %s: %v", id, err)
		}
	}
}

// triggersOnStackFrom counts the triggered-ability items in
// StackMeta sourced from cardID. Attack triggers fire once per
// attacking creature, so the count is the assertion.
func triggersOnStackFrom(g *game.Game, cardID uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == cardID {
			n++
		}
	}
	return n
}

// pushGraveyardPermanent seeds a permanent card with a real mana
// cost into a player's graveyard — Sun Titan's mana-value clause
// needs a parseable cost, which pushGraveyardCardForTest does not
// set.
func pushGraveyardPermanent(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		ManaCost:   manaCost,
		Power:      2,
		Toughness:  2,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// --- Hellrider --------------------------------------------------

// One trigger per attacking creature, and nothing happens until
// they resolve: the pings are stack objects with a response window,
// not a side effect of declaring.
func TestHellriderTriggersOncePerAttackerAndWaitsToResolve(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hell := pushDiesCreatureForTest(g, me.ID, "Hellrider", hellriderOracle, "Creature — Devil", 3, 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	lifeBefore := opp.Life

	// Hellrider counts itself — "a creature you control", not
	// "another" — so two attackers make two triggers.
	declareAttack(t, g, opp.ID, hell, bear)

	if g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("step = %s, want the triggers to land in declare_attackers", g.Turn.Step)
	}
	if got := triggersOnStackFrom(g, hell); got != 2 {
		t.Fatalf("Hellrider triggers on the stack = %d, want 2", got)
	}
	if opp.Life != lifeBefore {
		t.Errorf("damage landed before the triggers resolved (life %d, was %d)", opp.Life, lifeBefore)
	}

	passPriorityAroundTable(t, g)
	if got := lifeBefore - opp.Life; got != 2 {
		t.Errorf("Hellrider dealt %d total, want 2 (one per attacker)", got)
	}
}

// The damaged player is the defender THAT attacker was declared
// against, not "an opponent". Two creatures attacking two different
// seats ping two different seats.
func TestHellriderFollowsEachAttackersOwnDefender(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	pushDiesCreatureForTest(g, me.ID, "Hellrider", hellriderOracle, "Creature — Devil", 3, 3)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	firstBefore, secondBefore := first.Life, second.Life

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("DeclareAttacker a: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("DeclareAttacker b: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := firstBefore - first.Life; got != 1 {
		t.Errorf("first defender took %d, want 1", got)
	}
	if got := secondBefore - second.Life; got != 1 {
		t.Errorf("second defender took %d, want 1", got)
	}
}

// "A creature YOU control": an opponent's Hellrider stays quiet
// while my creatures attack.
func TestHellriderIgnoresAnotherPlayersAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hell := pushDiesCreatureForTest(g, opp.ID, "Hellrider", hellriderOracle, "Creature — Devil", 3, 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	declareAttack(t, g, opp.ID, bear)
	if got := triggersOnStackFrom(g, hell); got != 0 {
		t.Errorf("opponent's Hellrider triggered %d times on MY attacker", got)
	}
	if len(g.PendingTriggers) != 0 {
		t.Errorf("unexpected pending triggers: %d", len(g.PendingTriggers))
	}
}

// --- Krenko, Tin Street Kingpin ---------------------------------

// The printed "then" is load-bearing: the counter goes on first, so
// a 1/2 Krenko makes TWO Goblins on its first swing.
func TestKrenkoCounterThenGoblinsEqualToItsNewPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krenko := pushDiesCreatureForTest(g, me.ID, "Krenko, Tin Street Kingpin", krenkoKingpinOrcl,
		"Legendary Creature — Goblin", 1, 2)

	declareAttack(t, g, opp.ID, krenko)
	if triggersOnStackFrom(g, krenko) != 1 {
		t.Fatalf("Krenko did not put exactly one trigger on the stack")
	}
	if got := countBattlefieldNamed(g, me.ID, "Goblin"); got != 0 {
		t.Fatalf("%d Goblins made before the trigger resolved", got)
	}

	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, krenko)
	if !ok {
		t.Fatalf("Krenko left the battlefield")
	}
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("Krenko +1/+1 counters = %d, want 1", c.Counters["+1/+1"])
	}
	if got := countBattlefieldNamed(g, me.ID, "Goblin"); got != 2 {
		t.Errorf("Goblins = %d, want 2 (power 1 + the counter placed first)", got)
	}
}

// "Whenever THIS creature attacks" — another attacker does nothing.
func TestKrenkoIgnoresOtherAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krenko := pushDiesCreatureForTest(g, me.ID, "Krenko, Tin Street Kingpin", krenkoKingpinOrcl,
		"Legendary Creature — Goblin", 1, 2)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	declareAttack(t, g, opp.ID, bear)
	if got := triggersOnStackFrom(g, krenko); got != 0 {
		t.Errorf("Krenko triggered %d times on another creature's attack", got)
	}
	if got := countBattlefieldNamed(g, me.ID, "Goblin"); got != 0 {
		t.Errorf("Goblins = %d, want 0", got)
	}
}

// --- Sun Titan --------------------------------------------------

// The attack half: declare Sun Titan, say yes, pick a cheap
// enough permanent in the graveyard, and it comes back to the
// battlefield. The mana-value clause filters the legal set.
func TestSunTitanAttackReanimatesChosenPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := pushDiesCreatureForTest(g, me.ID, "Sun Titan", sunTitanOracle, "Creature — Giant", 6, 6)
	cheap := pushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	dear := pushGraveyardPermanent(me, "Serra Angel", "Creature — Angel", "{3}{W}{W}")

	declareAttack(t, g, opp.ID, titan)
	answerLatestTriggerPrompt(t, g, me.ID, true)

	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt after Yes")
	}
	if !hasID(prompt.PickTargetCards, cheap) {
		t.Errorf("legal set missing the mana-value-2 permanent")
	}
	if hasID(prompt.PickTargetCards, dear) {
		t.Errorf("mana-value-5 permanent offered as a Sun Titan target")
	}

	pickCard(t, g, me.ID, cheap)
	if triggerOnStack(g, titan) == nil {
		t.Fatalf("Sun Titan trigger not on the stack after the pick")
	}
	if g.Battlefield.Contains(cheap) {
		t.Errorf("reanimation happened before the trigger resolved")
	}

	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Errorf("Sun Titan did not return the chosen permanent to the battlefield")
	}
	if me.Graveyard.Contains(cheap) {
		t.Errorf("the returned card is still in the graveyard")
	}
}

// "Enters OR attacks" is one ability with two conditions — the
// enters half must still work after the attack half was added.
func TestSunTitanEntersHalfStillFires(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := pushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")

	castAndResolveCreature(t, g, "Sun Titan", "Creature — Giant", sunTitanOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(cheap) {
		t.Errorf("Sun Titan's enters trigger did not reanimate")
	}
}

// CR 603.3d: a targeted trigger with an empty legal set is removed
// without a prompt. An empty graveyard means no question is asked.
func TestSunTitanWithEmptyGraveyardAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := pushDiesCreatureForTest(g, me.ID, "Sun Titan", sunTitanOracle, "Creature — Giant", 6, 6)

	declareAttack(t, g, opp.ID, titan)
	for _, c := range g.PendingChoices {
		if c != nil && c.Source == titan {
			t.Errorf("Sun Titan prompted with an empty graveyard")
		}
	}
	if triggersOnStackFrom(g, titan) != 0 {
		t.Errorf("Sun Titan put a trigger on the stack with no legal target")
	}
}
