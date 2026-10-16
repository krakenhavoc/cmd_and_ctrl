package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// defender_waivers_test.go pins ADR 0106's 2026-10-16 amendment (#1936):
// "can attack players who attacked you during their last turn as though
// it didn't have defender" (Weathered Sentinels), asked per target, and
// the ModWaiveDefender mod's unconditional form (Hightide Hermit). The
// keyword stays in both (CR 609.4).

const sentinelsOracle = "test-weathered-sentinels"

// withSentinelsStatic installs the waiver static in the shape the catalog
// constructor writes it.
func withSentinelsStatic(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != sentinelsOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature()
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
				c.DefenderWaivers = append(c.DefenderWaivers, DefenderWaiver{
					Source: source.InstanceID, SourceName: source.Name, OnlyPlayersWhoAttackedYou: true,
				})
			},
		}}
	})
}

// pushReadyDefender puts a creature with defender under `controller`,
// not summoning sick.
func pushReadyDefender(g *Game, controller uuid.UUID, name, oracle string) uuid.UUID {
	id := pushTypedTestCard(g, Card{
		Name: name, TypeLine: "Artifact Creature — Wall", OracleID: oracle,
		Power: 2, Toughness: 5, Owner: controller, Controller: controller,
		Keywords: []string{"defender"},
	})
	g.WithWriteLock(func() { findBattlefieldCard(g, id).SummonedThisTurn = false })
	return id
}

func hasDefenderNow(g *Game, id uuid.UUID) bool {
	var out bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		out = HasKeyword(findBattlefieldCard(g, id), "defender")
	})
	return out
}

func TestSentinelsWaiverReachesOnlyPlayersWhoAttackedYou(t *testing.T) {
	withSentinelsStatic(t)
	g := newFourPlayerActiveGame(t)
	me, hitMe, hitMyWalker, hitElse := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	walker := pushTypedTestCard(g, Card{
		Name: "My Walker", TypeLine: "Legendary Planeswalker — Test", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{CounterLoyalty: 3},
	})
	theirWalker := pushTypedTestCard(g, Card{
		Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test", Owner: hitMe.ID, Controller: hitMe.ID,
		Counters: map[string]int{CounterLoyalty: 3},
	})
	// Seat 1 attacked me; seat 2 attacked only my planeswalker, which is
	// not me; seat 3 attacked seat 1.
	hitMe.LastTurnAttacks = []AttackRecord{{Attacker: uuid.New(), Defender: me.ID}}
	hitMyWalker.LastTurnAttacks = []AttackRecord{{Attacker: uuid.New(), Defender: walker, PlaneswalkerController: me.ID}}
	hitElse.LastTurnAttacks = []AttackRecord{{Attacker: uuid.New(), Defender: hitMe.ID}}
	sentinels := pushReadyDefender(g, me.ID, "Weathered Sentinels", sentinelsOracle)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, sentinels); !slices.Equal(got, []uuid.UUID{hitMe.ID}) {
		t.Fatalf("targets = %v: want only the player who attacked you (%s), not their planeswalker", got, hitMe.ID)
	}
	for _, refused := range []uuid.UUID{hitMyWalker.ID, hitElse.ID, theirWalker} {
		err := g.Clone().DeclareAttacker(sentinels, refused)
		if !errors.Is(err, ErrDefender) {
			t.Fatalf("declare at %s: err = %v, want ErrDefender", refused, err)
		}
		var we *DefenderWaiverError
		if !errors.As(err, &we) {
			t.Fatalf("refusal is %T, want *DefenderWaiverError", err)
		}
		if got, want := we.Sentence(), "Weathered Sentinels has defender: it can attack only players who attacked you during their last turn."; got != want {
			t.Errorf("sentence = %q, want %q", got, want)
		}
	}
	if err := g.DeclareAttacker(sentinels, hitMe.ID); err != nil {
		t.Fatalf("declare at the player who attacked you: %v", err)
	}
	if !hasDefenderNow(g, sentinels) {
		t.Error("the waiver took the keyword: it should still have defender (CR 609.4)")
	}
}

func TestSentinelsWaiverWithNobodyToAttackRefusesTheBulkVerb(t *testing.T) {
	withSentinelsStatic(t)
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentinels := pushReadyDefender(g, me.ID, "Weathered Sentinels", sentinelsOracle)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, sentinels); len(got) != 0 {
		t.Fatalf("targets = %v: nobody attacked you, so none", got)
	}
	if _, err := g.Clone().DeclareAttackers([]AttackDeclaration{{Attacker: sentinels, Target: opp.ID}}); !errors.Is(err, ErrNoLegalAttackers) {
		t.Fatalf("bulk declare: err = %v, want ErrNoLegalAttackers", err)
	}
	// The record is "you": the creature's controller. A player who
	// attacked someone else's seat does not open it.
	opp.LastTurnAttacks = []AttackRecord{{Attacker: uuid.New(), Defender: g.Seats[2].ID}}
	if got := attackTargetsOf(g, sentinels); len(got) != 0 {
		t.Fatalf("targets = %v: the attack was on another player", got)
	}
	// Read live at declaration: the waiver is on the creature, the
	// condition is asked of each target as it is named.
	opp.LastTurnAttacks = []AttackRecord{{Attacker: uuid.New(), Defender: me.ID}}
	if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: sentinels, Target: opp.ID}}); err != nil {
		t.Fatalf("bulk declare at the player who attacked you: %v", err)
	}
}

func TestSentinelsWaiverAnswersToItsCurrentController(t *testing.T) {
	withSentinelsStatic(t)
	g := newFourPlayerActiveGame(t)
	me, owner, hitOwner := g.Seats[0], g.Seats[1], g.Seats[2]
	// A Sentinels I took from seat 1. Seat 2 attacked seat 1, not me.
	hitOwner.LastTurnAttacks = []AttackRecord{{Attacker: uuid.New(), Defender: owner.ID}}
	sentinels := pushReadyDefender(g, me.ID, "Weathered Sentinels", sentinelsOracle)
	g.WithWriteLock(func() { findBattlefieldCard(g, sentinels).Owner = owner.ID })
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, sentinels); len(got) != 0 {
		t.Fatalf("targets = %v: \"you\" is its controller, whom nobody attacked", got)
	}
}

func TestWaiveDefenderModLetsItAttackAnyoneAndKeepsTheKeyword(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	hermit := pushReadyDefender(g, me.ID, "Hightide Hermit", "")
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.Clone().DeclareAttacker(hermit, g.Seats[1].ID); err != ErrDefender {
		t.Fatalf("before the waiver: err = %v, want ErrDefender", err)
	}

	registerScopedEffectForTest(t, g, hermit, []Mod{WaiveDefenderMod()}, IndefiniteDuration())
	got := attackTargetsOf(g, hermit)
	for _, p := range g.Seats[1:] {
		if !slices.Contains(got, p.ID) {
			t.Fatalf("targets = %v: want every opponent", got)
		}
	}
	if !hasDefenderNow(g, hermit) {
		t.Error("the waiver took the keyword: it should still have defender")
	}
	if err := g.DeclareAttacker(hermit, g.Seats[2].ID); err != nil {
		t.Fatalf("declare under the waiver: %v", err)
	}
}

func TestWaiveDefenderModSurvivesALaterLossOfAbilities(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	hermit := pushReadyDefender(g, me.ID, "Hightide Hermit", "")
	registerScopedEffectForTest(t, g, hermit, []Mod{WaiveDefenderMod()}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, hermit, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	// Defender comes back after both: the waiver is an effect, not an
	// ability, so it still lets the creature attack.
	registerScopedEffectForTest(t, g, hermit, []Mod{AddKeywordsMod("defender")}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)

	if !hasDefenderNow(g, hermit) {
		t.Fatal("setup: defender was granted back")
	}
	if got := attackTargetsOf(g, hermit); len(got) != 3 {
		t.Fatalf("targets = %v: want every opponent", got)
	}
}
