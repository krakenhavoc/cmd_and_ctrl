package testsa

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Hellrider --------------------------------------------------

// One trigger per attacking creature, and nothing happens until
// they resolve: the pings are stack objects with a response window,
// not a side effect of declaring.
func TestHellriderTriggersOncePerAttackerAndWaitsToResolve(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hell := effectstest.PushDiesCreatureForTest(g, me.ID, "Hellrider", effectstest.HellriderOracle, "Creature — Devil", 3, 3)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	lifeBefore := opp.Life

	// Hellrider counts itself — "a creature you control", not
	// "another" — so two attackers make two triggers.
	effectstest.DeclareAttack(t, g, opp.ID, hell, bear)

	if g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("step = %s, want the triggers to land in declare_attackers", g.Turn.Step)
	}
	if got := effectstest.TriggersOnStackFrom(g, hell); got != 2 {
		t.Fatalf("Hellrider triggers on the stack = %d, want 2", got)
	}
	if opp.Life != lifeBefore {
		t.Errorf("damage landed before the triggers resolved (life %d, was %d)", opp.Life, lifeBefore)
	}

	effectstest.PassPriorityAroundTable(t, g)
	if got := lifeBefore - opp.Life; got != 2 {
		t.Errorf("Hellrider dealt %d total, want 2 (one per attacker)", got)
	}
}

// The damaged player is the defender THAT attacker was declared
// against, not "an opponent". Two creatures attacking two different
// seats ping two different seats.
func TestHellriderFollowsEachAttackersOwnDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Hellrider", effectstest.HellriderOracle, "Creature — Devil", 3, 3)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	firstBefore, secondBefore := first.Life, second.Life

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("DeclareAttacker a: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("DeclareAttacker b: %v", err)
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hell := effectstest.PushDiesCreatureForTest(g, opp.ID, "Hellrider", effectstest.HellriderOracle, "Creature — Devil", 3, 3)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if got := effectstest.TriggersOnStackFrom(g, hell); got != 0 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krenko := effectstest.PushDiesCreatureForTest(g, me.ID, "Krenko, Tin Street Kingpin", effectstest.KrenkoKingpinOrcl,
		"Legendary Creature — Goblin", 1, 2)

	effectstest.DeclareAttack(t, g, opp.ID, krenko)
	if effectstest.TriggersOnStackFrom(g, krenko) != 1 {
		t.Fatalf("Krenko did not put exactly one trigger on the stack")
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Goblin"); got != 0 {
		t.Fatalf("%d Goblins made before the trigger resolved", got)
	}

	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCard(g, krenko)
	if !ok {
		t.Fatalf("Krenko left the battlefield")
	}
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("Krenko +1/+1 counters = %d, want 1", c.Counters["+1/+1"])
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Goblin"); got != 2 {
		t.Errorf("Goblins = %d, want 2 (power 1 + the counter placed first)", got)
	}
}

// TestKrenkoCounterThenGoblinsAfterAPausedPlacement is #1290's real-
// card exit criterion: with Doubling Season AND Hardened Scales both
// on the board, the +1/+1 counter placement pauses on a CR 616
// ordering prompt, and the Goblin count has to read Krenko's power
// AFTER the counter lands — not on the next line, which is where it
// used to be read before the fix moved it onto
// AddCounterByKeyedForEffect.
func TestKrenkoCounterThenGoblinsAfterAPausedPlacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	krenko := effectstest.PushDiesCreatureForTest(g, me.ID, "Krenko, Tin Street Kingpin", effectstest.KrenkoKingpinOrcl,
		"Legendary Creature — Goblin", 1, 2)

	effectstest.DeclareAttack(t, g, opp.ID, krenko)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Goblin"); got != 0 {
		t.Fatalf("%d Goblins made with the +1/+1 counter still owed to the CR 616 prompt", got)
	}

	// Doubling Season first: 1 → 2 → 3.
	effectstest.AnswerOrderBySource(t, g, ds, hs)

	c, ok := effectstest.BattlefieldCard(g, krenko)
	if !ok {
		t.Fatalf("Krenko left the battlefield")
	}
	if c.Counters["+1/+1"] != 3 {
		t.Errorf("Krenko +1/+1 counters = %d, want 3 ((1*2)+1 with Doubling Season first)", c.Counters["+1/+1"])
	}
	// Krenko's power is 1 (base) + 3 (counters) = 4, and Doubling
	// Season's OTHER half doubles the token creation itself, so the
	// real exit criterion is 8, not 4: the same board that proves the
	// counter read is correct also proves CreateToken still sees
	// Krenko's post-placement power rather than a stale pre-placement
	// one.
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Goblin"); got != 8 {
		t.Errorf("Goblins = %d, want 8 (4 power × 2 from Doubling Season's token half) — Goblins were made from Krenko's power before the paused placement landed", got)
	}
}

// "Whenever THIS creature attacks" — another attacker does nothing.
func TestKrenkoIgnoresOtherAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krenko := effectstest.PushDiesCreatureForTest(g, me.ID, "Krenko, Tin Street Kingpin", effectstest.KrenkoKingpinOrcl,
		"Legendary Creature — Goblin", 1, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if got := effectstest.TriggersOnStackFrom(g, krenko); got != 0 {
		t.Errorf("Krenko triggered %d times on another creature's attack", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Goblin"); got != 0 {
		t.Errorf("Goblins = %d, want 0", got)
	}
}

// --- Sun Titan --------------------------------------------------

// The attack half: declare Sun Titan, say yes, pick a cheap
// enough permanent in the graveyard, and it comes back to the
// battlefield. The mana-value clause filters the legal set.
func TestSunTitanAttackReanimatesChosenPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := effectstest.PushDiesCreatureForTest(g, me.ID, "Sun Titan", effectstest.SunTitanOracle, "Creature — Giant", 6, 6)
	cheap := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	dear := effectstest.PushGraveyardPermanent(me, "Serra Angel", "Creature — Angel", "{3}{W}{W}")

	effectstest.DeclareAttack(t, g, opp.ID, titan)

	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt after Yes")
	}
	if !effectstest.HasID(prompt.PickTargetCards, cheap) {
		t.Errorf("legal set missing the mana-value-2 permanent")
	}
	if effectstest.HasID(prompt.PickTargetCards, dear) {
		t.Errorf("mana-value-5 permanent offered as a Sun Titan target")
	}

	effectstest.PickCard(t, g, me.ID, cheap)
	if effectstest.TriggerOnStack(g, titan) == nil {
		t.Fatalf("Sun Titan trigger not on the stack after the pick")
	}
	if g.Battlefield.Contains(cheap) {
		t.Errorf("reanimation happened before the trigger resolved")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)

	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")

	effectstest.CastAndResolveCreature(t, g, "Sun Titan", "Creature — Giant", effectstest.SunTitanOracle)
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(cheap) {
		t.Errorf("Sun Titan's enters trigger did not reanimate")
	}
}

// CR 603.3d: a targeted trigger with an empty legal set is removed
// without a prompt. An empty graveyard means no question is asked.
func TestSunTitanWithEmptyGraveyardAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := effectstest.PushDiesCreatureForTest(g, me.ID, "Sun Titan", effectstest.SunTitanOracle, "Creature — Giant", 6, 6)

	effectstest.DeclareAttack(t, g, opp.ID, titan)
	for _, c := range g.PendingChoices {
		if c != nil && c.Source == titan {
			t.Errorf("Sun Titan prompted with an empty graveyard")
		}
	}
	if effectstest.TriggersOnStackFrom(g, titan) != 0 {
		t.Errorf("Sun Titan put a trigger on the stack with no legal target")
	}
}
