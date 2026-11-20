package testsg

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// homelands_next_untap_cards_test.go — the Homelands cards of #5753 on
// ADR 0058's one-shot "doesn't untap during … next untap step" mark:
// Reveka, Wizard Savant; Samite Alchemist; Spectral Bears; Labyrinth
// Minotaur; Joven's Ferrets. Each test plays the real catalog entry and
// checks that the mark is applied, shows as no_untap.next, and is used up
// by the right untap step.

const (
	revekaOracle            = "7085d5c6-da96-465e-90cf-8469fcacb474"
	samiteAlchemistOracle   = "120b565a-e22b-4d5d-9841-b11750b4ee49"
	spectralBearsOracle     = "539899f3-f3f2-4f7a-a6f6-e80d4a39f753"
	labyrinthMinotaurOracle = "1e7b075b-1c0f-4c44-a597-5c167a786cb1"
	jovensFerretsOracle     = "d2817b5f-f7a2-4254-a00f-e2e45d7a8cc0"
)

func TestHomelandsNextUntapCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		revekaOracle:            "Reveka, Wizard Savant",
		samiteAlchemistOracle:   "Samite Alchemist",
		spectralBearsOracle:     "Spectral Bears",
		labyrinthMinotaurOracle: "Labyrinth Minotaur",
		jovensFerretsOracle:     "Joven's Ferrets",
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

// hmlCreature puts a catalog creature with real P/T on the battlefield,
// not summoning sick.
func hmlCreature(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	id := effectstest.PushCatalogPermanent(g, owner, name, "Creature — Test", oracle, false)
	c := effectstest.FindBattlefieldCardForTest(g, id)
	c.Power, c.Toughness = power, toughness
	return id
}

// hmlNoUntapNext is the permanent's no_untap.next as `viewer` sees it:
// the players whose next untap step it sits out. Nil when the view
// carries no no_untap at all.
func hmlNoUntapNext(t *testing.T, g *game.Game, viewer, id uuid.UUID) []string {
	t.Helper()
	effectstest.ActivatedRowOfMarkKnown(g)
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID != id.String() {
			continue
		}
		if cv.NoUntap == nil {
			return nil
		}
		return cv.NoUntap.Next
	}
	t.Fatalf("card %s is not in the view", id)
	return nil
}

// hmlWantMarkedFor checks that `id` carries exactly one one-shot mark,
// shown in the view as `player`'s next untap step.
func hmlWantMarkedFor(t *testing.T, g *game.Game, id, player uuid.UUID) {
	t.Helper()
	skips := effectstest.HmlUntapSkips(t, g, id)
	if len(skips) != 1 || skips[0].While != nil || skips[0].Extra != 0 {
		t.Fatalf("marks = %+v, want one one-shot mark", skips)
	}
	if next := hmlNoUntapNext(t, g, player, id); !slices.Equal(next, []string{player.String()}) {
		t.Errorf("no_untap.next = %v, want [%s]", next, player)
	}
}

// --- Reveka, Wizard Savant --------------------------------------------------

// Reveka deals 2, stays tapped through her controller's next untap step,
// and untaps at the one after.
func TestRevekaDealsTwoAndSkipsYourNextUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reveka := hmlCreature(g, me.ID, "Reveka, Wizard Savant", revekaOracle, 0, 1)
	life := opp.Life

	if err := g.ActivateCatalogAbility(me.ID, reveka, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
	if !effectstest.TappedForTest(t, g, reveka) {
		t.Fatal("Reveka is not tapped")
	}
	hmlWantMarkedFor(t, g, reveka, me.ID)
	if skips := effectstest.HmlUntapSkips(t, g, reveka); skips[0].Player != me.ID {
		t.Errorf("mark keyed to %s, want the activator's untap step (%s)", skips[0].Player, me.ID)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedForTest(t, g, reveka) {
		t.Error("Reveka untapped during your next untap step")
	}
	if skips := effectstest.HmlUntapSkips(t, g, reveka); len(skips) != 0 {
		t.Errorf("the mark outlived the untap step it names: %+v", skips)
	}
	if next := hmlNoUntapNext(t, g, me.ID, reveka); len(next) != 0 {
		t.Errorf("no_untap.next = %v after the step, want none", next)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if effectstest.TappedForTest(t, g, reveka) {
		t.Error("Reveka stayed tapped at the untap step after")
	}
}

// The mark is used up by that untap step even when Reveka is already
// untapped by then (CR 502.3 only declines to untap her); it does not
// wait for a later step.
func TestRevekaMarkIsUsedUpEvenWhenSheIsUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reveka := hmlCreature(g, me.ID, "Reveka, Wizard Savant", revekaOracle, 0, 1)
	if err := g.ActivateCatalogAbility(me.ID, reveka, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.TapCard(reveka, false); err != nil {
		t.Fatalf("untap: %v", err)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if skips := effectstest.HmlUntapSkips(t, g, reveka); len(skips) != 0 {
		t.Fatalf("the mark survived the step on an untapped Reveka: %+v", skips)
	}
	if err := g.TapCard(reveka, true); err != nil {
		t.Fatalf("tap: %v", err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if effectstest.TappedForTest(t, g, reveka) {
		t.Error("the used-up mark held Reveka through a later untap step")
	}
}

// A target gone by resolution: the ability does nothing (CR 608.2b), so
// Reveka is not marked either.
func TestRevekaWithItsTargetGoneIsNotMarked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reveka := hmlCreature(g, me.ID, "Reveka, Wizard Savant", revekaOracle, 0, 1)
	victim := hmlCreature(g, opp.ID, "Victim", "", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, reveka, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(victim),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, victim); err != nil {
		t.Fatalf("bounce the target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if skips := effectstest.HmlUntapSkips(t, g, reveka); len(skips) != 0 {
		t.Errorf("Reveka was marked by an ability that did nothing: %+v", skips)
	}
}

// --- Samite Alchemist -------------------------------------------------------

// The creature is shielded from the next 4 damage, tapped, and held
// through its controller's next untap step.
func TestSamiteAlchemistShieldsTapsAndHoldsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alchemist := hmlCreature(g, me.ID, "Samite Alchemist", samiteAlchemistOracle, 0, 2)
	bear := hmlCreature(g, me.ID, "Bear", "", 2, 6)
	theirs := hmlCreature(g, opp.ID, "Theirs", "", 2, 2)

	effectstest.B06AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, alchemist, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Fatal("Samite Alchemist targeted a creature you don't control")
	}
	if err := g.ActivateCatalogAbility(me.ID, alchemist, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.TappedForTest(t, g, bear) {
		t.Fatal("the creature is not tapped")
	}
	hmlWantMarkedFor(t, g, bear, me.ID)
	if skips := effectstest.HmlUntapSkips(t, g, bear); skips[0].Player != me.ID {
		t.Errorf("mark keyed to %s, want the activator's untap step (%s)", skips[0].Player, me.ID)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 5) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 1 {
		t.Errorf("damage marked %d after 5 against a 4-point shield, want 1", got)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedForTest(t, g, bear) {
		t.Error("the creature untapped during your next untap step")
	}
	if skips := effectstest.HmlUntapSkips(t, g, bear); len(skips) != 0 {
		t.Errorf("the mark outlived its untap step: %+v", skips)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if effectstest.TappedForTest(t, g, bear) {
		t.Error("the creature stayed tapped at the untap step after")
	}
}

// The mark goes with the creature when it leaves the battlefield (CR
// 400.7): back as a new object, it untaps as usual.
func TestSamiteAlchemistMarkEndsWhenTheCreatureLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	alchemist := hmlCreature(g, me.ID, "Samite Alchemist", samiteAlchemistOracle, 0, 2)
	bear := hmlCreature(g, me.ID, "Bear", "", 2, 2)
	effectstest.B06AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, alchemist, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.HmlUntapSkips(t, g, bear)) != 1 {
		t.Fatal("the creature is not marked")
	}
	effectstest.BounceAndReturnForTest(t, g, bear, me.ID)
	if skips := effectstest.HmlUntapSkips(t, g, bear); len(skips) != 0 {
		t.Errorf("the returned creature kept the old object's mark: %+v", skips)
	}
}

// --- Spectral Bears ---------------------------------------------------------

// Attacking a player with no black nontoken permanent, the Bears stays
// tapped through your next untap step. A black TOKEN does not stop it.
func TestSpectralBearsSkipsYourNextUntapStepAgainstNoBlackPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bears := hmlCreature(g, me.ID, "Spectral Bears", spectralBearsOracle, 3, 3)
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Token Creature — Zombie",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})

	effectstest.DeclareAttack(t, g, opp.ID, bears)
	effectstest.PassPriorityAroundTable(t, g)
	hmlWantMarkedFor(t, g, bears, me.ID)

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedForTest(t, g, bears) {
		t.Error("the Bears untapped during your next untap step")
	}
	if skips := effectstest.HmlUntapSkips(t, g, bears); len(skips) != 0 {
		t.Errorf("the mark outlived its untap step: %+v", skips)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if effectstest.TappedForTest(t, g, bears) {
		t.Error("the Bears stayed tapped at the untap step after")
	}
}

// A black nontoken permanent when attackers are declared: no trigger.
func TestSpectralBearsDoesNotTriggerAgainstABlackPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bears := hmlCreature(g, me.ID, "Spectral Bears", spectralBearsOracle, 3, 3)
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Black Rock", TypeLine: "Artifact",
		Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID,
	})

	effectstest.DeclareAttack(t, g, opp.ID, bears)
	if n := effectstest.TriggersOnStackFrom(g, bears); n != 0 {
		t.Errorf("%d triggers from the Bears, want none", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if skips := effectstest.HmlUntapSkips(t, g, bears); len(skips) != 0 {
		t.Errorf("the Bears was marked: %+v", skips)
	}
}

// A black nontoken permanent that arrives before the trigger resolves:
// the intervening "if" fails the second check (CR 603.4) and nothing
// happens. Another player's black permanent never mattered.
func TestSpectralBearsChecksTheDefendingPlayerAgainAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bears := hmlCreature(g, me.ID, "Spectral Bears", spectralBearsOracle, 3, 3)
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Bystander's Black Rock", TypeLine: "Artifact",
		Colors: []string{"B"}, Owner: other.ID, Controller: other.ID,
	})

	effectstest.DeclareAttack(t, g, opp.ID, bears)
	if n := effectstest.TriggersOnStackFrom(g, bears); n != 1 {
		t.Fatalf("%d triggers from the Bears, want 1: only the defending player's permanents count", n)
	}
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Black Rock", TypeLine: "Artifact",
		Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.PassPriorityAroundTable(t, g)
	if skips := effectstest.HmlUntapSkips(t, g, bears); len(skips) != 0 {
		t.Errorf("the Bears was marked although the condition failed as it resolved: %+v", skips)
	}
}

// --- Labyrinth Minotaur -----------------------------------------------------

// The attacker the Minotaur blocks sits out its controller's next untap
// step and untaps at the one after.
func TestLabyrinthMinotaurHoldsTheCreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := hmlCreature(g, me.ID, "Attacker", "", 1, 5)
	minotaur := hmlCreature(g, opp.ID, "Labyrinth Minotaur", labyrinthMinotaurOracle, 1, 4)

	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(minotaur, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	hmlWantMarkedFor(t, g, attacker, me.ID)
	if skips := effectstest.HmlUntapSkips(t, g, attacker); skips[0].Player != uuid.Nil {
		t.Errorf("mark keyed to %s, want its controller's untap step (nil)", skips[0].Player)
	}
	if skips := effectstest.HmlUntapSkips(t, g, minotaur); len(skips) != 0 {
		t.Errorf("the Minotaur marked itself: %+v", skips)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedForTest(t, g, attacker) {
		t.Error("the blocked creature untapped during its controller's next untap step")
	}
	if skips := effectstest.HmlUntapSkips(t, g, attacker); len(skips) != 0 {
		t.Errorf("the mark outlived its untap step: %+v", skips)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if effectstest.TappedForTest(t, g, attacker) {
		t.Error("the creature stayed tapped at the untap step after")
	}
}

// An attacker that left and came back before the trigger resolved is a
// new object (CR 400.7): "that creature" is gone, and nothing is marked.
func TestLabyrinthMinotaurDoesNotMarkANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := hmlCreature(g, me.ID, "Attacker", "", 1, 5)
	minotaur := hmlCreature(g, opp.ID, "Labyrinth Minotaur", labyrinthMinotaurOracle, 1, 4)

	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(minotaur, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if n := effectstest.TriggersOnStackFrom(g, minotaur); n != 1 {
		t.Fatalf("%d triggers from the Minotaur, want 1", n)
	}
	effectstest.BounceAndReturnForTest(t, g, attacker, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if skips := effectstest.HmlUntapSkips(t, g, attacker); len(skips) != 0 {
		t.Errorf("the returned creature was marked: %+v", skips)
	}
}

// --- Joven's Ferrets --------------------------------------------------------

// The Ferrets gets +0/+2 as it attacks. At end of combat its blocker is
// tapped and sits out its controller's next untap step; a creature that
// did not block it is left alone.
func TestJovensFerretsTapsAndHoldsItsBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ferrets := hmlCreature(g, me.ID, "Joven's Ferrets", jovensFerretsOracle, 1, 1)
	blocker := hmlCreature(g, opp.ID, "Wall", "", 0, 4)
	bystander := hmlCreature(g, opp.ID, "Bystander", "", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, ferrets)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveToughness(t, g, ferrets); got != 3 {
		t.Errorf("Ferrets toughness %d after attacking, want 3", got)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, ferrets); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.TappedForTest(t, g, blocker) {
		t.Fatal("the blocker is not tapped")
	}
	hmlWantMarkedFor(t, g, blocker, opp.ID)
	if skips := effectstest.HmlUntapSkips(t, g, blocker); skips[0].Player != uuid.Nil {
		t.Errorf("mark keyed to %s, want its controller's untap step (nil)", skips[0].Player)
	}
	if effectstest.TappedForTest(t, g, bystander) || len(effectstest.HmlUntapSkips(t, g, bystander)) != 0 {
		t.Error("a creature that did not block the Ferrets was tapped or marked")
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.TappedForTest(t, g, blocker) {
		t.Error("the blocker untapped during its controller's next untap step")
	}
	if skips := effectstest.HmlUntapSkips(t, g, blocker); len(skips) != 0 {
		t.Errorf("the mark outlived its untap step: %+v", skips)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TappedForTest(t, g, blocker) {
		t.Error("the blocker stayed tapped at the untap step after")
	}
}

// A blocker that left the battlefield and came back before end of
// combat is a new object that never blocked the Ferrets (CR 400.7).
func TestJovensFerretsIgnoresABlockerThatCameBackAsANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ferrets := hmlCreature(g, me.ID, "Joven's Ferrets", jovensFerretsOracle, 1, 1)
	blocker := hmlCreature(g, opp.ID, "Wall", "", 0, 4)

	effectstest.DeclareAttack(t, g, opp.ID, ferrets)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, ferrets); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BounceAndReturnForTest(t, g, blocker, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.TappedForTest(t, g, blocker) || len(effectstest.HmlUntapSkips(t, g, blocker)) != 0 {
		t.Error("the returned creature was tapped or marked; it never blocked the Ferrets")
	}
}
