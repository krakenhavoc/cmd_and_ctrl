package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planeswalker_statics_2797_test.go — #2797: the planeswalker-matters
// engine pieces the Reality Fracture set needs, proved on fixture
// planeswalkers so the tests do not depend on a card that also needs
// "empower Jace" (#2796).
//
//  1. a loyalty ability granted to a CLASS of planeswalkers (Kiora of Salt
//     and Sand is the real card), counted against CR 606.3;
//  2. the CR 704.5i exemption (Sanctum Lurker's static);
//  3. "whenever you put one or more loyalty counters on a planeswalker"
//     (Inspired Tethermage's trigger).

const (
	// A planeswalker with a +2 and a −3 and nothing else.
	Pw2797WalkerOracle = "test-2797-walker"
	// A creature whose whole text is the Lurker's static.
	Pw2797ExemptOracle = "test-2797-zero-loyalty-exempt"
	// A creature that silences the exempting fixture, for the
	// "lost its abilities" case.
	Pw2797SilencerOracle = "test-2797-silencer"
	// A creature with the Tethermage trigger: a +1/+1 counter on itself
	// for each placement.
	Pw2797WatcherOracle = "test-2797-loyalty-watcher"
	// A second grantor of a loyalty ability to planeswalkers you control.
	Pw2797GranterOracle = "test-2797-granter"
)

func init() {
	ran := func(g *game.Game, item *game.StackItem) error {
		return g.AddCounterForEffect(item.SourceCardID, "ran", 1)
	}
	effects.Register(effects.Spec{
		OracleID: Pw2797WalkerOracle,
		Name:     "Fixture Walker",
		// A fixture, but the catalog's census asks every planeswalker.
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{Label: "+2: Fixture.", Cost: effects.LoyaltyCost(2), Effect: ran},
			{Label: "−3: Fixture.", Cost: effects.LoyaltyCost(-3), Effect: ran},
		},
	})
	effects.Register(effects.Spec{
		OracleID: Pw2797GranterOracle,
		Name:     "Fixture Granter",
		Grants: []effects.AbilityGrant{{
			Key: "test-2797-granter/minus-one",
			Activated: []effects.ActivatedAbility{{
				Label: "−1: Fixture.", Cost: effects.LoyaltyCost(-1), Effect: ran,
			}},
			Text: "[−1]: Fixture.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToYourPlaneswalkers("test-2797-granter/minus-one")},
	})
	effects.Register(effects.Spec{
		OracleID:              Pw2797ExemptOracle,
		Name:                  "Fixture Lurker",
		ZeroLoyaltyExemptions: effects.PlaneswalkersSurviveZeroLoyalty(),
	})
	effects.Register(effects.Spec{
		OracleID: Pw2797SilencerOracle,
		Name:     "Fixture Silencer",
		Static: []game.StaticAbility{{
			Layer:            game.Layer6Ability,
			RemovesAbilities: true,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.OracleID == Pw2797ExemptOracle
			},
			Apply: func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
		}},
	})
	effects.Register(effects.Spec{
		OracleID: Pw2797WatcherOracle,
		Name:     "Fixture Watcher",
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouPutLoyaltyCountersOnAPlaneswalker("Fixture Watcher — +1/+1 counter",
				func(g *game.Game, item *game.StackItem) error {
					return g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, 1)
				}),
		},
	})
}

// fixtureWalker seats a planeswalker with `loyalty` counters placed the way
// the engine places them (an event per placement), so the loyalty trigger's
// reading of "before" has an arrival to anchor on.
func FixtureWalker(g *game.Game, owner uuid.UUID, oracle, typeLine string, loyalty int) uuid.UUID {
	instance := uuid.New()
	// A distinct name per walker, so two legendary fixtures never meet the
	// legend rule (CR 704.5j) and park the game on a prompt.
	id := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: instance, Name: "Fixture Walker " + instance.String()[:8], TypeLine: typeLine,
		OracleID: oracle, Owner: owner, Controller: owner,
	})
	if loyalty > 0 {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	}
	return id
}

func OnBattlefieldNow(g *game.Game, id uuid.UUID) bool {
	_, ok := BattlefieldCard(g, id)
	return ok
}

const KioraOracle2797 = "e70bb7f8-8098-49c5-929c-68f63cd821c5"

// --- 2. planeswalkers that aren't put into graveyards for having 0 loyalty --

func RunStateChecks(g *game.Game) { g.RunStateChecksForTest() }

// settleTriggers puts triggers a direct test write left pending onto the
// stack (the state check an activation or resolution would have run) and
// resolves them, without moving the step.
func SettleTriggers(t *testing.T, g *game.Game) {
	t.Helper()
	RunStateChecks(g)
	PassPriorityAroundTable(t, g)
}
